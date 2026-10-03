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

func TestStoreListAcrossModesKeepsGlobalOrderAndCap(t *testing.T) {
	const rows = 1005
	for _, daily := range []bool{false, true} {
		t.Run(fmt.Sprintf("daily=%t", daily), func(t *testing.T) {
			ctx := context.Background()
			store, dir, now := openQueryLimitStore(t, daily)
			fixtures := make([]*model.Alert, rows)
			for i := range fixtures {
				// Equal timestamps exercise the ID tie break. Reverse insertion
				// prevents insertion order from satisfying the ordering assertion.
				at := now.Add(-time.Duration(i/2) * time.Second)
				if i >= 500 {
					at = at.Add(-24 * time.Hour)
				}
				fixture := makeAlert(at, fmt.Sprintf("list-keyword-%04d", i))
				fixture.RuleID = fmt.Sprintf("list-%04d", i)
				fixture.PayloadPreview = fmt.Sprintf("list payload %04d", i)
				fixtures[rows-1-i] = fixture
			}
			if err := store.WriteBatch(ctx, fixtures); err != nil {
				t.Fatal(err)
			}
			if daily {
				paths, err := filepath.Glob(filepath.Join(dir, "netsentry-*.db"))
				if err != nil || len(paths) != 2 {
					t.Fatalf("daily fixture paths=%v error=%v, want two shards", paths, err)
				}
			}
			baseline, total, err := store.Query(ctx, Query{Limit: rows})
			if err != nil || total != rows || len(baseline) != rows {
				t.Fatalf("baseline total=%d rows=%d error=%v", total, len(baseline), err)
			}
			want := make([]model.Alert, rows)
			for i, got := range baseline {
				fixture := fixtures[rows-1-i]
				if got == nil || got.RuleID != fmt.Sprintf("list-%04d", i) ||
					got.MatchedKeyword != fixture.MatchedKeyword ||
					got.PayloadPreview != fixture.PayloadPreview ||
					!got.LastSeen.Equal(fixture.Timestamp) || got.AggregatedCount != 1 {
					t.Fatalf("baseline row[%d]=%+v, want fixture %+v", i, got, fixture)
				}
				want[i] = *got
			}
			for repeat := 0; repeat < 2; repeat++ {
				got, err := store.List(ctx)
				if err != nil || len(got) != 1000 {
					t.Fatalf("List rows=%d error=%v, want 1000", len(got), err)
				}
				for i := range got {
					if got[i] == nil || *got[i] != want[i] {
						t.Fatalf("List row[%d]=%+v, want complete alert %+v", i, got[i], want[i])
					}
				}
				if !got[500].LastSeen.Before(now.Truncate(24*time.Hour)) || store.Health().Status != "ok" {
					t.Fatalf("List must include historical rows and stay healthy: %+v", store.Health())
				}
			}
			count, err := store.Count(ctx)
			if err != nil || count != rows {
				t.Fatalf("count after List=%d error=%v, want %d", count, err, rows)
			}
			after, total, err := store.Query(ctx, Query{Limit: rows})
			if err != nil || total != rows || len(after) != rows {
				t.Fatalf("full results after List: total=%d rows=%d error=%v", total, len(after), err)
			}
			for i := range after {
				if after[i] == nil || *after[i] != want[i] {
					t.Fatalf("List changed row[%d]: %+v, want %+v", i, after[i], want[i])
				}
			}
		})
	}
}

func TestStoreListHistoricalRowsWithEmptyCurrentShard(t *testing.T) {
	ctx := context.Background()
	store, dir, now := openQueryLimitStore(t, true)
	fixture := makeAlert(now.Add(-24*time.Hour), "historical-only")
	fixture.RuleID = "list-historical"
	if err := store.WriteBatch(ctx, []*model.Alert{fixture}); err != nil {
		t.Fatal(err)
	}
	var currentCount int
	if err := store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts").Scan(&currentCount); err != nil || currentCount != 0 {
		t.Fatalf("current shard count=%d error=%v, want zero", currentCount, err)
	}
	paths, err := filepath.Glob(filepath.Join(dir, "netsentry-*.db"))
	if err != nil || len(paths) != 2 {
		t.Fatalf("historical fixture paths=%v error=%v, want two shards", paths, err)
	}
	baseline, total, err := store.Query(ctx, Query{Limit: 1})
	if err != nil || total != 1 || len(baseline) != 1 || baseline[0].RuleID != fixture.RuleID {
		t.Fatalf("historical baseline=%+v total=%d error=%v", baseline, total, err)
	}
	want := *baseline[0]
	got, err := store.List(ctx)
	if err != nil || len(got) != 1 || got[0] == nil || *got[0] != want {
		t.Fatalf("historical List=%+v error=%v, want %+v", got, err, want)
	}
	if store.Health().Status != "ok" {
		t.Fatalf("historical List degraded storage: %+v", store.Health())
	}
}

func TestStoreListEmptyCanceledAndClosedAcrossModes(t *testing.T) {
	for _, daily := range []bool{false, true} {
		t.Run(fmt.Sprintf("daily=%t", daily), func(t *testing.T) {
			store, _, _ := openQueryLimitStore(t, daily)
			got, err := store.List(context.Background())
			if err != nil || len(got) != 0 || store.Health().Status != "ok" {
				t.Fatalf("empty List=%+v error=%v health=%+v", got, err, store.Health())
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			got, err = store.List(ctx)
			if !errors.Is(err, context.Canceled) || len(got) != 0 || store.Health().Status != "ok" {
				t.Fatalf("pre-canceled List=%+v error=%v health=%+v", got, err, store.Health())
			}
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			got, err = store.List(context.Background())
			if err == nil || len(got) != 0 || store.Health().Status != "closed" {
				t.Fatalf("closed List=%+v error=%v health=%+v", got, err, store.Health())
			}
		})
	}
}

func TestStoreListRejectsCorruptHistoricalShard(t *testing.T) {
	store, dir, now := openQueryLimitStore(t, true)
	path := filepath.Join(dir, "netsentry-"+now.Add(-24*time.Hour).Format("2006-01-02")+".db")
	before := []byte("corrupt historical SQLite fixture")
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := store.List(context.Background())
	if err == nil || len(got) != 0 || !strings.Contains(err.Error(), "alert shard") {
		t.Fatalf("corrupt historical List=%+v error=%v, want shard error", got, err)
	}
	if health := store.Health(); health.Status != "degraded" || health.LastError == "" {
		t.Fatalf("corrupt historical List health=%+v, want degraded diagnostic", health)
	}
	assertFileBytesUnchanged(t, path, before)
}
