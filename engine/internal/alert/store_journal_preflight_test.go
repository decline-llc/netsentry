package alert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestOpenInvalidJournalModePreservesPersistentInputs(t *testing.T) {
	for _, value := range []string{"INVALID", "WALS", " invalid ", "WAL; PRAGMA user_version=7"} {
		for _, daily := range []bool{false, true} {
			for _, fixture := range []string{"absent", "healthy", "corrupt-sidecars", "malformed-recovery", "file-parent"} {
				t.Run(fmt.Sprintf("%q/daily=%v/%s", value, daily, fixture), func(t *testing.T) {
					root := t.TempDir()
					opts := busyBoundsOptions(root, daily, false)
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
					opts.JournalMode = value
					clockCalls := 0
					opts.Now = func() time.Time { clockCalls++; return busyBoundsNow }
					caller := opts
					before := snapshotBusyBoundsTree(t, root)
					store, err := Open(context.Background(), opts)
					if store != nil {
						t.Cleanup(func() { _ = store.Close() })
						t.Fatal("invalid journal Open returned a Store")
					}
					if err == nil || err.Error() != fmt.Sprintf("unsupported sqlite journal mode %q", value) || clockCalls != 0 {
						t.Fatalf("Open error=%v, clock calls=%d", err, clockCalls)
					}
					assertJournalCallerOptions(t, opts, caller)
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

func TestOpenJournalPreflightPreservesEarlierDiagnostics(t *testing.T) {
	for _, cause := range []string{"canceled", "deadline", "durable-invalid", "durable-blank", "durable-supported-non-WAL", "busy-overflow"} {
		if cause == "busy-overflow" && strconv.IntSize == 32 {
			continue // This positive overflow cannot be passed through a native int.
		}
		t.Run(cause, func(t *testing.T) {
			root := t.TempDir()
			opts := busyBoundsOptions(root, false, false)
			opts.JournalMode = "INVALID"
			ctx := context.Background()
			var sentinel error
			want := "measurement durable writes require WAL journal mode"
			switch cause {
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				sentinel = context.Canceled
			case "deadline":
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Unix(0, 0))
				defer cancel()
				sentinel = context.DeadlineExceeded
			case "durable-invalid":
				opts.RequireDurableWrites = true
			case "durable-blank":
				opts.RequireDurableWrites, opts.JournalMode = true, " \t "
			case "durable-supported-non-WAL":
				opts.RequireDurableWrites, opts.JournalMode = true, " delete "
			case "busy-overflow":
				large := int64(2147483648)
				opts.BusyTimeoutMS = int(large)
				want = "sqlite busy timeout must not exceed 2147483647 milliseconds"
			}
			clockCalls := 0
			opts.Now = func() time.Time { clockCalls++; return busyBoundsNow }
			caller := opts
			before := snapshotBusyBoundsTree(t, root)
			store, err := Open(ctx, opts)
			if store != nil {
				t.Cleanup(func() { _ = store.Close() })
				t.Fatal("invalid Open returned a Store")
			}
			if sentinel != nil {
				if err != sentinel || !errors.Is(err, sentinel) {
					t.Fatalf("Open error=%v, want original %v", err, sentinel)
				}
			} else if err == nil || err.Error() != want {
				t.Fatalf("Open error=%v, want %q", err, want)
			}
			assertJournalCallerOptions(t, opts, caller)
			if clockCalls != 0 || !reflect.DeepEqual(snapshotBusyBoundsTree(t, root), before) {
				t.Fatal("earlier diagnostic reached clock or changed filesystem")
			}
		})
	}
}

func TestOpenSupportedJournalModesRetainPragmaAndReopen(t *testing.T) {
	cases := []struct {
		mode    string
		durable bool
	}{
		{"", false}, {" \t ", false}, {"DELETE", false}, {"TRUNCATE", false},
		{"PERSIST", false}, {"MEMORY", false}, {"WAL", false}, {"OFF", false},
		{" delete ", false}, {"truncate", false}, {" persist ", false},
		{"memory", false}, {" wal ", false}, {"off", false},
		{"", true}, {" wal \t", true},
	}
	for _, tc := range cases {
		for _, daily := range []bool{false, true} {
			t.Run(fmt.Sprintf("%q/daily=%v/durable=%v", tc.mode, daily, tc.durable), func(t *testing.T) {
				opts := busyBoundsOptions(t.TempDir(), daily, tc.durable)
				opts.JournalMode = tc.mode
				caller := opts
				want := strings.ToLower(strings.TrimSpace(tc.mode))
				if want == "" {
					want = "wal"
				}
				input := makeAlert(busyBoundsNow, "retained")
				alertBefore := *input
				for attempt := 0; attempt < 2; attempt++ {
					store, err := Open(context.Background(), opts)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = store.Close() })
					var mode string
					if err := store.db.QueryRowContext(context.Background(), "PRAGMA journal_mode").Scan(&mode); err != nil || mode != want {
						t.Fatalf("live PRAGMA=%q error=%v, want %q", mode, err, want)
					}
					if attempt == 0 {
						if err := store.WriteBatch(context.Background(), []*model.Alert{input}); err != nil {
							t.Fatal(err)
						}
						if *input != alertBefore {
							t.Fatal("write modified caller alert")
						}
					}
					rows, total, err := store.Query(context.Background(), Query{})
					if err != nil || total != 1 || len(rows) != 1 || rows[0].RuleID != input.RuleID || rows[0].MatchedKeyword != input.MatchedKeyword || rows[0].AggregatedCount != 1 {
						t.Fatalf("public Query rows=%+v total=%d error=%v", rows, total, err)
					}
					assertJournalCallerOptions(t, opts, caller)
					if err := store.Close(); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func assertJournalCallerOptions(t *testing.T, got, want Options) {
	t.Helper()
	if reflect.ValueOf(got.Now).Pointer() != reflect.ValueOf(want.Now).Pointer() {
		t.Fatal("Open changed caller clock")
	}
	got.Now, want.Now = nil, nil
	if !reflect.DeepEqual(got, want) {
		t.Fatal("Open changed caller options")
	}
}
