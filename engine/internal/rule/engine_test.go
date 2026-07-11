package rule

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func makePayloadRule(id, keyword string, caseIns bool) *model.Rule {
	cfg, _ := json.Marshal(model.PayloadMatchConfig{
		Keywords:        []string{keyword},
		CaseInsensitive: caseIns,
		Protocols:       []string{"TCP"},
		Ports:           []int{80},
		Direction:       "dest",
		Depth:           4096,
	})
	return &model.Rule{
		ID:       id,
		Name:     id,
		Type:     model.RuleTypePayloadMatch,
		Severity: model.SeverityHigh,
		Priority: 100,
		Enabled:  true,
		Config:   cfg,
		MITRETechs: []model.MITRETechnique{
			{Tactic: "Initial Access", TechniqueID: "T1190", TechniqueName: "Exploit Public-Facing Application"},
		},
	}
}

func makeIPRule(id, ip string, critical bool) *model.Rule {
	sev := model.SeverityHigh
	if critical {
		sev = model.SeverityCritical
	}
	cfg, _ := json.Marshal(model.IPBlacklistConfig{IPs: []string{ip}, Direction: "any"})
	return &model.Rule{
		ID:        id,
		Name:      id,
		Type:      model.RuleTypeIPBlacklist,
		Severity:  sev,
		Priority:  500,
		Enabled:   true,
		EarlyExit: critical,
		Config:    cfg,
	}
}

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func TestMatchPayload(t *testing.T) {
	e := NewEngine()
	if err := e.Reload([]*model.Rule{makePayloadRule("r1", "UNION SELECT", true)}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{
		SrcIP: "1.2.3.4", DstIP: "5.6.7.8", DstPort: 80, Protocol: 6,
		PayloadPreview: b64("GET /search?q=union select * from users"),
	}
	alerts := e.Match(pkt)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RuleID != "r1" {
		t.Errorf("wrong rule: %s", alerts[0].RuleID)
	}
}

func TestMatchIPBlacklist(t *testing.T) {
	e := NewEngine()
	if err := e.Reload([]*model.Rule{makeIPRule("r2", "10.0.0.99", false)}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{SrcIP: "10.0.0.99", DstIP: "192.168.1.1", DstPort: 80, Protocol: 6}
	alerts := e.Match(pkt)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
}

func TestEarlyExit(t *testing.T) {
	cfg, _ := json.Marshal(model.PayloadMatchConfig{
		Keywords: []string{"SELECT"}, CaseInsensitive: true,
	})
	payloadRule := &model.Rule{
		ID: "r-payload", Name: "payload", Type: model.RuleTypePayloadMatch,
		Severity: model.SeverityHigh, Priority: 100, Enabled: true, Config: cfg,
	}
	ipRule := makeIPRule("r-ip", "10.0.0.1", true) // critical + early_exit
	ipRule.Priority = 1000

	e := NewEngine()
	if err := e.Reload([]*model.Rule{ipRule, payloadRule}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{
		SrcIP: "10.0.0.1", DstIP: "192.168.1.1", DstPort: 80, Protocol: 6,
		PayloadPreview: b64("SELECT * FROM users"),
	}
	alerts := e.Match(pkt)
	// Only the IP rule should fire; payload rule skipped due to early exit.
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert (early exit), got %d", len(alerts))
	}
	if alerts[0].RuleID != "r-ip" {
		t.Errorf("expected r-ip, got %s", alerts[0].RuleID)
	}
}

func TestNoMatchOnDisabledRule(t *testing.T) {
	r := makePayloadRule("r3", "DROP TABLE", false)
	r.Enabled = false
	e := NewEngine()
	_ = e.Reload([]*model.Rule{r})
	pkt := &model.PacketInfo{
		SrcIP: "1.1.1.1", DstIP: "2.2.2.2", DstPort: 80, Protocol: 6,
		PayloadPreview: b64("DROP TABLE users"),
	}
	if alerts := e.Match(pkt); len(alerts) != 0 {
		t.Fatalf("expected 0 alerts for disabled rule, got %d", len(alerts))
	}
}

func TestConcurrentReload(t *testing.T) {
	e := NewEngine()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			_ = e.Reload([]*model.Rule{makePayloadRule("r1", "test", false)})
		}
		close(done)
	}()
	pkt := &model.PacketInfo{
		SrcIP: "1.1.1.1", DstIP: "2.2.2.2", DstPort: 80, Protocol: 6,
		PayloadPreview: b64("test payload"),
	}
	for i := 0; i < 500; i++ {
		_ = e.Match(pkt) // must not panic
	}
	<-done
}

func makePayloadRuleWithConfig(id string, cfg model.PayloadMatchConfig) *model.Rule {
	raw, _ := json.Marshal(cfg)
	return &model.Rule{
		ID:       id,
		Name:     id,
		Type:     model.RuleTypePayloadMatch,
		Severity: model.SeverityHigh,
		Priority: 100,
		Enabled:  true,
		Config:   raw,
	}
}

func TestPayloadRuleProtocolAndDestPort(t *testing.T) {
	e := NewEngine()
	r := makePayloadRuleWithConfig("proto-port", model.PayloadMatchConfig{
		Keywords:        []string{"union select"},
		CaseInsensitive: true,
		Protocols:       []string{"tcp"},
		Ports:           []int{80},
		Direction:       "dest",
	})
	if err := e.Reload([]*model.Rule{r}); err != nil {
		t.Fatal(err)
	}
	matching := &model.PacketInfo{DstPort: 80, Protocol: 6, PayloadPreview: b64("UNION SELECT")}
	if alerts := e.Match(matching); len(alerts) != 1 {
		t.Fatalf("expected tcp dst port match, got %d", len(alerts))
	}
	nonMatchingProtocol := &model.PacketInfo{DstPort: 80, Protocol: 17, PayloadPreview: b64("UNION SELECT")}
	if alerts := e.Match(nonMatchingProtocol); len(alerts) != 0 {
		t.Fatalf("expected udp packet to be rejected, got %d", len(alerts))
	}
	nonMatchingPort := &model.PacketInfo{DstPort: 443, Protocol: 6, PayloadPreview: b64("UNION SELECT")}
	if alerts := e.Match(nonMatchingPort); len(alerts) != 0 {
		t.Fatalf("expected dst port mismatch, got %d", len(alerts))
	}
}

func TestPayloadRuleSourceAndAnyDirectionPorts(t *testing.T) {
	e := NewEngine()
	sourceRule := makePayloadRuleWithConfig("source-port", model.PayloadMatchConfig{
		Keywords:  []string{"token"},
		Ports:     []int{12345},
		Direction: "source",
	})
	anyRule := makePayloadRuleWithConfig("any-port", model.PayloadMatchConfig{
		Keywords:  []string{"token"},
		Ports:     []int{8080},
		Direction: "any",
	})
	if err := e.Reload([]*model.Rule{sourceRule, anyRule}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{SrcPort: 12345, DstPort: 8080, Protocol: 6, PayloadPreview: b64("token")}
	alerts := e.Match(pkt)
	if len(alerts) != 2 {
		t.Fatalf("expected both source and any direction rules, got %d", len(alerts))
	}
}

func TestPayloadRuleOffsetAndDepth(t *testing.T) {
	e := NewEngine()
	r := makePayloadRuleWithConfig("window", model.PayloadMatchConfig{
		Keywords: []string{"needle"},
		Offset:   4,
		Depth:    8,
	})
	if err := e.Reload([]*model.Rule{r}); err != nil {
		t.Fatal(err)
	}
	inside := &model.PacketInfo{Protocol: 6, PayloadPreview: b64("xxxxneedle after")}
	if alerts := e.Match(inside); len(alerts) != 1 {
		t.Fatalf("expected keyword inside offset/depth window, got %d", len(alerts))
	}
	outside := &model.PacketInfo{Protocol: 6, PayloadPreview: b64("xxxx12345678needle")}
	if alerts := e.Match(outside); len(alerts) != 0 {
		t.Fatalf("expected keyword outside depth to be rejected, got %d", len(alerts))
	}
}

func TestPayloadRuleMixedCaseSensitivity(t *testing.T) {
	e := NewEngine()
	caseSensitive := makePayloadRuleWithConfig("case-sensitive", model.PayloadMatchConfig{
		Keywords:        []string{"SELECT"},
		CaseInsensitive: false,
	})
	caseInsensitive := makePayloadRuleWithConfig("case-insensitive", model.PayloadMatchConfig{
		Keywords:        []string{"union"},
		CaseInsensitive: true,
	})
	if err := e.Reload([]*model.Rule{caseSensitive, caseInsensitive}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{Protocol: 6, PayloadPreview: b64("select union")}
	alerts := e.Match(pkt)
	if len(alerts) != 1 || alerts[0].RuleID != "case-insensitive" {
		t.Fatalf("expected only case-insensitive rule, got %+v", alerts)
	}
}

func TestPayloadRuleRejectsInvalidConfig(t *testing.T) {
	e := NewEngine()
	r := makePayloadRuleWithConfig("bad", model.PayloadMatchConfig{
		Keywords:  []string{"x"},
		Protocols: []string{"gre"},
	})
	if err := e.Reload([]*model.Rule{r}); err == nil {
		t.Fatal("expected invalid protocol error")
	}
}

func TestReloadRejectsInvalidRuleSetAndRetainsState(t *testing.T) {
	e := NewEngine()
	valid := makePayloadRule("valid", "needle", false)
	if err := e.Reload([]*model.Rule{valid}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		rules []*model.Rule
	}{
		{name: "nil rule", rules: []*model.Rule{nil}},
		{name: "duplicate id", rules: []*model.Rule{valid, makePayloadRule("valid", "other", false)}},
		{name: "empty keywords", rules: []*model.Rule{makePayloadRuleWithConfig("empty", model.PayloadMatchConfig{})}},
		{name: "unsupported type", rules: []*model.Rule{{ID: "bad", Name: "bad", Type: model.RuleTypeFrequencyThreshold, Severity: model.SeverityHigh, Enabled: true, Config: json.RawMessage(`{}`)}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := e.Reload(test.rules); err == nil {
				t.Fatal("expected validation error")
			}
			if e.RuleCount() != 1 || e.Rules()[0].ID != "valid" {
				t.Fatalf("failed reload replaced active state: %+v", e.Rules())
			}
		})
	}
}

func TestReloadValidatesCanonicalMITREMapping(t *testing.T) {
	e := NewEngine()
	mismatch := makePayloadRule("mismatch", "needle", false)
	mismatch.MITRETechs[0].Tactic = "Execution"
	if err := e.Reload([]*model.Rule{mismatch}); err == nil {
		t.Fatal("expected mismatched MITRE tactic to be rejected")
	}

	multiple := makePayloadRule("multiple", "needle", false)
	multiple.MITRETechs = append(multiple.MITRETechs, model.MITRETechnique{
		Tactic: "Reconnaissance", TechniqueID: "T1595", TechniqueName: "Active Scanning",
	})
	if err := e.Reload([]*model.Rule{multiple}); err == nil {
		t.Fatal("expected multiple MITRE mappings to be rejected by v0.1 schema")
	}
}

func TestMatchNilPacketReturnsNoAlerts(t *testing.T) {
	e := NewEngine()
	if err := e.Reload([]*model.Rule{makePayloadRule("r1", "needle", false)}); err != nil {
		t.Fatal(err)
	}
	if alerts := e.Match(nil); len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %+v", alerts)
	}
}

func makeIPRuleWithConfig(id string, cfg model.IPBlacklistConfig) *model.Rule {
	raw, _ := json.Marshal(cfg)
	return &model.Rule{
		ID:       id,
		Name:     id,
		Type:     model.RuleTypeIPBlacklist,
		Severity: model.SeverityHigh,
		Priority: 100,
		Enabled:  true,
		Config:   raw,
	}
}

func makePortRuleWithConfig(id string, cfg model.PortBlacklistConfig) *model.Rule {
	raw, _ := json.Marshal(cfg)
	return &model.Rule{
		ID:       id,
		Name:     id,
		Type:     model.RuleTypePortBlacklist,
		Severity: model.SeverityHigh,
		Priority: 100,
		Enabled:  true,
		Config:   raw,
	}
}

func TestIPBlacklistDirectionAndProtocol(t *testing.T) {
	e := NewEngine()
	sourceRule := makeIPRuleWithConfig("ip-source", model.IPBlacklistConfig{
		IPs:       []string{"10.0.0.1"},
		Direction: "source",
		Protocols: []string{"tcp"},
	})
	destRule := makeIPRuleWithConfig("ip-dest", model.IPBlacklistConfig{
		IPs:       []string{"10.0.0.2"},
		Direction: "dest",
	})
	if err := e.Reload([]*model.Rule{sourceRule, destRule}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", Protocol: 6}
	alerts := e.Match(pkt)
	if len(alerts) != 2 {
		t.Fatalf("expected source and dest IP matches, got %d", len(alerts))
	}
	udpPkt := &model.PacketInfo{SrcIP: "10.0.0.1", DstIP: "10.0.0.3", Protocol: 17}
	if alerts := e.Match(udpPkt); len(alerts) != 0 {
		t.Fatalf("expected TCP-only source rule to reject UDP, got %d", len(alerts))
	}
}

func TestIPBlacklistCIDRStaysScopedToOwningRule(t *testing.T) {
	e := NewEngine()
	cidrRule := makeIPRuleWithConfig("cidr", model.IPBlacklistConfig{IPs: []string{"198.51.100.0/24"}, Direction: "any"})
	exactRule := makeIPRuleWithConfig("exact", model.IPBlacklistConfig{IPs: []string{"203.0.113.10"}, Direction: "any"})
	if err := e.Reload([]*model.Rule{cidrRule, exactRule}); err != nil {
		t.Fatal(err)
	}
	pkt := &model.PacketInfo{SrcIP: "198.51.100.25", DstIP: "192.168.1.1", Protocol: 6}
	alerts := e.Match(pkt)
	if len(alerts) != 1 || alerts[0].RuleID != "cidr" {
		t.Fatalf("expected only CIDR rule to match, got %+v", alerts)
	}
}

func TestPortBlacklistDirectionAndProtocol(t *testing.T) {
	e := NewEngine()
	sourceRule := makePortRuleWithConfig("port-source", model.PortBlacklistConfig{
		Ports:     []int{4444},
		Protocols: []string{"tcp"},
		Direction: "source",
	})
	anyRule := makePortRuleWithConfig("port-any", model.PortBlacklistConfig{
		Ports:     []int{53},
		Protocols: []string{"udp"},
		Direction: "any",
	})
	if err := e.Reload([]*model.Rule{sourceRule, anyRule}); err != nil {
		t.Fatal(err)
	}
	tcpPkt := &model.PacketInfo{SrcPort: 4444, DstPort: 80, Protocol: 6}
	alerts := e.Match(tcpPkt)
	if len(alerts) != 1 || alerts[0].RuleID != "port-source" {
		t.Fatalf("expected source TCP port rule, got %+v", alerts)
	}
	udpPkt := &model.PacketInfo{SrcPort: 5353, DstPort: 53, Protocol: 17}
	alerts = e.Match(udpPkt)
	if len(alerts) != 1 || alerts[0].RuleID != "port-any" {
		t.Fatalf("expected any UDP port rule, got %+v", alerts)
	}
	wrongProto := &model.PacketInfo{SrcPort: 5353, DstPort: 53, Protocol: 6}
	if alerts := e.Match(wrongProto); len(alerts) != 0 {
		t.Fatalf("expected UDP-only port rule to reject TCP, got %d", len(alerts))
	}
}

func TestIPAndPortRulesRejectInvalidConfig(t *testing.T) {
	e := NewEngine()
	badIP := makeIPRuleWithConfig("bad-ip", model.IPBlacklistConfig{IPs: []string{"not-an-ip"}})
	if err := e.Reload([]*model.Rule{badIP}); err == nil {
		t.Fatal("expected invalid IP error")
	}
	badIPDirection := makeIPRuleWithConfig("bad-ip-direction", model.IPBlacklistConfig{IPs: []string{"10.0.0.1"}, Direction: "sideways"})
	if err := e.Reload([]*model.Rule{badIPDirection}); err == nil {
		t.Fatal("expected invalid IP direction error")
	}
	badPort := makePortRuleWithConfig("bad-port", model.PortBlacklistConfig{Ports: []int{70000}})
	if err := e.Reload([]*model.Rule{badPort}); err == nil {
		t.Fatal("expected invalid port error")
	}
	badProto := makePortRuleWithConfig("bad-proto", model.PortBlacklistConfig{Ports: []int{53}, Protocols: []string{"sctp"}})
	if err := e.Reload([]*model.Rule{badProto}); err == nil {
		t.Fatal("expected invalid protocol error")
	}
}
