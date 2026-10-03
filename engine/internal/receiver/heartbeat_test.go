package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/stats"
)

func TestControlStateSequentialUpdatesPreserveOtherFrame(t *testing.T) {
	for _, heartbeatFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("heartbeat-first=%v", heartbeatFirst), func(t *testing.T) {
			s := newHeartbeatState()
			if got := s.Snapshot(); got != (State{}) {
				t.Fatalf("initial state = %+v, want zero", got)
			}
			hello, heartbeat := controlTestHello(1), controlTestHeartbeat(1)
			before := time.Now()
			if heartbeatFirst {
				s.SetHeartbeat(heartbeat)
				s.SetHello(hello)
			} else {
				s.SetHello(hello)
				s.SetHeartbeat(heartbeat)
			}
			assertControlStateFrames(t, s.Snapshot(), hello, heartbeat)
			stamp := s.Snapshot().LastHeartbeatAt
			if stamp.Before(before) || stamp.After(time.Now()) {
				t.Fatalf("receipt timestamp %v outside setter interval", stamp)
			}
			hello = controlTestHello(2)
			hello.SessionID = "new-hello-session"
			s.SetHello(hello)
			got := s.Snapshot()
			assertControlStateFrames(t, got, hello, heartbeat)
			if got.SessionID != hello.SessionID || got.LastHeartbeatAt != stamp {
				t.Fatalf("hello changed heartbeat time or last-setter identity: %+v", got)
			}
			heartbeat = controlTestHeartbeat(2)
			heartbeat.SessionID = "new-heartbeat-session"
			s.SetHeartbeat(heartbeat)
			got = s.Snapshot()
			assertControlStateFrames(t, got, hello, heartbeat)
			if got.SessionID != heartbeat.SessionID || got.LastHeartbeatAt.Before(stamp) {
				t.Fatalf("heartbeat changed hello or regressed receipt time: %+v", got)
			}
		})
	}
}

func TestControlStateConcurrentSettersRetainBothFramesAfterJoin(t *testing.T) {
	for round := 0; round < 256; round++ {
		s := newHeartbeatState()
		hello, heartbeat := controlTestHello(round+1), controlTestHeartbeat(round+1)
		hello.SessionID, heartbeat.SessionID = "hello-session", "heartbeat-session"
		runControlStateWriters(t, func() { s.SetHello(hello) }, func() { s.SetHeartbeat(heartbeat) })
		got := s.Snapshot()
		assertControlStateFrames(t, got, hello, heartbeat)
		if got.SessionID != hello.SessionID && got.SessionID != heartbeat.SessionID {
			t.Fatalf("round %d: last-setter identity = %q", round, got.SessionID)
		}
	}
}

func TestControlStateConcurrentReadersObserveWholePublishedFrames(t *testing.T) {
	s := newHeartbeatState()
	const updates = 1024
	ready, stop := make(chan struct{}), make(chan struct{})
	firstPairObserved := make(chan struct{})
	readerDone := make(chan error, 1)
	var once, observedOnce sync.Once
	stopReader := func() { once.Do(func() { close(stop) }) }
	releaseWriters := func() { observedOnce.Do(func() { close(firstPairObserved) }) }
	t.Cleanup(stopReader)
	t.Cleanup(releaseWriters)
	go func() {
		// Establish an observable first snapshot before releasing the writers.
		initial := s.Snapshot()
		close(ready)
		if initial != (State{}) {
			readerDone <- fmt.Errorf("initial reader state = %+v", initial)
			return
		}
		for {
			select {
			case <-stop:
				readerDone <- nil
				return
			default:
			}
			got := s.Snapshot()
			if got.Hello != (HelloFrame{}) && (got.Hello.PID < 1 || got.Hello.PID > updates || got.Hello != controlTestHello(got.Hello.PID)) {
				readerDone <- fmt.Errorf("partially published hello: %+v", got.Hello)
				return
			}
			if got.Heartbeat == (HeartbeatFrame{}) {
				if !got.LastHeartbeatAt.IsZero() {
					readerDone <- fmt.Errorf("timestamp published without heartbeat: %+v", got)
					return
				}
			} else if got.Heartbeat.Seq < 1 || got.Heartbeat.Seq > updates || got.Heartbeat != controlTestHeartbeat(int(got.Heartbeat.Seq)) || got.LastHeartbeatAt.IsZero() || got.LastHeartbeatAt.Location() != time.UTC {
				readerDone <- fmt.Errorf("partially published heartbeat/time: %+v", got)
				return
			}
			if got.SessionID != "" && got.SessionID != "shared-session" {
				readerDone <- fmt.Errorf("unexpected published session: %q", got.SessionID)
				return
			}
			if got.Hello.PID == 1 && got.Heartbeat.Seq == 1 {
				// Both writers wait for this read before continuing replacements.
				releaseWriters()
			}
		}
	}()
	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not publish initial readiness")
	}
	runControlStateWriters(t, func() {
		s.SetHello(controlTestHello(1))
		<-firstPairObserved
		for i := 2; i <= updates; i++ {
			s.SetHello(controlTestHello(i))
		}
	}, func() {
		s.SetHeartbeat(controlTestHeartbeat(1))
		<-firstPairObserved
		for i := 2; i <= updates; i++ {
			s.SetHeartbeat(controlTestHeartbeat(i))
		}
	})
	stopReader()
	select {
	case err := <-readerDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not finish")
	}
	assertControlStateFrames(t, s.Snapshot(), controlTestHello(updates), controlTestHeartbeat(updates))
}

func TestReceiverConcurrentControlFramesRetainHelloAndHeartbeat(t *testing.T) {
	for round := 0; round < 128; round++ {
		metrics := stats.New()
		r := New(Config{Stats: metrics, BufferSize: 1}, nil)
		first, second := &connectionSession{}, &connectionSession{}
		helloA, helloB, heartbeatA := controlTestHello(1), controlTestHello(2), controlTestHeartbeat(round+1)
		helloA.SessionID, helloB.SessionID, heartbeatA.SessionID = "session-a", "session-b", "session-a"
		encode := func(frame any) []byte {
			data, err := json.Marshal(frame)
			if err != nil {
				t.Fatal(err)
			}
			return data
		}
		if err := r.handleLine(context.Background(), encode(helloA), first); err != nil {
			t.Fatal(err)
		}
		helloLine, heartbeatLine := encode(helloB), encode(heartbeatA)
		errors := make(chan error, 2)
		runControlStateWriters(t, func() {
			errors <- r.handleLine(context.Background(), helloLine, second)
		}, func() {
			errors <- r.handleLine(context.Background(), heartbeatLine, first)
		})
		for i := 0; i < 2; i++ {
			if err := <-errors; err != nil {
				t.Fatalf("control frame: %v", err)
			}
		}
		got := r.State()
		assertControlStateFrames(t, got, helloB, heartbeatA)
		if got.SessionID != helloB.SessionID && got.SessionID != heartbeatA.SessionID {
			t.Fatalf("unexpected last setter identity: %+v", got)
		}
		if !first.helloReceived || first.sessionID != "session-a" || !second.helloReceived || second.sessionID != "session-b" {
			t.Fatalf("connection-local sessions changed: first=%+v second=%+v", first, second)
		}
		snapshot := metrics.Snapshot()
		if snapshot.FramesTotal != 3 || snapshot.ControlFrames != 3 || snapshot.DecodeErrors != 0 || snapshot.PacketsReceived != 0 || snapshot.PacketsProcessed != 0 || snapshot.PacketsCompleted != 0 || r.QueueDepth() != 0 {
			t.Fatalf("control-frame accounting changed: %+v queue=%d", snapshot, r.QueueDepth())
		}
	}
}

func TestControlStateSnapshotIsIndependentValue(t *testing.T) {
	s := newHeartbeatState()
	hello, heartbeat := controlTestHello(1), controlTestHeartbeat(1)
	s.SetHello(hello)
	s.SetHeartbeat(heartbeat)
	before := s.Snapshot()
	copy := before
	copy.SessionID = "changed"
	copy.Hello.Hostname = "changed"
	copy.Hello.PID = 42
	copy.Heartbeat.Sent = 0
	copy.Heartbeat.SessionID = "changed"
	copy.LastHeartbeatAt = time.Time{}
	if copy == before {
		t.Fatal("snapshot mutation fixture made no change")
	}
	if got := s.Snapshot(); got != before {
		t.Fatalf("mutating returned value changed stored state: got=%+v want=%+v", got, before)
	}
}

func runControlStateWriters(t *testing.T, first, second func()) {
	t.Helper()
	start, done := make(chan struct{}), make(chan struct{}, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for _, write := range []func(){first, second} {
		go func() {
			ready.Done()
			<-start
			write()
			done <- struct{}{}
		}()
	}
	ready.Wait()
	close(start)
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-deadline.C:
			t.Fatal("control-state writers did not finish")
		}
	}
}

func controlTestHello(sequence int) HelloFrame {
	return HelloFrame{Type: "hello", Version: "0.1.0", SessionID: "shared-session", PID: sequence, Hostname: fmt.Sprintf("host-%d", sequence), MaxPayloadLen: 4096}
}

func controlTestHeartbeat(sequence int) HeartbeatFrame {
	return HeartbeatFrame{Type: "heartbeat", SessionID: "shared-session", Seq: uint32(sequence), Sent: uint64(sequence * 3), Dropped: uint64(sequence), ParseErrors: uint64(sequence + 5), BufUtilPct: uint32(sequence % 100), AvgJSONSerializeUS: float64(sequence) / 8, UDSWriteErrors: uint64(sequence * 2)}
}

func assertControlStateFrames(t *testing.T, got State, hello HelloFrame, heartbeat HeartbeatFrame) {
	t.Helper()
	if got.Hello != hello || got.Heartbeat != heartbeat || got.LastHeartbeatAt.IsZero() || got.LastHeartbeatAt.Location() != time.UTC {
		t.Fatalf("stored frames/time disagree: got=%+v hello=%+v heartbeat=%+v", got, hello, heartbeat)
	}
}
