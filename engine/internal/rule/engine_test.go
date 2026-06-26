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
