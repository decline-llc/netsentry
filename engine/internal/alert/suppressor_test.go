package alert

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestSuppressionManagerAddsAndFilters(t *testing.T) {
	manager, err := NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Add(Suppression{ID: "src", Enabled: true, SrcCIDRs: []string{"10.0.0.0/24"}}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}
	rules := manager.List()
	if len(rules) != 1 || rules[0].ID != "src" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
	filtered := manager.Filter([]*model.Alert{
		alertForSuppression("rule-1", "10.0.0.10", "198.51.100.1"),
		alertForSuppression("rule-2", "198.51.100.2", "198.51.100.1"),
	})
	if len(filtered) != 1 || filtered[0].RuleID != "rule-2" {
		t.Fatalf("unexpected filtered alerts: %+v", filtered)
	}
}

func TestSuppressionManagerRejectsDuplicateID(t *testing.T) {
	manager, err := NewSuppressionManager([]Suppression{{ID: "dup", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Add(Suppression{ID: "dup", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}); err == nil {
		t.Fatal("expected duplicate suppression ID error")
	}
}

func TestSuppressionManagerRejectsEnabledRuleWithoutCIDR(t *testing.T) {
	if _, err := NewSuppressionManager([]Suppression{{ID: "empty", Enabled: true}}); err == nil {
		t.Fatal("expected missing CIDR error")
	}
}

func TestLoadAndSaveSuppressionsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	rules := []Suppression{{ID: "src", Enabled: true, RuleIDs: []string{"rule-1"}, SrcCIDRs: []string{"10.0.0.0/24"}}}
	if err := SaveSuppressionsToFile(path, rules); err != nil {
		t.Fatalf("save suppressions: %v", err)
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "src" || loaded[0].SrcCIDRs[0] != "10.0.0.0/24" {
		t.Fatalf("unexpected suppressions: %+v", loaded)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read suppressions file: %v", err)
	}
	if !strings.Contains(string(raw), `"suppressions"`) {
		t.Fatalf("expected canonical wrapper, got %s", string(raw))
	}
}

func TestRepositorySuppressionsFileUsesCanonicalWrappedSchema(t *testing.T) {
	path := filepath.Join("..", "..", "..", "configs", "suppressions.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var wrapped struct {
		Suppressions []Suppression `json:"suppressions"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		t.Fatalf("decode repository suppressions: %v", err)
	}
	if wrapped.Suppressions == nil {
		t.Fatal("repository suppressions file must use canonical suppressions wrapper")
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load repository suppressions: %v", err)
	}
	if _, err := NewSuppressionManager(loaded); err != nil {
		t.Fatalf("compile repository suppressions: %v", err)
	}
}

func TestLoadSuppressionsFromMissingFileReturnsEmpty(t *testing.T) {
	loaded, err := LoadSuppressionsFromFile(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("load missing suppressions: %v", err)
	}
	if len(loaded) != 0 {
		t.Fatalf("expected empty suppressions, got %+v", loaded)
	}
}

func TestPersistentSuppressionManagerWritesOnAdd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	manager, err := NewSuppressionManagerWithFile(nil, path)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Add(Suppression{ID: "persisted", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "persisted" {
		t.Fatalf("unexpected persisted suppressions: %+v", loaded)
	}
}

func TestSuppressionManagerUpdateDeleteAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	initial := []Suppression{{ID: "s1", Enabled: true, SrcCIDRs: []string{"10.0.0.0/24"}}}
	if err := SaveSuppressionsToFile(path, initial); err != nil {
		t.Fatalf("save initial suppressions: %v", err)
	}
	manager, err := NewSuppressionManagerWithFile(initial, path)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	if err := manager.Update("s1", Suppression{ID: "s1", Enabled: true, DstCIDRs: []string{"192.0.2.0/24"}}); err != nil {
		t.Fatalf("update suppression: %v", err)
	}
	if !manager.suppressor.Suppressed(alertForSuppression("rule-1", "198.51.100.1", "192.0.2.10")) {
		t.Fatal("expected updated destination suppression to be active")
	}

	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load updated suppressions: %v", err)
	}
	if len(loaded) != 1 || len(loaded[0].DstCIDRs) != 1 || loaded[0].DstCIDRs[0] != "192.0.2.0/24" {
		t.Fatalf("unexpected updated file suppressions: %+v", loaded)
	}

	if err := manager.Delete("s1"); err != nil {
		t.Fatalf("delete suppression: %v", err)
	}
	if listed := manager.List(); len(listed) != 0 {
		t.Fatalf("expected delete to clear suppression, got %+v", listed)
	}

	reloadedRules := []Suppression{{ID: "disk", Enabled: true, AnyCIDRs: []string{"203.0.113.0/24"}}}
	if err := SaveSuppressionsToFile(path, reloadedRules); err != nil {
		t.Fatalf("save reload suppressions: %v", err)
	}
	if err := manager.ReloadFromFile(); err != nil {
		t.Fatalf("reload suppressions: %v", err)
	}
	listed := manager.List()
	if len(listed) != 1 || listed[0].ID != "disk" {
		t.Fatalf("unexpected reloaded suppressions: %+v", listed)
	}
}

func TestSuppressionManagerUpdateDeleteMissingID(t *testing.T) {
	manager, err := NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Update("missing", Suppression{Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}); err == nil {
		t.Fatal("expected missing update error")
	}
	if err := manager.Delete("missing"); err == nil {
		t.Fatal("expected missing delete error")
	}
}

func alertForSuppression(ruleID, srcIP, dstIP string) *model.Alert {
	return &model.Alert{RuleID: ruleID, SrcIP: srcIP, DstIP: dstIP}
}
