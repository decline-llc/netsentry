package pipeline

import (
	"context"
	"errors"
	"testing"

	alertpkg "github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type jsonRedactionBoundaryWriter struct {
	seen []model.Alert
	err  error
}

func (w *jsonRedactionBoundaryWriter) WriteBatch(_ context.Context, alerts []*model.Alert) error {
	for _, alert := range alerts {
		// Copy the payload string and metadata at writer entry, before returning.
		w.seen = append(w.seen, *alert)
	}
	return w.err
}

func TestWorkerRealJSONRedactorRemovesEscapedSuffixBeforeWriter(t *testing.T) {
	writeErr := errors.New("injected writer failure")
	input := `{"password":"prefix\"suffix-canary","token":"prefix\\\"tail-canary","public":"keep"}`
	redacted := `{"password":"[REDACTED]","token":"[REDACTED]","public":"keep"}`
	cases := []struct {
		name                 string
		enabled              bool
		writerError          error
		wantPayload          string
		completed, generated uint64
		writeErrors          uint64
	}{
		{"enabled_success", true, nil, redacted, 1, 1, 0},
		{"disabled_success", false, nil, input, 1, 1, 0},
		{"enabled_writer_failure", true, writeErr, redacted, 0, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			packet := &model.PacketInfo{TimestampSec: 1719300000, TimestampUsec: 123456, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", PayloadPreview: input}
			originalPacket := *packet
			matched := &model.Alert{ID: "alert-1", EventID: "event-1", RuleID: "rule-1", RuleName: "fixture", Severity: model.SeverityHigh, PayloadPreview: input, SrcIP: packet.SrcIP, DstIP: packet.DstIP, DstPort: 443, MatchedKeyword: "match-marker", RawPayload: "raw-marker"}
			wantAlert := *matched
			wantAlert.PayloadPreview, wantAlert.Timestamp = tc.wantPayload, packet.Timestamp().UTC()
			writer := &jsonRedactionBoundaryWriter{err: tc.writerError}
			metrics := stats.New()
			worker := NewWorker(&fakeMatcher{alerts: []*model.Alert{matched}}, writer, nil, metrics)
			if tc.enabled {
				worker.SetRedactor(alertpkg.RedactSensitivePayloads)
			}
			packets := make(chan *model.PacketInfo, 1)
			packets <- packet
			close(packets)
			worker.Run(context.Background(), packets)
			if len(writer.seen) != 1 || writer.seen[0] != wantAlert {
				t.Fatalf("writer-entry snapshot = %+v, want %+v", writer.seen, wantAlert)
			}
			if *packet != originalPacket {
				t.Fatal("redactor changed the input packet")
			}
			snapshot := metrics.Snapshot()
			if snapshot.PacketsProcessed != 1 || snapshot.PacketsCompleted != tc.completed || snapshot.AlertsGenerated != tc.generated || snapshot.AlertWriteErrors != tc.writeErrors || snapshot.AlertWriteCount != 1 || snapshot.WorkerPanics != 0 || snapshot.PacketsReceived != 0 {
				t.Fatalf("original worker accounting changed: %+v", snapshot)
			}
		})
	}
}
