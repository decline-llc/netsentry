package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
)

const suppressionNullBodyDetail = "suppression request body must be a JSON object"

func suppressionNullFixture(t *testing.T, mode string) (string, string, *alert.SuppressionManager) {
	t.Helper()
	fixture := mode
	if mode == "memory" {
		fixture = "absent"
	}
	root, path, manager := suppressionReloadFixture(t, fixture)
	if mode == "memory" {
		var err error
		manager, err = alert.NewSuppressionManager(manager.List())
		if err != nil {
			t.Fatal(err)
		}
	}
	return root, path, manager
}

func TestHTTPSuppressionNullBodiesPreservePersistentAndActiveState(t *testing.T) {
	for _, mode := range []string{"absent", "healthy", "file-parent", "memory"} {
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			for _, auth := range []bool{false, true} {
				for i, body := range []string{"null", " \tnull\r\n", "\r\nnull\t "} {
					t.Run(fmt.Sprintf("%s/%s/auth=%v/body=%d", mode, method, auth, i), func(t *testing.T) {
						root, path, manager := suppressionNullFixture(t, mode)
						probes := suppressionReloadProbes()
						if !reflect.DeepEqual(manager.Filter(probes), probes[1:]) {
							t.Fatal("fixture must actively suppress the first probe")
						}
						before := suppressionSlashSnapshot(t, root, manager, probes)
						endpoint := "/api/suppressions"
						if method == http.MethodPut {
							endpoint += "/prior"
						}
						handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
						rec := suppressionReloadRequest(handler, method, endpoint, body, auth)
						assertSuppressionReloadError(t, rec, 400, "VALIDATION_ERROR", "Invalid suppression request", []string{suppressionNullBodyDetail})
						if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
							t.Fatal("null rejection changed files, membership, modes, rules, filter or caller probes")
						}
						if mode == "healthy" {
							assertSuppressionReloadState(t, manager, path, before.Rules, probes[1:], probes)
							if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
								t.Fatal("independent file observation changed state")
							}
						}
					})
				}
			}
		}
	}
}

func TestHTTPSuppressionNullBodyDiagnosticPrecedence(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, name := range []string{"auth", "manager", "incomplete", "empty", "second-object", "second-null", "malformed-tail", "unknown", "oversized-tail", "missing-target"} {
			t.Run(method+"/"+name, func(t *testing.T) {
				root, _, manager := suppressionNullFixture(t, "healthy")
				opts := api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}
				endpoint, body, authorized := "/api/suppressions", "null", true
				if method == http.MethodPut {
					endpoint += "/prior"
				}
				status, code, message := 400, "VALIDATION_ERROR", "Invalid suppression request"
				details := []string{suppressionNullBodyDetail}
				switch name {
				case "auth":
					opts.Suppressions, authorized = nil, false
					status, code, message, details = 401, "UNAUTHORIZED", "Valid bearer token required", nil
				case "manager":
					opts.Suppressions = nil
					status, code, message, details = 409, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured", nil
				case "incomplete":
					body, details = "nul", []string{"unexpected EOF"}
				case "empty":
					body, details = " \t\r\n", []string{"EOF"}
				case "second-object":
					body, details = "null {}", []string{"request body must contain exactly one JSON document"}
				case "second-null":
					body, details = "null null", []string{"request body must contain exactly one JSON document"}
				case "malformed-tail":
					body, details = "null {", []string{"invalid trailing JSON data: unexpected EOF"}
				case "unknown":
					body, details = `{"unknown":true}`, []string{`json: unknown field "unknown"`}
				case "oversized-tail":
					body = "null" + strings.Repeat(" ", 1<<20)
					status, code, message, details = 413, "REQUEST_TOO_LARGE", "Request body exceeds 1 MiB limit", nil
				case "missing-target":
					if method == http.MethodPut {
						endpoint = "/api/suppressions/missing"
					}
				}
				probes := suppressionReloadProbes()
				before := suppressionSlashSnapshot(t, root, manager, probes)
				rec := suppressionReloadRequest(api.NewServerWithOptions(nil, nil, nil, nil, opts).Handler(), method, endpoint, body, authorized)
				assertSuppressionReloadError(t, rec, status, code, message, details)
				if name == "auth" && rec.Header().Get("WWW-Authenticate") != `Bearer realm="netsentry"` {
					t.Fatal("auth challenge changed")
				}
				if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
					t.Fatal("diagnostic request changed persistent, active or caller state")
				}
			})
		}
	}
	for _, endpoint := range []string{"/api/suppressions/%2e", "/api/suppressions/prior%2Fchild"} {
		t.Run(endpoint, func(t *testing.T) {
			root, _, manager := suppressionNullFixture(t, "healthy")
			probes := suppressionReloadProbes()
			before := suppressionSlashSnapshot(t, root, manager, probes)
			handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			rec := suppressionReloadRequest(handler, http.MethodPut, endpoint, "null", false)
			assertSuppressionReloadError(t, rec, 404, "NOT_FOUND", "Suppression not found", nil)
			if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
				t.Fatal("path rejection changed persistent, active or caller state")
			}
		})
	}
}

func TestHTTPSuppressionObjectBodiesAndRetryRetainSemantics(t *testing.T) {
	for _, mode := range []string{"healthy", "memory"} {
		t.Run(mode, func(t *testing.T) {
			root, path, manager := suppressionNullFixture(t, mode)
			prior := manager.List()[0]
			probes := suppressionReloadProbes()
			originalProbes := suppressionSlashSnapshot(t, root, manager, probes).Probes
			handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			candidate := alert.Suppression{ID: "candidate"}
			steps := []struct {
				method, endpoint, body string
				status                 int
				want                   alert.Suppression
				allRules               bool
			}{
				{http.MethodPost, "/api/suppressions", `{"id":"candidate","enabled":false}`, 201, candidate, false},
				{http.MethodPut, "/api/suppressions/candidate", `{"enabled":false}`, 200, candidate, false},
				{http.MethodPut, "/api/suppressions/candidate", `{}`, 200, candidate, false},
				{http.MethodPut, "/api/suppressions/candidate", `{"enabled":true,"rule_ids":null,"src_cidrs":null,"dst_cidrs":null,"any_cidrs":["203.0.113.0/24"]}`, 200, alert.Suppression{ID: "candidate", Enabled: true, AnyCIDRs: []string{"203.0.113.0/24"}}, true},
			}
			for _, step := range steps {
				var fields map[string]json.RawMessage
				if err := json.Unmarshal([]byte(step.body), &fields); err != nil || fields == nil {
					t.Fatalf("object fixture invalid: %s / %v", step.body, err)
				}
				if step.method == http.MethodPut {
					if _, exists := fields["id"]; exists {
						t.Fatal("omitted-ID fixture emits an id member")
					}
				}
				if step.allRules && string(fields["rule_ids"]) != "null" {
					t.Fatal("optional-null fixture does not emit JSON null")
				}
				before := suppressionSlashSnapshot(t, root, manager, probes)
				rec := suppressionReloadRequest(handler, step.method, step.endpoint, "null", true)
				assertSuppressionReloadError(t, rec, 400, "VALIDATION_ERROR", "Invalid suppression request", []string{suppressionNullBodyDetail})
				if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
					t.Fatal("null request changed state before retry")
				}
				rec = suppressionReloadRequest(handler, step.method, step.endpoint, step.body, true)
				var response alert.Suppression
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != step.status || !reflect.DeepEqual(response, step.want) {
					t.Fatalf("object retry = %d/%s / %v; want %+v", rec.Code, rec.Body.String(), err, step.want)
				}
				wantRules, wantFilter := []alert.Suppression{prior, step.want}, probes[1:]
				if step.allRules {
					wantFilter = append(probes[1:2:2], probes[3])
				}
				if !reflect.DeepEqual(manager.List(), wantRules) || !reflect.DeepEqual(manager.Filter(probes), wantFilter) {
					t.Fatal("object retry changed neighboring rule or filter semantics")
				}
				if mode == "healthy" {
					assertSuppressionReloadState(t, manager, path, wantRules, wantFilter, probes)
				}
				if suppressionSlashSnapshot(t, root, manager, probes).Probes != originalProbes {
					t.Fatal("matching mutated caller probes")
				}
			}
		})
	}
}
