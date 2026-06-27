package alert

import (
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSuppressorMatchesSourceDestinationAndAnyCIDRs(t *testing.T) {
	suppressor, err := NewSuppressor([]Suppression{
		{ID: "src", Enabled: true, RuleIDs: []string{"rule-src"}, SrcCIDRs: []string{"10.0.0.0/24"}},
		{ID: "dst", Enabled: true, RuleIDs: []string{"rule-dst"}, DstCIDRs: []string{"192.0.2.10"}},
		{ID: "any", Enabled: true, AnyCIDRs: []string{"203.0.113.0/24"}},
	})
	if err != nil {
		t.Fatalf("new suppressor: %v", err)
	}

	cases := []struct {
		name  string
		alert *model.Alert
	}{
		{name: "source", alert: alertForSuppression("rule-src", "10.0.0.55", "198.51.100.1")},
		{name: "destination", alert: alertForSuppression("rule-dst", "198.51.100.1", "192.0.2.10")},
		{name: "any source", alert: alertForSuppression("rule-other", "203.0.113.20", "198.51.100.1")},
		{name: "any destination", alert: alertForSuppression("rule-other", "198.51.100.1", "203.0.113.20")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !suppressor.Suppressed(tc.alert) {
				t.Fatalf("expected alert to be suppressed: %+v", tc.alert)
			}
		})
	}
}

func TestSuppressorHonorsRuleIDAndEnabled(t *testing.T) {
	suppressor, err := NewSuppressor([]Suppression{
		{ID: "disabled", Enabled: false, SrcCIDRs: []string{"10.0.0.0/24"}},
		{ID: "scoped", Enabled: true, RuleIDs: []string{"rule-1"}, SrcCIDRs: []string{"10.0.0.0/24"}},
	})
	if err != nil {
		t.Fatalf("new suppressor: %v", err)
	}
	if suppressor.Suppressed(alertForSuppression("rule-2", "10.0.0.5", "198.51.100.1")) {
		t.Fatal("expected different rule ID to avoid suppression")
	}
	if !suppressor.Suppressed(alertForSuppression("rule-1", "10.0.0.5", "198.51.100.1")) {
		t.Fatal("expected matching rule ID to suppress alert")
	}
}

func TestSuppressorFilter(t *testing.T) {
	suppressor, err := NewSuppressor([]Suppression{{ID: "src", Enabled: true, SrcCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new suppressor: %v", err)
	}
	alerts := []*model.Alert{
		alertForSuppression("rule-1", "10.0.0.5", "198.51.100.1"),
		alertForSuppression("rule-2", "198.51.100.2", "198.51.100.1"),
	}
	filtered := suppressor.Filter(alerts)
	if len(filtered) != 1 || filtered[0].RuleID != "rule-2" {
		t.Fatalf("unexpected filtered alerts: %+v", filtered)
	}
}

func TestSuppressorRejectsInvalidCIDR(t *testing.T) {
	if _, err := NewSuppressor([]Suppression{{ID: "bad", Enabled: true, SrcCIDRs: []string{"not-a-cidr"}}}); err == nil {
		t.Fatal("expected invalid CIDR error")
	}
}

func alertForSuppression(ruleID, srcIP, dstIP string) *model.Alert {
	return &model.Alert{RuleID: ruleID, SrcIP: srcIP, DstIP: dstIP}
}
