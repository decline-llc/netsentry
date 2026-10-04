package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/stats"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSuppressionHTTPRejectsExplicitEmptyRuleScopeAndPermitsRetry(t *testing.T) {
	for _, operation := range []string{"create", "update", "reload"} {
		for _, ids := range [][]string{{""}, {"", ""}} {
			t.Run(fmt.Sprintf("%s/ids=%d", operation, len(ids)), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "suppression scope.json")
				initial := []alert.Suppression{{ID: "original", Enabled: true, RuleIDs: []string{"old-rule"}, AnyCIDRs: []string{"192.0.2.0/24"}}}
				writeFile := func(rules []alert.Suppression) {
					t.Helper()
					data, err := json.Marshal(struct {
						Suppressions []alert.Suppression `json:"suppressions"`
					}{rules})
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				readFile := func() []byte {
					t.Helper()
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					return data
				}
				writeFile(initial)
				manager, err := alert.NewSuppressionManagerWithFile(initial, path)
				if err != nil {
					t.Fatal(err)
				}
				server := NewServerWithOptions(&fakeStore{}, fakeQueue{}, &fakeRules{}, stats.New(), Options{Suppressions: manager})
				handler := server.Handler()
				candidate := alert.Suppression{ID: "candidate", Enabled: true, RuleIDs: ids, AnyCIDRs: []string{"203.0.113.0/24"}}
				method, endpoint, wantStatus, wantCode := http.MethodPost, "/api/suppressions", http.StatusBadRequest, "VALIDATION_ERROR"
				if operation == "update" {
					candidate.ID = "original"
					method, endpoint = http.MethodPut, "/api/suppressions/original"
				}
				if operation == "reload" {
					endpoint = "/api/suppressions/reload"
					wantStatus, wantCode = http.StatusInternalServerError, "INTERNAL_ERROR"
					writeFile([]alert.Suppression{candidate})
				}
				old := &model.Alert{RuleID: "old-rule", SrcIP: "192.0.2.9"}
				newHit := &model.Alert{RuleID: "new-rule", SrcIP: "203.0.113.9"}
				other := &model.Alert{RuleID: "unrelated-rule", SrcIP: "203.0.113.9"}
				probes := []*model.Alert{old, newHit, other}
				listBefore, fileBefore := manager.List(), readFile()
				wantBefore := []*model.Alert{newHit, other}
				if !reflect.DeepEqual(listBefore, initial) || !reflect.DeepEqual(manager.Filter(probes), wantBefore) {
					t.Fatal("initial manager fixture invalid")
				}
				request := func(rule alert.Suppression) *httptest.ResponseRecorder {
					t.Helper()
					data, err := json.Marshal(rule)
					if err != nil {
						t.Fatal(err)
					}
					req := httptest.NewRequest(method, endpoint, bytes.NewReader(data))
					req.Header.Set("X-Request-ID", "scope-regression")
					rec := httptest.NewRecorder()
					handler.ServeHTTP(rec, req)
					return rec
				}
				rec := request(candidate)
				var envelope errorEnvelope
				if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				wantDetails := []string{fmt.Sprintf("suppression %q rule_ids must include at least one nonempty rule ID", candidate.ID)}
				if rec.Code != wantStatus || rec.Header().Get("Content-Type") != "application/json" || envelope.Error.Code != wantCode || envelope.Error.RequestID != "scope-regression" || !reflect.DeepEqual(envelope.Error.Details, wantDetails) {
					t.Fatalf("response = %d %+v; body=%s", rec.Code, envelope, rec.Body.String())
				}
				if !bytes.Equal(readFile(), fileBefore) || !reflect.DeepEqual(manager.List(), listBefore) || !reflect.DeepEqual(manager.Filter(probes), wantBefore) {
					t.Fatal("HTTP rejection changed persistent/published state")
				}
				valid := candidate
				valid.RuleIDs = []string{"", "new-rule", "new-rule", ""}
				wantRules, wantFiltered, successStatus := []alert.Suppression{valid}, []*model.Alert{old, other}, http.StatusOK
				if operation == "create" {
					wantRules = append(append([]alert.Suppression(nil), initial...), valid)
					wantFiltered, successStatus = []*model.Alert{other}, http.StatusCreated
				}
				if operation == "reload" {
					writeFile(wantRules)
				}
				rec = request(valid)
				if rec.Code != successStatus || !reflect.DeepEqual(manager.List(), wantRules) || !reflect.DeepEqual(manager.Filter(probes), wantFiltered) {
					t.Fatalf("retry = %d %s; list=%+v filter=%+v", rec.Code, rec.Body.String(), manager.List(), manager.Filter(probes))
				}
				if operation == "reload" {
					var response suppressionReloadResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Reloaded != 1 {
						t.Fatalf("reload response = %+v, %v", response, err)
					}
				} else {
					var response alert.Suppression
					if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || !reflect.DeepEqual(response, valid) {
						t.Fatalf("mutation response = %+v, %v", response, err)
					}
				}
				loaded, err := alert.LoadSuppressionsFromFile(path)
				if err != nil || !reflect.DeepEqual(loaded, wantRules) {
					t.Fatalf("retry file = %+v, %v", loaded, err)
				}
			})
		}
	}
}
