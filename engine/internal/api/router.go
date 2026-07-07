package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/receiver"
	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
	"go.uber.org/zap"
)

type AlertStore interface {
	List(ctx context.Context) ([]*model.Alert, error)
	Count(ctx context.Context) (int, error)
}

type AlertQueryStore interface {
	Query(ctx context.Context, query alert.Query) ([]*model.Alert, int, error)
}

type StoragePathProvider interface {
	Path() string
}

type StorageHealthProvider interface {
	Health() alert.StorageHealth
}

type QueueDepthProvider interface {
	QueueDepth() int
}

type CaptureStateProvider interface {
	State() receiver.State
}

type SuppressionManager interface {
	List() []alert.Suppression
	Add(alert.Suppression) error
	Update(string, alert.Suppression) error
	Delete(string) error
	ReloadFromFile() error
}

type RuleManager interface {
	RuleCount() int
	Rules() []*model.Rule
	Reload([]*model.Rule) error
}

type Options struct {
	RulesSeedFile        string
	AuthEnabled          bool
	AuthToken            string
	HealthFreshnessLimit time.Duration
	Suppressions         SuppressionManager
	AuditLogger          *zap.Logger
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
	mux.HandleFunc("/api/suppressions", s.handleSuppressions)
	mux.HandleFunc("/api/suppressions/", s.handleSuppressionByID)
	mux.HandleFunc("/api/suppressions/reload", s.handleSuppressionsReload)
	mux.HandleFunc("/api/rules", s.handleRules)
	mux.HandleFunc("/api/rules/", s.handleRuleByID)
	mux.HandleFunc("/api/rules/reload", s.handleRulesReload)
	return s.audit(mux)
}

func (s *Server) requireMutationAuth(w http.ResponseWriter, r *http.Request) bool {
	if !s.opts.AuthEnabled {
		return true
	}
	const bearerPrefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, bearerPrefix) {
		writeAuthError(w, r)
		return false
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(auth, bearerPrefix)), []byte(s.opts.AuthToken)) != 1 {
		writeAuthError(w, r)
		return false
	}
	return true
}

func writeAuthError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="netsentry"`)
	writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Valid bearer token required")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}
	count, err := s.store.Count(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not count alerts")
		return
	}
	if r.URL.Query().Get("verbose") != "true" {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"alerts": count,
		})
		return
	}
	writeJSON(w, http.StatusOK, s.verboseHealth(count))
}

type verboseHealthResponse struct {
	Status     string           `json:"status"`
	Alerts     int              `json:"alerts"`
	Capture    captureHealth    `json:"capture"`
	Engine     engineHealth     `json:"engine"`
	Storage    storageHealth    `json:"storage"`
	Throughput throughputHealth `json:"throughput"`
}

type captureHealth struct {
	Status                string                  `json:"status"`
	SessionID             string                  `json:"session_id,omitempty"`
	Hello                 receiver.HelloFrame     `json:"hello,omitempty"`
	Heartbeat             receiver.HeartbeatFrame `json:"heartbeat,omitempty"`
	LastHeartbeatAt       string                  `json:"last_heartbeat_at,omitempty"`
	HeartbeatAgeSeconds   float64                 `json:"heartbeat_age_seconds,omitempty"`
	FreshnessLimitSeconds float64                 `json:"freshness_limit_seconds"`
}

type engineHealth struct {
	QueueDepth  int `json:"queue_depth"`
	RulesLoaded int `json:"rules_loaded"`
}

type storageHealth struct {
	Status         string  `json:"status"`
	Alerts         int     `json:"alerts"`
	AvailableBytes *uint64 `json:"available_bytes,omitempty"`
	LastError      string  `json:"last_error,omitempty"`
	LastErrorAt    string  `json:"last_error_at,omitempty"`
}

type throughputHealth struct {
	FramesTotal      uint64                    `json:"frames_total"`
	ControlFrames    uint64                    `json:"control_frames"`
	PacketsReceived  uint64                    `json:"packets_received"`
	PacketsProcessed uint64                    `json:"packets_processed"`
	DecodeErrors     uint64                    `json:"decode_errors"`
	AlertsGenerated  uint64                    `json:"alerts_generated"`
	WorkerPanics     uint64                    `json:"worker_panics"`
	AlertWriteErrors uint64                    `json:"alert_write_errors"`
	AlertsBySeverity map[model.Severity]uint64 `json:"alerts_by_severity"`
}

func (s *Server) verboseHealth(alertCount int) verboseHealthResponse {
	capture := s.captureHealth()
	storage := s.storageHealth(alertCount)
	status := "ok"
	if capture.Status == "stale" || storage.Status != "ok" {
		status = "degraded"
	}
	return verboseHealthResponse{
		Status:  status,
		Alerts:  alertCount,
		Capture: capture,
		Engine: engineHealth{
			QueueDepth:  s.queue.QueueDepth(),
			RulesLoaded: s.rules.RuleCount(),
		},
		Storage:    storage,
		Throughput: throughputFromStats(s.stats.Snapshot()),
	}
}

func (s *Server) storageHealth(alertCount int) storageHealth {
	out := storageHealth{
		Status: "ok",
		Alerts: alertCount,
	}
	if available, ok := s.storageAvailableBytes(); ok {
		out.AvailableBytes = &available
	}
	provider, ok := s.store.(StorageHealthProvider)
	if !ok {
		return out
	}
	health := provider.Health()
	if health.Status != "" {
		out.Status = health.Status
	}
	out.LastError = health.LastError
	if !health.LastErrorAt.IsZero() {
		out.LastErrorAt = health.LastErrorAt.Format(time.RFC3339Nano)
	}
	return out
}

func throughputFromStats(snapshot stats.Snapshot) throughputHealth {
	return throughputHealth{
		FramesTotal:      snapshot.FramesTotal,
		ControlFrames:    snapshot.ControlFrames,
		PacketsReceived:  snapshot.PacketsReceived,
		PacketsProcessed: snapshot.PacketsProcessed,
		DecodeErrors:     snapshot.DecodeErrors,
		AlertsGenerated:  snapshot.AlertsGenerated,
		WorkerPanics:     snapshot.WorkerPanics,
		AlertWriteErrors: snapshot.AlertWriteErrors,
		AlertsBySeverity: snapshot.AlertsBySeverity,
	}
}

func (s *Server) captureHealth() captureHealth {
	limit := s.opts.HealthFreshnessLimit
	if limit <= 0 {
		limit = 30 * time.Second
	}
	out := captureHealth{
		Status:                "unknown",
		FreshnessLimitSeconds: limit.Seconds(),
	}
	provider, ok := s.queue.(CaptureStateProvider)
	if !ok {
		return out
	}
	state := provider.State()
	out.SessionID = state.SessionID
	out.Hello = state.Hello
	out.Heartbeat = state.Heartbeat
	if state.LastHeartbeatAt.IsZero() {
		return out
	}
	age := time.Since(state.LastHeartbeatAt)
	out.LastHeartbeatAt = state.LastHeartbeatAt.Format(time.RFC3339Nano)
	out.HeartbeatAgeSeconds = age.Seconds()
	out.Status = "ok"
	if age > limit {
		out.Status = "stale"
	}
	return out
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}
	count, err := s.store.Count(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not count alerts")
		return
	}
	queueDepth := s.queue.QueueDepth()
	s.stats.ObserveQueueDepth(queueDepth)
	snapshot := s.stats.Snapshot()
	body := stats.RenderPrometheus(snapshot, s.metricsGauges(snapshot, count, queueDepth))
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (s *Server) metricsGauges(snapshot stats.Snapshot, alertCount int, queueDepth int) map[string]float64 {
	gauges := map[string]float64{
		"netsentry_alerts_current":     float64(alertCount),
		"netsentry_packet_queue_depth": float64(queueDepth),
		"netsentry_rules_loaded":       float64(s.rules.RuleCount()),
	}
	uptime := time.Since(snapshot.StartedAt).Seconds()
	if snapshot.StartedAt.IsZero() || uptime < 0 {
		uptime = 0
	}
	gauges["netsentry_process_uptime_seconds"] = uptime
	if uptime > 0 {
		gauges["netsentry_packets_processed_per_second"] = float64(snapshot.PacketsProcessed) / uptime
		gauges["netsentry_alerts_generated_per_second"] = float64(snapshot.AlertsGenerated) / uptime
	} else {
		gauges["netsentry_packets_processed_per_second"] = 0
		gauges["netsentry_alerts_generated_per_second"] = 0
	}
	if available, ok := s.storageAvailableBytes(); ok {
		gauges["netsentry_storage_available_bytes"] = float64(available)
	}
	if healthy, ok := s.storageHealthyGauge(); ok {
		gauges["netsentry_storage_healthy"] = healthy
	}
	capture := s.captureHealth()
	if capture.Status == "unknown" {
		return gauges
	}
	if capture.Status == "ok" {
		gauges["netsentry_capture_connected"] = 1
	} else {
		gauges["netsentry_capture_connected"] = 0
	}
	gauges["netsentry_capture_heartbeat_age_seconds"] = capture.HeartbeatAgeSeconds
	gauges["netsentry_capture_packets_sent"] = float64(capture.Heartbeat.Sent)
	gauges["netsentry_capture_packets_dropped"] = float64(capture.Heartbeat.Dropped)
	gauges["netsentry_capture_parse_errors"] = float64(capture.Heartbeat.ParseErrors)
	gauges["netsentry_capture_uds_write_errors"] = float64(capture.Heartbeat.UDSWriteErrors)
	gauges["netsentry_capture_avg_json_serialize_seconds"] = capture.Heartbeat.AvgJSONSerializeUS / float64(time.Second/time.Microsecond)
	return gauges
}

func (s *Server) storageHealthyGauge() (float64, bool) {
	provider, ok := s.store.(StorageHealthProvider)
	if !ok {
		return 0, false
	}
	if provider.Health().Status == "ok" {
		return 1, true
	}
	return 0, true
}

func (s *Server) storageAvailableBytes() (uint64, bool) {
	provider, ok := s.store.(StoragePathProvider)
	if !ok {
		return 0, false
	}
	path := provider.Path()
	if path == "" {
		return 0, false
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Dir(path), &stat); err != nil {
		return 0, false
	}
	return stat.Bavail * uint64(stat.Bsize), true
}

type suppressionListResponse struct {
	Data []alert.Suppression `json:"data"`
}

type suppressionReloadResponse struct {
	Reloaded int `json:"reloaded"`
}

func (s *Server) handleSuppressions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rules := []alert.Suppression{}
		if s.opts.Suppressions != nil {
			rules = s.opts.Suppressions.List()
		}
		writeJSON(w, http.StatusOK, suppressionListResponse{Data: rules})
	case http.MethodPost:
		if !s.requireMutationAuth(w, r) {
			return
		}
		if s.opts.Suppressions == nil {
			writeError(w, r, http.StatusConflict, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured")
			return
		}
		suppression, err := decodeSuppression(r)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid suppression request", err.Error())
			return
		}
		if err := s.opts.Suppressions.Add(*suppression); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				writeError(w, r, http.StatusConflict, "SUPPRESSION_ALREADY_EXISTS", "Suppression already exists")
				return
			}
			if strings.HasPrefix(err.Error(), "persist suppressions:") {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not persist suppression", err.Error())
				return
			}
			writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid suppression request", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, suppression)
	default:
		writeMethodNotAllowed(w, r, "GET, POST")
	}
}

func (s *Server) handleSuppressionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/suppressions/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") || id == "reload" {
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Suppression not found")
		return
	}
	switch r.Method {
	case http.MethodPut:
		if !s.requireMutationAuth(w, r) {
			return
		}
		s.handleSuppressionUpdate(w, r, id)
	case http.MethodDelete:
		if !s.requireMutationAuth(w, r) {
			return
		}
		s.handleSuppressionDelete(w, r, id)
	default:
		writeMethodNotAllowed(w, r, "PUT, DELETE")
	}
}

func (s *Server) handleSuppressionUpdate(w http.ResponseWriter, r *http.Request, id string) {
	if s.opts.Suppressions == nil {
		writeError(w, r, http.StatusConflict, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured")
		return
	}
	suppression, err := decodeSuppression(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid suppression request", err.Error())
		return
	}
	if suppression.ID != "" && suppression.ID != id {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Suppression ID must match path")
		return
	}
	suppression.ID = id
	if err := s.opts.Suppressions.Update(id, *suppression); err != nil {
		s.writeSuppressionMutationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, suppression)
}

func (s *Server) handleSuppressionDelete(w http.ResponseWriter, r *http.Request, id string) {
	if s.opts.Suppressions == nil {
		writeError(w, r, http.StatusConflict, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured")
		return
	}
	if err := s.opts.Suppressions.Delete(id); err != nil {
		s.writeSuppressionMutationError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSuppressionsReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r, http.MethodPost)
		return
	}
	if !s.requireMutationAuth(w, r) {
		return
	}
	if s.opts.Suppressions == nil {
		writeError(w, r, http.StatusConflict, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured")
		return
	}
	if err := s.opts.Suppressions.ReloadFromFile(); err != nil {
		if strings.Contains(err.Error(), "not configured") {
			writeError(w, r, http.StatusConflict, "SUPPRESSIONS_RELOAD_UNAVAILABLE", "Suppressions file is not configured")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not reload suppressions", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, suppressionReloadResponse{Reloaded: len(s.opts.Suppressions.List())})
}

func (s *Server) writeSuppressionMutationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case strings.Contains(err.Error(), "not found"):
		writeError(w, r, http.StatusNotFound, "SUPPRESSION_NOT_FOUND", "Suppression not found")
	case strings.HasPrefix(err.Error(), "persist suppressions:"):
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not persist suppression", err.Error())
	default:
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid suppression request", err.Error())
	}
}

func decodeSuppression(r *http.Request) (*alert.Suppression, error) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var suppression alert.Suppression
	if err := dec.Decode(&suppression); err != nil {
		return nil, err
	}
	return &suppression, nil
}

type alertListResponse struct {
	Data       []*model.Alert `json:"data"`
	Pagination pagination     `json:"pagination"`
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}
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
	if queryStore, ok := s.store.(AlertQueryStore); ok {
		alerts, total, err := queryStore.Query(r.Context(), filters.toAlertQuery(p))
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list alerts")
			return
		}
		p.Total = total
		writeJSON(w, http.StatusOK, alertListResponse{
			Data:       alerts,
			Pagination: p,
		})
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
		if !s.requireMutationAuth(w, r) {
			return
		}
		s.handleRuleCreate(w, r)
	default:
		writeMethodNotAllowed(w, r, "GET, POST")
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
		if !s.requireMutationAuth(w, r) {
			return
		}
		s.handleRuleUpdate(w, r, id)
	case http.MethodDelete:
		if !s.requireMutationAuth(w, r) {
			return
		}
		s.handleRuleDelete(w, r, id)
	default:
		writeMethodNotAllowed(w, r, "PUT, DELETE")
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
		writeMethodNotAllowed(w, r, http.MethodPost)
		return
	}
	if !s.requireMutationAuth(w, r) {
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
