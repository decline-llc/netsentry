package rule

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

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
