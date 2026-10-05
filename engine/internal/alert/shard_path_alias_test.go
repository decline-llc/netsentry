package alert_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

func shardAliasAlert(at time.Time, id string) *model.Alert {
	return &model.Alert{RuleID: id, RuleName: "Alias fixture", Severity: model.SeverityHigh,
		SrcIP: "192.0.2.1", DstIP: "198.51.100.9", DstPort: 80, Protocol: "TCP",
		Timestamp: at, PayloadPreview: id, MatchedKeyword: id}
}

// Seed through non-daily public calls so fixture discovery cannot duplicate rows.
func seedShardAlias(t *testing.T, path, mode string, now time.Time, input []*model.Alert) []*model.Alert {
	t.Helper()
	store, err := alert.Open(context.Background(), alert.Options{Path: path, JournalMode: mode, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.WriteBatch(context.Background(), input); err != nil {
		_ = store.Close()
		t.Fatal(err)
	}
	rows, err := store.List(context.Background())
	if err != nil || len(rows) != len(input) {
		_ = store.Close()
		t.Fatalf("seed: rows=%d error=%v", len(rows), err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return rows
}

type shardAliasArtifact struct {
	Mode os.FileMode
	Data string
}

func shardAliasTree(t *testing.T, dir string) map[string]shardAliasArtifact {
	t.Helper()
	result := map[string]shardAliasArtifact{}
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
		item := shardAliasArtifact{Mode: info.Mode()}
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

func shardAliasObserver(t *testing.T, absolutePath string) *sql.DB {
	t.Helper()
	u := url.URL{Scheme: "file", Path: absolutePath}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("readonly_shm", "1")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close read-only observer: %v", err)
		}
	})
	return db
}

func shardAliasObservedCount(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow("SELECT COUNT(*) FROM alerts").Scan(&got); err != nil || got != want {
		t.Fatalf("independent count=%d error=%v, want %d", got, err, want)
	}
}

func shardAliasDirectory(t *testing.T, dir, spelling string) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, dir)
	if err != nil {
		t.Fatal(err)
	}
	switch spelling {
	case "relative":
		return rel
	case "relative_dot":
		return "./" + rel + "/."
	case "absolute_dot":
		return dir + "/."
	case "absolute_parent":
		return dir + "/../" + filepath.Base(dir)
	default:
		return dir
	}
}

func shardAliasResolvedPath(t *testing.T, store *alert.Store, opts alert.Options, now time.Time, active string) {
	t.Helper()
	want := filepath.Join(opts.Dir, "netsentry-"+now.UTC().Format("2006-01-02")+".db")
	if store.Path() != want || store.Path() == opts.Path {
		t.Fatalf("daily Path=%q, want derived %q; ignored explicit Path=%q", store.Path(), want, opts.Path)
	}
	absolute, err := filepath.Abs(store.Path())
	if err != nil || absolute != active {
		t.Fatalf("resolved active=%q error=%v, want seeded %q", absolute, err, active)
	}
	if _, err := os.Stat(opts.Path); !os.IsNotExist(err) {
		t.Fatalf("ignored explicit Path acquired an artifact: %v", err)
	}
}

// The retained name is used by the departmental handoff. These public fixtures
// cover Dir spelling compatibility, not independently configured Path aliases:
// daily startup derives Path from Dir/Now and ignores Options.Path.
func TestDailyShardLexicalActivePathAliasesCountRowsOnce(t *testing.T) {
	for _, mode := range []string{"WAL", "DELETE"} {
		for _, spelling := range []string{"absolute", "relative", "relative_dot", "absolute_dot", "absolute_parent"} {
			t.Run(mode+"/"+spelling, func(t *testing.T) {
				ctx := context.Background()
				now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
				dir := filepath.Join(t.TempDir(), "daily paths with spaces")
				active := filepath.Join(dir, "netsentry-2026-10-04.db")
				history := filepath.Join(dir, "netsentry-2026-10-03.db")
				input := []*model.Alert{shardAliasAlert(now, "latest"), shardAliasAlert(now.Add(-10*time.Minute), "middle"), shardAliasAlert(now.Add(-24*time.Hour), "history")}
				original := []model.Alert{*input[0], *input[1], *input[2]}
				want := seedShardAlias(t, active, mode, now, input[:2])
				want = append(want, seedShardAlias(t, history, "DELETE", now, input[2:])...)
				opts := alert.Options{Path: filepath.Join(filepath.Dir(dir), "ignored explicit Path", "operator.db"),
					Dir: shardAliasDirectory(t, dir, spelling), DailyShard: true, JournalMode: mode, Now: func() time.Time { return now }}
				store, err := alert.Open(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := store.Close(); err != nil {
						t.Errorf("close store: %v", err)
					}
				})
				shardAliasResolvedPath(t, store, opts, now, active)
				activeObserver, historyObserver := shardAliasObserver(t, active), shardAliasObserver(t, history)
				shardAliasObservedCount(t, activeObserver, 2)
				shardAliasObservedCount(t, historyObserver, 1)
				before := shardAliasTree(t, dir)
				since, until := now.Add(-15*time.Minute), now
				oldest := now.Add(-24 * time.Hour)
				cases := []struct {
					query alert.Query
					total int
					rows  []*model.Alert
				}{
					{alert.Query{}, 3, want},
					{alert.Query{Limit: -1}, 3, want},
					{alert.Query{Limit: 1, Offset: 1}, 3, want[1:2]},
					{alert.Query{Limit: 1, Offset: 2}, 3, want[2:]},
					{alert.Query{Limit: 1, Offset: 3}, 3, nil},
					{alert.Query{Limit: 1, Offset: 4}, 3, nil},
					{alert.Query{Since: &since, Until: &until}, 2, want[:2]},
					{alert.Query{Until: &oldest}, 1, want[2:]},
					{alert.Query{RuleID: "history"}, 1, want[2:]},
				}
				for repeat := 0; repeat < 2; repeat++ {
					count, err := store.Count(ctx)
					if err != nil || count != 3 {
						t.Fatalf("Count=%d error=%v, want 3", count, err)
					}
					listed, err := store.List(ctx)
					if err != nil || !reflect.DeepEqual(listed, want) {
						t.Fatalf("List=%+v error=%v, want %+v", listed, err, want)
					}
					for _, tc := range cases {
						got, total, err := store.Query(ctx, tc.query)
						if err != nil || total != tc.total || len(got) != len(tc.rows) {
							t.Fatalf("Query=%+v rows=%d total=%d error=%v, want rows=%d total=%d", tc.query, len(got), total, err, len(tc.rows), tc.total)
						}
						for i, row := range tc.rows {
							if got[i] == nil || *got[i] != *row {
								t.Fatalf("row[%d]=%+v, want complete row %+v", i, got[i], row)
							}
						}
					}
				}
				if since != now.Add(-15*time.Minute) || until != now || oldest != now.Add(-24*time.Hour) {
					t.Fatal("changed query timestamps")
				}
				for i := range input {
					if *input[i] != original[i] {
						t.Fatal("changed caller alert")
					}
				}
				shardAliasObservedCount(t, activeObserver, 2)
				shardAliasObservedCount(t, historyObserver, 1)
				if !reflect.DeepEqual(shardAliasTree(t, dir), before) || store.Health().Status != "ok" {
					t.Fatal("read calls changed persistent bytes/modes/membership or health")
				}
				shardAliasResolvedPath(t, store, opts, now, active)
			})
		}
	}
}

func TestDailyShardAliasPreservesMissingDirectoryAndHistoricalErrors(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	for _, fixture := range []string{"missing_directory", "invalid_calendar", "corrupt_history"} {
		t.Run(fixture, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "daily errors with spaces")
			active := filepath.Join(dir, "netsentry-2026-10-04.db")
			want := seedShardAlias(t, active, "DELETE", now, []*model.Alert{shardAliasAlert(now, "active")})
			opts := alert.Options{Path: filepath.Join(filepath.Dir(dir), "ignored explicit Path", "operator.db"),
				Dir: shardAliasDirectory(t, dir, "relative_dot"), DailyShard: true, JournalMode: "DELETE", Now: func() time.Time { return now }}
			store, err := alert.Open(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("close: %v", err)
				}
			})
			shardAliasResolvedPath(t, store, opts, now, active)
			date := "2026-02-30"
			if fixture == "corrupt_history" {
				date = "2026-10-03"
			}
			if fixture != "missing_directory" {
				if err := os.WriteFile(filepath.Join(dir, "netsentry-"+date+".db"), []byte("unrelated non-database bytes"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			observer := shardAliasObserver(t, active)
			shardAliasObservedCount(t, observer, 1)
			if fixture == "missing_directory" {
				// Open creates its active directory. Move it only after opening and
				// observing the seeded database, so ReadDir really sees ENOENT while
				// the live store and independent observer retain the same database.
				moved := dir + " moved"
				if err := os.Rename(dir, moved); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.Rename(moved, dir); err != nil {
						t.Errorf("restore opened directory: %v", err)
					}
				})
				if _, err := os.Stat(opts.Dir); !os.IsNotExist(err) {
					t.Fatalf("discovery directory still exists: %v", err)
				}
			}
			before := shardAliasTree(t, filepath.Dir(dir))
			count, countErr := store.Count(ctx)
			listed, listErr := store.List(ctx)
			queried, total, queryErr := store.Query(ctx, alert.Query{})
			if fixture == "corrupt_history" {
				for _, err := range []error{countErr, listErr, queryErr} {
					if err == nil || !strings.Contains(err.Error(), "netsentry-2026-10-03.db") {
						t.Fatalf("missing real historical error: %v", err)
					}
				}
				if count != 0 || total != 0 || len(listed) != 0 || len(queried) != 0 || store.Health().Status == "ok" {
					t.Fatal("corrupt shard exposed partial success")
				}
			} else if countErr != nil || listErr != nil || queryErr != nil || count != 1 || total != 1 || !reflect.DeepEqual(listed, want) || !reflect.DeepEqual(queried, want) {
				t.Fatalf("fallback/control: Count=%d Query total=%d errors=%v/%v/%v", count, total, countErr, listErr, queryErr)
			}
			shardAliasObservedCount(t, observer, 1)
			if !reflect.DeepEqual(shardAliasTree(t, filepath.Dir(dir)), before) {
				t.Fatal("reads changed active/history artifacts")
			}
			if store.Path() != filepath.Join(opts.Dir, "netsentry-2026-10-04.db") {
				t.Fatal("reads changed derived daily Path")
			}
			if _, err := os.Stat(opts.Path); !os.IsNotExist(err) {
				t.Fatalf("reads created ignored explicit Path: %v", err)
			}
		})
	}
}
