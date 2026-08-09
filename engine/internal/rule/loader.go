package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/decline-llc/netsentry/pkg/model"
)

type rulesFile struct {
	Rules []*model.Rule `json:"rules"`
}

type replacementFile interface {
	Name() string
	Write([]byte) (int, error)
	Chmod(os.FileMode) error
	Sync() error
	Close() error
}

type replacementDirectory interface {
	Sync() error
	Close() error
}

type replacementOps struct {
	stat       func(string) (os.FileInfo, error)
	createTemp func(string, string) (replacementFile, error)
	rename     func(string, string) error
	openDir    func(string) (replacementDirectory, error)
	remove     func(string) error
}

var osReplacementOps = replacementOps{
	stat: os.Stat,
	createTemp: func(dir, pattern string) (replacementFile, error) {
		return os.CreateTemp(dir, pattern)
	},
	rename: os.Rename,
	openDir: func(path string) (replacementDirectory, error) {
		return os.Open(path)
	},
	remove: os.Remove,
}

type replacementError struct {
	phase     string
	committed bool
	err       error
}

func (e *replacementError) Error() string {
	if e.committed {
		return fmt.Sprintf("rules file replacement committed but %s failed: %v", e.phase, e.err)
	}
	return fmt.Sprintf("%s rules file: %v", e.phase, e.err)
}

func (e *replacementError) Unwrap() error {
	return e.err
}

func (e *replacementError) ReplacementCommitted() bool {
	return e.committed
}

// ReplacementCommitted reports whether err describes a replacement that
// already crossed the atomic rename boundary.
func ReplacementCommitted(err error) bool {
	var classified interface {
		ReplacementCommitted() bool
	}
	return errors.As(err, &classified) && classified.ReplacementCommitted()
}

type rawRule struct {
	model.Rule
	MITRETactic        string          `json:"mitre_tactic"`
	MITRETechniqueID   string          `json:"mitre_technique_id"`
	MITRETechniqueName string          `json:"mitre_technique_name"`
	PayloadMatch       json.RawMessage `json:"payload_match"`
	IPBlacklist        json.RawMessage `json:"ip_blacklist"`
	PortBlacklist      json.RawMessage `json:"port_blacklist"`
}

// LoadFromFile reads a rules JSON file and returns the parsed rules.
func LoadFromFile(path string) ([]*model.Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules %s: %w", path, err)
	}
	rules, err := parseRules(data)
	if err != nil {
		return nil, fmt.Errorf("parse rules %s: %w", path, err)
	}
	for _, r := range rules {
		applyRuleDefaults(r)
	}
	return rules, nil
}

// SaveToFile writes rules using the canonical wrapped schema.
func SaveToFile(path string, rules []*model.Rule) error {
	data, err := json.MarshalIndent(rulesFile{Rules: rules}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rules: %w", err)
	}
	data = append(data, '\n')

	return replaceRulesFile(path, data, osReplacementOps)
}

func replaceRulesFile(path string, data []byte, ops replacementOps) (result error) {
	mode := os.FileMode(0o644)
	info, err := ops.stat(path)
	if err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return &replacementError{phase: "stat", err: err}
	}

	dir := filepath.Dir(path)
	tmp, err := ops.createTemp(dir, ".rules-*.json")
	if err != nil {
		return &replacementError{phase: "create temporary", err: err}
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if err := ops.remove(tmpName); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, &replacementError{
				phase:     "remove temporary",
				committed: committed,
				err:       err,
			})
		}
	}()

	written, err := tmp.Write(data)
	if err == nil && written != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return closeReplacementAfterFailure(tmp, &replacementError{phase: "write temporary", err: err})
	}
	if err := tmp.Chmod(mode); err != nil {
		return closeReplacementAfterFailure(tmp, &replacementError{phase: "chmod temporary", err: err})
	}
	if err := tmp.Sync(); err != nil {
		return closeReplacementAfterFailure(tmp, &replacementError{phase: "sync temporary", err: err})
	}
	if err := tmp.Close(); err != nil {
		return &replacementError{phase: "close temporary", err: err}
	}
	if err := ops.rename(tmpName, path); err != nil {
		return &replacementError{phase: "rename temporary", err: err}
	}
	committed = true

	parent, err := ops.openDir(dir)
	if err != nil {
		return &replacementError{phase: "open parent directory", committed: true, err: err}
	}
	if err := parent.Sync(); err != nil {
		return closeDirectoryAfterFailure(parent, &replacementError{phase: "sync parent directory", committed: true, err: err})
	}
	if err := parent.Close(); err != nil {
		return &replacementError{phase: "close parent directory", committed: true, err: err}
	}
	return nil
}

func closeReplacementAfterFailure(file replacementFile, primary error) error {
	if err := file.Close(); err != nil {
		return errors.Join(primary, &replacementError{phase: "close temporary after failure", err: err})
	}
	return primary
}

func closeDirectoryAfterFailure(dir replacementDirectory, primary error) error {
	if err := dir.Close(); err != nil {
		return errors.Join(primary, &replacementError{phase: "close parent directory after failure", committed: true, err: err})
	}
	return primary
}

func parseRules(data []byte) ([]*model.Rule, error) {
	var wrapped struct {
		Rules []rawRule `json:"rules"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Rules != nil {
		return normalizeRules(wrapped.Rules), nil
	}

	var list []rawRule
	if err := json.Unmarshal(data, &list); err != nil {
		var rf rulesFile
		if wrappedErr := json.Unmarshal(data, &rf); wrappedErr == nil {
			return rf.Rules, nil
		}
		return nil, err
	}
	return normalizeRules(list), nil
}

func normalizeRules(raw []rawRule) []*model.Rule {
	rules := make([]*model.Rule, 0, len(raw))
	for i := range raw {
		r := raw[i].Rule
		if len(r.Config) == 0 {
			switch r.Type {
			case model.RuleTypePayloadMatch:
				r.Config = raw[i].PayloadMatch
			case model.RuleTypeIPBlacklist:
				r.Config = raw[i].IPBlacklist
			case model.RuleTypePortBlacklist:
				r.Config = raw[i].PortBlacklist
			}
		}
		if len(r.MITRETechs) == 0 && (raw[i].MITRETactic != "" || raw[i].MITRETechniqueID != "" || raw[i].MITRETechniqueName != "") {
			r.MITRETechs = []model.MITRETechnique{{
				Tactic:        raw[i].MITRETactic,
				TechniqueID:   raw[i].MITRETechniqueID,
				TechniqueName: raw[i].MITRETechniqueName,
			}}
		}
		rules = append(rules, &r)
	}
	return rules
}

func applyRuleDefaults(r *model.Rule) {
	if r.Priority == 0 {
		r.Priority = 100
	}
	if len(r.Config) == 0 {
		r.Config = json.RawMessage("{}")
	}
}
