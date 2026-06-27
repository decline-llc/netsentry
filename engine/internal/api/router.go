package api

import (
	"context"
	"net/http"

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

type RuleCounter interface {
	RuleCount() int
}

type Server struct {
	store AlertStore
	queue QueueDepthProvider
	rules RuleCounter
	stats *stats.Stats
}

func NewServer(store AlertStore, queue QueueDepthProvider, rules RuleCounter, metrics *stats.Stats) *Server {
	return &Server{store: store, queue: queue, rules: rules, stats: metrics}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/alerts", s.handleAlerts)
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
