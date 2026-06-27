package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type fakeStore struct {
	alerts []*model.Alert
	err    error
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

type fakeQueue struct{ depth int }

func (q fakeQueue) QueueDepth() int { return q.depth }

type fakeRules struct{ count int }

func (r fakeRules) RuleCount() int { return r.count }

func TestAlertsPaginationEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{alerts: []*model.Alert{
		{RuleID: "rule-1"},
		{RuleID: "rule-2"},
		{RuleID: "rule-3"},
	}}, fakeQueue{}, fakeRules{}, stats.New())

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
	server := NewServer(&fakeStore{}, fakeQueue{}, fakeRules{}, stats.New())
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

func TestStoreErrorUsesErrorEnvelope(t *testing.T) {
	server := NewServer(&fakeStore{err: errors.New("disk offline")}, fakeQueue{}, fakeRules{}, stats.New())
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
	server := NewServer(&fakeStore{alerts: []*model.Alert{{RuleID: "rule-1"}}}, fakeQueue{depth: 7}, fakeRules{count: 3}, stats.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"netsentry_alerts_current 1", "netsentry_packet_queue_depth 7", "netsentry_rules_loaded 3"} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics missing %q:\n%s", want, body)
		}
	}
}
