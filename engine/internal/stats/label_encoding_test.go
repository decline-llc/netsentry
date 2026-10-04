package stats

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestPrometheusSeverityLabelEncoding(t *testing.T) {
	cases := []struct{ name, value, encoded string }{
		{"canonical", "high", "high"},
		{"backslash", `custom\label`, `custom\\label`},
		{"quote", `custom"label`, `custom\"label`},
		{"newline", "custom\nlabel", `custom\nlabel`},
		{"tab", "custom\tlabel", "custom\tlabel"},
		{"carriage_return", "custom\rlabel", "custom\rlabel"},
		{"control", "custom\x01label", "custom\x01label"},
		{"delete", "custom\x7flabel", "custom\x7flabel"},
		{"unicode", "告警é🙂", "告警é🙂"},
		{"nonbreaking_space", "custom\u00a0label", "custom\u00a0label"},
		{"literal_escape", `custom\nlabel`, `custom\\nlabel`},
		{"mixed_injection", "custom\\\"\nfake_metric 99\t\r", `custom\\\"\nfake_metric 99` + "\t\r"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := &Stats{}
			entry := &model.Alert{Severity: model.Severity(tc.value)}
			before := *entry
			metrics.ObserveAlerts([]*model.Alert{entry, nil, entry})
			snapshot := metrics.Snapshot()
			wantCounts := map[model.Severity]uint64{model.Severity(tc.value): 2}
			if snapshot.AlertsGenerated != 2 || !reflect.DeepEqual(snapshot.AlertsBySeverity, wantCounts) {
				t.Fatalf("snapshot = %+v, want raw severity identity and count 2", snapshot)
			}
			body := RenderPrometheus(snapshot, nil)
			want := `netsentry_alerts_by_severity_total{severity="` + tc.encoded + "\"} 2"
			var lines []string
			for _, line := range strings.Split(body, "\n") {
				if strings.HasPrefix(line, "netsentry_alerts_by_severity_total{") {
					lines = append(lines, line)
				}
				if strings.HasPrefix(line, "fake_metric ") {
					t.Fatalf("injected sample: %q", line)
				}
			}
			if !reflect.DeepEqual(lines, []string{want}) {
				t.Fatalf("severity lines = %q, want %q", lines, want)
			}
			if got := decodePrometheusTestLabel(t, strings.TrimSuffix(strings.TrimPrefix(lines[0], `netsentry_alerts_by_severity_total{severity="`), `"} 2`)); got != tc.value {
				t.Fatalf("decoded value = %q, want %q", got, tc.value)
			}
			if *entry != before || !reflect.DeepEqual(snapshot.AlertsBySeverity, wantCounts) || RenderPrometheus(snapshot, nil) != body {
				t.Fatal("renderer changed input, snapshot or repeatability")
			}
		})
	}
}

// Independent format reader: no Go unquoting and no production escape helper.
func decodePrometheusTestLabel(t *testing.T, encoded string) string {
	t.Helper()
	var out strings.Builder
	for i := 0; i < len(encoded); i++ {
		switch encoded[i] {
		case '"', '\n':
			t.Fatalf("unescaped delimiter at byte %d: %q", i, encoded)
		case '\\':
			i++
			if i == len(encoded) {
				t.Fatalf("terminal escape: %q", encoded)
			}
			switch encoded[i] {
			case '\\', '"':
				out.WriteByte(encoded[i])
			case 'n':
				out.WriteByte('\n')
			default:
				t.Fatalf("unsupported escape at byte %d: %q", i, encoded)
			}
		default:
			out.WriteByte(encoded[i])
		}
	}
	return out.String()
}

func TestPrometheusSeverityLabelsPreserveSortAndCounts(t *testing.T) {
	metrics := New()
	labels := []string{"z\t", "a\n", `a\n`, "a\"", "告警"}
	for _, label := range labels {
		metrics.ObserveAlerts([]*model.Alert{{Severity: model.Severity(label)}})
	}
	snapshot := metrics.Snapshot()
	var raw []string
	for label := range snapshot.AlertsBySeverity {
		raw = append(raw, string(label))
	}
	sort.Strings(raw)
	var decoded []string
	for _, line := range strings.Split(RenderPrometheus(snapshot, nil), "\n") {
		prefix := `netsentry_alerts_by_severity_total{severity="`
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		end := strings.LastIndex(line, `"} `)
		if end < len(prefix) {
			t.Fatalf("invalid sample: %q", line)
		}
		label := decodePrometheusTestLabel(t, line[len(prefix):end])
		decoded = append(decoded, label)
		if want := fmt.Sprint(snapshot.AlertsBySeverity[model.Severity(label)]); line[end+3:] != want {
			t.Fatalf("count for %q: %q, want %s", label, line, want)
		}
	}
	if !reflect.DeepEqual(decoded, raw) || snapshot.AlertsGenerated != uint64(len(labels)) {
		t.Fatalf("decoded order = %q, want %q; total=%d", decoded, raw, snapshot.AlertsGenerated)
	}
}
