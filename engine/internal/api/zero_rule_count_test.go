package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestAPIZeroValueRuleEngineAcrossReloadLifecycle(t *testing.T) {
	var engine rule.Engine
	server := NewServer(&fakeStore{}, fakeQueue{depth: 3}, &engine, stats.New())
	handler := server.Handler()
	fixture := &model.Rule{ID: "api-zero", Name: "API zero fixture", Type: model.RuleTypePortBlacklist, Severity: model.SeverityHigh, Priority: 100, Enabled: true, Config: json.RawMessage(`{"ports":[80]}`)}
	phases := []struct {
		name  string
		apply func() error
		count int
		rules []*model.Rule
	}{
		{name: "unpublished"},
		{name: "loaded", apply: func() error { return engine.Reload([]*model.Rule{fixture}) }, count: 1, rules: []*model.Rule{fixture}},
		{name: "rejected_reload", apply: func() error {
			err := engine.Reload([]*model.Rule{nil})
			if err == nil || !strings.Contains(err.Error(), "rule at index 0 is null") {
				return fmt.Errorf("invalid reload error = %v", err)
			}
			return nil
		}, count: 1, rules: []*model.Rule{fixture}},
		{name: "cleared", apply: func() error { return engine.Reload(nil) }},
	}
	for _, phase := range phases {
		t.Run(phase.name, func(t *testing.T) {
			if phase.apply != nil {
				if err := phase.apply(); err != nil {
					t.Fatal(err)
				}
			}
			for _, path := range []string{"/api/health", "/api/health?verbose=true", "/api/metrics", "/api/rules"} {
				t.Run(path, func(t *testing.T) {
					response := httptest.NewRecorder()
					handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
					if response.Code != http.StatusOK {
						t.Fatalf("status = %d, body %s", response.Code, response.Body.String())
					}
					if path == "/api/metrics" {
						if got := response.Header().Get("Content-Type"); got != "text/plain; version=0.0.4; charset=utf-8" {
							t.Fatalf("metrics content type = %q", got)
						}
						for _, line := range []string{fmt.Sprintf("netsentry_rules_loaded %d", phase.count), "netsentry_alerts_current 0", "netsentry_packet_queue_depth 3", "netsentry_packets_processed_total 0", "netsentry_alerts_generated_total 0"} {
							if strings.Count(response.Body.String(), "\n"+line+"\n") != 1 {
								t.Fatalf("missing or repeated line %q in %s", line, response.Body.String())
							}
						}
						return
					}
					if got := response.Header().Get("Content-Type"); got != "application/json" {
						t.Fatalf("JSON content type = %q", got)
					}
					var document map[string]json.RawMessage
					if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
						t.Fatal(err)
					}
					switch path {
					case "/api/rules":
						var got ruleListResponse
						if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
							t.Fatal(err)
						}
						if len(document) != 1 || !reflect.DeepEqual(got.Data, phase.rules) {
							t.Fatalf("rule JSON = %s, want %+v", response.Body.String(), phase.rules)
						}
					case "/api/health":
						if len(document) != 2 || string(document["status"]) != `"ok"` || string(document["alerts"]) != "0" {
							t.Fatalf("basic health changed: %s", response.Body.String())
						}
					default:
						var got verboseHealthResponse
						if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
							t.Fatal(err)
						}
						if len(document) != 6 || got.Status != "ok" || got.Alerts != 0 || got.Engine.RulesLoaded != phase.count || got.Engine.QueueDepth != 3 || got.Storage.Alerts != 0 || got.Storage.Status != "ok" || got.Throughput.PacketsProcessed != 0 || got.Throughput.AlertsGenerated != 0 {
							t.Fatalf("verbose health changed: %s", response.Body.String())
						}
					}
				})
			}
		})
	}
}
