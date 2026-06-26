package model

import "time"

// PacketInfo is the deserialized form of a C-side PacketInfo JSON frame.
type PacketInfo struct {
	TimestampSec  int64  `json:"timestamp_sec"`
	TimestampUsec int32  `json:"timestamp_usec"`
	SrcIP         string `json:"src_ip"`
	DstIP         string `json:"dst_ip"`
	SrcPort       uint16 `json:"src_port"`
	DstPort       uint16 `json:"dst_port"`
	Protocol      uint8  `json:"protocol"`
	TCPFlags      string `json:"tcp_flags,omitempty"`
	PayloadLen    uint16 `json:"payload_len"`
	PayloadPreview string `json:"payload_preview,omitempty"`
	IsFragment    bool   `json:"is_fragment"`
	Truncated     bool   `json:"truncated"`
}

func (p *PacketInfo) Timestamp() time.Time {
	return time.Unix(p.TimestampSec, int64(p.TimestampUsec)*1000)
}
