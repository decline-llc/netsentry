package pipeline

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type completionObserver struct {
	failStage   string
	panicStage  string
	calls       []string
	onProcessed func()
}

func (o *completionObserver) observe(stage string) error {
	o.calls = append(o.calls, stage)
	if stage == "processed" && o.onProcessed != nil {
		o.onProcessed()
	}
	if stage == o.panicStage {
		panic("observer fault")
	}
	if stage == o.failStage {
		return errors.New("observer fault")
	}
	return nil
}

func (o *completionObserver) Arrival(*model.PacketInfo) error { return o.observe("arrival") }
func (o *completionObserver) Durable(*model.PacketInfo, []*model.Alert, time.Time) error {
	return o.observe("durable")
}
func (o *completionObserver) Processed(*model.PacketInfo, time.Time) error {
	return o.observe("processed")
}

func TestWorkerCompletionTerminalAndFailureBoundaries(t *testing.T) {
	cases := []struct {
		name                         string
		alerts, suppress, observer   bool
		writerError, matcherPanic    bool
		redactorPanic                bool
		failStage, panicStage        string
		completed, generated, writes uint64
		writeErrors, panics          uint64
		persisted                    int
		calls                        []string
	}{
		{name: "no_alerts", completed: 1},
		{name: "suppressed", alerts: true, suppress: true, completed: 1},
		{name: "persisted", alerts: true, completed: 1, generated: 1, writes: 1, persisted: 1},
		{name: "no_alerts_exported", observer: true, completed: 1, calls: []string{"arrival", "processed"}},
		{name: "suppressed_exported", alerts: true, suppress: true, observer: true, completed: 1,
			calls: []string{"arrival", "processed"}},
		{name: "persisted_exported", alerts: true, observer: true, completed: 1, generated: 1, writes: 1, persisted: 1,
			calls: []string{"arrival", "durable", "processed"}},
		{name: "writer_error", alerts: true, writerError: true, writes: 1, writeErrors: 1},
		{name: "arrival_error", alerts: true, observer: true, failStage: "arrival", calls: []string{"arrival"}},
		{name: "durable_error", alerts: true, observer: true, failStage: "durable", writes: 1, persisted: 1,
			calls: []string{"arrival", "durable"}},
		{name: "processed_error_no_alerts", observer: true, failStage: "processed", calls: []string{"arrival", "processed"}},
		{name: "processed_error_suppressed", alerts: true, suppress: true, observer: true, failStage: "processed",
			calls: []string{"arrival", "processed"}},
		{name: "processed_error_persisted", alerts: true, observer: true, failStage: "processed", generated: 1, writes: 1,
			persisted: 1, calls: []string{"arrival", "durable", "processed"}},
		{name: "matcher_panic", matcherPanic: true, panics: 1},
		{name: "redactor_panic", alerts: true, redactorPanic: true, panics: 1},
		{name: "arrival_panic", alerts: true, observer: true, panicStage: "arrival", panics: 1, calls: []string{"arrival"}},
		{name: "durable_panic", alerts: true, observer: true, panicStage: "durable", panics: 1, writes: 1, persisted: 1,
			calls: []string{"arrival", "durable"}},
		{name: "processed_panic", observer: true, panicStage: "processed", panics: 1, calls: []string{"arrival", "processed"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := stats.New()
			matcher := &fakeMatcher{panic: tc.matcherPanic}
			if tc.alerts {
				matcher.alerts = []*model.Alert{{RuleID: "completion", Severity: model.SeverityHigh}}
			}
			writer := &fakeWriter{}
			if tc.writerError {
				writer.err = errors.New("write fault")
			}
			worker := NewWorker(matcher, writer, nil, metrics)
			if tc.suppress {
				worker.SetSuppressor(fakeSuppressor{})
			}
			if tc.redactorPanic {
				worker.SetRedactor(func([]*model.Alert) { panic("redactor fault") })
			}
			observer := &completionObserver{failStage: tc.failStage, panicStage: tc.panicStage}
			observer.onProcessed = func() {
				if got := metrics.Snapshot().PacketsCompleted; got != 0 {
					t.Errorf("count before Processed returns = %d, want 0", got)
				}
			}
			if tc.observer {
				worker.SetObserver(observer)
			}
			packets := make(chan *model.PacketInfo, 1)
			packets <- &model.PacketInfo{TimestampSec: 1}
			close(packets)
			worker.Run(context.Background(), packets)
			snapshot := metrics.Snapshot()
			if snapshot.PacketsCompleted != tc.completed || snapshot.PacketsProcessed != 1 || snapshot.PacketsReceived != 0 {
				t.Fatalf("completion/legacy counters: %+v", snapshot)
			}
			if snapshot.AlertsGenerated != tc.generated || snapshot.AlertWriteCount != tc.writes ||
				snapshot.AlertWriteErrors != tc.writeErrors || snapshot.WorkerPanics != tc.panics {
				t.Fatalf("legacy alert/error counters: %+v", snapshot)
			}
			if len(writer.alerts) != tc.persisted || !reflect.DeepEqual(observer.calls, tc.calls) {
				t.Fatalf("write/export ownership changed: alerts=%d, calls=%v", len(writer.alerts), observer.calls)
			}
		})
	}
}

func TestWorkerCompletionWaitsForProcessedObserverReturn(t *testing.T) {
	metrics := stats.New()
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	worker := NewWorker(&fakeMatcher{}, &fakeWriter{}, nil, metrics)
	worker.SetObserver(&completionObserver{onProcessed: func() {
		close(entered)
		<-release
	}})
	packets := make(chan *model.PacketInfo, 1)
	packets <- &model.PacketInfo{TimestampSec: 1}
	close(packets)
	go func() {
		defer close(done)
		worker.Run(context.Background(), packets)
	}()
	defer func() {
		unblock()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("worker did not return after releasing observer")
		}
	}()
	select {
	case <-entered:
	case <-done:
		t.Fatal("worker returned before terminal observer readiness")
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not reach terminal observer")
	}
	if snapshot := metrics.Snapshot(); snapshot.PacketsCompleted != 0 || snapshot.PacketsProcessed != 1 {
		t.Fatalf("counter while observer is pending: %+v", snapshot)
	}
	unblock()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not return after observer success")
	}
	if got := metrics.Snapshot().PacketsCompleted; got != 1 {
		t.Fatalf("counter after observer success = %d, want 1", got)
	}
}

func TestWorkerCompletionSkipsNilAndPreCancelledEmptyInput(t *testing.T) {
	metrics := stats.New()
	worker := NewWorker(&fakeMatcher{}, &fakeWriter{}, nil, metrics)
	packets := make(chan *model.PacketInfo, 1)
	packets <- nil
	close(packets)
	worker.Run(context.Background(), packets)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Leave input empty and open so cancellation is the only ready select case.
	worker.Run(ctx, make(chan *model.PacketInfo))
	if snapshot := metrics.Snapshot(); snapshot.PacketsCompleted != 0 || snapshot.PacketsProcessed != 0 {
		t.Fatalf("nil/cancelled input counted work: %+v", snapshot)
	}
}

func TestWorkerCompletionCountsConcurrentWorkersAfterJoin(t *testing.T) {
	const total = 1024
	metrics := stats.New()
	packets := make(chan *model.PacketInfo, total)
	for i := 0; i < total; i++ {
		packets <- &model.PacketInfo{TimestampSec: int64(i + 1)}
	}
	close(packets)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		worker := NewWorker(&fakeMatcher{}, &fakeWriter{}, nil, metrics)
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			worker.Run(context.Background(), packets)
		}()
	}
	close(start)
	workers.Wait()
	// Compare after join; active snapshots do not sample all atomics transactionally.
	if snapshot := metrics.Snapshot(); snapshot.PacketsCompleted != total || snapshot.PacketsProcessed != total ||
		snapshot.PacketsReceived != 0 || snapshot.MatchCount != total || snapshot.AlertWriteCount != 0 {
		t.Fatalf("concurrent terminal counts: %+v", snapshot)
	}
}
