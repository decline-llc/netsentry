package api

import (
	"context"
	"net/http"

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
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, ruleListResponse{Data: s.rules.Rules()})
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
