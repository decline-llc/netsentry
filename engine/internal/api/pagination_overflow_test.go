package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestParsePaginationOffsetBoundaries(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	type testCase struct {
		name, query, diagnostic string
		page, size              int
	}
	cases := []testCase{
		{name: "defaults", page: 1, size: 20},
		{name: "ordinary", query: "page=2&per_page=5", page: 2, size: 5},
		{name: "empty_defaults", query: "page=&per_page=", page: 1, size: 20},
		{name: "zero_page", query: "page=0", diagnostic: "page must be a positive integer"},
		{name: "negative_page", query: "page=-1", diagnostic: "page must be a positive integer"},
		{name: "nonnumeric_page", query: "page=no", diagnostic: "page must be a positive integer"},
		{name: "page_outside_int", query: "page=" + strconv.FormatUint(uint64(maxInt)+1, 10), diagnostic: "page must be a positive integer"},
		{name: "zero_size", query: "per_page=0", diagnostic: "per_page must be a positive integer"},
		{name: "negative_size", query: "per_page=-1", diagnostic: "per_page must be a positive integer"},
		{name: "nonnumeric_size", query: "per_page=no", diagnostic: "per_page must be a positive integer"},
		{name: "oversize_before_offset", query: fmt.Sprintf("page=%d&per_page=101", maxInt), diagnostic: "per_page must be <= 100"},
		{name: "default_size_overflow", query: "page=" + strconv.Itoa(maxInt), diagnostic: "page and per_page exceed maximum pagination offset"},
	}
	for _, size := range []int{1, 20, 100} {
		// At size 1 the page itself reaches MaxInt before its offset can overflow.
		page := maxInt
		if size > 1 {
			page = maxInt/size + 1
		}
		cases = append(cases, testCase{name: fmt.Sprintf("last_valid_%d", size), query: fmt.Sprintf("page=%d&per_page=%d", page, size), page: page, size: size})
		if size > 1 {
			cases = append(cases, testCase{name: fmt.Sprintf("first_overflow_%d", size), query: fmt.Sprintf("page=%d&per_page=%d", page+1, size), diagnostic: "page and per_page exceed maximum pagination offset"})
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := parsePagination(httptest.NewRequest(http.MethodGet, "/api/alerts?"+tc.query, nil))
			if tc.diagnostic != "" {
				if err == nil || err.Error() != tc.diagnostic || p != (pagination{}) {
					t.Fatalf("pagination=%+v error=%v, want zero pagination and %q", p, err, tc.diagnostic)
				}
				return
			}
			if err != nil || p != (pagination{Page: tc.page, PerPage: tc.size}) {
				t.Fatalf("pagination=%+v error=%v, want page=%d size=%d", p, err, tc.page, tc.size)
			}
		})
	}
}

func TestPageBoundsClampsBeforeAdding(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name                          string
		total, page, size, start, end int
	}{
		{"empty", 0, 1, 20, 0, 0},
		{"first", 5, 1, 2, 0, 2},
		{"middle", 5, 2, 2, 2, 4},
		{"last", 5, 3, 2, 4, 5},
		{"exact_end", 4, 3, 2, 4, 4},
		{"past_end", 5, 4, 2, 5, 5},
		{"large_valid_empty", 5, maxInt/100 + 1, 100, 5, 5},
		{"max_total_last_one", maxInt, maxInt, 1, maxInt - 1, maxInt},
		{"max_total_clamp_before_add", maxInt, maxInt/100 + 1, 100, (maxInt / 100) * 100, maxInt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end := pageBounds(tc.total, pagination{Page: tc.page, PerPage: tc.size})
			if start != tc.start || end != tc.end || start < 0 || start > end || end > tc.total {
				t.Fatalf("bounds=(%d,%d), want (%d,%d), total=%d", start, end, tc.start, tc.end, tc.total)
			}
		})
	}
}

type paginationProbeStore struct {
	alerts                []*model.Alert
	listCalls, countCalls int
}

func (s *paginationProbeStore) List(context.Context) ([]*model.Alert, error) {
	s.listCalls++
	return s.alerts, nil
}

func (s *paginationProbeStore) Count(context.Context) (int, error) {
	s.countCalls++
	return len(s.alerts), nil
}

type paginationQueryProbe struct {
	*paginationProbeStore
	queryCalls int
	query      alert.Query
	results    []*model.Alert
}

func (s *paginationQueryProbe) Query(_ context.Context, query alert.Query) ([]*model.Alert, int, error) {
	s.queryCalls++
	s.query = query
	return s.results, 2, nil
}

func TestAlertsRejectOverflowBeforeStore(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, query := range []string{
		"page=" + strconv.Itoa(maxInt),
		fmt.Sprintf("page=%d&per_page=100", maxInt/100+2),
	} {
		for _, backend := range []string{"list", "query"} {
			t.Run(backend+"/"+query, func(t *testing.T) {
				probe := &paginationProbeStore{}
				queryProbe := &paginationQueryProbe{paginationProbeStore: probe}
				var store AlertStore = probe
				if backend == "query" {
					store = queryProbe
				}
				server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
				req := httptest.NewRequest(http.MethodGet, "/api/alerts?"+query, nil)
				req.Header.Set("X-Request-ID", "req-offset")
				rec := httptest.NewRecorder()
				server.Handler().ServeHTTP(rec, req)
				if rec.Code != http.StatusBadRequest {
					t.Fatalf("status=%d body=%s, want 400", rec.Code, rec.Body.String())
				}
				for _, want := range []string{`"code":"VALIDATION_ERROR"`, `"request_id":"req-offset"`, "page and per_page exceed maximum pagination offset"} {
					if !strings.Contains(rec.Body.String(), want) {
						t.Fatalf("response missing %q: %s", want, rec.Body.String())
					}
				}
				if probe.listCalls != 0 || probe.countCalls != 0 || queryProbe.queryCalls != 0 {
					t.Fatalf("invalid pagination reached store: list=%d count=%d query=%d", probe.listCalls, probe.countCalls, queryProbe.queryCalls)
				}
			})
		}
	}
}

func TestAlertsPaginationKeepsRepresentableOffsets(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	high1 := &model.Alert{RuleID: "high-1", Severity: model.SeverityHigh}
	high2 := &model.Alert{RuleID: "high-2", Severity: model.SeverityHigh}
	for _, tc := range []struct {
		name       string
		page, size int
		want       []*model.Alert
	}{
		{"filtered_second", 2, 1, []*model.Alert{high2}},
		{"max_page_size_one", maxInt, 1, []*model.Alert{}},
		{"max_offset_size_twenty", maxInt/20 + 1, 20, []*model.Alert{}},
		{"max_offset_size_hundred", maxInt/100 + 1, 100, []*model.Alert{}},
	} {
		for _, backend := range []string{"list", "query"} {
			t.Run(backend+"/"+tc.name, func(t *testing.T) {
				probe := &paginationProbeStore{alerts: []*model.Alert{high1, {RuleID: "low", Severity: model.SeverityLow}, high2}}
				queryProbe := &paginationQueryProbe{paginationProbeStore: probe, results: tc.want}
				var store AlertStore = probe
				if backend == "query" {
					store = queryProbe
				}
				server := NewServer(store, fakeQueue{}, &fakeRules{}, stats.New())
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/alerts?page=%d&per_page=%d&severity=high", tc.page, tc.size), nil)
				server.Handler().ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Fatalf("status=%d body=%s, want 200", rec.Code, rec.Body.String())
				}
				var got alertListResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Pagination != (pagination{Page: tc.page, PerPage: tc.size, Total: 2}) || len(got.Data) != len(tc.want) || got.Data == nil {
					t.Fatalf("unexpected envelope: %+v", got)
				}
				for i, want := range tc.want {
					if got.Data[i] == nil || *got.Data[i] != *want {
						t.Fatalf("alert[%d]=%+v, want %+v", i, got.Data[i], want)
					}
				}
				if probe.countCalls != 0 {
					t.Fatalf("unexpected Count calls: %d", probe.countCalls)
				}
				if backend == "query" {
					if probe.listCalls != 0 || queryProbe.queryCalls != 1 || queryProbe.query.Offset != (tc.page-1)*tc.size || queryProbe.query.Offset < 0 || queryProbe.query.Limit != tc.size || queryProbe.query.Severity != model.SeverityHigh {
						t.Fatalf("unexpected query/calls: %+v list=%d query=%d", queryProbe.query, probe.listCalls, queryProbe.queryCalls)
					}
				} else if probe.listCalls != 1 || queryProbe.queryCalls != 0 {
					t.Fatalf("unexpected fallback calls: list=%d query=%d", probe.listCalls, queryProbe.queryCalls)
				}
			})
		}
	}
}
