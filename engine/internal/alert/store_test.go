package alert

import (
	"context"
	"os"
	"path/filepath"
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
