package receiver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/pkg/model"
)

const defaultUDSPath = "/tmp/netsentry.sock"

// Config controls the Unix socket receiver.
type Config struct {
	Path       string
	SocketMode os.FileMode
	BufferSize int
}

// Receiver owns a UDS listener and a context-aware packet channel.
type Receiver struct {
	cfg     Config
	logger  *zap.Logger
	packets chan *model.PacketInfo
	state   *heartbeatState
	ln      net.Listener
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
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Receiver{
		cfg:     cfg,
		logger:  logger,
		packets: make(chan *model.PacketInfo, cfg.BufferSize),
		state:   newHeartbeatState(),
	}
}

// Packets returns the decoded data frames.
func (r *Receiver) Packets() <-chan *model.PacketInfo { return r.packets }

// State returns the latest capture control-frame state.
func (r *Receiver) State() State { return r.state.Snapshot() }

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
		_ = ln.Close()
		_ = os.Remove(r.cfg.Path)
	}()

	go r.acceptLoop(ctx, ln)
	return nil
}

func (r *Receiver) acceptLoop(ctx context.Context, ln net.Listener) {
	r.logger.Info("uds listener started", zap.String("path", r.cfg.Path))
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return
			}
			r.logger.Warn("accept uds connection", zap.Error(err))
			continue
		}
		go r.handleConn(ctx, conn)
	}
}

func (r *Receiver) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if err := r.handleLine(ctx, scanner.Bytes()); err != nil {
			r.logger.Warn("handle uds frame", zap.Error(err))
		}
	}
	if err := scanner.Err(); err != nil {
		r.logger.Warn("read uds connection", zap.Error(err))
	}
}

func (r *Receiver) handleLine(ctx context.Context, line []byte) error {
	var meta struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &meta); err != nil {
		return fmt.Errorf("decode frame metadata: %w", err)
	}

	switch meta.Type {
	case "hello":
		var h HelloFrame
		if err := json.Unmarshal(line, &h); err != nil {
			return fmt.Errorf("decode hello frame: %w", err)
		}
		if h.SessionID == "" || h.Version == "" {
			return fmt.Errorf("invalid hello frame")
		}
		r.state.SetHello(h)
		return nil
	case "heartbeat":
		var h HeartbeatFrame
		if err := json.Unmarshal(line, &h); err != nil {
			return fmt.Errorf("decode heartbeat frame: %w", err)
		}
		if h.SessionID == "" {
			return fmt.Errorf("invalid heartbeat frame")
		}
		r.state.SetHeartbeat(h)
		return nil
	case "":
		var pkt model.PacketInfo
		if err := json.Unmarshal(line, &pkt); err != nil {
			return fmt.Errorf("decode packet frame: %w", err)
		}
		select {
		case r.packets <- &pkt:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	default:
		return fmt.Errorf("unknown control frame type %q", meta.Type)
	}
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
	case pkt := <-packets:
		return pkt, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, context.DeadlineExceeded
	}
}
