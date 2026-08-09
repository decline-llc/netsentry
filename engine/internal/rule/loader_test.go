package rule

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
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

var errInjectedRuleReplacement = errors.New("injected rule replacement failure")

type faultReplacementFile struct {
	replacementFile
	phase string
}

func (f *faultReplacementFile) Write(data []byte) (int, error) {
	switch f.phase {
	case "short-write":
		if len(data) == 0 {
			return 0, nil
		}
		n, err := f.replacementFile.Write(data[:len(data)-1])
		if err != nil {
			return n, err
		}
		return n, nil
	case "write":
		return 0, errInjectedRuleReplacement
	default:
		return f.replacementFile.Write(data)
	}
}

func (f *faultReplacementFile) Chmod(mode os.FileMode) error {
	if f.phase == "chmod" {
		return errInjectedRuleReplacement
	}
	return f.replacementFile.Chmod(mode)
}

func (f *faultReplacementFile) Sync() error {
	if f.phase == "file-sync" {
		return errInjectedRuleReplacement
	}
	return f.replacementFile.Sync()
}

func (f *faultReplacementFile) Close() error {
	err := f.replacementFile.Close()
	if f.phase == "file-close" {
		return errInjectedRuleReplacement
	}
	return err
}

type faultReplacementDirectory struct {
	replacementDirectory
	phase string
}

func (d *faultReplacementDirectory) Sync() error {
	if d.phase == "directory-sync" {
		return errInjectedRuleReplacement
	}
	return d.replacementDirectory.Sync()
}

func (d *faultReplacementDirectory) Close() error {
	err := d.replacementDirectory.Close()
	if d.phase == "directory-close" {
		return errInjectedRuleReplacement
	}
	return err
}

func replacementOpsWithFault(phase string) replacementOps {
	ops := osReplacementOps
	baseStat := ops.stat
	ops.stat = func(path string) (os.FileInfo, error) {
		if phase == "stat" {
			return nil, errInjectedRuleReplacement
		}
		return baseStat(path)
	}
	baseCreateTemp := ops.createTemp
	ops.createTemp = func(dir, pattern string) (replacementFile, error) {
		if phase == "create" {
			return nil, errInjectedRuleReplacement
		}
		file, err := baseCreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return &faultReplacementFile{replacementFile: file, phase: phase}, nil
	}
	baseRename := ops.rename
	ops.rename = func(oldPath, newPath string) error {
		if phase == "rename" {
			return errInjectedRuleReplacement
		}
		return baseRename(oldPath, newPath)
	}
	baseOpenDir := ops.openDir
	ops.openDir = func(path string) (replacementDirectory, error) {
		if phase == "directory-open" {
			return nil, errInjectedRuleReplacement
		}
		dir, err := baseOpenDir(path)
		if err != nil {
			return nil, err
		}
		return &faultReplacementDirectory{replacementDirectory: dir, phase: phase}, nil
	}
	return ops
}

func TestSaveToFileReplacementLifecyclePreservesPriorFileBeforeRename(t *testing.T) {
	for _, phase := range []string{
		"stat",
		"create",
		"short-write",
		"write",
		"chmod",
		"file-sync",
		"file-close",
		"rename",
	} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "rules.json")
			before := []byte("{\"rules\":[]}\n")
			if err := os.WriteFile(path, before, 0o600); err != nil {
				t.Fatal(err)
			}

			rules := []*model.Rule{{ID: "new-rule", Name: "New", Type: model.RuleTypePayloadMatch}}
			data, err := json.MarshalIndent(rulesFile{Rules: rules}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			err = replaceRulesFile(path, append(data, '\n'), replacementOpsWithFault(phase))
			wantErr := errInjectedRuleReplacement
			if phase == "short-write" {
				wantErr = io.ErrShortWrite
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
			if ReplacementCommitted(err) {
				t.Fatalf("phase %s incorrectly reported a committed replacement: %v", phase, err)
			}
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !slices.Equal(after, before) {
				t.Fatalf("phase %s changed prior file: got %q want %q", phase, after, before)
			}
			assertNoRuleTemporaryFiles(t, dir)
		})
	}
}

func TestSaveToFilePostRenameFailuresReportCommittedReplacement(t *testing.T) {
	for _, phase := range []string{"directory-open", "directory-sync", "directory-close"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "rules.json")
			if err := os.WriteFile(path, []byte("{\"rules\":[]}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			rules := []*model.Rule{{ID: "committed-rule", Name: "Committed", Type: model.RuleTypePayloadMatch}}
			data, err := json.MarshalIndent(rulesFile{Rules: rules}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			want := append(data, '\n')

			err = replaceRulesFile(path, want, replacementOpsWithFault(phase))
			if !errors.Is(err, errInjectedRuleReplacement) || !ReplacementCommitted(err) {
				t.Fatalf("error = %v, want committed injected failure", err)
			}
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !slices.Equal(after, want) {
				t.Fatalf("phase %s canonical bytes = %q, want %q", phase, after, want)
			}
			assertNoRuleTemporaryFiles(t, dir)
		})
	}
}

type recordingReplacementFile struct {
	replacementFile
	events *[]string
}

func (f *recordingReplacementFile) Write(data []byte) (int, error) {
	*f.events = append(*f.events, "write")
	return f.replacementFile.Write(data)
}

func (f *recordingReplacementFile) Chmod(mode os.FileMode) error {
	*f.events = append(*f.events, "chmod")
	return f.replacementFile.Chmod(mode)
}

func (f *recordingReplacementFile) Sync() error {
	*f.events = append(*f.events, "file-sync")
	return f.replacementFile.Sync()
}

func (f *recordingReplacementFile) Close() error {
	*f.events = append(*f.events, "file-close")
	return f.replacementFile.Close()
}

type recordingReplacementDirectory struct {
	replacementDirectory
	events *[]string
}

func (d *recordingReplacementDirectory) Sync() error {
	*d.events = append(*d.events, "directory-sync")
	return d.replacementDirectory.Sync()
}

func (d *recordingReplacementDirectory) Close() error {
	*d.events = append(*d.events, "directory-close")
	return d.replacementDirectory.Close()
}

func TestSaveToFileRunsCompleteDurabilityLifecycleInOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte("{\"rules\":[]}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	events := make([]string, 0, 11)
	ops := osReplacementOps
	baseStat := ops.stat
	ops.stat = func(path string) (os.FileInfo, error) {
		events = append(events, "stat")
		return baseStat(path)
	}
	baseCreateTemp := ops.createTemp
	ops.createTemp = func(dir, pattern string) (replacementFile, error) {
		events = append(events, "create")
		file, err := baseCreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return &recordingReplacementFile{replacementFile: file, events: &events}, nil
	}
	baseRename := ops.rename
	ops.rename = func(oldPath, newPath string) error {
		events = append(events, "rename")
		return baseRename(oldPath, newPath)
	}
	baseOpenDir := ops.openDir
	ops.openDir = func(path string) (replacementDirectory, error) {
		events = append(events, "directory-open")
		dir, err := baseOpenDir(path)
		if err != nil {
			return nil, err
		}
		return &recordingReplacementDirectory{replacementDirectory: dir, events: &events}, nil
	}
	baseRemove := ops.remove
	ops.remove = func(path string) error {
		events = append(events, "remove-temp")
		return baseRemove(path)
	}

	rules := []*model.Rule{{ID: "ordered-rule", Name: "Ordered", Type: model.RuleTypePayloadMatch}}
	data, err := json.MarshalIndent(rulesFile{Rules: rules}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := replaceRulesFile(path, append(data, '\n'), ops); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"stat", "create", "write", "chmod", "file-sync", "file-close",
		"rename", "directory-open", "directory-sync", "directory-close", "remove-temp",
	}
	if !slices.Equal(events, want) {
		t.Fatalf("lifecycle events = %v, want %v", events, want)
	}
	assertNoRuleTemporaryFiles(t, dir)
}

func assertNoRuleTemporaryFiles(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".rules-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary rule files remain: %v", matches)
	}
}
