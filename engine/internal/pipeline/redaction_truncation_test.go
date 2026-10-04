package pipeline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	alertpkg "github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestWorkerRedactsActualEngineTruncatedJSONBeforeWriter(t *testing.T) {
	cuts := []struct{ name, tail, completion string }{
		{"plain", "canary", ""},
		{"dangling_escape", "canary\\", "n"},
		{"escaped_quote", `canary\"`, ""},
		{"partial_unicode", `canary\u00`, "41"},
		{"even_backslashes", `canary\\`, ""},
	}
	writeErr := errors.New("injected writer failure")
	for _, key := range []string{"password", "TOKEN"} {
		for _, cut := range cuts {
			for _, mode := range []string{"enabled", "disabled", "writer_failure"} {
				t.Run(key+"/"+cut.name+"/"+mode, func(t *testing.T) {
					prefix := `{"public":"keep","` + key + `":"`
					preview := prefix + strings.Repeat("x", 200-len(prefix)-len(cut.tail)) + cut.tail
					payload := preview + cut.completion + strings.Repeat("y", 100) + `"}`
					if len(payload) <= 200 || len(preview) != 200 || !json.Valid([]byte(payload)) || json.Valid([]byte(preview)) {
						t.Fatalf("fixture does not cross actual preview boundary: %q", payload)
					}
					engine := rule.NewEngine()
					candidate := &model.Rule{ID: "truncated-json", Name: "preview fixture", Type: model.RuleTypePayloadMatch, Enabled: true, Severity: model.SeverityHigh, Config: json.RawMessage(`{"keywords":["keep"]}`)}
					if err := engine.Reload([]*model.Rule{candidate}); err != nil {
						t.Fatal(err)
					}
					packet := &model.PacketInfo{TimestampSec: 1719300000, TimestampUsec: 123456, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, Protocol: 6, PayloadPreview: base64.StdEncoding.EncodeToString([]byte(payload))}
					originalPacket := *packet
					matched := engine.Match(packet)
					if len(matched) != 1 || matched[0].PayloadPreview != preview || len(matched[0].PayloadPreview) != 200 || matched[0].MatchedKeyword != "keep" {
						t.Fatalf("actual matcher count/preview = %+v, want one 200-byte preview", matched)
					}
					wantAlert := *matched[0]
					wantAlert.Timestamp = packet.Timestamp().UTC()
					if mode != "disabled" {
						wantAlert.PayloadPreview = prefix + "[REDACTED]"
					}
					writer := &jsonRedactionBoundaryWriter{}
					completed, generated, writeErrors := uint64(1), uint64(1), uint64(0)
					if mode == "writer_failure" {
						writer.err = writeErr
						completed, generated, writeErrors = 0, 0, 1
					}
					metrics := stats.New()
					worker := NewWorker(engine, writer, nil, metrics)
					if mode != "disabled" {
						worker.SetRedactor(alertpkg.RedactSensitivePayloads)
					}
					packets := make(chan *model.PacketInfo, 1)
					packets <- packet
					close(packets)
					worker.Run(context.Background(), packets)
					if len(writer.seen) != 1 || writer.seen[0] != wantAlert {
						t.Fatalf("writer-entry count/content = %+v, want %+v", writer.seen, wantAlert)
					}
					if *packet != originalPacket || matched[0].PayloadPreview != preview {
						t.Fatal("worker changed packet or earlier independent match")
					}
					snapshot := metrics.Snapshot()
					if snapshot.PacketsProcessed != 1 || snapshot.PacketsCompleted != completed || snapshot.AlertsGenerated != generated || snapshot.AlertWriteErrors != writeErrors || snapshot.AlertWriteCount != 1 || snapshot.WorkerPanics != 0 {
						t.Fatalf("worker accounting changed: %+v", snapshot)
					}
				})
			}
		}
	}
}
