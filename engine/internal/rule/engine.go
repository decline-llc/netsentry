package rule

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync/atomic"

	ac "github.com/decline-llc/netsentry/internal/rule/ahocorasick"
	"github.com/decline-llc/netsentry/pkg/model"
)

// ruleState is an immutable snapshot of the active rule set.
type ruleState struct {
	acMatcher       *ac.Matcher
	acRuleIdx       []int // parallel to ac patterns: which rule owns each keyword
	payloadRules    map[string]compiledPayloadRule
	ipNets          []*net.IPNet
	ipRules         map[string]*model.Rule
	portRules       map[uint16][]*model.Rule
	allByPriority   []*model.Rule
	caseInsensitive bool
}

type compiledPayloadRule struct {
	keywords        []string
	caseInsensitive bool
	protocols       map[uint8]struct{}
	ports           map[uint16]struct{}
	direction       string
	depth           int
	offset          int
}

// Engine is the lock-free rule matching engine.
type Engine struct {
	state atomic.Pointer[ruleState]
}

// NewEngine creates an Engine with an empty rule set.
func NewEngine() *Engine {
	e := &Engine{}
	e.state.Store(&ruleState{
		payloadRules: make(map[string]compiledPayloadRule),
		ipRules:      make(map[string]*model.Rule),
		portRules:    make(map[uint16][]*model.Rule),
	})
	return e
}

// Reload atomically replaces the rule set.
func (e *Engine) Reload(rules []*model.Rule) error {
	s, err := buildState(rules)
	if err != nil {
		return err
	}
	e.state.Store(s)
	return nil
}

// RuleCount returns the number of rules currently loaded.
func (e *Engine) RuleCount() int {
	return len(e.state.Load().allByPriority)
}

// Match runs all enabled rules against pkt and returns triggered alerts.
func (e *Engine) Match(pkt *model.PacketInfo) []*model.Alert {
	s := e.state.Load()
	if s == nil || len(s.allByPriority) == 0 {
		return nil
	}

	payload := decodePayload(pkt.PayloadPreview)

	// Run AC automaton once; collect which rule indices matched.
	acHitRules := map[string]string{} // rule ID → matched keyword
	if s.acMatcher != nil && len(payload) > 0 {
		for _, patIdx := range s.acMatcher.Match(payload) {
			if patIdx < len(s.acRuleIdx) {
				ruleIdx := s.acRuleIdx[patIdx]
				if ruleIdx < len(s.allByPriority) {
					r := s.allByPriority[ruleIdx]
					if _, seen := acHitRules[r.ID]; !seen {
						acHitRules[r.ID] = s.acMatcher.Patterns()[patIdx]
					}
				}
			}
		}
	}

	var alerts []*model.Alert
	for _, rule := range s.allByPriority {
		if !rule.Enabled {
			continue
		}
		var alert *model.Alert
		switch rule.Type {
		case model.RuleTypeIPBlacklist:
			alert = matchIPBlacklist(rule, pkt, s)
		case model.RuleTypePortBlacklist:
			alert = matchPortBlacklist(rule, pkt, s)
		case model.RuleTypePayloadMatch:
			if _, ok := acHitRules[rule.ID]; ok {
				if cfg, exists := s.payloadRules[rule.ID]; exists {
					if kw, matched := matchPayloadRule(cfg, pkt, payload); matched {
						preview := previewPayload(payload, 200)
						alert = buildAlert(rule, pkt, preview, kw)
					}
				}
			}
		}
		if alert != nil {
			alerts = append(alerts, alert)
			if rule.EarlyExit && rule.Severity == model.SeverityCritical {
				break
			}
		}
	}
	return alerts
}

// ---- state builder ----------------------------------------------------------

func buildState(rules []*model.Rule) (*ruleState, error) {
	s := &ruleState{
		payloadRules: make(map[string]compiledPayloadRule),
		ipRules:      make(map[string]*model.Rule),
		portRules:    make(map[uint16][]*model.Rule),
	}

	sorted := make([]*model.Rule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})
	s.allByPriority = sorted

	// Build a priority-indexed lookup for AC rule tracking.
	ruleByID := make(map[string]int, len(sorted))
	for i, r := range sorted {
		ruleByID[r.ID] = i
	}

	var keywords []string
	var keywordRuleIdx []int
	caseIns := false

	for _, r := range sorted {
		if !r.Enabled {
			continue
		}
		switch r.Type {
		case model.RuleTypePayloadMatch:
			var cfg model.PayloadMatchConfig
			if err := json.Unmarshal(r.Config, &cfg); err != nil {
				return nil, fmt.Errorf("rule %s: %w", r.ID, err)
			}
			compiled, err := compilePayloadRule(cfg)
			if err != nil {
				return nil, fmt.Errorf("rule %s: %w", r.ID, err)
			}
			s.payloadRules[r.ID] = compiled
			if cfg.CaseInsensitive {
				caseIns = true
			}
			ruleIdx := ruleByID[r.ID]
			for _, kw := range cfg.Keywords {
				if cfg.CaseInsensitive {
					kw = strings.ToLower(kw)
				}
				keywords = append(keywords, kw)
				keywordRuleIdx = append(keywordRuleIdx, ruleIdx)
			}

		case model.RuleTypeIPBlacklist:
			var cfg model.IPBlacklistConfig
			if err := json.Unmarshal(r.Config, &cfg); err != nil {
				return nil, fmt.Errorf("rule %s: %w", r.ID, err)
			}
			for _, ipStr := range cfg.IPs {
				if strings.Contains(ipStr, "/") {
					_, ipNet, err := net.ParseCIDR(ipStr)
					if err == nil {
						s.ipNets = append(s.ipNets, ipNet)
					}
				} else {
					s.ipRules[ipStr] = r
				}
			}

		case model.RuleTypePortBlacklist:
			var cfg model.PortBlacklistConfig
			if err := json.Unmarshal(r.Config, &cfg); err != nil {
				return nil, fmt.Errorf("rule %s: %w", r.ID, err)
			}
			for _, port := range cfg.Ports {
				p := uint16(port)
				s.portRules[p] = append(s.portRules[p], r)
			}
		}
	}

	if len(keywords) > 0 {
		s.acMatcher = ac.NewMatcher(keywords, caseIns)
		s.acRuleIdx = keywordRuleIdx
	}
	s.caseInsensitive = caseIns
	return s, nil
}

func compilePayloadRule(cfg model.PayloadMatchConfig) (compiledPayloadRule, error) {
	compiled := compiledPayloadRule{
		keywords:        append([]string(nil), cfg.Keywords...),
		caseInsensitive: cfg.CaseInsensitive,
		protocols:       make(map[uint8]struct{}),
		ports:           make(map[uint16]struct{}),
		direction:       strings.ToLower(cfg.Direction),
		depth:           cfg.Depth,
		offset:          cfg.Offset,
	}
	if compiled.direction == "" {
		compiled.direction = "dest"
	}
	if compiled.direction != "dest" && compiled.direction != "source" && compiled.direction != "any" {
		return compiledPayloadRule{}, fmt.Errorf("unsupported payload direction %q", cfg.Direction)
	}
	if compiled.depth < 0 || compiled.offset < 0 {
		return compiledPayloadRule{}, fmt.Errorf("payload depth and offset must be non-negative")
	}
	for _, proto := range cfg.Protocols {
		p, ok := parseProtocol(proto)
		if !ok {
			return compiledPayloadRule{}, fmt.Errorf("unsupported protocol %q", proto)
		}
		if p == 0 {
			continue
		}
		compiled.protocols[p] = struct{}{}
	}
	for _, port := range cfg.Ports {
		if port < 0 || port > 65535 {
			return compiledPayloadRule{}, fmt.Errorf("port %d out of range", port)
		}
		compiled.ports[uint16(port)] = struct{}{}
	}
	return compiled, nil
}

func parseProtocol(proto string) (uint8, bool) {
	switch strings.ToLower(strings.TrimSpace(proto)) {
	case "", "any":
		return 0, true
	case "tcp":
		return 6, true
	case "udp":
		return 17, true
	case "icmp":
		return 1, true
	default:
		return 0, false
	}
}

func matchPayloadRule(cfg compiledPayloadRule, pkt *model.PacketInfo, payload []byte) (string, bool) {
	if len(cfg.protocols) > 0 {
		if _, ok := cfg.protocols[pkt.Protocol]; !ok {
			return "", false
		}
	}
	if len(cfg.ports) > 0 && !payloadPortMatches(cfg, pkt) {
		return "", false
	}

	if cfg.offset > len(payload) {
		return "", false
	}
	window := payload[cfg.offset:]
	if cfg.depth > 0 && cfg.depth < len(window) {
		window = window[:cfg.depth]
	}

	haystack := string(window)
	if cfg.caseInsensitive {
		haystack = strings.ToLower(haystack)
	}
	for _, kw := range cfg.keywords {
		needle := kw
		if cfg.caseInsensitive {
			needle = strings.ToLower(needle)
		}
		if strings.Contains(haystack, needle) {
			return kw, true
		}
	}
	return "", false
}

func payloadPortMatches(cfg compiledPayloadRule, pkt *model.PacketInfo) bool {
	_, srcOK := cfg.ports[pkt.SrcPort]
	_, dstOK := cfg.ports[pkt.DstPort]
	switch cfg.direction {
	case "source":
		return srcOK
	case "any":
		return srcOK || dstOK
	default:
		return dstOK
	}
}

// ---- per-type matchers ------------------------------------------------------

func matchIPBlacklist(rule *model.Rule, pkt *model.PacketInfo, s *ruleState) *model.Alert {
	src := net.ParseIP(pkt.SrcIP)
	dst := net.ParseIP(pkt.DstIP)

	if _, ok := s.ipRules[pkt.SrcIP]; ok {
		return buildAlert(rule, pkt, "", "ip_blacklist: "+pkt.SrcIP)
	}
	if _, ok := s.ipRules[pkt.DstIP]; ok {
		return buildAlert(rule, pkt, "", "ip_blacklist: "+pkt.DstIP)
	}
	for _, ipNet := range s.ipNets {
		if (src != nil && ipNet.Contains(src)) || (dst != nil && ipNet.Contains(dst)) {
			return buildAlert(rule, pkt, "", "ip_blacklist: "+ipNet.String())
		}
	}
	return nil
}

func matchPortBlacklist(rule *model.Rule, pkt *model.PacketInfo, s *ruleState) *model.Alert {
	if rules, ok := s.portRules[pkt.DstPort]; ok {
		for _, r := range rules {
			if r.ID == rule.ID {
				return buildAlert(rule, pkt, "", fmt.Sprintf("port_blacklist: %d", pkt.DstPort))
			}
		}
	}
	return nil
}

// ---- helpers ----------------------------------------------------------------

func decodePayload(preview string) []byte {
	if preview == "" {
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(preview)
	if err != nil {
		return []byte(preview)
	}
	return b
}

func previewPayload(b []byte, maxLen int) string {
	if len(b) > maxLen {
		return string(b[:maxLen])
	}
	return string(b)
}

func buildAlert(rule *model.Rule, pkt *model.PacketInfo, payloadPreview, matchedKeyword string) *model.Alert {
	mitreTactic, mitreTechID, mitreTechName := "", "", ""
	if len(rule.MITRETechs) > 0 {
		mitreTactic = rule.MITRETechs[0].Tactic
		mitreTechID = rule.MITRETechs[0].TechniqueID
		mitreTechName = rule.MITRETechs[0].TechniqueName
	}
	return &model.Alert{
		RuleID:             rule.ID,
		RuleName:           rule.Name,
		SrcIP:              pkt.SrcIP,
		DstIP:              pkt.DstIP,
		DstPort:            pkt.DstPort,
		Protocol:           protocolName(pkt.Protocol),
		Severity:           rule.Severity,
		MitreTactic:        mitreTactic,
		MitreTechniqueID:   mitreTechID,
		MitreTechniqueName: mitreTechName,
		PayloadPreview:     payloadPreview,
		MatchedKeyword:     matchedKeyword,
	}
}

func protocolName(p uint8) string {
	switch p {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 1:
		return "ICMP"
	default:
		return fmt.Sprintf("PROTO_%d", p)
	}
}
