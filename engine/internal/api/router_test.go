package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/receiver"
	"github.com/decline-llc/netsentry/internal/rule"
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

type fakeRecoveryStore struct {
	fakeStore
	recoverErr error
	calls      int
}

func (s *fakeRecoveryStore) Recover(context.Context) error {
	s.calls++
	return s.recoverErr
}

type fakeQueryStore struct {
	fakeStore
	query       alert.Query
	queryAlerts []*model.Alert
	queryTotal  int
	queryErr    error
	called      bool
}

func (s *fakeQueryStore) Query(ctx context.Context, query alert.Query) ([]*model.Alert, int, error) {
	s.called = true
	s.query = query
	if s.queryErr != nil {
		return nil, 0, s.queryErr
	}
	return s.queryAlerts, s.queryTotal, nil
}

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

type faultSuppressions struct {
	err error
}

func (s *faultSuppressions) List() []alert.Suppression {
	return nil
}

func (s *faultSuppressions) Add(alert.Suppression) error {
	return s.err
}

func (s *faultSuppressions) Update(string, alert.Suppression) error {
	return s.err
}

func (s *faultSuppressions) Delete(string) error {
	return s.err
}

func (s *faultSuppressions) ReloadFromFile() error {
	return s.err
}

type committedSuppressionPersistenceError struct{}

func (committedSuppressionPersistenceError) Error() string {
	return "injected committed suppression persistence failure"
}

func (committedSuppressionPersistenceError) ReplacementCommitted() bool {
	return true
}

type synchronizedRules struct {
	mu                sync.RWMutex
	rules             []*model.Rule
	firstRulesEntered chan struct{}
	releaseFirstRules chan struct{}
	blockFirstRules   sync.Once
	rulesCalls        chan struct{}
	reloadCalls       chan struct{}
}

func newSynchronizedRules(rules []*model.Rule) *synchronizedRules {
	return &synchronizedRules{
		rules:             cloneRules(rules),
		firstRulesEntered: make(chan struct{}),
		releaseFirstRules: make(chan struct{}),
		rulesCalls:        make(chan struct{}, 8),
		reloadCalls:       make(chan struct{}, 8),
	}
}

func (r *synchronizedRules) RuleCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rules)
}

func (r *synchronizedRules) Rules() []*model.Rule {
	r.rulesCalls <- struct{}{}
	r.blockFirstRules.Do(func() {
		close(r.firstRulesEntered)
		<-r.releaseFirstRules
	})
	return r.snapshot()
}

func (r *synchronizedRules) Reload(rules []*model.Rule) error {
	r.mu.Lock()
	r.rules = cloneRules(rules)
	r.mu.Unlock()
	r.reloadCalls <- struct{}{}
	return nil
}

func (r *synchronizedRules) snapshot() []*model.Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneRules(r.rules)
}

type asyncHTTPResult struct {
	status int
	body   string
}

func serveAsync(handler http.Handler, req *http.Request) (<-chan struct{}, <-chan asyncHTTPResult) {
	started := make(chan struct{})
	done := make(chan asyncHTTPResult, 1)
	go func() {
		close(started)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		done <- asyncHTTPResult{status: rec.Code, body: rec.Body.String()}
	}()
	return started, done
}

func waitHTTPResult(t *testing.T, done <-chan asyncHTTPResult) asyncHTTPResult {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for rule-management request")
		return asyncHTTPResult{}
	}
}

func assertTransactionBlocked(t *testing.T, crossed <-chan struct{}) {
	t.Helper()
	select {
	case <-crossed:
		t.Fatal("concurrent rule transaction crossed the blocked transaction boundary")
	case <-time.After(200 * time.Millisecond):
	}
}

func testPayloadRule(id, name, keyword string) *model.Rule {
	return &model.Rule{
		ID:       id,
		Name:     name,
		Type:     model.RuleTypePayloadMatch,
		Severity: model.SeverityHigh,
		Enabled:  true,
		Config:   json.RawMessage(fmt.Sprintf(`{"keywords":[%q]}`, keyword)),
	}
}

func ruleRequestBody(t *testing.T, rule *model.Rule) string {
	t.Helper()
	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal rule request: %v", err)
	}
	return string(data)
}

func assertRuleState(t *testing.T, path string, manager *synchronizedRules, want map[string]string) {
	t.Helper()
	persisted, err := rule.LoadFromFile(path)
	if err != nil {
		t.Fatalf("load persisted rules: %v", err)
	}
	for label, rules := range map[string][]*model.Rule{
		"persisted": persisted,
		"active":    manager.snapshot(),
	} {
		if len(rules) != len(want) {
			t.Fatalf("%s rule count = %d, want %d: %+v", label, len(rules), len(want), rules)
		}
		for _, got := range rules {
			if got == nil || want[got.ID] != got.Name {
				t.Fatalf("unexpected %s rule: %+v, want %+v", label, got, want)
			}
		}
	}
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

func TestHealthVerboseReportsEmergencyStorage(t *testing.T) {
	store := &fakeStore{
		alerts: []*model.Alert{{RuleID: "rule-1"}},
		health: alert.StorageHealth{
			Status:      "emergency",
			LastError:   "database or disk is full",
			LastErrorAt: time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC),
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
	for _, want := range []string{`"status":"degraded"`, `"storage":{"status":"emergency"`, `"last_error":"database or disk is full"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestHealthVerboseIncludesStorageAvailableBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netsentry.db")
	server := NewServer(&fakeStoreWithPath{path: path}, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Storage struct {
			AvailableBytes uint64 `json:"available_bytes"`
		} `json:"storage"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Storage.AvailableBytes == 0 {
		t.Fatalf("expected storage available bytes, body = %s", rec.Body.String())
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

func TestStorageRecoveryRequiresPostAndConfiguredAuthentication(t *testing.T) {
	store := &fakeRecoveryStore{}
	server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/storage/recovery", nil)
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("GET recovery status=%d allow=%q body=%s", rec.Code, rec.Header().Get("Allow"), rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/storage/recovery", nil)
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"STORAGE_RECOVERY_AUTH_REQUIRED"`) {
		t.Fatalf("unauthenticated configuration status=%d body=%s", rec.Code, rec.Body.String())
	}
	if store.calls != 0 {
		t.Fatalf("recovery calls=%d, want 0 without configured auth", store.calls)
	}
}

func TestStorageRecoveryRequiresAndAcceptsBearerToken(t *testing.T) {
	store := &fakeRecoveryStore{}
	server := NewServerWithOptions(store, fakeQueue{}, &fakeRules{}, stats.New(), Options{
		AuthEnabled: true,
		AuthToken:   "secret",
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/storage/recovery", nil)
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || store.calls != 0 {
		t.Fatalf("missing token status=%d calls=%d body=%s", rec.Code, store.calls, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/storage/recovery", nil)
	req.Header.Set("Authorization", "Bearer secret")
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || store.calls != 1 {
		t.Fatalf("valid token status=%d calls=%d body=%s", rec.Code, store.calls, rec.Body.String())
	}
	if rec.Header().Get("X-NetSentry-Recovery-Phase") != "complete" || !strings.Contains(rec.Body.String(), `"phase":"complete"`) {
		t.Fatalf("unexpected recovery success response: headers=%v body=%s", rec.Header(), rec.Body.String())
	}
}

func TestStorageRecoveryMapsStableConflictAndFailurePhases(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		status   int
		code     string
		phase    string
		writable string
	}{
		{name: "not needed", err: alert.ErrStorageRecoveryNotNeeded, status: http.StatusConflict, code: "STORAGE_RECOVERY_NOT_NEEDED", phase: "not_needed"},
		{name: "in progress", err: alert.ErrStorageRecoveryInProgress, status: http.StatusConflict, code: "STORAGE_RECOVERY_IN_PROGRESS", phase: "in_progress"},
		{name: "preflight", err: &alert.RecoveryError{Phase: "preflight", Err: errors.New("private path detail")}, status: http.StatusServiceUnavailable, code: "STORAGE_RECOVERY_FAILED", phase: "preflight", writable: "writable_attempted=false"},
		{name: "replay", err: &alert.RecoveryError{Phase: "replay", WritableAttempted: true, Err: errors.New("private path detail")}, status: http.StatusServiceUnavailable, code: "STORAGE_RECOVERY_FAILED", phase: "replay", writable: "writable_attempted=true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeRecoveryStore{recoverErr: tt.err}
			server := NewServerWithOptions(store, fakeQueue{}, &fakeRules{}, stats.New(), Options{AuthEnabled: true, AuthToken: "secret"})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/storage/recovery", nil)
			req.Header.Set("Authorization", "Bearer secret")
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != tt.status || !strings.Contains(rec.Body.String(), `"code":"`+tt.code+`"`) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if got := rec.Header().Get("X-NetSentry-Recovery-Phase"); got != tt.phase {
				t.Fatalf("phase header=%q, want %q", got, tt.phase)
			}
			if strings.Contains(rec.Body.String(), "private path detail") {
				t.Fatalf("response leaked underlying recovery error: %s", rec.Body.String())
			}
			if tt.writable != "" && !strings.Contains(rec.Body.String(), tt.writable) {
				t.Fatalf("response missing %q: %s", tt.writable, rec.Body.String())
			}
		})
	}
}

func TestHealthVerboseReportsRecoveryWithoutWaitingForStoreCount(t *testing.T) {
	started := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	store := &fakeStore{
		err: errors.New("count must not run while recovering"),
		health: alert.StorageHealth{
			Status:            "recovering",
			RecoveryPhase:     "preflight",
			RecoveryStartedAt: started,
		},
	}
	server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health?verbose=true", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, want := range []string{`"status":"recovering"`, `"recovery_phase":"preflight"`, `"recovery_started_at":"` + started.Format(time.RFC3339Nano) + `"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("health missing %q: %s", want, rec.Body.String())
		}
	}
}

func TestMutationEndpointsRejectUnsupportedMethods(t *testing.T) {
	manager, err := alert.NewSuppressionManager([]alert.Suppression{{ID: "s1", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new suppressions: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{rules: []*model.Rule{{ID: "rule-1"}}}, stats.New(), Options{
		Suppressions: manager,
	})
	tests := []struct {
		name   string
		method string
		path   string
		allow  string
	}{
		{name: "suppressions collection", method: http.MethodPut, path: "/api/suppressions", allow: "GET, POST"},
		{name: "suppression resource", method: http.MethodPost, path: "/api/suppressions/s1", allow: "PUT, DELETE"},
		{name: "suppressions reload", method: http.MethodGet, path: "/api/suppressions/reload", allow: http.MethodPost},
		{name: "rules collection", method: http.MethodPut, path: "/api/rules", allow: "GET, POST"},
		{name: "rule resource", method: http.MethodPost, path: "/api/rules/rule-1", allow: "PUT, DELETE"},
		{name: "rules reload", method: http.MethodGet, path: "/api/rules/reload", allow: http.MethodPost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Request-ID", "req-method")
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			if allow := rec.Header().Get("Allow"); allow != tt.allow {
				t.Fatalf("Allow = %q, want %q", allow, tt.allow)
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

func TestAlertsUsesStoreQueryWhenAvailable(t *testing.T) {
	since := time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC)
	store := &fakeQueryStore{
		queryAlerts: []*model.Alert{{RuleID: "rule-1"}},
		queryTotal:  12,
	}
	server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/alerts?page=2&per_page=5&severity=high&since=2026-07-02T11:00:00Z&matched_keyword=union&min_count=2", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !store.called {
		t.Fatal("expected query-capable store to be used")
	}
	if store.query.Limit != 5 || store.query.Offset != 5 {
		t.Fatalf("query pagination = limit %d offset %d, want 5/5", store.query.Limit, store.query.Offset)
	}
	if store.query.Severity != model.SeverityHigh || store.query.MatchedKeyword != "union" {
		t.Fatalf("unexpected query filters: %+v", store.query)
	}
	if store.query.Since == nil || !store.query.Since.Equal(since) {
		t.Fatalf("query since = %v, want %s", store.query.Since, since)
	}
	if store.query.MinCount == nil || *store.query.MinCount != 2 {
		t.Fatalf("query min_count = %v, want 2", store.query.MinCount)
	}
	var got struct {
		Data       []model.Alert `json:"data"`
		Pagination pagination    `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Data) != 1 || got.Pagination.Total != 12 {
		t.Fatalf("unexpected query response: %+v", got)
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

func TestSuppressionMutationsReportCommittedDurabilityFailure(t *testing.T) {
	err := fmt.Errorf("persist suppressions: %w", committedSuppressionPersistenceError{})
	manager := &faultSuppressions{err: err}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create", method: http.MethodPost, path: "/api/suppressions", body: `{"id":"new","enabled":true,"any_cidrs":["192.0.2.0/24"]}`},
		{name: "update", method: http.MethodPut, path: "/api/suppressions/existing", body: `{"enabled":true,"any_cidrs":["192.0.2.0/24"]}`},
		{name: "delete", method: http.MethodDelete, path: "/api/suppressions/existing"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			server.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"SUPPRESSIONS_DURABILITY_UNCERTAIN"`) {
				t.Fatalf("unexpected body: %s", rec.Body.String())
			}
		})
	}
}

func TestSuppressionMutationPreRenamePersistenceFailureRemainsInternalError(t *testing.T) {
	manager := &faultSuppressions{err: errors.New("persist suppressions: injected pre-rename failure")}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/suppressions", strings.NewReader(`{"id":"new","enabled":true,"any_cidrs":["192.0.2.0/24"]}`))
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
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

func TestRuleMutationRejectsOversizedBody(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{RulesSeedFile: path})
	body := `{"id":"rule-new","name":"` + strings.Repeat("x", maxMutationBodyBytes) + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(body))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"REQUEST_TOO_LARGE"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRuleMutationRejectsTrailingJSONDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{RulesSeedFile: path})
	body := `{"id":"rule-new","name":"New","type":"payload_match","severity":"high","enabled":true,"config":{"keywords":["needle"]}} {}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(body))
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "exactly one JSON document") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
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

func TestConcurrentRuleCreatesSerializeCompleteTransactions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := rule.SaveToFile(path, nil); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	manager := newSynchronizedRules(nil)
	handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), Options{RulesSeedFile: path}).Handler()

	firstRule := testPayloadRule("rule-first", "First", "first")
	_, firstDone := serveAsync(handler, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, firstRule))))
	<-manager.firstRulesEntered
	<-manager.rulesCalls

	secondRule := testPayloadRule("rule-second", "Second", "second")
	secondStarted, secondDone := serveAsync(handler, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, secondRule))))
	<-secondStarted
	assertTransactionBlocked(t, manager.rulesCalls)
	close(manager.releaseFirstRules)

	for label, result := range map[string]asyncHTTPResult{
		"first":  waitHTTPResult(t, firstDone),
		"second": waitHTTPResult(t, secondDone),
	} {
		if result.status != http.StatusCreated {
			t.Fatalf("%s create status = %d, body = %s", label, result.status, result.body)
		}
	}
	assertRuleState(t, path, manager, map[string]string{
		"rule-first":  "First",
		"rule-second": "Second",
	})
}

func TestConcurrentRuleUpdateDeleteSerializeCompleteTransactions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	initial := []*model.Rule{
		testPayloadRule("rule-update", "Before", "before"),
		testPayloadRule("rule-delete", "Delete", "delete"),
	}
	if err := rule.SaveToFile(path, initial); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	manager := newSynchronizedRules(initial)
	handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), Options{RulesSeedFile: path}).Handler()

	updated := testPayloadRule("rule-update", "After", "after")
	_, updateDone := serveAsync(handler, httptest.NewRequest(http.MethodPut, "/api/rules/rule-update", strings.NewReader(ruleRequestBody(t, updated))))
	<-manager.firstRulesEntered
	<-manager.rulesCalls

	deleteStarted, deleteDone := serveAsync(handler, httptest.NewRequest(http.MethodDelete, "/api/rules/rule-delete", nil))
	<-deleteStarted
	assertTransactionBlocked(t, manager.rulesCalls)
	close(manager.releaseFirstRules)

	updateResult := waitHTTPResult(t, updateDone)
	if updateResult.status != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updateResult.status, updateResult.body)
	}
	deleteResult := waitHTTPResult(t, deleteDone)
	if deleteResult.status != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", deleteResult.status, deleteResult.body)
	}
	assertRuleState(t, path, manager, map[string]string{"rule-update": "After"})
}

func TestConcurrentRuleMutationReloadSerializeCompleteTransactions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	initial := []*model.Rule{testPayloadRule("rule-existing", "Existing", "existing")}
	if err := rule.SaveToFile(path, initial); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	manager := newSynchronizedRules(initial)
	handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), Options{RulesSeedFile: path}).Handler()

	created := testPayloadRule("rule-created", "Created", "created")
	_, createDone := serveAsync(handler, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, created))))
	<-manager.firstRulesEntered
	<-manager.rulesCalls

	reloadStarted, reloadDone := serveAsync(handler, httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil))
	<-reloadStarted
	assertTransactionBlocked(t, manager.reloadCalls)
	close(manager.releaseFirstRules)

	createResult := waitHTTPResult(t, createDone)
	if createResult.status != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createResult.status, createResult.body)
	}
	reloadResult := waitHTTPResult(t, reloadDone)
	if reloadResult.status != http.StatusOK {
		t.Fatalf("reload status = %d, body = %s", reloadResult.status, reloadResult.body)
	}
	assertRuleState(t, path, manager, map[string]string{
		"rule-existing": "Existing",
		"rule-created":  "Created",
	})
}

func TestRejectedRuleTransactionsPreserveStateAndReleaseSerialization(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "rules.json")
		initial := []*model.Rule{testPayloadRule("rule-existing", "Existing", "existing")}
		if err := rule.SaveToFile(path, initial); err != nil {
			t.Fatalf("write rules seed: %v", err)
		}
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read rules seed: %v", err)
		}
		manager := &fakeRules{rules: cloneRules(initial)}
		handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), Options{RulesSeedFile: path}).Handler()

		invalid := testPayloadRule("rule-invalid", "Invalid", "")
		invalid.Config = json.RawMessage(`{"keywords":[]}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, invalid))))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid create status = %d, body = %s", rec.Code, rec.Body.String())
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read rules seed after rejection: %v", err)
		}
		if string(after) != string(before) || len(manager.rules) != 1 || manager.rules[0].ID != "rule-existing" {
			t.Fatalf("validation rejection changed file or active state: file=%s rules=%+v", after, manager.rules)
		}

		valid := testPayloadRule("rule-valid", "Valid", "valid")
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, valid))))
		if rec.Code != http.StatusCreated || len(manager.rules) != 2 {
			t.Fatalf("valid create after rejection status = %d, body = %s, rules = %+v", rec.Code, rec.Body.String(), manager.rules)
		}
	})

	t.Run("persistence", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "rules.json")
		initial := []*model.Rule{testPayloadRule("rule-existing", "Existing", "existing")}
		if err := rule.SaveToFile(path, initial); err != nil {
			t.Fatalf("write rules seed: %v", err)
		}
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read rules seed: %v", err)
		}
		manager := &fakeRules{rules: cloneRules(initial)}
		handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), Options{RulesSeedFile: path}).Handler()
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatalf("make seed directory read-only: %v", err)
		}
		defer func() { _ = os.Chmod(dir, 0o700) }()

		created := testPayloadRule("rule-created", "Created", "created")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, created))))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("unwritable create status = %d, body = %s", rec.Code, rec.Body.String())
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read rules seed after persistence failure: %v", err)
		}
		if string(after) != string(before) || len(manager.rules) != 1 || manager.rules[0].ID != "rule-existing" {
			t.Fatalf("persistence failure changed file or active state: file=%s rules=%+v", after, manager.rules)
		}

		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("restore seed directory permissions: %v", err)
		}
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, created))))
		if rec.Code != http.StatusCreated || len(manager.rules) != 2 {
			t.Fatalf("valid create after persistence failure status = %d, body = %s, rules = %+v", rec.Code, rec.Body.String(), manager.rules)
		}
	})
}

type committedRuleSaveError struct {
	err error
}

func (e *committedRuleSaveError) Error() string {
	return "rules file replacement committed but sync parent directory failed: " + e.err.Error()
}

func (e *committedRuleSaveError) Unwrap() error {
	return e.err
}

func (e *committedRuleSaveError) ReplacementCommitted() bool {
	return true
}

func TestRuleMutationPostRenameDurabilityFailurePublishesCommittedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	initial := []*model.Rule{testPayloadRule("rule-existing", "Existing", "existing")}
	if err := rule.SaveToFile(path, initial); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	manager := &fakeRules{rules: cloneRules(initial), count: len(initial)}
	opts := Options{RulesSeedFile: path}
	opts.saveRules = func(path string, rules []*model.Rule) error {
		if err := rule.SaveToFile(path, rules); err != nil {
			return err
		}
		return &committedRuleSaveError{err: errors.New("injected directory sync failure")}
	}
	handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), opts).Handler()

	created := testPayloadRule("rule-created", "Created", "created")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, created))))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), `"code":"RULES_DURABILITY_UNCERTAIN"`) {
		t.Fatalf("post-rename failure status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "applied but crash durability could not be confirmed") {
		t.Fatalf("post-rename response does not report committed outcome: %s", rec.Body.String())
	}
	assertFakeRuleState(t, path, manager, map[string]string{
		"rule-existing": "Existing",
		"rule-created":  "Created",
	})
}

func TestRuleMutationPreRenameFailurePreservesStateAndAllowsRetry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	initial := []*model.Rule{testPayloadRule("rule-existing", "Existing", "existing")}
	if err := rule.SaveToFile(path, initial); err != nil {
		t.Fatalf("write rules seed: %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read rules seed: %v", err)
	}
	manager := &fakeRules{rules: cloneRules(initial), count: len(initial)}
	failNext := true
	opts := Options{RulesSeedFile: path}
	opts.saveRules = func(path string, rules []*model.Rule) error {
		if failNext {
			failNext = false
			return errors.New("injected pre-rename failure")
		}
		return rule.SaveToFile(path, rules)
	}
	handler := NewServerWithOptions(&fakeStore{}, fakeQueue{}, manager, stats.New(), opts).Handler()
	created := testPayloadRule("rule-created", "Created", "created")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, created))))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("pre-rename failure status = %d, body = %s", rec.Code, rec.Body.String())
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read rules seed after rejection: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("pre-rename failure changed canonical bytes: got %q want %q", after, before)
	}
	assertFakeRuleState(t, path, manager, map[string]string{"rule-existing": "Existing"})

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/rules", strings.NewReader(ruleRequestBody(t, created))))
	if rec.Code != http.StatusCreated {
		t.Fatalf("retry status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertFakeRuleState(t, path, manager, map[string]string{
		"rule-existing": "Existing",
		"rule-created":  "Created",
	})
}

func assertFakeRuleState(t *testing.T, path string, manager *fakeRules, want map[string]string) {
	t.Helper()
	persisted, err := rule.LoadFromFile(path)
	if err != nil {
		t.Fatalf("load persisted rules: %v", err)
	}
	for label, rules := range map[string][]*model.Rule{
		"persisted": persisted,
		"active":    manager.rules,
	} {
		if len(rules) != len(want) {
			t.Fatalf("%s rule count = %d, want %d: %+v", label, len(rules), len(want), rules)
		}
		for _, got := range rules {
			if got == nil || want[got.ID] != got.Name {
				t.Fatalf("unexpected %s rule: %+v, want %+v", label, got, want)
			}
		}
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
	queue := &fakeQueue{depth: 7}
	metrics := stats.New()
	metrics.IncPacketProcessed()
	metrics.IncPacketProcessed()
	metrics.ObserveAlerts([]*model.Alert{{RuleID: "rule-1", Severity: model.SeverityHigh}})
	server := NewServer(&fakeStore{alerts: []*model.Alert{{RuleID: "rule-1"}}}, queue, &fakeRules{count: 3}, metrics)
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
		"# HELP netsentry_packet_queue_depth_high_water Highest packet queue depth observed by metrics scrapes.",
		"# HELP netsentry_packets_processed_per_second Process-lifetime average packets processed per second.",
		"# HELP netsentry_alerts_generated_per_second Process-lifetime average alerts generated per second.",
		"# HELP netsentry_process_uptime_seconds Seconds since this NetSentry engine process started.",
		"# HELP netsentry_rules_loaded Current number of loaded rules.",
		"netsentry_alerts_current 1",
		"netsentry_packet_queue_depth 7",
		"netsentry_packet_queue_depth_high_water 7",
		"netsentry_packets_processed_per_second ",
		"netsentry_alerts_generated_per_second ",
		"netsentry_process_uptime_seconds ",
		"netsentry_rules_loaded 3",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q:\n%s", want, body)
		}
	}

	queue.depth = 2
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)
	body = rec.Body.String()
	for _, want := range []string{
		"netsentry_packet_queue_depth 2",
		"netsentry_packet_queue_depth_high_water 7",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q after lower queue depth:\n%s", want, body)
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

func TestHealthAndMetricsCountDailyShardStore(t *testing.T) {
	dir := t.TempDir()
	days := []time.Time{
		time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
	}
	for _, day := range days {
		writeDailyShardAPIAlert(t, dir, day)
	}
	store := openDailyShardAPIStore(t, dir, days[len(days)-1])
	defer store.Close()

	server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var health struct {
		Alerts int `json:"alerts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if health.Alerts != len(days) {
		t.Fatalf("health alerts = %d, want %d", health.Alerts, len(days))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "netsentry_alerts_current 3") {
		t.Fatalf("metrics missing daily shard alert count:\n%s", rec.Body.String())
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

func TestAuditLogsStorageRecoveryPhaseWithoutPrivateError(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	store := &fakeRecoveryStore{recoverErr: &alert.RecoveryError{
		Phase: "preflight",
		Err:   errors.New("private/operator/path.db"),
	}}
	server := NewServerWithOptions(store, fakeQueue{}, &fakeRules{}, stats.New(), Options{
		AuthEnabled: true,
		AuthToken:   "secret",
		AuditLogger: zap.New(core),
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/storage/recovery", nil)
	req.Header.Set("Authorization", "Bearer secret")
	server.Handler().ServeHTTP(rec, req)

	entries := observed.FilterMessage("api audit").All()
	if len(entries) != 1 {
		t.Fatalf("audit entries=%d, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["target"] != "storage" || fields["phase"] != "preflight" || fields["authorized"] != true {
		t.Fatalf("unexpected storage audit fields: %+v", fields)
	}
	if strings.Contains(entries[0].Message+fmt.Sprint(fields), "private/operator/path.db") || strings.Contains(rec.Body.String(), "private/operator/path.db") {
		t.Fatalf("recovery audit or response leaked private error: fields=%+v body=%s", fields, rec.Body.String())
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

func openDailyShardAPIStore(t *testing.T, dir string, now time.Time) *alert.Store {
	t.Helper()
	store, err := alert.Open(context.Background(), alert.Options{
		Dir:               dir,
		DailyShard:        true,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("open daily shard alert store: %v", err)
	}
	return store
}

func writeDailyShardAPIAlert(t *testing.T, dir string, ts time.Time) {
	t.Helper()
	store := openDailyShardAPIStore(t, dir, ts)
	defer store.Close()
	alert := &model.Alert{
		RuleID:             "rule-" + ts.Format("20060102"),
		RuleName:           "Daily Shard Test",
		Timestamp:          ts,
		SrcIP:              "10.0.0.1",
		DstIP:              "10.0.0.2",
		DstPort:            80,
		Protocol:           "TCP",
		Severity:           model.SeverityHigh,
		MitreTactic:        "Initial Access",
		MitreTechniqueID:   "T1190",
		MitreTechniqueName: "Exploit Public-Facing Application",
		PayloadPreview:     "GET / HTTP/1.1",
		MatchedKeyword:     "daily-shard",
	}
	if err := store.WriteBatch(context.Background(), []*model.Alert{alert}); err != nil {
		t.Fatalf("write daily shard alert: %v", err)
	}
}
