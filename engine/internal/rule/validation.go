package rule

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/decline-llc/netsentry/pkg/model"
)

// canonicalMITRE contains the Enterprise ATT&CK techniques currently used by
// NetSentry's v0.1 rule schema. Keeping the tuple together prevents a valid ID
// from being published with a mismatched tactic or technique name.
var canonicalMITRE = map[string]model.MITRETechnique{
	"T1083":     {Tactic: "Discovery", TechniqueID: "T1083", TechniqueName: "File and Directory Discovery"},
	"T1059":     {Tactic: "Execution", TechniqueID: "T1059", TechniqueName: "Command and Scripting Interpreter"},
	"T1059.004": {Tactic: "Execution", TechniqueID: "T1059.004", TechniqueName: "Command and Scripting Interpreter: Unix Shell"},
	"T1071":     {Tactic: "Command and Control", TechniqueID: "T1071", TechniqueName: "Application Layer Protocol"},
	"T1190":     {Tactic: "Initial Access", TechniqueID: "T1190", TechniqueName: "Exploit Public-Facing Application"},
	"T1595":     {Tactic: "Reconnaissance", TechniqueID: "T1595", TechniqueName: "Active Scanning"},
}

func validateRuleSet(rules []*model.Rule) error {
	seen := make(map[string]struct{}, len(rules))
	for index, rule := range rules {
		if rule == nil {
			return fmt.Errorf("rule at index %d is null", index)
		}
		if strings.TrimSpace(rule.ID) == "" {
			return fmt.Errorf("rule at index %d: id is required", index)
		}
		if _, exists := seen[rule.ID]; exists {
			return fmt.Errorf("duplicate rule id %q", rule.ID)
		}
		seen[rule.ID] = struct{}{}
		if strings.TrimSpace(rule.Name) == "" {
			return fmt.Errorf("rule %s: name is required", rule.ID)
		}
		switch rule.Severity {
		case model.SeverityLow, model.SeverityMedium, model.SeverityHigh, model.SeverityCritical:
		default:
			return fmt.Errorf("rule %s: unsupported severity %q", rule.ID, rule.Severity)
		}
		if err := validateRuleMITRE(rule); err != nil {
			return err
		}
		if err := validateRuleConfig(rule); err != nil {
			return err
		}
	}
	return nil
}

func validateRuleMITRE(rule *model.Rule) error {
	if len(rule.MITRETechs) > 1 {
		return fmt.Errorf("rule %s: v0.1 alert schema supports at most one MITRE technique", rule.ID)
	}
	if len(rule.MITRETechs) == 0 {
		return nil
	}
	got := rule.MITRETechs[0]
	want, ok := canonicalMITRE[got.TechniqueID]
	if !ok {
		return fmt.Errorf("rule %s: unsupported MITRE technique id %q; update the versioned catalog first", rule.ID, got.TechniqueID)
	}
	if got.Tactic != want.Tactic || got.TechniqueName != want.TechniqueName {
		return fmt.Errorf(
			"rule %s: MITRE %s must use tactic %q and name %q",
			rule.ID, got.TechniqueID, want.Tactic, want.TechniqueName,
		)
	}
	return nil
}

func validateRuleConfig(rule *model.Rule) error {
	switch rule.Type {
	case model.RuleTypePayloadMatch:
		var cfg model.PayloadMatchConfig
		if err := json.Unmarshal(rule.Config, &cfg); err != nil {
			return fmt.Errorf("rule %s: decode payload config: %w", rule.ID, err)
		}
		if len(cfg.Keywords) == 0 {
			return fmt.Errorf("rule %s: payload rule requires at least one keyword", rule.ID)
		}
		for _, keyword := range cfg.Keywords {
			if strings.TrimSpace(keyword) == "" {
				return fmt.Errorf("rule %s: payload keywords must not be empty", rule.ID)
			}
		}
		if _, err := compilePayloadRule(cfg); err != nil {
			return fmt.Errorf("rule %s: %w", rule.ID, err)
		}
	case model.RuleTypeIPBlacklist:
		var cfg model.IPBlacklistConfig
		if err := json.Unmarshal(rule.Config, &cfg); err != nil {
			return fmt.Errorf("rule %s: decode ip config: %w", rule.ID, err)
		}
		if len(cfg.IPs) == 0 {
			return fmt.Errorf("rule %s: ip blacklist requires at least one IP or CIDR", rule.ID)
		}
		if _, err := compileIPRule(cfg); err != nil {
			return fmt.Errorf("rule %s: %w", rule.ID, err)
		}
	case model.RuleTypePortBlacklist:
		var cfg model.PortBlacklistConfig
		if err := json.Unmarshal(rule.Config, &cfg); err != nil {
			return fmt.Errorf("rule %s: decode port config: %w", rule.ID, err)
		}
		if len(cfg.Ports) == 0 {
			return fmt.Errorf("rule %s: port blacklist requires at least one port", rule.ID)
		}
		if _, err := compilePortRule(cfg); err != nil {
			return fmt.Errorf("rule %s: %w", rule.ID, err)
		}
	default:
		return fmt.Errorf("rule %s: unsupported rule type %q", rule.ID, rule.Type)
	}
	return nil
}
