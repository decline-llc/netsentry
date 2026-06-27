package alert

import (
	"context"
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
