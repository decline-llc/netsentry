package rule_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/pkg/model"
)

func legacyDecodeFixture(configField string) map[string]any {
	return map[string]any{
		"id": "legacy-decode", "name": "legacy decode fixture", "type": "payload_match", "severity": "high", "enabled": true,
		configField:    map[string]any{"keywords": []string{"needle"}},
		"mitre_tactic": "Initial Access", "mitre_technique_id": "T1190", "mitre_technique_name": "Exploit Public-Facing Application",
		"operator_note": "unknown fields retain existing tolerance",
	}
}

func writeLegacyDecodeFixture(t *testing.T, path string, document any) []byte {
	t.Helper()
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoadRulesRetainsMalformedLegacyWrappedDecodeError(t *testing.T) {
	for _, field := range []string{"mitre_tactic", "mitre_technique_id", "mitre_technique_name"} {
		for i, bad := range []any{7, true, map[string]any{}, []any{}} {
			for _, configField := range []string{"config", "payload_match"} {
				for _, nullPrefix := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%d/%s/null=%v", field, i, configField, nullPrefix), func(t *testing.T) {
						engine := rule.NewEngine()
						seed := &model.Rule{ID: "old", Name: "old snapshot", Type: model.RuleTypePortBlacklist, Severity: model.SeverityHigh, Enabled: true, Priority: 100, Config: json.RawMessage(`{"ports":[80]}`)}
						if err := engine.Reload([]*model.Rule{seed}); err != nil {
							t.Fatal(err)
						}
						packet := &model.PacketInfo{DstPort: 80, Protocol: 6}
						priorRules, priorAlerts := engine.Rules(), engine.Match(packet)
						fixture := legacyDecodeFixture(configField)
						fixture[field] = bad
						items := []any{fixture}
						if nullPrefix {
							items = append([]any{nil}, items...)
						}
						dir := filepath.Join(t.TempDir(), "seed directory with spaces")
						if err := os.Mkdir(dir, 0o700); err != nil {
							t.Fatal(err)
						}
						path := filepath.Join(dir, "rules.json")
						data := writeLegacyDecodeFixture(t, path, map[string]any{"rules": items})
						loaded, err := rule.LoadFromFile(path)
						var decodeErr *json.UnmarshalTypeError
						if loaded != nil || !errors.As(err, &decodeErr) || !strings.HasSuffix(decodeErr.Field, field) || decodeErr.Type.Kind() != reflect.String || decodeErr.Value != []string{"number", "bool", "object", "array"}[i] || !strings.HasPrefix(err.Error(), "parse rules "+path+": ") {
							t.Fatalf("load = %+v, error = %v, want original wrapped field type error", loaded, err)
						}
						if after, readErr := os.ReadFile(path); readErr != nil || string(after) != string(data) {
							t.Fatalf("rejected file changed: %v", readErr)
						}
						if engine.RuleCount() != 1 || !reflect.DeepEqual(engine.Rules(), priorRules) || !reflect.DeepEqual(engine.Match(packet), priorAlerts) {
							t.Fatal("load rejection changed existing engine snapshot")
						}
					})
				}
			}
		}
	}
}

func TestLoadRulesPreservesValidLegacyAndCanonicalFormats(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		for _, configField := range []string{"config", "payload_match"} {
			t.Run(fmt.Sprintf("wrapped=%v/%s", wrapped, configField), func(t *testing.T) {
				fixture := legacyDecodeFixture(configField)
				if configField == "config" {
					for _, field := range []string{"mitre_tactic", "mitre_technique_id", "mitre_technique_name"} {
						delete(fixture, field)
					}
					fixture["mitre_techniques"] = []model.MITRETechnique{{Tactic: "Initial Access", TechniqueID: "T1190", TechniqueName: "Exploit Public-Facing Application"}}
				}
				var document any = []any{fixture}
				if wrapped {
					document = map[string]any{"rules": document, "operator_note": "keep"}
				}
				path := filepath.Join(t.TempDir(), "valid rules.json")
				data := writeLegacyDecodeFixture(t, path, document)
				loaded, err := rule.LoadFromFile(path)
				if err != nil || len(loaded) != 1 {
					t.Fatalf("load = %+v, error %v", loaded, err)
				}
				wantMITRE := []model.MITRETechnique{{Tactic: "Initial Access", TechniqueID: "T1190", TechniqueName: "Exploit Public-Facing Application"}}
				if loaded[0].Priority != 100 || !reflect.DeepEqual(loaded[0].MITRETechs, wantMITRE) {
					t.Fatalf("defaults/legacy tuple changed: %+v", loaded[0])
				}
				engine := rule.NewEngine()
				if err := engine.Reload(loaded); err != nil {
					t.Fatal(err)
				}
				packet := &model.PacketInfo{Protocol: 6, PayloadPreview: base64.StdEncoding.EncodeToString([]byte("needle"))}
				alerts := engine.Match(packet)
				if engine.RuleCount() != 1 || len(alerts) != 1 || alerts[0].RuleID != "legacy-decode" || alerts[0].MitreTechniqueID != "T1190" || alerts[0].MatchedKeyword != "needle" {
					t.Fatalf("normalized config/tuple matching changed: %+v", alerts)
				}
				if after, err := os.ReadFile(path); err != nil || string(after) != string(data) {
					t.Fatalf("valid file changed: %v", err)
				}
			})
		}
	}
}

func TestLoadRulesKeepsExistingEmptyNullAndSyntaxBoundaries(t *testing.T) {
	for _, document := range []string{`{}`, `{"rules":null}`, `null`, `[]`, `{"rules":[]}`, `{"rules":[],"operator_note":1}`} {
		t.Run(document, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "empty rules.json")
			if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
				t.Fatal(err)
			}
			loaded, err := rule.LoadFromFile(path)
			if err != nil || len(loaded) != 0 || ((loaded == nil) != (document == `{}` || document == `{"rules":null}`)) {
				t.Fatalf("existing empty-container parsing changed: %+v, %v", loaded, err)
			}
			if after, err := os.ReadFile(path); err != nil || string(after) != document {
				t.Fatalf("empty file changed: %v", err)
			}
		})
	}
	for _, document := range []string{`{"rules":[null]}`, `[null]`} {
		t.Run(document, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "null rule.json")
			if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
				t.Fatal(err)
			}
			loaded, err := rule.LoadFromFile(path)
			if err != nil || len(loaded) != 1 || loaded[0] == nil || loaded[0].ID != "" || loaded[0].Priority != 100 || string(loaded[0].Config) != "{}" {
				t.Fatalf("lone-null normalization changed: %+v, %v", loaded, err)
			}
			if err := rule.NewEngine().Reload(loaded); err == nil || err.Error() != "rule at index 0: id is required" {
				t.Fatalf("downstream semantic diagnostic changed: %v", err)
			}
			if after, err := os.ReadFile(path); err != nil || string(after) != document {
				t.Fatalf("null-entry file changed: %v", err)
			}
		})
	}
	for _, document := range []string{`{"rules":[`, `[`, `{"rules":[{"id":"bad",}]}`} {
		t.Run(document, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "malformed rules.json")
			if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
				t.Fatal(err)
			}
			loaded, err := rule.LoadFromFile(path)
			var syntaxErr *json.SyntaxError
			if loaded != nil || !errors.As(err, &syntaxErr) {
				t.Fatalf("malformed syntax no longer rejected: %+v, %v", loaded, err)
			}
			if after, err := os.ReadFile(path); err != nil || string(after) != document {
				t.Fatalf("malformed file changed: %v", err)
			}
		})
	}
}
