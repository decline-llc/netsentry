package api

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type auditResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *auditResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (s *Server) audit(next http.Handler) http.Handler {
	if s.opts.AuditLogger == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		reqID := requestID(r)
		if r.Header.Get("X-Request-ID") == "" {
			r.Header.Set("X-Request-ID", reqID)
		}
		recorder := &auditResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		if r.Method == http.MethodGet {
			return
		}
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		s.opts.AuditLogger.Info("api audit",
			zap.String("request_id", reqID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", status),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Bool("authorized", status != http.StatusUnauthorized),
			zap.Duration("duration", time.Since(started)),
			zap.String("target", auditTarget(r.URL.Path)),
		)
	})
}

func auditTarget(path string) string {
	switch {
	case path == "/api/rules" || strings.HasPrefix(path, "/api/rules/"):
		return "rules"
	case path == "/api/suppressions" || strings.HasPrefix(path, "/api/suppressions/"):
		return "suppressions"
	default:
		return "api"
	}
}
