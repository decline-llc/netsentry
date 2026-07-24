package alert

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestStoreAggregatesAlertsInWindow(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	base := time.Date(2026, 6, 27, 3, 0, 10, 0, time.UTC)
	alerts := []*model.Alert{
		makeAlert(base, "first"),
		makeAlert(base.Add(20*time.Second), "second"),
	}
	if err := store.WriteBatch(ctx, alerts); err != nil {
		t.Fatalf("write alerts: %v", err)
	}

	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 aggregated row, got %d", len(listed))
	}
	if listed[0].AggregatedCount != 2 {
		t.Fatalf("aggregated count = %d, want 2", listed[0].AggregatedCount)
	}
	if listed[0].MatchedKeyword != "second" {
		t.Fatalf("matched keyword = %q, want latest", listed[0].MatchedKeyword)
	}
}

func TestStoreAggregatesOutOfOrderAlertsWithoutRegressingLatestFields(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	base := time.Date(2026, 6, 27, 3, 0, 0, 0, time.UTC)
	latest := makeAlert(base.Add(40*time.Second), "latest")
	latest.PayloadPreview = "latest payload"
	earlier := makeAlert(base.Add(5*time.Second), "earlier")
	earlier.PayloadPreview = "earlier payload"
	if err := store.WriteBatch(ctx, []*model.Alert{latest}); err != nil {
		t.Fatalf("write latest alert: %v", err)
	}
	if err := store.WriteBatch(ctx, []*model.Alert{earlier}); err != nil {
		t.Fatalf("write earlier alert: %v", err)
	}

	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 aggregated row, got %d", len(listed))
	}
	got := listed[0]
	if got.AggregatedCount != 2 {
		t.Fatalf("aggregated count = %d, want 2", got.AggregatedCount)
	}
	if !got.FirstSeen.Equal(earlier.Timestamp) {
		t.Fatalf("first_seen = %s, want %s", got.FirstSeen, earlier.Timestamp)
	}
	if !got.LastSeen.Equal(latest.Timestamp) {
		t.Fatalf("last_seen = %s, want %s", got.LastSeen, latest.Timestamp)
	}
	if got.MatchedKeyword != "latest" || got.PayloadPreview != "latest payload" {
		t.Fatalf("latest fields regressed: keyword=%q payload=%q", got.MatchedKeyword, got.PayloadPreview)
	}
}

func TestStoreKeepsAggregationKeysSeparate(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	base := time.Date(2026, 6, 27, 3, 0, 10, 0, time.UTC)
	sameWindow := makeAlert(base, "same")
	otherRule := makeAlert(base.Add(10*time.Second), "rule")
	otherRule.RuleID = "rule-2"
	otherSrc := makeAlert(base.Add(20*time.Second), "src")
	otherSrc.SrcIP = "10.0.0.3"
	otherDst := makeAlert(base.Add(30*time.Second), "dst")
	otherDst.DstIP = "10.0.0.4"
	otherPort := makeAlert(base.Add(40*time.Second), "port")
	otherPort.DstPort = 443

	if err := store.WriteBatch(ctx, []*model.Alert{sameWindow, otherRule, otherSrc, otherDst, otherPort}); err != nil {
		t.Fatalf("write alerts: %v", err)
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if count != 5 {
		t.Fatalf("count = %d, want 5", count)
	}
}

func TestStoreCreatesAlertQueryIndexes(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	rows, err := store.db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'alerts'")
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	defer rows.Close()
	indexes := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan index name: %v", err)
		}
		indexes[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate indexes: %v", err)
	}
	for _, want := range []string{
		"idx_alerts_last_seen",
		"idx_alerts_rule_window",
		"idx_alerts_rule_last_seen",
		"idx_alerts_severity_last_seen",
		"idx_alerts_src_last_seen",
		"idx_alerts_dst_last_seen",
		"idx_alerts_protocol_port_last_seen",
		"idx_alerts_mitre_technique_last_seen",
		"idx_alerts_count_last_seen",
	} {
		if !indexes[want] {
			t.Fatalf("expected alert query index %q, got %+v", want, indexes)
		}
	}
}

func TestStoreQueryFiltersCountsAndPagesAlerts(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	base := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	matchFirst := makeAlert(base, "UNION SELECT")
	matchSecond := makeAlert(base.Add(10*time.Second), "UNION SELECT")
	otherRule := makeAlert(base.Add(2*time.Minute), "UNION SELECT")
	otherRule.RuleID = "rule-2"
	otherMITRE := makeAlert(base.Add(3*time.Minute), "scanner")
	otherMITRE.MitreTactic = "Discovery"
	otherMITRE.MitreTechniqueID = "T1046"
	if err := store.WriteBatch(ctx, []*model.Alert{matchFirst, matchSecond, otherRule, otherMITRE}); err != nil {
		t.Fatalf("write alerts: %v", err)
	}

	dstPort := uint16(80)
	minCount := 2
	since := base.Add(-time.Minute)
	until := base.Add(time.Minute)
	query := Query{
		RuleID:           "rule-1",
		Severity:         model.SeverityHigh,
		SrcIP:            "10.0.0.1",
		DstIP:            "10.0.0.2",
		Protocol:         "tcp",
		DstPort:          &dstPort,
		Since:            &since,
		Until:            &until,
		MitreTactic:      "initial access",
		MitreTechniqueID: "t1190",
		MatchedKeyword:   "union",
		MinCount:         &minCount,
		Limit:            10,
	}
	alerts, total, err := store.Query(ctx, query)
	if err != nil {
		t.Fatalf("query alerts: %v", err)
	}
	if total != 1 || len(alerts) != 1 {
		t.Fatalf("query returned total=%d len=%d, want 1/1", total, len(alerts))
	}
	if alerts[0].RuleID != "rule-1" || alerts[0].AggregatedCount != 2 {
		t.Fatalf("unexpected matched alert: %+v", alerts[0])
	}

	alerts, total, err = store.Query(ctx, Query{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("page alerts: %v", err)
	}
	if total != 3 || len(alerts) != 1 {
		t.Fatalf("page returned total=%d len=%d, want 3/1", total, len(alerts))
	}
}

func TestStoreReplaysRecoveryLogIdempotently(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alerts.db")
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	logged := normalizeAlerts([]*model.Alert{makeAlert(now, "replayed")}, now, time.Minute)

	logOnly := &Store{recoveryLogPath: recoveryLogPath}
	if err := logOnly.appendRecoveryLog(logged); err != nil {
		t.Fatalf("append recovery log: %v", err)
	}

	store, err := Open(ctx, Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("open store with recovery log: %v", err)
	}
	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list replayed alerts: %v", err)
	}
	if len(listed) != 1 || listed[0].MatchedKeyword != "replayed" || listed[0].AggregatedCount != 1 {
		t.Fatalf("unexpected replayed alerts: %+v", listed)
	}
	if info, err := os.Stat(recoveryLogPath); err != nil || info.Size() != 0 {
		t.Fatalf("recovery log should be truncated after replay, info=%+v err=%v", info, err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	if err := logOnly.appendRecoveryLog(logged); err != nil {
		t.Fatalf("append duplicate recovery log: %v", err)
	}
	store, err = Open(ctx, Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("reopen store with duplicate recovery log: %v", err)
	}
	defer store.Close()
	listed, err = store.List(ctx)
	if err != nil {
		t.Fatalf("list after duplicate replay: %v", err)
	}
	if len(listed) != 1 || listed[0].AggregatedCount != 1 {
		t.Fatalf("duplicate replay should not increment aggregate: %+v", listed)
	}
}

func TestStoreRejectsMalformedRecoveryLogWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alerts.db")
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	now := time.Date(2026, 7, 19, 1, 0, 0, 0, time.UTC)
	logOnly := &Store{recoveryLogPath: recoveryLogPath}
	if err := logOnly.appendRecoveryLog(normalizeAlerts([]*model.Alert{
		makeAlert(now, "valid-prefix"),
	}, now, time.Minute)); err != nil {
		t.Fatalf("append valid recovery record: %v", err)
	}
	file, err := os.OpenFile(recoveryLogPath, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("open recovery log for malformed suffix: %v", err)
	}
	if _, err := file.WriteString("{\"rule_id\":\n"); err != nil {
		_ = file.Close()
		t.Fatalf("append malformed recovery record: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close malformed recovery log: %v", err)
	}
	wantBytes, err := os.ReadFile(recoveryLogPath)
	if err != nil {
		t.Fatalf("read malformed recovery log before startup: %v", err)
	}

	store, err := Open(ctx, Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if store != nil {
		_ = store.Close()
		t.Fatal("malformed recovery log returned a usable store")
	}
	if !errors.Is(err, ErrRecoveryLogIntegrity) {
		t.Fatalf("startup error = %v, want ErrRecoveryLogIntegrity", err)
	}
	assertFileBytes(t, recoveryLogPath, wantBytes)
	assertFileDoesNotExist(t, dbPath)
	assertNoPersistedRecoveryPrefix(t, dbPath, filepath.Join(dir, "empty-recovery.jsonl"), now)
}

func TestStoreRejectsTruncatedRecoveryLogWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alerts.db")
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	now := time.Date(2026, 7, 19, 1, 5, 0, 0, time.UTC)
	logOnly := &Store{recoveryLogPath: recoveryLogPath}
	if err := logOnly.appendRecoveryLog(normalizeAlerts([]*model.Alert{
		makeAlert(now, "unterminated-final-record"),
	}, now, time.Minute)); err != nil {
		t.Fatalf("append recovery record: %v", err)
	}
	wantBytes, err := os.ReadFile(recoveryLogPath)
	if err != nil {
		t.Fatalf("read recovery log before truncation: %v", err)
	}
	wantBytes = bytes.TrimSuffix(wantBytes, []byte("\n"))
	if err := os.WriteFile(recoveryLogPath, wantBytes, 0o600); err != nil {
		t.Fatalf("write truncated recovery log: %v", err)
	}

	store, err := Open(ctx, Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if store != nil {
		_ = store.Close()
		t.Fatal("truncated recovery log returned a usable store")
	}
	if !errors.Is(err, ErrRecoveryLogIntegrity) || !strings.Contains(err.Error(), "truncated final JSONL record") {
		t.Fatalf("startup error = %v, want truncated ErrRecoveryLogIntegrity", err)
	}
	assertFileBytes(t, recoveryLogPath, wantBytes)
	assertFileDoesNotExist(t, dbPath)
	assertNoPersistedRecoveryPrefix(t, dbPath, filepath.Join(dir, "empty-recovery.jsonl"), now)
}

func TestStoreRejectsSemanticallyInvalidRecoveryLogWithoutModification(t *testing.T) {
	now := time.Date(2026, 7, 20, 5, 30, 0, 0, time.UTC)
	valid := normalizeAlert(makeAlert(now, "valid-prefix"), now, time.Minute)
	validJSON, err := json.Marshal(valid)
	if err != nil {
		t.Fatalf("marshal valid recovery record: %v", err)
	}

	fieldCases := []string{
		"id",
		"event_id",
		"rule_id",
		"src_ip",
		"dst_ip",
		"protocol",
		"timestamp",
		"first_seen",
		"last_seen",
		"window_start",
		"aggregated_count",
	}
	tests := []struct {
		name      string
		invalid   func(t *testing.T) []byte
		condition string
	}{
		{
			name:      "null",
			invalid:   func(t *testing.T) []byte { return []byte("null") },
			condition: "required field id is empty",
		},
		{
			name:      "empty object",
			invalid:   func(t *testing.T) []byte { return []byte("{}") },
			condition: "required field id is empty",
		},
	}
	for _, field := range fieldCases {
		field := field
		tests = append(tests, struct {
			name      string
			invalid   func(t *testing.T) []byte
			condition string
		}{
			name: "missing " + field,
			invalid: func(t *testing.T) []byte {
				t.Helper()
				var record map[string]any
				if err := json.Unmarshal(validJSON, &record); err != nil {
					t.Fatalf("decode valid recovery record: %v", err)
				}
				delete(record, field)
				encoded, err := json.Marshal(record)
				if err != nil {
					t.Fatalf("marshal invalid recovery record: %v", err)
				}
				return encoded
			},
			condition: "required field " + field,
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "alerts.db")
			recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
			contents := append(append(append([]byte(nil), validJSON...), '\n'), test.invalid(t)...)
			contents = append(contents, '\n')
			if err := os.WriteFile(recoveryLogPath, contents, 0o600); err != nil {
				t.Fatalf("write invalid recovery log: %v", err)
			}

			store, err := Open(context.Background(), Options{
				Path:              dbPath,
				RecoveryLogPath:   recoveryLogPath,
				JournalMode:       "WAL",
				BusyTimeoutMS:     1000,
				AggregationWindow: time.Minute,
				Now:               func() time.Time { return now },
			})
			if store != nil {
				_ = store.Close()
				t.Fatal("semantically invalid recovery log returned a usable store")
			}
			if !errors.Is(err, ErrRecoveryLogIntegrity) || !strings.Contains(err.Error(), "record 2") || !strings.Contains(err.Error(), test.condition) {
				t.Fatalf("startup error = %v, want record 2 semantic ErrRecoveryLogIntegrity containing %q", err, test.condition)
			}
			assertFileBytes(t, recoveryLogPath, contents)
			assertFileDoesNotExist(t, dbPath)
			assertNoPersistedRecoveryPrefix(t, dbPath, filepath.Join(dir, "empty-recovery.jsonl"), now)
		})
	}
}

func TestStoreRejectsInconsistentNormalizedRecoveryLogWithoutModification(t *testing.T) {
	now := time.Date(2026, 7, 21, 4, 0, 0, 0, time.UTC)
	window := time.Minute
	valid := normalizeAlert(makeAlert(now, "valid-prefix"), now, window)
	validJSON, err := json.Marshal(valid)
	if err != nil {
		t.Fatalf("marshal valid recovery record: %v", err)
	}

	tests := []struct {
		name      string
		mutate    func(*model.Alert)
		condition string
	}{
		{
			name:      "durable id",
			mutate:    func(alert *model.Alert) { alert.ID = "inconsistent-id" },
			condition: "field id does not match normalized identity",
		},
		{
			name:      "first seen",
			mutate:    func(alert *model.Alert) { alert.FirstSeen = alert.Timestamp.Add(-time.Second) },
			condition: "field first_seen does not match timestamp",
		},
		{
			name:      "last seen",
			mutate:    func(alert *model.Alert) { alert.LastSeen = alert.Timestamp.Add(time.Second) },
			condition: "field last_seen does not match timestamp",
		},
		{
			name:      "window start",
			mutate:    func(alert *model.Alert) { alert.WindowStart = alert.WindowStart.Add(window) },
			condition: "field window_start does not match aggregation window",
		},
		{
			name:      "aggregate count",
			mutate:    func(alert *model.Alert) { alert.AggregatedCount = 2 },
			condition: "required field aggregated_count must equal 1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invalid := valid
			test.mutate(&invalid)
			invalidJSON, err := json.Marshal(invalid)
			if err != nil {
				t.Fatalf("marshal inconsistent recovery record: %v", err)
			}
			contents := append(append(append([]byte(nil), validJSON...), '\n'), invalidJSON...)
			contents = append(contents, '\n')

			for _, existing := range []bool{false, true} {
				state := "missing database"
				if existing {
					state = "existing database"
				}
				t.Run(state, func(t *testing.T) {
					dir := t.TempDir()
					dbPath := filepath.Join(dir, "alerts.db")
					recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
					var beforeDB []byte
					if existing {
						createSQLiteFixture(t, dbPath, schemaSQL, `DROP INDEX idx_alerts_last_seen`)
						beforeDB = readFileBytes(t, dbPath)
					}
					if err := os.WriteFile(recoveryLogPath, contents, 0o600); err != nil {
						t.Fatalf("write inconsistent recovery log: %v", err)
					}

					store, err := Open(context.Background(), Options{
						Path:              dbPath,
						RecoveryLogPath:   recoveryLogPath,
						JournalMode:       "WAL",
						BusyTimeoutMS:     1000,
						AggregationWindow: window,
						Now:               func() time.Time { return now },
					})
					if store != nil {
						_ = store.Close()
						t.Fatal("inconsistent recovery log returned a usable store")
					}
					if !errors.Is(err, ErrRecoveryLogIntegrity) || !strings.Contains(err.Error(), "record 2") || !strings.Contains(err.Error(), test.condition) {
						t.Fatalf("startup error = %v, want record 2 ErrRecoveryLogIntegrity containing %q", err, test.condition)
					}
					assertFileBytes(t, recoveryLogPath, contents)
					if existing {
						assertFileBytes(t, dbPath, beforeDB)
						assertSQLiteIndexMissing(t, dbPath, "idx_alerts_last_seen")
					} else {
						assertFileDoesNotExist(t, dbPath)
					}
				})
			}
		})
	}
}

func TestStoreRecoveryPreflightPreservesCompatibleDatabase(t *testing.T) {
	tests := []struct {
		name     string
		contents []byte
	}{
		{name: "malformed", contents: []byte("{\"rule_id\":\n")},
		{name: "semantic", contents: []byte("{}\n")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "alerts.db")
			recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
			createSQLiteFixture(t, dbPath, schemaSQL, `DROP INDEX idx_alerts_last_seen`)
			beforeDB := readFileBytes(t, dbPath)
			if err := os.WriteFile(recoveryLogPath, test.contents, 0o600); err != nil {
				t.Fatalf("write invalid recovery log: %v", err)
			}

			store, err := Open(context.Background(), Options{
				Path:            dbPath,
				RecoveryLogPath: recoveryLogPath,
				JournalMode:     "WAL",
			})
			if store != nil {
				_ = store.Close()
				t.Fatal("invalid recovery log returned a usable store")
			}
			if !errors.Is(err, ErrRecoveryLogIntegrity) {
				t.Fatalf("startup error = %v, want ErrRecoveryLogIntegrity", err)
			}
			assertFileBytes(t, recoveryLogPath, test.contents)
			assertFileBytes(t, dbPath, beforeDB)
			assertSQLiteIndexMissing(t, dbPath, "idx_alerts_last_seen")
		})
	}
}

func TestStoreRecoveryPreflightReplaysIntoCompatibleDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alerts.db")
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	createSQLiteFixture(t, dbPath, schemaSQL, `DROP INDEX idx_alerts_last_seen`)
	now := time.Date(2026, 7, 20, 5, 35, 0, 0, time.UTC)
	logOnly := &Store{recoveryLogPath: recoveryLogPath}
	if err := logOnly.appendRecoveryLog(normalizeAlerts([]*model.Alert{
		makeAlert(now, "valid-existing-database"),
	}, now, time.Minute)); err != nil {
		t.Fatalf("append valid recovery record: %v", err)
	}

	store, err := Open(context.Background(), Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("open compatible database with valid recovery input: %v", err)
	}
	defer store.Close()
	alerts, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("list replayed alerts: %v", err)
	}
	if len(alerts) != 1 || alerts[0].MatchedKeyword != "valid-existing-database" {
		t.Fatalf("unexpected replayed alerts: %+v", alerts)
	}
	if info, err := os.Stat(recoveryLogPath); err != nil || info.Size() != 0 {
		t.Fatalf("recovery log should be truncated after replay, info=%+v err=%v", info, err)
	}
	var indexCount int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_alerts_last_seen'`).Scan(&indexCount); err != nil {
		t.Fatalf("inspect recreated optional index: %v", err)
	}
	if indexCount != 1 {
		t.Fatalf("optional index count = %d, want 1 after valid initialization", indexCount)
	}
}

func TestStoreKeepsValidRecoveryLogWhenPersistenceFails(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	now := time.Date(2026, 7, 19, 1, 10, 0, 0, time.UTC)
	store, err := Open(ctx, Options{
		Path:              filepath.Join(dir, "alerts.db"),
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.appendRecoveryLog(normalizeAlerts([]*model.Alert{
		makeAlert(now, "pending-valid-record"),
	}, now, time.Minute)); err != nil {
		_ = store.Close()
		t.Fatalf("append valid recovery record: %v", err)
	}
	wantBytes, err := os.ReadFile(recoveryLogPath)
	if err != nil {
		_ = store.Close()
		t.Fatalf("read valid recovery log: %v", err)
	}
	if err := store.db.Close(); err != nil {
		t.Fatalf("close database before replay: %v", err)
	}
	if err := store.ReplayRecoveryLog(ctx); err == nil {
		t.Fatal("replay with closed database succeeded")
	}
	assertFileBytes(t, recoveryLogPath, wantBytes)
}

func TestStoreWriteBatchPersistsExistingRecoveryLogBeforeTruncate(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	oldAlert := makeAlert(now, "old-pending")
	oldLogged := normalizeAlerts([]*model.Alert{oldAlert}, now, time.Minute)

	store, err := Open(ctx, Options{
		Path:              filepath.Join(dir, "alerts.db"),
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.appendRecoveryLog(oldLogged); err != nil {
		t.Fatalf("append old pending recovery log: %v", err)
	}
	if err := store.WriteBatch(ctx, []*model.Alert{makeAlert(now.Add(10*time.Second), "new-current")}); err != nil {
		t.Fatalf("write current alert: %v", err)
	}
	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(listed) != 1 || listed[0].AggregatedCount != 2 || listed[0].MatchedKeyword != "new-current" {
		t.Fatalf("expected old pending and new current alerts to persist before truncate: %+v", listed)
	}
}

func TestStoreWriteBatchRejectsInvalidRecoveryLogBeforeAppend(t *testing.T) {
	now := time.Date(2026, 7, 22, 2, 0, 0, 0, time.UTC)
	valid := normalizeAlert(makeAlert(now, "existing-invalid"), now, time.Minute)
	validJSON, err := json.Marshal(valid)
	if err != nil {
		t.Fatalf("marshal valid recovery record: %v", err)
	}
	inconsistent := valid
	inconsistent.ID = "inconsistent-id"
	inconsistentJSON, err := json.Marshal(inconsistent)
	if err != nil {
		t.Fatalf("marshal inconsistent recovery record: %v", err)
	}

	tests := []struct {
		name      string
		contents  []byte
		condition string
	}{
		{name: "malformed", contents: []byte("{\"rule_id\":\n"), condition: "decode record 1"},
		{name: "truncated", contents: validJSON, condition: "truncated final JSONL record"},
		{name: "semantic", contents: []byte("{}\n"), condition: "required field id is empty"},
		{name: "normalized invariant", contents: append(inconsistentJSON, '\n'), condition: "field id does not match normalized identity"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "alerts.db")
			recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
			store, err := Open(context.Background(), Options{
				Path:              dbPath,
				RecoveryLogPath:   recoveryLogPath,
				JournalMode:       "WAL",
				BusyTimeoutMS:     1000,
				AggregationWindow: time.Minute,
				Now:               func() time.Time { return now },
			})
			if err != nil {
				t.Fatalf("open store: %v", err)
			}
			defer store.Close()

			if err := os.WriteFile(recoveryLogPath, test.contents, 0o600); err != nil {
				t.Fatalf("write invalid runtime recovery log: %v", err)
			}
			beforeDB := readFileBytes(t, dbPath)
			err = store.WriteBatch(context.Background(), []*model.Alert{
				makeAlert(now.Add(time.Second), "must-not-append"),
			})
			if !errors.Is(err, ErrRecoveryLogIntegrity) || !strings.Contains(err.Error(), test.condition) {
				t.Fatalf("WriteBatch error = %v, want ErrRecoveryLogIntegrity containing %q", err, test.condition)
			}
			assertFileBytes(t, recoveryLogPath, test.contents)
			assertFileBytes(t, dbPath, beforeDB)
			assertSQLiteAlertCount(t, dbPath, 0)
		})
	}
}

func TestStoreWritesRecoveryRecordAboveFormerScannerLimit(t *testing.T) {
	now := time.Date(2026, 7, 22, 2, 30, 0, 0, time.UTC)
	alert := makeRecoveryAlertWithEncodedSize(t, 70<<10, now)
	dir := t.TempDir()
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	store, err := Open(context.Background(), Options{
		Path:              filepath.Join(dir, "alerts.db"),
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.WriteBatch(context.Background(), []*model.Alert{alert}); err != nil {
		t.Fatalf("write large recovery record: %v", err)
	}
	listed, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("list large alert: %v", err)
	}
	if len(listed) != 1 || listed[0].RuleName != alert.RuleName {
		t.Fatalf("large alert was not preserved: count=%d", len(listed))
	}
	if info, err := os.Stat(recoveryLogPath); err != nil || info.Size() != 0 {
		t.Fatalf("recovery log should be truncated after large-record persistence, info=%+v err=%v", info, err)
	}
}

func TestStoreReplaysRecoveryRecordAboveFormerScannerLimit(t *testing.T) {
	now := time.Date(2026, 7, 22, 2, 35, 0, 0, time.UTC)
	alert := makeRecoveryAlertWithEncodedSize(t, 70<<10, now)
	dir := t.TempDir()
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	logOnly := &Store{recoveryLogPath: recoveryLogPath}
	if err := logOnly.appendRecoveryLog([]*model.Alert{alert}); err != nil {
		t.Fatalf("append large recovery record: %v", err)
	}

	store, err := Open(context.Background(), Options{
		Path:              filepath.Join(dir, "alerts.db"),
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("replay large recovery record: %v", err)
	}
	defer store.Close()
	listed, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("list replayed large alert: %v", err)
	}
	if len(listed) != 1 || listed[0].RuleName != alert.RuleName {
		t.Fatalf("large replay was not preserved: count=%d", len(listed))
	}
}

func TestStoreAcceptsRecoveryRecordAtDurableLimit(t *testing.T) {
	now := time.Date(2026, 7, 22, 2, 40, 0, 0, time.UTC)
	alert := makeRecoveryAlertWithEncodedSize(t, maxRecoveryRecordBytes, now)
	logOnly := &Store{
		recoveryLogPath:   filepath.Join(t.TempDir(), "alerts-recovery.jsonl"),
		aggregationWindow: time.Minute,
	}
	if err := logOnly.appendRecoveryLog([]*model.Alert{alert}); err != nil {
		t.Fatalf("append boundary recovery record: %v", err)
	}
	alerts, err := logOnly.readRecoveryLog()
	if err != nil {
		t.Fatalf("read boundary recovery record: %v", err)
	}
	if len(alerts) != 1 || alerts[0].RuleName != alert.RuleName {
		t.Fatalf("boundary recovery record changed: count=%d", len(alerts))
	}
}

func TestStoreRejectsOversizedRecoveryBatchBeforeAppend(t *testing.T) {
	now := time.Date(2026, 7, 22, 2, 45, 0, 0, time.UTC)
	valid := normalizeAlert(makeAlert(now, "existing-valid"), now, time.Minute)
	oversized := makeRecoveryAlertWithEncodedSize(t, maxRecoveryRecordBytes+1, now.Add(time.Second))
	recoveryLogPath := filepath.Join(t.TempDir(), "alerts-recovery.jsonl")
	logOnly := &Store{recoveryLogPath: recoveryLogPath}
	if err := logOnly.appendRecoveryLog([]*model.Alert{&valid}); err != nil {
		t.Fatalf("seed recovery log: %v", err)
	}
	wantBytes := readFileBytes(t, recoveryLogPath)

	err := logOnly.appendRecoveryLog([]*model.Alert{&valid, oversized})
	if !errors.Is(err, ErrRecoveryRecordTooLarge) || !strings.Contains(err.Error(), "record 2") {
		t.Fatalf("append oversized batch error = %v, want record 2 ErrRecoveryRecordTooLarge", err)
	}
	assertFileBytes(t, recoveryLogPath, wantBytes)
}

func TestStoreWriteBatchRejectsOversizedRecoveryRecordWithoutModification(t *testing.T) {
	now := time.Date(2026, 7, 22, 2, 50, 0, 0, time.UTC)
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alerts.db")
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	store, err := Open(context.Background(), Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	wantLog := readFileBytes(t, recoveryLogPath)
	wantDB := readFileBytes(t, dbPath)
	valid := makeAlert(now, "valid-prefix-must-not-append")
	oversized := makeRecoveryAlertWithEncodedSize(t, maxRecoveryRecordBytes+1, now.Add(time.Second))

	err = store.WriteBatch(context.Background(), []*model.Alert{valid, oversized})
	if !errors.Is(err, ErrRecoveryRecordTooLarge) || !strings.Contains(err.Error(), "record 2") {
		t.Fatalf("WriteBatch oversized error = %v, want record 2 ErrRecoveryRecordTooLarge", err)
	}
	assertFileBytes(t, recoveryLogPath, wantLog)
	assertFileBytes(t, dbPath, wantDB)
	assertSQLiteAlertCount(t, dbPath, 0)
}

func TestStoreWritesDistinctEventsInSameAggregationWindow(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	base := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	if err := store.WriteBatch(ctx, []*model.Alert{
		makeAlert(base, "first"),
		makeAlert(base.Add(10*time.Second), "second"),
	}); err != nil {
		t.Fatalf("write alerts: %v", err)
	}
	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(listed) != 1 || listed[0].AggregatedCount != 2 {
		t.Fatalf("same-window distinct events should aggregate to count 2: %+v", listed)
	}
}

func TestStoreRejectsUnsupportedJournalMode(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, Options{
		Path:              filepath.Join(t.TempDir(), "alerts.db"),
		JournalMode:       "INVALID",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
	})
	if err == nil {
		store.Close()
		t.Fatal("expected unsupported journal mode error")
	}
}

func TestStoreRejectsCorruptExistingDatabaseWithoutModification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	before := []byte("not a sqlite database\x00corrupt bytes")
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if store != nil {
		_ = store.Close()
		t.Fatal("corrupt existing database returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
}

func TestStoreRejectsTruncatedExistingDatabaseWithoutModification(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "alerts.db")
	store, err := Open(ctx, Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("create valid database: %v", err)
	}
	if err := store.WriteBatch(ctx, []*model.Alert{makeAlert(time.Now().UTC(), "before-truncate")}); err != nil {
		t.Fatalf("seed valid database: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close valid database: %v", err)
	}
	valid, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(valid) < 1024 {
		t.Fatalf("valid database unexpectedly small: %d bytes", len(valid))
	}
	before := append([]byte(nil), valid[:len(valid)/2]...)
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}

	store, err = Open(ctx, Options{Path: path, JournalMode: "DELETE"})
	if store != nil {
		_ = store.Close()
		t.Fatal("truncated existing database returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
}

func TestStoreInitializesExistingEmptyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if err != nil {
		t.Fatalf("initialize existing empty database: %v", err)
	}
	defer store.Close()
	if count, err := store.Count(context.Background()); err != nil || count != 0 {
		t.Fatalf("new database count=%d err=%v, want 0/nil", count, err)
	}
}

func TestStoreRejectsUnrelatedSQLiteSchemaWithoutModification(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "schema fixtures")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "operator data.db")
	createSQLiteFixture(t, path, `CREATE TABLE operator_data (id INTEGER PRIMARY KEY, value TEXT)`)
	before := readFileBytes(t, path)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if store != nil {
		_ = store.Close()
		t.Fatal("unrelated SQLite schema returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "required table alerts is missing") {
		t.Fatalf("startup error = %v, want missing alerts table", err)
	}
}

func TestStoreRejectsMissingRequiredColumnWithoutModification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL, `ALTER TABLE alerts RENAME COLUMN dst_port TO legacy_port`)
	before := readFileBytes(t, path)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if store != nil {
		_ = store.Close()
		t.Fatal("schema missing a required column returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "required column alerts.dst_port is missing") {
		t.Fatalf("startup error = %v, want missing dst_port", err)
	}
}

func TestStoreRejectsIncompatibleRequiredColumnWithoutModification(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		replacement string
		column      string
	}{
		{name: "type", old: "dst_port INTEGER NOT NULL", replacement: "dst_port TEXT NOT NULL", column: "dst_port"},
		{name: "not null", old: "dst_port INTEGER NOT NULL", replacement: "dst_port INTEGER", column: "dst_port"},
		{name: "primary key", old: "id TEXT PRIMARY KEY", replacement: "id TEXT NOT NULL", column: "id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alerts.db")
			incompatible := strings.Replace(schemaSQL, tt.old, tt.replacement, 1)
			createSQLiteFixture(t, path, incompatible)
			before := readFileBytes(t, path)

			store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
			if store != nil {
				_ = store.Close()
				t.Fatal("schema with an incompatible required column returned a usable store")
			}
			assertIntegrityRejectionPreservesFile(t, path, before, err)
			if !strings.Contains(err.Error(), "required column alerts."+tt.column+" has an incompatible definition") {
				t.Fatalf("startup error = %v, want incompatible %s", err, tt.column)
			}
		})
	}
}

func TestStoreRejectsMissingAlertEventsTableWithoutModification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL, `DROP TABLE alert_events`)
	before := readFileBytes(t, path)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if store != nil {
		_ = store.Close()
		t.Fatal("schema missing alert_events returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "required table alert_events is missing") {
		t.Fatalf("startup error = %v, want missing alert_events", err)
	}
}

func TestStoreRejectsUnknownRequiredColumnsWithoutModification(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		table     string
		column    string
	}{
		{
			name:      "alerts without default",
			statement: `ALTER TABLE alerts ADD COLUMN operator_required TEXT NOT NULL`,
			table:     "alerts",
			column:    "operator_required",
		},
		{
			name:      "alert events without default",
			statement: `ALTER TABLE alert_events ADD COLUMN operator_required TEXT NOT NULL`,
			table:     "alert_events",
			column:    "operator_required",
		},
		{
			name:      "literal null default",
			statement: `ALTER TABLE alerts ADD COLUMN operator_null TEXT NOT NULL DEFAULT NULL`,
			table:     "alerts",
			column:    "operator_null",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alerts.db")
			createSQLiteFixture(t, path, schemaSQL, test.statement)
			before := readFileBytes(t, path)

			store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
			if store != nil {
				_ = store.Close()
				t.Fatal("schema with an unknown required column returned a usable store")
			}
			assertIntegrityRejectionPreservesFile(t, path, before, err)
			condition := "unknown column " + test.table + "." + test.column + " is NOT NULL without a usable default"
			if !strings.Contains(err.Error(), condition) {
				t.Fatalf("startup error = %v, want %q", err, condition)
			}
		})
	}
}

func TestStoreAllowsCompatibleExtraColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL,
		`ALTER TABLE alerts ADD COLUMN operator_note TEXT`,
		`ALTER TABLE alert_events ADD COLUMN ingest_source TEXT NOT NULL DEFAULT 'legacy'`,
	)
	store, err := Open(context.Background(), Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("open schema with compatible extra columns: %v", err)
	}
	defer store.Close()
	if err := store.WriteBatch(context.Background(), []*model.Alert{
		makeAlert(time.Date(2026, 7, 22, 3, 0, 0, 0, time.UTC), "compatible-extra-columns"),
	}); err != nil {
		t.Fatalf("write schema with compatible extra columns: %v", err)
	}
	var source string
	if err := store.db.QueryRow(`SELECT ingest_source FROM alert_events`).Scan(&source); err != nil {
		t.Fatalf("read defaulted extra column: %v", err)
	}
	if source != "legacy" {
		t.Fatalf("defaulted extra column = %q, want legacy", source)
	}
}

func TestStoreRejectsMissingAggregationConstraintWithoutModification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	withoutAggregationKey := strings.Replace(schemaSQL,
		"    updated_at TEXT NOT NULL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		"    updated_at TEXT NOT NULL", 1)
	createSQLiteFixture(t, path, withoutAggregationKey)
	before := readFileBytes(t, path)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if store != nil {
		_ = store.Close()
		t.Fatal("schema without aggregation uniqueness returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "aggregation uniqueness constraint is missing") {
		t.Fatalf("startup error = %v, want missing aggregation constraint", err)
	}
}

func TestStoreRejectsWriteBlockingUniqueIndexesWithoutModification(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		table     string
		index     string
	}{
		{
			name:      "alerts subset",
			statement: `CREATE UNIQUE INDEX operator_alert_rule_unique ON alerts(rule_id)`,
			table:     "alerts",
			index:     "operator_alert_rule_unique",
		},
		{
			name:      "alert events timestamp",
			statement: `CREATE UNIQUE INDEX operator_event_time_unique ON alert_events(created_at)`,
			table:     "alert_events",
			index:     "operator_event_time_unique",
		},
		{
			name:      "alerts expression",
			statement: `CREATE UNIQUE INDEX operator_alert_rule_expression_unique ON alerts(lower(rule_id))`,
			table:     "alerts",
			index:     "operator_alert_rule_expression_unique",
		},
		{
			name:      "alerts partial subset",
			statement: `CREATE UNIQUE INDEX operator_alert_rule_partial_unique ON alerts(rule_id) WHERE severity = 'high'`,
			table:     "alerts",
			index:     "operator_alert_rule_partial_unique",
		},
		{
			name:      "alerts collated identity",
			statement: `CREATE UNIQUE INDEX operator_alert_identity_nocase_unique ON alerts(id COLLATE NOCASE)`,
			table:     "alerts",
			index:     "operator_alert_identity_nocase_unique",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alerts.db")
			createSQLiteFixture(t, path, schemaSQL, test.statement)
			before := readFileBytes(t, path)

			store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
			if store != nil {
				_ = store.Close()
				t.Fatal("schema with a write-blocking unique index returned a usable store")
			}
			assertIntegrityRejectionPreservesFile(t, path, before, err)
			condition := "unique index " + test.table + "." + test.index + " can reject valid alert writes"
			if !strings.Contains(err.Error(), condition) {
				t.Fatalf("startup error = %v, want %q", err, condition)
			}
		})
	}
}

func TestStoreAllowsWriteCompatibleAdditionalIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL,
		`CREATE INDEX operator_rule_lookup ON alerts(rule_id)`,
		`CREATE UNIQUE INDEX operator_alert_identity_unique ON alerts(id, rule_name)`,
		`CREATE UNIQUE INDEX operator_alert_aggregate_extension_unique ON alerts(rule_id, src_ip, dst_ip, dst_port, window_start, severity)`,
		`CREATE UNIQUE INDEX operator_event_identity_unique ON alert_events(event_id, created_at)`,
	)
	store, err := Open(context.Background(), Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("open schema with compatible additional indexes: %v", err)
	}
	defer store.Close()

	base := time.Date(2026, 7, 23, 4, 30, 0, 0, time.UTC)
	first := makeAlert(base, "compatible-index-one")
	second := makeAlert(base.Add(time.Second), "compatible-index-two")
	second.SrcIP = "10.0.0.3"
	if err := store.WriteBatch(context.Background(), []*model.Alert{first, second}); err != nil {
		t.Fatalf("write schema with compatible additional indexes: %v", err)
	}
	if count, err := store.Count(context.Background()); err != nil || count != 2 {
		t.Fatalf("compatible additional index count=%d err=%v, want 2/nil", count, err)
	}
}

func TestStoreRejectsWriteCriticalTriggersWithoutModification(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		table     string
		trigger   string
	}{
		{
			name:      "alerts before insert",
			statement: `CREATE TRIGGER operator_alert_insert BEFORE INSERT ON alerts BEGIN SELECT RAISE(ABORT, 'blocked alert insert'); END`,
			table:     "alerts",
			trigger:   "operator_alert_insert",
		},
		{
			name:      "alerts after update",
			statement: `CREATE TRIGGER operator_alert_update AFTER UPDATE ON alerts BEGIN SELECT RAISE(ABORT, 'blocked alert update'); END`,
			table:     "alerts",
			trigger:   "operator_alert_update",
		},
		{
			name:      "alert events after insert",
			statement: `CREATE TRIGGER operator_event_insert AFTER INSERT ON alert_events BEGIN SELECT RAISE(ABORT, 'blocked event insert'); END`,
			table:     "alert_events",
			trigger:   "operator_event_insert",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alerts.db")
			createSQLiteFixture(t, path, schemaSQL, test.statement)
			before := readFileBytes(t, path)

			store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
			if store != nil {
				_ = store.Close()
				t.Fatal("schema with a write-critical trigger returned a usable store")
			}
			assertIntegrityRejectionPreservesFile(t, path, before, err)
			condition := "trigger " + test.table + "." + test.trigger + " can alter or reject valid alert writes"
			if !strings.Contains(err.Error(), condition) {
				t.Fatalf("startup error = %v, want %q", err, condition)
			}
		})
	}
}

func TestStoreRejectsGeneratedColumnsWithoutModification(t *testing.T) {
	alertsVirtual := strings.Replace(
		schemaSQL,
		"    updated_at TEXT NOT NULL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		"    updated_at TEXT NOT NULL,\n    operator_generated TEXT GENERATED ALWAYS AS (rule_id || ':' || src_ip) VIRTUAL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		1,
	)
	alertsStored := strings.Replace(
		schemaSQL,
		"    updated_at TEXT NOT NULL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		"    updated_at TEXT NOT NULL,\n    operator_generated TEXT GENERATED ALWAYS AS (rule_id || ':' || src_ip) STORED,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		1,
	)
	eventsVirtual := strings.Replace(
		schemaSQL,
		"CREATE TABLE IF NOT EXISTS alert_events (\n    event_id TEXT PRIMARY KEY,\n    created_at TEXT NOT NULL\n);",
		"CREATE TABLE IF NOT EXISTS alert_events (\n    event_id TEXT PRIMARY KEY,\n    created_at TEXT NOT NULL,\n    operator_generated INTEGER GENERATED ALWAYS AS (length(created_at)) VIRTUAL\n);",
		1,
	)
	eventsStored := strings.Replace(
		schemaSQL,
		"CREATE TABLE IF NOT EXISTS alert_events (\n    event_id TEXT PRIMARY KEY,\n    created_at TEXT NOT NULL\n);",
		"CREATE TABLE IF NOT EXISTS alert_events (\n    event_id TEXT PRIMARY KEY,\n    created_at TEXT NOT NULL,\n    operator_generated INTEGER GENERATED ALWAYS AS (length(created_at)) STORED\n);",
		1,
	)
	tests := []struct {
		name   string
		schema string
		table  string
	}{
		{name: "alerts virtual", schema: alertsVirtual, table: "alerts"},
		{name: "alerts stored", schema: alertsStored, table: "alerts"},
		{name: "alert events virtual", schema: eventsVirtual, table: "alert_events"},
		{name: "alert events stored", schema: eventsStored, table: "alert_events"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alerts.db")
			createSQLiteFixture(t, path, test.schema)
			before := readFileBytes(t, path)

			store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
			if store != nil {
				_ = store.Close()
				t.Fatal("schema with a generated column returned a usable store")
			}
			assertIntegrityRejectionPreservesFile(t, path, before, err)
			condition := "unknown generated column " + test.table + ".operator_generated can alter or reject valid alert writes"
			if !strings.Contains(err.Error(), condition) {
				t.Fatalf("startup error = %v, want %q", err, condition)
			}
		})
	}
}

func TestStoreRejectsWriteCriticalCheckConstraintsWithoutModification(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		table  string
	}{
		{
			name: "alerts table constraint",
			schema: strings.Replace(
				schemaSQL,
				"    updated_at TEXT NOT NULL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
				"    updated_at TEXT NOT NULL,\n    CHECK (severity <> 'blocked'),\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
				1,
			),
			table: "alerts",
		},
		{
			name: "alerts column constraint",
			schema: strings.Replace(
				schemaSQL,
				"    severity TEXT NOT NULL,",
				"    severity TEXT NOT NULL CHECK (length(severity) > 0),",
				1,
			),
			table: "alerts",
		},
		{
			name: "alert events case variant constraint",
			schema: strings.Replace(
				schemaSQL,
				"    created_at TEXT NOT NULL\n);",
				"    created_at TEXT NOT NULL cHeCk (length(created_at) > 0)\n);",
				1,
			),
			table: "alert_events",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alerts.db")
			createSQLiteFixture(t, path, test.schema)
			before := readFileBytes(t, path)

			store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
			if store != nil {
				_ = store.Close()
				t.Fatal("schema with a write-critical CHECK constraint returned a usable store")
			}
			assertIntegrityRejectionPreservesFile(t, path, before, err)
			condition := "CHECK constraint on " + test.table + " can reject valid alert writes"
			if !strings.Contains(err.Error(), condition) {
				t.Fatalf("startup error = %v, want %q", err, condition)
			}
		})
	}
}

func TestStoreAllowsCheckKeywordOutsideConstraint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	compatibleSchema := strings.Replace(
		schemaSQL,
		"    updated_at TEXT NOT NULL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		"    updated_at TEXT NOT NULL,\n"+
			"    precheck_note TEXT DEFAULT 'CHECK',\n"+
			"    \"CHECK\" TEXT, -- CHECK (ignored_comment)\n"+
			"    [CHECK note] TEXT, /* CHECK (ignored_block_comment) */\n"+
			"    `CHECK value` TEXT,\n"+
			"    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		1,
	)
	createSQLiteFixture(t, path, compatibleSchema)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("open schema with non-constraint CHECK text: %v", err)
	}
	defer store.Close()
	if err := store.WriteBatch(context.Background(), []*model.Alert{
		makeAlert(time.Date(2026, 7, 24, 7, 0, 0, 0, time.UTC), "compatible-check-text"),
	}); err != nil {
		t.Fatalf("write schema with non-constraint CHECK text: %v", err)
	}
}

func TestStoreAllowsUnrelatedTableCheckConstraint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL,
		`CREATE TABLE operator_data (id INTEGER PRIMARY KEY, value TEXT NOT NULL CHECK (length(value) > 0))`,
	)
	store, err := Open(context.Background(), Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("open schema with unrelated CHECK constraint: %v", err)
	}
	defer store.Close()
	if err := store.WriteBatch(context.Background(), []*model.Alert{
		makeAlert(time.Date(2026, 7, 24, 7, 5, 0, 0, time.UTC), "unrelated-check-constraint"),
	}); err != nil {
		t.Fatalf("write schema with unrelated CHECK constraint: %v", err)
	}
}

func TestStoreAllowsUnrelatedTableTrigger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL,
		`CREATE TABLE operator_data (id INTEGER PRIMARY KEY, value TEXT NOT NULL)`,
		`CREATE TABLE operator_audit (operator_id INTEGER NOT NULL, value TEXT NOT NULL)`,
		`CREATE TRIGGER operator_data_audit AFTER INSERT ON operator_data BEGIN INSERT INTO operator_audit (operator_id, value) VALUES (NEW.id, NEW.value); END`,
	)
	store, err := Open(context.Background(), Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("open schema with unrelated table trigger: %v", err)
	}
	defer store.Close()

	base := time.Date(2026, 7, 23, 4, 45, 0, 0, time.UTC)
	first := makeAlert(base, "unrelated-trigger-one")
	second := makeAlert(base.Add(time.Second), "unrelated-trigger-two")
	second.SrcIP = "10.0.0.3"
	if err := store.WriteBatch(context.Background(), []*model.Alert{first, second}); err != nil {
		t.Fatalf("write schema with unrelated table trigger: %v", err)
	}
	if count, err := store.Count(context.Background()); err != nil || count != 2 {
		t.Fatalf("unrelated trigger alert count=%d err=%v, want 2/nil", count, err)
	}
	if _, err := store.db.Exec(`INSERT INTO operator_data (id, value) VALUES (1, 'retained')`); err != nil {
		t.Fatalf("exercise unrelated table trigger: %v", err)
	}
	var value string
	if err := store.db.QueryRow(`SELECT value FROM operator_audit WHERE operator_id = 1`).Scan(&value); err != nil {
		t.Fatalf("read unrelated trigger output: %v", err)
	}
	if value != "retained" {
		t.Fatalf("unrelated trigger output = %q, want retained", value)
	}
}

func TestStoreRejectsWriteCriticalTriggerWithCaseVariantTableWithoutModification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	caseVariantSchema := strings.Replace(
		schemaSQL,
		"CREATE TABLE IF NOT EXISTS alerts (",
		"CREATE TABLE IF NOT EXISTS ALERTS (",
		1,
	)
	createSQLiteFixture(t, path, caseVariantSchema,
		`CREATE TRIGGER operator_alert_insert BEFORE INSERT ON ALERTS BEGIN SELECT RAISE(ABORT, 'blocked alert insert'); END`,
	)
	before := readFileBytes(t, path)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "WAL"})
	if store != nil {
		_ = store.Close()
		t.Fatal("case-variant write-critical trigger returned a usable store")
	}
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "trigger alerts.operator_alert_insert can alter or reject valid alert writes") {
		t.Fatalf("startup error = %v, want case-variant write-critical trigger", err)
	}
}

func TestStoreReopensCompatibleExistingDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "alerts.db")
	store, err := Open(ctx, Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("create compatible database: %v", err)
	}
	if err := store.WriteBatch(ctx, []*model.Alert{makeAlert(time.Now().UTC(), "compatible")}); err != nil {
		t.Fatalf("seed compatible database: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close compatible database: %v", err)
	}

	store, err = Open(ctx, Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("reopen compatible database: %v", err)
	}
	defer store.Close()
	if count, err := store.Count(ctx); err != nil || count != 1 {
		t.Fatalf("compatible database count=%d err=%v, want 1/nil", count, err)
	}
}

func TestStoreRecreatesOptionalQueryIndexOnCompatibleDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alerts.db")
	createSQLiteFixture(t, path, schemaSQL, `DROP INDEX idx_alerts_last_seen`)

	store, err := Open(context.Background(), Options{Path: path, JournalMode: "DELETE"})
	if err != nil {
		t.Fatalf("reopen compatible database without optional index: %v", err)
	}
	defer store.Close()
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_alerts_last_seen'`).Scan(&count); err != nil {
		t.Fatalf("inspect recreated query index: %v", err)
	}
	if count != 1 {
		t.Fatalf("recreated query index count = %d, want 1", count)
	}
}

func TestStoreRejectsIncompatibleHistoricalShardWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()

	historical := now.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historical.Format("2006-01-02")+".db")
	createSQLiteFixture(t, path, `CREATE TABLE operator_data (id INTEGER PRIMARY KEY)`)
	before := readFileBytes(t, path)

	err := store.WriteBatch(ctx, []*model.Alert{makeAlert(historical, "incompatible-historical-shard")})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
}

func TestStoreRejectsHistoricalShardWithUnknownRequiredColumnWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()

	historical := now.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historical.Format("2006-01-02")+".db")
	createSQLiteFixture(t, path, schemaSQL,
		`ALTER TABLE alerts ADD COLUMN operator_required TEXT NOT NULL`,
	)
	before := readFileBytes(t, path)

	err := store.WriteBatch(ctx, []*model.Alert{makeAlert(historical, "required-column-historical-shard")})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "unknown column alerts.operator_required is NOT NULL without a usable default") {
		t.Fatalf("historical shard error = %v, want unknown required column", err)
	}
}

func TestStoreRejectsHistoricalShardWithWriteBlockingUniqueIndexWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 7, 23, 4, 30, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()

	historical := now.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historical.Format("2006-01-02")+".db")
	createSQLiteFixture(t, path, schemaSQL,
		`CREATE UNIQUE INDEX operator_alert_rule_unique ON alerts(rule_id)`,
	)
	before := readFileBytes(t, path)

	err := store.WriteBatch(ctx, []*model.Alert{makeAlert(historical, "unique-index-historical-shard")})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "unique index alerts.operator_alert_rule_unique can reject valid alert writes") {
		t.Fatalf("historical shard error = %v, want write-blocking unique index", err)
	}
}

func TestStoreRejectsHistoricalShardWithWriteCriticalTriggerWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 7, 23, 4, 45, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()

	historical := now.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historical.Format("2006-01-02")+".db")
	createSQLiteFixture(t, path, schemaSQL,
		`CREATE TRIGGER operator_event_insert BEFORE INSERT ON alert_events BEGIN SELECT RAISE(ABORT, 'blocked event insert'); END`,
	)
	before := readFileBytes(t, path)

	err := store.WriteBatch(ctx, []*model.Alert{makeAlert(historical, "trigger-historical-shard")})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "trigger alert_events.operator_event_insert can alter or reject valid alert writes") {
		t.Fatalf("historical shard error = %v, want write-critical trigger", err)
	}
}

func TestStoreRejectsHistoricalShardWithGeneratedColumnWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 7, 23, 4, 55, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()

	historical := now.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historical.Format("2006-01-02")+".db")
	generatedSchema := strings.Replace(
		schemaSQL,
		"    updated_at TEXT NOT NULL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		"    updated_at TEXT NOT NULL,\n    operator_generated TEXT GENERATED ALWAYS AS (rule_id || ':' || src_ip) VIRTUAL,\n    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)",
		1,
	)
	createSQLiteFixture(t, path, generatedSchema)
	before := readFileBytes(t, path)

	err := store.WriteBatch(ctx, []*model.Alert{makeAlert(historical, "generated-column-historical-shard")})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "unknown generated column alerts.operator_generated can alter or reject valid alert writes") {
		t.Fatalf("historical shard error = %v, want generated column", err)
	}
}

func TestStoreRejectsHistoricalShardWithCheckConstraintWithoutModification(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 7, 24, 7, 10, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()

	historical := now.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historical.Format("2006-01-02")+".db")
	constrainedSchema := strings.Replace(
		schemaSQL,
		"    created_at TEXT NOT NULL\n);",
		"    created_at TEXT NOT NULL CHECK (length(created_at) > 0)\n);",
		1,
	)
	createSQLiteFixture(t, path, constrainedSchema)
	before := readFileBytes(t, path)

	err := store.WriteBatch(ctx, []*model.Alert{makeAlert(historical, "check-constraint-historical-shard")})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
	if !strings.Contains(err.Error(), "CHECK constraint on alert_events can reject valid alert writes") {
		t.Fatalf("historical shard error = %v, want CHECK constraint", err)
	}
}

func assertIntegrityRejectionPreservesFile(t *testing.T, path string, before []byte, err error) {
	t.Helper()
	if !errors.Is(err, ErrDatabaseIntegrity) {
		t.Fatalf("startup error = %v, want ErrDatabaseIntegrity", err)
	}
	if !strings.Contains(err.Error(), "file was not modified") {
		t.Fatalf("startup error does not state preservation: %v", err)
	}
	assertFileBytesUnchanged(t, path, before)
}

func assertFileBytesUnchanged(t *testing.T, path string, before []byte) {
	t.Helper()
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read rejected database: %v", readErr)
	}
	if !bytes.Equal(after, before) {
		t.Fatalf("rejected database changed: before=%d bytes after=%d bytes", len(before), len(after))
	}
}

func createSQLiteFixture(t *testing.T, path string, statements ...string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open SQLite fixture: %v", err)
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			_ = db.Close()
			t.Fatalf("create SQLite fixture: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close SQLite fixture: %v", err)
	}
}

func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func TestStoreWriteBatchHonorsCanceledContext(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := store.WriteBatch(canceled, []*model.Alert{makeAlert(time.Now().UTC(), "canceled")}); err == nil {
		t.Fatal("expected canceled context error")
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
	if health := store.Health(); health.Status != "ok" {
		t.Fatalf("canceled context should not degrade storage health: %+v", health)
	}
}

func TestStoreHealthMarksDegradedAndRecovers(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	store.markDegraded(os.ErrPermission)
	health := store.Health()
	if health.Status != "degraded" || health.LastError == "" || health.LastErrorAt.IsZero() {
		t.Fatalf("unexpected degraded health: %+v", health)
	}

	if err := store.WriteBatch(ctx, []*model.Alert{makeAlert(time.Now().UTC(), "recover")}); err != nil {
		t.Fatalf("write alerts: %v", err)
	}
	health = store.Health()
	if health.Status != "ok" || health.LastError != "" || !health.LastErrorAt.IsZero() {
		t.Fatalf("unexpected recovered health: %+v", health)
	}
}

func TestStoreClassifiesEmergencyStorageErrors(t *testing.T) {
	for _, err := range []error{
		syscall.ENOSPC,
		syscall.EDQUOT,
		syscall.EROFS,
		syscall.EIO,
		errors.New("constraint failed: SQLITE_FULL: database or disk is full"),
		errors.New("attempt to write a readonly database"),
		errors.New("disk I/O error"),
	} {
		if !isEmergencyStorageError(err) {
			t.Fatalf("expected emergency classification for %v", err)
		}
	}
	if isEmergencyStorageError(os.ErrPermission) {
		t.Fatal("plain permission errors should stay degraded, not emergency")
	}
}

func TestStoreEmergencyModePersistsRecoveryLogUntilRestart(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alerts.db")
	recoveryLogPath := filepath.Join(dir, "alerts-recovery.jsonl")
	now := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
	store, err := Open(ctx, Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	store.markStorageError(syscall.ENOSPC)
	if _, err := store.List(ctx); err != nil {
		t.Fatalf("list alerts while emergency: %v", err)
	}
	health := store.Health()
	if health.Status != "emergency" || health.LastError == "" || health.LastErrorAt.IsZero() {
		t.Fatalf("expected emergency health to persist after successful read: %+v", health)
	}
	if err := store.db.Close(); err != nil {
		t.Fatalf("close db before emergency write: %v", err)
	}

	err = store.WriteBatch(ctx, []*model.Alert{makeAlert(now, "pending-emergency")})
	if !errors.Is(err, ErrStorageEmergency) {
		t.Fatalf("write in emergency error = %v, want ErrStorageEmergency", err)
	}
	if info, err := os.Stat(recoveryLogPath); err != nil || info.Size() == 0 {
		t.Fatalf("expected pending alert in recovery log, info=%+v err=%v", info, err)
	}

	reopened, err := Open(ctx, Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer reopened.Close()
	listed, err := reopened.List(ctx)
	if err != nil {
		t.Fatalf("list replayed alerts: %v", err)
	}
	if len(listed) != 1 || listed[0].MatchedKeyword != "pending-emergency" {
		t.Fatalf("expected recovery replay after restart, got %+v", listed)
	}
	if info, err := os.Stat(recoveryLogPath); err != nil || info.Size() != 0 {
		t.Fatalf("recovery log should be truncated after restart replay, info=%+v err=%v", info, err)
	}
}

func TestStoreSeparatesAggregationWindows(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 60*time.Second)
	defer store.Close()

	base := time.Date(2026, 6, 27, 3, 0, 10, 0, time.UTC)
	if err := store.WriteBatch(ctx, []*model.Alert{
		makeAlert(base, "first"),
		makeAlert(base.Add(70*time.Second), "second"),
	}); err != nil {
		t.Fatalf("write alerts: %v", err)
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestStoreResolvesDailyShardPath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(ctx, Options{
		Dir:               dir,
		DailyShard:        true,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("open daily shard store: %v", err)
	}
	defer store.Close()

	want := filepath.Join(dir, "netsentry-2026-06-27.db")
	if store.Path() != want {
		t.Fatalf("store path = %q, want %q", store.Path(), want)
	}
	if err := store.WriteBatch(ctx, []*model.Alert{makeAlert(time.Date(2026, 6, 27, 12, 1, 0, 0, time.UTC), "daily")}); err != nil {
		t.Fatalf("write daily shard alert: %v", err)
	}
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count daily shard alerts: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestStoreRotatesDailyShardWritesAcrossAlertDates(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	firstDay := time.Date(2026, 6, 27, 23, 59, 0, 0, time.UTC)
	secondDay := time.Date(2026, 6, 28, 0, 1, 0, 0, time.UTC)
	store := openDailyShardStoreAt(t, dir, firstDay)
	defer store.Close()

	if err := store.WriteBatch(ctx, []*model.Alert{
		makeAlert(firstDay, "first-day"),
		makeAlert(secondDay, "second-day"),
	}); err != nil {
		t.Fatalf("write cross-day alerts: %v", err)
	}

	for _, date := range []string{"2026-06-27", "2026-06-28"} {
		path := filepath.Join(dir, "netsentry-"+date+".db")
		if !fileExists(path) {
			t.Fatalf("expected daily shard %s to exist", path)
		}
	}
	if store.Path() != filepath.Join(dir, "netsentry-2026-06-27.db") {
		t.Fatalf("store path changed after rotation: %q", store.Path())
	}

	alerts, total, err := store.Query(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatalf("query rotated daily shards: %v", err)
	}
	if total != 2 || len(alerts) != 2 {
		t.Fatalf("query returned total=%d len=%d, want 2/2", total, len(alerts))
	}
	if alerts[0].MatchedKeyword != "second-day" || alerts[1].MatchedKeyword != "first-day" {
		t.Fatalf("alerts not ordered across rotated shards: %+v", alerts)
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count rotated daily shards: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestStoreRejectsCorruptHistoricalShardWriteWithoutModification(t *testing.T) {
	dir := t.TempDir()
	currentDay := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	historicalDay := currentDay.AddDate(0, 0, -1)
	path := filepath.Join(dir, "netsentry-"+historicalDay.Format("2006-01-02")+".db")
	before := []byte("not a sqlite shard\x00corrupt bytes")
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}

	store := openDailyShardStoreAt(t, dir, currentDay)
	defer store.Close()
	err := store.WriteBatch(context.Background(), []*model.Alert{
		makeAlert(historicalDay, "corrupt-historical-shard"),
	})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
}

func TestStoreRejectsTruncatedHistoricalShardWriteWithoutModification(t *testing.T) {
	dir := t.TempDir()
	historicalDay := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	seed := openDailyShardStoreAt(t, dir, historicalDay)
	if err := seed.WriteBatch(context.Background(), []*model.Alert{
		makeAlert(historicalDay, "before-truncate"),
	}); err != nil {
		t.Fatalf("seed historical shard: %v", err)
	}
	if err := seed.Close(); err != nil {
		t.Fatalf("close historical shard: %v", err)
	}

	path := filepath.Join(dir, "netsentry-"+historicalDay.Format("2006-01-02")+".db")
	valid, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(valid) < 1024 {
		t.Fatalf("valid historical shard unexpectedly small: %d bytes", len(valid))
	}
	before := append([]byte(nil), valid[:len(valid)/2]...)
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}

	currentDay := historicalDay.AddDate(0, 0, 1)
	store := openDailyShardStoreAt(t, dir, currentDay)
	defer store.Close()
	err = store.WriteBatch(context.Background(), []*model.Alert{
		makeAlert(historicalDay, "truncated-historical-shard"),
	})
	assertIntegrityRejectionPreservesFile(t, path, before, err)
}

func TestStoreWritesExistingValidHistoricalShards(t *testing.T) {
	for _, test := range []struct {
		name string
		seed func(t *testing.T, dir string, day time.Time)
	}{
		{
			name: "empty",
			seed: func(t *testing.T, dir string, day time.Time) {
				t.Helper()
				path := filepath.Join(dir, "netsentry-"+day.Format("2006-01-02")+".db")
				if err := os.WriteFile(path, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "healthy",
			seed: func(t *testing.T, dir string, day time.Time) {
				t.Helper()
				writeDailyShardAlert(t, dir, day, makeAlert(day, "existing-healthy"))
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			historicalDay := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
			test.seed(t, dir, historicalDay)

			store := openDailyShardStoreAt(t, dir, historicalDay.AddDate(0, 0, 1))
			defer store.Close()
			if err := store.WriteBatch(context.Background(), []*model.Alert{
				makeAlert(historicalDay.Add(time.Minute), "new-historical-write"),
			}); err != nil {
				t.Fatalf("write existing %s historical shard: %v", test.name, err)
			}
		})
	}
}

func TestStoreMalformedHistoricalShardReadsPreserveFile(t *testing.T) {
	operations := []struct {
		name string
		run  func(context.Context, *Store) error
	}{
		{
			name: "query",
			run: func(ctx context.Context, store *Store) error {
				_, _, err := store.Query(ctx, Query{Limit: 10})
				return err
			},
		},
		{
			name: "count",
			run: func(ctx context.Context, store *Store) error {
				_, err := store.Count(ctx)
				return err
			},
		},
	}
	for _, fixture := range []string{"corrupt", "truncated"} {
		for _, operation := range operations {
			t.Run(fixture+"/"+operation.name, func(t *testing.T) {
				dir, path, before, currentDay := malformedHistoricalShard(t, fixture)
				store := openDailyShardStoreAt(t, dir, currentDay)
				defer store.Close()

				if err := operation.run(context.Background(), store); err == nil {
					t.Fatalf("%s malformed historical shard unexpectedly passed %s", fixture, operation.name)
				}
				assertFileBytesUnchanged(t, path, before)
			})
		}
	}
}

func TestStoreReadsActiveWALHistoricalShardReadOnly(t *testing.T) {
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "daily shards")
	historicalDay := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	historical := openDailyShardStoreAt(t, dir, historicalDay)
	defer historical.Close()
	if err := historical.WriteBatch(ctx, []*model.Alert{
		makeAlert(historicalDay, "active-wal-historical"),
	}); err != nil {
		t.Fatalf("write active WAL historical shard: %v", err)
	}

	current := openDailyShardStoreAt(t, dir, historicalDay.AddDate(0, 0, 1))
	defer current.Close()
	alerts, total, err := current.Query(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatalf("query active WAL historical shard: %v", err)
	}
	if total != 1 || len(alerts) != 1 || alerts[0].MatchedKeyword != "active-wal-historical" {
		t.Fatalf("unexpected active WAL query result: total=%d alerts=%+v", total, alerts)
	}
	if count, err := current.Count(ctx); err != nil || count != 1 {
		t.Fatalf("active WAL count=%d err=%v, want 1/nil", count, err)
	}
}

func malformedHistoricalShard(t *testing.T, kind string) (string, string, []byte, time.Time) {
	t.Helper()
	dir := t.TempDir()
	historicalDay := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(dir, "netsentry-"+historicalDay.Format("2006-01-02")+".db")
	var before []byte
	switch kind {
	case "corrupt":
		before = []byte("not a sqlite shard\x00corrupt read fixture")
	case "truncated":
		seed := openDailyShardStoreAt(t, dir, historicalDay)
		if err := seed.WriteBatch(context.Background(), []*model.Alert{
			makeAlert(historicalDay, "before-read-truncate"),
		}); err != nil {
			t.Fatalf("seed historical read shard: %v", err)
		}
		if err := seed.Close(); err != nil {
			t.Fatalf("close historical read shard: %v", err)
		}
		valid, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(valid) < 1024 {
			t.Fatalf("valid historical read shard unexpectedly small: %d bytes", len(valid))
		}
		before = append([]byte(nil), valid[:len(valid)/2]...)
	default:
		t.Fatalf("unknown malformed shard fixture %q", kind)
	}
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, path, before, historicalDay.AddDate(0, 0, 1)
}

func TestStoreQueriesAcrossDailyShards(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	firstDay := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	secondDay := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	thirdDay := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)

	writeDailyShardAlert(t, dir, firstDay, makeAlert(firstDay, "first-day"))
	writeDailyShardAlert(t, dir, secondDay, makeAlert(secondDay, "second-day"))
	writeDailyShardAlert(t, dir, thirdDay, makeAlert(thirdDay, "third-day"))

	store := openDailyShardStoreAt(t, dir, thirdDay)
	defer store.Close()

	alerts, total, err := store.Query(ctx, Query{Limit: 2})
	if err != nil {
		t.Fatalf("query across daily shards: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(alerts) != 2 {
		t.Fatalf("len(alerts) = %d, want 2", len(alerts))
	}
	if alerts[0].MatchedKeyword != "third-day" || alerts[1].MatchedKeyword != "second-day" {
		t.Fatalf("alerts not ordered across shards: %+v", alerts)
	}

	alerts, total, err = store.Query(ctx, Query{Limit: 1, Offset: 2})
	if err != nil {
		t.Fatalf("page across daily shards: %v", err)
	}
	if total != 3 || len(alerts) != 1 || alerts[0].MatchedKeyword != "first-day" {
		t.Fatalf("paged query returned total=%d alerts=%+v, want first-day", total, alerts)
	}
}

func TestStoreCountsAcrossDailyShards(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	firstDay := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	secondDay := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	thirdDay := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)

	writeDailyShardAlert(t, dir, firstDay, makeAlert(firstDay, "first-day"))
	writeDailyShardAlert(t, dir, secondDay, makeAlert(secondDay, "second-day"))
	writeDailyShardAlert(t, dir, thirdDay, makeAlert(thirdDay, "third-day"))

	store := openDailyShardStoreAt(t, dir, thirdDay)
	defer store.Close()

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count across daily shards: %v", err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestStoreDailyShardQueryHonorsTimeRange(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	firstDay := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	secondDay := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	thirdDay := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)

	writeDailyShardAlert(t, dir, firstDay, makeAlert(firstDay, "first-day"))
	writeDailyShardAlert(t, dir, secondDay, makeAlert(secondDay, "second-day"))
	writeDailyShardAlert(t, dir, thirdDay, makeAlert(thirdDay, "third-day"))

	store := openDailyShardStoreAt(t, dir, thirdDay)
	defer store.Close()

	since := secondDay.Add(-time.Hour)
	until := secondDay.Add(time.Hour)
	alerts, total, err := store.Query(ctx, Query{Since: &since, Until: &until, Limit: 10})
	if err != nil {
		t.Fatalf("query daily shard time range: %v", err)
	}
	if total != 1 || len(alerts) != 1 {
		t.Fatalf("time range returned total=%d len=%d, want 1/1", total, len(alerts))
	}
	if alerts[0].MatchedKeyword != "second-day" {
		t.Fatalf("matched keyword = %q, want second-day", alerts[0].MatchedKeyword)
	}
}

func TestStorePrunesExpiredAlerts(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	store, err := Open(ctx, Options{
		Path:              filepath.Join(t.TempDir(), "alerts.db"),
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		RetentionDays:     7,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.WriteBatch(ctx, []*model.Alert{
		makeAlert(now.AddDate(0, 0, -8), "expired"),
		makeAlert(now.AddDate(0, 0, -6), "fresh"),
	}); err != nil {
		t.Fatalf("write alerts: %v", err)
	}

	pruned, err := store.PruneExpired(ctx)
	if err != nil {
		t.Fatalf("prune expired alerts: %v", err)
	}
	if pruned != 1 {
		t.Fatalf("pruned = %d, want 1", pruned)
	}
	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(listed) != 1 || listed[0].MatchedKeyword != "fresh" {
		t.Fatalf("expected only fresh alert, got %+v", listed)
	}
}

func TestStorePrunesExpiredShardFiles(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	store, err := Open(ctx, Options{
		Dir:               dir,
		DailyShard:        true,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		RetentionDays:     7,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("open daily shard store: %v", err)
	}
	defer store.Close()

	oldBase := filepath.Join(dir, "netsentry-2026-06-19.db")
	for _, path := range []string{oldBase, oldBase + "-wal", oldBase + "-shm"} {
		if err := touchFile(path); err != nil {
			t.Fatalf("touch old shard file %s: %v", path, err)
		}
	}
	fresh := filepath.Join(dir, "netsentry-2026-06-21.db")
	unrelated := filepath.Join(dir, "notes-2026-06-19.db")
	for _, path := range []string{fresh, unrelated} {
		if err := touchFile(path); err != nil {
			t.Fatalf("touch retained file %s: %v", path, err)
		}
	}

	deleted, err := store.PruneExpiredShardFiles(ctx, dir)
	if err != nil {
		t.Fatalf("prune expired shard files: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("deleted = %d, want 3", deleted)
	}
	for _, path := range []string{oldBase, oldBase + "-wal", oldBase + "-shm"} {
		if fileExists(path) {
			t.Fatalf("expected %s to be removed", path)
		}
	}
	for _, path := range []string{store.Path(), fresh, unrelated} {
		if !fileExists(path) {
			t.Fatalf("expected %s to remain", path)
		}
	}
}

func touchFile(path string) error {
	return os.WriteFile(path, []byte("x"), 0o600)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func openTestStore(t *testing.T, window time.Duration) *Store {
	t.Helper()
	store, err := Open(context.Background(), Options{
		Path:              filepath.Join(t.TempDir(), "alerts.db"),
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: window,
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s changed: got %d bytes, want %d", path, len(got), len(want))
	}
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s exists after rejected recovery preflight, stat error=%v", path, err)
	}
}

func assertSQLiteIndexMissing(t *testing.T, path, index string) {
	t.Helper()
	db, err := openReadOnlyDatabase(path)
	if err != nil {
		t.Fatalf("open preserved database read-only: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?`, index).Scan(&count); err != nil {
		t.Fatalf("inspect preserved database index: %v", err)
	}
	if count != 0 {
		t.Fatalf("optional index %s was created before recovery rejection", index)
	}
}

func assertSQLiteAlertCount(t *testing.T, path string, want int) {
	t.Helper()
	db, err := openReadOnlyDatabase(path)
	if err != nil {
		t.Fatalf("open preserved database read-only: %v", err)
	}
	defer db.Close()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM alerts`).Scan(&got); err != nil {
		t.Fatalf("count preserved database alerts: %v", err)
	}
	if got != want {
		t.Fatalf("preserved database alert count = %d, want %d", got, want)
	}
}

func assertNoPersistedRecoveryPrefix(t *testing.T, dbPath, recoveryLogPath string, now time.Time) {
	t.Helper()
	store, err := Open(context.Background(), Options{
		Path:              dbPath,
		RecoveryLogPath:   recoveryLogPath,
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("reopen store to inspect partial replay: %v", err)
	}
	defer store.Close()
	count, err := store.Count(context.Background())
	if err != nil {
		t.Fatalf("count alerts after rejected recovery log: %v", err)
	}
	if count != 0 {
		t.Fatalf("rejected recovery log persisted %d alert rows", count)
	}
}

func openDailyShardStoreAt(t *testing.T, dir string, now time.Time) *Store {
	t.Helper()
	store, err := Open(context.Background(), Options{
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
		t.Fatalf("open daily shard store at %s: %v", now, err)
	}
	return store
}

func writeDailyShardAlert(t *testing.T, dir string, now time.Time, alert *model.Alert) {
	t.Helper()
	store := openDailyShardStoreAt(t, dir, now)
	defer store.Close()
	if err := store.WriteBatch(context.Background(), []*model.Alert{alert}); err != nil {
		t.Fatalf("write daily shard alert at %s: %v", now, err)
	}
}

func makeAlert(ts time.Time, keyword string) *model.Alert {
	return &model.Alert{
		RuleID:             "rule-1",
		RuleName:           "Test Rule",
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
		MatchedKeyword:     keyword,
	}
}

func makeRecoveryAlertWithEncodedSize(t *testing.T, size int, ts time.Time) *model.Alert {
	t.Helper()
	alert := normalizeAlert(makeAlert(ts, "large-record"), ts, time.Minute)
	alert.RuleName = ""
	base, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("marshal recovery record base: %v", err)
	}
	if size < len(base) {
		t.Fatalf("requested recovery record size %d is smaller than base %d", size, len(base))
	}
	alert.RuleName = strings.Repeat("r", size-len(base))
	encoded, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("marshal sized recovery record: %v", err)
	}
	if len(encoded) != size {
		t.Fatalf("encoded recovery record size = %d, want %d", len(encoded), size)
	}
	return &alert
}
