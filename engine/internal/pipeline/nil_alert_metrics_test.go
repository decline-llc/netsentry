package pipeline

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	alertpkg "github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func nilMetricFixtures() []*model.Alert {
	return []*model.Alert{
		nil,
		{RuleID: "nil-metrics-high", RuleName: "High fixture", Severity: model.SeverityHigh,
			SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, Protocol: "TCP", MatchedKeyword: "high"},
		nil,
		{RuleID: "nil-metrics-low", RuleName: "Low fixture", Severity: model.SeverityLow,
			SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, Protocol: "TCP", MatchedKeyword: "low"},
		nil,
	}
}

func assertWorkerNilMetricSnapshot(t *testing.T, snapshot stats.Snapshot, generated, completed, writeErrors uint64) {
	t.Helper()
	want := map[model.Severity]uint64{
		model.SeverityLow: generated / 2, model.SeverityMedium: 0, model.SeverityHigh: generated / 2, model.SeverityCritical: 0,
	}
	if snapshot.AlertsGenerated != generated || !reflect.DeepEqual(snapshot.AlertsBySeverity, want) ||
		snapshot.PacketsProcessed != 1 || snapshot.PacketsCompleted != completed || snapshot.PacketsReceived != 0 ||
		snapshot.MatchCount != 1 || snapshot.AlertWriteCount != 1 || snapshot.AlertWriteErrors != writeErrors || snapshot.WorkerPanics != 0 {
		t.Fatalf("worker metrics=%+v, want generated=%d severity=%v completed=%d writeErrors=%d", snapshot, generated, want, completed, writeErrors)
	}
}

func TestWorkerNilAlertMetricsMatchRealStoreRows(t *testing.T) {
	for _, allNil := range []bool{false, true} {
		name := "mixed"
		if allNil {
			name = "all_nil"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			store, err := alertpkg.Open(ctx, alertpkg.Options{
				Path:        filepath.Join(t.TempDir(), "nil metric fixtures", "alerts.db"),
				JournalMode: "DELETE", AggregationWindow: time.Minute,
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("close store: %v", err)
				}
			})
			batch := nilMetricFixtures()
			wantRows := 2
			if allNil {
				batch, wantRows = []*model.Alert{nil, nil}, 0
			}
			metrics := stats.New()
			worker := NewWorker(&fakeMatcher{alerts: batch}, store, nil, metrics)
			packet := &model.PacketInfo{TimestampSec: 1791028800, TimestampUsec: 123456, SrcIP: "192.0.2.1", DstIP: "198.51.100.1"}
			before := *packet
			packets := make(chan *model.PacketInfo, 1)
			packets <- packet
			close(packets)
			worker.Run(ctx, packets)
			if *packet != before {
				t.Fatal("Worker changed packet input")
			}
			assertWorkerNilMetricSnapshot(t, metrics.Snapshot(), uint64(wantRows), 1, 0)
			rows, total, err := store.Query(ctx, alertpkg.Query{Limit: 10})
			if err != nil || total != wantRows || len(rows) != wantRows {
				t.Fatalf("real rows=%d total=%d error=%v, want %d", len(rows), total, err, wantRows)
			}
			for i, row := range rows {
				fixture := batch[1+2*i]
				if row == nil || row.RuleID != fixture.RuleID || row.Severity != fixture.Severity ||
					row.MatchedKeyword != fixture.MatchedKeyword || row.AggregatedCount != 1 || !row.LastSeen.Equal(packet.Timestamp().UTC()) {
					t.Fatalf("real ordered row[%d]=%+v, want fixture %+v", i, row, fixture)
				}
			}
			count, err := store.Count(ctx)
			if err != nil || count != wantRows || store.Health().Status != "ok" {
				t.Fatalf("real count=%d error=%v health=%+v", count, err, store.Health())
			}
		})
	}
}

func TestWorkerNilAlertMetricsPreserveFailureAndExportGates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		writerFail bool
		failStage  string
		generated  uint64
		completed  uint64
		writeError uint64
		written    int
		calls      []string
	}{
		{"writer_failure", true, "", 0, 0, 1, 0, []string{"arrival"}},
		{"durable_export_failure", false, "durable", 0, 0, 0, 5, []string{"arrival", "durable"}},
		{"processed_export_failure", false, "processed", 2, 0, 0, 5, []string{"arrival", "durable", "processed"}},
		{"export_success", false, "", 2, 1, 0, 5, []string{"arrival", "durable", "processed"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metrics := stats.New()
			batch := nilMetricFixtures()
			writer := &fakeWriter{}
			if tc.writerFail {
				writer.err = errors.New("injected write rejection")
			}
			worker := NewWorker(&fakeMatcher{alerts: batch}, writer, nil, metrics)
			observer := &completionObserver{failStage: tc.failStage}
			worker.SetObserver(observer)
			packets := make(chan *model.PacketInfo, 1)
			packets <- &model.PacketInfo{TimestampSec: 1791028800}
			close(packets)
			worker.Run(context.Background(), packets)
			assertWorkerNilMetricSnapshot(t, metrics.Snapshot(), tc.generated, tc.completed, tc.writeError)
			if len(writer.alerts) != tc.written || !reflect.DeepEqual(observer.calls, tc.calls) {
				t.Fatalf("writer/export contract: batch entries=%d calls=%v, want %d/%v", len(writer.alerts), observer.calls, tc.written, tc.calls)
			}
			for i, written := range writer.alerts {
				if written != batch[i] {
					t.Fatalf("writer input[%d]=%p, want original %p", i, written, batch[i])
				}
			}
		})
	}
}
