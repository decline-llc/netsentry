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

func TestStoreLiteralMemoryFilenamePersistsAcrossReopen(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, relative := range []string{":memory:", "./:memory:", ":MEMORY:", "directory with spaces/:memory:"} {
			t.Run(pathEncodingMode(durable)+"/"+relative, func(t *testing.T) {
				root := memoryFilenameRoot(t)
				// Keep the exact sentinel in the public input; an absolute input
				// would already bypass SQLite's in-memory recognition.
				t.Chdir(root)
				opts := filePrefixOptions(relative, durable)
				literal := filepath.Join(root, relative)
				files := []string{filepath.Clean(relative), filepath.Clean(relative + ".alerts.jsonl")}
				store := openPathEncodingStore(t, opts, relative)
				first := pathEncodingAlert(time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC))
				writePathEncodingAlerts(t, store, first)
				assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, literal, 1, 1)
				assertPathEncodingFiles(t, root, files)

				store = openPathEncodingStore(t, opts, relative)
				assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
				second := pathEncodingAlert(first.Timestamp.Add(time.Second))
				writePathEncodingAlerts(t, store, second)
				assertPathEncodingRows(t, store, []*model.Alert{second}, []int{2})
				closePathEncodingStore(t, store)
				assertPathEncodingObservedRows(t, literal, 1, 2)
				assertPathEncodingFiles(t, root, files)
			})
		}
	}
}

func TestStoreLiteralMemoryFilenameUsesExistingFile(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(pathEncodingMode(durable), func(t *testing.T) {
			root := memoryFilenameRoot(t)
			t.Chdir(root)
			literal := filepath.Join(root, ":memory:")
			first := pathEncodingAlert(time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC))
			seed := openPathEncodingStore(t, filePrefixOptions(literal, durable), literal)
			writePathEncodingAlerts(t, seed, first)
			closePathEncodingStore(t, seed)
			// Establish an independent encoded absolute read-only observer
			// before opening the relative writer; reuse it across both opens.
			observer := openCancellationReadOnlyObserver(t, literal)
			assertMemoryFilenameObserver(t, observer, 1)
			store := openPathEncodingStore(t, filePrefixOptions(":memory:", durable), ":memory:")
			assertPathEncodingRows(t, store, []*model.Alert{first}, []int{1})
			second := pathEncodingAlert(first.Timestamp.Add(time.Second))
			writePathEncodingAlerts(t, store, second)
			assertPathEncodingRows(t, store, []*model.Alert{second}, []int{2})
			closePathEncodingStore(t, store)
			assertMemoryFilenameObserver(t, observer, 2)
			store = openPathEncodingStore(t, filePrefixOptions(":memory:", durable), ":memory:")
			assertPathEncodingRows(t, store, []*model.Alert{second}, []int{2})
			closePathEncodingStore(t, store)
			assertMemoryFilenameObserver(t, observer, 2)
			assertPathEncodingFiles(t, root, []string{":memory:", ":memory:.alerts.jsonl"})
		})
	}
}

func TestStoreLiteralMemoryFilenameRejectedInputsPreserveTree(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, fixture := range []string{"recovery", "database"} {
			t.Run(pathEncodingMode(durable)+"/"+fixture, func(t *testing.T) {
				root := memoryFilenameRoot(t)
				t.Chdir(root)
				want := alert.ErrRecoveryLogIntegrity
				if fixture == "recovery" {
					writeOpenCancellationFile(t, ":memory:.alerts.jsonl", []byte("{malformed recovery\n"))
				} else {
					want = alert.ErrDatabaseIntegrity
					writeOpenCancellationFile(t, ":memory:", []byte("retained corrupt SQLite database\x00"))
				}
				before := snapshotOpenCancellationTree(t, root)
				store, err := alert.Open(context.Background(), filePrefixOptions(":memory:", durable))
				if store != nil {
					t.Cleanup(func() { _ = store.Close() })
					t.Fatal("rejected persistent input returned a Store")
				}
				if !errors.Is(err, want) {
					t.Fatalf("Open error=%v, want %v", err, want)
				}
				if fixture == "database" && !strings.Contains(err.Error(), "not a database") {
					t.Fatalf("rejection did not reach corrupt SQLite input: %v", err)
				}
				if after := snapshotOpenCancellationTree(t, root); !reflect.DeepEqual(after, before) {
					t.Fatalf("rejected input changed tree: before=%v after=%v", before, after)
				}
			})
		}
	}
}

func memoryFilenameRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "base with spaces % #")
	if err := os.Mkdir(root, 0o750); err != nil {
		t.Fatalf("create encoded-path fixture: %v", err)
	}
	return root
}

func assertMemoryFilenameObserver(t *testing.T, observer *sql.DB, wantAggregate int) {
	t.Helper()
	var rows, aggregate int
	if err := observer.QueryRow("SELECT COUNT(*), SUM(aggregated_count) FROM alerts WHERE rule_id = ?", "path-encoding").Scan(&rows, &aggregate); err != nil {
		t.Fatalf("independent literal-memory-file observer: %v", err)
	}
	if rows != 1 || aggregate != wantAggregate {
		t.Fatalf("observer rows=%d aggregate=%d, want 1/%d", rows, aggregate, wantAggregate)
	}
}
