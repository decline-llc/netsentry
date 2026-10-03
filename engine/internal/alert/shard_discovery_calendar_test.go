package alert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestDailyShardReadsIgnoreInvalidCalendarsWithoutModification(t *testing.T) {
	for _, currentRow := range []bool{false, true} {
		t.Run(fmt.Sprintf("current_row_%t", currentRow), func(t *testing.T) {
			ctx := context.Background()
			dir := filepath.Join(t.TempDir(), "daily shard read fixtures")
			now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
			store := openDailyShardStoreAt(t, dir, now)
			t.Cleanup(func() { _ = store.Close() })
			leap := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
			ordinary := time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC)
			fixtures := []*model.Alert{makeAlert(leap, "leap"), makeAlert(ordinary, "ordinary")}
			if currentRow {
				fixtures = append(fixtures, makeAlert(now, "current"))
			}
			for _, fixture := range fixtures {
				fixture.RuleID = "calendar-" + fixture.MatchedKeyword
			}
			if err := store.WriteBatch(ctx, fixtures); err != nil {
				t.Fatalf("write valid calendar shards: %v", err)
			}
			baseline, total, err := store.Query(ctx, Query{Limit: -1})
			if err != nil || total != len(fixtures) || len(baseline) != len(fixtures) {
				t.Fatalf("baseline rows=%d total=%d err=%v", len(baseline), total, err)
			}
			for i, got := range baseline {
				fixture := fixtures[len(fixtures)-1-i]
				if got == nil || got.RuleID != fixture.RuleID || got.MatchedKeyword != fixture.MatchedKeyword ||
					!got.LastSeen.Equal(fixture.Timestamp) || got.AggregatedCount != 1 {
					t.Fatalf("baseline row %d = %+v, want fixed fixture %+v", i, got, fixture)
				}
			}
			retained := make(map[string][]byte)
			for _, name := range []string{
				"netsentry-2025-02-29.db", "netsentry-2025-02-30.db", "netsentry-2025-04-31.db",
				"netsentry-2025-00-01.db", "netsentry-2025-13-01.db",
				"netsentry-2025-01-00.db", "netsentry-2025-01-32.db",
				"netsentry-2025-2-01.db", "notes-2025-01-01.db",
			} {
				for path, contents := range calendarFixtureSet(t, dir, name) {
					retained[path] = contents
				}
			}
			for _, suffix := range []string{"-wal", "-shm"} {
				path := filepath.Join(dir, "netsentry-2020-01-01.db"+suffix)
				retained[path] = calendarFixtureFile(t, path)
			}
			folder := filepath.Join(dir, "netsentry-2020-01-02.db")
			if err := os.Mkdir(folder, 0o700); err != nil {
				t.Fatalf("create date-shaped directory: %v", err)
			}
			path := filepath.Join(folder, "nested")
			retained[path] = calendarFixtureFile(t, path)
			checkBytes := func() {
				t.Helper()
				for path, contents := range retained {
					assertFileBytes(t, path, contents)
				}
			}
			since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
			until := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
			ordinaryIndex := 0
			if currentRow {
				ordinaryIndex = 1
			}
			operations := []struct {
				name  string
				run   func(context.Context) ([]*model.Alert, int, error)
				want  []*model.Alert
				total int
			}{
				{"list", func(ctx context.Context) ([]*model.Alert, int, error) {
					rows, err := store.List(ctx)
					return rows, len(rows), err
				}, baseline, len(baseline)},
				{"count", func(ctx context.Context) ([]*model.Alert, int, error) {
					count, err := store.Count(ctx)
					return nil, count, err
				}, nil, len(baseline)},
				{"query_all", func(ctx context.Context) ([]*model.Alert, int, error) {
					return store.Query(ctx, Query{Limit: -1})
				}, baseline, len(baseline)},
				{"query_page", func(ctx context.Context) ([]*model.Alert, int, error) {
					return store.Query(ctx, Query{Limit: 1, Offset: 1})
				}, baseline[1:2], len(baseline)},
				{"query_range", func(ctx context.Context) ([]*model.Alert, int, error) {
					return store.Query(ctx, Query{Since: &since, Until: &until, Limit: -1})
				}, baseline[ordinaryIndex : ordinaryIndex+1], 1},
			}
			for repeat := 0; repeat < 2; repeat++ {
				for _, operation := range operations {
					got, total, err := operation.run(ctx)
					if err != nil || total != operation.total {
						t.Fatalf("%s total=%d err=%v, want %d", operation.name, total, err, operation.total)
					}
					assertCalendarRows(t, operation.name, got, operation.want)
					if health := store.Health(); health.Status != "ok" {
						t.Fatalf("%s health=%+v, want ok", operation.name, health)
					}
					checkBytes()
				}
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			for _, operation := range operations {
				got, total, err := operation.run(canceled)
				if !errors.Is(err, context.Canceled) || total != 0 || len(got) != 0 || store.Health().Status != "ok" {
					t.Fatalf("canceled %s rows=%d total=%d err=%v health=%+v", operation.name, len(got), total, err, store.Health())
				}
				checkBytes()
			}
			if err := store.Close(); err != nil {
				t.Fatalf("close daily store: %v", err)
			}
			checkBytes()
		})
	}
}

func TestDailyShardCalendarDiscoveryStillRejectsValidDateCorruption(t *testing.T) {
	for _, operation := range []string{"list", "query", "count"} {
		t.Run(operation, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "valid date corrupt shard")
			now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
			store := openDailyShardStoreAt(t, dir, now)
			defer store.Close()
			path := filepath.Join(dir, "netsentry-2026-06-26.db")
			before := calendarFixtureFile(t, path)
			var rows []*model.Alert
			var count int
			var err error
			switch operation {
			case "list":
				rows, err = store.List(context.Background())
			case "query":
				rows, count, err = store.Query(context.Background(), Query{Limit: -1})
			case "count":
				count, err = store.Count(context.Background())
			}
			if err == nil || !strings.Contains(err.Error(), "alert shard") || len(rows) != 0 || count != 0 {
				t.Fatalf("%s rows=%d count=%d err=%v, want shard error", operation, len(rows), count, err)
			}
			if health := store.Health(); health.Status != "degraded" || health.LastError == "" {
				t.Fatalf("%s health=%+v, want degraded diagnostic", operation, health)
			}
			assertFileBytes(t, path, before)
		})
	}
}

func assertCalendarRows(t *testing.T, operation string, got, want []*model.Alert) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s rows=%d, want %d", operation, len(got), len(want))
	}
	for i := range got {
		if got[i] == nil || *got[i] != *want[i] {
			t.Fatalf("%s row %d = %+v, want full alert %+v", operation, i, got[i], want[i])
		}
	}
}
