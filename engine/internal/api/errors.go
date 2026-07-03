package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Details   []string `json:"details"`
	RequestID string   `json:"request_id"`
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details ...string) {
	writeJSON(w, status, errorEnvelope{Error: apiError{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: requestID(r),
	}})
}

func writeMethodNotAllowed(w http.ResponseWriter, r *http.Request, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
}

func requestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return fmt.Sprintf("req_%x", time.Now().UTC().UnixNano())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
