package rule

import (
	"encoding/base64"
	"encoding/json"
	"runtime"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func BenchmarkEngineMatch(b *testing.B) {
	engine := NewEngine()
	if err := engine.Reload(benchmarkRuleSet(b)); err != nil {
		b.Fatalf("reload benchmark rules: %v", err)
	}

	type benchmarkCase struct {
		name       string
		rawPayload []byte
		packet     *model.PacketInfo
		want       map[string]struct{}
	}
	cases := []benchmarkCase{
		{
			name:       "no_hit",
			rawPayload: []byte("GET /status HTTP/1.1\r\nHost: example.test\r\n\r\nordinary request body"),
			packet: &model.PacketInfo{
				SrcIP:    "192.0.2.10",
				DstIP:    "192.0.2.20",
				SrcPort:  32100,
				DstPort:  443,
				Protocol: 6,
			},
			want: map[string]struct{}{},
		},
		{
			name:       "multi_hit",
			rawPayload: []byte("POST /search?q=UNION SELECT HTTP/1.1\r\n\r\n../../etc/passwd;/bin/sh"),
			packet: &model.PacketInfo{
				SrcIP:    "198.51.100.42",
				DstIP:    "203.0.113.8",
				SrcPort:  4444,
				DstPort:  80,
				Protocol: 6,
			},
			want: map[string]struct{}{
				"payload-sql":       {},
				"payload-traversal": {},
				"payload-shell":     {},
				"ip-source":         {},
				"ip-dest-net":       {},
				"port-source":       {},
				"port-dest":         {},
			},
		},
	}

	for _, benchmark := range cases {
		benchmark.packet.PayloadLen = uint16(len(benchmark.rawPayload))
		benchmark.packet.PayloadPreview = base64.StdEncoding.EncodeToString(benchmark.rawPayload)

		b.Run(benchmark.name, func(b *testing.B) {
			alerts := engine.Match(benchmark.packet)
			assertBenchmarkAlertIDs(b, alerts, benchmark.want)

			b.ReportAllocs()
			b.SetBytes(int64(len(benchmark.rawPayload)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				alerts = engine.Match(benchmark.packet)
			}
			b.StopTimer()

			runtime.KeepAlive(alerts)
			assertBenchmarkAlertIDs(b, alerts, benchmark.want)
		})
	}
}

func benchmarkRuleSet(b *testing.B) []*model.Rule {
	b.Helper()
	return []*model.Rule{
		benchmarkRule(b, "payload-sql", model.RuleTypePayloadMatch, 800, model.PayloadMatchConfig{
			Keywords:        []string{"union select", "drop table"},
			CaseInsensitive: true,
			Protocols:       []string{"tcp"},
			Ports:           []int{80},
			Direction:       "dest",
			Depth:           4096,
		}),
		benchmarkRule(b, "payload-traversal", model.RuleTypePayloadMatch, 700, model.PayloadMatchConfig{
			Keywords:  []string{"../", "/etc/passwd"},
			Protocols: []string{"tcp"},
			Ports:     []int{80},
			Direction: "dest",
			Depth:     4096,
		}),
		benchmarkRule(b, "payload-shell", model.RuleTypePayloadMatch, 600, model.PayloadMatchConfig{
			Keywords:        []string{"/bin/sh", "cmd.exe"},
			CaseInsensitive: true,
			Protocols:       []string{"tcp"},
			Ports:           []int{80},
			Direction:       "dest",
			Depth:           4096,
		}),
		benchmarkRule(b, "ip-source", model.RuleTypeIPBlacklist, 500, model.IPBlacklistConfig{
			IPs:       []string{"198.51.100.42"},
			Direction: "source",
			Protocols: []string{"tcp"},
		}),
		benchmarkRule(b, "ip-dest-net", model.RuleTypeIPBlacklist, 400, model.IPBlacklistConfig{
			IPs:       []string{"203.0.113.0/24"},
			Direction: "dest",
			Protocols: []string{"tcp"},
		}),
		benchmarkRule(b, "port-source", model.RuleTypePortBlacklist, 300, model.PortBlacklistConfig{
			Ports:     []int{4444},
			Protocols: []string{"tcp"},
			Direction: "source",
		}),
		benchmarkRule(b, "port-dest", model.RuleTypePortBlacklist, 200, model.PortBlacklistConfig{
			Ports:     []int{80},
			Protocols: []string{"tcp"},
			Direction: "dest",
		}),
	}
}

func benchmarkRule(b *testing.B, id string, ruleType model.RuleType, priority int, config any) *model.Rule {
	b.Helper()
	raw, err := json.Marshal(config)
	if err != nil {
		b.Fatalf("marshal benchmark rule %s: %v", id, err)
	}
	return &model.Rule{
		ID:       id,
		Name:     id,
		Type:     ruleType,
		Severity: model.SeverityHigh,
		Priority: priority,
		Enabled:  true,
		Config:   raw,
	}
}

func assertBenchmarkAlertIDs(b *testing.B, alerts []*model.Alert, want map[string]struct{}) {
	b.Helper()
	if len(alerts) != len(want) {
		b.Fatalf("got %d alerts, want %d: %+v", len(alerts), len(want), alerts)
	}
	seen := make(map[string]struct{}, len(alerts))
	for _, alert := range alerts {
		if alert == nil {
			b.Fatal("got nil alert")
		}
		if _, ok := want[alert.RuleID]; !ok {
			b.Fatalf("unexpected alert rule ID %q", alert.RuleID)
		}
		if _, duplicate := seen[alert.RuleID]; duplicate {
			b.Fatalf("duplicate alert rule ID %q", alert.RuleID)
		}
		seen[alert.RuleID] = struct{}{}
	}
}
