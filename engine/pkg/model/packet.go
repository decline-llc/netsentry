package model

import (
	"fmt"
	"regexp"
	"time"
)

// PacketInfo is the deserialized form of a C-side PacketInfo JSON frame.
type PacketInfo struct {
	SLO            *SLOMetadata `json:"slo,omitempty"`
	TimestampSec   int64        `json:"timestamp_sec"`
	TimestampUsec  int32        `json:"timestamp_usec"`
	SrcIP          string       `json:"src_ip"`
	DstIP          string       `json:"dst_ip"`
	SrcPort        uint16       `json:"src_port"`
	DstPort        uint16       `json:"dst_port"`
	Protocol       uint8        `json:"protocol"`
	TCPFlags       string       `json:"tcp_flags,omitempty"`
	PayloadLen     uint16       `json:"payload_len"`
	PayloadPreview string       `json:"payload_preview,omitempty"`
	IsFragment     bool         `json:"is_fragment"`
	Truncated      bool         `json:"truncated"`
}

// SLOMetadata is supplied by an instrumented ingress, not inferred at matching.
type SLOMetadata struct {
	RunID         string `json:"run_id"`
	PacketID      string `json:"packet_id"`
	ArrivalUnixNS int64  `json:"arrival_unix_ns"`
}

var sloIdentifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

func ValidSLOIdentifier(value string) bool { return sloIdentifier.MatchString(value) }

func (m *SLOMetadata) Validate() error {
	if m == nil || !ValidSLOIdentifier(m.RunID) || !ValidSLOIdentifier(m.PacketID) || m.ArrivalUnixNS <= 0 {
		return fmt.Errorf("invalid SLO run/packet identity or arrival timestamp")
	}
	return nil
}

func (p *PacketInfo) Timestamp() time.Time {
	return time.Unix(p.TimestampSec, int64(p.TimestampUsec)*1000)
}

// ProtocolName returns the canonical alert-storage name for an IP protocol.
func ProtocolName(protocol uint8) string {
	switch protocol {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 1:
		return "ICMP"
	default:
		return fmt.Sprintf("PROTO_%d", protocol)
	}
}
