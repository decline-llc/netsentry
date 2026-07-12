package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/config"
	"github.com/decline-llc/netsentry/internal/pipeline"
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
