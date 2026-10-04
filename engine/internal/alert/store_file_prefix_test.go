package alert_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestStoreLiteralFilePrefixPrimaryPathsPreserveAlternateTarget(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, relative := range []string{
			"ordinary/netsentry.db",
			"FILE:netsentry.db",
			"file:",
			"file:netsentry.db",
			"file:netsentry % #.db",
			"file:directory/netsentry.db",
			"file:directory with spaces % #/netsentry.db",
		} {
			t.Run(pathEncodingMode(durable)+"/"+relative, func(t *testing.T) {
				root := t.TempDir()
				// A leading file: reaches SQLite only when the actual input is relative.
				t.Chdir(root)
				literalPath := filepath.Join(root, relative)
				opts := filePrefixOptions(relative, durable)
				files := []string{relative, relative + ".alerts.jsonl"}
				var observer *sql.DB
				var decoyFiles []string
				var decoyBefore map[string]openCancellationTreeEntry
				if strings.HasPrefix(relative, "file:") && relative != "file:" {
					decoy := strings.TrimPrefix(relative, "file:")
					seedFilePrefixDecoy(t, decoy)
					decoyFiles = []string{decoy, decoy + ".alerts.jsonl"}
					files = append(files, decoyFiles...)
					// Open and query before the writer; reuse this handle afterward.
					observer = openCancellationReadOnlyObserver(t, filepath.Join(root, decoy))
					assertFilePrefixDecoyRows(t, observer)
					decoyBefore = snapshotFilePrefixFiles(t, decoyFiles)
				}
				store := openPathEncodingStore(t, opts, relative)
				first := pathEncodingAlert(time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC))
				writePathEncodingAlerts(t, store, first)
				assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, literalPath, 1, 1)
				assertPathEncodingFiles(t, root, files)
				assertFilePrefixDecoyPreserved(t, observer, decoyFiles, decoyBefore)

				store = openPathEncodingStore(t, opts, relative)
				assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
				second := pathEncodingAlert(first.Timestamp.Add(time.Second))
				writePathEncodingAlerts(t, store, second)
				assertPathEncodingRows(t, store, []*model.Alert{second}, []int{2})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, literalPath, 1, 2)
				assertPathEncodingFiles(t, root, files)
				assertFilePrefixDecoyPreserved(t, observer, decoyFiles, decoyBefore)
			})
		}
	}
}

func TestStoreLiteralFilePrefixDailyShardPaths(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, directory := range []string{"ordinary", "file:directory", "file:directory with spaces % #"} {
			t.Run(pathEncodingMode(durable)+"/"+directory, func(t *testing.T) {
				root := t.TempDir()
				t.Chdir(root)
				now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
				currentPath := filepath.Join(directory, "netsentry-2026-10-03.db")
				historicalPath := filepath.Join(directory, "netsentry-2026-10-02.db")
				opts := filePrefixOptions(currentPath, durable)
				opts.Dir, opts.DailyShard, opts.Now = directory, true, func() time.Time { return now }
				store := openPathEncodingStore(t, opts, currentPath)
				current := pathEncodingAlert(now)
				historical := pathEncodingAlert(now.AddDate(0, 0, -1))
				// Historical timestamp takes the actual non-current openShard route.
				writePathEncodingAlerts(t, store, current, historical)
				assertPathEncodingRows(t, store, []*model.Alert{current, historical}, []int{1, 1})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, filepath.Join(root, currentPath), 1, 1)
				assertPathEncodingObservedRows(t, filepath.Join(root, historicalPath), 1, 1)
				files := []string{currentPath, historicalPath, filepath.Join(directory, "netsentry-alerts-recovery.jsonl")}
				assertPathEncodingFiles(t, root, files)

				store = openPathEncodingStore(t, opts, currentPath)
				assertPathEncodingRows(t, store, []*model.Alert{current, historical}, []int{1, 1})
				// Retry also reaches existing historical-file read-only preflight.
				second := pathEncodingAlert(historical.Timestamp.Add(time.Second))
				writePathEncodingAlerts(t, store, second)
				assertPathEncodingRows(t, store, []*model.Alert{current, second}, []int{1, 2})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, filepath.Join(root, currentPath), 1, 1)
				assertPathEncodingObservedRows(t, filepath.Join(root, historicalPath), 1, 2)
				assertPathEncodingFiles(t, root, files)
			})
		}
	}
}

func TestStoreLiteralFilePrefixRejectedInputsPreserveTree(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, fixture := range []string{"recovery", "database"} {
			t.Run(pathEncodingMode(durable)+"/"+fixture, func(t *testing.T) {
				root := t.TempDir()
				t.Chdir(root)
				path := filepath.Join("file:directory with spaces % #", "netsentry.db")
				want := alert.ErrRecoveryLogIntegrity
				if fixture == "recovery" {
					writeOpenCancellationFile(t, path+".alerts.jsonl", []byte("{malformed recovery\n"))
				} else {
					want = alert.ErrDatabaseIntegrity
					writeOpenCancellationFile(t, path, []byte("retained corrupt SQLite database\x00"))
				}
				before := snapshotOpenCancellationTree(t, root)
				store, err := alert.Open(context.Background(), filePrefixOptions(path, durable))
				if store != nil {
					t.Cleanup(func() { _ = store.Close() })
					t.Fatal("rejected persistent input returned a Store")
				}
				if !errors.Is(err, want) {
					t.Fatalf("Open error=%v, want %v", err, want)
				}
				if fixture == "database" && !strings.Contains(err.Error(), "not a database") {
					t.Fatalf("integrity rejection did not reach corrupt SQLite input: %v", err)
				}
				if after := snapshotOpenCancellationTree(t, root); !reflect.DeepEqual(after, before) {
					t.Fatalf("rejected input changed tree: before=%v after=%v", before, after)
				}
			})
		}
	}
}

func filePrefixOptions(path string, durable bool) alert.Options {
	mode := "DELETE"
	if durable {
		mode = "WAL"
	}
	return alert.Options{Path: path, Dir: filepath.Dir(path), JournalMode: mode, RequireDurableWrites: durable}
}

func seedFilePrefixDecoy(t *testing.T, path string) {
	t.Helper()
	store := openPathEncodingStore(t, filePrefixOptions(path, false), path)
	decoy := pathEncodingAlert(time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC))
	decoy.RuleID, decoy.RuleName = "uri-decoy", "URI alternate-target fixture"
	writePathEncodingAlerts(t, store, decoy)
	closePathEncodingStore(t, store)
}

func assertFilePrefixDecoyRows(t *testing.T, observer *sql.DB) {
	t.Helper()
	var rows, aggregates, decoys int
	if err := observer.QueryRow("SELECT COUNT(*), SUM(aggregated_count), SUM(rule_id = 'uri-decoy') FROM alerts").Scan(&rows, &aggregates, &decoys); err != nil {
		t.Fatalf("independent decoy observer: %v", err)
	}
	if rows != 1 || aggregates != 1 || decoys != 1 {
		t.Fatalf("decoy rows=%d aggregates=%d decoys=%d, want 1/1/1", rows, aggregates, decoys)
	}
}

func snapshotFilePrefixFiles(t *testing.T, files []string) map[string]openCancellationTreeEntry {
	t.Helper()
	out := make(map[string]openCancellationTreeEntry, len(files))
	for _, path := range files {
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatalf("inspect decoy file: %v", err)
		}
		if !info.Mode().IsRegular() {
			t.Fatalf("decoy path is not a regular file: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read decoy bytes: %v", err)
		}
		out[path] = openCancellationTreeEntry{Mode: info.Mode(), Data: data}
	}
	return out
}

func assertFilePrefixDecoyPreserved(t *testing.T, observer *sql.DB, files []string, before map[string]openCancellationTreeEntry) {
	t.Helper()
	if observer == nil {
		return
	}
	assertFilePrefixDecoyRows(t, observer)
	if after := snapshotFilePrefixFiles(t, files); !reflect.DeepEqual(after, before) {
		t.Fatalf("literal-path writer modified alternate URI target: before=%v after=%v", before, after)
	}
}
