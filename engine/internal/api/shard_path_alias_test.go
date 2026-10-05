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

type shardHTTPAliasArtifact struct {
	Mode os.FileMode
	Data string
}

func shardHTTPAliasTree(t *testing.T, dir string) map[string]shardHTTPAliasArtifact {
	t.Helper()
	result := map[string]shardHTTPAliasArtifact{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		item := shardHTTPAliasArtifact{Mode: info.Mode()}
		if !entry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			item.Data = string(data)
		}
		result[rel] = item
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// Keep the handoff's declaration name while covering actual Dir/Now resolution.
// Options.Path cannot independently configure a daily active pathname alias.
func TestDailyShardAliasHTTPListHealthAndMetricsUseActualCounts(t *testing.T) {
	for _, spelling := range []string{"absolute", "relative", "relative_dot", "absolute_dot", "absolute_parent"} {
		t.Run(spelling, func(t *testing.T) {
			ctx := context.Background()
			now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
			dir := filepath.Join(t.TempDir(), "HTTP daily paths with spaces")
			active := filepath.Join(dir, "netsentry-2026-10-04.db")
			history := filepath.Join(dir, "netsentry-2026-10-03.db")
			input := []*model.Alert{
				{RuleID: "latest", RuleName: "Daily HTTP", Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.9", DstPort: 80, Protocol: "TCP", Timestamp: now},
				{RuleID: "middle", RuleName: "Daily HTTP", Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.9", DstPort: 80, Protocol: "TCP", Timestamp: now.Add(-10 * time.Minute)},
				{RuleID: "history", RuleName: "Daily HTTP", Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.9", DstPort: 80, Protocol: "TCP", Timestamp: now.Add(-24 * time.Hour)},
			}
			original := []model.Alert{*input[0], *input[1], *input[2]}
			var want []*model.Alert
			// Seed exact absolute resources through primary mode, independently
			// of the daily path selection and discovery being exercised.
			for _, fixture := range []struct {
				path  string
				input []*model.Alert
			}{{active, input[:2]}, {history, input[2:]}} {
				seed, err := alert.Open(ctx, alert.Options{Path: fixture.path, JournalMode: "DELETE", Now: func() time.Time { return now }})
				if err != nil {
					t.Fatal(err)
				}
				if err := seed.WriteBatch(ctx, fixture.input); err != nil {
					_ = seed.Close()
					t.Fatal(err)
				}
				rows, err := seed.List(ctx)
				if err != nil || len(rows) != len(fixture.input) {
					_ = seed.Close()
					t.Fatalf("seed rows=%d error=%v", len(rows), err)
				}
				want = append(want, rows...)
				if err := seed.Close(); err != nil {
					t.Fatal(err)
				}
			}
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			relDir, err := filepath.Rel(cwd, dir)
			if err != nil {
				t.Fatal(err)
			}
			publicDir := dir
			switch spelling {
			case "relative":
				publicDir = relDir
			case "relative_dot":
				publicDir = "./" + relDir + "/."
			case "absolute_dot":
				publicDir = dir + "/."
			case "absolute_parent":
				publicDir = dir + "/../" + filepath.Base(dir)
			}
			ignored := filepath.Join(filepath.Dir(dir), "ignored explicit Path", "operator.db")
			expectedPath := filepath.Join(publicDir, "netsentry-2026-10-04.db")
			store, err := alert.Open(ctx, alert.Options{Path: ignored, Dir: publicDir, DailyShard: true, JournalMode: "DELETE", Now: func() time.Time { return now }})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("close: %v", err)
				}
			})
			assertPath := func() {
				t.Helper()
				absolute, err := filepath.Abs(store.Path())
				if err != nil || store.Path() != expectedPath || absolute != active || store.Path() == ignored {
					t.Fatalf("daily Path=%q absolute=%q error=%v, want derived %q and seeded %q", store.Path(), absolute, err, expectedPath, active)
				}
				if _, err := os.Stat(ignored); !os.IsNotExist(err) {
					t.Fatalf("ignored explicit Path acquired an artifact: %v", err)
				}
			}
			assertPath()
			before := shardHTTPAliasTree(t, filepath.Dir(dir))
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
			pages := []struct {
				query string
				page  int
				total int
				rows  []*model.Alert
			}{
				{"page=1", 1, 3, want[:1]},
				{"page=2", 2, 3, want[1:2]},
				{"page=3", 3, 3, want[2:]},
				{"page=4", 4, 3, nil},
				{"page=5", 5, 3, nil},
				{"rule_id=history&page=1", 1, 1, want[2:]},
				{"rule_id=history&page=2", 2, 1, nil},
			}
			for repeat := 0; repeat < 2; repeat++ {
				for _, page := range pages {
					response := get("/api/alerts?per_page=1&" + page.query)
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
					if list.Pagination.Total != page.total || list.Pagination.PerPage != 1 || list.Pagination.Page != page.page || len(list.Data) != len(page.rows) {
						t.Fatalf("%s incorrect pagination/rows: %+v", page.query, list)
					}
					for i, row := range page.rows {
						if list.Data[i] == nil || *list.Data[i] != *row {
							t.Fatalf("%s complete row[%d]=%+v, want %+v", page.query, i, list.Data[i], row)
						}
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
					if health.Alerts != 3 || health.Status != "ok" || (strings.Contains(target, "verbose") && health.Storage.Alerts != 3) {
						t.Fatalf("health count: %+v", health)
					}
				}
				response := get("/api/metrics")
				if response.Header().Get("Content-Type") != "text/plain; version=0.0.4; charset=utf-8" || strings.Count(response.Body.String(), "\nnetsentry_alerts_current 3\n") != 1 {
					t.Fatalf("metrics count: %s", response.Body.String())
				}
			}
			if !reflect.DeepEqual(shardHTTPAliasTree(t, filepath.Dir(dir)), before) {
				t.Fatal("HTTP reads changed active/history bytes/modes/membership")
			}
			for i := range input {
				if *input[i] != original[i] {
					t.Fatal("HTTP reads changed caller alert")
				}
			}
			assertPath()
		})
	}
}
