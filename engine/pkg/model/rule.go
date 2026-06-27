package model

import "encoding/json"

// RuleType identifies the matching strategy for a rule.
type RuleType string

const (
	RuleTypePayloadMatch       RuleType = "payload_match"
	RuleTypeIPBlacklist        RuleType = "ip_blacklist"
	RuleTypePortBlacklist      RuleType = "port_blacklist"
	RuleTypeFrequencyThreshold RuleType = "frequency_threshold"
)

// MITRETechnique maps a rule to a MITRE ATT&CK technique.
type MITRETechnique struct {
	Tactic        string `json:"tactic"`
	TechniqueID   string `json:"technique_id"`
	TechniqueName string `json:"technique_name"`
}

// Rule is the in-memory representation of a detection rule.
type Rule struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        RuleType         `json:"type"`
	Severity    Severity         `json:"severity"`
	Priority    int              `json:"priority"`
	Enabled     bool             `json:"enabled"`
	EarlyExit   bool             `json:"early_exit"`
	Config      json.RawMessage  `json:"config"`
	MITRETechs  []MITRETechnique `json:"mitre_techniques"`
	Description string           `json:"description"`
}

// PayloadMatchConfig holds config for payload_match rules.
type PayloadMatchConfig struct {
	Keywords        []string `json:"keywords"`
	CaseInsensitive bool     `json:"case_insensitive"`
	Protocols       []string `json:"protocols"`
	Ports           []int    `json:"ports"`
	Direction       string   `json:"direction"`
	Depth           int      `json:"depth"`
	Offset          int      `json:"offset"`
}

// IPBlacklistConfig holds config for ip_blacklist rules.
type IPBlacklistConfig struct {
	IPs       []string `json:"ips"`
	Direction string   `json:"direction"`
	Protocols []string `json:"protocols"`
}

// PortBlacklistConfig holds config for port_blacklist rules.
type PortBlacklistConfig struct {
	Ports     []int    `json:"ports"`
	Protocols []string `json:"protocols"`
	Direction string   `json:"direction"`
}
