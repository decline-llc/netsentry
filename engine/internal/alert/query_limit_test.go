package alert

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func openQueryLimitStore(t *testing.T, daily bool) (*Store, string, time.Time) {
	t.Helper()
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	dir := filepath.Join(t.TempDir(), "query limit fixtures")
	store, err := Open(context.Background(), Options{
		Path: filepath.Join(dir, "alerts.db"), Dir: dir, DailyShard: daily,
		JournalMode: "DELETE", AggregationWindow: time.Minute,
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	return store, dir, now
}

func TestStoreQueryNegativeLimitsDoNotTruncateAcrossModes(t *testing.T) {
	const rows = 1005
	maxInt := int(^uint(0) >> 1)
	for _, daily := range []bool{false, true} {
		t.Run(fmt.Sprintf("daily=%t", daily), func(t *testing.T) {
			ctx := context.Background()
			store, dir, now := openQueryLimitStore(t, daily)
			fixtures := make([]*model.Alert, rows)
			allIndices := make([]int, 0, rows)
			highIndices := make([]int, 0, rows)
			for i := range fixtures {
				at := now.Add(-time.Duration(i) * time.Second)
				if i >= 500 {
					at = at.Add(-24 * time.Hour)
				}
				fixtures[i] = makeAlert(at, fmt.Sprintf("keyword-%04d", i))
				fixtures[i].RuleID = fmt.Sprintf("limit-%04d", i)
				allIndices = append(allIndices, i)
				if i == 2 || i == 4 {
					fixtures[i].Severity = model.SeverityLow
				} else {
					highIndices = append(highIndices, i)
				}
			}
			if len(highIndices) != 1003 {
				t.Fatal("filtered fixture must exceed 1000 rows")
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
			// A positive limit avoids depending on the negative-limit repair when
			// establishing immutable expected values for the boundary queries.
			baseline, total, err := store.Query(ctx, Query{Limit: rows})
			if err != nil || total != rows || len(baseline) != rows {
				t.Fatalf("fixture total=%d rows=%d error=%v, want %d", total, len(baseline), err, rows)
			}
			snapshots := make([]model.Alert, rows)
			for i, got := range baseline {
				if got == nil || got.RuleID != fmt.Sprintf("limit-%04d", i) || got.Severity != fixtures[i].Severity || got.AggregatedCount != 1 {
					t.Fatalf("fixture order/severity/count at %d: %+v", i, got)
				}
				snapshots[i] = *got
			}
			for _, tc := range []struct {
				name  string
				query Query
				total int
				want  []int
			}{
				{"negative_one", Query{Limit: -1}, rows, allIndices},
				{"negative_other", Query{Limit: -2}, rows, allIndices},
				{"negative_offset", Query{Limit: -1, Offset: 1}, rows, allIndices[1:]},
				{"negative_filtered_offset", Query{Limit: -1, Offset: 1, Severity: model.SeverityHigh}, 1003, highIndices[1:]},
				{"zero_default", Query{}, rows, allIndices[:1000]},
				{"zero_offset", Query{Offset: 5}, rows, allIndices[5:]},
				{"positive_large", Query{Limit: 1001}, rows, allIndices[:1001]},
				{"positive_normal", Query{Limit: 2, Offset: 1}, rows, allIndices[1:3]},
				{"negative_tail", Query{Limit: -1, Offset: 1000}, rows, allIndices[1000:]},
				{"at_end", Query{Limit: -1, Offset: rows}, rows, nil},
				{"past_end", Query{Limit: -1, Offset: rows + 1}, rows, nil},
				{"max_offset", Query{Limit: -1, Offset: maxInt}, rows, nil},
				{"normalized_negative_offset", Query{Limit: -1, Offset: -1}, rows, allIndices},
			} {
				t.Run(tc.name, func(t *testing.T) {
					got, total, err := store.Query(ctx, tc.query)
					if err != nil || total != tc.total || len(got) != len(tc.want) {
						t.Fatalf("Query=%+v total=%d rows=%d error=%v, want total=%d rows=%d", tc.query, total, len(got), err, tc.total, len(tc.want))
					}
					for i, index := range tc.want {
						if got[i] == nil || *got[i] != snapshots[index] {
							t.Fatalf("row[%d]=%+v, want complete alert %+v", i, got[i], snapshots[index])
						}
					}
					if health := store.Health(); health.Status != "ok" {
						t.Fatalf("query degraded storage: %+v", health)
					}
				})
			}
			count, err := store.Count(ctx)
			if err != nil || count != rows {
				t.Fatalf("count after queries=%d error=%v, want %d", count, err, rows)
			}
			after, total, err := store.Query(ctx, Query{Limit: rows})
			if err != nil || total != rows || len(after) != rows {
				t.Fatalf("full result after queries: total=%d rows=%d error=%v", total, len(after), err)
			}
			for i, want := range snapshots {
				if after[i] == nil || *after[i] != want {
					t.Fatalf("query changed row[%d]: %+v, want %+v", i, after[i], want)
				}
			}
		})
	}
}

func TestStoreQueryLimitSemanticsForEmptyResults(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, daily := range []bool{false, true} {
		for _, filtered := range []bool{false, true} {
			t.Run(fmt.Sprintf("daily=%t/filtered=%t", daily, filtered), func(t *testing.T) {
				ctx := context.Background()
				store, _, now := openQueryLimitStore(t, daily)
				countWant := 0
				if filtered {
					if err := store.WriteBatch(ctx, []*model.Alert{makeAlert(now, "nonmatching")}); err != nil {
						t.Fatal(err)
					}
					countWant = 1
				}
				for _, limit := range []int{-1, -2, 0, 1} {
					for _, offset := range []int{-1, 0, 1, maxInt} {
						query := Query{Limit: limit, Offset: offset}
						if filtered {
							query.RuleID = "missing-rule"
						}
						got, total, err := store.Query(ctx, query)
						if err != nil || total != 0 || len(got) != 0 {
							t.Fatalf("empty Query=%+v total=%d rows=%d error=%v", query, total, len(got), err)
						}
					}
				}
				count, err := store.Count(ctx)
				if err != nil || count != countWant || store.Health().Status != "ok" {
					t.Fatalf("empty query changed count/health: count=%d error=%v health=%+v", count, err, store.Health())
				}
			})
		}
	}
}
