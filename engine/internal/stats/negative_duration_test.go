package stats

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

// This independent fixture pins all exported bounds rather than reading the
// implementation's bucket array to derive the expected distribution.
var durationTestBounds = [...]float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

func observeDurationForTest(s *Stats, match bool, d time.Duration) {
	if match {
		s.ObserveMatchDuration(d)
	} else {
		s.ObserveAlertWriteDuration(d)
	}
}

func assertDurationObservation(t *testing.T, snapshot Snapshot, match, both bool, count uint64, sum time.Duration, counts [13]uint64) {
	t.Helper()
	zeroBuckets, buckets := make([]HistogramBucket, 13), make([]HistogramBucket, 13)
	for i, bound := range durationTestBounds {
		zeroBuckets[i] = HistogramBucket{Le: bound}
		buckets[i] = HistogramBucket{Le: bound, Count: counts[i]}
	}
	want := Snapshot{
		StartedAt: snapshot.StartedAt, MatchBuckets: zeroBuckets,
		AlertWriteBuckets: zeroBuckets,
		AlertsBySeverity: map[model.Severity]uint64{
			model.SeverityLow: 0, model.SeverityMedium: 0,
			model.SeverityHigh: 0, model.SeverityCritical: 0,
		},
	}
	prefix := "netsentry_alert_write_duration_seconds"
	if match {
		want.MatchCount, want.MatchDurationNS, want.MatchBuckets = count, uint64(sum), buckets
		prefix = "netsentry_rule_match_duration_seconds"
	} else {
		want.AlertWriteCount, want.AlertWriteNS, want.AlertWriteBuckets = count, uint64(sum), buckets
	}
	if both {
		want.MatchCount, want.MatchDurationNS, want.MatchBuckets = count, uint64(sum), buckets
		want.AlertWriteCount, want.AlertWriteNS, want.AlertWriteBuckets = count, uint64(sum), buckets
	}
	if !reflect.DeepEqual(snapshot, want) {
		t.Fatalf("snapshot = %+v, want %+v", snapshot, want)
	}
	body := RenderPrometheus(snapshot, nil)
	lines := []string{
		"# TYPE " + prefix + " histogram",
		"# TYPE " + prefix + "_total counter",
		fmt.Sprintf("%s_total %v", prefix, sum.Seconds()),
		fmt.Sprintf("%s_sum %v", prefix, sum.Seconds()),
		fmt.Sprintf("%s_count %d", prefix, count),
		fmt.Sprintf("%s_bucket{le=\"+Inf\"} %d", prefix, count),
	}
	for i, bound := range durationTestBounds {
		lines = append(lines, fmt.Sprintf("%s_bucket{le=\"%v\"} %d", prefix, bound, counts[i]))
	}
	if match {
		lines = append(lines, fmt.Sprintf("netsentry_rule_match_total %d", count))
	}
	for _, wantLine := range lines {
		matches := 0
		for _, line := range strings.Split(body, "\n") {
			if line == wantLine {
				matches++
			}
		}
		if matches != 1 {
			t.Fatalf("expected exactly one line %q, found %d in:\n%s", wantLine, matches, body)
		}
	}
}

func TestNegativeDurationObservationsPreserveSnapshotAndMetrics(t *testing.T) {
	for _, match := range []bool{true, false} {
		for _, seed := range []struct {
			name string
			d    time.Duration
			add  bool
		}{
			{name: "fresh"},
			{name: "zero", add: true},
			{name: "positive", d: 250 * time.Millisecond, add: true},
		} {
			t.Run(fmt.Sprintf("match_%t/%s", match, seed.name), func(t *testing.T) {
				s := New()
				var count uint64
				var buckets [13]uint64
				if seed.add {
					observeDurationForTest(s, match, seed.d)
					count = 1
					first := 0
					if seed.d != 0 {
						first = 8
					}
					for i := first; i < 13; i++ {
						buckets[i] = 1
					}
				}
				assertDurationObservation(t, s.Snapshot(), match, false, count, seed.d, buckets)
				before, textBefore := s.Snapshot(), RenderPrometheus(s.Snapshot(), nil)
				for _, negative := range []time.Duration{-1, -time.Second, time.Duration(-1 << 63)} {
					observeDurationForTest(s, match, negative)
					if after := s.Snapshot(); !reflect.DeepEqual(after, before) {
						t.Fatalf("negative %v changed snapshot: %+v -> %+v", negative, before, after)
					}
					if textAfter := RenderPrometheus(s.Snapshot(), nil); textAfter != textBefore {
						t.Fatalf("negative %v changed exposition:\n%s", negative, textAfter)
					}
					assertDurationObservation(t, s.Snapshot(), match, false, count, seed.d, buckets)
				}
			})
		}
	}
	var absent *Stats
	for _, match := range []bool{true, false} {
		for _, d := range []time.Duration{-1, time.Duration(-1 << 63), 0, time.Second} {
			observeDurationForTest(absent, match, d)
		}
	}
	if !reflect.DeepEqual(absent.Snapshot(), Snapshot{}) {
		t.Fatal("nil receiver snapshot changed")
	}
}

func TestNonnegativeDurationObservationsRetainEveryBucketBoundary(t *testing.T) {
	cases := []struct {
		name  string
		d     time.Duration
		first int
	}{
		{name: "zero", first: 0},
		{name: "one_nanosecond", d: 1, first: 0},
		{name: "above_largest_bucket", d: 6 * time.Second, first: 13},
	}
	boundsNS := [...]time.Duration{100000, 500000, 1000000, 5000000, 10000000, 25000000, 50000000, 100000000, 250000000, 500000000, 1000000000, 2500000000, 5000000000}
	for i, d := range boundsNS {
		cases = append(cases, struct {
			name  string
			d     time.Duration
			first int
		}{fmt.Sprintf("boundary_%d", i), d, i}, struct {
			name  string
			d     time.Duration
			first int
		}{fmt.Sprintf("above_boundary_%d", i), d + 1, i + 1})
	}
	for _, match := range []bool{true, false} {
		for _, tc := range cases {
			t.Run(fmt.Sprintf("match_%t/%s", match, tc.name), func(t *testing.T) {
				s := New()
				observeDurationForTest(s, match, tc.d)
				var buckets [13]uint64
				for i := tc.first; i < 13; i++ {
					buckets[i] = 1
				}
				assertDurationObservation(t, s.Snapshot(), match, false, 1, tc.d, buckets)
			})
		}
	}
}

func TestConcurrentDurationObservationsAfterWritersJoin(t *testing.T) {
	s := New()
	const writers, observations = 4, 100
	start := make(chan struct{})
	var done sync.WaitGroup
	done.Add(writers)
	for i := 0; i < writers; i++ {
		go func() {
			defer done.Done()
			<-start
			for n := 0; n < observations; n++ {
				for _, match := range []bool{true, false} {
					for _, d := range []time.Duration{-1, time.Duration(-1 << 63), 0, 250 * time.Millisecond} {
						observeDurationForTest(s, match, d)
					}
				}
			}
		}()
	}
	close(start)
	done.Wait()
	// Snapshot counters are individually sampled. Assert totals only after all
	// writes have stopped; this does not promise a transactional live snapshot.
	snapshot := s.Snapshot()
	for _, match := range []bool{true, false} {
		assertDurationObservation(t, snapshot, match, true, 800, 100*time.Second, [13]uint64{400, 400, 400, 400, 400, 400, 400, 400, 800, 800, 800, 800, 800})
	}
}
