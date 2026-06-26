package pipeline

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Worker reads packets, matches rules, and writes generated alerts.
type Worker struct {
	matcher Matcher
	writer  AlertWriter
	logger  *zap.Logger
	now     func() time.Time
}

// NewWorker constructs a single packet processing worker.
func NewWorker(matcher Matcher, writer AlertWriter, logger *zap.Logger) *Worker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Worker{
		matcher: matcher,
		writer:  writer,
		logger:  logger,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

// Run processes packets until the input channel is closed or ctx is cancelled.
func (w *Worker) Run(ctx context.Context, packets <-chan *model.PacketInfo) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("pipeline worker panic", zap.Any("panic", r))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case pkt, ok := <-packets:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			w.processPacket(ctx, pkt)
		}
	}
}

func (w *Worker) processPacket(ctx context.Context, pkt *model.PacketInfo) {
	alerts := w.matcher.Match(pkt)
	if len(alerts) == 0 {
		return
	}

	packetTime := pkt.Timestamp().UTC()
	if packetTime.IsZero() {
		packetTime = w.now()
	}
	for _, alert := range alerts {
		if alert != nil {
			alert.Timestamp = packetTime
		}
	}

	if err := w.writer.WriteBatch(ctx, alerts); err != nil {
		w.logger.Warn("write pipeline alerts", zap.Error(err))
		return
	}
	w.logger.Info("packet matched rules",
		zap.String("src_ip", pkt.SrcIP),
		zap.String("dst_ip", pkt.DstIP),
		zap.Int("alerts", len(alerts)))
}
