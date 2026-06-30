package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type fakeMatcher struct {
	alerts []*model.Alert
	panic  bool
}

func (m *fakeMatcher) Match(pkt *model.PacketInfo) []*model.Alert {
	if m.panic {
		panic("matcher failed")
	}
	return m.alerts
}

type sequenceMatcher struct {
	calls int
}

func (m *sequenceMatcher) Match(pkt *model.PacketInfo) []*model.Alert {
	m.calls++
	if m.calls == 1 {
		panic("first packet failed")
	}
	return []*model.Alert{{RuleID: "rule-after-panic"}}
}

type fakeWriter struct {
	alerts []*model.Alert
	err    error
}

type fakeSuppressor struct {
	alerts []*model.Alert
}

func (s fakeSuppressor) Filter(alerts []*model.Alert) []*model.Alert {
	return s.alerts
}

func (w *fakeWriter) WriteBatch(ctx context.Context, alerts []*model.Alert) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if w.err != nil {
		return w.err
	}
	w.alerts = append(w.alerts, alerts...)
	return nil
}

func TestWorkerWritesAlertsWithPacketTimestamp(t *testing.T) {
	metrics := stats.New()
	writer := &fakeWriter{}
	matcher := &fakeMatcher{alerts: []*model.Alert{{RuleID: "rule-1"}}}
	worker := NewWorker(matcher, writer, zap.NewNop(), metrics)

	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1719300000, TimestampUsec: 123456, SrcIP: "10.0.0.1", DstIP: "10.0.0.2"}
	close(packets)

	worker.Run(context.Background(), packets)
	if len(writer.alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(writer.alerts))
	}
	if writer.alerts[0].Timestamp.IsZero() || writer.alerts[0].Timestamp.Nanosecond() != 123456000 {
		t.Fatalf("unexpected timestamp: %s", writer.alerts[0].Timestamp)
	}
	if got := metrics.Snapshot().AlertWriteCount; got != 1 {
		t.Fatalf("alert write observations = %d, want 1", got)
	}
}

func TestWorkerFiltersSuppressedAlerts(t *testing.T) {
	writer := &fakeWriter{}
	matcher := &fakeMatcher{alerts: []*model.Alert{{RuleID: "suppressed"}, {RuleID: "kept"}}}
	worker := NewWorker(matcher, writer, zap.NewNop())
	worker.SetSuppressor(fakeSuppressor{alerts: []*model.Alert{{RuleID: "kept"}}})

	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1}
	close(packets)

	worker.Run(context.Background(), packets)
	if len(writer.alerts) != 1 || writer.alerts[0].RuleID != "kept" {
		t.Fatalf("unexpected written alerts: %+v", writer.alerts)
	}
}

func TestWorkerSkipsWriteWhenAllAlertsSuppressed(t *testing.T) {
	writer := &fakeWriter{}
	matcher := &fakeMatcher{alerts: []*model.Alert{{RuleID: "suppressed"}}}
	worker := NewWorker(matcher, writer, zap.NewNop())
	worker.SetSuppressor(fakeSuppressor{})

	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1}
	close(packets)

	worker.Run(context.Background(), packets)
	if len(writer.alerts) != 0 {
		t.Fatalf("expected no written alerts, got %+v", writer.alerts)
	}
}

func TestWorkerStopsOnContextCancel(t *testing.T) {
	worker := NewWorker(&fakeMatcher{}, &fakeWriter{}, zap.NewNop())
	packets := make(chan *model.PacketInfo)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, packets)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func TestWorkerHandlesWriterError(t *testing.T) {
	writer := &fakeWriter{err: errors.New("disk full")}
	matcher := &fakeMatcher{alerts: []*model.Alert{{RuleID: "rule-1"}}}
	worker := NewWorker(matcher, writer, zap.NewNop())
	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1}
	close(packets)

	worker.Run(context.Background(), packets)
	if len(writer.alerts) != 0 {
		t.Fatalf("writer should not store alerts on error")
	}
}

func TestWorkerRecoversMatcherPanic(t *testing.T) {
	worker := NewWorker(&fakeMatcher{panic: true}, &fakeWriter{}, zap.NewNop())
	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1}
	close(packets)

	worker.Run(context.Background(), packets)
}

func TestWorkerContinuesAfterPacketPanic(t *testing.T) {
	metrics := stats.New()
	writer := &fakeWriter{}
	worker := NewWorker(&sequenceMatcher{}, writer, zap.NewNop(), metrics)
	packets := make(chan *model.PacketInfo, 2)
	packets <- &model.PacketInfo{TimestampSec: 1, SrcIP: "10.0.0.1", DstIP: "10.0.0.2"}
	packets <- &model.PacketInfo{TimestampSec: 2, SrcIP: "10.0.0.3", DstIP: "10.0.0.4"}
	close(packets)

	worker.Run(context.Background(), packets)

	if len(writer.alerts) != 1 || writer.alerts[0].RuleID != "rule-after-panic" {
		t.Fatalf("worker did not continue after panic, alerts=%+v", writer.alerts)
	}
	if got := metrics.Snapshot().WorkerPanics; got != 1 {
		t.Fatalf("expected 1 worker panic, got %d", got)
	}
}

func TestWorkerRedactsAlertsBeforeWrite(t *testing.T) {
	writer := &fakeWriter{}
	matcher := &fakeMatcher{alerts: []*model.Alert{{RuleID: "rule-1", PayloadPreview: "password=secret"}}}
	worker := NewWorker(matcher, writer, zap.NewNop())
	worker.SetRedactor(func(alerts []*model.Alert) {
		for _, alert := range alerts {
			alert.PayloadPreview = "redacted:" + alert.PayloadPreview
		}
	})

	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1}
	close(packets)

	worker.Run(context.Background(), packets)
	if len(writer.alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(writer.alerts))
	}
	if writer.alerts[0].PayloadPreview != "redacted:password=secret" {
		t.Fatalf("unexpected payload preview: %s", writer.alerts[0].PayloadPreview)
	}
}
