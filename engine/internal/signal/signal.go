package signal

import (
	"context"
	"os"
	ossignal "os/signal"
	"syscall"
)

// WaitForShutdown blocks until SIGINT or SIGTERM is received, then cancels
// the returned context. Callers should use the context to drive graceful shutdown.
func WaitForShutdown() (context.Context, context.CancelFunc) {
	return ossignal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
