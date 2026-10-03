package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/stats"
)

func TestCompletionMetricsEndpointPreservesHealthJSON(t *testing.T) {
	metrics := stats.New()
	for i := 0; i < 2; i++ {
		metrics.IncPacketReceived()
		metrics.IncPacketProcessed()
	}
	metrics.IncPacketCompleted()
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, metrics)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/plain; version=0.0.4; charset=utf-8" {
		t.Fatalf("metrics response: %d, %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, line := range []string{
		"# HELP netsentry_packets_completed_total Packets completing pipeline processing and any enabled lifecycle export.",
		"# TYPE netsentry_packets_completed_total counter",
		"netsentry_packets_completed_total 1",
		"netsentry_packets_processed_total 2",
		"netsentry_packets_received_total 2",
	} {
		if strings.Count(body, line+"\n") != 1 {
			t.Fatalf("missing or duplicated export line %q: %s", line, body)
		}
	}
	if !strings.Contains(body, "\nnetsentry_packets_processed_per_second ") {
		t.Fatal("legacy processed rate gauge missing")
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("health response: %d, %s", response.Code, response.Body.String())
	}
	var health struct {
		Throughput map[string]json.RawMessage `json:"throughput"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
		t.Fatal(err)
	}
	keys := []string{"frames_total", "control_frames", "packets_received", "packets_processed", "decode_errors",
		"alerts_generated", "worker_panics", "alert_write_errors", "alerts_by_severity"}
	if len(health.Throughput) != len(keys) {
		t.Fatalf("health throughput shape changed: %v", health.Throughput)
	}
	for _, key := range keys {
		if _, ok := health.Throughput[key]; !ok {
			t.Fatalf("health key %q missing", key)
		}
	}
	if string(health.Throughput["packets_received"]) != "2" || string(health.Throughput["packets_processed"]) != "2" {
		t.Fatalf("health legacy counter values changed: %v", health.Throughput)
	}
}

func TestCompletionMetricsEndpointWithNilStats(t *testing.T) {
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "\nnetsentry_packets_completed_total 0\n") {
		t.Fatalf("nil-stats completion response: %d, %s", response.Code, response.Body.String())
	}
}
