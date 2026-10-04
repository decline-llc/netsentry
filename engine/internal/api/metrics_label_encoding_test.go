package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestMetricsSeverityLabelEncodingPreservesHealthIdentity(t *testing.T) {
	metrics := stats.New()
	label := "custom\t\r\u00a0\\\"\ninjected_metric 99"
	entry := &model.Alert{Severity: model.Severity(label)}
	before := *entry
	metrics.ObserveAlerts([]*model.Alert{entry, nil, entry})
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, metrics)
	want := `netsentry_alerts_by_severity_total{severity="custom` + "\t\r\u00a0" + `\\\"\ninjected_metric 99"} 2`
	var first []string
	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/plain; version=0.0.4; charset=utf-8" {
			t.Fatalf("metrics response: %d %v", response.Code, response.Header())
		}
		var lines []string
		for _, line := range strings.Split(response.Body.String(), "\n") {
			if strings.HasPrefix(line, "netsentry_alerts_by_severity_total{") {
				lines = append(lines, line)
			}
			if strings.HasPrefix(line, "injected_metric ") {
				t.Fatalf("injected sample: %q", line)
			}
		}
		if len(lines) != 5 || strings.Count(response.Body.String(), want+"\n") != 1 || !strings.Contains(response.Body.String(), "\nnetsentry_alerts_generated_total 2\n") {
			t.Fatalf("missing exact label/counts: %q", response.Body.String())
		}
		for _, canonical := range []string{"critical", "high", "low", "medium"} {
			if strings.Count(response.Body.String(), `netsentry_alerts_by_severity_total{severity="`+canonical+"\"} 0\n") != 1 {
				t.Fatalf("canonical %s label changed", canonical)
			}
		}
		if i == 0 {
			first = lines
		} else if !reflect.DeepEqual(lines, first) {
			t.Fatalf("label lines changed on repeated scrape: %q vs %q", lines, first)
		}
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("health status=%d", response.Code)
	}
	var health struct {
		Throughput map[string]json.RawMessage `json:"throughput"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
		t.Fatal(err)
	}
	keys := []string{"frames_total", "control_frames", "packets_received", "packets_processed", "decode_errors", "alerts_generated", "worker_panics", "alert_write_errors", "alerts_by_severity"}
	if len(health.Throughput) != len(keys) {
		t.Fatalf("health shape changed: %v", health.Throughput)
	}
	for _, key := range keys {
		if _, ok := health.Throughput[key]; !ok {
			t.Fatalf("missing health field %s", key)
		}
	}
	var counts map[model.Severity]uint64
	if err := json.Unmarshal(health.Throughput["alerts_by_severity"], &counts); err != nil {
		t.Fatal(err)
	}
	if counts[model.Severity(label)] != 2 || len(counts) != 5 || string(health.Throughput["alerts_generated"]) != "2" || *entry != before {
		t.Fatalf("raw label identity/counts/input changed: %v", counts)
	}
}
