package alert

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Column default and primary-key index collation are deliberately independent.
// A NOCASE primary key would fail write-safety preflight before reaching reads.
func idOrderSchema(collation string, legacyIndex bool) string {
	ddl := strings.Replace(schemaSQL, "id TEXT PRIMARY KEY,", "id TEXT COLLATE "+collation+",", 1)
	ddl = strings.Replace(ddl, "    UNIQUE(rule_id,", "    PRIMARY KEY(id COLLATE BINARY),\n    UNIQUE(rule_id,", 1)
	if legacyIndex {
		ddl = strings.Replace(ddl, "DESC, id COLLATE BINARY ASC);", "DESC, id ASC);", 1)
	}
	return ddl
}

func idOrderInput(at time.Time, id string) *model.Alert {
	return &model.Alert{RuleID: id, RuleName: "Order fixture " + id,
		Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.2",
		DstPort: 80, Protocol: "TCP", Timestamp: at,
		MatchedKeyword: id, PayloadPreview: "Payload " + id}
}

// Compute full expected contents independently of store normalization/sorting.
func idOrderExpected(input *model.Alert) *model.Alert {
	out := *input
	out.Timestamp = input.Timestamp.UTC()
	out.WindowStart = out.Timestamp.Truncate(time.Minute)
	out.FirstSeen, out.LastSeen, out.AggregatedCount = out.Timestamp, out.Timestamp, 1
	out.ID = fmt.Sprintf("%s-192.0.2.1-198.51.100.2-80-%d", input.RuleID, out.WindowStart.Unix())
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00192.0.2.1\x00198.51.100.2\x0080\x00TCP\x00%s\x00%s\x00%d",
		input.RuleID, input.MatchedKeyword, out.Timestamp.Format(time.RFC3339Nano), out.Timestamp.UnixNano())))
	out.EventID = "evt_" + hex.EncodeToString(sum[:16])
	return &out
}

func assertIDOrderFixtureMetadata(t *testing.T, db *sql.DB, collation string, legacy bool) string {
	t.Helper()
	var ddl, indexDDL string
	if err := db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='alerts'").Scan(&ddl); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ddl, "id TEXT COLLATE "+collation+",") || !strings.Contains(ddl, "PRIMARY KEY(id COLLATE BINARY)") {
		t.Fatalf("fixture lost column/primary-key separation: %s", ddl)
	}
	if err := db.QueryRow("SELECT sql FROM sqlite_master WHERE name='idx_alerts_last_seen_time_id'").Scan(&indexDDL); err != nil {
		t.Fatal(err)
	}
	wantTerm := "id COLLATE BINARY ASC"
	if legacy {
		wantTerm = "id ASC"
	}
	if !strings.Contains(indexDDL, wantTerm) {
		t.Fatalf("fixture index = %s, want %s", indexDDL, wantTerm)
	}
	return indexDDL
}

func TestStoreIDTieOrderingIgnoresColumnCollation(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 5, 12, 0, 0, 100000000, time.UTC)
	for _, collation := range []string{"BINARY", "NOCASE", "RTRIM"} {
		for _, legacy := range []bool{false, true} {
			for _, daily := range []bool{false, true} {
				for _, mode := range []string{"DELETE", "WAL"} {
					t.Run(fmt.Sprintf("%s/legacy=%v/daily=%v/%s", collation, legacy, daily, mode), func(t *testing.T) {
						dir := filepath.Join(t.TempDir(), "order path % #")
						if err := os.MkdirAll(dir, 0o750); err != nil {
							t.Fatal(err)
						}
						path := filepath.Join(dir, "primary.db")
						if daily {
							path = filepath.Join(dir, "netsentry-2026-10-05.db")
						}
						createSQLiteFixture(t, path, idOrderSchema(collation, legacy))
						opts := Options{Path: path, Dir: dir, DailyShard: daily, JournalMode: mode, Now: func() time.Time { return now }}
						if daily {
							opts.Path = filepath.Join(dir, "ignored.db")
						}
						input := []*model.Alert{idOrderInput(now, "a"), idOrderInput(now, "B"), idOrderInput(now, "c"),
							idOrderInput(now.Add(time.Nanosecond), "newer"), idOrderInput(now.Add(-time.Second), "older")}
						beforeInputs := make([]model.Alert, len(input))
						for i, a := range input {
							beforeInputs[i] = *a
						}
						want := []*model.Alert{idOrderExpected(input[3]), idOrderExpected(input[1]), idOrderExpected(input[0]), idOrderExpected(input[2]), idOrderExpected(input[4])}
						var historicalPath string
						var historicalBefore map[string]idOrderArtifact
						if daily {
							historicalPath = filepath.Join(dir, "netsentry-2026-10-04.db")
							createSQLiteFixture(t, historicalPath, idOrderSchema(collation, legacy))
							seed, err := Open(ctx, Options{Path: historicalPath, JournalMode: "DELETE", Now: opts.Now})
							if err != nil {
								t.Fatal(err)
							}
							history := idOrderInput(now.Add(-24*time.Hour), "history")
							if err := seed.WriteBatch(ctx, []*model.Alert{history}); err != nil {
								_ = seed.Close()
								t.Fatal(err)
							}
							if err := seed.Close(); err != nil {
								t.Fatal(err)
							}
							want = append(want, idOrderExpected(history))
						}
						store, err := Open(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = store.Close() })
						if store.Path() != path {
							t.Fatalf("resolved path=%s, want %s", store.Path(), path)
						}
						indexBefore := assertIDOrderFixtureMetadata(t, store.db, collation, legacy)
						if err := store.WriteBatch(ctx, input); err != nil {
							t.Fatal(err)
						}
						if daily {
							historicalBefore = idOrderHistoricalArtifacts(t, historicalPath)
						}
						for reopen := 0; reopen < 2; reopen++ {
							listed, err := store.List(ctx)
							if err != nil || !reflect.DeepEqual(listed, want) {
								t.Fatalf("List/reopen=%d: got=%+v want=%+v error=%v", reopen, listed, want, err)
							}
							for _, page := range []struct{ offset, limit, start, end int }{
								{0, -1, 0, len(want)}, {0, 0, 0, len(want)}, {0, 2, 0, 2},
								{1, 1, 1, 2}, {2, 1, 2, 3}, {3, 2, 3, 5}, {len(want), 1, len(want), len(want)},
							} {
								got, total, err := store.Query(ctx, Query{Offset: page.offset, Limit: page.limit})
								if err != nil || total != len(want) || len(got) != page.end-page.start || (len(got) > 0 && !reflect.DeepEqual(got, want[page.start:page.end])) {
									t.Fatalf("Query/reopen=%d/page=%+v: got=%+v total=%d error=%v", reopen, page, got, total, err)
								}
							}
							// Inclusive exact-time filtering leaves only the tied group.
							got, total, err := store.Query(ctx, Query{Since: &now, Until: &now, Offset: 1, Limit: 1})
							if err != nil || total != 3 || !reflect.DeepEqual(got, want[2:3]) {
								t.Fatalf("filtered tie page=%+v total=%d error=%v", got, total, err)
							}
							if count, err := store.Count(ctx); err != nil || count != len(want) || store.Health().Status != "ok" {
								t.Fatalf("count=%d health=%+v error=%v", count, store.Health(), err)
							}
							if got := assertIDOrderFixtureMetadata(t, store.db, collation, legacy); got != indexBefore {
								t.Fatalf("existing index rebuilt: before=%s after=%s", indexBefore, got)
							}
							if daily && !reflect.DeepEqual(idOrderHistoricalArtifacts(t, historicalPath), historicalBefore) {
								t.Fatal("historical shard bytes/modes/membership changed")
							}
							if reopen == 0 {
								if err := store.Close(); err != nil {
									t.Fatal(err)
								}
								store, err = Open(ctx, opts)
								if err != nil {
									t.Fatal(err)
								}
							}
						}
						for i, a := range input {
							if *a != beforeInputs[i] {
								t.Fatalf("caller input %d changed", i)
							}
						}
						if daily {
							if _, err := os.Stat(opts.Path); !os.IsNotExist(err) {
								t.Fatalf("ignored daily Path created: %v", err)
							}
						}
					})
				}
			}
		}
	}
}

type idOrderArtifact struct {
	Mode os.FileMode
	Data string
}

func idOrderHistoricalArtifacts(t *testing.T, path string) map[string]idOrderArtifact {
	t.Helper()
	result := map[string]idOrderArtifact{}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		info, err := os.Stat(path + suffix)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path + suffix)
		if err != nil {
			t.Fatal(err)
		}
		result[suffix] = idOrderArtifact{Mode: info.Mode(), Data: string(data)}
	}
	return result
}

func TestStoreNewIDOrderIndexUsesBinaryCollation(t *testing.T) {
	ctx := context.Background()
	for _, collation := range []string{"BINARY", "NOCASE", "RTRIM"} {
		t.Run(collation, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "index.db")
			// Only the tables exist. Open must create the actual new index.
			ddl := strings.Split(idOrderSchema(collation, false), "CREATE INDEX")[0]
			createSQLiteFixture(t, path, ddl)
			store, err := Open(ctx, Options{Path: path, JournalMode: "DELETE"})
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			rows, err := store.db.Query("PRAGMA index_xinfo(idx_alerts_last_seen_time_id)")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for rows.Next() {
				var seq, cid, descending, key int
				var name, coll sql.NullString
				if err := rows.Scan(&seq, &cid, &name, &descending, &coll, &key); err != nil {
					_ = rows.Close()
					t.Fatal(err)
				}
				if key == 1 && name.Valid && name.String == "id" {
					found = true
					if !coll.Valid || coll.String != "BINARY" || descending != 0 {
						t.Errorf("id index term: collation=%v descending=%d", coll, descending)
					}
				}
			}
			if err := rows.Err(); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			if err := rows.Close(); err != nil || !found {
				t.Fatalf("binary id term found=%v close error=%v", found, err)
			}
			since := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
			where, args := alertQueryWhere(Query{Since: &since})
			for _, statement := range []struct {
				sql  string
				args []any
			}{
				{alertSelectColumns + alertOrderSQL + " LIMIT 1000", nil},
				{alertSelectColumns + where + alertOrderSQL + " LIMIT ? OFFSET ?", append(args, 2, 1)},
			} {
				rows, err := store.db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+statement.sql, statement.args...)
				if err != nil {
					t.Fatal(err)
				}
				var details []string
				for rows.Next() {
					var id, parent, unused int
					var detail string
					if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
						_ = rows.Close()
						t.Fatal(err)
					}
					details = append(details, detail)
				}
				if err := rows.Err(); err != nil {
					_ = rows.Close()
					t.Fatal(err)
				}
				if err := rows.Close(); err != nil {
					t.Fatal(err)
				}
				plan := strings.Join(details, "\n")
				if !strings.Contains(plan, "idx_alerts_last_seen_time_id") || strings.Contains(plan, "USE TEMP B-TREE") {
					t.Fatalf("new expression index did not cover binary order: %s", plan)
				}
			}
		})
	}
}
