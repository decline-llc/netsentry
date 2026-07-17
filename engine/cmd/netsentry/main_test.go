package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/config"
	"github.com/decline-llc/netsentry/internal/pipeline"
	"github.com/decline-llc/netsentry/internal/receiver"
	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type noOpMatcher struct{}

func (noOpMatcher) Match(*model.PacketInfo) []*model.Alert {
	return nil
}

type noOpAlertWriter struct{}

func (noOpAlertWriter) WriteBatch(context.Context, []*model.Alert) error {
	return nil
}

func TestStartPipelineWorkerClosesDoneAfterContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	packets := make(chan *model.PacketInfo)
	worker := pipeline.NewWorker(noOpMatcher{}, noOpAlertWriter{}, zap.NewNop())

	done := startPipelineWorker(ctx, worker, packets)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pipeline worker did not stop after context cancellation")
	}
}

type countingMatcher struct {
	calls atomic.Int64
}

func (m *countingMatcher) Match(*model.PacketInfo) []*model.Alert {
	m.calls.Add(1)
	return nil
}

func TestStartPipelineWorkersDrainInput(t *testing.T) {
	matcher := &countingMatcher{}
	worker := pipeline.NewWorker(matcher, noOpAlertWriter{}, zap.NewNop())
	packets := make(chan *model.PacketInfo, 32)
	for i := 0; i < 32; i++ {
		packets <- &model.PacketInfo{}
	}
	close(packets)

	done := startPipelineWorkers(context.Background(), worker, packets, 4)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pipeline worker pool did not drain closed input")
	}
	if got := matcher.calls.Load(); got != 32 {
		t.Fatalf("match calls = %d, want 32", got)
	}
}

func TestNewHTTPServerUsesSecureResourceLimits(t *testing.T) {
	srv := newHTTPServer(config.EngineConfig{
		APIListenHost: "127.0.0.1",
		APIPort:       9090,
	}, nil, nil, nil, nil, nil, zap.NewNop())
	if srv.Addr != "127.0.0.1:9090" {
		t.Fatalf("server address = %q", srv.Addr)
	}
	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Fatalf("server timeouts must all be bounded: %+v", srv)
	}
	if srv.MaxHeaderBytes != 16<<10 {
		t.Fatalf("MaxHeaderBytes = %d, want %d", srv.MaxHeaderBytes, 16<<10)
	}
}

func TestStartHTTPServerReturnsBindError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = startHTTPServer(ctx, config.EngineConfig{
		APIListenHost: "127.0.0.1",
		APIPort:       ln.Addr().(*net.TCPAddr).Port,
	}, nil, nil, nil, nil, nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected occupied HTTP address to fail synchronously")
	}
}

func TestServeHTTPServerStopsWhenContextAlreadyCanceled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: http.NewServeMux(), ReadHeaderTimeout: time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := serveHTTPServer(ctx, srv, ln, zap.NewNop())
	select {
	case <-done:
	case <-time.After(time.Second):
		_ = srv.Close()
		t.Fatal("HTTP server did not stop when context was already canceled")
	}
}

type stagedShutdownMatcher struct {
	calls         atomic.Int64
	secondStarted chan struct{}
	releaseSecond chan struct{}
	releaseOnce   sync.Once
}

func newStagedShutdownMatcher() *stagedShutdownMatcher {
	return &stagedShutdownMatcher{
		secondStarted: make(chan struct{}),
		releaseSecond: make(chan struct{}),
	}
}

func (m *stagedShutdownMatcher) Match(pkt *model.PacketInfo) []*model.Alert {
	if m.calls.Add(1) == 2 {
		close(m.secondStarted)
		<-m.releaseSecond
	}
	return []*model.Alert{{
		RuleID:         "shutdown-rule",
		RuleName:       "shutdown drill",
		Severity:       model.SeverityHigh,
		Protocol:       "tcp",
		SrcIP:          pkt.SrcIP,
		DstIP:          pkt.DstIP,
		DstPort:        pkt.DstPort,
		MatchedKeyword: "shutdown",
	}}
}

func (m *stagedShutdownMatcher) release() {
	m.releaseOnce.Do(func() { close(m.releaseSecond) })
}

type shutdownTrackingWriter struct {
	store            *alert.Store
	closed           atomic.Bool
	writesAfterClose atomic.Int64
	firstWriteDone   chan struct{}
	firstWriteOnce   sync.Once
}

func newShutdownTrackingWriter(store *alert.Store) *shutdownTrackingWriter {
	return &shutdownTrackingWriter{store: store, firstWriteDone: make(chan struct{})}
}

func (w *shutdownTrackingWriter) WriteBatch(ctx context.Context, alerts []*model.Alert) error {
	if w.closed.Load() {
		w.writesAfterClose.Add(1)
		return errors.New("write attempted after store close")
	}
	err := w.store.WriteBatch(ctx, alerts)
	if err == nil {
		w.firstWriteOnce.Do(func() { close(w.firstWriteDone) })
	}
	return err
}

func (w *shutdownTrackingWriter) closeStore() error {
	w.closed.Store(true)
	return w.store.Close()
}

func TestActiveLoadFullEngineShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	store, err := alert.Open(context.Background(), alert.Options{
		Path:              filepath.Join(dir, "alerts.db"),
		JournalMode:       "WAL",
		AggregationWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	writer := newShutdownTrackingWriter(store)
	defer func() {
		if !writer.closed.Load() {
			_ = writer.closeStore()
		}
	}()

	metrics := stats.New()
	udsPath := filepath.Join(dir, "netsentry.sock")
	recv := receiver.New(receiver.Config{
		Path:           udsPath,
		MaxConnections: 2,
		BufferSize:     4,
		Stats:          metrics,
	}, zap.NewNop())
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("start UDS receiver: %v", err)
	}
	defer func() {
		recv.Stop()
		recv.Wait()
	}()

	matcher := newStagedShutdownMatcher()
	defer matcher.release()
	worker := pipeline.NewWorker(matcher, writer, zap.NewNop(), metrics)
	workerDone := startPipelineWorker(ctx, worker, recv.Packets())

	srv := newHTTPServer(config.EngineConfig{
		APIListenHost:               "127.0.0.1",
		APIPort:                     8080,
		HealthFreshnessLimitSeconds: 30,
	}, store, recv, rule.NewEngine(), metrics, nil, zap.NewNop())
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for HTTP API: %v", err)
	}
	srv.Addr = ln.Addr().String()
	httpDone := serveHTTPServer(ctx, srv, ln, zap.NewNop())
	defer srv.Close()

	conn, err := dialShutdownUDS(udsPath)
	if err != nil {
		t.Fatalf("dial UDS receiver: %v", err)
	}
	defer conn.Close()
	writeShutdownFrame(t, conn, map[string]any{
		"type": "hello", "version": "0.1.0", "session_id": "shutdown-drill",
		"pid": 1, "hostname": "test", "max_payload_len": 4096,
	})
	writeShutdownPacket(t, conn, 1)

	select {
	case <-writer.firstWriteDone:
	case <-time.After(2 * time.Second):
		t.Fatal("first active-load alert was not persisted")
	}

	client := &http.Client{
		Timeout:   time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := client.Get("http://" + srv.Addr + "/api/alerts")
	if err != nil {
		t.Fatalf("query active HTTP API: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("active HTTP API status = %d, want 200", resp.StatusCode)
	}

	writeShutdownPacket(t, conn, 2)
	select {
	case <-matcher.secondStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("second packet did not reach the in-flight matcher")
	}

	cancel()
	shutdownDone := make(chan struct{})
	go func() {
		waitForEngineShutdown(recv, workerDone, httpDone)
		close(shutdownDone)
	}()
	select {
	case <-workerDone:
		t.Fatal("worker exited while its matcher was still in flight")
	default:
	}

	matcher.release()
	select {
	case <-shutdownDone:
	case <-time.After(3 * time.Second):
		t.Fatal("full-engine shutdown exceeded its bounded timeout")
	}

	count, err := store.Count(context.Background())
	if err != nil {
		t.Fatalf("count persisted alerts before store close: %v", err)
	}
	if count != 1 {
		t.Fatalf("persisted alert count = %d, want 1", count)
	}
	if err := writer.closeStore(); err != nil {
		t.Fatalf("close SQLite store: %v", err)
	}
	if got := writer.writesAfterClose.Load(); got != 0 {
		t.Fatalf("writes after store close = %d, want 0", got)
	}
	if _, err := os.Stat(udsPath); !os.IsNotExist(err) {
		t.Fatalf("UDS path remains after shutdown: %v", err)
	}
	if _, err := client.Get("http://" + srv.Addr + "/api/health"); err == nil {
		t.Fatal("HTTP API accepted a request after shutdown")
	}
}

func dialShutdownUDS(path string) (net.Conn, error) {
	deadline := time.Now().Add(time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", path)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}
	return nil, lastErr
}

func writeShutdownPacket(t *testing.T, conn net.Conn, timestamp int64) {
	t.Helper()
	writeShutdownFrame(t, conn, map[string]any{
		"timestamp_sec": timestamp, "timestamp_usec": 0,
		"src_ip": "10.0.0.1", "dst_ip": "10.0.0.2",
		"src_port": 12345, "dst_port": 443, "protocol": 6,
		"payload_len": 0, "payload_preview": "",
		"is_fragment": false, "truncated": false,
	})
}

func writeShutdownFrame(t *testing.T, conn net.Conn, frame any) {
	t.Helper()
	encoded, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal UDS frame: %v", err)
	}
	if _, err := conn.Write(append(encoded, '\n')); err != nil {
		t.Fatalf("write UDS frame: %v", err)
	}
}
