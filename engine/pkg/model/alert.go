package model

import "time"

// Severity levels for alerts and rules.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Alert represents a detected security event.
type Alert struct {
	ID             string    `json:"id"`
	EventID        string    `json:"event_id"`
	RuleID         string    `json:"rule_id"`
	RuleName       string    `json:"rule_name"`
	Timestamp      time.Time `json:"timestamp"`
	SrcIP          string    `json:"src_ip"`
	DstIP          string    `json:"dst_ip"`
	DstPort        uint16    `json:"dst_port"`
	Protocol       string    `json:"protocol"`
	Severity       Severity  `json:"severity"`
	AggregatedCount int      `json:"aggregated_count"`
	FirstSeen      time.Time `json:"first_seen"`
	LastSeen       time.Time `json:"last_seen"`
	WindowStart    time.Time `json:"window_start"`
	MitreTactic         string `json:"mitre_tactic"`
	MitreTechniqueID    string `json:"mitre_technique_id"`
	MitreTechniqueName  string `json:"mitre_technique_name"`
	PayloadPreview      string `json:"payload_preview"`
	MatchedKeyword      string `json:"matched_keyword"`
	RawPayload          string `json:"raw_payload,omitempty"`
}
