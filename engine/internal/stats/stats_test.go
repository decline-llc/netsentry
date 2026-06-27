package stats

import (
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSnapshotPreinitializesSeverityLabels(t *testing.T) {
	s := New()
	s.ObserveAlerts([]*model.Alert{{Severity: model.SeverityHigh}})

	snapshot := s.Snapshot()
	if snapshot.AlertsGenerated != 1 {
		t.Fatalf("alerts generated = %d, want 1", snapshot.AlertsGenerated)
	}
	for _, severity := range []model.Severity{model.SeverityLow, model.SeverityMedium, model.SeverityHigh, model.SeverityCritical} {
		if _, ok := snapshot.AlertsBySeverity[severity]; !ok {
			t.Fatalf("missing severity label %q", severity)
		}
	}
	if snapshot.AlertsBySeverity[model.SeverityHigh] != 1 {
		t.Fatalf("high severity count = %d, want 1", snapshot.AlertsBySeverity[model.SeverityHigh])
	}
}

func TestRenderPrometheusIncludesCountersLabelsAndGauges(t *testing.T) {
	s := New()
	s.IncFrame()
	s.IncPacketReceived()
	s.ObserveMatchDuration(250 * time.Millisecond)
	s.ObserveAlerts([]*model.Alert{{Severity: model.SeverityCritical}})

	body := RenderPrometheus(s.Snapshot(), map[string]float64{
		"netsentry_rules_loaded": 3,
	})
	for _, want := range []string{
		"# TYPE netsentry_frames_total counter",
		"netsentry_frames_total 1",
		`netsentry_alerts_by_severity_total{severity="critical"} 1`,
		`netsentry_alerts_by_severity_total{severity="low"} 0`,
		"netsentry_rule_match_duration_seconds_total 0.25",
		"netsentry_rules_loaded 3",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics body missing %q:\n%s", want, body)
		}
	}
}
