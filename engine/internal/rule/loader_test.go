package rule

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestRepositoryRuleFilesUseCanonicalWrappedSchema(t *testing.T) {
	for _, name := range []string{"rules.json", "rules.example.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "configs", name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			var raw struct {
				Rules []map[string]json.RawMessage `json:"rules"`
			}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("decode %s: %v", name, err)
			}
			if len(raw.Rules) == 0 {
				t.Fatalf("%s should contain wrapped rules", name)
			}
			for i, rule := range raw.Rules {
				if _, ok := rule["config"]; !ok {
					t.Fatalf("%s rule %d missing canonical config object", name, i)
				}
				if _, ok := rule["mitre_techniques"]; !ok {
					t.Fatalf("%s rule %d missing canonical mitre_techniques array", name, i)
				}
				for _, legacyKey := range []string{"payload_match", "ip_blacklist", "port_blacklist", "mitre_tactic", "mitre_technique_id", "mitre_technique_name"} {
					if _, ok := rule[legacyKey]; ok {
						t.Fatalf("%s rule %d still uses legacy key %q", name, i, legacyKey)
					}
				}
			}

			rules, err := LoadFromFile(path)
			if err != nil {
				t.Fatalf("load %s: %v", name, err)
			}
			e := NewEngine()
			if err := e.Reload(rules); err != nil {
				t.Fatalf("compile %s: %v", name, err)
			}
			if e.RuleCount() != len(raw.Rules) {
				t.Fatalf("%s rule count = %d, want %d", name, e.RuleCount(), len(raw.Rules))
			}
		})
	}
}

func TestLoadFromFileSupportsLegacyArrayRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	data := `[
		{
			"id": "legacy-001",
			"name": "Legacy Payload",
			"type": "payload_match",
			"enabled": true,
			"severity": "high",
			"mitre_tactic": "Initial Access",
			"mitre_technique_id": "T1190",
			"mitre_technique_name": "Exploit Public-Facing Application",
			"payload_match": {
				"keywords": ["union select"],
				"case_insensitive": true
			}
		}
	]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Config == nil || string(rules[0].Config) == "{}" {
		t.Fatalf("expected legacy payload_match config to be normalized, got %s", rules[0].Config)
	}
	if len(rules[0].MITRETechs) != 1 || rules[0].MITRETechs[0].TechniqueID != "T1190" {
		t.Fatalf("expected legacy MITRE fields to be normalized, got %#v", rules[0].MITRETechs)
	}

	e := NewEngine()
	if err := e.Reload(rules); err != nil {
		t.Fatal(err)
	}
	alerts := e.Match(&model.PacketInfo{DstPort: 80, Protocol: 6, PayloadPreview: b64("UNION SELECT 1")})
	if len(alerts) != 1 {
		t.Fatalf("expected normalized rule to match, got %d alerts", len(alerts))
	}
}

func TestSaveToFileWritesWrappedSchemaAndPreservesMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"rules":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	rules := []*model.Rule{
		{
			ID:       "saved-001",
			Name:     "Saved Payload",
			Type:     model.RuleTypePayloadMatch,
			Severity: model.SeverityHigh,
			Enabled:  true,
			Config:   json.RawMessage(`{"keywords":["needle"],"case_insensitive":true}`),
		},
	}
	if err := SaveToFile(path, rules); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("mode = %v, want 0600", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var wrapped struct {
		Rules []model.Rule `json:"rules"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		t.Fatalf("decode saved rules: %v\n%s", err, data)
	}
	if len(wrapped.Rules) != 1 || wrapped.Rules[0].ID != "saved-001" {
		t.Fatalf("unexpected saved rules: %+v", wrapped.Rules)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].ID != "saved-001" {
		t.Fatalf("unexpected loaded rules: %+v", loaded)
	}
}
