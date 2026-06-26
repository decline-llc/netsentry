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
