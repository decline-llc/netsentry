package rule_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/pkg/model"
)

func emptyIPFixture(t *testing.T, id string, enabled bool, cfg model.IPBlacklistConfig) *model.Rule {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &model.Rule{ID: id, Name: "Address fixture", Type: model.RuleTypeIPBlacklist, Severity: model.SeverityHigh, Priority: 100, Enabled: enabled, Config: data}
}

func assertIPSnapshot(t *testing.T, e *rule.Engine, rules []*model.Rule, packet *model.PacketInfo, alerts []*model.Alert) {
	t.Helper()
	if e.RuleCount() != len(rules) || !reflect.DeepEqual(e.Rules(), rules) || !reflect.DeepEqual(e.Match(packet), alerts) {
		t.Fatal("rejected reload changed rule count, published rules or matching")
	}
}

func TestEmptyIPBlacklistReloadRejectsAndPreservesSnapshot(t *testing.T) {
	cases := []struct {
		name string
		ips  []string
	}{
		{"nil", nil}, {"empty", []string{}}, {"empty_entry", []string{""}},
		{"ascii_whitespace", []string{" \t\r\n"}}, {"unicode_whitespace", []string{"\u2003\u00a0"}},
		{"multiple_blanks", []string{"", " \t", "\u2003"}},
	}
	for _, enabled := range []bool{true, false} {
		for _, tc := range cases {
			t.Run(tc.name+"/enabled="+map[bool]string{true: "true", false: "false"}[enabled], func(t *testing.T) {
				e := rule.NewEngine()
				original := emptyIPFixture(t, "original", true, model.IPBlacklistConfig{IPs: []string{"198.51.100.1"}})
				if err := e.Reload([]*model.Rule{original}); err != nil {
					t.Fatal(err)
				}
				packet := &model.PacketInfo{SrcIP: "198.51.100.1", DstIP: "192.0.2.1", Protocol: 6}
				packetBefore := *packet
				wantRules, wantAlerts := e.Rules(), e.Match(packet)
				if len(wantAlerts) != 1 || wantAlerts[0].RuleID != "original" {
					t.Fatalf("fixture alerts: %+v", wantAlerts)
				}
				invalid := emptyIPFixture(t, "blank", enabled, model.IPBlacklistConfig{IPs: tc.ips})
				before := *invalid
				before.Config = append(json.RawMessage(nil), invalid.Config...)
				candidate := []*model.Rule{original, invalid}
				pointers := append([]*model.Rule(nil), candidate...)
				err := e.Reload(candidate)
				if err == nil || err.Error() != "rule blank: ip blacklist requires at least one IP or CIDR" {
					t.Fatalf("error = %v", err)
				}
				assertIPSnapshot(t, e, wantRules, packet, wantAlerts)
				if !reflect.DeepEqual(*invalid, before) || !reflect.DeepEqual(candidate, pointers) || *packet != packetBefore {
					t.Fatal("rejection changed candidate or packet")
				}
				replacement := emptyIPFixture(t, "replacement", true, model.IPBlacklistConfig{IPs: []string{"", " 198.51.100.1 "}})
				if err := e.Reload([]*model.Rule{replacement}); err != nil {
					t.Fatal(err)
				}
				if got := e.Match(packet); e.RuleCount() != 1 || len(got) != 1 || got[0].RuleID != "replacement" {
					t.Fatalf("valid retry failed: %+v", got)
				}
			})
		}
	}
	// The new post-loop guard must not obscure errors already detected earlier.
	for _, tc := range []struct {
		name       string
		cfg        model.IPBlacklistConfig
		diagnostic string
	}{
		{"direction", model.IPBlacklistConfig{IPs: []string{""}, Direction: "sideways"}, "direction"},
		{"protocol", model.IPBlacklistConfig{IPs: []string{""}, Protocols: []string{"sctp"}}, "protocol"},
		{"invalid_ip", model.IPBlacklistConfig{IPs: []string{"", "not-an-ip"}}, "invalid IP"},
		{"invalid_cidr", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.1/99"}}, "invalid CIDR"},
	} {
		t.Run("diagnostic/"+tc.name, func(t *testing.T) {
			e := rule.NewEngine()
			err := e.Reload([]*model.Rule{emptyIPFixture(t, "bad", true, tc.cfg)})
			if err == nil || !strings.Contains(err.Error(), tc.diagnostic) || strings.Contains(err.Error(), "requires at least") {
				t.Fatalf("error = %v", err)
			}
			if e.RuleCount() != 0 || len(e.Rules()) != 0 || len(e.Match(&model.PacketInfo{})) != 0 {
				t.Fatal("invalid rule published")
			}
		})
	}
}

func TestNonemptyIPBlacklistRetainsMatchingFiltersAndOwnership(t *testing.T) {
	cases := []struct {
		name    string
		cfg     model.IPBlacklistConfig
		enabled bool
		packet  model.PacketInfo
		match   bool
	}{
		{"padded_exact", model.IPBlacklistConfig{IPs: []string{" 198.51.100.1\t"}}, true, model.PacketInfo{SrcIP: "198.51.100.1"}, true},
		{"padded_cidr", model.IPBlacklistConfig{IPs: []string{" 198.51.100.0/24 "}}, true, model.PacketInfo{DstIP: "198.51.100.42"}, true},
		{"mixed_duplicates", model.IPBlacklistConfig{IPs: []string{"", "\u2003", "198.51.100.1", "198.51.100.1", "198.51.100.0/24"}}, true, model.PacketInfo{SrcIP: "198.51.100.1"}, true},
		{"source_tcp", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.1"}, Direction: "source", Protocols: []string{"tcp"}}, true, model.PacketInfo{SrcIP: "198.51.100.1", Protocol: 6}, true},
		{"source_wrong_side", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.1"}, Direction: "source"}, true, model.PacketInfo{DstIP: "198.51.100.1"}, false},
		{"source_wrong_protocol", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.1"}, Direction: "source", Protocols: []string{"tcp"}}, true, model.PacketInfo{SrcIP: "198.51.100.1", Protocol: 17}, false},
		{"dest_udp", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.0/24"}, Direction: "dest", Protocols: []string{"udp"}}, true, model.PacketInfo{DstIP: "198.51.100.42", Protocol: 17}, true},
		{"dest_wrong_side", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.0/24"}, Direction: "dest"}, true, model.PacketInfo{SrcIP: "198.51.100.42"}, false},
		{"any_source", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.0/24"}, Direction: "any"}, true, model.PacketInfo{SrcIP: "198.51.100.42"}, true},
		{"any_dest", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.0/24"}, Direction: "any"}, true, model.PacketInfo{DstIP: "198.51.100.42"}, true},
		{"outside", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.0/24"}}, true, model.PacketInfo{SrcIP: "203.0.113.1", DstIP: "192.0.2.1"}, false},
		{"disabled", model.IPBlacklistConfig{IPs: []string{"", "198.51.100.1"}}, false, model.PacketInfo{SrcIP: "198.51.100.1"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rule.NewEngine()
			r := emptyIPFixture(t, "valid", tc.enabled, tc.cfg)
			before := *r
			before.Config = append(json.RawMessage(nil), r.Config...)
			packetBefore := tc.packet
			if err := e.Reload([]*model.Rule{r}); err != nil {
				t.Fatal(err)
			}
			got := e.Match(&tc.packet)
			want := 0
			if tc.match {
				want = 1
			}
			if len(got) != want || (want == 1 && got[0].RuleID != "valid") || e.RuleCount() != 1 {
				t.Fatalf("alerts=%+v count=%d", got, e.RuleCount())
			}
			if !reflect.DeepEqual(*r, before) || tc.packet != packetBefore {
				t.Fatal("valid reload/match changed inputs")
			}
		})
	}
	e := rule.NewEngine()
	if err := e.Reload([]*model.Rule{
		emptyIPFixture(t, "cidr", true, model.IPBlacklistConfig{IPs: []string{"", "198.51.100.0/24"}}),
		emptyIPFixture(t, "exact", true, model.IPBlacklistConfig{IPs: []string{"", "203.0.113.10"}}),
	}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ ip, id string }{{"198.51.100.42", "cidr"}, {"203.0.113.10", "exact"}} {
		got := e.Match(&model.PacketInfo{SrcIP: tc.ip})
		if len(got) != 1 || got[0].RuleID != tc.id {
			t.Fatalf("owning rule scope: %+v", got)
		}
	}
}

func TestLoadedIPBlacklistValidationPreservesFileAndPublishedState(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		for _, legacy := range []bool{false, true} {
			for _, valid := range []bool{false, true} {
				t.Run(map[bool]string{true: "wrapped", false: "array"}[wrapped]+"/"+map[bool]string{true: "legacy", false: "canonical"}[legacy]+"/"+map[bool]string{true: "valid", false: "blank"}[valid], func(t *testing.T) {
					ips := []string{"", "\u2003"}
					if valid {
						ips = append(ips, " 198.51.100.1 ")
					}
					configKey := "config"
					if legacy {
						configKey = "ip_blacklist"
					}
					entries := []map[string]any{{"id": "loaded", "name": "Loaded fixture", "type": "ip_blacklist", "severity": "high", "enabled": true, configKey: map[string]any{"ips": ips}}}
					var value any = entries
					if wrapped {
						value = map[string]any{"rules": entries}
					}
					data, err := json.MarshalIndent(value, "", "  ")
					if err != nil {
						t.Fatal(err)
					}
					data = append(data, '\n')
					path := filepath.Join(t.TempDir(), "rules with spaces.json")
					if err := os.WriteFile(path, data, 0o600); err != nil {
						t.Fatal(err)
					}
					e := rule.NewEngine()
					if err := e.Reload([]*model.Rule{emptyIPFixture(t, "original", true, model.IPBlacklistConfig{IPs: []string{"198.51.100.1"}})}); err != nil {
						t.Fatal(err)
					}
					packet := &model.PacketInfo{SrcIP: "198.51.100.1"}
					wantRules, wantAlerts := e.Rules(), e.Match(packet)
					loaded, err := rule.LoadFromFile(path)
					if err != nil || len(loaded) != 1 {
						t.Fatalf("parse: rules=%+v err=%v", loaded, err)
					}
					before := *loaded[0]
					before.Config = append(json.RawMessage(nil), loaded[0].Config...)
					err = e.Reload(loaded)
					if valid {
						if err != nil {
							t.Fatal(err)
						}
						got := e.Match(packet)
						if e.RuleCount() != 1 || len(got) != 1 || got[0].RuleID != "loaded" {
							t.Fatalf("loaded match: %+v", got)
						}
					} else {
						if err == nil || err.Error() != "rule loaded: ip blacklist requires at least one IP or CIDR" {
							t.Fatalf("reload error=%v", err)
						}
						assertIPSnapshot(t, e, wantRules, packet, wantAlerts)
					}
					after, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(after, data) || !reflect.DeepEqual(*loaded[0], before) {
						t.Fatal("load/reload changed file or loaded rule")
					}
				})
			}
		}
	}
}
