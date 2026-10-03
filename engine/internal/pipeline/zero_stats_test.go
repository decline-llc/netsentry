package pipeline_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/pipeline"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestWorkerZeroStatsCompletesAfterRealStoreWrite(t *testing.T) {
	ctx := context.Background()
	store, err := alert.Open(ctx, alert.Options{
		Path:        filepath.Join(t.TempDir(), "zero stats rows", "alerts.db"),
		JournalMode: "DELETE", AggregationWindow: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	batch := []*model.Alert{
		nil,
		{RuleID: "zero-stats-high", RuleName: "High fixture", Severity: model.SeverityHigh,
			SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, Protocol: "TCP", MatchedKeyword: "high"},
		nil,
		{RuleID: "zero-stats-low", RuleName: "Low fixture", Severity: model.SeverityLow,
			SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, Protocol: "TCP", MatchedKeyword: "low"},
	}
	var metrics stats.Stats
	matcher := &zeroStatsMatcher{alerts: batch}
	worker := pipeline.NewWorker(matcher, store, nil, &metrics)
	packet := &model.PacketInfo{TimestampSec: 1791028800, TimestampUsec: 123456, SrcIP: "192.0.2.1", DstIP: "198.51.100.1"}
	before := *packet
	packets := make(chan *model.PacketInfo, 2)
	packets <- nil
	packets <- packet
	close(packets)
	worker.Run(ctx, packets)
	if *packet != before || matcher.calls != 1 || matcher.packet != packet {
		t.Fatalf("packet/matcher contract changed: packet=%+v calls=%d", packet, matcher.calls)
	}
	snapshot := metrics.Snapshot()
	wantCounts := map[model.Severity]uint64{model.SeverityHigh: 1, model.SeverityLow: 1}
	if snapshot.AlertsGenerated != 2 || !reflect.DeepEqual(snapshot.AlertsBySeverity, wantCounts) ||
		snapshot.PacketsProcessed != 1 || snapshot.PacketsCompleted != 1 || snapshot.PacketsReceived != 0 ||
		snapshot.WorkerPanics != 0 || snapshot.AlertWriteErrors != 0 || snapshot.MatchCount != 1 || snapshot.AlertWriteCount != 1 || !snapshot.StartedAt.IsZero() {
		t.Fatalf("worker zero Stats=%+v", snapshot)
	}
	rows, total, err := store.Query(ctx, alert.Query{Limit: -1})
	if err != nil || total != 2 || len(rows) != 2 {
		t.Fatalf("stored rows=%d total=%d err=%v, want two", len(rows), total, err)
	}
	wanted := map[string]*model.Alert{batch[1].RuleID: batch[1], batch[3].RuleID: batch[3]}
	for _, row := range rows {
		if row == nil {
			t.Fatal("nil stored row")
		}
		fixture := wanted[row.RuleID]
		if fixture == nil || row.Severity != fixture.Severity || row.MatchedKeyword != fixture.MatchedKeyword ||
			row.AggregatedCount != 1 || !row.LastSeen.Equal(packet.Timestamp().UTC()) {
			t.Fatalf("stored row=%+v, want matching fixture %+v", row, fixture)
		}
		delete(wanted, row.RuleID)
	}
	count, err := store.Count(ctx)
	if err != nil || count != 2 || len(wanted) != 0 || store.Health().Status != "ok" {
		t.Fatalf("store count=%d err=%v missing=%v health=%+v", count, err, wanted, store.Health())
	}
}

func TestWorkerZeroStatsPreservesNoAlertAndFailedWriteGates(t *testing.T) {
	for _, failedWrite := range []bool{false, true} {
		name := "no_alert"
		if failedWrite {
			name = "writer_failure"
		}
		t.Run(name, func(t *testing.T) {
			var metrics stats.Stats
			matcher := &zeroStatsMatcher{}
			writer := &zeroStatsWriter{}
			wantCompleted, wantErrors, wantWrites := uint64(1), uint64(0), 0
			if failedWrite {
				matcher.alerts = []*model.Alert{{RuleID: "write-failure", Severity: model.SeverityHigh}}
				writer.err = errors.New("fixture write rejection")
				wantCompleted, wantErrors, wantWrites = 0, 1, 1
			}
			worker := pipeline.NewWorker(matcher, writer, nil, &metrics)
			packets := make(chan *model.PacketInfo, 1)
			packets <- &model.PacketInfo{TimestampSec: 1791028800}
			close(packets)
			worker.Run(context.Background(), packets)
			snapshot := metrics.Snapshot()
			if snapshot.PacketsProcessed != 1 || snapshot.PacketsCompleted != wantCompleted || snapshot.AlertsGenerated != 0 ||
				snapshot.AlertWriteErrors != wantErrors || snapshot.WorkerPanics != 0 || snapshot.MatchCount != 1 ||
				snapshot.AlertWriteCount != uint64(wantWrites) || len(snapshot.AlertsBySeverity) != 0 || !snapshot.StartedAt.IsZero() ||
				writer.calls != wantWrites || matcher.calls != 1 {
				t.Fatalf("gate stats=%+v writer calls=%d matcher calls=%d", snapshot, writer.calls, matcher.calls)
			}
		})
	}
}

type zeroStatsMatcher struct {
	alerts []*model.Alert
	calls  int
	packet *model.PacketInfo
}

func (m *zeroStatsMatcher) Match(packet *model.PacketInfo) []*model.Alert {
	m.calls++
	m.packet = packet
	return m.alerts
}

type zeroStatsWriter struct {
	calls int
	err   error
}

func (w *zeroStatsWriter) WriteBatch(context.Context, []*model.Alert) error {
	w.calls++
	return w.err
}
