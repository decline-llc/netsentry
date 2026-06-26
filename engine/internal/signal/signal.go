package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdown blocks until SIGINT or SIGTERM is received, then cancels
// the returned context. Callers should use the context to drive graceful shutdown.
func WaitForShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()
	return ctx, cancel
}
