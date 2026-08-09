package receiver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestHandleLineControlFrames(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	ctx := context.Background()
	session := &connectionSession{}

	if err := r.handleLine(ctx, []byte(`{"type":"hello","version":"0.1.0","session_id":"abcd1234","pid":123,"hostname":"host","max_payload_len":4096}`), session); err != nil {
		t.Fatalf("hello: %v", err)
	}
	if got := r.State(); got.SessionID != "abcd1234" || got.Hello.Version != "0.1.0" {
		t.Fatalf("unexpected hello state: %+v", got)
	}

	if err := r.handleLine(ctx, []byte(`{"type":"heartbeat","session_id":"abcd1234","seq":7,"sent":11,"dropped":2,"parse_errors":3,"buf_util_pct":4,"avg_json_serialize_us":1.5,"uds_write_errors":6}`), session); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if got := r.State(); got.Heartbeat.Seq != 7 || got.Heartbeat.UDSWriteErrors != 6 {
		t.Fatalf("unexpected heartbeat state: %+v", got)
	}
}

func TestHandleLineDataFrame(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	ctx := context.Background()

	line := []byte(`{"timestamp_sec":1719300000,"timestamp_usec":123456,"src_ip":"10.0.0.1","dst_ip":"10.0.0.2","src_port":12345,"dst_port":80,"protocol":6,"tcp_flags":"ACK","payload_len":4,"payload_preview":"dGVzdA==","is_fragment":false,"truncated":false}`)
	if err := r.handleLine(ctx, line, establishedSession("data-session")); err != nil {
		t.Fatalf("packet: %v", err)
	}
	pkt := <-r.Packets()
	if pkt.SrcIP != "10.0.0.1" || pkt.DstPort != 80 || pkt.Protocol != 6 {
		t.Fatalf("unexpected packet: %+v", pkt)
	}
}

func TestHandleLineBadJSON(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	if err := r.handleLine(context.Background(), []byte(`{"timestamp_sec"`), &connectionSession{}); err == nil {
		t.Fatal("expected bad JSON error")
	}
}

func TestHandleLineRejectsInvalidPacketContract(t *testing.T) {
	metrics := stats.New()
	r := New(Config{BufferSize: 1, Stats: metrics}, zap.NewNop())
	tests := []string{
		`{"timestamp_sec":1,"timestamp_usec":1000000,"src_ip":"10.0.0.1","dst_ip":"10.0.0.2","protocol":6,"payload_len":0,"payload_preview":""}`,
		`{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"invalid","dst_ip":"10.0.0.2","protocol":6,"payload_len":0,"payload_preview":""}`,
		`{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"10.0.0.1","dst_ip":"10.0.0.2","protocol":6,"payload_len":4,"payload_preview":"bad!"}`,
		`{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"10.0.0.1","dst_ip":"10.0.0.2","protocol":6,"payload_len":3,"payload_preview":"dGVzdA=="}`,
	}
	for _, line := range tests {
		if err := r.handleLine(context.Background(), []byte(line), establishedSession("contract-session")); err == nil {
			t.Fatalf("expected invalid packet frame to be rejected: %s", line)
		}
	}
	if got := metrics.Snapshot().DecodeErrors; got != uint64(len(tests)) {
		t.Fatalf("decode errors = %d, want %d", got, len(tests))
	}
}

func TestHandleLineRejectsNonIPv4PacketAddresses(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{
			name: "IPv6 source",
			line: `{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"2001:db8::1","dst_ip":"10.0.0.2","protocol":6,"payload_len":0,"payload_preview":""}`,
		},
		{
			name: "IPv6 destination",
			line: `{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"10.0.0.1","dst_ip":"2001:db8::2","protocol":6,"payload_len":0,"payload_preview":""}`,
		},
		{
			name: "IPv4-mapped IPv6 source",
			line: `{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"::ffff:192.0.2.1","dst_ip":"10.0.0.2","protocol":6,"payload_len":0,"payload_preview":""}`,
		},
		{
			name: "IPv4-mapped IPv6 destination",
			line: `{"timestamp_sec":1,"timestamp_usec":0,"src_ip":"10.0.0.1","dst_ip":"::ffff:192.0.2.2","protocol":6,"payload_len":0,"payload_preview":""}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metrics := stats.New()
			r := New(Config{BufferSize: 1, Stats: metrics}, zap.NewNop())
			err := r.handleLine(
				context.Background(),
				[]byte(test.line),
				establishedSession("ipv4-contract-session"),
			)
			if err == nil || !strings.Contains(err.Error(), "source and destination must be IPv4 addresses") {
				t.Fatalf("packet error = %v, want IPv4 contract rejection", err)
			}
			if got := metrics.Snapshot().DecodeErrors; got != 1 {
				t.Fatalf("decode errors = %d, want 1", got)
			}
			select {
			case packet := <-r.Packets():
				t.Fatalf("rejected packet was queued: %+v", packet)
			default:
			}
		})
	}
}

func TestStartPreservesPreExistingRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	want := []byte("operator data\n")
	if err := os.WriteFile(path, want, 0o640); err != nil {
		t.Fatalf("write regular file: %v", err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatalf("chmod regular file: %v", err)
	}

	r := New(Config{Path: path}, zap.NewNop())
	err := r.Start(context.Background())
	if err == nil || !strings.Contains(err.Error(), "existing path is not a Unix socket") {
		t.Fatalf("start error = %v, want non-socket rejection", err)
	}
	if r.ln != nil {
		t.Fatal("receiver installed a listener after rejecting a regular file")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read preserved regular file: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("regular file bytes = %q, want %q", got, want)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("stat preserved regular file: %v", err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0o640 {
		t.Fatalf("regular file mode = %o, want 640", gotMode)
	}
}

func TestStartPreservesPreExistingSymlink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "netsentry.sock")
	targetName := "operator.sock"
	targetPath := filepath.Join(dir, targetName)
	target, err := net.Listen("unix", targetPath)
	if err != nil {
		t.Fatalf("create symlink target socket: %v", err)
	}
	defer target.Close()
	if err := os.Symlink(targetName, path); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	r := New(Config{Path: path}, zap.NewNop())
	err = r.Start(context.Background())
	if err == nil || !strings.Contains(err.Error(), "existing path is not a Unix socket") {
		t.Fatalf("start error = %v, want symlink rejection", err)
	}
	if r.ln != nil {
		t.Fatal("receiver installed a listener after rejecting a symlink")
	}
	gotTarget, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("read preserved symlink: %v", err)
	}
	if gotTarget != targetName {
		t.Fatalf("symlink target = %q, want %q", gotTarget, targetName)
	}
	targetInfo, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("stat symlink target socket: %v", err)
	}
	if targetInfo.Mode()&os.ModeSocket == 0 {
		t.Fatalf("symlink target mode = %v, want Unix socket", targetInfo.Mode())
	}
}

func TestStartCanceledContextPreservesAbsentPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := New(Config{Path: path}, zap.NewNop())
	err := r.Start(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("start error = %v, want context.Canceled", err)
	}
	if r.ln != nil {
		t.Fatal("receiver installed a listener for an already-canceled context")
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("socket path after rejected startup: %v, want absent", err)
	}
}

func TestStartCanceledContextPreservesExistingUnixSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	existing, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("create existing Unix socket: %v", err)
	}
	existing.(*net.UnixListener).SetUnlinkOnClose(false)
	t.Cleanup(func() {
		_ = existing.Close()
		_ = os.Remove(path)
	})
	before, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("stat existing Unix socket: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := New(Config{Path: path}, zap.NewNop())
	err = r.Start(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("start error = %v, want context.Canceled", err)
	}
	if r.ln != nil {
		t.Fatal("receiver installed a listener for an already-canceled context")
	}
	after, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("stat preserved Unix socket: %v", err)
	}
	if after.Mode()&os.ModeSocket == 0 {
		t.Fatalf("preserved path mode = %v, want Unix socket", after.Mode())
	}
	if !os.SameFile(before, after) {
		t.Fatal("already-canceled startup replaced the existing Unix socket identity")
	}
}

func TestStartReclaimsPreExistingUnixSocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	stale, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("create stale Unix socket: %v", err)
	}
	stale.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := stale.Close(); err != nil {
		t.Fatalf("close stale Unix socket: %v", err)
	}

	r := New(Config{Path: path}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver over stale Unix socket: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		r.Wait()
	})
	cancel()
	waitForReceiverShutdown(t, r)
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("owned socket path remains after shutdown: %v", err)
	}
}

func TestStopPreservesReplacementRegularFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	r := New(Config{Path: path}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		r.Wait()
	})
	want := []byte("replacement data\n")
	if err := os.Remove(path); err != nil {
		t.Fatalf("displace owned socket: %v", err)
	}
	if err := os.WriteFile(path, want, 0o640); err != nil {
		t.Fatalf("write replacement regular file: %v", err)
	}

	cancel()
	waitForReceiverShutdown(t, r)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read replacement regular file: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("replacement bytes = %q, want %q", got, want)
	}
}

func TestStopPreservesReplacementSymlink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dir := t.TempDir()
	path := filepath.Join(dir, "netsentry.sock")
	r := New(Config{Path: path}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		r.Wait()
	})
	targetName := "replacement-target"
	targetPath := filepath.Join(dir, targetName)
	want := []byte("replacement target data\n")
	if err := os.WriteFile(targetPath, want, 0o600); err != nil {
		t.Fatalf("write replacement target: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("displace owned socket: %v", err)
	}
	if err := os.Symlink(targetName, path); err != nil {
		t.Fatalf("create replacement symlink: %v", err)
	}

	cancel()
	waitForReceiverShutdown(t, r)
	gotTarget, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("read replacement symlink: %v", err)
	}
	if gotTarget != targetName {
		t.Fatalf("replacement symlink target = %q, want %q", gotTarget, targetName)
	}
	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read replacement target file: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("replacement target bytes = %q, want %q", got, want)
	}
}

func TestWaitForPacketReturnsEOFForClosedChannel(t *testing.T) {
	packets := make(chan *model.PacketInfo)
	close(packets)
	pkt, err := WaitForPacket(context.Background(), packets)
	if pkt != nil || !errors.Is(err, io.EOF) {
		t.Fatalf("packet=%+v err=%v, want nil/io.EOF", pkt, err)
	}
}

func TestHandleLineContextCancelWhileBlocked(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	session := establishedSession("cancel-session")

	line := []byte(`{"timestamp_sec":1,"timestamp_usec":2,"src_ip":"10.0.0.1","dst_ip":"10.0.0.2","src_port":1,"dst_port":2,"protocol":6,"payload_len":0,"is_fragment":false,"truncated":false}`)
	if err := r.handleLine(ctx, line, session); err != nil {
		t.Fatalf("first packet: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- r.handleLine(ctx, line, session) }()

	select {
	case err := <-done:
		t.Fatalf("send should block before cancel, got %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected context cancellation error")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for context cancellation")
	}
}

func TestHandleConnRejectsSessionProtocolViolations(t *testing.T) {
	tests := []struct {
		name   string
		frames []any
	}{
		{
			name:   "packet before hello",
			frames: []any{testPacketFrame(80)},
		},
		{
			name: "heartbeat before hello",
			frames: []any{HeartbeatFrame{
				Type: "heartbeat", SessionID: "session-one", Seq: 1,
			}},
		},
		{
			name: "duplicate hello",
			frames: []any{
				HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-one", PID: 1, Hostname: "host", MaxPayloadLen: 4096},
				HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-two", PID: 2, Hostname: "host", MaxPayloadLen: 4096},
			},
		},
		{
			name: "mismatched heartbeat session",
			frames: []any{
				HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-one", PID: 1, Hostname: "host", MaxPayloadLen: 4096},
				HeartbeatFrame{Type: "heartbeat", SessionID: "session-two", Seq: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := stats.New()
			r := New(Config{BufferSize: 1, Stats: metrics}, zap.NewNop())
			server, client := net.Pipe()
			defer client.Close()
			done := make(chan struct{})
			go func() {
				r.handleConn(context.Background(), server)
				close(done)
			}()

			for _, frame := range tt.frames {
				if err := writeJSONFrame(client, frame); err != nil {
					t.Fatalf("write frame: %v", err)
				}
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("protocol violation did not close the connection")
			}
			if got := metrics.Snapshot().DecodeErrors; got != 1 {
				t.Fatalf("decode errors = %d, want exactly 1", got)
			}
			if got := len(r.Packets()); got != 0 {
				t.Fatalf("rejected connection queued %d packet(s)", got)
			}
		})
	}
}

func TestProtocolViolationClosesOnlyOffendingConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	metrics := stats.New()
	r := New(Config{BufferSize: 1, Stats: metrics}, zap.NewNop())

	validServer, validClient := net.Pipe()
	defer validClient.Close()
	validDone := make(chan struct{})
	go func() {
		r.handleConn(ctx, validServer)
		close(validDone)
	}()
	if err := writeJSONFrame(validClient, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "valid", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write valid hello: %v", err)
	}

	badServer, badClient := net.Pipe()
	badDone := make(chan struct{})
	go func() {
		r.handleConn(ctx, badServer)
		close(badDone)
	}()
	if err := writeJSONFrame(badClient, testPacketFrame(81)); err != nil {
		t.Fatalf("write violating packet: %v", err)
	}
	select {
	case <-badDone:
	case <-time.After(time.Second):
		t.Fatal("offending connection remained open")
	}
	_ = badClient.Close()

	if err := writeJSONFrame(validClient, testPacketFrame(443)); err != nil {
		t.Fatalf("valid connection closed with offender: %v", err)
	}
	packetCtx, packetCancel := context.WithTimeout(context.Background(), time.Second)
	pkt, err := WaitForPacket(packetCtx, r.Packets())
	packetCancel()
	if err != nil || pkt.DstPort != 443 {
		t.Fatalf("valid connection packet=%+v err=%v", pkt, err)
	}
	if got := metrics.Snapshot().DecodeErrors; got != 1 {
		t.Fatalf("decode errors = %d, want exactly 1", got)
	}

	cancel()
	select {
	case <-validDone:
	case <-time.After(time.Second):
		t.Fatal("valid connection did not stop after cancellation")
	}
}

func TestStartProtocolViolationReleasesConnectionCapacity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	metrics := stats.New()
	r := New(Config{Path: path, MaxConnections: 1, BufferSize: 1, Stats: metrics}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	defer func() {
		cancel()
		r.Wait()
	}()

	offender, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial offending connection: %v", err)
	}
	if err := writeJSONFrame(offender, testPacketFrame(80)); err != nil {
		t.Fatalf("write violating packet: %v", err)
	}
	if err := offender.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set offender read deadline: %v", err)
	}
	if _, err := offender.Read(make([]byte, 1)); err == nil {
		t.Fatal("offending connection remained open")
	}
	_ = offender.Close()

	releaseCapacity := claimReleasedConnectionCapacity(t, r)
	releaseCapacity()

	replacement, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial replacement connection: %v", err)
	}
	defer replacement.Close()
	if err := writeJSONFrame(replacement, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "replacement", PID: 2, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write replacement hello: %v", err)
	}
	if err := writeJSONFrame(replacement, testPacketFrame(443)); err != nil {
		t.Fatalf("write replacement packet: %v", err)
	}
	packetCtx, packetCancel := context.WithTimeout(context.Background(), time.Second)
	pkt, err := WaitForPacket(packetCtx, r.Packets())
	packetCancel()
	if err != nil || pkt.DstPort != 443 {
		t.Fatalf("replacement packet=%+v err=%v", pkt, err)
	}
	if got := metrics.Snapshot().DecodeErrors; got != 1 {
		t.Fatalf("decode errors = %d, want exactly 1", got)
	}
}

func TestStartReceivesFramesOverUnixSocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	path := filepath.Join(t.TempDir(), "netsentry.sock")
	r := New(Config{Path: path, BufferSize: 4}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	defer func() {
		cancel()
		r.Wait()
	}()

	conn, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial receiver: %v", err)
	}
	defer conn.Close()

	frames := []any{
		HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "abcd1234", PID: 1, Hostname: "host", MaxPayloadLen: 4096},
		map[string]any{"timestamp_sec": 1, "timestamp_usec": 2, "src_ip": "10.0.0.1", "dst_ip": "10.0.0.2", "src_port": 1, "dst_port": 80, "protocol": 6, "payload_len": 0, "is_fragment": false, "truncated": false},
	}
	for _, frame := range frames {
		if err := writeJSONFrame(conn, frame); err != nil {
			t.Fatalf("write frame: %v", err)
		}
	}

	packetCtx, packetCancel := context.WithTimeout(context.Background(), time.Second)
	defer packetCancel()
	pkt, err := WaitForPacket(packetCtx, r.Packets())
	if err != nil {
		t.Fatalf("wait packet: %v", err)
	}
	if pkt.DstPort != 80 || r.State().SessionID != "abcd1234" {
		t.Fatalf("unexpected receiver state packet=%+v state=%+v", pkt, r.State())
	}
}

func TestStartAcceptsReconnectOnSameUnixSocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	path := filepath.Join(t.TempDir(), "netsentry.sock")
	r := New(Config{Path: path, BufferSize: 4}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	defer func() {
		cancel()
		r.Wait()
	}()

	first, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial first receiver connection: %v", err)
	}
	if err := writeJSONFrame(first, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-one", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write first hello: %v", err)
	}
	if err := writeJSONFrame(first, map[string]any{"timestamp_sec": 1, "timestamp_usec": 2, "src_ip": "10.0.0.1", "dst_ip": "10.0.0.2", "src_port": 1000, "dst_port": 80, "protocol": 6, "payload_len": 0, "is_fragment": false, "truncated": false}); err != nil {
		t.Fatalf("write first packet: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first connection: %v", err)
	}

	packetCtx, packetCancel := context.WithTimeout(context.Background(), time.Second)
	firstPkt, err := WaitForPacket(packetCtx, r.Packets())
	packetCancel()
	if err != nil {
		t.Fatalf("wait first packet: %v", err)
	}
	if firstPkt.DstPort != 80 {
		t.Fatalf("unexpected first packet: %+v", firstPkt)
	}

	second, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial second receiver connection: %v", err)
	}
	defer second.Close()
	if err := writeJSONFrame(second, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-two", PID: 2, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write second hello: %v", err)
	}
	if err := writeJSONFrame(second, map[string]any{"timestamp_sec": 3, "timestamp_usec": 4, "src_ip": "10.0.0.3", "dst_ip": "10.0.0.4", "src_port": 2000, "dst_port": 443, "protocol": 6, "payload_len": 0, "is_fragment": false, "truncated": false}); err != nil {
		t.Fatalf("write second packet: %v", err)
	}

	packetCtx, packetCancel = context.WithTimeout(context.Background(), time.Second)
	secondPkt, err := WaitForPacket(packetCtx, r.Packets())
	packetCancel()
	if err != nil {
		t.Fatalf("wait second packet: %v", err)
	}
	if secondPkt.DstPort != 443 {
		t.Fatalf("unexpected second packet: %+v", secondPkt)
	}
	if got := r.State(); got.SessionID != "session-two" {
		t.Fatalf("receiver state did not update after reconnect: %+v", got)
	}
}

func TestConcurrentConnectionsKeepSessionStateLocal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	metrics := stats.New()
	r := New(Config{Path: path, BufferSize: 2, Stats: metrics}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	defer func() {
		cancel()
		r.Wait()
	}()

	first, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial first connection: %v", err)
	}
	defer first.Close()
	second, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial second connection: %v", err)
	}
	defer second.Close()

	if err := writeJSONFrame(first, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-one", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write first hello: %v", err)
	}
	if err := writeJSONFrame(second, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-two", PID: 2, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write second hello: %v", err)
	}
	if err := writeJSONFrame(first, HeartbeatFrame{Type: "heartbeat", SessionID: "session-one", Seq: 1}); err != nil {
		t.Fatalf("write first heartbeat: %v", err)
	}
	if err := writeJSONFrame(second, HeartbeatFrame{Type: "heartbeat", SessionID: "session-two", Seq: 2}); err != nil {
		t.Fatalf("write second heartbeat: %v", err)
	}
	if err := writeJSONFrame(first, testPacketFrame(8443)); err != nil {
		t.Fatalf("write first packet after concurrent hellos: %v", err)
	}
	packetCtx, packetCancel := context.WithTimeout(context.Background(), time.Second)
	pkt, err := WaitForPacket(packetCtx, r.Packets())
	packetCancel()
	if err != nil || pkt.DstPort != 8443 {
		t.Fatalf("packet=%+v err=%v", pkt, err)
	}
	if got := metrics.Snapshot().DecodeErrors; got != 0 {
		t.Fatalf("connection-local sessions produced %d decode errors", got)
	}
}

func TestStartRejectsConnectionsAboveLimitAndReusesCapacity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	path := filepath.Join(t.TempDir(), "netsentry.sock")
	r := New(Config{Path: path, MaxConnections: 1, BufferSize: 4}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	defer func() {
		cancel()
		r.Wait()
	}()

	first, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial first receiver connection: %v", err)
	}
	if err := writeJSONFrame(first, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "first", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write first hello: %v", err)
	}
	if err := writeJSONFrame(first, testPacketFrame(80)); err != nil {
		t.Fatalf("write first packet: %v", err)
	}
	firstPacketCtx, firstPacketCancel := context.WithTimeout(context.Background(), time.Second)
	firstPacket, err := WaitForPacket(firstPacketCtx, r.Packets())
	firstPacketCancel()
	if err != nil || firstPacket.DstPort != 80 {
		t.Fatalf("first packet=%+v err=%v", firstPacket, err)
	}

	excess, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial excess receiver connection: %v", err)
	}
	if err := excess.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set excess connection deadline: %v", err)
	}
	if _, err := excess.Read(make([]byte, 1)); err == nil {
		t.Fatal("excess receiver connection should be closed")
	}
	_ = excess.Close()

	if err := first.Close(); err != nil {
		t.Fatalf("close first receiver connection: %v", err)
	}

	releaseCapacity := claimReleasedConnectionCapacity(t, r)
	releaseCapacity()

	replacement, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial replacement connection: %v", err)
	}
	defer replacement.Close()
	if err := writeJSONFrame(replacement, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "replacement", PID: 2, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write replacement hello: %v", err)
	}
	if err := writeJSONFrame(replacement, testPacketFrame(443)); err != nil {
		t.Fatalf("write replacement packet: %v", err)
	}
	replacementPacketCtx, replacementPacketCancel := context.WithTimeout(context.Background(), time.Second)
	replacementPacket, err := WaitForPacket(replacementPacketCtx, r.Packets())
	replacementPacketCancel()
	if err != nil || replacementPacket.DstPort != 443 {
		t.Fatalf("replacement packet=%+v err=%v", replacementPacket, err)
	}
}

func TestHandleConnTimesOutBeforeFirstCompleteFrame(t *testing.T) {
	metrics := stats.New()
	r := New(Config{ReadTimeout: 50 * time.Millisecond, BufferSize: 1, Stats: metrics}, zap.NewNop())
	server, client := net.Pipe()
	defer client.Close()
	done := make(chan struct{})
	go func() {
		r.handleConn(context.Background(), server)
		close(done)
	}()

	if _, err := client.Write([]byte("{")); err != nil {
		t.Fatalf("write partial frame: %v", err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("partial frame did not expire before the first complete frame")
	}
	if got := metrics.Snapshot().DecodeErrors; got != 0 {
		t.Fatalf("idle timeout incremented decode errors: %d", got)
	}
}

func TestHandleConnRefreshesReadDeadlineAfterEachFrame(t *testing.T) {
	metrics := stats.New()
	r := New(Config{ReadTimeout: time.Second, BufferSize: 1, Stats: metrics}, zap.NewNop())
	server, client := net.Pipe()
	recorded := &readDeadlineRecordingConn{Conn: server}
	done := make(chan struct{})
	go func() {
		r.handleConn(context.Background(), recorded)
		close(done)
	}()

	if err := writeJSONFrame(client, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "deadline-refresh", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write hello: %v", err)
	}
	if err := writeJSONFrame(client, map[string]any{"timestamp_sec": 1, "timestamp_usec": 2, "src_ip": "10.0.0.1", "dst_ip": "10.0.0.2", "src_port": 1, "dst_port": 80, "protocol": 6, "payload_len": 0, "is_fragment": false, "truncated": false}); err != nil {
		t.Fatalf("write packet: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("close client: %v", err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connection handler did not exit after client close")
	}
	if got := recorded.readDeadlineCount(); got != 3 {
		t.Fatalf("read deadline calls = %d, want initial plus one per complete frame", got)
	}
	if got := metrics.Snapshot().DecodeErrors; got != 0 {
		t.Fatalf("valid frames incremented decode errors: %d", got)
	}
}

func TestStartIdleTimeoutReleasesConnectionCapacity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	path := filepath.Join(t.TempDir(), "netsentry.sock")
	metrics := stats.New()
	r := New(Config{Path: path, MaxConnections: 1, ReadTimeout: 100 * time.Millisecond, BufferSize: 1, Stats: metrics}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}
	defer func() {
		cancel()
		r.Wait()
	}()

	first, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial first connection: %v", err)
	}
	if err := writeJSONFrame(first, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "idle-first", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write first hello: %v", err)
	}
	if err := writeJSONFrame(first, testPacketFrame(80)); err != nil {
		t.Fatalf("write first packet: %v", err)
	}
	firstPacketCtx, firstPacketCancel := context.WithTimeout(context.Background(), time.Second)
	firstPacket, err := WaitForPacket(firstPacketCtx, r.Packets())
	firstPacketCancel()
	if err != nil || firstPacket.DstPort != 80 {
		t.Fatalf("first packet=%+v err=%v", firstPacket, err)
	}
	if err := first.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set client read deadline: %v", err)
	}
	if _, err := first.Read(make([]byte, 1)); err == nil {
		t.Fatal("idle connection remained open")
	} else {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			t.Fatal("client timed out before the receiver closed the idle connection")
		}
	}
	_ = first.Close()

	releaseCapacity := claimReleasedConnectionCapacity(t, r)
	releaseCapacity()

	replacement, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial replacement connection: %v", err)
	}
	defer replacement.Close()
	if err := writeJSONFrame(replacement, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "idle-replacement", PID: 2, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write replacement hello: %v", err)
	}
	if err := writeJSONFrame(replacement, testPacketFrame(443)); err != nil {
		t.Fatalf("write replacement packet: %v", err)
	}
	replacementPacketCtx, replacementPacketCancel := context.WithTimeout(context.Background(), time.Second)
	replacementPacket, err := WaitForPacket(replacementPacketCtx, r.Packets())
	replacementPacketCancel()
	if err != nil || replacementPacket.DstPort != 443 {
		t.Fatalf("replacement packet=%+v err=%v", replacementPacket, err)
	}
	if got := metrics.Snapshot().DecodeErrors; got != 0 {
		t.Fatalf("idle expiry incremented decode errors: %d", got)
	}
}

func TestStartStopsActiveConnectionOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	path := filepath.Join(t.TempDir(), "netsentry.sock")
	r := New(Config{Path: path, BufferSize: 1}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}

	conn, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial receiver: %v", err)
	}
	defer conn.Close()

	cancel()
	done := make(chan struct{})
	go func() {
		r.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("receiver did not stop after context cancellation")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("socket path should be removed after stop, err=%v", err)
	}
}

func TestStartStopsMultipleActiveConnectionsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	path := filepath.Join(t.TempDir(), "netsentry.sock")
	r := New(Config{Path: path, BufferSize: 4}, zap.NewNop())
	if err := r.Start(ctx); err != nil {
		t.Fatalf("start receiver: %v", err)
	}

	first, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial first receiver connection: %v", err)
	}
	defer first.Close()
	second, err := dialUnix(path)
	if err != nil {
		t.Fatalf("dial second receiver connection: %v", err)
	}
	defer second.Close()

	if err := writeJSONFrame(first, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-one", PID: 1, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write first hello: %v", err)
	}
	if err := writeJSONFrame(second, HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "session-two", PID: 2, Hostname: "host", MaxPayloadLen: 4096}); err != nil {
		t.Fatalf("write second hello: %v", err)
	}

	cancel()
	done := make(chan struct{})
	go func() {
		r.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("receiver did not stop multiple active connections after context cancellation")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("socket path should be removed after stop, err=%v", err)
	}
	if conn, err := net.DialTimeout("unix", path, 50*time.Millisecond); err == nil {
		conn.Close()
		t.Fatal("dial should fail after receiver shutdown")
	}
}

type readDeadlineRecordingConn struct {
	net.Conn
	mu        sync.Mutex
	deadlines []time.Time
}

func waitForReceiverShutdown(t *testing.T, r *Receiver) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		r.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("receiver did not stop")
	}
}

func (c *readDeadlineRecordingConn) SetReadDeadline(deadline time.Time) error {
	c.mu.Lock()
	c.deadlines = append(c.deadlines, deadline)
	c.mu.Unlock()
	return c.Conn.SetReadDeadline(deadline)
}

func (c *readDeadlineRecordingConn) readDeadlineCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.deadlines)
}

func writeJSONFrame(conn net.Conn, frame any) error {
	b, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(b, '\n'))
	return err
}

func dialUnix(path string) (net.Conn, error) {
	var lastErr error
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("unix", path)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}
	return nil, lastErr
}

func waitForSession(t *testing.T, r *Receiver, sessionID string) {
	t.Helper()
	if waitForSessionWithin(r, sessionID, time.Second) {
		return
	}
	t.Fatalf("receiver session did not become %q: %+v", sessionID, r.State())
}

func waitForSessionWithin(r *Receiver, sessionID string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if r.State().SessionID == sessionID {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func claimReleasedConnectionCapacity(t *testing.T, r *Receiver) func() {
	t.Helper()
	if r.slots == nil {
		t.Fatal("receiver connection slots are not initialized")
	}
	select {
	case <-r.slots:
		return func() { r.slots <- struct{}{} }
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for released receiver connection capacity")
		return func() {}
	}
}

func establishedSession(sessionID string) *connectionSession {
	return &connectionSession{helloReceived: true, sessionID: sessionID}
}

func testPacketFrame(dstPort uint16) map[string]any {
	return map[string]any{
		"timestamp_sec": 1, "timestamp_usec": 2,
		"src_ip": "10.0.0.1", "dst_ip": "10.0.0.2",
		"src_port": 1, "dst_port": dstPort, "protocol": 6,
		"payload_len": 0, "payload_preview": "",
		"is_fragment": false, "truncated": false,
	}
}
