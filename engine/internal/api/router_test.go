package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/receiver"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type fakeStore struct {
	alerts []*model.Alert
	err    error
	health alert.StorageHealth
}

func (s *fakeStore) List(ctx context.Context) ([]*model.Alert, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.alerts, nil
}

func (s *fakeStore) Count(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return len(s.alerts), nil
}

func (s *fakeStore) Health() alert.StorageHealth {
	if s.health.Status == "" {
		return alert.StorageHealth{Status: "ok"}
	}
	return s.health
}

type fakeStoreWithPath struct {
	fakeStore
	path string
}

func (s *fakeStoreWithPath) Path() string { return s.path }

type fakeQueue struct{ depth int }

func (q fakeQueue) QueueDepth() int { return q.depth }

type fakeHealthQueue struct {
	depth int
	state receiver.State
}

func (q fakeHealthQueue) QueueDepth() int { return q.depth }

func (q fakeHealthQueue) State() receiver.State { return q.state }

type fakeRules struct {
	count     int
	rules     []*model.Rule
	reloadErr error
	reloaded  []*model.Rule
}

func (r *fakeRules) RuleCount() int { return r.count }

func (r *fakeRules) Rules() []*model.Rule { return r.rules }

func (r *fakeRules) Reload(rules []*model.Rule) error {
	if r.reloadErr != nil {
		return r.reloadErr
	}
	r.reloaded = rules
	r.rules = rules
	r.count = len(rules)
	return nil
}

func TestHealthMinimalResponse(t *testing.T) {
	server := NewServer(&fakeStore{alerts: []*model.Alert{{RuleID: "rule-1"}}}, fakeQueue{depth: 7}, &fakeRules{count: 3}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["status"] != "ok" || got["alerts"] != float64(1) {
		t.Fatalf("unexpected health response: %+v", got)
	}
	if _, ok := got["capture"]; ok {
		t.Fatalf("minimal health should not include verbose fields: %+v", got)
	}
}

func TestHealthVerboseResponse(t *testing.T) {
	metrics := stats.New()
	metrics.IncFrame()
	metrics.IncPacketReceived()
	queue := fakeHealthQueue{
		depth: 4,
		state: receiver.State{
			SessionID: "session-1",
			Hello: receiver.HelloFrame{
				Type:      "hello",
				Version:   "0.1.0",
				SessionID: "session-1",
				PID:       42,
			},
			Heartbeat: receiver.HeartbeatFrame{
				Type:      "heartbeat",
				SessionID: "session-1",
				Seq:       7,
				Sent:      10,
			},
			LastHeartbeatAt: time.Now().UTC(),
		},
	}
	server := NewServerWithOptions(&fakeStore{alerts: []*model.Alert{{RuleID: "rule-1"}}}, queue, &fakeRules{count: 3}, metrics, Options{HealthFreshnessLimit: time.Minute})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Status  string `json:"status"`
		Alerts  int    `json:"alerts"`
		Capture struct {
			Status    string `json:"status"`
			SessionID string `json:"session_id"`
		} `json:"capture"`
		Engine struct {
			QueueDepth  int `json:"queue_depth"`
			RulesLoaded int `json:"rules_loaded"`
		} `json:"engine"`
		Storage struct {
			Status string `json:"status"`
			Alerts int    `json:"alerts"`
		} `json:"storage"`
		Throughput struct {
			FramesTotal     uint64 `json:"frames_total"`
			PacketsReceived uint64 `json:"packets_received"`
		} `json:"throughput"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v\n%s", err, rec.Body.String())
	}
	if got.Status != "ok" || got.Alerts != 1 {
		t.Fatalf("unexpected health status: %+v", got)
	}
	if got.Capture.Status != "ok" || got.Capture.SessionID != "session-1" {
		t.Fatalf("unexpected capture health: %+v", got.Capture)
	}
	if got.Engine.QueueDepth != 4 || got.Engine.RulesLoaded != 3 {
		t.Fatalf("unexpected engine health: %+v", got.Engine)
	}
	if got.Storage.Status != "ok" || got.Storage.Alerts != 1 {
		t.Fatalf("unexpected storage health: %+v", got.Storage)
	}
	if got.Throughput.FramesTotal != 1 || got.Throughput.PacketsReceived != 1 {
		t.Fatalf("unexpected throughput: %+v", got.Throughput)
	}
}

func TestHealthVerboseReportsStaleCapture(t *testing.T) {
	queue := fakeHealthQueue{
		state: receiver.State{
			SessionID:       "session-old",
			LastHeartbeatAt: time.Now().Add(-2 * time.Minute).UTC(),
		},
	}
	server := NewServerWithOptions(&fakeStore{}, queue, &fakeRules{}, stats.New(), Options{HealthFreshnessLimit: time.Second})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"status":"degraded"`, `"capture":{"status":"stale"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestHealthVerboseReportsDegradedStorage(t *testing.T) {
	store := &fakeStore{
		alerts: []*model.Alert{{RuleID: "rule-1"}},
		health: alert.StorageHealth{
			Status:      "degraded",
			LastError:   "disk full",
			LastErrorAt: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"status":"degraded"`, `"storage":{"status":"degraded"`, `"last_error":"disk full"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestReadOnlyEndpointsRejectNonGET(t *testing.T) {
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New())
	for _, path := range []string{"/api/health", "/api/metrics", "/api/alerts"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.Header.Set("X-Request-ID", "req-method")
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
				t.Fatalf("Allow = %q, want %q", allow, http.MethodGet)
			}
			body := rec.Body.String()
			for _, want := range []string{`"code":"METHOD_NOT_ALLOWED"`, `"request_id":"req-method"`} {
				if !strings.Contains(body, want) {
					t.Fatalf("response missing %q: %s", want, body)
				}
			}
		})
	}
}

func TestAlertsPaginationEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{alerts: []*model.Alert{
		{RuleID: "rule-1"},
		{RuleID: "rule-2"},
		{RuleID: "rule-3"},
	}}, fakeQueue{}, &fakeRules{}, stats.New())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts?page=2&per_page=2", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Data       []model.Alert `json:"data"`
		Pagination pagination    `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].RuleID != "rule-3" {
		t.Fatalf("unexpected page data: %+v", got.Data)
	}
	if got.Pagination.Page != 2 || got.Pagination.PerPage != 2 || got.Pagination.Total != 3 {
		t.Fatalf("unexpected pagination: %+v", got.Pagination)
	}
}

func TestAlertsInvalidPaginationUsesErrorEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts?page=0", nil)
	req.Header.Set("X-Request-ID", "req-test")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"code":"VALIDATION_ERROR"`, `"request_id":"req-test"`, "page must be a positive integer"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestAlertsFilters(t *testing.T) {
	recent := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	older := recent.Add(-2 * time.Hour)
	server := NewServer(&fakeStore{alerts: []*model.Alert{
		{
			RuleID:           "rule-1",
			Severity:         model.SeverityHigh,
			SrcIP:            "10.0.0.1",
			DstIP:            "10.0.0.2",
			DstPort:          80,
			Protocol:         "TCP",
			LastSeen:         recent,
			MitreTactic:      "Initial Access",
			MitreTechniqueID: "T1190",
			MatchedKeyword:   "UNION SELECT",
			AggregatedCount:  3,
		},
		{
			RuleID:           "rule-2",
			Severity:         model.SeverityLow,
			SrcIP:            "10.0.0.3",
			DstIP:            "10.0.0.4",
			DstPort:          53,
			Protocol:         "UDP",
			LastSeen:         older,
			MitreTactic:      "Discovery",
			MitreTechniqueID: "T1046",
			MatchedKeyword:   "scanner",
			AggregatedCount:  1,
		},
	}}, fakeQueue{}, &fakeRules{}, stats.New())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts?severity=high&src_ip=10.0.0.1&protocol=tcp&dst_port=80&since=2026-07-02T11:00:00Z&until=2026-07-02T13:00:00Z&mitre_tactic=initial+access&mitre_technique_id=t1190&matched_keyword=union&min_count=2", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Data       []model.Alert `json:"data"`
		Pagination pagination    `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].RuleID != "rule-1" {
		t.Fatalf("unexpected filtered alerts: %+v", got.Data)
	}
	if got.Pagination.Total != 1 {
		t.Fatalf("total = %d, want 1", got.Pagination.Total)
	}
}

func TestAlertsInvalidFilterUsesErrorEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts?dst_port=70000", nil)
	req.Header.Set("X-Request-ID", "req-filter")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"code":"VALIDATION_ERROR"`, `"request_id":"req-filter"`, "dst_port must be an integer from 0 to 65535"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestAlertsInvalidTimeRangeUsesErrorEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts?since=2026-07-02T13:00:00Z&until=2026-07-02T12:00:00Z", nil)
	req.Header.Set("X-Request-ID", "req-time")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"code":"VALIDATION_ERROR"`, `"request_id":"req-time"`, "until must be greater than or equal to since"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestSuppressionsList(t *testing.T) {
	manager, err := alert.NewSuppressionManager([]alert.Suppression{{ID: "existing", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/suppressions", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Data []alert.Suppression `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].ID != "existing" {
		t.Fatalf("unexpected suppressions: %+v", got.Data)
	}
}

func TestSuppressionsCreateRequiresBearerToken(t *testing.T) {
	manager, err := alert.NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager, AuthEnabled: true, AuthToken: "secret"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/suppressions", strings.NewReader(`{"id":"s1","enabled":true,"any_cidrs":["10.0.0.0/24"]}`))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestSuppressionsCreateAddsRule(t *testing.T) {
	manager, err := alert.NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager, AuthEnabled: true, AuthToken: "secret"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/suppressions", strings.NewReader(`{"id":"s1","enabled":true,"rule_ids":["rule-1"],"src_cidrs":["10.0.0.0/24"]}`))
	req.Header.Set("Authorization", "Bearer secret")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	listed := manager.List()
	if len(listed) != 1 || listed[0].ID != "s1" {
		t.Fatalf("unexpected suppressions: %+v", listed)
	}
}

func TestSuppressionsCreatePersistsRule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	manager, err := alert.NewSuppressionManagerWithFile(nil, path)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/suppressions", strings.NewReader(`{"id":"s1","enabled":true,"any_cidrs":["10.0.0.0/24"]}`))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	loaded, err := alert.LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "s1" {
		t.Fatalf("unexpected persisted suppressions: %+v", loaded)
	}
}

func TestSuppressionsCreateRejectsInvalidCIDR(t *testing.T) {
	manager, err := alert.NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/suppressions", strings.NewReader(`{"id":"bad","enabled":true,"any_cidrs":["not-a-cidr"]}`))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"VALIDATION_ERROR"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSuppressionsUpdatePersistsRule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	initial := []alert.Suppression{{ID: "s1", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}}
	if err := alert.SaveSuppressionsToFile(path, initial); err != nil {
		t.Fatalf("save suppressions: %v", err)
	}
	manager, err := alert.NewSuppressionManagerWithFile(initial, path)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/suppressions/s1", strings.NewReader(`{"enabled":true,"dst_cidrs":["192.0.2.0/24"]}`))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	loaded, err := alert.LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "s1" || len(loaded[0].DstCIDRs) != 1 {
		t.Fatalf("unexpected persisted suppressions: %+v", loaded)
	}
}

func TestSuppressionsDeleteRemovesRule(t *testing.T) {
	manager, err := alert.NewSuppressionManager([]alert.Suppression{{ID: "s1", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/suppressions/s1", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if listed := manager.List(); len(listed) != 0 {
		t.Fatalf("expected delete to clear suppressions, got %+v", listed)
	}
}

func TestSuppressionsDeleteMissingRuleUsesNotFoundEnvelope(t *testing.T) {
	manager, err := alert.NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/suppressions/missing", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"SUPPRESSION_NOT_FOUND"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSuppressionsReloadFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	initial := []alert.Suppression{{ID: "old", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}}
	if err := alert.SaveSuppressionsToFile(path, initial); err != nil {
		t.Fatalf("save initial suppressions: %v", err)
	}
	manager, err := alert.NewSuppressionManagerWithFile(initial, path)
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	next := []alert.Suppression{{ID: "new", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}}
	if err := alert.SaveSuppressionsToFile(path, next); err != nil {
		t.Fatalf("save next suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/suppressions/reload", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	listed := manager.List()
	if len(listed) != 1 || listed[0].ID != "new" {
		t.Fatalf("unexpected reloaded suppressions: %+v", listed)
	}
	if !strings.Contains(rec.Body.String(), `"reloaded":1`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRulesList(t *testing.T) {
	rules := &fakeRules{rules: []*model.Rule{{ID: "rule-2", Name: "Second"}, {ID: "rule-1", Name: "First"}}}
	server := NewServer(&fakeStore{}, fakeQueue{}, rules, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/rules", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Data []model.Rule `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Data) != 2 || got.Data[0].ID != "rule-2" || got.Data[1].ID != "rule-1" {
		t.Fatalf("unexpected rules: %+v", got.Data)
	}
}

func TestRulesListDoesNotRequireAuth(t *testing.T) {
	rules := &fakeRules{rules: []*model.Rule{{ID: "rule-1", Name: "First"}}}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, rules, stats.New(), Options{AuthEnabled: true, AuthToken: "secret"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/rules", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRuleMutationRequiresBearerToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{
		RulesSeedFile: path,
		AuthEnabled:   true,
		AuthToken:     "secret",
	})
	body := `{"id":"rule-new","name":"New Rule","type":"payload_match","severity":"high","enabled":true,"config":{"keywords":["needle"]}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(body))
	req.Header.Set("X-Request-ID", "req-auth")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	bodyText := rec.Body.String()
	for _, want := range []string{`"code":"UNAUTHORIZED"`, `"request_id":"req-auth"`} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("response missing %q: %s", want, bodyText)
		}
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != `Bearer realm="netsentry"` {
		t.Fatalf("WWW-Authenticate = %q", got)
	}
}

func TestRuleMutationAcceptsBearerToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	rules := &fakeRules{}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, rules, stats.New(), Options{
		RulesSeedFile: path,
		AuthEnabled:   true,
		AuthToken:     "secret",
	})
	body := `{"id":"rule-new","name":"New Rule","type":"payload_match","severity":"high","enabled":true,"config":{"keywords":["needle"]}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer secret")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(rules.reloaded) != 1 || rules.reloaded[0].ID != "rule-new" {
		t.Fatalf("unexpected reloaded rules: %+v", rules.reloaded)
	}
}

func TestRulesReloadFromSeedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	data := `{"rules":[{"id":"rule-reload","name":"Reloaded","type":"payload_match","severity":"high","enabled":true,"config":{"keywords":["test"]}}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	rules := &fakeRules{}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, rules, stats.New(), Options{RulesSeedFile: path})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(rules.reloaded) != 1 || rules.reloaded[0].ID != "rule-reload" {
		t.Fatalf("unexpected reloaded rules: %+v", rules.reloaded)
	}
	if !strings.Contains(rec.Body.String(), `"reloaded":1`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRulesReloadWithoutSeedFileUsesErrorEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil)
	req.Header.Set("X-Request-ID", "req-rules")
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"code":"RULES_RELOAD_UNAVAILABLE"`, `"request_id":"req-rules"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestRulesCreatePersistsAndReloads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	rules := &fakeRules{}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, rules, stats.New(), Options{RulesSeedFile: path})
	body := `{"id":"rule-new","name":"New Rule","type":"payload_match","severity":"high","enabled":true,"config":{"keywords":["needle"]}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(body))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(rules.reloaded) != 1 || rules.reloaded[0].ID != "rule-new" {
		t.Fatalf("unexpected reloaded rules: %+v", rules.reloaded)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read rules seed: %v", err)
	}
	if !strings.Contains(string(written), `"rules"`) || !strings.Contains(string(written), `"rule-new"`) {
		t.Fatalf("rules file was not updated: %s", string(written))
	}
}

func TestRulesUpdateRequiresMatchingID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	rules := &fakeRules{rules: []*model.Rule{{ID: "rule-1", Name: "Old", Type: model.RuleTypePayloadMatch, Severity: model.SeverityLow, Enabled: true, Config: json.RawMessage(`{"keywords":["old"]}`)}}}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, rules, stats.New(), Options{RulesSeedFile: path})
	body := `{"id":"other","name":"Updated","type":"payload_match","severity":"high","enabled":true,"config":{"keywords":["new"]}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/rules/rule-1", strings.NewReader(body))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Rule ID in path and body must match") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRulesDeletePersistsAndReloads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	rules := &fakeRules{rules: []*model.Rule{
		{ID: "rule-1", Name: "One", Type: model.RuleTypePayloadMatch, Severity: model.SeverityLow, Enabled: true, Config: json.RawMessage(`{"keywords":["one"]}`)},
		{ID: "rule-2", Name: "Two", Type: model.RuleTypePayloadMatch, Severity: model.SeverityHigh, Enabled: true, Config: json.RawMessage(`{"keywords":["two"]}`)},
	}}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, rules, stats.New(), Options{RulesSeedFile: path})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/rules/rule-1", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(rules.reloaded) != 1 || rules.reloaded[0].ID != "rule-2" {
		t.Fatalf("unexpected reloaded rules: %+v", rules.reloaded)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read rules seed: %v", err)
	}
	if strings.Contains(string(written), "rule-1") || !strings.Contains(string(written), "rule-2") {
		t.Fatalf("rules file was not updated: %s", string(written))
	}
}

func TestStoreErrorUsesErrorEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{err: errors.New("disk offline")}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestMetricsEndpoint(t *testing.T) {
	server := NewServer(&fakeStore{alerts: []*model.Alert{{RuleID: "rule-1"}}}, fakeQueue{depth: 7}, &fakeRules{count: 3}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"# HELP netsentry_alerts_current Current number of aggregated alerts in storage.",
		"# HELP netsentry_packet_queue_depth Current packet queue depth.",
		"# HELP netsentry_rules_loaded Current number of loaded rules.",
		"netsentry_alerts_current 1",
		"netsentry_packet_queue_depth 7",
		"netsentry_rules_loaded 3",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q:\n%s", want, body)
		}
	}
}

func TestMetricsEndpointIncludesCaptureHeartbeatGauges(t *testing.T) {
	queue := fakeHealthQueue{
		depth: 2,
		state: receiver.State{
			SessionID: "session-1",
			Heartbeat: receiver.HeartbeatFrame{
				SessionID:          "session-1",
				Sent:               10,
				Dropped:            1,
				ParseErrors:        2,
				AvgJSONSerializeUS: 2.5,
				UDSWriteErrors:     3,
			},
			LastHeartbeatAt: time.Now().Add(-time.Second).UTC(),
		},
	}
	server := NewServerWithOptions(&fakeStore{}, queue, &fakeRules{count: 4}, stats.New(), Options{HealthFreshnessLimit: time.Minute})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"# HELP netsentry_capture_connected Whether the capture heartbeat is currently fresh.",
		"netsentry_capture_connected 1",
		"netsentry_capture_packets_sent 10",
		"netsentry_capture_packets_dropped 1",
		"netsentry_capture_parse_errors 2",
		"netsentry_capture_uds_write_errors 3",
		"netsentry_capture_avg_json_serialize_seconds 2.5e-06",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q:\n%s", want, body)
		}
	}
}

func TestMetricsEndpointIncludesStorageAvailableGauge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netsentry.db")
	server := NewServer(&fakeStoreWithPath{path: path}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"# HELP netsentry_storage_available_bytes Available bytes on the alert storage filesystem.",
		"# TYPE netsentry_storage_available_bytes gauge",
		"netsentry_storage_available_bytes ",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q:\n%s", want, body)
		}
	}
}

func TestMetricsEndpointIncludesStorageHealthGauge(t *testing.T) {
	store := &fakeStore{health: alert.StorageHealth{Status: "degraded", LastError: "disk full"}}
	server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"# HELP netsentry_storage_healthy Whether alert storage is currently healthy.",
		"# TYPE netsentry_storage_healthy gauge",
		"netsentry_storage_healthy 0",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q:\n%s", want, body)
		}
	}
}

func TestAuditLogsMutationRequests(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{AuditLogger: zap.New(core)})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil)
	req.Header.Set("X-Request-ID", "req-audit")
	server.Handler().ServeHTTP(rec, req)

	entries := observed.FilterMessage("api audit").All()
	if len(entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["request_id"] != "req-audit" || fields["method"] != http.MethodPost || fields["path"] != "/api/rules/reload" {
		t.Fatalf("unexpected audit fields: %+v", fields)
	}
	if fields["status"] != int64(http.StatusConflict) || fields["target"] != "rules" {
		t.Fatalf("unexpected audit result fields: %+v", fields)
	}
}

func TestAuditSkipsGetRequests(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{AuditLogger: zap.New(core)})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	server.Handler().ServeHTTP(rec, req)

	if got := observed.FilterMessage("api audit").Len(); got != 0 {
		t.Fatalf("audit entries = %d, want 0", got)
	}
}

func TestAuditSharesGeneratedRequestIDWithErrorEnvelope(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{AuditLogger: zap.New(core)})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil)
	server.Handler().ServeHTTP(rec, req)

	entries := observed.FilterMessage("api audit").All()
	if len(entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(entries))
	}
	requestID, ok := entries[0].ContextMap()["request_id"].(string)
	if !ok || requestID == "" {
		t.Fatalf("missing audit request id: %+v", entries[0].ContextMap())
	}
	if !strings.Contains(rec.Body.String(), `"request_id":"`+requestID+`"`) {
		t.Fatalf("response and audit request ids differ: body=%s audit=%s", rec.Body.String(), requestID)
	}
}
