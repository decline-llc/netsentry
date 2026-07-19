package alert

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	"github.com/decline-llc/netsentry/pkg/model"
)

var ErrStorageEmergency = errors.New("alert storage is in emergency mode; clear disk/storage fault and restart NetSentry")

// ErrRecoveryLogIntegrity reports that durable recovery input is malformed or
// truncated and was left unchanged for operator-led recovery.
var ErrRecoveryLogIntegrity = errors.New("alert recovery log failed integrity check; file was not modified")

// ErrDatabaseIntegrity reports that an existing database failed a read-only
// integrity preflight and was not opened for writable initialization.
var ErrDatabaseIntegrity = errors.New("existing SQLite database failed integrity check; file was not modified")

// Options controls the SQLite alert store.
type Options struct {
	Path              string
	Dir               string
	DailyShard        bool
	JournalMode       string
	BusyTimeoutMS     int
	AggregationWindow time.Duration
	RetentionDays     int
	RecoveryLogPath   string
	Now               func() time.Time
}

// Store persists alerts and aggregates repeated hits in a fixed time window.
type Store struct {
	db                *sql.DB
	path              string
	dir               string
	dailyShard        bool
	journalMode       string
	busyTimeoutMS     int
	aggregationWindow time.Duration
	retentionDays     int
	recoveryLogPath   string
	now               func() time.Time
	writeMu           sync.Mutex
	healthMu          sync.RWMutex
	health            StorageHealth
}

// StorageHealth describes the current alert storage state.
type StorageHealth struct {
	Status      string    `json:"status"`
	LastError   string    `json:"last_error,omitempty"`
	LastErrorAt time.Time `json:"last_error_at,omitempty"`
}

// Query filters, counts, and pages aggregated alert rows.
type Query struct {
	RuleID           string
	Severity         model.Severity
	SrcIP            string
	DstIP            string
	Protocol         string
	DstPort          *uint16
	Since            *time.Time
	Until            *time.Time
	MitreTactic      string
	MitreTechniqueID string
	MatchedKeyword   string
	MinCount         *int
	Limit            int
	Offset           int
}

// Open creates the SQLite database and initializes its schema.
func Open(ctx context.Context, opts Options) (*Store, error) {
	if opts.JournalMode == "" {
		opts.JournalMode = "WAL"
	}
	if opts.BusyTimeoutMS <= 0 {
		opts.BusyTimeoutMS = 5000
	}
	if opts.AggregationWindow <= 0 {
		opts.AggregationWindow = time.Minute
	}

	dbPath := resolveDBPath(opts)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return nil, fmt.Errorf("create sqlite alert dir: %w", err)
	}
	existingNonEmpty, err := existingNonEmptyDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	if existingNonEmpty {
		if err := validateExistingDatabase(ctx, dbPath); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite alerts store: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{
		db:                db,
		path:              dbPath,
		dir:               defaultDBDir(opts.Dir),
		dailyShard:        opts.DailyShard,
		journalMode:       opts.JournalMode,
		busyTimeoutMS:     opts.BusyTimeoutMS,
		aggregationWindow: opts.AggregationWindow,
		retentionDays:     opts.RetentionDays,
		recoveryLogPath:   resolveRecoveryLogPath(opts, dbPath),
		now:               clock(opts.Now),
		health:            StorageHealth{Status: "ok"},
	}
	if err := store.init(ctx, opts); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ReplayRecoveryLog(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := store.PruneExpired(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := store.PruneExpiredShardFiles(ctx, defaultDBDir(opts.Dir)); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func existingNonEmptyDatabase(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect sqlite alerts database: %w", err)
	}
	return info.Mode().IsRegular() && info.Size() > 0, nil
}

func validateExistingDatabase(ctx context.Context, path string) error {
	db, err := openReadOnlyDatabase(path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseIntegrity, err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "PRAGMA quick_check")
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %v", ErrDatabaseIntegrity, err)
	}
	defer rows.Close()

	checked := false
	for rows.Next() {
		var result string
		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("%w: read quick_check result: %v", ErrDatabaseIntegrity, err)
		}
		checked = true
		if !strings.EqualFold(strings.TrimSpace(result), "ok") {
			return fmt.Errorf("%w: %s", ErrDatabaseIntegrity, strings.TrimSpace(result))
		}
	}
	if err := rows.Err(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %v", ErrDatabaseIntegrity, err)
	}
	if !checked {
		return fmt.Errorf("%w: quick_check returned no result", ErrDatabaseIntegrity)
	}
	return nil
}

func openReadOnlyDatabase(path string) (*sql.DB, error) {
	dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func resolveDBPath(opts Options) string {
	if !opts.DailyShard {
		if opts.Path != "" {
			return opts.Path
		}
		return filepath.Join(defaultDBDir(opts.Dir), "netsentry.db")
	}

	now := time.Now().UTC()
	if opts.Now != nil {
		now = opts.Now().UTC()
	}
	return filepath.Join(defaultDBDir(opts.Dir), fmt.Sprintf("netsentry-%s.db", now.Format("2006-01-02")))
}

func resolveRecoveryLogPath(opts Options, dbPath string) string {
	if opts.RecoveryLogPath != "" {
		return opts.RecoveryLogPath
	}
	if opts.DailyShard {
		return filepath.Join(defaultDBDir(opts.Dir), "netsentry-alerts-recovery.jsonl")
	}
	return dbPath + ".alerts.jsonl"
}

func defaultDBDir(dir string) string {
	if dir == "" {
		return "data"
	}
	return dir
}

func clock(now func() time.Time) func() time.Time {
	if now != nil {
		return func() time.Time { return now().UTC() }
	}
	return func() time.Time { return time.Now().UTC() }
}

// Path returns the concrete SQLite database path in use.
func (s *Store) Path() string { return s.path }

// Health returns the latest known alert storage status.
func (s *Store) Health() StorageHealth {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()
	return s.health
}

func (s *Store) markHealthy() {
	s.healthMu.Lock()
	defer s.healthMu.Unlock()
	if s.health.Status == "emergency" {
		return
	}
	s.health = StorageHealth{Status: "ok"}
}

func (s *Store) markDegraded(err error) {
	if err == nil {
		return
	}
	s.healthMu.Lock()
	defer s.healthMu.Unlock()
	s.health = StorageHealth{
		Status:      "degraded",
		LastError:   err.Error(),
		LastErrorAt: s.now(),
	}
}

func (s *Store) markStorageError(err error) {
	if err == nil {
		return
	}
	if isEmergencyStorageError(err) {
		s.markEmergency(err)
		return
	}
	s.markDegraded(err)
}

func (s *Store) markEmergency(err error) {
	if err == nil {
		return
	}
	s.healthMu.Lock()
	defer s.healthMu.Unlock()
	s.health = StorageHealth{
		Status:      "emergency",
		LastError:   err.Error(),
		LastErrorAt: s.now(),
	}
}

func (s *Store) emergencyErr() error {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()
	if s.health.Status != "emergency" {
		return nil
	}
	if s.health.LastError == "" {
		return ErrStorageEmergency
	}
	return fmt.Errorf("%w: %s", ErrStorageEmergency, s.health.LastError)
}

func isEmergencyStorageError(err error) bool {
	for _, target := range []error{
		syscall.ENOSPC,
		syscall.EDQUOT,
		syscall.EROFS,
		syscall.EIO,
	} {
		if errors.Is(err, target) {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range []string{
		"sql logic error: database or disk is full",
		"database or disk is full",
		"sqlite_full",
		"readonly database",
		"attempt to write a readonly database",
		"read-only file system",
		"disk i/o error",
		"sqlite_ioerr",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func (s *Store) init(ctx context.Context, opts Options) error {
	journalMode := strings.ToUpper(strings.TrimSpace(opts.JournalMode))
	if journalMode == "" {
		journalMode = "WAL"
	}
	switch journalMode {
	case "DELETE", "TRUNCATE", "PERSIST", "MEMORY", "WAL", "OFF":
	default:
		return fmt.Errorf("unsupported sqlite journal mode %q", opts.JournalMode)
	}
	if _, err := s.db.ExecContext(ctx, "PRAGMA journal_mode="+journalMode); err != nil {
		return fmt.Errorf("set sqlite journal mode: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout=%d", opts.BusyTimeoutMS)); err != nil {
		return fmt.Errorf("set sqlite busy timeout: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("init sqlite alerts schema: %w", err)
	}
	return nil
}

func (s *Store) openShard(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create sqlite alert shard dir: %w", err)
	}
	existingNonEmpty, err := existingNonEmptyDatabase(path)
	if err != nil {
		return nil, err
	}
	if existingNonEmpty {
		if err := validateExistingDatabase(ctx, path); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite alert shard: %w", err)
	}
	db.SetMaxOpenConns(1)
	opts := Options{
		JournalMode:   s.journalMode,
		BusyTimeoutMS: s.busyTimeoutMS,
	}
	shard := &Store{db: db}
	if err := shard.init(ctx, opts); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (s *Store) shardPathFor(ts time.Time) string {
	return filepath.Join(s.dir, fmt.Sprintf("netsentry-%s.db", ts.UTC().Format("2006-01-02")))
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS alerts (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    rule_id TEXT NOT NULL,
    rule_name TEXT NOT NULL,
    severity TEXT NOT NULL,
    protocol TEXT NOT NULL,
    src_ip TEXT NOT NULL,
    dst_ip TEXT NOT NULL,
    dst_port INTEGER NOT NULL,
    mitre_tactic TEXT NOT NULL,
    mitre_technique_id TEXT NOT NULL,
    mitre_technique_name TEXT NOT NULL,
    payload_preview TEXT NOT NULL,
    matched_keyword TEXT NOT NULL,
    aggregated_count INTEGER NOT NULL,
    first_seen TEXT NOT NULL,
    last_seen TEXT NOT NULL,
    window_start TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(rule_id, src_ip, dst_ip, dst_port, window_start)
);

CREATE TABLE IF NOT EXISTS alert_events (
    event_id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_alerts_last_seen ON alerts(last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_window ON alerts(rule_id, window_start);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_last_seen ON alerts(rule_id, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_severity_last_seen ON alerts(severity, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_src_last_seen ON alerts(src_ip, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_dst_last_seen ON alerts(dst_ip, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_protocol_port_last_seen ON alerts(protocol, dst_port, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_mitre_technique_last_seen ON alerts(mitre_technique_id, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_count_last_seen ON alerts(aggregated_count, last_seen DESC);
`

const upsertAlertSQL = `
INSERT INTO alerts (
    id, event_id, rule_id, rule_name, severity, protocol, src_ip, dst_ip, dst_port,
    mitre_tactic, mitre_technique_id, mitre_technique_name,
    payload_preview, matched_keyword,
    aggregated_count, first_seen, last_seen, window_start, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(rule_id, src_ip, dst_ip, dst_port, window_start) DO UPDATE SET
    aggregated_count = alerts.aggregated_count + excluded.aggregated_count,
    first_seen = CASE
        WHEN excluded.first_seen < alerts.first_seen THEN excluded.first_seen
        ELSE alerts.first_seen
    END,
    last_seen = CASE
        WHEN excluded.last_seen > alerts.last_seen THEN excluded.last_seen
        ELSE alerts.last_seen
    END,
    payload_preview = CASE
        WHEN excluded.last_seen >= alerts.last_seen THEN excluded.payload_preview
        ELSE alerts.payload_preview
    END,
    matched_keyword = CASE
        WHEN excluded.last_seen >= alerts.last_seen THEN excluded.matched_keyword
        ELSE alerts.matched_keyword
    END,
    updated_at = excluded.updated_at;
`

const insertAlertEventSQL = `
INSERT OR IGNORE INTO alert_events (event_id, created_at) VALUES (?, ?)
`

// WriteBatch inserts alerts and aggregates repeats in the configured window.
func (s *Store) WriteBatch(ctx context.Context, alerts []*model.Alert) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(alerts) == 0 {
		return nil
	}
	now := s.now()
	normalized := normalizeAlerts(alerts, now, s.aggregationWindow)
	if len(normalized) == 0 {
		return nil
	}
	if err := s.appendRecoveryLog(normalized); err != nil {
		s.markStorageError(err)
		return err
	}
	pending, err := s.readRecoveryLog()
	if err != nil {
		s.markStorageError(err)
		return err
	}
	if err := s.emergencyErr(); err != nil {
		return err
	}
	if !s.dailyShard {
		if err := s.writeBatchToDB(ctx, s.db, pending, now); err != nil {
			return err
		}
		if err := s.truncateRecoveryLog(); err != nil {
			s.markStorageError(err)
			return err
		}
		s.markHealthy()
		return nil
	}

	byPath := map[string][]*model.Alert{}
	for _, alert := range pending {
		path := s.shardPathFor(alert.Timestamp)
		byPath[path] = append(byPath[path], alert)
	}
	for path, shardAlerts := range byPath {
		if err := ctx.Err(); err != nil {
			return err
		}
		db := s.db
		closeDB := false
		if path != s.path {
			var err error
			db, err = s.openShard(ctx, path)
			if err != nil {
				s.markStorageError(err)
				return fmt.Errorf("open alert shard %s: %w", path, err)
			}
			closeDB = true
		}
		err := s.writeBatchToDB(ctx, db, shardAlerts, now)
		if closeDB {
			if closeErr := db.Close(); err == nil && closeErr != nil {
				err = fmt.Errorf("close alert shard %s: %w", path, closeErr)
			}
		}
		if err != nil {
			s.markStorageError(err)
			return fmt.Errorf("write alert shard %s: %w", path, err)
		}
	}
	if err := s.truncateRecoveryLog(); err != nil {
		s.markStorageError(err)
		return err
	}
	s.markHealthy()
	return nil
}

func (s *Store) writeBatchToDB(ctx context.Context, db *sql.DB, alerts []*model.Alert, now time.Time) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		s.markStorageError(err)
		return fmt.Errorf("begin alert transaction: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, upsertAlertSQL)
	if err != nil {
		_ = tx.Rollback()
		s.markStorageError(err)
		return fmt.Errorf("prepare alert upsert: %w", err)
	}
	defer stmt.Close()
	eventStmt, err := tx.PrepareContext(ctx, insertAlertEventSQL)
	if err != nil {
		_ = tx.Rollback()
		s.markStorageError(err)
		return fmt.Errorf("prepare alert event insert: %w", err)
	}
	defer eventStmt.Close()

	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		normalized := normalizeAlert(alert, now, s.aggregationWindow)
		result, err := eventStmt.ExecContext(ctx, normalized.EventID, formatTime(now))
		if err != nil {
			_ = tx.Rollback()
			s.markStorageError(err)
			return fmt.Errorf("insert alert event %s: %w", normalized.EventID, err)
		}
		inserted, err := result.RowsAffected()
		if err != nil {
			_ = tx.Rollback()
			s.markStorageError(err)
			return fmt.Errorf("check alert event insert %s: %w", normalized.EventID, err)
		}
		if inserted == 0 {
			continue
		}
		if _, err := stmt.ExecContext(ctx,
			normalized.ID,
			normalized.EventID,
			normalized.RuleID,
			normalized.RuleName,
			string(normalized.Severity),
			normalized.Protocol,
			normalized.SrcIP,
			normalized.DstIP,
			int(normalized.DstPort),
			normalized.MitreTactic,
			normalized.MitreTechniqueID,
			normalized.MitreTechniqueName,
			normalized.PayloadPreview,
			normalized.MatchedKeyword,
			normalized.AggregatedCount,
			formatTime(normalized.FirstSeen),
			formatTime(normalized.LastSeen),
			formatTime(normalized.WindowStart),
			formatTime(now),
			formatTime(now),
		); err != nil {
			_ = tx.Rollback()
			s.markStorageError(err)
			return fmt.Errorf("upsert alert %s: %w", normalized.RuleID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		s.markStorageError(err)
		return fmt.Errorf("commit alert transaction: %w", err)
	}
	return nil
}

func normalizeAlerts(alerts []*model.Alert, now time.Time, window time.Duration) []*model.Alert {
	normalized := make([]*model.Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		next := normalizeAlert(alert, now, window)
		normalized = append(normalized, &next)
	}
	return normalized
}

func normalizeAlert(alert *model.Alert, now time.Time, window time.Duration) model.Alert {
	out := *alert
	if out.Timestamp.IsZero() {
		out.Timestamp = now
	}
	out.Timestamp = out.Timestamp.UTC()
	out.WindowStart = out.Timestamp.Truncate(window)
	out.FirstSeen = out.Timestamp
	out.LastSeen = out.Timestamp
	out.AggregatedCount = 1
	out.ID = fmt.Sprintf("%s-%s-%s-%d-%d", out.RuleID, out.SrcIP, out.DstIP, out.DstPort, out.WindowStart.Unix())
	if strings.TrimSpace(out.EventID) == "" {
		out.EventID = alertEventID(out)
	}
	return out
}

func alertEventID(alert model.Alert) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s\x00%s\x00%d",
		alert.RuleID,
		alert.SrcIP,
		alert.DstIP,
		alert.DstPort,
		alert.Protocol,
		alert.MatchedKeyword,
		alert.Timestamp.Format(time.RFC3339Nano),
		alert.Timestamp.UnixNano(),
	)))
	return "evt_" + hex.EncodeToString(sum[:16])
}

func (s *Store) appendRecoveryLog(alerts []*model.Alert) error {
	if s.recoveryLogPath == "" || len(alerts) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.recoveryLogPath), 0o750); err != nil {
		return fmt.Errorf("create alert recovery log dir: %w", err)
	}
	file, err := os.OpenFile(s.recoveryLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open alert recovery log: %w", err)
	}
	enc := json.NewEncoder(file)
	for _, alert := range alerts {
		if err := enc.Encode(alert); err != nil {
			_ = file.Close()
			return fmt.Errorf("write alert recovery log: %w", err)
		}
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync alert recovery log: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close alert recovery log: %w", err)
	}
	return nil
}

func (s *Store) readRecoveryLog() ([]*model.Alert, error) {
	if s.recoveryLogPath == "" {
		return nil, nil
	}
	file, err := os.Open(s.recoveryLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open alert recovery log: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect alert recovery log: %w", err)
	}
	if info.Size() > 0 {
		if _, err := file.Seek(-1, io.SeekEnd); err != nil {
			return nil, fmt.Errorf("inspect alert recovery log terminator: %w", err)
		}
		var last [1]byte
		if _, err := file.Read(last[:]); err != nil {
			return nil, fmt.Errorf("inspect alert recovery log terminator: %w", err)
		}
		if last[0] != '\n' {
			return nil, fmt.Errorf("%w: truncated final JSONL record", ErrRecoveryLogIntegrity)
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("rewind alert recovery log: %w", err)
		}
	}

	var alerts []*model.Alert
	scanner := bufio.NewScanner(file)
	record := 0
	for scanner.Scan() {
		record++
		var alert model.Alert
		if err := json.Unmarshal(scanner.Bytes(), &alert); err != nil {
			return nil, fmt.Errorf("%w: decode record %d: %v", ErrRecoveryLogIntegrity, record, err)
		}
		alerts = append(alerts, &alert)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: read records: %v", ErrRecoveryLogIntegrity, err)
	}
	return alerts, nil
}

// ReplayRecoveryLog restores alert writes that were durably logged before a
// previous process exited. Existing event IDs are skipped, so replay is safe to
// repeat after SQLite committed but the recovery log was not truncated.
func (s *Store) ReplayRecoveryLog(ctx context.Context) error {
	alerts, err := s.readRecoveryLog()
	if err != nil {
		return err
	}
	if len(alerts) == 0 {
		return s.truncateRecoveryLog()
	}
	now := s.now()
	if !s.dailyShard {
		if err := s.writeBatchToDB(ctx, s.db, alerts, now); err != nil {
			return err
		}
		return s.truncateRecoveryLog()
	}
	byPath := map[string][]*model.Alert{}
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		ts := alert.Timestamp
		if ts.IsZero() {
			ts = now
		}
		path := s.shardPathFor(ts)
		byPath[path] = append(byPath[path], alert)
	}
	for path, shardAlerts := range byPath {
		db := s.db
		closeDB := false
		if path != s.path {
			var err error
			db, err = s.openShard(ctx, path)
			if err != nil {
				return fmt.Errorf("open alert shard %s for recovery replay: %w", path, err)
			}
			closeDB = true
		}
		err := s.writeBatchToDB(ctx, db, shardAlerts, now)
		if closeDB {
			if closeErr := db.Close(); err == nil && closeErr != nil {
				err = fmt.Errorf("close alert shard %s after recovery replay: %w", path, closeErr)
			}
		}
		if err != nil {
			return fmt.Errorf("replay alert shard %s: %w", path, err)
		}
	}
	return s.truncateRecoveryLog()
}

func (s *Store) truncateRecoveryLog() error {
	if s.recoveryLogPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.recoveryLogPath), 0o750); err != nil {
		return fmt.Errorf("create alert recovery log dir: %w", err)
	}
	file, err := os.OpenFile(s.recoveryLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("truncate alert recovery log: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close truncated alert recovery log: %w", err)
	}
	return nil
}

// List returns alerts ordered by most recent activity.
func (s *Store) List(ctx context.Context) ([]*model.Alert, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, event_id, rule_id, rule_name, severity, protocol, src_ip, dst_ip, dst_port,
       mitre_tactic, mitre_technique_id, mitre_technique_name,
       payload_preview, matched_keyword,
       aggregated_count, first_seen, last_seen, window_start
FROM alerts
ORDER BY last_seen DESC, id ASC
LIMIT 1000`)
	if err != nil {
		s.markStorageError(err)
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*model.Alert
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		s.markStorageError(err)
		return nil, fmt.Errorf("iterate alerts: %w", err)
	}
	s.markHealthy()
	return alerts, nil
}

const alertSelectColumns = `
SELECT id, event_id, rule_id, rule_name, severity, protocol, src_ip, dst_ip, dst_port,
       mitre_tactic, mitre_technique_id, mitre_technique_name,
       payload_preview, matched_keyword,
       aggregated_count, first_seen, last_seen, window_start
FROM alerts`

// Query returns filtered and paginated alerts plus the total filtered row count.
func (s *Store) Query(ctx context.Context, query Query) ([]*model.Alert, int, error) {
	if s.dailyShard {
		return s.queryDailyShards(ctx, query)
	}
	alerts, total, err := queryAlertsDB(ctx, s.db, query)
	if err != nil {
		s.markStorageError(err)
		return nil, 0, err
	}
	s.markHealthy()
	return alerts, total, nil
}

func queryAlertsDB(ctx context.Context, db *sql.DB, query Query) ([]*model.Alert, int, error) {
	where, args := alertQueryWhere(query)
	countSQL := "SELECT COUNT(*) FROM alerts" + where
	var total int
	if err := db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count filtered alerts: %w", err)
	}

	limit := query.Limit
	if limit < 0 {
		limit = total
	} else if limit <= 0 {
		limit = 1000
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	rows, err := db.QueryContext(ctx, alertSelectColumns+where+`
ORDER BY last_seen DESC, id ASC
LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*model.Alert
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate queried alerts: %w", err)
	}
	return alerts, total, nil
}

func (s *Store) queryDailyShards(ctx context.Context, query Query) ([]*model.Alert, int, error) {
	paths, err := s.alertShardPaths(ctx, query)
	if err != nil {
		s.markStorageError(err)
		return nil, 0, err
	}

	var all []*model.Alert
	total := 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		shardQuery := query
		shardQuery.Limit = -1
		shardQuery.Offset = 0
		var (
			alerts []*model.Alert
			count  int
			err    error
		)
		if path == s.path {
			alerts, count, err = queryAlertsDB(ctx, s.db, shardQuery)
		} else {
			db, openErr := openReadOnlyDatabase(path)
			if openErr != nil {
				s.markStorageError(openErr)
				return nil, 0, fmt.Errorf("open alert shard %s: %w", path, openErr)
			}
			alerts, count, err = queryAlertsDB(ctx, db, shardQuery)
			closeErr := db.Close()
			if err == nil && closeErr != nil {
				err = fmt.Errorf("close alert shard %s: %w", path, closeErr)
			}
		}
		if err != nil {
			s.markStorageError(err)
			return nil, 0, fmt.Errorf("query alert shard %s: %w", path, err)
		}
		total += count
		all = append(all, alerts...)
	}

	sortAlerts(all)
	limit := query.Limit
	if limit <= 0 {
		limit = 1000
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	start, end := sliceBounds(len(all), limit, offset)
	s.markHealthy()
	return all[start:end], total, nil
}

func alertQueryWhere(query Query) (string, []any) {
	var clauses []string
	var args []any
	add := func(clause string, value any) {
		clauses = append(clauses, clause)
		args = append(args, value)
	}
	if query.RuleID != "" {
		add("rule_id = ?", query.RuleID)
	}
	if query.Severity != "" {
		add("severity = ?", string(query.Severity))
	}
	if query.SrcIP != "" {
		add("src_ip = ?", query.SrcIP)
	}
	if query.DstIP != "" {
		add("dst_ip = ?", query.DstIP)
	}
	if query.Protocol != "" {
		add("UPPER(protocol) = ?", strings.ToUpper(query.Protocol))
	}
	if query.DstPort != nil {
		add("dst_port = ?", int(*query.DstPort))
	}
	if query.Since != nil {
		add("julianday(last_seen) >= julianday(?)", formatTime(*query.Since))
	}
	if query.Until != nil {
		add("julianday(last_seen) <= julianday(?)", formatTime(*query.Until))
	}
	if query.MitreTactic != "" {
		add("LOWER(mitre_tactic) = LOWER(?)", query.MitreTactic)
	}
	if query.MitreTechniqueID != "" {
		add("LOWER(mitre_technique_id) = LOWER(?)", query.MitreTechniqueID)
	}
	if query.MatchedKeyword != "" {
		add("instr(LOWER(matched_keyword), LOWER(?)) > 0", query.MatchedKeyword)
	}
	if query.MinCount != nil {
		add("aggregated_count >= ?", *query.MinCount)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (s *Store) alertShardPaths(ctx context.Context, query Query) ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{s.path}, nil
		}
		return nil, fmt.Errorf("read alert shard dir: %w", err)
	}
	seen := map[string]bool{}
	var paths []string
	addPath := func(path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() {
			continue
		}
		match := dailyShardNameRe.FindStringSubmatch(entry.Name())
		if match == nil || !shardDateMatchesQuery(match[1], query) {
			continue
		}
		addPath(filepath.Join(s.dir, entry.Name()))
	}
	addPath(s.path)
	sort.Strings(paths)
	return paths, nil
}

func shardDateMatchesQuery(date string, query Query) bool {
	if query.Since != nil && date < query.Since.UTC().Format("2006-01-02") {
		return false
	}
	if query.Until != nil && date > query.Until.UTC().Format("2006-01-02") {
		return false
	}
	return true
}

func sortAlerts(alerts []*model.Alert) {
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].LastSeen.Equal(alerts[j].LastSeen) {
			return alerts[i].ID < alerts[j].ID
		}
		return alerts[i].LastSeen.After(alerts[j].LastSeen)
	})
}

func sliceBounds(length, limit, offset int) (int, int) {
	if offset >= length {
		return length, length
	}
	end := offset + limit
	if end > length {
		end = length
	}
	return offset, end
}

// Count returns the number of aggregated alert rows.
func (s *Store) Count(ctx context.Context) (int, error) {
	if !s.dailyShard {
		count, err := countAlertsDB(ctx, s.db)
		if err != nil {
			s.markStorageError(err)
			return 0, err
		}
		return count, nil
	}

	paths, err := s.alertShardPaths(ctx, Query{})
	if err != nil {
		s.markStorageError(err)
		return 0, err
	}
	total := 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		var count int
		if path == s.path {
			count, err = countAlertsDB(ctx, s.db)
		} else {
			var db *sql.DB
			db, err = openReadOnlyDatabase(path)
			if err == nil {
				count, err = countAlertsDB(ctx, db)
				closeErr := db.Close()
				if err == nil && closeErr != nil {
					err = fmt.Errorf("close alert shard %s: %w", path, closeErr)
				}
			}
		}
		if err != nil {
			s.markStorageError(err)
			return 0, fmt.Errorf("count alert shard %s: %w", path, err)
		}
		total += count
	}
	return total, nil
}

func countAlertsDB(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts").Scan(&count); err != nil {
		return 0, fmt.Errorf("count alerts: %w", err)
	}
	return count, nil
}

// PruneExpired deletes alerts older than the configured retention window from
// the currently opened SQLite database. RetentionDays <= 0 disables pruning.
func (s *Store) PruneExpired(ctx context.Context) (int64, error) {
	if s.retentionDays <= 0 {
		return 0, nil
	}
	cutoff := s.retentionCutoff()
	result, err := s.db.ExecContext(ctx, "DELETE FROM alerts WHERE last_seen < ?", formatTime(cutoff))
	if err != nil {
		return 0, fmt.Errorf("prune expired alerts: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count pruned alerts: %w", err)
	}
	return rows, nil
}

var dailyShardNameRe = regexp.MustCompile(`^netsentry-(\d{4}-\d{2}-\d{2})\.db$`)

// PruneExpiredShardFiles deletes old daily shard database files and their WAL/SHM
// sidecars. It only touches files named netsentry-YYYY-MM-DD.db.
func (s *Store) PruneExpiredShardFiles(ctx context.Context, dir string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if s.retentionDays <= 0 {
		return 0, nil
	}
	cutoffDate := s.retentionCutoff().Format("2006-01-02")
	entries, err := os.ReadDir(defaultDBDir(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read alert shard dir: %w", err)
	}

	deleted := 0
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return deleted, err
		}
		if entry.IsDir() {
			continue
		}
		match := dailyShardNameRe.FindStringSubmatch(entry.Name())
		if match == nil || match[1] >= cutoffDate {
			continue
		}
		base := filepath.Join(defaultDBDir(dir), entry.Name())
		removed, err := removeShardSet(base)
		if err != nil {
			return deleted, err
		}
		deleted += removed
	}
	return deleted, nil
}

func (s *Store) retentionCutoff() time.Time {
	return s.now().AddDate(0, 0, -s.retentionDays)
}

func removeShardSet(base string) (int, error) {
	deleted := 0
	for _, path := range []string{base, base + "-wal", base + "-shm"} {
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return deleted, fmt.Errorf("remove alert shard %s: %w", path, err)
		}
		deleted++
	}
	return deleted, nil
}

func scanAlert(rows *sql.Rows) (*model.Alert, error) {
	var alert model.Alert
	var severity string
	var dstPort int
	var firstSeen, lastSeen, windowStart string
	if err := rows.Scan(
		&alert.ID,
		&alert.EventID,
		&alert.RuleID,
		&alert.RuleName,
		&severity,
		&alert.Protocol,
		&alert.SrcIP,
		&alert.DstIP,
		&dstPort,
		&alert.MitreTactic,
		&alert.MitreTechniqueID,
		&alert.MitreTechniqueName,
		&alert.PayloadPreview,
		&alert.MatchedKeyword,
		&alert.AggregatedCount,
		&firstSeen,
		&lastSeen,
		&windowStart,
	); err != nil {
		return nil, fmt.Errorf("scan alert: %w", err)
	}
	parsedFirstSeen, err := parseTime(firstSeen)
	if err != nil {
		return nil, err
	}
	parsedLastSeen, err := parseTime(lastSeen)
	if err != nil {
		return nil, err
	}
	parsedWindowStart, err := parseTime(windowStart)
	if err != nil {
		return nil, err
	}
	alert.Severity = model.Severity(severity)
	alert.DstPort = uint16(dstPort)
	alert.FirstSeen = parsedFirstSeen
	alert.LastSeen = parsedLastSeen
	alert.Timestamp = parsedLastSeen
	alert.WindowStart = parsedWindowStart
	return &alert, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse alert timestamp %q: %w", value, err)
	}
	return parsed, nil
}

// Close releases database resources.
func (s *Store) Close() error {
	return s.db.Close()
}
