package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestHTTPSuppressionErrorPhrasesRetainActualErrorCategories(t *testing.T) {
	for _, id := range []string{"not found", "already exists", "not configured"} {
		for _, operation := range []string{"invalid-create", "invalid-update", "duplicate", "missing-update", "missing-delete", "configured-reload"} {
			t.Run(id+"/"+operation, func(t *testing.T) {
				root, path, manager := suppressionReloadFixture(t, "healthy")
				candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
				probes := suppressionReloadProbes()
				wantFilter := probes[1:]
				if operation == "invalid-update" || operation == "duplicate" {
					if err := manager.Add(candidate); err != nil {
						t.Fatal(err)
					}
					wantFilter = []*model.Alert{probes[1], probes[3], probes[4]}
				}
				method, endpoint := http.MethodPost, "/api/suppressions"
				status, code, message := 400, "VALIDATION_ERROR", "Invalid suppression request"
				details := []string{fmt.Sprintf("suppression %s any cidrs: invalid IP or CIDR %q", id, "bad-cidr")}
				switch operation {
				case "invalid-create":
					candidate.AnyCIDRs = []string{"bad-cidr"}
				case "invalid-update":
					candidate.AnyCIDRs = []string{"bad-cidr"}
					method, endpoint = http.MethodPut, "/api/suppressions/"+url.PathEscape(id)
				case "duplicate":
					status, code, message, details = 409, "SUPPRESSION_ALREADY_EXISTS", "Suppression already exists", nil
				case "missing-update", "missing-delete":
					candidate.ID = "missing " + id
					method, endpoint = http.MethodPut, "/api/suppressions/"+url.PathEscape(candidate.ID)
					if operation == "missing-delete" {
						method = http.MethodDelete
					}
					status, code, message, details = 404, "SUPPRESSION_NOT_FOUND", "Suppression not found", nil
				case "configured-reload":
					candidate.AnyCIDRs = []string{"bad-cidr"}
					if err := alert.SaveSuppressionsToFile(path, []alert.Suppression{candidate}); err != nil {
						t.Fatal(err)
					}
					endpoint = "/api/suppressions/reload"
					status, code, message = 500, "INTERNAL_ERROR", "Could not reload suppressions"
				}
				if !reflect.DeepEqual(manager.Filter(probes), wantFilter) {
					t.Fatal("fixture filter does not match independent probes")
				}
				body := suppressionReloadBody(t, candidate)
				before := suppressionSlashSnapshot(t, root, manager, probes)
				handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
				rec := suppressionReloadRequest(handler, method, endpoint, body, true)
				assertSuppressionReloadError(t, rec, status, code, message, details)
				if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) || body != suppressionReloadBody(t, candidate) {
					t.Fatal("error classification changed files, active state or caller data")
				}
			})
		}
	}
	manager, err := alert.NewSuppressionManager(nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager}).Handler()
	rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions/reload", "", false)
	assertSuppressionReloadError(t, rec, 409, "SUPPRESSIONS_RELOAD_UNAVAILABLE", "Suppressions file is not configured", nil)
}

func TestHTTPSuppressionMisleadingPersistencePathsPreserveStateAndPermitRetry(t *testing.T) {
	for _, phrase := range []string{"not found", "already exists", "not configured"} {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			t.Run(phrase+"/"+method, func(t *testing.T) {
				root := t.TempDir()
				parent := filepath.Join(root, phrase+" parent with spaces")
				path := filepath.Join(parent, "suppressions.json")
				if err := os.WriteFile(parent, []byte("retained non-directory occupant"), 0o600); err != nil {
					t.Fatal(err)
				}
				prior := suppressionReloadRule("prior", "old-rule", "192.0.2.0/24")
				manager, err := alert.NewSuppressionManagerWithFile([]alert.Suppression{prior}, path)
				if err != nil {
					t.Fatal(err)
				}
				candidate := suppressionReloadRule("candidate", "candidate-rule", "203.0.113.0/24")
				endpoint, success := "/api/suppressions", 201
				wantRules := []alert.Suppression{prior, candidate}
				probes := suppressionReloadProbes()
				wantFilter := []*model.Alert{probes[1], probes[3], probes[4]}
				if method != http.MethodPost {
					candidate.ID, endpoint = "prior", "/api/suppressions/prior"
					wantRules, wantFilter, success = []alert.Suppression{candidate}, []*model.Alert{probes[0], probes[1], probes[3], probes[4]}, 200
				}
				if method == http.MethodDelete {
					wantRules, wantFilter, success = []alert.Suppression{}, probes, 204
				}
				if !reflect.DeepEqual(manager.Filter(probes), probes[1:]) {
					t.Fatal("fixture must suppress prior independently")
				}
				body := suppressionReloadBody(t, candidate)
				before := suppressionSlashSnapshot(t, root, manager, probes)
				_, statErr := os.Stat(path)
				if statErr == nil {
					t.Fatal("fixture path unexpectedly exists")
				}
				detail := "persist suppressions: stat suppressions file: " + statErr.Error()
				handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
				rec := suppressionReloadRequest(handler, method, endpoint, body, true)
				assertSuppressionReloadError(t, rec, 500, "INTERNAL_ERROR", "Could not persist suppression", []string{detail})
				if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
					t.Fatal("persistence failure changed tree, modes, membership, active state or caller probes")
				}
				if err := os.Remove(parent); err != nil {
					t.Fatal(err)
				}
				if err := alert.SaveSuppressionsToFile(path, manager.List()); err != nil {
					t.Fatal(err)
				}
				rec = suppressionReloadRequest(handler, method, endpoint, body, true)
				if rec.Code != success {
					t.Fatalf("retry=%d/%s, want %d", rec.Code, rec.Body.String(), success)
				}
				if method != http.MethodDelete {
					var response alert.Suppression
					if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || !reflect.DeepEqual(response, candidate) {
						t.Fatalf("retry response=%+v/%v, want %+v", response, err, candidate)
					}
				}
				assertSuppressionReloadState(t, manager, path, wantRules, wantFilter, probes)
				if suppressionSlashSnapshot(t, root, manager, probes).Probes != before.Probes || suppressionReloadBody(t, candidate) != body {
					t.Fatal("retry changed caller probes or input")
				}
			})
		}
	}
}

func TestHTTPSuppressionErrorPhraseIDsRetainExactCRUDAndReload(t *testing.T) {
	for _, id := range []string{"not found", "already exists", "not configured"} {
		t.Run(id, func(t *testing.T) {
			root, path, manager := suppressionReloadFixture(t, "healthy")
			prior := manager.List()[0]
			probes := suppressionReloadProbes()
			beforeProbes := suppressionSlashSnapshot(t, root, manager, probes).Probes
			candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
			handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			endpoint := "/api/suppressions/" + url.PathEscape(id)
			for step, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
				requestPath, body, status := endpoint, "", 204
				wantRules, wantFilter := []alert.Suppression{prior}, probes[1:]
				if step == 0 {
					requestPath, body, status = "/api/suppressions", suppressionReloadBody(t, candidate), 201
					wantRules, wantFilter = []alert.Suppression{prior, candidate}, []*model.Alert{probes[1], probes[3], probes[4]}
				} else if step == 1 {
					candidate.AnyCIDRs = []string{"198.51.100.0/24"}
					body, status = suppressionSlashBodyWithoutID(t, candidate), 200
					wantRules, wantFilter = []alert.Suppression{prior, candidate}, []*model.Alert{probes[1], probes[2], probes[4]}
				}
				rec := suppressionReloadRequest(handler, method, requestPath, body, true)
				if rec.Code != status {
					t.Fatalf("%s=%d/%s, want %d", method, rec.Code, rec.Body.String(), status)
				}
				if method != http.MethodDelete {
					var response alert.Suppression
					if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || !reflect.DeepEqual(response, candidate) {
						t.Fatalf("exact identity response=%+v/%v, want %+v", response, err, candidate)
					}
				}
				assertSuppressionReloadState(t, manager, path, wantRules, wantFilter, probes)
				for _, read := range []string{"list", "reload"} {
					verb, route := http.MethodGet, "/api/suppressions"
					if read == "reload" {
						verb, route = http.MethodPost, "/api/suppressions/reload"
					}
					before := suppressionSlashSnapshot(t, root, manager, probes)
					rec := suppressionReloadRequest(handler, verb, route, "", true)
					if rec.Code != 200 {
						t.Fatalf("%s=%d/%s", read, rec.Code, rec.Body.String())
					}
					if read == "list" {
						var response struct{ Data []alert.Suppression }
						if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || !reflect.DeepEqual(response.Data, wantRules) {
							t.Fatalf("list=%+v/%v, want %+v", response, err, wantRules)
						}
					} else {
						var response struct{ Reloaded int }
						if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Reloaded != len(wantRules) {
							t.Fatalf("reload=%+v/%v, want %d", response, err, len(wantRules))
						}
					}
					if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
						t.Fatal("list/reload changed files, exact identities, neighbors, filter or callers")
					}
				}
				if suppressionSlashSnapshot(t, root, manager, probes).Probes != beforeProbes {
					t.Fatal("CRUD changed caller probes")
				}
			}
		})
	}
}
