package pipeline

import (
	"github.com/decline-llc/netsentry/pkg/model"
)

// Matcher is the interface every rule type must implement.
type Matcher interface {
	// ID returns the unique identifier for this matcher (usually the rule ID).
	ID() string
	// Match inspects the packet and returns a non-nil Alert on a hit.
	Match(pkt *model.PacketInfo) *model.Alert
}

// AlertWriter persists or forwards alerts produced by the pipeline.
type AlertWriter interface {
	Write(alert *model.Alert) error
	Close() error
}
