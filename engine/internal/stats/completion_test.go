package stats

import (
	"strings"
	"sync"
	"testing"
)

func TestPacketCompletionCounterNilZeroAndIndependence(t *testing.T) {
	var absent *Stats
	absent.IncPacketCompleted()
	for _, snapshot := range []Snapshot{absent.Snapshot(), New().Snapshot()} {
		if snapshot.PacketsCompleted != 0 {
			t.Fatalf("initial completion = %d, want 0", snapshot.PacketsCompleted)
		}
		if body := RenderPrometheus(snapshot, nil); !strings.Contains(body, "\nnetsentry_packets_completed_total 0\n") {
			t.Fatalf("missing zero completion counter: %s", body)
		}
	}
	s := New()
	s.IncPacketReceived()
	s.IncPacketProcessed()
	s.IncPacketProcessed()
	for i := 0; i < 3; i++ {
		s.IncPacketCompleted()
	}
	snapshot := s.Snapshot()
	if snapshot.PacketsCompleted != 3 || snapshot.PacketsProcessed != 2 || snapshot.PacketsReceived != 1 {
		t.Fatalf("new counter changed legacy counts: %+v", snapshot)
	}
	body := RenderPrometheus(snapshot, nil)
	for _, line := range []string{
		"# HELP netsentry_packets_completed_total Packets completing pipeline processing and any enabled lifecycle export.",
		"# TYPE netsentry_packets_completed_total counter",
		"netsentry_packets_completed_total 3",
		"# HELP netsentry_packets_processed_total Packets processed by the pipeline.",
		"netsentry_packets_processed_total 2",
		"netsentry_packets_received_total 1",
	} {
		if strings.Count(body, line+"\n") != 1 {
			t.Fatalf("expected one exact export line %q in:\n%s", line, body)
		}
	}
}

func TestPacketCompletionCounterConcurrentIncrements(t *testing.T) {
	s := New()
	start := make(chan struct{})
	var writers sync.WaitGroup
	for i := 0; i < 4; i++ {
		writers.Add(1)
		go func() {
			defer writers.Done()
			<-start
			for j := 0; j < 1000; j++ {
				s.IncPacketCompleted()
			}
		}()
	}
	close(start)
	writers.Wait()
	if snapshot := s.Snapshot(); snapshot.PacketsCompleted != 4000 || snapshot.PacketsProcessed != 0 ||
		snapshot.PacketsReceived != 0 {
		t.Fatalf("concurrent independent count: %+v", snapshot)
	}
}
