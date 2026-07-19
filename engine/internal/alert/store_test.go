package alert

import (
	"bytes"
	"context"
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
	assertNoPersistedRecoveryPrefix(t, dbPath, filepath.Join(dir, "empty-recovery.jsonl"), now)
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
