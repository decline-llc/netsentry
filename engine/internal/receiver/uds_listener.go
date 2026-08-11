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
	"net/netip"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

const defaultUDSPath = "/tmp/netsentry.sock"

const (
	maxUDSFrameBytes           = 64 << 10
	maxPacketPayloadBytes      = 4096
	existingSocketProbeTimeout = time.Second
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
	cfg                 Config
	logger              *zap.Logger
	packets             chan *model.PacketInfo
	state               *heartbeatState
	ln                  net.Listener
	socket              os.FileInfo
	stats               *stats.Stats
	wg                  sync.WaitGroup
	slots               chan struct{}
	probeExistingSocket func(string) (net.Conn, error)
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
		probeExistingSocket: func(path string) (net.Conn, error) {
			return net.DialTimeout("unix", path, existingSocketProbeTimeout)
		},
	}
}

// Packets returns the decoded data frames.
func (r *Receiver) Packets() <-chan *model.PacketInfo { return r.packets }

// QueueDepth returns the current number of packets waiting for pipeline work.
func (r *Receiver) QueueDepth() int { return len(r.packets) }

// State returns the latest capture control-frame state.
func (r *Receiver) State() State { return r.state.Snapshot() }

// Stop closes the listening socket and removes its owned socket path. Existing
// connections also stop when the context passed to Start is cancelled.
func (r *Receiver) Stop() {
	if err := removeOwnedSocket(r.cfg.Path, r.socket); err != nil {
		r.logger.Warn("remove owned uds socket", zap.Error(err))
	}
	if r.ln != nil {
		_ = r.ln.Close()
	}
}

// Wait blocks until the receiver accept loop and connection handlers exit.
func (r *Receiver) Wait() {
	r.wg.Wait()
}

// Start begins accepting UDS connections until ctx is cancelled.
func (r *Receiver) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("start uds listener %q: %w", r.cfg.Path, err)
	}
	if err := removeExistingSocket(r.cfg.Path, r.probeExistingSocket); err != nil {
		return fmt.Errorf("prepare uds listener %q: %w", r.cfg.Path, err)
	}
	ln, err := net.Listen("unix", r.cfg.Path)
	if err != nil {
		return fmt.Errorf("start uds listener %q: %w", r.cfg.Path, err)
	}
	unixListener, ok := ln.(*net.UnixListener)
	if !ok {
		_ = ln.Close()
		return fmt.Errorf("start uds listener %q: unexpected listener type %T", r.cfg.Path, ln)
	}
	unixListener.SetUnlinkOnClose(false)
	socket, err := os.Lstat(r.cfg.Path)
	if err != nil {
		_ = ln.Close()
		return fmt.Errorf("inspect started uds listener %q: %w", r.cfg.Path, err)
	}
	if socket.Mode()&os.ModeSocket == 0 {
		_ = ln.Close()
		return fmt.Errorf("inspect started uds listener %q: path is not a Unix socket", r.cfg.Path)
	}
	r.ln = ln
	r.socket = socket
	r.slots = make(chan struct{}, r.cfg.MaxConnections)
	for i := 0; i < r.cfg.MaxConnections; i++ {
		r.slots <- struct{}{}
	}
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
		r.acceptLoop(ctx, ln, r.slots)
	}()
	return nil
}

func removeExistingSocket(path string, probe func(string) (net.Conn, error)) error {
	original, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect existing path: %w", err)
	}
	if original.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("existing path is not a Unix socket")
	}
	conn, probeErr := probe(path)
	if conn != nil {
		_ = conn.Close()
	}
	if probeErr == nil {
		return fmt.Errorf("existing Unix socket is active")
	}
	if !errors.Is(probeErr, syscall.ECONNREFUSED) {
		if errors.Is(probeErr, os.ErrNotExist) {
			if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
				return nil
			} else if err != nil {
				return fmt.Errorf("inspect existing path after liveness probe: %w", err)
			}
		}
		return fmt.Errorf("probe existing Unix socket: %w", probeErr)
	}
	current, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect existing path after liveness probe: %w", err)
	}
	if current.Mode()&os.ModeSocket == 0 || !sameUnixSocketIdentity(original, current) {
		return fmt.Errorf("existing path changed during Unix socket liveness probe")
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove existing Unix socket: %w", err)
	}
	return nil
}

func sameUnixSocketIdentity(original, current os.FileInfo) bool {
	if !os.SameFile(original, current) {
		return false
	}
	originalStat, originalOK := original.Sys().(*syscall.Stat_t)
	currentStat, currentOK := current.Sys().(*syscall.Stat_t)
	return originalOK && currentOK && originalStat.Ctim == currentStat.Ctim
}

func removeOwnedSocket(path string, owned os.FileInfo) error {
	if owned == nil {
		return nil
	}
	current, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect socket path: %w", err)
	}
	if current.Mode()&os.ModeSocket == 0 || !os.SameFile(current, owned) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove socket path: %w", err)
	}
	return nil
}

func (r *Receiver) acceptLoop(ctx context.Context, ln net.Listener, slots chan struct{}) {
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
		select {
		case <-slots:
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				defer func() { slots <- struct{}{} }()
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
	if !isIPv4Address(pkt.SrcIP) || !isIPv4Address(pkt.DstIP) {
		return fmt.Errorf("invalid packet frame: source and destination must be IPv4 addresses")
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

func isIPv4Address(value string) bool {
	address, err := netip.ParseAddr(value)
	return err == nil && address.Is4()
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
