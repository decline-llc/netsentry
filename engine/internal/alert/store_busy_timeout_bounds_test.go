package alert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

var busyBoundsNow = time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)

func busyBoundsOptions(root string, daily, durable bool) Options {
	dir := filepath.Join(root, "busy path % #", "nested")
	mode := "DELETE"
	if durable {
		mode = "WAL"
	}
	return Options{Path: filepath.Join(dir, "netsentry.db"), Dir: dir, DailyShard: daily,
		JournalMode: mode, RequireDurableWrites: durable, BusyTimeoutMS: 5000,
		Now: func() time.Time { return busyBoundsNow }}
}

func TestOpenBusyTimeoutOverflowPreservesPersistentInputs(t *testing.T) {
	for _, value := range []int64{2147483648, 4294967295, 9223372036854775807} {
		// These positive values cannot be passed through a 32-bit native int.
		if strconv.IntSize == 32 {
			continue
		}
		for _, daily := range []bool{false, true} {
			for _, durable := range []bool{false, true} {
				for _, fixture := range []string{"absent", "healthy", "corrupt-sidecars", "malformed-recovery", "file-parent"} {
					t.Run(fmt.Sprintf("%d/daily=%v/durable=%v/%s", value, daily, durable, fixture), func(t *testing.T) {
						root := t.TempDir()
						opts := busyBoundsOptions(root, daily, durable)
						path := opts.Path
						if daily {
							path = filepath.Join(opts.Dir, "netsentry-2026-10-03.db")
						}
						var observer *sql.DB
						switch fixture {
						case "healthy":
							seed, err := Open(context.Background(), opts)
							if err != nil {
								t.Fatal(err)
							}
							t.Cleanup(func() { _ = seed.Close() })
							if err := seed.WriteBatch(context.Background(), []*model.Alert{makeAlert(busyBoundsNow, "retained")}); err != nil {
								t.Fatal(err)
							}
							if err := seed.Close(); err != nil {
								t.Fatal(err)
							}
							// Independent encoded read-only connection, warmed before rejection.
							dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: url.Values{"mode": {"ro"}}.Encode()}).String()
							observer, err = sql.Open("sqlite", dsn)
							if err != nil {
								t.Fatal(err)
							}
							observer.SetMaxOpenConns(1)
							t.Cleanup(func() { _ = observer.Close() })
							assertBusyBoundsRetainedRow(t, observer)
						case "corrupt-sidecars":
							for _, suffix := range []string{"", "-wal", "-shm"} {
								writeBusyBoundsFile(t, path+suffix, []byte("retained corrupt input\x00"+suffix))
							}
						case "malformed-recovery":
							opts.RecoveryLogPath = filepath.Join(root, "recovery input.jsonl")
							writeBusyBoundsFile(t, opts.RecoveryLogPath, []byte("{malformed recovery\n"))
						case "file-parent":
							writeBusyBoundsFile(t, filepath.Dir(opts.Dir), []byte("retained parent occupant"))
						}
						opts.BusyTimeoutMS = int(value)
						clockCalls := 0
						opts.Now = func() time.Time { clockCalls++; return busyBoundsNow }
						before := snapshotBusyBoundsTree(t, root)
						store, err := Open(context.Background(), opts)
						if store != nil {
							t.Cleanup(func() { _ = store.Close() })
							t.Fatal("overflow Open returned a Store")
						}
						if err == nil || err.Error() != "sqlite busy timeout must not exceed 2147483647 milliseconds" || clockCalls != 0 {
							t.Fatalf("Open error=%v, clock calls=%d", err, clockCalls)
						}
						if observer != nil {
							assertBusyBoundsRetainedRow(t, observer)
						}
						if after := snapshotBusyBoundsTree(t, root); !reflect.DeepEqual(after, before) {
							t.Fatalf("rejection changed tree bytes/modes/membership: before=%v after=%v", before, after)
						}
					})
				}
			}
		}
	}
}

func TestOpenBusyTimeoutBoundsPreserveValidationPrecedence(t *testing.T) {
	if strconv.IntSize == 32 {
		return // No positive native int can exceed SQLite's maximum.
	}
	for _, cause := range []string{"canceled", "deadline", "durable-journal"} {
		t.Run(cause, func(t *testing.T) {
			root := t.TempDir()
			opts := busyBoundsOptions(root, false, false)
			large := int64(2147483648)
			opts.BusyTimeoutMS = int(large)
			ctx := context.Background()
			var want error
			switch cause {
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = context.Canceled
			case "deadline":
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Unix(0, 0))
				defer cancel()
				want = context.DeadlineExceeded
			case "durable-journal":
				opts.RequireDurableWrites = true
			}
			before := snapshotBusyBoundsTree(t, root)
			store, err := Open(ctx, opts)
			if store != nil {
				t.Cleanup(func() { _ = store.Close() })
				t.Fatal("invalid Open returned a Store")
			}
			if want != nil {
				if err != want || !errors.Is(err, want) {
					t.Fatalf("Open error=%v, want unchanged context error %v", err, want)
				}
			} else if err == nil || err.Error() != "measurement durable writes require WAL journal mode" {
				t.Fatalf("Open error=%v, want existing journal diagnostic", err)
			}
			if after := snapshotBusyBoundsTree(t, root); !reflect.DeepEqual(after, before) {
				t.Fatal("validation rejection changed persistent inputs")
			}
		})
	}
}

func TestOpenBusyTimeoutRepresentableValuesRetainEffectivePragma(t *testing.T) {
	minimumInt := -int(^uint(0)>>1) - 1
	for _, value := range []int{minimumInt, -1, 0, 1, 5000, 2147483647} {
		for _, daily := range []bool{false, true} {
			for _, durable := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/daily=%v/durable=%v", value, daily, durable), func(t *testing.T) {
					opts := busyBoundsOptions(t.TempDir(), daily, durable)
					opts.BusyTimeoutMS = value
					want := value
					if want <= 0 {
						want = 5000
					}
					input := makeAlert(busyBoundsNow, "retained")
					before := *input
					for attempt := 0; attempt < 2; attempt++ {
						store, err := Open(context.Background(), opts)
						if err != nil {
							t.Fatal(err)
						}
						defer store.Close()
						var got int
						if err := store.db.QueryRowContext(context.Background(), "PRAGMA busy_timeout").Scan(&got); err != nil || got != want || store.busyTimeoutMS != want {
							t.Fatalf("live PRAGMA=%d stored=%d error=%v, want %d", got, store.busyTimeoutMS, err, want)
						}
						if attempt == 0 {
							if err := store.WriteBatch(context.Background(), []*model.Alert{input}); err != nil {
								t.Fatal(err)
							}
							if *input != before {
								t.Fatal("write modified caller alert")
							}
						}
						rows, total, err := store.Query(context.Background(), Query{})
						if err != nil || total != 1 || len(rows) != 1 || rows[0].RuleID != input.RuleID || rows[0].MatchedKeyword != input.MatchedKeyword || rows[0].AggregatedCount != 1 {
							t.Fatalf("public Query rows=%+v total=%d error=%v", rows, total, err)
						}
						if err := store.Close(); err != nil {
							t.Fatal(err)
						}
					}
				})
			}
		}
	}
}

func assertBusyBoundsRetainedRow(t *testing.T, db *sql.DB) {
	t.Helper()
	var count int
	var keyword string
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*), MAX(matched_keyword) FROM alerts").Scan(&count, &keyword); err != nil || count != 1 || keyword != "retained" {
		t.Fatalf("independent read-only row count=%d keyword=%q error=%v", count, keyword, err)
	}
}

func writeBusyBoundsFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
}

type busyBoundsTreeEntry struct {
	Mode os.FileMode
	Data []byte
}

func snapshotBusyBoundsTree(t *testing.T, root string) map[string]busyBoundsTreeEntry {
	t.Helper()
	out := make(map[string]busyBoundsTreeEntry)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		item := busyBoundsTreeEntry{Mode: info.Mode()}
		if info.Mode().IsRegular() {
			item.Data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}
		out[rel] = item
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
