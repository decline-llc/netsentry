package receiver

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/goleak"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestHandleLineControlFrames(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	ctx := context.Background()

	if err := r.handleLine(ctx, []byte(`{"type":"hello","version":"0.1.0","session_id":"abcd1234","pid":123,"hostname":"host","max_payload_len":4096}`)); err != nil {
		t.Fatalf("hello: %v", err)
	}
	if got := r.State(); got.SessionID != "abcd1234" || got.Hello.Version != "0.1.0" {
		t.Fatalf("unexpected hello state: %+v", got)
	}

	if err := r.handleLine(ctx, []byte(`{"type":"heartbeat","session_id":"abcd1234","seq":7,"sent":11,"dropped":2,"parse_errors":3,"buf_util_pct":4,"avg_json_serialize_us":1.5,"uds_write_errors":6}`)); err != nil {
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
	if err := r.handleLine(ctx, line); err != nil {
		t.Fatalf("packet: %v", err)
	}
	pkt := <-r.Packets()
	if pkt.SrcIP != "10.0.0.1" || pkt.DstPort != 80 || pkt.Protocol != 6 {
		t.Fatalf("unexpected packet: %+v", pkt)
	}
}

func TestHandleLineBadJSON(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	if err := r.handleLine(context.Background(), []byte(`{"timestamp_sec"`)); err == nil {
		t.Fatal("expected bad JSON error")
	}
}

func TestHandleLineContextCancelWhileBlocked(t *testing.T) {
	r := New(Config{BufferSize: 1}, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())

	line := []byte(`{"timestamp_sec":1,"timestamp_usec":2,"src_ip":"10.0.0.1","dst_ip":"10.0.0.2","src_port":1,"dst_port":2,"protocol":6,"payload_len":0,"is_fragment":false,"truncated":false}`)
	if err := r.handleLine(ctx, line); err != nil {
		t.Fatalf("first packet: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- r.handleLine(ctx, line) }()

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
