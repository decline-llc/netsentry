package alert

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func BenchmarkStoreWriteBatch(b *testing.B) {
	cases := []struct {
		name      string
		batchSize int
	}{
		{name: "single_alert", batchSize: 1},
		{name: "batch_32_alerts", batchSize: 32},
	}

	for _, benchmark := range cases {
		b.Run(benchmark.name, func(b *testing.B) {
			b.StopTimer()
			ctx := context.Background()
			store := openBenchmarkStore(b)
			base := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

			b.ReportAllocs()
			b.ResetTimer()
			for iteration := 0; iteration < b.N; iteration++ {
				alerts := benchmarkWriteAlerts(base, iteration, benchmark.batchSize)

				b.StartTimer()
				err := store.WriteBatch(ctx, alerts)
				b.StopTimer()
				if err != nil {
					b.Fatalf("write benchmark batch: %v", err)
				}

				assertBenchmarkWriteState(b, store, benchmark.batchSize, 1)
				clearBenchmarkStoreRows(b, store)
			}
			b.ReportMetric(float64(benchmark.batchSize), "alerts/op")
		})
	}
}

func BenchmarkStoreQuery(b *testing.B) {
	b.StopTimer()
	ctx := context.Background()
	store := openBenchmarkStore(b)
	base := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	const fixtureCount = 512
	if err := store.WriteBatch(ctx, benchmarkQueryAlerts(base, fixtureCount)); err != nil {
		b.Fatalf("seed query benchmark: %v", err)
	}
	assertBenchmarkWriteState(b, store, fixtureCount, fixtureCount)

	since := base.Add(224 * time.Minute)
	until := base.Add(287 * time.Minute)
	cases := []struct {
		name      string
		query     Query
		indexName string
		wantTotal int
		validate  func(*model.Alert) bool
	}{
		{
			name:      "exact_rule",
			query:     Query{RuleID: "benchmark-rule-03", Limit: 16},
			indexName: "idx_alerts_rule_last_seen",
			wantTotal: fixtureCount / 8,
			validate: func(alert *model.Alert) bool {
				return alert.RuleID == "benchmark-rule-03"
			},
		},
		{
			name: "timestamp_range",
			query: Query{
				Since: &since,
				Until: &until,
				Limit: 16,
			},
			indexName: "idx_alerts_last_seen_time_id",
			wantTotal: 64,
			validate: func(alert *model.Alert) bool {
				return !alert.LastSeen.Before(since) && !alert.LastSeen.After(until)
			},
		},
	}

	for _, benchmark := range cases {
		b.Run(benchmark.name, func(b *testing.B) {
			b.StopTimer()
			assertBenchmarkQueryUsesIndex(b, store, benchmark.query, benchmark.indexName)
			alerts, total, err := store.Query(ctx, benchmark.query)
			if err != nil {
				b.Fatalf("verify benchmark query: %v", err)
			}
			assertBenchmarkQueryResult(b, alerts, total, benchmark.query.Limit, benchmark.wantTotal, benchmark.validate)

			b.ReportAllocs()
			b.ResetTimer()
			b.StartTimer()
			for iteration := 0; iteration < b.N; iteration++ {
				alerts, total, err = store.Query(ctx, benchmark.query)
				if err != nil {
					b.Fatalf("run benchmark query: %v", err)
				}
			}
			b.StopTimer()

			runtime.KeepAlive(alerts)
			assertBenchmarkQueryResult(b, alerts, total, benchmark.query.Limit, benchmark.wantTotal, benchmark.validate)
		})
	}
}

func openBenchmarkStore(b *testing.B) *Store {
	b.Helper()
	dir := b.TempDir()
	store, err := Open(context.Background(), Options{
		Path:              filepath.Join(dir, "alerts.db"),
		JournalMode:       "WAL",
		BusyTimeoutMS:     1000,
		AggregationWindow: time.Minute,
		Now: func() time.Time {
			return time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		b.Fatalf("open benchmark store: %v", err)
	}
	b.Cleanup(func() {
		if err := store.Close(); err != nil {
			b.Errorf("close benchmark store: %v", err)
		}
	})
	return store
}

func benchmarkWriteAlerts(base time.Time, iteration, batchSize int) []*model.Alert {
	alerts := make([]*model.Alert, 0, batchSize)
	first := int64(iteration) * int64(batchSize)
	for offset := 0; offset < batchSize; offset++ {
		sequence := first + int64(offset)
		alert := benchmarkAlert(base.Add(time.Duration(sequence)), fmt.Sprintf("write-%d", sequence))
		alert.EventID = alertEventID(normalizeAlert(alert, alert.Timestamp, time.Minute))
		alerts = append(alerts, alert)
	}
	return alerts
}

func benchmarkQueryAlerts(base time.Time, count int) []*model.Alert {
	alerts := make([]*model.Alert, 0, count)
	for index := 0; index < count; index++ {
		alert := benchmarkAlert(base.Add(time.Duration(index)*time.Minute), fmt.Sprintf("query-%03d", index))
		alert.RuleID = fmt.Sprintf("benchmark-rule-%02d", index%8)
		if index%4 == 0 {
			alert.Severity = model.SeverityCritical
		}
		alert.EventID = alertEventID(normalizeAlert(alert, alert.Timestamp, time.Minute))
		alerts = append(alerts, alert)
	}
	return alerts
}

func benchmarkAlert(timestamp time.Time, keyword string) *model.Alert {
	return &model.Alert{
		RuleID:             "benchmark-rule",
		RuleName:           "Benchmark Rule",
		Timestamp:          timestamp,
		SrcIP:              "192.0.2.10",
		DstIP:              "198.51.100.20",
		DstPort:            443,
		Protocol:           "TCP",
		Severity:           model.SeverityHigh,
		MitreTactic:        "Initial Access",
		MitreTechniqueID:   "T1190",
		MitreTechniqueName: "Exploit Public-Facing Application",
		PayloadPreview:     "deterministic benchmark payload",
		MatchedKeyword:     keyword,
	}
}

func assertBenchmarkWriteState(b *testing.B, store *Store, wantEvents, wantRows int) {
	b.Helper()
	var events, rows, aggregateCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM alert_events").Scan(&events); err != nil {
		b.Fatalf("count benchmark alert events: %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*), COALESCE(SUM(aggregated_count), 0) FROM alerts").Scan(&rows, &aggregateCount); err != nil {
		b.Fatalf("count benchmark aggregate rows: %v", err)
	}
	if events != wantEvents || rows != wantRows || aggregateCount != wantEvents {
		b.Fatalf(
			"benchmark write state = events %d rows %d aggregate %d, want %d/%d/%d",
			events,
			rows,
			aggregateCount,
			wantEvents,
			wantRows,
			wantEvents,
		)
	}
	recovery, err := os.ReadFile(store.recoveryLogPath)
	if err != nil {
		b.Fatalf("read benchmark recovery log: %v", err)
	}
	if len(recovery) != 0 {
		b.Fatalf("benchmark recovery log has %d bytes after successful write, want 0", len(recovery))
	}
	if health := store.Health(); health.Status != "ok" {
		b.Fatalf("benchmark store health = %+v, want ok", health)
	}
}

func clearBenchmarkStoreRows(b *testing.B, store *Store) {
	b.Helper()
	tx, err := store.db.Begin()
	if err != nil {
		b.Fatalf("begin benchmark cleanup: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM alert_events"); err != nil {
		_ = tx.Rollback()
		b.Fatalf("clear benchmark alert events: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM alerts"); err != nil {
		_ = tx.Rollback()
		b.Fatalf("clear benchmark aggregate rows: %v", err)
	}
	if err := tx.Commit(); err != nil {
		b.Fatalf("commit benchmark cleanup: %v", err)
	}
}

func assertBenchmarkQueryUsesIndex(b *testing.B, store *Store, query Query, indexName string) {
	b.Helper()
	where, args := alertQueryWhere(query)
	limit := query.Limit
	if limit <= 0 {
		limit = 1000
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	planArgs := append([]any{}, args...)
	planArgs = append(planArgs, limit, offset)
	rows, err := store.db.Query(
		"EXPLAIN QUERY PLAN "+alertSelectColumns+where+alertOrderSQL+"\nLIMIT ? OFFSET ?",
		planArgs...,
	)
	if err != nil {
		b.Fatalf("explain benchmark query: %v", err)
	}
	defer rows.Close()

	var details []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			b.Fatalf("scan benchmark query plan: %v", err)
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		b.Fatalf("iterate benchmark query plan: %v", err)
	}
	plan := strings.Join(details, "\n")
	if !strings.Contains(plan, indexName) {
		b.Fatalf("benchmark query plan does not use %s:\n%s", indexName, plan)
	}
}

func assertBenchmarkQueryResult(
	b *testing.B,
	alerts []*model.Alert,
	total, limit, wantTotal int,
	validate func(*model.Alert) bool,
) {
	b.Helper()
	wantPage := limit
	if wantTotal < wantPage {
		wantPage = wantTotal
	}
	if total != wantTotal || len(alerts) != wantPage {
		b.Fatalf("benchmark query = total %d page %d, want %d/%d", total, len(alerts), wantTotal, wantPage)
	}
	for _, alert := range alerts {
		if alert == nil || !validate(alert) {
			b.Fatalf("benchmark query returned unexpected alert: %+v", alert)
		}
	}
}
