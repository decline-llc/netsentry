package pipeline

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

// Worker reads packets, matches rules, and writes generated alerts.
type Worker struct {
	matcher    Matcher
	writer     AlertWriter
	logger     *zap.Logger
	now        func() time.Time
	stats      *stats.Stats
	suppressor SuppressionFilter
	redactor   AlertRedactor
}

// NewWorker constructs a single packet processing worker.
func NewWorker(matcher Matcher, writer AlertWriter, logger *zap.Logger, statsOpt ...*stats.Stats) *Worker {
	if logger == nil {
		logger = zap.NewNop()
	}
	var metrics *stats.Stats
	if len(statsOpt) > 0 {
		metrics = statsOpt[0]
	}
	return &Worker{
		matcher: matcher,
		writer:  writer,
		logger:  logger,
		now:     func() time.Time { return time.Now().UTC() },
		stats:   metrics,
	}
}

// SetSuppressor configures an optional alert suppression filter.
func (w *Worker) SetSuppressor(filter SuppressionFilter) {
	w.suppressor = filter
}

// SetRedactor configures an optional alert payload redactor.
func (w *Worker) SetRedactor(redactor AlertRedactor) {
	w.redactor = redactor
}

// Run processes packets until the input channel is closed or ctx is cancelled.
func (w *Worker) Run(ctx context.Context, packets <-chan *model.PacketInfo) {
	defer func() {
		if r := recover(); r != nil {
			w.stats.IncWorkerPanic()
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
	defer func() {
		if r := recover(); r != nil {
			w.stats.IncWorkerPanic()
			w.logger.Error("pipeline packet panic", zap.Any("panic", r), zap.String("src_ip", pkt.SrcIP), zap.String("dst_ip", pkt.DstIP))
		}
	}()

	w.stats.IncPacketProcessed()
	start := time.Now()
	alerts := w.matcher.Match(pkt)
	w.stats.ObserveMatchDuration(time.Since(start))
	if len(alerts) == 0 {
		return
	}
	if w.suppressor != nil {
		alerts = w.suppressor.Filter(alerts)
		if len(alerts) == 0 {
			return
		}
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

	if w.redactor != nil {
		w.redactor(alerts)
	}

	writeStart := time.Now()
	err := w.writer.WriteBatch(ctx, alerts)
	w.stats.ObserveAlertWriteDuration(time.Since(writeStart))
	if err != nil {
		w.stats.IncAlertWriteError()
		w.logger.Warn("write pipeline alerts", zap.Error(err))
		return
	}
	w.stats.ObserveAlerts(alerts)
	w.logger.Info("packet matched rules",
		zap.String("src_ip", pkt.SrcIP),
		zap.String("dst_ip", pkt.DstIP),
		zap.Int("alerts", len(alerts)))
}
