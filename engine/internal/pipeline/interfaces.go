package pipeline

import (
	"context"

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
