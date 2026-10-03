package alert

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSliceBoundsClampsLargeLimitBeforeAdding(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name                              string
		length, limit, offset, start, end int
	}{
		{"empty", 0, maxInt, 0, 0, 0},
		{"first", 3, 1, 0, 0, 1},
		{"exact_last", 3, 2, 1, 1, 3},
		{"partial_last", 3, 2, 2, 2, 3},
		{"small_slice_large_limit", 3, maxInt, 1, 1, 3},
		{"small_slice_last_large_limit", 3, maxInt, 2, 2, 3},
		{"at_end", 3, maxInt, 3, 3, 3},
		{"past_end", 3, maxInt, 4, 3, 3},
		{"max_offset", 3, maxInt, maxInt, 3, 3},
		{"max_length", maxInt, maxInt, 1, 1, maxInt},
		{"max_length_last", maxInt, 2, maxInt - 1, maxInt - 1, maxInt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end := sliceBounds(tc.length, tc.limit, tc.offset)
			if start != tc.start || end != tc.end || start < 0 || start > end || end > tc.length {
				t.Fatalf("bounds=(%d,%d), want (%d,%d), length=%d", start, end, tc.start, tc.end, tc.length)
			}
		})
	}
}

func TestStoreQueryLargeLimitsAcrossPrimaryAndDailyShards(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, daily := range []bool{false, true} {
		name := "primary"
		if daily {
			name = "daily"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
			dir := filepath.Join(t.TempDir(), "query bounds fixtures")
			store, err := Open(ctx, Options{
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
			latest := makeAlert(now, "latest")
			latest.RuleID = "latest"
			middle := makeAlert(now.Add(-time.Hour), "middle")
			middle.RuleID = "middle"
			middle.Severity = model.SeverityLow
			oldest := makeAlert(now.Add(-24*time.Hour), "oldest")
			oldest.RuleID = "oldest"
			if err := store.WriteBatch(ctx, []*model.Alert{oldest, latest, middle}); err != nil {
				t.Fatal(err)
			}
			if daily {
				paths, err := filepath.Glob(filepath.Join(dir, "netsentry-*.db"))
				if err != nil || len(paths) != 2 {
					t.Fatalf("daily fixture paths=%v error=%v, want two actual shards", paths, err)
				}
			}
			baseline, total, err := store.Query(ctx, Query{Limit: 10})
			if err != nil || total != 3 || len(baseline) != 3 {
				t.Fatalf("fixture count=%d rows=%d error=%v", total, len(baseline), err)
			}
			snapshots := make([]model.Alert, 3)
			for i, id := range []string{"latest", "middle", "oldest"} {
				if baseline[i] == nil || baseline[i].RuleID != id || baseline[i].AggregatedCount != 1 {
					t.Fatalf("fixture row[%d]=%+v, want %s with count=1", i, baseline[i], id)
				}
				snapshots[i] = *baseline[i]
			}
			for _, tc := range []struct {
				name  string
				query Query
				total int
				want  []int
			}{
				{"max_first", Query{Limit: maxInt}, 3, []int{0, 1, 2}},
				{"max_middle", Query{Limit: maxInt, Offset: 1}, 3, []int{1, 2}},
				{"max_last", Query{Limit: maxInt, Offset: 2}, 3, []int{2}},
				{"max_filtered", Query{Limit: maxInt, Offset: 1, Severity: model.SeverityHigh}, 2, []int{2}},
				{"normal", Query{Limit: 1, Offset: 1}, 3, []int{1}},
				{"default_limit", Query{Offset: 1}, 3, []int{1, 2}},
				{"negative_offset", Query{Limit: 1, Offset: -1}, 3, []int{0}},
				{"at_end", Query{Limit: maxInt, Offset: 3}, 3, nil},
				{"past_end", Query{Limit: maxInt, Offset: 4}, 3, nil},
				{"max_offset", Query{Limit: maxInt, Offset: maxInt}, 3, nil},
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
			if err != nil || count != 3 {
				t.Fatalf("count after queries=%d error=%v, want 3", count, err)
			}
			after, total, err := store.Query(ctx, Query{Limit: 10})
			if err != nil || total != 3 || len(after) != 3 {
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
