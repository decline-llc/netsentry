package alert

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Options controls the SQLite alert store.
type Options struct {
	Path              string
	Dir               string
	DailyShard        bool
	JournalMode       string
	BusyTimeoutMS     int
	AggregationWindow time.Duration
	RetentionDays     int
	Now               func() time.Time
}

// Store persists alerts and aggregates repeated hits in a fixed time window.
type Store struct {
	db                *sql.DB
	path              string
	aggregationWindow time.Duration
	retentionDays     int
	now               func() time.Time
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

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite alerts store: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{
		db:                db,
		path:              dbPath,
		aggregationWindow: opts.AggregationWindow,
		retentionDays:     opts.RetentionDays,
		now:               clock(opts.Now),
	}
	if err := store.init(ctx, opts); err != nil {
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

CREATE INDEX IF NOT EXISTS idx_alerts_last_seen ON alerts(last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_rule_window ON alerts(rule_id, window_start);
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

// WriteBatch inserts alerts and aggregates repeats in the configured window.
func (s *Store) WriteBatch(ctx context.Context, alerts []*model.Alert) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(alerts) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin alert transaction: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, upsertAlertSQL)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare alert upsert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		normalized := normalizeAlert(alert, now, s.aggregationWindow)
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
			return fmt.Errorf("upsert alert %s: %w", normalized.RuleID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit alert transaction: %w", err)
	}
	return nil
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
	out.EventID = out.ID
	return out
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
		return nil, fmt.Errorf("iterate alerts: %w", err)
	}
	return alerts, nil
}

// Count returns the number of aggregated alert rows.
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts").Scan(&count); err != nil {
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
