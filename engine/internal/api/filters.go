package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/decline-llc/netsentry/pkg/model"
)

type alertFilters struct {
	RuleID   string
	Severity model.Severity
	SrcIP    string
	DstIP    string
	Protocol string
	DstPort  *uint16
}

func parseAlertFilters(r *http.Request) (alertFilters, error) {
	q := r.URL.Query()
	filters := alertFilters{
		RuleID:   strings.TrimSpace(q.Get("rule_id")),
		SrcIP:    strings.TrimSpace(q.Get("src_ip")),
		DstIP:    strings.TrimSpace(q.Get("dst_ip")),
		Protocol: strings.ToUpper(strings.TrimSpace(q.Get("protocol"))),
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
	return filters, nil
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
	return true
}
