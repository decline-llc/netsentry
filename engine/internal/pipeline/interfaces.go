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
