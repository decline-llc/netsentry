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

func TestWorkerRedactsRealEngineEscapedJSONNamesBeforeWriter(t *testing.T) {
	for _, key := range []string{`pass\u0077ord`, `\u0054OKEN`, `\u0074\u006f\u006b\u0065\u006e`} {
		for _, cut := range []string{"complete", "plain", "dangling_escape", "partial_unicode"} {
			prefix := `{"public":"keep","` + key + `" : "`
			payload := prefix + `quote\"tail-canary","public2":"stay"}`
			preview, redacted := payload, prefix+`[REDACTED]","public2":"stay"}`
			if cut != "complete" {
				tail, completion := "canary", ""
				switch cut {
				case "dangling_escape":
					tail, completion = "canary\\", "n"
				case "partial_unicode":
					tail, completion = `canary\u00`, "41"
				}
				preview = prefix + strings.Repeat("x", 200-len(prefix)-len(tail)) + tail
				payload = preview + completion + strings.Repeat("y", 100) + `"}`
				redacted = prefix + `[REDACTED]`
			}
			for _, mode := range []string{"enabled", "disabled", "writer_failure"} {
				t.Run(key+"/"+cut+"/"+mode, func(t *testing.T) {
					if !json.Valid([]byte(payload)) || (cut != "complete" && (len(payload) <= 200 || len(preview) != 200 || json.Valid([]byte(preview)))) {
						t.Fatal("fixture does not reach promised complete/truncated boundary")
					}
					engine := rule.NewEngine()
					candidate := &model.Rule{ID: "escaped-name", Name: "key fixture", Type: model.RuleTypePayloadMatch, Enabled: true, Severity: model.SeverityHigh, Config: json.RawMessage(`{"keywords":["keep"]}`)}
					if err := engine.Reload([]*model.Rule{candidate}); err != nil {
						t.Fatal(err)
					}
					packet := &model.PacketInfo{TimestampSec: 1719300000, TimestampUsec: 123456, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, Protocol: 6, PayloadPreview: base64.StdEncoding.EncodeToString([]byte(payload))}
					packetBefore := *packet
					matched := engine.Match(packet)
					if len(matched) != 1 || matched[0].PayloadPreview != preview {
						t.Fatalf("actual Engine preview count/content = %+v, want %q", matched, preview)
					}
					writer := &jsonRedactionBoundaryWriter{}
					metrics := stats.New()
					worker := NewWorker(engine, writer, nil, metrics)
					wantPayload := redacted
					completed, generated, writeErrors := uint64(1), uint64(1), uint64(0)
					if mode == "disabled" {
						wantPayload = preview
					} else {
						worker.SetRedactor(alertpkg.RedactSensitivePayloads)
					}
					if mode == "writer_failure" {
						writer.err = errors.New("injected writer failure")
						completed, generated, writeErrors = 0, 0, 1
					}
					want := model.Alert{RuleID: "escaped-name", RuleName: "key fixture", Timestamp: packet.Timestamp().UTC(), SrcIP: packet.SrcIP, DstIP: packet.DstIP, DstPort: 443, Protocol: "TCP", Severity: model.SeverityHigh, PayloadPreview: wantPayload, MatchedKeyword: "keep"}
					packets := make(chan *model.PacketInfo, 1)
					packets <- packet
					close(packets)
					worker.Run(context.Background(), packets)
					if len(writer.seen) != 1 || writer.seen[0] != want || *packet != packetBefore || matched[0].PayloadPreview != preview {
						t.Fatalf("writer boundary or packet changed: %+v, want %+v", writer.seen, want)
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
