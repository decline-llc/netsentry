package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

type alertFilters struct {
	RuleID           string
	Severity         model.Severity
	SrcIP            string
	DstIP            string
	Protocol         string
	DstPort          *uint16
	Since            *time.Time
	Until            *time.Time
	MitreTactic      string
	MitreTechniqueID string
	MatchedKeyword   string
	MinCount         *int
}

func parseAlertFilters(r *http.Request) (alertFilters, error) {
	q := r.URL.Query()
	filters := alertFilters{
		RuleID:           strings.TrimSpace(q.Get("rule_id")),
		SrcIP:            strings.TrimSpace(q.Get("src_ip")),
		DstIP:            strings.TrimSpace(q.Get("dst_ip")),
		Protocol:         strings.ToUpper(strings.TrimSpace(q.Get("protocol"))),
		MitreTactic:      strings.TrimSpace(q.Get("mitre_tactic")),
		MitreTechniqueID: strings.TrimSpace(q.Get("mitre_technique_id")),
		MatchedKeyword:   strings.TrimSpace(q.Get("matched_keyword")),
	}
	if severity := strings.TrimSpace(q.Get("severity")); severity != "" {
		filters.Severity = model.Severity(strings.ToLower(severity))
		if !validSeverity(filters.Severity) {
			return alertFilters{}, fmt.Errorf("severity must be one of low, medium, high, critical")
		}
	}
	if rawPort := strings.TrimSpace(q.Get("dst_port")); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil || port < 0 || port > 65535 {
			return alertFilters{}, fmt.Errorf("dst_port must be an integer from 0 to 65535")
		}
		p := uint16(port)
		filters.DstPort = &p
	}
	if rawSince := strings.TrimSpace(q.Get("since")); rawSince != "" {
		since, err := parseAlertFilterTime(rawSince)
		if err != nil {
			return alertFilters{}, fmt.Errorf("since must be an RFC3339 timestamp")
		}
		filters.Since = &since
	}
	if rawUntil := strings.TrimSpace(q.Get("until")); rawUntil != "" {
		until, err := parseAlertFilterTime(rawUntil)
		if err != nil {
			return alertFilters{}, fmt.Errorf("until must be an RFC3339 timestamp")
		}
		filters.Until = &until
	}
	if filters.Since != nil && filters.Until != nil && filters.Until.Before(*filters.Since) {
		return alertFilters{}, fmt.Errorf("until must be greater than or equal to since")
	}
	if rawMinCount := strings.TrimSpace(q.Get("min_count")); rawMinCount != "" {
		minCount, err := strconv.Atoi(rawMinCount)
		if err != nil || minCount < 1 {
			return alertFilters{}, fmt.Errorf("min_count must be a positive integer")
		}
		filters.MinCount = &minCount
	}
	return filters, nil
}

func parseAlertFilterTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func validSeverity(severity model.Severity) bool {
	switch severity {
	case model.SeverityLow, model.SeverityMedium, model.SeverityHigh, model.SeverityCritical:
		return true
	default:
		return false
	}
}

func applyAlertFilters(alerts []*model.Alert, filters alertFilters) []*model.Alert {
	out := make([]*model.Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert == nil || !alertMatchesFilters(alert, filters) {
			continue
		}
		out = append(out, alert)
	}
	return out
}

func alertMatchesFilters(alert *model.Alert, filters alertFilters) bool {
	if filters.RuleID != "" && alert.RuleID != filters.RuleID {
		return false
	}
	if filters.Severity != "" && alert.Severity != filters.Severity {
		return false
	}
	if filters.SrcIP != "" && alert.SrcIP != filters.SrcIP {
		return false
	}
	if filters.DstIP != "" && alert.DstIP != filters.DstIP {
		return false
	}
	if filters.Protocol != "" && strings.ToUpper(alert.Protocol) != filters.Protocol {
		return false
	}
	if filters.DstPort != nil && alert.DstPort != *filters.DstPort {
		return false
	}
	if filters.Since != nil && alert.LastSeen.Before(*filters.Since) {
		return false
	}
	if filters.Until != nil && alert.LastSeen.After(*filters.Until) {
		return false
	}
	if filters.MitreTactic != "" && !strings.EqualFold(alert.MitreTactic, filters.MitreTactic) {
		return false
	}
	if filters.MitreTechniqueID != "" && !strings.EqualFold(alert.MitreTechniqueID, filters.MitreTechniqueID) {
		return false
	}
	if filters.MatchedKeyword != "" && !strings.Contains(strings.ToLower(alert.MatchedKeyword), strings.ToLower(filters.MatchedKeyword)) {
		return false
	}
	if filters.MinCount != nil && alert.AggregatedCount < *filters.MinCount {
		return false
	}
	return true
}
