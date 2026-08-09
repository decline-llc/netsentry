package alert

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	mu             sync.RWMutex
	rules          []Suppression
	suppressor     *Suppressor
	filePath       string
	replacementOps suppressionReplacementOps
}

type suppressionsFile struct {
	Suppressions []Suppression `json:"suppressions"`
}

type suppressionReplacementFile interface {
	Name() string
	Write([]byte) (int, error)
	Chmod(os.FileMode) error
	Sync() error
	Close() error
}

type suppressionReplacementDirectory interface {
	Sync() error
	Close() error
}

type suppressionReplacementOps struct {
	stat       func(string) (os.FileInfo, error)
	mkdirAll   func(string, os.FileMode) error
	createTemp func(string, string) (suppressionReplacementFile, error)
	rename     func(string, string) error
	openDir    func(string) (suppressionReplacementDirectory, error)
	remove     func(string) error
}

var osSuppressionReplacementOps = suppressionReplacementOps{
	stat:     os.Stat,
	mkdirAll: os.MkdirAll,
	createTemp: func(dir, pattern string) (suppressionReplacementFile, error) {
		return os.CreateTemp(dir, pattern)
	},
	rename: os.Rename,
	openDir: func(path string) (suppressionReplacementDirectory, error) {
		return os.Open(path)
	},
	remove: os.Remove,
}

type suppressionReplacementError struct {
	phase     string
	committed bool
	err       error
}

func (e *suppressionReplacementError) Error() string {
	if e.committed {
		return fmt.Sprintf("suppressions file replacement committed but %s failed: %v", e.phase, e.err)
	}
	return fmt.Sprintf("%s suppressions file: %v", e.phase, e.err)
}

func (e *suppressionReplacementError) Unwrap() error {
	return e.err
}

func (e *suppressionReplacementError) ReplacementCommitted() bool {
	return e.committed
}

// SuppressionReplacementCommitted reports whether err describes a suppression
// file replacement that already crossed the atomic rename boundary.
func SuppressionReplacementCommitted(err error) bool {
	var classified interface {
		ReplacementCommitted() bool
	}
	return errors.As(err, &classified) && classified.ReplacementCommitted()
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
	return &SuppressionManager{
		rules:          cloneSuppressions(rules),
		suppressor:     suppressor,
		filePath:       path,
		replacementOps: osSuppressionReplacementOps,
	}, nil
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
	return saveSuppressionsToFile(path, rules, osSuppressionReplacementOps)
}

func saveSuppressionsToFile(path string, rules []Suppression, ops suppressionReplacementOps) (result error) {
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
	info, err := ops.stat(path)
	if err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return &suppressionReplacementError{phase: "stat", err: err}
	}
	dir := filepath.Dir(path)
	if err := ops.mkdirAll(dir, 0o750); err != nil {
		return &suppressionReplacementError{phase: "create parent directory", err: err}
	}
	tmp, err := ops.createTemp(dir, ".suppressions-*.json")
	if err != nil {
		return &suppressionReplacementError{phase: "create temporary", err: err}
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if err := ops.remove(tmpName); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, &suppressionReplacementError{
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
		return closeSuppressionReplacementAfterFailure(tmp, &suppressionReplacementError{phase: "write temporary", err: err})
	}
	if err := tmp.Chmod(mode); err != nil {
		return closeSuppressionReplacementAfterFailure(tmp, &suppressionReplacementError{phase: "chmod temporary", err: err})
	}
	if err := tmp.Sync(); err != nil {
		return closeSuppressionReplacementAfterFailure(tmp, &suppressionReplacementError{phase: "sync temporary", err: err})
	}
	if err := tmp.Close(); err != nil {
		return &suppressionReplacementError{phase: "close temporary", err: err}
	}
	if err := ops.rename(tmpName, path); err != nil {
		return &suppressionReplacementError{phase: "rename temporary", err: err}
	}
	committed = true

	parent, err := ops.openDir(dir)
	if err != nil {
		return &suppressionReplacementError{phase: "open parent directory", committed: true, err: err}
	}
	if err := parent.Sync(); err != nil {
		return closeSuppressionDirectoryAfterFailure(parent, &suppressionReplacementError{phase: "sync parent directory", committed: true, err: err})
	}
	if err := parent.Close(); err != nil {
		return &suppressionReplacementError{phase: "close parent directory", committed: true, err: err}
	}
	return nil
}

func closeSuppressionReplacementAfterFailure(file suppressionReplacementFile, primary error) error {
	if err := file.Close(); err != nil {
		return errors.Join(primary, &suppressionReplacementError{phase: "close temporary after failure", err: err})
	}
	return primary
}

func closeSuppressionDirectoryAfterFailure(dir suppressionReplacementDirectory, primary error) error {
	if err := dir.Close(); err != nil {
		return errors.Join(primary, &suppressionReplacementError{phase: "close parent directory after failure", committed: true, err: err})
	}
	return primary
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
	return m.replaceLocked(candidate, true)
}

// Update replaces an existing suppression rule by ID, then atomically swaps the compiled filter.
func (m *SuppressionManager) Update(id string, rule Suppression) error {
	if m == nil {
		return fmt.Errorf("suppression manager is not configured")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	candidate := cloneSuppressions(m.rules)
	idx := findSuppression(candidate, id)
	if idx < 0 {
		return fmt.Errorf("suppression %q not found", id)
	}
	rule.ID = id
	candidate[idx] = cloneSuppression(rule)
	return m.replaceLocked(candidate, true)
}

// Delete removes an existing suppression rule by ID, then atomically swaps the compiled filter.
func (m *SuppressionManager) Delete(id string) error {
	if m == nil {
		return fmt.Errorf("suppression manager is not configured")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	candidate := cloneSuppressions(m.rules)
	idx := findSuppression(candidate, id)
	if idx < 0 {
		return fmt.Errorf("suppression %q not found", id)
	}
	candidate = append(candidate[:idx], candidate[idx+1:]...)
	return m.replaceLocked(candidate, true)
}

// ReloadFromFile reloads configured suppressions from disk and atomically swaps the compiled filter.
func (m *SuppressionManager) ReloadFromFile() error {
	if m == nil {
		return fmt.Errorf("suppression manager is not configured")
	}
	if m.filePath == "" {
		return fmt.Errorf("suppressions file is not configured")
	}
	rules, err := LoadSuppressionsFromFile(m.filePath)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.replaceLocked(rules, false)
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

func (m *SuppressionManager) replaceLocked(candidate []Suppression, persist bool) error {
	if err := validateSuppressionSet(candidate); err != nil {
		return err
	}
	suppressor, err := NewSuppressor(candidate)
	if err != nil {
		return err
	}
	if persist && m.filePath != "" {
		if err := saveSuppressionsToFile(m.filePath, candidate, m.replacementOps); err != nil {
			if !SuppressionReplacementCommitted(err) {
				return fmt.Errorf("persist suppressions: %w", err)
			}
			m.rules = cloneSuppressions(candidate)
			m.suppressor = suppressor
			return fmt.Errorf("persist suppressions: %w", err)
		}
	}
	m.rules = cloneSuppressions(candidate)
	m.suppressor = suppressor
	return nil
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

func findSuppression(rules []Suppression, id string) int {
	for i, rule := range rules {
		if rule.ID == id {
			return i
		}
	}
	return -1
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
