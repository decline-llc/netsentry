package rule

import (
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func snapshotRule(kind string) *model.Rule {
	var r *model.Rule
	switch kind {
	case "payload":
		r = makePayloadRule("snapshot", "needle", false)
	case "ip":
		r = makeIPRuleWithConfig("snapshot", model.IPBlacklistConfig{IPs: []string{"198.51.100.1"}})
	case "port":
		r = makePortRuleWithConfig("snapshot", model.PortBlacklistConfig{Ports: []int{80}})
	}
	r.Description = "original description"
	r.MITRETechs = []model.MITRETechnique{
		{Tactic: "Initial Access", TechniqueID: "T1190", TechniqueName: "Exploit Public-Facing Application"},
	}
	return r
}

func snapshotPacket() *model.PacketInfo {
	return &model.PacketInfo{
		SrcIP: "198.51.100.1", DstIP: "192.0.2.1", DstPort: 80, Protocol: 6,
		PayloadPreview: b64("needle"),
	}
}

func TestReloadSnapshotIsolatesCallerMutations(t *testing.T) {
	mutations := []struct {
		name string
		edit func([]*model.Rule)
	}{
		{"slice_entry", func(r []*model.Rule) { r[0] = nil }},
		{"id", func(r []*model.Rule) { r[0].ID = "changed" }},
		{"name", func(r []*model.Rule) { r[0].Name = "changed" }},
		{"type", func(r []*model.Rule) { r[0].Type = model.RuleTypeFrequencyThreshold }},
		{"severity", func(r []*model.Rule) { r[0].Severity = model.SeverityCritical }},
		{"priority", func(r []*model.Rule) { r[0].Priority++ }},
		{"enabled", func(r []*model.Rule) { r[0].Enabled = false }},
		{"early_exit", func(r []*model.Rule) { r[0].EarlyExit = true }},
		{"description", func(r []*model.Rule) { r[0].Description = "changed" }},
		{"config_bytes", func(r []*model.Rule) { r[0].Config[0] = '!' }},
		{"mitre_tactic", func(r []*model.Rule) { r[0].MITRETechs[0].Tactic = "changed" }},
		{"mitre_id", func(r []*model.Rule) { r[0].MITRETechs[0].TechniqueID = "changed" }},
		{"mitre_name", func(r []*model.Rule) { r[0].MITRETechs[0].TechniqueName = "changed" }},
	}
	for _, kind := range []string{"payload", "ip", "port"} {
		for _, mutation := range mutations {
			t.Run(kind+"/"+mutation.name, func(t *testing.T) {
				e := NewEngine()
				input := []*model.Rule{snapshotRule(kind)}
				if err := e.Reload(input); err != nil {
					t.Fatal(err)
				}
				wantRules := e.Rules()
				wantAlerts := e.Match(snapshotPacket())
				if len(wantAlerts) != 1 {
					t.Fatalf("fixture must match exactly one rule, got %+v", wantAlerts)
				}
				mutation.edit(input)
				if got := e.Rules(); !reflect.DeepEqual(got, wantRules) {
					t.Fatalf("caller mutation changed published rules: got %+v, want %+v", got, wantRules)
				}
				if e.RuleCount() != 1 {
					t.Fatalf("caller mutation changed rule count: %d", e.RuleCount())
				}
				if got := e.Match(snapshotPacket()); !reflect.DeepEqual(got, wantAlerts) {
					t.Fatalf("caller mutation changed matching: got %+v, want %+v", got, wantAlerts)
				}
			})
		}
	}
}

func TestReloadSnapshotPreservesInputAndDefensiveOutput(t *testing.T) {
	e := NewEngine()
	low, high := snapshotRule("payload"), snapshotRule("ip")
	low.ID, high.ID = "low", "high"
	low.Priority, high.Priority = 1, 2
	input := []*model.Rule{low, high}
	// Independent fixtures describe the caller data before Reload.
	wantLow, wantHigh := snapshotRule("payload"), snapshotRule("ip")
	wantLow.ID, wantHigh.ID = "low", "high"
	wantLow.Priority, wantHigh.Priority = 1, 2
	if err := e.Reload(input); err != nil {
		t.Fatal(err)
	}
	if input[0] != low || input[1] != high || !reflect.DeepEqual(input, []*model.Rule{wantLow, wantHigh}) {
		t.Fatal("Reload changed caller order or rule data")
	}
	wantRules, wantAlerts := e.Rules(), e.Match(snapshotPacket())
	if len(wantRules) != 2 || wantRules[0].ID != "high" || wantRules[1].ID != "low" {
		t.Fatalf("unexpected priority ordering: %+v", wantRules)
	}
	returned := e.Rules()
	returned[0].Name = "changed"
	returned[0].Config[0] = '!'
	returned[0].MITRETechs[0].Tactic = "changed"
	returned[1] = nil
	if !reflect.DeepEqual(e.Rules(), wantRules) || !reflect.DeepEqual(e.Match(snapshotPacket()), wantAlerts) {
		t.Fatal("mutating Rules output changed the published snapshot or alerts")
	}
}

func TestReloadSnapshotRetainsStateOnRejection(t *testing.T) {
	for _, invalid := range []struct {
		name  string
		rules []*model.Rule
	}{
		{"nil_rule", []*model.Rule{nil}},
		{"duplicate", []*model.Rule{snapshotRule("payload"), snapshotRule("payload")}},
		{"malformed_config", []*model.Rule{{ID: "bad", Name: "bad", Type: model.RuleTypePayloadMatch,
			Severity: model.SeverityHigh, Enabled: true, Config: []byte("{")}}},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			e := NewEngine()
			caller := snapshotRule("payload")
			if err := e.Reload([]*model.Rule{caller}); err != nil {
				t.Fatal(err)
			}
			state, wantRules, wantAlerts := e.state.Load(), e.Rules(), e.Match(snapshotPacket())
			err := e.Reload(invalid.rules)
			if err == nil {
				t.Fatal("invalid reload succeeded")
			}
			if invalid.name == "nil_rule" && err.Error() != "rule at index 0 is null" {
				t.Fatalf("nil diagnostic changed: %v", err)
			}
			caller.Enabled = false
			caller.Config[0] = '!'
			caller.MITRETechs[0].Tactic = "changed"
			if e.state.Load() != state || !reflect.DeepEqual(e.Rules(), wantRules) ||
				!reflect.DeepEqual(e.Match(snapshotPacket()), wantAlerts) {
				t.Fatal("rejected reload or prior caller mutation changed the active snapshot")
			}
		})
	}
}

func TestReloadSnapshotClearsNilAndEmptySets(t *testing.T) {
	for _, input := range [][]*model.Rule{nil, {}} {
		e := NewEngine()
		if err := e.Reload([]*model.Rule{snapshotRule("payload")}); err != nil {
			t.Fatal(err)
		}
		if err := e.Reload(input); err != nil {
			t.Fatal(err)
		}
		if e.RuleCount() != 0 || len(e.Rules()) != 0 || len(e.Match(snapshotPacket())) != 0 {
			t.Fatal("nil/empty reload did not clear the active rules")
		}
	}
}

func TestReloadSnapshotConcurrentPostReturnCallerMutation(t *testing.T) {
	e := NewEngine()
	caller := snapshotRule("payload")
	if err := e.Reload([]*model.Rule{caller}); err != nil {
		t.Fatal(err)
	}
	wantRules, wantAlerts := e.Rules(), e.Match(snapshotPacket())
	start, mutated, readersReady, done := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		<-start // Reload has already returned before caller writes begin.
		caller.Name = "changed"
		close(mutated)
		<-readersReady
		for i := 0; i < 1000; i++ {
			caller.Name = "changed"
			caller.Enabled = i%2 == 0
			caller.Config[0] = byte(i)
			caller.MITRETechs[0].Tactic = "changed"
		}
	}()
	defer func() { <-done }()
	close(start)
	<-mutated // Ensure the regression reaches a real post-return mutation.
	close(readersReady)
	for i := 0; i < 1000; i++ {
		if !reflect.DeepEqual(e.Rules(), wantRules) || !reflect.DeepEqual(e.Match(snapshotPacket()), wantAlerts) {
			t.Fatal("post-return caller writes changed concurrent snapshot reads")
		}
	}
}
