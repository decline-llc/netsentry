package signal

import (
	"context"
	"testing"
	"time"
)

func TestWaitForShutdownCancelStopsContext(t *testing.T) {
	ctx, cancel := WaitForShutdown()
	cancel()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("shutdown context was not cancelled")
	}
	if err := ctx.Err(); err != context.Canceled {
		t.Fatalf("unexpected context error: %v", err)
	}
}
