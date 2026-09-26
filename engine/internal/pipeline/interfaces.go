package pipeline

import (
	"context"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Matcher evaluates packets and returns every alert produced by active rules.
type Matcher interface {
	Match(pkt *model.PacketInfo) []*model.Alert
}

// AlertWriter persists or forwards alerts produced by the pipeline.
type AlertWriter interface {
	WriteBatch(ctx context.Context, alerts []*model.Alert) error
}

// SuppressionFilter removes alerts that should not be written.
type SuppressionFilter interface {
	Filter(alerts []*model.Alert) []*model.Alert
}

// AlertRedactor mutates alerts before they are persisted or returned by storage-backed APIs.
type AlertRedactor func(alerts []*model.Alert)

// Observer is configured before workers start. Implementations must be safe for
// concurrent workers. Durable is called only after successful writer return.
type Observer interface {
	Arrival(pkt *model.PacketInfo) error
	Durable(pkt *model.PacketInfo, alerts []*model.Alert, at time.Time) error
	Processed(pkt *model.PacketInfo, at time.Time) error
}
