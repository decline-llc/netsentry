package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

type AlertStore interface {
	List(ctx context.Context) ([]*model.Alert, error)
	Count(ctx context.Context) (int, error)
}

type QueueDepthProvider interface {
	QueueDepth() int
}

type RuleManager interface {
	RuleCount() int
	Rules() []*model.Rule
	Reload([]*model.Rule) error
}

type Options struct {
	RulesSeedFile string
}

type Server struct {
	store AlertStore
	queue QueueDepthProvider
	rules RuleManager
	stats *stats.Stats
	opts  Options
}

func NewServer(store AlertStore, queue QueueDepthProvider, rules RuleManager, metrics *stats.Stats) *Server {
	return NewServerWithOptions(store, queue, rules, metrics, Options{})
}

func NewServerWithOptions(store AlertStore, queue QueueDepthProvider, rules RuleManager, metrics *stats.Stats, opts Options) *Server {
	return &Server{store: store, queue: queue, rules: rules, stats: metrics, opts: opts}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/alerts", s.handleAlerts)
	mux.HandleFunc("/api/rules", s.handleRules)
	mux.HandleFunc("/api/rules/", s.handleRuleByID)
	mux.HandleFunc("/api/rules/reload", s.handleRulesReload)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	count, err := s.store.Count(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not count alerts")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"alerts": count,
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	count, err := s.store.Count(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not count alerts")
		return
	}
	body := stats.RenderPrometheus(s.stats.Snapshot(), map[string]float64{
		"netsentry_alerts_current":     float64(count),
		"netsentry_packet_queue_depth": float64(s.queue.QueueDepth()),
		"netsentry_rules_loaded":       float64(s.rules.RuleCount()),
	})
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

type alertListResponse struct {
	Data       []*model.Alert `json:"data"`
	Pagination pagination     `json:"pagination"`
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	p, err := parsePagination(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid pagination parameters", err.Error())
		return
	}
	filters, err := parseAlertFilters(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid alert filters", err.Error())
		return
	}
	alerts, err := s.store.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list alerts")
		return
	}
	filtered := applyAlertFilters(alerts, filters)
	p.Total = len(filtered)
	start, end := pageBounds(len(filtered), p)
	writeJSON(w, http.StatusOK, alertListResponse{
		Data:       filtered[start:end],
		Pagination: p,
	})
}

type ruleListResponse struct {
	Data []*model.Rule `json:"data"`
}

type ruleReloadResponse struct {
	Reloaded int `json:"reloaded"`
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, ruleListResponse{Data: s.rules.Rules()})
	case http.MethodPost:
		s.handleRuleCreate(w, r)
	default:
		writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (s *Server) handleRuleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/rules/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Rule not found")
		return
	}
	switch r.Method {
	case http.MethodPut:
		s.handleRuleUpdate(w, r, id)
	case http.MethodDelete:
		s.handleRuleDelete(w, r, id)
	default:
		writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (s *Server) handleRuleCreate(w http.ResponseWriter, r *http.Request) {
	if s.opts.RulesSeedFile == "" {
		writeError(w, r, http.StatusConflict, "RULES_WRITE_UNAVAILABLE", "Rules seed file is not configured")
		return
	}
	newRule, err := decodeRule(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rule request", err.Error())
		return
	}
	if err := validateRuleBasics(newRule); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rule request", err.Error())
		return
	}
	rules := s.rules.Rules()
	if findRuleIndex(rules, newRule.ID) >= 0 {
		writeError(w, r, http.StatusConflict, "RULE_ALREADY_EXISTS", "Rule already exists")
		return
	}
	candidate := append(cloneRules(rules), newRule)
	if err := s.persistAndReloadRules(candidate); err != nil {
		writeRuleMutationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, newRule)
}

func (s *Server) handleRuleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	if s.opts.RulesSeedFile == "" {
		writeError(w, r, http.StatusConflict, "RULES_WRITE_UNAVAILABLE", "Rules seed file is not configured")
		return
	}
	updated, err := decodeRule(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rule request", err.Error())
		return
	}
	if updated.ID == "" {
		updated.ID = id
	}
	if updated.ID != id {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Rule ID in path and body must match")
		return
	}
	if err := validateRuleBasics(updated); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rule request", err.Error())
		return
	}
	rules := cloneRules(s.rules.Rules())
	idx := findRuleIndex(rules, id)
	if idx < 0 {
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Rule not found")
		return
	}
	rules[idx] = updated
	if err := s.persistAndReloadRules(rules); err != nil {
		writeRuleMutationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleRuleDelete(w http.ResponseWriter, r *http.Request, id string) {
	if s.opts.RulesSeedFile == "" {
		writeError(w, r, http.StatusConflict, "RULES_WRITE_UNAVAILABLE", "Rules seed file is not configured")
		return
	}
	rules := cloneRules(s.rules.Rules())
	idx := findRuleIndex(rules, id)
	if idx < 0 {
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Rule not found")
		return
	}
	rules = append(rules[:idx], rules[idx+1:]...)
	if err := s.persistAndReloadRules(rules); err != nil {
		writeRuleMutationError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeRule(r *http.Request) (*model.Rule, error) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var rule model.Rule
	if err := dec.Decode(&rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func validateRuleBasics(r *model.Rule) error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if strings.ContainsAny(r.ID, "/?#") {
		return fmt.Errorf("id cannot contain /, ?, or #")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	switch r.Type {
	case model.RuleTypePayloadMatch, model.RuleTypeIPBlacklist, model.RuleTypePortBlacklist:
	default:
		return fmt.Errorf("unsupported rule type %q", r.Type)
	}
	switch r.Severity {
	case model.SeverityLow, model.SeverityMedium, model.SeverityHigh, model.SeverityCritical:
	default:
		return fmt.Errorf("unsupported severity %q", r.Severity)
	}
	return nil
}

func (s *Server) persistAndReloadRules(rules []*model.Rule) error {
	validator := rule.NewEngine()
	if err := validator.Reload(rules); err != nil {
		return fmt.Errorf("validate rules: %w", err)
	}
	if err := rule.SaveToFile(s.opts.RulesSeedFile, rules); err != nil {
		return fmt.Errorf("save rules: %w", err)
	}
	loaded, err := rule.LoadFromFile(s.opts.RulesSeedFile)
	if err != nil {
		return fmt.Errorf("reload saved rules: %w", err)
	}
	if err := s.rules.Reload(loaded); err != nil {
		return fmt.Errorf("reload rules: %w", err)
	}
	return nil
}

func writeRuleMutationError(w http.ResponseWriter, r *http.Request, err error) {
	if strings.HasPrefix(err.Error(), "validate rules:") {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rule request", err.Error())
		return
	}
	writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not persist rules", err.Error())
}

func findRuleIndex(rules []*model.Rule, id string) int {
	for i, r := range rules {
		if r != nil && r.ID == id {
			return i
		}
	}
	return -1
}

func cloneRules(rules []*model.Rule) []*model.Rule {
	out := make([]*model.Rule, 0, len(rules))
	for _, r := range rules {
		if r == nil {
			continue
		}
		clone := *r
		clone.Config = append(json.RawMessage(nil), r.Config...)
		clone.MITRETechs = append([]model.MITRETechnique(nil), r.MITRETechs...)
		out = append(out, &clone)
	}
	return out
}

func (s *Server) handleRulesReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	if s.opts.RulesSeedFile == "" {
		writeError(w, r, http.StatusConflict, "RULES_RELOAD_UNAVAILABLE", "Rules seed file is not configured")
		return
	}
	rules, err := rule.LoadFromFile(s.opts.RulesSeedFile)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load rules", err.Error())
		return
	}
	if err := s.rules.Reload(rules); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Could not reload rules", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ruleReloadResponse{Reloaded: len(rules)})
}
