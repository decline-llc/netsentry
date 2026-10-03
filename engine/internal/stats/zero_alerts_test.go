package stats_test

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestZeroAndConstructedStatsObserveAlertsPreserveSnapshotAndInputs(t *testing.T) {
	repeated := &model.Alert{RuleID: "repeated", Severity: model.SeverityHigh}
	for _, constructor := range []struct {
		name string
		new  func() *stats.Stats
	}{
		{"zero", func() *stats.Stats { return &stats.Stats{} }},
		{"constructed", stats.New},
	} {
		for _, tc := range []struct {
			name   string
			batch  []*model.Alert
			total  uint64
			counts map[model.Severity]uint64
		}{
			{"nil", nil, 0, nil},
			{"empty", []*model.Alert{}, 0, nil},
			{"all_nil", []*model.Alert{nil, nil}, 0, nil},
			{"single", []*model.Alert{{Severity: model.SeverityHigh}}, 1, map[model.Severity]uint64{model.SeverityHigh: 1}},
			{"mixed", []*model.Alert{nil, {Severity: model.SeverityHigh}, nil, {Severity: model.SeverityLow}}, 2,
				map[model.Severity]uint64{model.SeverityHigh: 1, model.SeverityLow: 1}},
			{"default", []*model.Alert{nil, {}}, 1, map[model.Severity]uint64{model.SeverityLow: 1}},
			{"all_severities", []*model.Alert{{Severity: model.SeverityLow}, {Severity: model.SeverityMedium}, {Severity: model.SeverityHigh}, {Severity: model.SeverityCritical}}, 4,
				map[model.Severity]uint64{model.SeverityLow: 1, model.SeverityMedium: 1, model.SeverityHigh: 1, model.SeverityCritical: 1}},
			{"repeated", []*model.Alert{repeated, nil, repeated}, 2, map[model.Severity]uint64{model.SeverityHigh: 2}},
			{"dynamic", []*model.Alert{nil, {Severity: "custom"}}, 1, map[model.Severity]uint64{"custom": 1}},
		} {
			t.Run(constructor.name+"/"+tc.name, func(t *testing.T) {
				metrics := constructor.new()
				metrics.IncFrame()
				metrics.IncControlFrame()
				metrics.IncPacketReceived()
				metrics.IncPacketProcessed()
				metrics.IncPacketCompleted()
				metrics.IncDecodeError()
				metrics.IncWorkerPanic()
				metrics.IncAlertWriteError()
				metrics.ObserveQueueDepth(7)
				metrics.ObserveMatchDuration(100 * time.Millisecond)
				metrics.ObserveAlertWriteDuration(250 * time.Millisecond)
				before := metrics.Snapshot()
				if before.StartedAt.IsZero() != (constructor.name == "zero") {
					t.Fatalf("unexpected initial start time: %v", before.StartedAt)
				}
				initialCounts := map[model.Severity]uint64{}
				if constructor.name == "constructed" {
					initialCounts = map[model.Severity]uint64{model.SeverityLow: 0, model.SeverityMedium: 0, model.SeverityHigh: 0, model.SeverityCritical: 0}
				}
				if !reflect.DeepEqual(before.AlertsBySeverity, initialCounts) {
					t.Fatalf("initial labels=%v, want %v", before.AlertsBySeverity, initialCounts)
				}
				pointers := append([]*model.Alert(nil), tc.batch...)
				values := zeroAlertValues(tc.batch)
				for observation := uint64(1); observation <= 2; observation++ {
					metrics.ObserveAlerts(tc.batch)
					want := before
					want.AlertsGenerated = tc.total * observation
					want.AlertsBySeverity = make(map[model.Severity]uint64)
					for severity, count := range before.AlertsBySeverity {
						want.AlertsBySeverity[severity] = count
					}
					for severity, count := range tc.counts {
						want.AlertsBySeverity[severity] += count * observation
					}
					assertZeroAlertSnapshot(t, metrics.Snapshot(), want)
					assertZeroAlertInputs(t, tc.batch, pointers, values)
					// Editing a returned map must not change the owned counters.
					snapshot := metrics.Snapshot()
					snapshot.AlertsBySeverity[model.SeverityHigh] = 99999
					snapshot.AlertsBySeverity["mutated"] = 123
					assertZeroAlertSnapshot(t, metrics.Snapshot(), want)
				}
			})
		}
	}
	var absent *stats.Stats
	absent.ObserveAlerts([]*model.Alert{nil, repeated})
	if got := absent.Snapshot(); !reflect.DeepEqual(got, stats.Snapshot{}) {
		t.Fatalf("nil Stats snapshot=%+v, want zero snapshot", got)
	}
}

func TestZeroStatsConcurrentFirstAlertObservationAfterJoin(t *testing.T) {
	const writers, observations = 4, 100
	var metrics stats.Stats
	before := metrics.Snapshot()
	batch := []*model.Alert{nil, {Severity: model.SeverityHigh}, {}, {Severity: "custom"}, nil}
	pointers := append([]*model.Alert(nil), batch...)
	values := zeroAlertValues(batch)
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
	want := before
	want.AlertsGenerated = 1200
	want.AlertsBySeverity = map[model.Severity]uint64{model.SeverityHigh: 400, model.SeverityLow: 400, "custom": 400}
	assertZeroAlertSnapshot(t, metrics.Snapshot(), want)
	assertZeroAlertInputs(t, batch, pointers, values)
}

func zeroAlertValues(batch []*model.Alert) []model.Alert {
	values := make([]model.Alert, len(batch))
	for i, alert := range batch {
		if alert != nil {
			values[i] = *alert
		}
	}
	return values
}

func assertZeroAlertInputs(t *testing.T, batch, pointers []*model.Alert, values []model.Alert) {
	t.Helper()
	for i, alert := range batch {
		if alert != pointers[i] || (alert != nil && *alert != values[i]) {
			t.Fatalf("input[%d] changed: %+v", i, alert)
		}
	}
}

func assertZeroAlertSnapshot(t *testing.T, got, want stats.Snapshot) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot=%+v, want complete snapshot %+v", got, want)
	}
	var sum uint64
	var expected []string
	for severity, count := range want.AlertsBySeverity {
		sum += count
		expected = append(expected, fmt.Sprintf("netsentry_alerts_by_severity_total{severity=%q} %d", severity, count))
	}
	if sum != want.AlertsGenerated {
		t.Fatalf("quiescent severity sum=%d, want %d", sum, want.AlertsGenerated)
	}
	text := stats.RenderPrometheus(got, nil)
	var actual []string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "netsentry_alerts_by_severity_total{") {
			actual = append(actual, line)
		}
	}
	sort.Strings(expected)
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("severity exposition=%v, want %v", actual, expected)
	}
	line := fmt.Sprintf("\nnetsentry_alerts_generated_total %d\n", want.AlertsGenerated)
	if strings.Count(text, line) != 1 {
		t.Fatalf("missing/duplicate total line %q", line)
	}
}
