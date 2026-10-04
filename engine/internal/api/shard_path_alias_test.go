package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type shardAliasQueue struct{}

func (shardAliasQueue) QueueDepth() int { return 0 }

func TestDailyShardAliasHTTPListHealthAndMetricsUseActualCounts(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	dir := filepath.Join(t.TempDir(), "HTTP daily paths with spaces")
	active := filepath.Join(dir, "netsentry-2026-10-04.db")
	input := []*model.Alert{{RuleID: "active", RuleName: "Alias HTTP", Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.9", DstPort: 80, Protocol: "TCP", Timestamp: now}}
	seed, err := alert.Open(ctx, alert.Options{Path: active, JournalMode: "DELETE"})
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.WriteBatch(ctx, input); err != nil {
		_ = seed.Close()
		t.Fatal(err)
	}
	want, err := seed.List(ctx)
	if err != nil || len(want) != 1 {
		_ = seed.Close()
		t.Fatalf("seed rows=%d error=%v", len(want), err)
	}
	if err := seed.Close(); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relDir, err := filepath.Rel(cwd, dir)
	if err != nil {
		t.Fatal(err)
	}
	store, err := alert.Open(ctx, alert.Options{Path: active, Dir: relDir, DailyShard: true, JournalMode: "DELETE"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	})
	before, err := os.ReadFile(active)
	if err != nil {
		t.Fatal(err)
	}
	infoBefore, err := os.Stat(active)
	if err != nil {
		t.Fatal(err)
	}
	entriesBefore, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	server := api.NewServer(store, shardAliasQueue{}, rule.NewEngine(), stats.New()).Handler()
	get := func(target string) *httptest.ResponseRecorder {
		t.Helper()
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", target, response.Code, response.Body.String())
		}
		return response
	}
	for repeat := 0; repeat < 2; repeat++ {
		for _, page := range []string{"1", "2", "3"} {
			response := get("/api/alerts?per_page=1&page=" + page)
			var list struct {
				Data       []*model.Alert `json:"data"`
				Pagination struct {
					Page    int `json:"page"`
					PerPage int `json:"per_page"`
					Total   int `json:"total"`
				} `json:"pagination"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &list); err != nil {
				t.Fatal(err)
			}
			if list.Pagination.Total != 1 || list.Pagination.PerPage != 1 || list.Pagination.Page != int(page[0]-'0') {
				t.Fatalf("incorrect pagination: %+v", list.Pagination)
			}
			if page == "1" {
				if !reflect.DeepEqual(list.Data, want) {
					t.Fatalf("HTTP complete rows=%+v, want %+v", list.Data, want)
				}
			} else if len(list.Data) != 0 {
				t.Fatalf("alias appeared on page %s: %+v", page, list.Data)
			}
		}
		for _, target := range []string{"/api/health", "/api/health?verbose=true"} {
			var health struct {
				Status  string `json:"status"`
				Alerts  int    `json:"alerts"`
				Storage struct {
					Alerts int `json:"alerts"`
				} `json:"storage"`
			}
			if err := json.Unmarshal(get(target).Body.Bytes(), &health); err != nil {
				t.Fatal(err)
			}
			if health.Alerts != 1 || health.Status != "ok" || (strings.Contains(target, "verbose") && health.Storage.Alerts != 1) {
				t.Fatalf("health count: %+v", health)
			}
		}
		response := get("/api/metrics")
		if response.Header().Get("Content-Type") != "text/plain; version=0.0.4; charset=utf-8" || strings.Count(response.Body.String(), "\nnetsentry_alerts_current 1\n") != 1 {
			t.Fatalf("metrics count: %s", response.Body.String())
		}
	}
	after, err := os.ReadFile(active)
	if err != nil {
		t.Fatal(err)
	}
	infoAfter, err := os.Stat(active)
	if err != nil {
		t.Fatal(err)
	}
	entriesAfter, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	membership := func(entries []os.DirEntry) []string {
		out := make([]string, len(entries))
		for i, e := range entries {
			out[i] = e.Name()
		}
		return out
	}
	if string(after) != string(before) || infoAfter.Mode() != infoBefore.Mode() || !reflect.DeepEqual(membership(entriesAfter), membership(entriesBefore)) || store.Path() != active {
		t.Fatal("HTTP reads changed artifacts or original Path")
	}
}
