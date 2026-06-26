package rule

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/decline-llc/netsentry/pkg/model"
)

type rulesFile struct {
	Rules []*model.Rule `json:"rules"`
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
