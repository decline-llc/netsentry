package alert_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

// Capture every durable column through a separately encoded read-only handle.
func activeRetentionRows(t *testing.T, db *sql.DB) map[string][][]any {
	t.Helper()
	out := map[string][][]any{}
	for _, table := range []string{"alerts", "alert_events"} {
		order := "id"
		if table == "alert_events" {
			order = "event_id"
		}
		rows, err := db.Query("SELECT * FROM " + table + " ORDER BY " + order)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			for i, value := range values {
				if data, ok := value.([]byte); ok {
					values[i] = string(data)
				}
			}
			out[table] = append(out[table], values)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return out
}

func activeRetentionSet(t *testing.T, dir, name string) []string {
	t.Helper()
	var paths []string
	for _, suffix := range []string{"", "-wal", "-shm"} {
		path := filepath.Join(dir, name+suffix)
		if err := os.WriteFile(path, []byte("inactive fixture "+name+suffix), 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	return paths
}

func activeRetentionPublicRows(t *testing.T, store *alert.Store, want []*model.Alert) {
	t.Helper()
	ctx := context.Background()
	if count, err := store.Count(ctx); err != nil || count != len(want) {
		t.Fatalf("Count=%d error=%v, want %d", count, err, len(want))
	}
	if rows, err := store.List(ctx); err != nil || !reflect.DeepEqual(rows, want) {
		t.Fatalf("List=%+v error=%v, want %+v", rows, err, want)
	}
	rows, total, err := store.Query(ctx, alert.Query{Limit: -1})
	if err != nil || total != len(want) || !reflect.DeepEqual(rows, want) {
		t.Fatalf("Query=%+v total=%d error=%v, want %+v", rows, total, err, want)
	}
}

func TestExpiredActiveShardSurvivesStartupAndLexicalCleanup(t *testing.T) {
	for _, mode := range []string{"WAL", "DELETE"} {
		for _, daily := range []bool{false, true} {
			for _, spelling := range []string{"identical", "relative_dir", "relative_path", "both_relative_dot", "absolute_dot"} {
				t.Run(fmt.Sprintf("%s/daily_%t/%s", mode, daily, spelling), func(t *testing.T) {
					ctx := context.Background()
					now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
					dir := filepath.Join(t.TempDir(), "retention paths with spaces")
					active := filepath.Join(dir, "netsentry-2026-09-01.db")
					input := []*model.Alert{shardAliasAlert(now, "active")}
					original := *input[0]
					want := seedShardAlias(t, active, mode, now, input)
					seedObserver := shardAliasObserver(t, active)
					seedRows := activeRetentionRows(t, seedObserver)
					if err := seedObserver.Close(); err != nil {
						t.Fatal(err)
					}
					initialInfo, err := os.Stat(active)
					if err != nil {
						t.Fatal(err)
					}
					removedAtStartup := activeRetentionSet(t, dir, "netsentry-2026-09-02.db")
					activeRetentionSet(t, dir, "netsentry-2026-09-27.db") // Exact cutoff is retained.
					activeRetentionSet(t, dir, "netsentry-2026-09-31.db") // Invalid calendar is retained.
					controlBefore := shardAliasTree(t, dir)
					cwd, err := os.Getwd()
					if err != nil {
						t.Fatal(err)
					}
					relDir, err := filepath.Rel(cwd, dir)
					if err != nil {
						t.Fatal(err)
					}
					opts := alert.Options{Path: active, Dir: dir, DailyShard: daily, JournalMode: mode,
						RetentionDays: 7, Now: func() time.Time { return now }}
					switch spelling {
					case "relative_dir":
						opts.Dir = relDir
					case "relative_path":
						opts.Path = filepath.Join(relDir, filepath.Base(active))
					case "both_relative_dot":
						opts.Dir, opts.Path = relDir, "./"+relDir+"/./"+filepath.Base(active)
					case "absolute_dot":
						opts.Path = dir + "/./" + filepath.Base(active)
					}
					store, err := alert.Open(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = store.Close() })
					info, err := os.Stat(active)
					if err != nil || !os.SameFile(initialInfo, info) || initialInfo.Mode() != info.Mode() || store.Path() != opts.Path {
						t.Fatalf("startup active identity/path changed: info=%v error=%v Path=%q", info, err, store.Path())
					}
					for _, path := range removedAtStartup {
						if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
							t.Fatalf("startup expired artifact remains: %s error=%v", path, err)
						}
					}
					controlsAfter := shardAliasTree(t, dir)
					for _, name := range []string{"netsentry-2026-09-27.db", "netsentry-2026-09-31.db"} {
						for _, suffix := range []string{"", "-wal", "-shm"} {
							rel := name + suffix
							if controlsAfter[rel] != controlBefore[rel] {
								t.Fatalf("retained control %s changed", rel)
							}
							// Retained controls contain arbitrary bytes. Remove them before
							// daily queries, which must reject corrupt valid-date databases.
							if err := os.Remove(filepath.Join(dir, rel)); err != nil {
								t.Fatal(err)
							}
						}
					}
					observer := shardAliasObserver(t, active)
					if got := activeRetentionRows(t, observer); !reflect.DeepEqual(got, seedRows) {
						t.Fatalf("startup durable rows changed: %+v, want %+v", got, seedRows)
					}
					activeRetentionPublicRows(t, store, want)
					before := shardAliasTree(t, dir)
					if mode == "WAL" {
						for _, suffix := range []string{"-wal", "-shm"} {
							artifact, ok := before[filepath.Base(active)+suffix]
							if !ok || !artifact.Mode.IsRegular() {
								t.Fatalf("WAL preservation fixture lacks real active %s", suffix)
							}
						}
					}
					identities := map[string]os.FileInfo{}
					for _, suffix := range []string{"", "-wal", "-shm"} {
						path := active + suffix
						info, err := os.Stat(path)
						if err != nil {
							if mode == "DELETE" && suffix != "" && errors.Is(err, os.ErrNotExist) {
								continue
							}
							t.Fatal(err)
						}
						identities[path] = info
					}
					removed := activeRetentionSet(t, dir, "netsentry-2026-09-03.db")
					for repeat := 0; repeat < 2; repeat++ {
						deleted, err := store.PruneExpiredShardFiles(ctx, opts.Dir)
						wantDeleted := 3
						if repeat > 0 {
							wantDeleted = 0
						}
						if err != nil || deleted != wantDeleted {
							t.Fatalf("cleanup=%d error=%v, want %d", deleted, err, wantDeleted)
						}
						if got := shardAliasTree(t, dir); !reflect.DeepEqual(got, before) {
							t.Fatalf("cleanup changed active bytes/modes/membership: %+v, want %+v", got, before)
						}
						for _, path := range removed {
							if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
								t.Fatalf("expired artifact remains: %s error=%v", path, err)
							}
						}
						if got := activeRetentionRows(t, observer); !reflect.DeepEqual(got, seedRows) {
							t.Fatal("cleanup changed durable alerts/events")
						}
						activeRetentionPublicRows(t, store, want)
						for path, originalInfo := range identities {
							info, err := os.Stat(path)
							if err != nil || !os.SameFile(originalInfo, info) {
								t.Fatalf("cleanup replaced active artifact %s: %v", path, err)
							}
						}
					}
					if *input[0] != original || store.Path() != opts.Path || store.Health().Status != "ok" {
						t.Fatal("caller input, original Path or health changed")
					}
					// Primary writes still target the protected pathname; daily writes
					// route to today's separate shard while retaining the explicit active file.
					next := shardAliasAlert(now.Add(time.Minute), "continued")
					nextBefore := *next
					if err := store.WriteBatch(ctx, []*model.Alert{next}); err != nil {
						t.Fatal(err)
					}
					rows, err := store.List(ctx)
					if err != nil || len(rows) != 2 || rows[0].RuleID != "continued" || !reflect.DeepEqual(rows[1], want[0]) {
						t.Fatalf("continued write rows=%+v error=%v", rows, err)
					}
					if *next != nextBefore || rows[0].RuleName != next.RuleName || rows[0].PayloadPreview != next.PayloadPreview || rows[0].MatchedKeyword != next.MatchedKeyword || !rows[0].Timestamp.Equal(next.Timestamp) || rows[0].AggregatedCount != 1 || rows[0].ID == "" || rows[0].EventID == "" {
						t.Fatal("continued write changed caller input or emitted row content")
					}
					want = rows
					wantActiveCount := 2
					if daily {
						wantActiveCount = 1
					}
					shardAliasObservedCount(t, observer, wantActiveCount)
					postWriteRows := activeRetentionRows(t, observer)
					if err := observer.Close(); err != nil {
						t.Fatal(err)
					}
					if err := store.Close(); err != nil {
						t.Fatal(err)
					}
					store, err = alert.Open(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}
					activeRetentionPublicRows(t, store, want)
					freshObserver := shardAliasObserver(t, active)
					shardAliasObservedCount(t, freshObserver, wantActiveCount)
					if !reflect.DeepEqual(activeRetentionRows(t, freshObserver), postWriteRows) {
						t.Fatal("close/reopen changed durable active alerts/events")
					}
					if err := freshObserver.Close(); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestLongLivedActiveShardCleanupPreservesFileAndSeparateRowTTL(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	store, err := alert.Open(ctx, alert.Options{Dir: dir, DailyShard: true, JournalMode: "WAL",
		RetentionDays: 7, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.WriteBatch(ctx, []*model.Alert{shardAliasAlert(now, "aged")}); err != nil {
		t.Fatal(err)
	}
	active := store.Path()
	observer := shardAliasObserver(t, active)
	rowsBefore := activeRetentionRows(t, observer)
	before := shardAliasTree(t, dir)
	now = now.AddDate(0, 0, 33)
	deleted, err := store.PruneExpiredShardFiles(ctx, dir)
	if err != nil || deleted != 0 || !reflect.DeepEqual(shardAliasTree(t, dir), before) || !reflect.DeepEqual(activeRetentionRows(t, observer), rowsBefore) {
		t.Fatalf("aged active cleanup=%d error=%v; artifacts/rows must remain", deleted, err)
	}
	pruned, err := store.PruneExpired(ctx)
	if err != nil || pruned != 1 {
		t.Fatalf("separate row TTL=%d error=%v, want 1", pruned, err)
	}
	shardAliasObservedCount(t, observer, 0)
	if _, err := os.Stat(active); err != nil || store.Path() != active {
		t.Fatalf("row TTL lost active file/path: %v", err)
	}
}

func TestActiveShardCleanupGuardsAndOtherDirectory(t *testing.T) {
	ctx := context.Background()
	for _, retention := range []int{0, 7} {
		t.Run(fmt.Sprintf("retention_%d", retention), func(t *testing.T) {
			dir := t.TempDir()
			active := filepath.Join(dir, "netsentry-2026-09-01.db")
			store, err := alert.Open(ctx, alert.Options{Path: active, Dir: dir, RetentionDays: retention,
				Now: func() time.Time { return time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC) }})
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			activeRetentionSet(t, dir, "netsentry-2026-09-02.db")
			before := shardAliasTree(t, dir)
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			deleted, err := store.PruneExpiredShardFiles(canceled, dir)
			if deleted != 0 || !errors.Is(err, context.Canceled) || !reflect.DeepEqual(shardAliasTree(t, dir), before) {
				t.Fatalf("canceled cleanup=%d error=%v or artifacts changed", deleted, err)
			}
			deleted, err = store.PruneExpiredShardFiles(ctx, filepath.Join(dir, "missing"))
			if err != nil || deleted != 0 {
				t.Fatalf("missing directory cleanup=%d error=%v", deleted, err)
			}
			other := t.TempDir()
			activeRetentionSet(t, other, filepath.Base(active))
			otherBefore := shardAliasTree(t, other)
			deleted, err = store.PruneExpiredShardFiles(ctx, other)
			if err != nil || (retention == 7 && deleted != 3) || (retention == 0 && deleted != 0) {
				t.Fatalf("other-directory cleanup=%d error=%v", deleted, err)
			}
			if retention == 0 {
				if !reflect.DeepEqual(shardAliasTree(t, other), otherBefore) {
					t.Fatal("disabled retention changed other directory")
				}
				if deleted, err := store.PruneExpiredShardFiles(ctx, dir); err != nil || deleted != 0 {
					t.Fatalf("disabled active cleanup=%d error=%v", deleted, err)
				}
			} else if len(shardAliasTree(t, other)) != 1 {
				t.Fatal("same basename in other directory was incorrectly protected")
			}
			if !reflect.DeepEqual(shardAliasTree(t, dir), before) {
				t.Fatal("guards/other-directory cleanup changed active artifacts")
			}
		})
	}
}
