package rule_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/pkg/model"
)

func identityRule(t *testing.T, id string, cfg model.IPBlacklistConfig) *model.Rule {
	t.Helper()
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &model.Rule{ID: id, Name: id, Type: model.RuleTypeIPBlacklist,
		Severity: model.SeverityHigh, Priority: 100, Enabled: true, Config: raw}
}

func assertIdentityAlert(t *testing.T, alerts []*model.Alert, r *model.Rule, p *model.PacketInfo, reason string) {
	t.Helper()
	want := &model.Alert{RuleID: r.ID, RuleName: r.Name, SrcIP: p.SrcIP,
		DstIP: p.DstIP, DstPort: p.DstPort, Protocol: model.ProtocolName(p.Protocol),
		Severity: r.Severity, MatchedKeyword: "ip_blacklist: " + reason}
	if len(alerts) != 1 || !reflect.DeepEqual(alerts[0], want) {
		t.Fatalf("alerts = %+v, want exactly %+v", alerts, want)
	}
}

func TestExactIPBlacklistMatchesEquivalentAddressSpellings(t *testing.T) {
	cases := []struct {
		name    string
		ips     []string
		address string
		match   bool
	}{
		{"native_ipv4", []string{"198.51.100.1"}, "198.51.100.1", true},
		{"mapped_dotted_rule", []string{"::ffff:198.51.100.1"}, "198.51.100.1", true},
		{"mapped_hex_rule", []string{"::ffff:c633:6401"}, "198.51.100.1", true},
		{"mapped_expanded_rule", []string{"0000:0000:0000:0000:0000:FFFF:C633:6401"}, "198.51.100.1", true},
		{"mapped_dotted_packet", []string{"198.51.100.1"}, "::ffff:198.51.100.1", true},
		{"mapped_hex_packet", []string{"198.51.100.1"}, "::FFFF:C633:6401", true},
		{"expanded_mapped_packet", []string{"::ffff:198.51.100.1"}, "0:0:0:0:0:ffff:c633:6401", true},
		{"ipv6_expanded_rule", []string{"2001:0DB8:0000:0000:0000:0000:0000:00AB"}, "2001:db8::ab", true},
		{"ipv6_expanded_packet", []string{"2001:db8::ab"}, "2001:0DB8:0000:0000:0000:0000:0000:00AB", true},
		{"ipv6_case", []string{"2001:DB8::AB"}, "2001:db8::ab", true},
		{"mixed_blanks_duplicates", []string{"", " \t", " ::FFFF:198.51.100.1 ", "198.51.100.1", "::ffff:c633:6401"}, "198.51.100.1", true},
		{"unequal_ipv4", []string{"::ffff:198.51.100.1"}, "198.51.100.2", false},
		{"unequal_ipv6", []string{"2001:DB8::AB"}, "2001:db8::ac", false},
		{"compatible_ipv6_is_not_mapped", []string{"198.51.100.1"}, "::198.51.100.1", false},
		{"invalid_packet", []string{"198.51.100.1"}, "not-an-ip", false},
		{"spaced_packet", []string{"198.51.100.1"}, " 198.51.100.1 ", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := identityRule(t, "exact", model.IPBlacklistConfig{IPs: tc.ips, Direction: "source"})
			before := *r
			before.Config = append(json.RawMessage(nil), r.Config...)
			e := rule.NewEngine()
			if err := e.Reload([]*model.Rule{r}); err != nil {
				t.Fatal(err)
			}
			packet := &model.PacketInfo{SrcIP: tc.address, DstIP: "192.0.2.1", DstPort: 80, Protocol: 6}
			packetBefore := *packet
			alerts := e.Match(packet)
			if tc.match {
				assertIdentityAlert(t, alerts, r, packet, tc.address)
			} else if len(alerts) != 0 {
				t.Fatalf("unexpected alerts: %+v", alerts)
			}
			if *packet != packetBefore || !reflect.DeepEqual(*r, before) ||
				e.RuleCount() != 1 || !reflect.DeepEqual(e.Rules(), []*model.Rule{r}) {
				t.Fatal("match/reload changed packet, rule config or published snapshot")
			}
		})
	}
}

func TestExactIPIdentityRetainsFiltersAndCIDRPrecedence(t *testing.T) {
	cases := []struct {
		name      string
		direction string
		src, dst  string
		protocol  uint8
		enabled   bool
		reason    string
	}{
		{"source", "source", "198.51.100.1", "192.0.2.1", 6, true, "198.51.100.1"},
		{"source_ignores_destination", "source", "192.0.2.1", "198.51.100.1", 6, true, ""},
		{"destination", "dest", "192.0.2.1", "198.51.100.1", 6, true, "198.51.100.1"},
		{"destination_ignores_source", "dest", "198.51.100.1", "192.0.2.1", 6, true, ""},
		{"any_source_first", "any", "::FFFF:C633:6401", "198.51.100.1", 6, true, "::FFFF:C633:6401"},
		{"any_destination", "any", "192.0.2.1", "::ffff:198.51.100.1", 6, true, "::ffff:198.51.100.1"},
		{"protocol_rejects", "any", "198.51.100.1", "192.0.2.1", 17, true, ""},
		{"disabled", "any", "198.51.100.1", "192.0.2.1", 6, false, ""},
		{"cidr_only_hit", "source", "198.51.100.2", "192.0.2.1", 6, true, "198.51.100.0/24"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := identityRule(t, "filtered", model.IPBlacklistConfig{
				IPs:       []string{"198.51.100.0/24", "::ffff:198.51.100.1"},
				Direction: tc.direction, Protocols: []string{"tcp"}})
			r.Enabled = tc.enabled
			e := rule.NewEngine()
			if err := e.Reload([]*model.Rule{r}); err != nil {
				t.Fatal(err)
			}
			packet := &model.PacketInfo{SrcIP: tc.src, DstIP: tc.dst, DstPort: 80, Protocol: tc.protocol}
			before := *packet
			alerts := e.Match(packet)
			if tc.reason == "" {
				if len(alerts) != 0 {
					t.Fatalf("unexpected alerts: %+v", alerts)
				}
			} else {
				assertIdentityAlert(t, alerts, r, packet, tc.reason)
			}
			if *packet != before || !reflect.DeepEqual(e.Rules(), []*model.Rule{r}) {
				t.Fatal("filters modified input or snapshot")
			}
		})
	}
}

func TestExactIPIdentityRetainsRuleScopeOrderingAndRejectedReload(t *testing.T) {
	e := rule.NewEngine()
	first := identityRule(t, "first", model.IPBlacklistConfig{IPs: []string{"::ffff:198.51.100.1"}})
	first.Priority, first.Severity = 300, model.SeverityCritical
	second := identityRule(t, "second", model.IPBlacklistConfig{IPs: []string{"198.51.100.1"}})
	second.Priority = 200
	unrelated := identityRule(t, "unrelated", model.IPBlacklistConfig{IPs: []string{"198.51.100.2"}})
	packet := &model.PacketInfo{SrcIP: "198.51.100.1", DstIP: "192.0.2.1", Protocol: 6}
	for _, exit := range []bool{false, true} {
		first.EarlyExit = exit
		if err := e.Reload([]*model.Rule{unrelated, second, first}); err != nil {
			t.Fatal(err)
		}
		alerts := e.Match(packet)
		if exit {
			assertIdentityAlert(t, alerts, first, packet, packet.SrcIP)
		} else {
			if len(alerts) != 2 || alerts[0].RuleID != first.ID || alerts[1].RuleID != second.ID {
				t.Fatalf("per-rule order/count changed: %+v", alerts)
			}
			assertIdentityAlert(t, alerts[:1], first, packet, packet.SrcIP)
			assertIdentityAlert(t, alerts[1:], second, packet, packet.SrcIP)
		}
	}
	wantRules, wantAlerts := e.Rules(), e.Match(packet)
	for _, tc := range []struct {
		ips  []string
		want string
	}{
		{[]string{"not-an-ip"}, `rule bad: invalid IP "not-an-ip"`},
		{[]string{"198.51.100.1/99"}, `rule bad: invalid CIDR "198.51.100.1/99"`},
		{[]string{"", " \t"}, "rule bad: ip blacklist requires at least one IP or CIDR"},
	} {
		bad := identityRule(t, "bad", model.IPBlacklistConfig{IPs: tc.ips})
		before := append(json.RawMessage(nil), bad.Config...)
		err := e.Reload([]*model.Rule{second, bad})
		if err == nil || err.Error() != tc.want {
			t.Fatalf("rejection = %v, want %q", err, tc.want)
		}
		if e.RuleCount() != 3 || !reflect.DeepEqual(e.Rules(), wantRules) ||
			!reflect.DeepEqual(e.Match(packet), wantAlerts) || !bytes.Equal(bad.Config, before) {
			t.Fatal("rejected reload modified inputs or active snapshot")
		}
	}
	if err := e.Reload([]*model.Rule{second}); err != nil {
		t.Fatal(err)
	}
	assertIdentityAlert(t, e.Match(packet), second, packet, packet.SrcIP)
}

func TestExactIPIdentityFileRoundTripPreservesRuleSpelling(t *testing.T) {
	r := identityRule(t, "file", model.IPBlacklistConfig{IPs: []string{" ::FFFF:198.51.100.1 ", "", "198.51.100.1"}, Direction: "source"})
	path := filepath.Join(t.TempDir(), "mapped addresses.json")
	if err := rule.SaveToFile(path, []*model.Rule{r}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := rule.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got, want model.IPBlacklistConfig
	if len(loaded) != 1 {
		t.Fatalf("loaded %d rules", len(loaded))
	}
	if err := json.Unmarshal(loaded[0].Config, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(r.Config, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("file round trip normalized config: %+v", got)
	}
	configBefore := append(json.RawMessage(nil), loaded[0].Config...)
	e := rule.NewEngine()
	if err := e.Reload(loaded); err != nil {
		t.Fatal(err)
	}
	packet := &model.PacketInfo{SrcIP: "198.51.100.1", DstIP: "192.0.2.1", Protocol: 6}
	assertIdentityAlert(t, e.Match(packet), loaded[0], packet, packet.SrcIP)
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) || !bytes.Equal(configBefore, loaded[0].Config) || !reflect.DeepEqual(e.Rules(), loaded) {
		t.Fatal("load/reload/match modified file, caller config or published rule spelling")
	}
}
