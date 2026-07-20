package receiver

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

const defaultUDSPath = "/tmp/netsentry.sock"

const (
	maxUDSFrameBytes      = 64 << 10
	maxPacketPayloadBytes = 4096
)

var errSessionProtocol = errors.New("uds session protocol violation")

type connectionSession struct {
	helloReceived bool
	sessionID     string
}

// Config controls the Unix socket receiver.
type Config struct {
	Path           string
	SocketMode     os.FileMode
	MaxConnections int
	ReadTimeout    time.Duration
	BufferSize     int
	Stats          *stats.Stats
}

// Receiver owns a UDS listener and a context-aware packet channel.
type Receiver struct {
	cfg     Config
	logger  *zap.Logger
	packets chan *model.PacketInfo
	state   *heartbeatState
	ln      net.Listener
	stats   *stats.Stats
	wg      sync.WaitGroup
}

// New constructs a receiver. Start must be called before packets arrive.
func New(cfg Config, logger *zap.Logger) *Receiver {
	if cfg.Path == "" {
		cfg.Path = defaultUDSPath
	}
	if cfg.SocketMode == 0 {
		cfg.SocketMode = 0o600
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1024
	}
	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = 4
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Receiver{
		cfg:     cfg,
		logger:  logger,
		packets: make(chan *model.PacketInfo, cfg.BufferSize),
		state:   newHeartbeatState(),
		stats:   cfg.Stats,
	}
}

// Packets returns the decoded data frames.
func (r *Receiver) Packets() <-chan *model.PacketInfo { return r.packets }

// QueueDepth returns the current number of packets waiting for pipeline work.
func (r *Receiver) QueueDepth() int { return len(r.packets) }

// State returns the latest capture control-frame state.
func (r *Receiver) State() State { return r.state.Snapshot() }

// Stop closes the listening socket and removes the socket path. Existing
// connections also stop when the context passed to Start is cancelled.
func (r *Receiver) Stop() {
	if r.ln != nil {
		_ = r.ln.Close()
	}
	_ = os.Remove(r.cfg.Path)
}

// Wait blocks until the receiver accept loop and connection handlers exit.
func (r *Receiver) Wait() {
	r.wg.Wait()
}

// Start begins accepting UDS connections until ctx is cancelled.
func (r *Receiver) Start(ctx context.Context) error {
	_ = os.Remove(r.cfg.Path)
	ln, err := net.Listen("unix", r.cfg.Path)
	if err != nil {
		return fmt.Errorf("start uds listener %q: %w", r.cfg.Path, err)
	}
	r.ln = ln
	if err := os.Chmod(r.cfg.Path, r.cfg.SocketMode); err != nil {
		r.logger.Warn("chmod uds socket", zap.Error(err))
	}

	go func() {
		<-ctx.Done()
		r.Stop()
	}()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.acceptLoop(ctx, ln)
	}()
	return nil
}

func (r *Receiver) acceptLoop(ctx context.Context, ln net.Listener) {
	r.logger.Info("uds listener started", zap.String("path", r.cfg.Path))
	connectionSlots := make(chan struct{}, r.cfg.MaxConnections)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return
			}
			r.logger.Warn("accept uds connection", zap.Error(err))
			continue
		}
		select {
		case connectionSlots <- struct{}{}:
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				defer func() { <-connectionSlots }()
				r.handleConn(ctx, conn)
			}()
		default:
			r.logger.Warn("reject uds connection: capacity exhausted", zap.Int("max_connections", r.cfg.MaxConnections))
			_ = conn.Close()
		}
	}
}

func (r *Receiver) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	if err := conn.SetReadDeadline(time.Now().Add(r.cfg.ReadTimeout)); err != nil {
		r.logger.Warn("set uds read deadline", zap.Error(err))
		return
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 16*1024), maxUDSFrameBytes)
	session := &connectionSession{}
	for scanner.Scan() {
		if isNetworkDeadline(scanner.Err()) {
			r.logger.Debug("close idle uds connection", zap.Duration("read_timeout", r.cfg.ReadTimeout))
			return
		}
		if err := r.handleLine(ctx, scanner.Bytes(), session); err != nil {
			r.logger.Warn("handle uds frame", zap.Error(err))
			if errors.Is(err, errSessionProtocol) {
				return
			}
		}
		if err := conn.SetReadDeadline(time.Now().Add(r.cfg.ReadTimeout)); err != nil {
			r.logger.Warn("refresh uds read deadline", zap.Error(err))
			return
		}
	}
	if err := scanner.Err(); err != nil {
		if isNetworkDeadline(err) {
			r.logger.Debug("close idle uds connection", zap.Duration("read_timeout", r.cfg.ReadTimeout))
			return
		}
		r.stats.IncDecodeError()
		r.logger.Warn("read uds connection", zap.Error(err))
	}
}

func isNetworkDeadline(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.Is(err, os.ErrDeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout()
}

func (r *Receiver) handleLine(ctx context.Context, line []byte, session *connectionSession) error {
	r.stats.IncFrame()
	var meta struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &meta); err != nil {
		r.stats.IncDecodeError()
		return fmt.Errorf("decode frame metadata: %w", err)
	}

	switch meta.Type {
	case "hello":
		if session.helloReceived {
			return r.sessionProtocolError("duplicate hello frame")
		}
		var h HelloFrame
		if err := json.Unmarshal(line, &h); err != nil {
			r.stats.IncDecodeError()
			return fmt.Errorf("decode hello frame: %w", err)
		}
		if h.SessionID == "" || h.Version == "" {
			r.stats.IncDecodeError()
			return fmt.Errorf("invalid hello frame")
		}
		session.helloReceived = true
		session.sessionID = h.SessionID
		r.state.SetHello(h)
		r.stats.IncControlFrame()
		return nil
	case "heartbeat":
		if !session.helloReceived {
			return r.sessionProtocolError("heartbeat before hello")
		}
		var h HeartbeatFrame
		if err := json.Unmarshal(line, &h); err != nil {
			r.stats.IncDecodeError()
			return fmt.Errorf("decode heartbeat frame: %w", err)
		}
		if h.SessionID == "" {
			r.stats.IncDecodeError()
			return fmt.Errorf("invalid heartbeat frame")
		}
		if h.SessionID != session.sessionID {
			return r.sessionProtocolError("heartbeat session_id does not match hello")
		}
		r.state.SetHeartbeat(h)
		r.stats.IncControlFrame()
		return nil
	case "":
		if !session.helloReceived {
			return r.sessionProtocolError("packet before hello")
		}
		var pkt model.PacketInfo
		if err := json.Unmarshal(line, &pkt); err != nil {
			r.stats.IncDecodeError()
			return fmt.Errorf("decode packet frame: %w", err)
		}
		if err := validatePacketFrame(&pkt); err != nil {
			r.stats.IncDecodeError()
			return err
		}
		select {
		case r.packets <- &pkt:
			r.stats.IncPacketReceived()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	default:
		r.stats.IncDecodeError()
		return fmt.Errorf("unknown control frame type %q", meta.Type)
	}
}

func (r *Receiver) sessionProtocolError(message string) error {
	r.stats.IncDecodeError()
	return fmt.Errorf("%w: %s", errSessionProtocol, message)
}

func validatePacketFrame(pkt *model.PacketInfo) error {
	if pkt == nil {
		return fmt.Errorf("invalid packet frame: null packet")
	}
	if pkt.TimestampUsec < 0 || pkt.TimestampUsec >= 1_000_000 {
		return fmt.Errorf("invalid packet frame: timestamp_usec out of range")
	}
	if net.ParseIP(pkt.SrcIP) == nil || net.ParseIP(pkt.DstIP) == nil {
		return fmt.Errorf("invalid packet frame: source and destination IPs are required")
	}
	if pkt.PayloadLen > maxPacketPayloadBytes {
		return fmt.Errorf("invalid packet frame: payload_len exceeds %d", maxPacketPayloadBytes)
	}
	decoded, err := base64.StdEncoding.DecodeString(pkt.PayloadPreview)
	if err != nil {
		return fmt.Errorf("invalid packet frame: payload_preview is not base64: %w", err)
	}
	if len(decoded) != int(pkt.PayloadLen) {
		return fmt.Errorf("invalid packet frame: payload_len does not match payload_preview")
	}
	return nil
}

// ParseSocketMode converts config values such as "0600" into a file mode.
func ParseSocketMode(mode string) os.FileMode {
	if mode == "" {
		return 0o600
	}
	v, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0o600
	}
	return os.FileMode(v)
}

// WaitForPacket is a small helper for tests and integration callers.
func WaitForPacket(ctx context.Context, packets <-chan *model.PacketInfo) (*model.PacketInfo, error) {
	select {
	case pkt, ok := <-packets:
		if !ok {
			return nil, io.EOF
		}
		return pkt, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, context.DeadlineExceeded
	}
}
