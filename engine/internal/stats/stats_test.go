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
		"netsentry_capture_connected":       1,
		"netsentry_rules_loaded":            3,
		"netsentry_storage_available_bytes": 1024,
	})
	for _, want := range []string{
		"# TYPE netsentry_frames_total counter",
		"netsentry_frames_total 1",
		`netsentry_alerts_by_severity_total{severity="critical"} 1`,
		`netsentry_alerts_by_severity_total{severity="low"} 0`,
		"netsentry_rule_match_duration_seconds_total 0.25",
		"# HELP netsentry_rule_match_duration_seconds Rule match duration distribution.",
		"# TYPE netsentry_rule_match_duration_seconds histogram",
		`netsentry_rule_match_duration_seconds_bucket{le="0.1"} 0`,
		`netsentry_rule_match_duration_seconds_bucket{le="0.25"} 1`,
		`netsentry_rule_match_duration_seconds_bucket{le="+Inf"} 1`,
		"netsentry_rule_match_duration_seconds_sum 0.25",
		"netsentry_rule_match_duration_seconds_count 1",
		"# HELP netsentry_capture_connected Whether the capture heartbeat is currently fresh.",
		"netsentry_capture_connected 1",
		"# HELP netsentry_rules_loaded Current number of loaded rules.",
		"netsentry_rules_loaded 3",
		"# HELP netsentry_storage_available_bytes Available bytes on the alert storage filesystem.",
		"netsentry_storage_available_bytes 1024",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics body missing %q:\n%s", want, body)
		}
	}
}
