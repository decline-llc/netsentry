package alert_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestStoreWritableQuestionMarkPrimaryPath(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, relative := range []string{
			"ordinary/netsentry.db",
			"directory with spaces/netsentry.db",
			"directory?literal/netsentry.db",
			"ordinary/netsentry?literal.db",
			"space % # ? directory/netsentry % # ?.db",
			"ordinary/netsentry.db?_pragma=synchronous(OFF)",
		} {
			t.Run(pathEncodingMode(durable)+"/"+relative, func(t *testing.T) {
				root := t.TempDir()
				path := filepath.Join(root, relative)
				opts := alert.Options{Path: path, Dir: filepath.Dir(path), JournalMode: "DELETE", RequireDurableWrites: durable}
				if durable {
					opts.JournalMode = "WAL"
				}
				store := openPathEncodingStore(t, opts, path)
				first := pathEncodingAlert(time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC))
				writePathEncodingAlerts(t, store, first)
				assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
				closePathEncodingStore(t, store)
				assertPathEncodingFiles(t, root, []string{relative, relative + ".alerts.jsonl"})
				assertPathEncodingObservedRows(t, path, 1, 1)

				store = openPathEncodingStore(t, opts, path)
				assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
				second := pathEncodingAlert(first.Timestamp.Add(time.Second))
				writePathEncodingAlerts(t, store, second)
				assertPathEncodingRows(t, store, []*model.Alert{second}, []int{2})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, path, 1, 2)
				assertPathEncodingFiles(t, root, []string{relative, relative + ".alerts.jsonl"})
			})
		}
	}
}

func TestStoreWritableQuestionMarkDailyShardPaths(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, directory := range []string{"ordinary", "directory?literal", "space % # ? directory"} {
			t.Run(pathEncodingMode(durable)+"/"+directory, func(t *testing.T) {
				root := t.TempDir()
				dir := filepath.Join(root, directory)
				now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
				currentPath := filepath.Join(dir, "netsentry-2026-10-03.db")
				historicalPath := filepath.Join(dir, "netsentry-2026-10-02.db")
				opts := alert.Options{Dir: dir, DailyShard: true, Now: func() time.Time { return now }, JournalMode: "DELETE", RequireDurableWrites: durable}
				if durable {
					opts.JournalMode = "WAL"
				}
				store := openPathEncodingStore(t, opts, currentPath)
				current := pathEncodingAlert(now)
				historical := pathEncodingAlert(now.AddDate(0, 0, -1))
				// The historical timestamp forces the real non-current openShard path.
				writePathEncodingAlerts(t, store, current, historical)
				assertPathEncodingRows(t, store, []*model.Alert{current, historical}, []int{1, 1})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, currentPath, 1, 1)
				assertPathEncodingObservedRows(t, historicalPath, 1, 1)
				files := []string{
					filepath.Join(directory, "netsentry-2026-10-03.db"),
					filepath.Join(directory, "netsentry-2026-10-02.db"),
					filepath.Join(directory, "netsentry-alerts-recovery.jsonl"),
				}
				assertPathEncodingFiles(t, root, files)
				store = openPathEncodingStore(t, opts, currentPath)
				assertPathEncodingRows(t, store, []*model.Alert{current, historical}, []int{1, 1})
				// A second historical write also reaches existing-file read-only preflight.
				second := pathEncodingAlert(historical.Timestamp.Add(time.Second))
				writePathEncodingAlerts(t, store, second)
				assertPathEncodingRows(t, store, []*model.Alert{current, second}, []int{1, 2})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, currentPath, 1, 1)
				assertPathEncodingObservedRows(t, historicalPath, 1, 2)
				assertPathEncodingFiles(t, root, files)
			})
		}
	}
}

func TestStoreWritableQuestionMarkRejectedRecoveryPreservesTree(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(pathEncodingMode(durable), func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "space % # ? directory", "netsentry?.db")
			writeOpenCancellationFile(t, path+".alerts.jsonl", []byte("{malformed recovery\n"))
			before := snapshotOpenCancellationTree(t, root)
			store, err := alert.Open(context.Background(), alert.Options{Path: path, Dir: filepath.Dir(path), RequireDurableWrites: durable})
			if store != nil {
				t.Cleanup(func() { _ = store.Close() })
				t.Fatal("rejected recovery returned a Store")
			}
			if !errors.Is(err, alert.ErrRecoveryLogIntegrity) {
				t.Fatalf("Open error = %v, want recovery integrity error", err)
			}
			if after := snapshotOpenCancellationTree(t, root); !reflect.DeepEqual(after, before) {
				t.Fatalf("rejected recovery changed tree: before=%v after=%v", before, after)
			}
		})
	}
}

func pathEncodingMode(durable bool) string {
	if durable {
		return "durable-WAL"
	}
	return "ordinary-DELETE"
}

func pathEncodingAlert(timestamp time.Time) *model.Alert {
	return &model.Alert{
		EventID: "path-encoding-" + timestamp.Format(time.RFC3339Nano),
		RuleID:  "path-encoding", RuleName: "Path encoding fixture",
		Timestamp: timestamp, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 80,
		Protocol: "TCP", Severity: model.SeverityHigh,
		MatchedKeyword: "needle", PayloadPreview: "path encoding payload",
	}
}

func openPathEncodingStore(t *testing.T, opts alert.Options, wantPath string) *alert.Store {
	t.Helper()
	store, err := alert.Open(context.Background(), opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if store.Path() != wantPath {
		t.Fatalf("Store.Path = %q, want %q", store.Path(), wantPath)
	}
	return store
}

func closePathEncodingStore(t *testing.T, store *alert.Store) {
	t.Helper()
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func writePathEncodingAlerts(t *testing.T, store *alert.Store, alerts ...*model.Alert) {
	t.Helper()
	before := make([]model.Alert, len(alerts))
	for i, value := range alerts {
		before[i] = *value
	}
	if err := store.WriteBatch(context.Background(), alerts); err != nil {
		t.Fatalf("WriteBatch: %v", err)
	}
	for i, value := range alerts {
		if !reflect.DeepEqual(*value, before[i]) {
			t.Fatalf("WriteBatch mutated input %d", i)
		}
	}
}

func assertPathEncodingRows(t *testing.T, store *alert.Store, want []*model.Alert, counts []int) {
	t.Helper()
	rows, total, err := store.Query(context.Background(), alert.Query{Limit: -1})
	if err != nil || total != len(want) || len(rows) != len(want) {
		t.Fatalf("Query rows=%d total=%d error=%v, want %d", len(rows), total, err, len(want))
	}
	for i, row := range rows {
		value := want[i]
		if row.RuleID != value.RuleID || row.RuleName != value.RuleName || row.SrcIP != value.SrcIP ||
			row.DstIP != value.DstIP || row.DstPort != value.DstPort || row.Protocol != value.Protocol ||
			row.Severity != value.Severity || row.PayloadPreview != value.PayloadPreview ||
			row.MatchedKeyword != value.MatchedKeyword || row.AggregatedCount != counts[i] ||
			!row.Timestamp.Equal(value.Timestamp) || !row.LastSeen.Equal(value.Timestamp) ||
			!row.WindowStart.Equal(value.Timestamp.Truncate(time.Minute)) {
			t.Fatalf("Query row %d = %+v, want fixture %+v with aggregate %d", i, row, value, counts[i])
		}
	}
	listed, err := store.List(context.Background())
	if err != nil || !reflect.DeepEqual(listed, rows) {
		t.Fatalf("List = %+v error=%v, want Query rows %+v", listed, err, rows)
	}
	count, err := store.Count(context.Background())
	if err != nil || count != len(want) {
		t.Fatalf("Count = %d error=%v, want %d", count, err, len(want))
	}
}

func assertPathEncodingObservedRows(t *testing.T, path string, wantRows, wantAggregate int) {
	t.Helper()
	observer := openCancellationReadOnlyObserver(t, path)
	var rows, aggregate int
	if err := observer.QueryRow("SELECT COUNT(*), SUM(aggregated_count) FROM alerts WHERE rule_id = ?", "path-encoding").Scan(&rows, &aggregate); err != nil {
		t.Fatalf("independent literal-path read: %v", err)
	}
	if rows != wantRows || aggregate != wantAggregate {
		t.Fatalf("independent rows=%d aggregate=%d, want %d/%d", rows, aggregate, wantRows, wantAggregate)
	}
	if err := observer.Close(); err != nil {
		t.Fatalf("close independent observer: %v", err)
	}
}

func assertPathEncodingFiles(t *testing.T, root string, want []string) {
	t.Helper()
	var files []string
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// WAL sidecars may legitimately remain after a read-only observer.
		// Accept them only beside the exact expected database paths.
		for _, suffix := range []string{"-wal", "-shm"} {
			for _, expected := range want {
				if filepath.Ext(expected) != ".jsonl" && rel == expected+suffix {
					return nil
				}
			}
		}
		files = append(files, rel)
		if filepath.Ext(path) == ".jsonl" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(data) != 0 {
				t.Fatalf("recovery log is not cleared: %q", data)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("enumerate literal-path files: %v", err)
	}
	sort.Strings(files)
	expected := append([]string(nil), want...)
	sort.Strings(expected)
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("files = %q, want %q; truncated/alternate databases must not appear", files, expected)
	}
}
