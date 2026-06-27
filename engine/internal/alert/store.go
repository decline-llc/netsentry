package alert

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Options controls the SQLite alert store.
type Options struct {
	Path              string
	JournalMode       string
	BusyTimeoutMS     int
	AggregationWindow time.Duration
}

// Store persists alerts and aggregates repeated hits in a fixed time window.
type Store struct {
	db                *sql.DB
	aggregationWindow time.Duration
}

// Open creates the SQLite database and initializes its schema.
func Open(ctx context.Context, opts Options) (*Store, error) {
	if opts.Path == "" {
		opts.Path = "data/netsentry.db"
	}
	if opts.JournalMode == "" {
		opts.JournalMode = "WAL"
	}
	if opts.BusyTimeoutMS <= 0 {
		opts.BusyTimeoutMS = 5000
	}
	if opts.AggregationWindow <= 0 {
		opts.AggregationWindow = time.Minute
	}

	db, err := sql.Open("sqlite", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite alerts store: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db, aggregationWindow: opts.AggregationWindow}
	if err := store.init(ctx, opts); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
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
    last_seen = excluded.last_seen,
    payload_preview = excluded.payload_preview,
    matched_keyword = excluded.matched_keyword,
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
