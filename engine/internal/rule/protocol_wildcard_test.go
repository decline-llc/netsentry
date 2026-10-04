package rule_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/pkg/model"
)

var wildcardRuleTypes = []model.RuleType{model.RuleTypePayloadMatch, model.RuleTypeIPBlacklist, model.RuleTypePortBlacklist}

func wildcardRule(t *testing.T, kind model.RuleType, protocols []string) *model.Rule {
	t.Helper()
	var config any
	switch kind {
	case model.RuleTypePayloadMatch:
		config = model.PayloadMatchConfig{Keywords: []string{"NEEDLE"}, Protocols: protocols, Ports: []int{80}, Direction: "dest", Offset: 1, Depth: 6}
	case model.RuleTypeIPBlacklist:
		config = model.IPBlacklistConfig{IPs: []string{"198.51.100.0/24"}, Protocols: protocols, Direction: "dest"}
	case model.RuleTypePortBlacklist:
		config = model.PortBlacklistConfig{Ports: []int{80}, Protocols: protocols, Direction: "dest"}
	}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	return &model.Rule{ID: "fixture", Name: "Protocol fixture", Type: kind, Severity: model.SeverityHigh, Enabled: true, Priority: 100, Config: data}
}

func wildcardPacket(protocol uint8) model.PacketInfo {
	return model.PacketInfo{SrcIP: "192.0.2.1", DstIP: "198.51.100.9", SrcPort: 40000, DstPort: 80, Protocol: protocol, PayloadPreview: base64.StdEncoding.EncodeToString([]byte("xNEEDLE tail"))}
}

func TestExplicitAnyProtocolPreservesWildcardAcrossRuleTypes(t *testing.T) {
	cases := []struct {
		name      string
		protocols []string
		allowed   map[uint8]bool // nil means unrestricted
	}{
		{"nil", nil, nil}, {"empty", []string{}, nil},
		{"blank_only", []string{"", " \t"}, nil},
		{"any_only", []string{"any"}, nil},
		{"any_first", []string{"any", "TCP"}, nil},
		{"any_last", []string{"UDP", "any"}, nil},
		{"mixed_case_spaces_duplicates", []string{"tcp", " AnY \t", "ICMP", "any", "UDP", ""}, nil},
		{"blank_with_named", []string{"", " \t", "TCP"}, map[uint8]bool{6: true}},
		{"named_union_duplicates", []string{"tcp", "UDP", " TCP "}, map[uint8]bool{6: true, 17: true}},
	}
	for _, kind := range wildcardRuleTypes {
		for _, tc := range cases {
			t.Run(string(kind)+"/"+tc.name, func(t *testing.T) {
				engine := rule.NewEngine()
				candidate := wildcardRule(t, kind, tc.protocols)
				before := *candidate
				before.Config = append(json.RawMessage(nil), candidate.Config...)
				if err := engine.Reload([]*model.Rule{candidate}); err != nil {
					t.Fatal(err)
				}
				for _, protocol := range []uint8{0, 1, 6, 17, 255} {
					packet := wildcardPacket(protocol)
					packetBefore := packet
					var want []*model.Alert
					if tc.allowed == nil || tc.allowed[protocol] {
						preview, keyword := "", "port_blacklist: 80"
						switch kind {
						case model.RuleTypePayloadMatch:
							preview, keyword = "xNEEDLE tail", "NEEDLE"
						case model.RuleTypeIPBlacklist:
							keyword = "ip_blacklist: 198.51.100.0/24"
						}
						want = []*model.Alert{{RuleID: "fixture", RuleName: "Protocol fixture", SrcIP: packet.SrcIP, DstIP: packet.DstIP, DstPort: 80, Protocol: model.ProtocolName(protocol), Severity: model.SeverityHigh, PayloadPreview: preview, MatchedKeyword: keyword}}
					}
					if got := engine.Match(&packet); !reflect.DeepEqual(got, want) {
						t.Fatalf("protocol %d: alerts = %+v; want %+v", protocol, got, want)
					}
					if packet != packetBefore || !reflect.DeepEqual(*candidate, before) || engine.RuleCount() != 1 || !reflect.DeepEqual(engine.Rules(), []*model.Rule{candidate}) {
						t.Fatal("matching changed packet, caller rule or published config/count")
					}
				}
			})
		}
	}
}

func TestAnyProtocolPreservesOtherRuleGates(t *testing.T) {
	for _, kind := range wildcardRuleTypes {
		t.Run(string(kind), func(t *testing.T) {
			engine := rule.NewEngine()
			candidate := wildcardRule(t, kind, []string{"TCP", "any"})
			if err := engine.Reload([]*model.Rule{candidate}); err != nil {
				t.Fatal(err)
			}
			packet := wildcardPacket(17)
			if got := engine.Match(&packet); len(got) != 1 {
				t.Fatalf("wildcard control = %+v", got)
			}
			wrongSide := packet
			wrongSide.SrcIP, wrongSide.DstIP = packet.DstIP, packet.SrcIP
			wrongSide.SrcPort, wrongSide.DstPort = packet.DstPort, packet.SrcPort
			if got := engine.Match(&wrongSide); len(got) != 0 {
				t.Fatalf("wildcard bypassed direction/port/IP gate: %+v", got)
			}
			if kind == model.RuleTypePayloadMatch {
				for _, payload := range []string{"xneedle tail", "NEEDLE tail", "xxxxNEEDLE", "xNEEDL tail"} {
					miss := packet
					miss.PayloadPreview = base64.StdEncoding.EncodeToString([]byte(payload))
					if got := engine.Match(&miss); len(got) != 0 {
						t.Fatalf("wildcard bypassed case/window/keyword for %q: %+v", payload, got)
					}
				}
			}
			candidate.Enabled = false
			if err := engine.Reload([]*model.Rule{candidate}); err != nil {
				t.Fatal(err)
			}
			if got := engine.Match(&packet); len(got) != 0 || engine.RuleCount() != 1 {
				t.Fatalf("disabled wildcard rule = %+v; count=%d", got, engine.RuleCount())
			}
		})
	}
}

func TestAnyProtocolRejectsInvalidEntriesAndPreservesSnapshot(t *testing.T) {
	for _, kind := range wildcardRuleTypes {
		for _, protocols := range [][]string{{"sctp", "any"}, {"any", "sctp"}, {"TCP", "any", "sctp", "UDP"}, {"any", "sctp", "any"}} {
			for _, enabled := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%v/enabled=%v", kind, protocols, enabled), func(t *testing.T) {
					engine := rule.NewEngine()
					seed := wildcardRule(t, kind, []string{"TCP"})
					if err := engine.Reload([]*model.Rule{seed}); err != nil {
						t.Fatal(err)
					}
					packet := wildcardPacket(6)
					priorRules, priorAlerts := engine.Rules(), engine.Match(&packet)
					candidate := wildcardRule(t, kind, protocols)
					candidate.Enabled = enabled
					before := *candidate
					before.Config = append(json.RawMessage(nil), candidate.Config...)
					if err := engine.Reload([]*model.Rule{candidate}); err == nil || err.Error() != `rule fixture: unsupported protocol "sctp"` {
						t.Fatalf("invalid mixture error = %v", err)
					}
					if engine.RuleCount() != 1 || !reflect.DeepEqual(engine.Rules(), priorRules) || !reflect.DeepEqual(engine.Match(&packet), priorAlerts) || !reflect.DeepEqual(*candidate, before) {
						t.Fatal("invalid mixture changed published snapshot, matching or caller config")
					}
					valid := wildcardRule(t, kind, []string{"any", "TCP"})
					if err := engine.Reload([]*model.Rule{valid}); err != nil {
						t.Fatal(err)
					}
					udp := wildcardPacket(17)
					if got := engine.Match(&udp); len(got) != 1 || got[0].RuleID != "fixture" {
						t.Fatalf("valid retry = %+v", got)
					}
				})
			}
		}
	}
	for _, kind := range wildcardRuleTypes {
		candidate := wildcardRule(t, kind, []string{"any", "sctp"})
		var config map[string]any
		if err := json.Unmarshal(candidate.Config, &config); err != nil {
			t.Fatal(err)
		}
		config["direction"] = "sideways"
		candidate.Config, _ = json.Marshal(config)
		label := map[model.RuleType]string{model.RuleTypePayloadMatch: "payload", model.RuleTypeIPBlacklist: "ip", model.RuleTypePortBlacklist: "port"}[kind]
		if err := rule.NewEngine().Reload([]*model.Rule{candidate}); err == nil || err.Error() != fmt.Sprintf("rule fixture: unsupported %s direction %q", label, "sideways") {
			t.Fatalf("direction diagnostic for %s = %v", kind, err)
		}
		if kind == model.RuleTypePayloadMatch {
			config["direction"], config["offset"] = "dest", -1
			candidate.Config, _ = json.Marshal(config)
			if err := rule.NewEngine().Reload([]*model.Rule{candidate}); err == nil || err.Error() != "rule fixture: payload depth and offset must be non-negative" {
				t.Fatalf("window diagnostic = %v", err)
			}
		}
	}
}
