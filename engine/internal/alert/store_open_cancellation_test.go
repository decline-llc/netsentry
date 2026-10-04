package alert_test

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestOpenAlreadyDoneContextPreservesInputs(t *testing.T) {
	for _, cause := range []string{"canceled", "deadline"} {
		for _, directory := range []string{"ordinary", "directory with spaces"} {
			for _, fixture := range []string{"absent", "healthy", "corrupt-artifacts", "malformed-recovery", "file-parent"} {
				t.Run(cause+"/"+directory+"/"+fixture, func(t *testing.T) {
					root := t.TempDir()
					path := filepath.Join(root, directory, "nested", "netsentry.db")
					opts := alert.Options{Path: path, Dir: filepath.Dir(path), JournalMode: "DELETE"}
					var observer *sql.DB
					switch fixture {
					case "absent":
						// No target directories exist before the rejected call.
					case "healthy":
						seedOpenCancellationStore(t, opts)
						observer = openCancellationReadOnlyObserver(t, path)
						assertOpenCancellationRowCount(t, observer, 1)
					case "corrupt-artifacts":
						writeOpenCancellationFile(t, path, []byte("corrupt database\x00retained"))
						writeOpenCancellationFile(t, path+"-wal", []byte("retained WAL\x00"))
						writeOpenCancellationFile(t, path+"-shm", []byte("retained SHM\x00"))
						writeOpenCancellationFile(t, path+".alerts.jsonl", []byte("{malformed recovery\n"))
					case "malformed-recovery":
						opts.RecoveryLogPath = filepath.Join(root, "retained recovery.jsonl")
						writeOpenCancellationFile(t, opts.RecoveryLogPath, []byte("{\"broken\":\n"))
					case "file-parent":
						parent := filepath.Join(root, directory)
						writeOpenCancellationFile(t, parent, []byte("retained regular-file occupant"))
					}
					before := snapshotOpenCancellationTree(t, root)
					ctx, want := alreadyDoneOpenContext(t, cause)
					store, err := alert.Open(ctx, opts)
					if store != nil {
						t.Cleanup(func() { _ = store.Close() })
						t.Fatalf("Open returned a Store for already-done context")
					}
					if !errors.Is(err, want) || err != want {
						t.Fatalf("Open error = %v, want unchanged %v", err, want)
					}
					if observer != nil {
						// Reuse the observer opened before rejection; no writable reopen.
						assertOpenCancellationRowCount(t, observer, 1)
					}
					if after := snapshotOpenCancellationTree(t, root); !reflect.DeepEqual(after, before) {
						t.Fatalf("Open changed tree membership, bytes or modes:\nbefore=%v\nafter=%v", before, after)
					}
				})
			}
		}
	}
}

func TestOpenAlreadyDoneContextPrecedesDurableOptionValidation(t *testing.T) {
	for _, cause := range []string{"canceled", "deadline"} {
		t.Run(cause, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "absent", "netsentry.db")
			before := snapshotOpenCancellationTree(t, root)
			ctx, want := alreadyDoneOpenContext(t, cause)
			store, err := alert.Open(ctx, alert.Options{
				Path: path, Dir: filepath.Dir(path), JournalMode: "DELETE", RequireDurableWrites: true,
			})
			if store != nil {
				t.Cleanup(func() { _ = store.Close() })
				t.Fatal("Open returned a Store for already-done context")
			}
			if !errors.Is(err, want) || err != want {
				t.Fatalf("Open error = %v, want context error before durable-mode diagnostic %v", err, want)
			}
			if after := snapshotOpenCancellationTree(t, root); !reflect.DeepEqual(after, before) {
				t.Fatalf("invalid-option rejected Open changed tree: before=%v after=%v", before, after)
			}
		})
	}
}

func TestOpenLiveContextCreatesReadableStore(t *testing.T) {
	for _, kind := range []string{"background", "cancelable"} {
		for _, directory := range []string{"ordinary", "directory with spaces"} {
			t.Run(kind+"/"+directory, func(t *testing.T) {
				ctx := context.Background()
				if kind == "cancelable" {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					t.Cleanup(cancel)
				}
				path := filepath.Join(t.TempDir(), directory, "nested", "netsentry.db")
				store, err := alert.Open(ctx, alert.Options{Path: path, Dir: filepath.Dir(path), JournalMode: "DELETE"})
				if err != nil {
					t.Fatalf("Open live context: %v", err)
				}
				t.Cleanup(func() { _ = store.Close() })
				row := openCancellationAlert()
				if err := store.WriteBatch(ctx, []*model.Alert{row}); err != nil {
					t.Fatalf("write live store: %v", err)
				}
				rows, err := store.List(ctx)
				if err != nil || len(rows) != 1 {
					t.Fatalf("List live store = %v, %v; want one row", rows, err)
				}
				if rows[0].RuleID != row.RuleID || rows[0].PayloadPreview != row.PayloadPreview || rows[0].AggregatedCount != 1 {
					t.Fatalf("live stored row = %+v, want fixture content and count 1", rows[0])
				}
				if count, err := store.Count(ctx); err != nil || count != 1 {
					t.Fatalf("Count live store = %d, %v; want 1", count, err)
				}
				if err := store.Close(); err != nil {
					t.Fatalf("close live store: %v", err)
				}
				assertOpenCancellationRowCount(t, openCancellationReadOnlyObserver(t, path), 1)
			})
		}
	}
}

func alreadyDoneOpenContext(t *testing.T, cause string) (context.Context, error) {
	t.Helper()
	if cause == "canceled" {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx, context.Canceled
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(1, 0))
	t.Cleanup(cancel)
	return ctx, context.DeadlineExceeded
}

func openCancellationAlert() *model.Alert {
	return &model.Alert{
		RuleID: "open-cancellation-fixture", RuleName: "Open cancellation fixture",
		Timestamp: time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC),
		SrcIP:     "192.0.2.1", DstIP: "198.51.100.1", DstPort: 80,
		Protocol: "TCP", Severity: model.SeverityHigh,
		MatchedKeyword: "needle", PayloadPreview: "retained payload",
	}
}

func seedOpenCancellationStore(t *testing.T, opts alert.Options) {
	t.Helper()
	store, err := alert.Open(context.Background(), opts)
	if err != nil {
		t.Fatalf("seed Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.WriteBatch(context.Background(), []*model.Alert{openCancellationAlert()}); err != nil {
		t.Fatalf("seed WriteBatch: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("seed Close: %v", err)
	}
}

func openCancellationReadOnlyObserver(t *testing.T, path string) *sql.DB {
	t.Helper()
	dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: url.Values{"mode": {"ro"}}.Encode()}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("read-only observer: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertOpenCancellationRowCount(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM alerts").Scan(&count); err != nil {
		t.Fatalf("read-only observer count: %v", err)
	}
	if count != want {
		t.Fatalf("read-only observer count = %d, want %d", count, want)
	}
}

func writeOpenCancellationFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("fixture directory: %v", err)
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatalf("fixture file: %v", err)
	}
}

type openCancellationTreeEntry struct {
	Mode os.FileMode
	Data []byte
}

func snapshotOpenCancellationTree(t *testing.T, root string) map[string]openCancellationTreeEntry {
	t.Helper()
	entries := make(map[string]openCancellationTreeEntry)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := openCancellationTreeEntry{Mode: info.Mode()}
		if info.Mode().IsRegular() {
			value.Data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}
		entries[rel] = value
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot tree: %v", err)
	}
	return entries
}
