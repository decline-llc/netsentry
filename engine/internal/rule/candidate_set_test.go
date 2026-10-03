package rule

import (
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestPayloadCandidateSetPreservesPerRuleAlertSelection(t *testing.T) {
	payloadRule := func(id string, priority int, cfg model.PayloadMatchConfig) *model.Rule {
		r := makePayloadRuleWithConfig(id, cfg)
		r.Priority = priority
		return r
	}
	critical := payloadRule("critical", 300, model.PayloadMatchConfig{Keywords: []string{"needle"}})
	critical.Severity, critical.EarlyExit = model.SeverityCritical, true
	disabled := payloadRule("disabled", 200, model.PayloadMatchConfig{Keywords: []string{"needle"}})
	disabled.Enabled = false
	type expectedHit struct {
		id, keyword string
		severity    model.Severity
	}
	cases := []struct {
		name    string
		rules   []*model.Rule
		payload string
		hits    []expectedHit
	}{
		{"duplicate_keywords_original_order", []*model.Rule{payloadRule("duplicate", 100, model.PayloadMatchConfig{Keywords: []string{"needle", "needle", "need"}})}, "needle", []expectedHit{{"duplicate", "needle", model.SeverityHigh}}},
		{"shared_keywords", []*model.Rule{payloadRule("low", 100, model.PayloadMatchConfig{Keywords: []string{"needle"}}), payloadRule("high", 200, model.PayloadMatchConfig{Keywords: []string{"needle"}})}, "needle", []expectedHit{{"high", "needle", model.SeverityHigh}, {"low", "needle", model.SeverityHigh}}},
		{"mixed_case", []*model.Rule{payloadRule("sensitive", 200, model.PayloadMatchConfig{Keywords: []string{"NEEDLE"}}), payloadRule("insensitive", 100, model.PayloadMatchConfig{Keywords: []string{"NEEDLE"}, CaseInsensitive: true})}, "needle", []expectedHit{{"insensitive", "NEEDLE", model.SeverityHigh}}},
		{"window_hit", []*model.Rule{payloadRule("window", 100, model.PayloadMatchConfig{Keywords: []string{"needle"}, Offset: 4, Depth: 6})}, "xxxxneedle tail", []expectedHit{{"window", "needle", model.SeverityHigh}}},
		{"window_miss", []*model.Rule{payloadRule("window", 100, model.PayloadMatchConfig{Keywords: []string{"needle"}, Offset: 4, Depth: 4})}, "xxxxneedle tail", nil},
		{"disabled_shared_keyword", []*model.Rule{disabled, payloadRule("enabled", 100, model.PayloadMatchConfig{Keywords: []string{"needle"}})}, "needle", []expectedHit{{"enabled", "needle", model.SeverityHigh}}},
		{"protocol_port_rejection", []*model.Rule{payloadRule("udp-only", 200, model.PayloadMatchConfig{Keywords: []string{"needle"}, Protocols: []string{"UDP"}}), payloadRule("port-only", 100, model.PayloadMatchConfig{Keywords: []string{"needle"}, Ports: []int{81}})}, "needle", nil},
		{"critical_early_exit", []*model.Rule{payloadRule("low", 100, model.PayloadMatchConfig{Keywords: []string{"needle"}}), critical}, "needle", []expectedHit{{"critical", "needle", model.SeverityCritical}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Reload(tc.rules); err != nil {
				t.Fatal(err)
			}
			packet := &model.PacketInfo{SrcIP: "198.51.100.1", DstIP: "192.0.2.1", SrcPort: 40000, DstPort: 80, Protocol: 6, PayloadPreview: b64(tc.payload)}
			before := *packet
			var want []*model.Alert
			for _, hit := range tc.hits {
				want = append(want, &model.Alert{RuleID: hit.id, RuleName: hit.id, SrcIP: "198.51.100.1", DstIP: "192.0.2.1", DstPort: 80, Protocol: "TCP", Severity: hit.severity, PayloadPreview: tc.payload, MatchedKeyword: hit.keyword})
			}
			if got := engine.Match(packet); !reflect.DeepEqual(got, want) {
				t.Fatalf("alerts = %+v, want %+v", got, want)
			}
			if *packet != before || engine.RuleCount() != len(tc.rules) {
				t.Fatalf("packet or loaded count changed: packet %+v, count %d", packet, engine.RuleCount())
			}
		})
	}
}
