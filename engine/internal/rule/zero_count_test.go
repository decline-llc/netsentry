package rule

import (
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestZeroValueEngineCountAcrossReloadLifecycle(t *testing.T) {
	var engine Engine
	packet := &model.PacketInfo{SrcIP: "198.51.100.1", DstIP: "192.0.2.1", DstPort: 80, Protocol: 6}
	assertEmpty := func(t *testing.T, e *Engine) {
		t.Helper()
		if count := e.RuleCount(); count != 0 {
			t.Fatalf("empty rule count = %d, want 0", count)
		}
		if rules := e.Rules(); rules != nil {
			t.Fatalf("empty rules = %+v, want nil", rules)
		}
		if alerts := e.Match(packet); alerts != nil {
			t.Fatalf("empty match = %+v, want nil", alerts)
		}
		if alerts := e.Match(nil); alerts != nil {
			t.Fatalf("nil packet match = %+v, want nil", alerts)
		}
	}
	// This direct call must reach RuleCount before Reload ever publishes a state.
	assertEmpty(t, &engine)
	assertEmpty(t, NewEngine())
	enabled := makePortRuleWithConfig("enabled", model.PortBlacklistConfig{Ports: []int{80}})
	disabled := makePortRuleWithConfig("disabled", model.PortBlacklistConfig{Ports: []int{80}})
	enabled.Priority, disabled.Enabled = 200, false
	input := []*model.Rule{disabled, enabled}
	if err := engine.Reload(input); err != nil {
		t.Fatal(err)
	}
	wantRules := []*model.Rule{enabled, disabled}
	wantAlerts := []*model.Alert{{RuleID: "enabled", RuleName: "enabled", SrcIP: "198.51.100.1", DstIP: "192.0.2.1", DstPort: 80, Protocol: "TCP", Severity: model.SeverityHigh, MatchedKeyword: "port_blacklist: 80"}}
	assertLoaded := func() {
		t.Helper()
		if count := engine.RuleCount(); count != 2 {
			t.Fatalf("loaded count = %d, want both enabled and disabled rules", count)
		}
		if got := engine.Rules(); !reflect.DeepEqual(got, wantRules) {
			t.Fatalf("published rules = %+v, want %+v", got, wantRules)
		}
		if got := engine.Match(packet); !reflect.DeepEqual(got, wantAlerts) {
			t.Fatalf("matched alerts = %+v, want %+v", got, wantAlerts)
		}
	}
	assertLoaded()
	if err := engine.Reload([]*model.Rule{nil}); err == nil || !strings.Contains(err.Error(), "rule at index 0 is null") {
		t.Fatalf("invalid reload error = %v", err)
	}
	assertLoaded()
	if input[0] != disabled || input[1] != enabled {
		t.Fatal("reload changed input order")
	}
	if err := engine.Reload(nil); err != nil {
		t.Fatal(err)
	}
	assertEmpty(t, &engine)
}
