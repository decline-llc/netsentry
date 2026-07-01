package alert

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/decline-llc/netsentry/pkg/model"
)

// Suppression describes an alert suppression rule scoped by rule ID and IP range.
type Suppression struct {
	ID       string   `json:"id"`
	Enabled  bool     `json:"enabled"`
	RuleIDs  []string `json:"rule_ids"`
	SrcCIDRs []string `json:"src_cidrs"`
	DstCIDRs []string `json:"dst_cidrs"`
	AnyCIDRs []string `json:"any_cidrs"`
}

// SuppressionManager owns the active in-memory suppression rules and compiled filter.
type SuppressionManager struct {
	mu         sync.RWMutex
	rules      []Suppression
	suppressor *Suppressor
	filePath   string
}

type suppressionsFile struct {
	Suppressions []Suppression `json:"suppressions"`
}

// NewSuppressionManager constructs an in-memory suppression manager.
func NewSuppressionManager(rules []Suppression) (*SuppressionManager, error) {
	return NewSuppressionManagerWithFile(rules, "")
}

// NewSuppressionManagerWithFile constructs a suppression manager that persists
// successful changes to path when path is not empty.
func NewSuppressionManagerWithFile(rules []Suppression, path string) (*SuppressionManager, error) {
	if err := validateSuppressionSet(rules); err != nil {
		return nil, err
	}
	suppressor, err := NewSuppressor(rules)
	if err != nil {
		return nil, err
	}
	return &SuppressionManager{rules: cloneSuppressions(rules), suppressor: suppressor, filePath: path}, nil
}

// LoadSuppressionsFromFile reads suppression rules from a canonical JSON file.
// A missing file is treated as an empty suppression set.
func LoadSuppressionsFromFile(path string) ([]Suppression, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read suppressions %s: %w", path, err)
	}
	var wrapped suppressionsFile
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("parse suppressions %s: %w", path, err)
	}
	if err := validateSuppressionSet(wrapped.Suppressions); err != nil {
		return nil, fmt.Errorf("validate suppressions %s: %w", path, err)
	}
	return cloneSuppressions(wrapped.Suppressions), nil
}

// SaveSuppressionsToFile writes suppression rules using the canonical wrapped schema.
func SaveSuppressionsToFile(path string, rules []Suppression) error {
	if path == "" {
		return nil
	}
	if err := validateSuppressionSet(rules); err != nil {
		return err
	}
	data, err := json.MarshalIndent(suppressionsFile{Suppressions: cloneSuppressions(rules)}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal suppressions: %w", err)
	}
	data = append(data, '\n')

	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create suppressions dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".suppressions-*.json")
	if err != nil {
		return fmt.Errorf("create temp suppressions file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp suppressions file: %w", err)
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp suppressions file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp suppressions file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace suppressions file: %w", err)
	}
	return nil
}

// List returns the configured suppression rules in insertion order.
func (m *SuppressionManager) List() []Suppression {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneSuppressions(m.rules)
}

// Add validates and appends one suppression rule, then atomically swaps the compiled filter.
func (m *SuppressionManager) Add(rule Suppression) error {
	if m == nil {
		return fmt.Errorf("suppression manager is not configured")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	candidate := append(cloneSuppressions(m.rules), cloneSuppression(rule))
	if err := validateSuppressionSet(candidate); err != nil {
		return err
	}
	suppressor, err := NewSuppressor(candidate)
	if err != nil {
		return err
	}
	if m.filePath != "" {
		if err := SaveSuppressionsToFile(m.filePath, candidate); err != nil {
			return fmt.Errorf("persist suppressions: %w", err)
		}
	}
	m.rules = candidate
	m.suppressor = suppressor
	return nil
}

// Filter returns only alerts not matching active suppressions.
func (m *SuppressionManager) Filter(alerts []*model.Alert) []*model.Alert {
	if m == nil || len(alerts) == 0 {
		return alerts
	}
	m.mu.RLock()
	suppressor := m.suppressor
	m.mu.RUnlock()
	if suppressor == nil {
		return alerts
	}
	return suppressor.Filter(alerts)
}

func validateSuppressionSet(rules []Suppression) error {
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		if strings.TrimSpace(rule.ID) == "" {
			return fmt.Errorf("id is required")
		}
		if _, ok := seen[rule.ID]; ok {
			return fmt.Errorf("suppression %q already exists", rule.ID)
		}
		seen[rule.ID] = struct{}{}
		if rule.Enabled && len(rule.SrcCIDRs) == 0 && len(rule.DstCIDRs) == 0 && len(rule.AnyCIDRs) == 0 {
			return fmt.Errorf("suppression %q must include at least one CIDR", rule.ID)
		}
	}
	return nil
}

func cloneSuppressions(rules []Suppression) []Suppression {
	out := make([]Suppression, 0, len(rules))
	for _, rule := range rules {
		out = append(out, cloneSuppression(rule))
	}
	return out
}

func cloneSuppression(rule Suppression) Suppression {
	return Suppression{
		ID:       rule.ID,
		Enabled:  rule.Enabled,
		RuleIDs:  append([]string(nil), rule.RuleIDs...),
		SrcCIDRs: append([]string(nil), rule.SrcCIDRs...),
		DstCIDRs: append([]string(nil), rule.DstCIDRs...),
		AnyCIDRs: append([]string(nil), rule.AnyCIDRs...),
	}
}

// Suppressor filters alerts using precompiled CIDR and exact-IP suppression rules.
type Suppressor struct {
	rules []compiledSuppression
}

type compiledSuppression struct {
	id      string
	ruleIDs map[string]struct{}
	src     []netip.Prefix
	dst     []netip.Prefix
	any     []netip.Prefix
}

// NewSuppressor compiles suppression rules. Empty rule IDs match every rule.
func NewSuppressor(rules []Suppression) (*Suppressor, error) {
	compiled := make([]compiledSuppression, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		c := compiledSuppression{
			id:      rule.ID,
			ruleIDs: make(map[string]struct{}, len(rule.RuleIDs)),
		}
		for _, ruleID := range rule.RuleIDs {
			if ruleID != "" {
				c.ruleIDs[ruleID] = struct{}{}
			}
		}
		var err error
		if c.src, err = compilePrefixes(rule.SrcCIDRs); err != nil {
			return nil, fmt.Errorf("suppression %s src cidrs: %w", rule.ID, err)
		}
		if c.dst, err = compilePrefixes(rule.DstCIDRs); err != nil {
			return nil, fmt.Errorf("suppression %s dst cidrs: %w", rule.ID, err)
		}
		if c.any, err = compilePrefixes(rule.AnyCIDRs); err != nil {
			return nil, fmt.Errorf("suppression %s any cidrs: %w", rule.ID, err)
		}
		compiled = append(compiled, c)
	}
	return &Suppressor{rules: compiled}, nil
}

func compilePrefixes(values []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		prefix, err := parsePrefix(value)
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func parsePrefix(value string) (netip.Prefix, error) {
	if prefix, err := netip.ParsePrefix(value); err == nil {
		return prefix.Masked(), nil
	}
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("invalid IP or CIDR %q", value)
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

// Suppressed reports whether alert matches any suppression rule.
func (s *Suppressor) Suppressed(alert *model.Alert) bool {
	if s == nil || alert == nil {
		return false
	}
	src, srcErr := netip.ParseAddr(alert.SrcIP)
	dst, dstErr := netip.ParseAddr(alert.DstIP)
	srcOK := srcErr == nil
	dstOK := dstErr == nil
	for _, rule := range s.rules {
		if !rule.matchesRuleID(alert.RuleID) {
			continue
		}
		if srcOK && containsPrefix(rule.src, src) {
			return true
		}
		if dstOK && containsPrefix(rule.dst, dst) {
			return true
		}
		if (srcOK && containsPrefix(rule.any, src)) || (dstOK && containsPrefix(rule.any, dst)) {
			return true
		}
	}
	return false
}

func (r compiledSuppression) matchesRuleID(ruleID string) bool {
	if len(r.ruleIDs) == 0 {
		return true
	}
	_, ok := r.ruleIDs[ruleID]
	return ok
}

func containsPrefix(prefixes []netip.Prefix, addr netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// Filter returns a new slice containing only non-suppressed alerts.
func (s *Suppressor) Filter(alerts []*model.Alert) []*model.Alert {
	if s == nil || len(alerts) == 0 {
		return alerts
	}
	out := make([]*model.Alert, 0, len(alerts))
	for _, alert := range alerts {
		if !s.Suppressed(alert) {
			out = append(out, alert)
		}
	}
	return out
}
