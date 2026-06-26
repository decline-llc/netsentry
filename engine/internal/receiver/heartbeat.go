package receiver

import "sync/atomic"

// HelloFrame is the capture-side handshake frame.
type HelloFrame struct {
	Type          string `json:"type"`
	Version       string `json:"version"`
	SessionID     string `json:"session_id"`
	PID           int    `json:"pid"`
	Hostname      string `json:"hostname"`
	MaxPayloadLen int    `json:"max_payload_len"`
}

// HeartbeatFrame is the capture-side health frame.
type HeartbeatFrame struct {
	Type               string  `json:"type"`
	SessionID          string  `json:"session_id"`
	Seq                uint32  `json:"seq"`
	Sent               uint64  `json:"sent"`
	Dropped            uint64  `json:"dropped"`
	ParseErrors        uint64  `json:"parse_errors"`
	BufUtilPct         uint32  `json:"buf_util_pct"`
	AvgJSONSerializeUS float64 `json:"avg_json_serialize_us"`
	UDSWriteErrors     uint64  `json:"uds_write_errors"`
}

// State keeps the most recent capture control frames.
type State struct {
	SessionID string
	Hello     HelloFrame
	Heartbeat HeartbeatFrame
}

type heartbeatState struct {
	value atomic.Value // stores State
}

func newHeartbeatState() *heartbeatState {
	hs := &heartbeatState{}
	hs.value.Store(State{})
	return hs
}

func (s *heartbeatState) Snapshot() State {
	v, _ := s.value.Load().(State)
	return v
}

func (s *heartbeatState) SetHello(h HelloFrame) {
	cur := s.Snapshot()
	cur.SessionID = h.SessionID
	cur.Hello = h
	s.value.Store(cur)
}

func (s *heartbeatState) SetHeartbeat(h HeartbeatFrame) {
	cur := s.Snapshot()
	cur.SessionID = h.SessionID
	cur.Heartbeat = h
	s.value.Store(cur)
}
