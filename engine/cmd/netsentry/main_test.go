package main

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

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
