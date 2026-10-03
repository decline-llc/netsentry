package stats

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func assertNilAlertCounts(t *testing.T, metrics *Stats, total uint64, counts map[model.Severity]uint64) {
	t.Helper()
	snapshot := metrics.Snapshot()
	if snapshot.AlertsGenerated != total || !reflect.DeepEqual(snapshot.AlertsBySeverity, counts) {
		t.Fatalf("alert total=%d severity=%v, want %d/%v", snapshot.AlertsGenerated, snapshot.AlertsBySeverity, total, counts)
	}
	var sum uint64
	for _, value := range snapshot.AlertsBySeverity {
		sum += value
	}
	if sum != total {
		t.Fatalf("quiescent severity sum=%d, want total %d", sum, total)
	}
	lines := strings.Split(RenderPrometheus(snapshot, nil), "\n")
	want := []string{
		"# TYPE netsentry_alerts_generated_total counter",
		fmt.Sprintf("netsentry_alerts_generated_total %d", total),
		"# TYPE netsentry_alerts_by_severity_total counter",
	}
	for severity, count := range counts {
		want = append(want, fmt.Sprintf("netsentry_alerts_by_severity_total{severity=%q} %d", severity, count))
	}
	for _, line := range want {
		matches := 0
		for _, got := range lines {
			if got == line {
				matches++
			}
		}
		if matches != 1 {
			t.Fatalf("metric line %q occurs %d times, want once", line, matches)
		}
	}
}

func TestObserveAlertsCountsOnlyNonNilEntries(t *testing.T) {
	repeated := &model.Alert{RuleID: "repeated", Severity: model.SeverityHigh}
	for _, tc := range []struct {
		name   string
		alerts []*model.Alert
		total  uint64
		counts map[model.Severity]uint64
	}{
		{"nil_batch", nil, 0, nil},
		{"empty_batch", []*model.Alert{}, 0, nil},
		{"all_nil", []*model.Alert{nil, nil}, 0, nil},
		{"single", []*model.Alert{{Severity: model.SeverityHigh}}, 1, map[model.Severity]uint64{model.SeverityHigh: 1}},
		{"mixed", []*model.Alert{nil, {Severity: model.SeverityHigh}, nil, {Severity: model.SeverityLow}, nil}, 2,
			map[model.Severity]uint64{model.SeverityHigh: 1, model.SeverityLow: 1}},
		{"default_severity", []*model.Alert{nil, {}, nil}, 1, map[model.Severity]uint64{model.SeverityLow: 1}},
		{"all_severities", []*model.Alert{{Severity: model.SeverityLow}, {Severity: model.SeverityMedium}, {Severity: model.SeverityHigh}, {Severity: model.SeverityCritical}}, 4,
			map[model.Severity]uint64{model.SeverityLow: 1, model.SeverityMedium: 1, model.SeverityHigh: 1, model.SeverityCritical: 1}},
		{"repeated_pointer", []*model.Alert{repeated, nil, repeated}, 2, map[model.Severity]uint64{model.SeverityHigh: 2}},
		{"dynamic_label", []*model.Alert{nil, {Severity: "custom"}}, 1, map[model.Severity]uint64{"custom": 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metrics := New()
			pointers := append([]*model.Alert(nil), tc.alerts...)
			values := make([]model.Alert, len(tc.alerts))
			for i, alert := range tc.alerts {
				if alert != nil {
					values[i] = *alert
				}
			}
			for observation := uint64(1); observation <= 2; observation++ {
				metrics.ObserveAlerts(tc.alerts)
				counts := map[model.Severity]uint64{
					model.SeverityLow: 0, model.SeverityMedium: 0, model.SeverityHigh: 0, model.SeverityCritical: 0,
				}
				for severity, count := range tc.counts {
					counts[severity] = count * observation
				}
				assertNilAlertCounts(t, metrics, tc.total*observation, counts)
				for i, alert := range tc.alerts {
					if alert != pointers[i] || (alert != nil && *alert != values[i]) {
						t.Fatalf("observation changed input[%d]: %+v", i, alert)
					}
				}
			}
		})
	}
	var absent *Stats
	absent.ObserveAlerts([]*model.Alert{nil, repeated})
	if snapshot := absent.Snapshot(); !reflect.DeepEqual(snapshot, Snapshot{}) {
		t.Fatalf("nil Stats snapshot=%+v, want zero value", snapshot)
	}
}

func TestObserveAlertsConcurrentNilBatchesAfterJoin(t *testing.T) {
	const writers, observations = 4, 100
	metrics := New()
	batch := []*model.Alert{nil, {Severity: model.SeverityHigh}, nil, {}, nil}
	start := make(chan struct{})
	var done sync.WaitGroup
	for i := 0; i < writers; i++ {
		done.Add(1)
		go func() {
			defer done.Done()
			<-start
			for j := 0; j < observations; j++ {
				metrics.ObserveAlerts(batch)
				metrics.ObserveAlerts([]*model.Alert{nil, nil})
			}
		}()
	}
	close(start)
	done.Wait()
	// Snapshot atomics are independently sampled; compare aggregates after join.
	assertNilAlertCounts(t, metrics, 800, map[model.Severity]uint64{
		model.SeverityLow: 400, model.SeverityMedium: 0, model.SeverityHigh: 400, model.SeverityCritical: 0,
	})
}
