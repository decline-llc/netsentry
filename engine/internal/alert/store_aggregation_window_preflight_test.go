package alert

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

var aggregationPreflightNow = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

func aggregationPreflightOptions(root string, daily, durable bool) Options {
	opts := busyBoundsOptions(root, daily, durable)
	opts.Now = func() time.Time { return aggregationPreflightNow }
	return opts
}

func aggregationPreflightPath(opts Options) string {
	if opts.DailyShard {
		// Daily startup ignores Path and selects this resource through Dir/Now.
		return filepath.Join(opts.Dir, "netsentry-2026-10-05.db")
	}
	return opts.Path
}

func aggregationPreflightObserver(t *testing.T, path string) *sql.DB {
	t.Helper()
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	u := url.URL{Scheme: "file", Path: absolute, RawQuery: url.Values{"mode": {"ro"}, "readonly_shm": {"1"}}.Encode()}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close independent observer: %v", err)
		}
	})
	return db
}

// Observe every durable column independently, retaining the handle across calls.
func aggregationPreflightTables(t *testing.T, db *sql.DB) map[string][][]any {
	t.Helper()
	result := map[string][][]any{}
	for _, table := range []string{"alerts", "alert_events"} {
		order := "id"
		if table == "alert_events" {
			order = "event_id"
		}
		rows, err := db.Query("SELECT * FROM " + table + " ORDER BY " + order)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			for i, value := range values {
				if b, ok := value.([]byte); ok {
					values[i] = string(b)
				}
			}
			result[table] = append(result[table], values)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func TestOpenSubsecondAggregationWindowPreservesPersistentInputs(t *testing.T) {
	for _, window := range []time.Duration{time.Nanosecond, time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, time.Second - time.Nanosecond} {
		for _, daily := range []bool{false, true} {
			for _, durable := range []bool{false, true} {
				for _, fixture := range []string{"absent", "healthy", "corrupt-sidecars", "malformed-recovery", "file-parent"} {
					t.Run(fmt.Sprintf("%s/daily=%v/durable=%v/%s", window, daily, durable, fixture), func(t *testing.T) {
						root := t.TempDir()
						opts := aggregationPreflightOptions(root, daily, durable)
						path := aggregationPreflightPath(opts)
						var observer *sql.DB
						var retained map[string][][]any
						switch fixture {
						case "healthy":
							seed, err := Open(context.Background(), opts)
							if err != nil {
								t.Fatal(err)
							}
							t.Cleanup(func() { _ = seed.Close() })
							if seed.Path() != path {
								t.Fatalf("seed Path=%q, want resolved %q", seed.Path(), path)
							}
							if err := seed.WriteBatch(context.Background(), []*model.Alert{makeAlert(aggregationPreflightNow, "retained")}); err != nil {
								t.Fatal(err)
							}
							if err := seed.Close(); err != nil {
								t.Fatal(err)
							}
							observer = aggregationPreflightObserver(t, path)
							assertBusyBoundsRetainedRow(t, observer)
							retained = aggregationPreflightTables(t, observer)
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
						opts.AggregationWindow = window
						clockCalls := 0
						opts.Now = func() time.Time { clockCalls++; return aggregationPreflightNow }
						caller := opts
						before := snapshotBusyBoundsTree(t, root)
						store, err := Open(context.Background(), opts)
						if store != nil {
							t.Cleanup(func() { _ = store.Close() })
							t.Fatal("subsecond Open returned a Store")
						}
						if err == nil || err.Error() != "positive alert aggregation window must be at least one second" || clockCalls != 0 {
							t.Fatalf("Open error=%v, clock calls=%d", err, clockCalls)
						}
						assertJournalCallerOptions(t, opts, caller)
						if observer != nil {
							assertBusyBoundsRetainedRow(t, observer)
							if !reflect.DeepEqual(aggregationPreflightTables(t, observer), retained) {
								t.Fatal("rejection changed durable columns")
							}
						}
						if !reflect.DeepEqual(snapshotBusyBoundsTree(t, root), before) {
							t.Fatal("rejection changed tree bytes/modes/membership")
						}
					})
				}
			}
		}
	}
}

func TestOpenAggregationWindowPreflightPreservesEarlierDiagnostics(t *testing.T) {
	for _, cause := range []string{"canceled", "deadline", "durable-journal", "busy-overflow", "unsupported-journal"} {
		if cause == "busy-overflow" && strconv.IntSize == 32 {
			continue // Positive native int cannot exceed SQLite's signed limit here.
		}
		t.Run(cause, func(t *testing.T) {
			root := t.TempDir()
			opts := aggregationPreflightOptions(root, true, false)
			opts.AggregationWindow = 500 * time.Millisecond
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
			case "durable-journal":
				opts.RequireDurableWrites = true
			case "busy-overflow":
				overflow := int64(2147483648)
				opts.BusyTimeoutMS = int(overflow)
				want = "sqlite busy timeout must not exceed 2147483647 milliseconds"
			case "unsupported-journal":
				opts.JournalMode = "INVALID"
				want = `unsupported sqlite journal mode "INVALID"`
			}
			clockCalls := 0
			opts.Now = func() time.Time { clockCalls++; return aggregationPreflightNow }
			caller := opts
			before := snapshotBusyBoundsTree(t, root)
			store, err := Open(ctx, opts)
			if store != nil {
				t.Cleanup(func() { _ = store.Close() })
				t.Fatal("invalid Open returned a Store")
			}
			if sentinel != nil {
				if err != sentinel || !errors.Is(err, sentinel) {
					t.Fatalf("Open error=%v, want unchanged %v", err, sentinel)
				}
			} else if err == nil || err.Error() != want {
				t.Fatalf("Open error=%v, want %q", err, want)
			}
			assertJournalCallerOptions(t, opts, caller)
			if clockCalls != 0 || !reflect.DeepEqual(snapshotBusyBoundsTree(t, root), before) {
				t.Fatal("earlier rejection called clock or changed artifacts")
			}
		})
	}
}

// Independent expectations retain the durable whole-second row-ID format and
// deterministic event-ID specification; no production normalizer is used.
func aggregationPreflightExpected(input *model.Alert, window time.Duration) *model.Alert {
	out := *input
	out.WindowStart = out.Timestamp.Truncate(window)
	out.ID = fmt.Sprintf("%s-%s-%s-%d-%d", out.RuleID, out.SrcIP, out.DstIP, out.DstPort, out.WindowStart.Unix())
	out.FirstSeen, out.LastSeen, out.AggregatedCount = out.Timestamp, out.Timestamp, 1
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s\x00%s\x00%d", out.RuleID, out.SrcIP, out.DstIP, out.DstPort, out.Protocol, out.MatchedKeyword, out.Timestamp.Format(time.RFC3339Nano), out.Timestamp.UnixNano())))
	out.EventID = fmt.Sprintf("evt_%x", digest[:16])
	return &out
}

func aggregationPreflightExpectedTables(want []*model.Alert) map[string][][]any {
	result := map[string][][]any{}
	created := aggregationPreflightNow.Format(time.RFC3339Nano)
	for _, row := range want {
		result["alerts"] = append(result["alerts"], []any{
			row.ID, row.EventID, row.RuleID, row.RuleName, string(row.Severity), row.Protocol,
			row.SrcIP, row.DstIP, int64(row.DstPort), row.MitreTactic, row.MitreTechniqueID,
			row.MitreTechniqueName, row.PayloadPreview, row.MatchedKeyword, int64(row.AggregatedCount),
			row.FirstSeen.Format(time.RFC3339Nano), row.LastSeen.Format(time.RFC3339Nano),
			row.WindowStart.Format(time.RFC3339Nano), created, created,
		})
		result["alert_events"] = append(result["alert_events"], []any{row.EventID, created})
	}
	for _, table := range []string{"alerts", "alert_events"} {
		sort.Slice(result[table], func(i, j int) bool {
			return result[table][i][0].(string) < result[table][j][0].(string)
		})
	}
	return result
}

func TestOpenAcceptedAggregationWindowsRetainRowsIdentityAndReopen(t *testing.T) {
	for _, window := range []time.Duration{time.Duration(-1 << 63), -time.Nanosecond, 0, time.Second, time.Second + time.Nanosecond, 1500 * time.Millisecond, time.Minute} {
		for _, daily := range []bool{false, true} {
			for _, durable := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/daily=%v/durable=%v", window, daily, durable), func(t *testing.T) {
					ctx := context.Background()
					opts := aggregationPreflightOptions(t.TempDir(), daily, durable)
					opts.AggregationWindow = window
					caller := opts
					effective := window
					if effective <= 0 {
						effective = time.Minute
					}
					first := makeAlert(aggregationPreflightNow.Add(250*time.Millisecond), "first window")
					second := makeAlert(first.Timestamp.Add(effective), "second window")
					input := []*model.Alert{first, second}
					original := []model.Alert{*first, *second}
					want := []*model.Alert{aggregationPreflightExpected(second, effective), aggregationPreflightExpected(first, effective)}
					wantTables := aggregationPreflightExpectedTables(want)
					if want[0].ID == want[1].ID || !want[0].WindowStart.After(want[1].WindowStart) || want[0].EventID == want[1].EventID {
						t.Fatal("accepted fixture did not establish two distinct windows/IDs/events")
					}
					var retained map[string][][]any
					for attempt := 0; attempt < 2; attempt++ {
						store, err := Open(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = store.Close() })
						if store.Path() != aggregationPreflightPath(opts) {
							t.Fatal("Open selected an unexpected resource")
						}
						if attempt == 0 {
							if err := store.WriteBatch(ctx, input); err != nil {
								t.Fatal(err)
							}
						}
						listed, listErr := store.List(ctx)
						queried, total, queryErr := store.Query(ctx, Query{})
						count, countErr := store.Count(ctx)
						if listErr != nil || queryErr != nil || countErr != nil || total != 2 || count != 2 || !reflect.DeepEqual(listed, want) || !reflect.DeepEqual(queried, want) {
							t.Fatalf("public rows: List=%+v Query=%+v total=%d Count=%d errors=%v/%v/%v, want %+v", listed, queried, total, count, listErr, queryErr, countErr, want)
						}
						observer := aggregationPreflightObserver(t, aggregationPreflightPath(opts))
						columns := aggregationPreflightTables(t, observer)
						if !reflect.DeepEqual(columns, wantTables) {
							t.Fatalf("independent complete durable rows/events=%+v, want %+v", columns, wantTables)
						}
						if attempt == 0 {
							retained = columns
						} else if !reflect.DeepEqual(columns, retained) {
							t.Fatal("close/reopen changed durable columns")
						}
						if err := observer.Close(); err != nil {
							t.Fatal(err)
						}
						if store.Health().Status != "ok" {
							t.Fatal("accepted window degraded storage")
						}
						if err := store.Close(); err != nil {
							t.Fatal(err)
						}
						assertJournalCallerOptions(t, opts, caller)
						for i := range input {
							if *input[i] != original[i] {
								t.Fatal("accepted window modified caller alert")
							}
						}
					}
				})
			}
		}
	}
}
