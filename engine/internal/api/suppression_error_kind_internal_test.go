package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
)

type suppressionClassificationCommittedError struct{ error }

func (suppressionClassificationCommittedError) ReplacementCommitted() bool { return true }

// This seam verifies classification; it does not simulate filesystem durability.
func TestSuppressionMutationClassificationUsesCausesThroughWrappers(t *testing.T) {
	text := "not found already exists not configured"
	missing := fmt.Errorf("suppression %q %w", text, alert.ErrSuppressionNotFound)
	committed := fmt.Errorf("persist suppressions: %w", suppressionClassificationCommittedError{errors.New(text)})
	persist := fmt.Errorf("persist suppressions: %w", errors.New(text))
	cases := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
		details []string
	}{
		{"wrapped-missing", fmt.Errorf("outer: %w", missing), 404, "SUPPRESSION_NOT_FOUND", "Suppression not found", nil},
		{"joined-missing", errors.Join(errors.New(text), missing), 404, "SUPPRESSION_NOT_FOUND", "Suppression not found", nil},
		{"committed", committed, 500, "SUPPRESSIONS_DURABILITY_UNCERTAIN", "Suppression mutation was applied but durability could not be confirmed", []string{committed.Error()}},
		{"persistence", persist, 500, "INTERNAL_ERROR", "Could not persist suppression", []string{persist.Error()}},
		{"plain-text", errors.New(text), 400, "VALIDATION_ERROR", "Invalid suppression request", []string{text}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/suppressions/prior", nil)
			req.Header.Set("X-Request-ID", "error-kind-fixture")
			rec := httptest.NewRecorder()
			(&Server{}).writeSuppressionMutationError(rec, req, tc.err)
			var response errorEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			want := apiError{Code: tc.code, Message: tc.message, Details: tc.details, RequestID: "error-kind-fixture"}
			if rec.Code != tc.status || rec.Header().Get("Content-Type") != "application/json" || !reflect.DeepEqual(response.Error, want) {
				t.Fatalf("response=%d/%s, want %d/%+v", rec.Code, rec.Body.String(), tc.status, want)
			}
		})
	}
}
