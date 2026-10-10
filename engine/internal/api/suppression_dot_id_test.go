package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/pkg/model"
)

func suppressionDotFixture(t *testing.T, fixture string) (string, string, *alert.SuppressionManager) {
	t.Helper()
	base := fixture
	if fixture == "legacy-dots" {
		base = "healthy"
	}
	root, path, manager := suppressionReloadFixture(t, base)
	probes := suppressionReloadProbes()
	want := probes[1:]
	if fixture == "legacy-dots" {
		for _, candidate := range []alert.Suppression{
			suppressionReloadRule(".", "legacy-rule", "198.51.100.0/24"),
			suppressionReloadRule("..", "candidate-rule", "203.0.113.0/24"),
		} {
			if err := manager.Add(candidate); err != nil {
				t.Fatal(err)
			}
		}
		want = probes[3:]
		assertSuppressionDotState(t, root, path, manager, manager.List(), want, probes)
	}
	if !reflect.DeepEqual(manager.Filter(probes), want) {
		t.Fatal("initial independent filtering fixture invalid")
	}
	return root, path, manager
}

func assertSuppressionDotState(t *testing.T, root, path string, manager *alert.SuppressionManager, rules []alert.Suppression, want, probes []*model.Alert) {
	t.Helper()
	before := suppressionSlashSnapshot(t, root, manager, probes)
	assertSuppressionReloadState(t, manager, path, rules, want, probes)
	if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
		t.Fatal("read-only load/rebuilt-manager/filter observation changed artifacts or callers")
	}
}

func TestHTTPSuppressionCreateRejectsDotIDsWithoutMutation(t *testing.T) {
	for _, id := range []string{".", ".."} {
		for _, fixture := range []string{"absent", "healthy", "legacy-dots", "file-parent"} {
			for _, enabled := range []bool{false, true} {
				for _, auth := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/enabled=%v/auth=%v", id, fixture, enabled, auth), func(t *testing.T) {
						root, _, manager := suppressionDotFixture(t, fixture)
						candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
						candidate.Enabled = enabled
						body := suppressionReloadBody(t, candidate)
						probes := suppressionReloadProbes()
						before := suppressionSlashSnapshot(t, root, manager, probes)
						handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
						rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions", body, auth)
						assertSuppressionReloadError(t, rec, 400, "VALIDATION_ERROR", "Invalid suppression request", []string{"id cannot be . or .."})
						if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) || suppressionReloadBody(t, candidate) != body {
							t.Fatal("dot creation rejection changed tree, rules, filter or callers")
						}
					})
				}
			}
		}
	}
}

func TestHTTPSuppressionDotManagementAndRawRedirectBoundaries(t *testing.T) {
	for _, target := range []string{"/api/suppressions/%2e", "/api/suppressions/%2E", "/api/suppressions/%2e%2e", "/api/suppressions/%2E%2E"} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			for _, mode := range []string{"no-auth", "authorized", "missing-auth"} {
				t.Run(target+"/"+method+"/"+mode, func(t *testing.T) {
					root, _, manager := suppressionDotFixture(t, "legacy-dots")
					probes := suppressionReloadProbes()
					before := suppressionSlashSnapshot(t, root, manager, probes)
					candidate := suppressionReloadRule(".", "candidate-rule", "198.51.100.0/24")
					caller := suppressionReloadBody(t, candidate)
					body := suppressionSlashBodyWithoutID(t, candidate)
					if mode == "missing-auth" {
						body = "{malformed"
					}
					handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: mode != "no-auth", AuthToken: "fixture-auth-value"}).Handler()
					rec := suppressionReloadRequest(handler, method, target, body, mode == "authorized")
					assertSuppressionReloadError(t, rec, 404, "NOT_FOUND", "Suppression not found", nil)
					if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) || suppressionReloadBody(t, candidate) != caller {
						t.Fatal("decoded dot rejection changed requested/neighbor state or callers")
					}
				})
			}
		}
	}
	for _, id := range []string{".", ".."} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			t.Run("raw/"+id+"/"+method, func(t *testing.T) {
				root, _, manager := suppressionDotFixture(t, "legacy-dots")
				probes := suppressionReloadProbes()
				before := suppressionSlashSnapshot(t, root, manager, probes)
				handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
				rec := suppressionReloadRequest(handler, method, "/api/suppressions/"+id, "{malformed", false)
				location := "/api/suppressions"
				if id == ".." {
					location = "/api"
				}
				if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != location || !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
					t.Fatalf("raw dot redirect = %d/%q, want 307/%q with preserved state", rec.Code, rec.Header().Get("Location"), location)
				}
			})
		}
	}
}

func TestHTTPSuppressionDotCreationDiagnosticPrecedence(t *testing.T) {
	for _, name := range []string{"auth", "manager", "malformed", "unknown-field", "required", "slash", "reload", "before-cidr", "before-compiler"} {
		t.Run(name, func(t *testing.T) {
			root, _, manager := suppressionDotFixture(t, "legacy-dots")
			probes := suppressionReloadProbes()
			before := suppressionSlashSnapshot(t, root, manager, probes)
			opts := api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}
			candidate := suppressionReloadRule(".", "candidate-rule", "203.0.113.0/24")
			status, code, message, authorized := 400, "VALIDATION_ERROR", "Invalid suppression request", true
			details := []string{"id cannot be . or .."}
			switch name {
			case "auth":
				opts.Suppressions, authorized = nil, false
				status, code, message, details = 401, "UNAUTHORIZED", "Valid bearer token required", nil
			case "manager":
				opts.Suppressions = nil
				status, code, message, details = 409, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured", nil
			case "malformed":
				details = []string{"unexpected EOF"}
			case "unknown-field":
				details = []string{`json: unknown field "unknown"`}
			case "required":
				candidate.ID, details = "", []string{"id is required"}
			case "slash":
				candidate.ID, details = "./", []string{suppressionSlashIDDetail}
			case "reload":
				candidate.ID, details = "reload", []string{suppressionReloadIDDetail}
			case "before-cidr":
				candidate.AnyCIDRs = nil
			case "before-compiler":
				candidate.AnyCIDRs = []string{"invalid-prefix"}
			}
			body := suppressionReloadBody(t, candidate)
			if name == "auth" || name == "manager" || name == "malformed" {
				body = "{"
			} else if name == "unknown-field" {
				body = `{"id":".","unknown":true}`
			}
			rec := suppressionReloadRequest(api.NewServerWithOptions(nil, nil, nil, nil, opts).Handler(), http.MethodPost, "/api/suppressions", body, authorized)
			assertSuppressionReloadError(t, rec, status, code, message, details)
			if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
				t.Fatal("diagnostic control changed artifacts, rules/filter or callers")
			}
		})
	}
}

func assertSuppressionDotReload(t *testing.T, root string, handler http.Handler, manager *alert.SuppressionManager, count int, probes []*model.Alert) {
	t.Helper()
	before := suppressionSlashSnapshot(t, root, manager, probes)
	rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions/reload", "", true)
	var response struct{ Reloaded int }
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || response.Reloaded != count || !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
		t.Fatalf("reload = %d/%s, error = %v or changed read-only state", rec.Code, rec.Body.String(), err)
	}
}

func TestHTTPSuppressionOtherDotAndPercentIDsRetainExactCRUD(t *testing.T) {
	for _, id := range []string{"...", ".prior", "prior.", "prior..id", "%2e", "%2E"} {
		t.Run(id, func(t *testing.T) {
			root, path, manager := suppressionDotFixture(t, "healthy")
			prior := manager.List()[0]
			probes := suppressionReloadProbes()
			handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
			body := suppressionReloadBody(t, candidate)
			rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions", body, true)
			var response alert.Suppression
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 201 || !reflect.DeepEqual(response, candidate) || suppressionReloadBody(t, candidate) != body {
				t.Fatalf("create = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
			}
			rules := []alert.Suppression{prior, candidate}
			assertSuppressionDotState(t, root, path, manager, rules, []*model.Alert{probes[1], probes[3], probes[4]}, probes)
			list := suppressionReloadRequest(handler, http.MethodGet, "/api/suppressions", "", false)
			var listed struct{ Data []alert.Suppression }
			if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || list.Code != 200 || !reflect.DeepEqual(listed.Data, rules) {
				t.Fatalf("list = %d/%s, error = %v", list.Code, list.Body.String(), err)
			}
			updated := suppressionReloadRule(id, "candidate-rule", "198.51.100.0/24")
			body = suppressionReloadBody(t, updated)
			target := "/api/suppressions/" + url.PathEscape(id)
			rec = suppressionReloadRequest(handler, http.MethodPut, target, suppressionSlashBodyWithoutID(t, updated), true)
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || !reflect.DeepEqual(response, updated) || suppressionReloadBody(t, updated) != body {
				t.Fatalf("update = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
			}
			rules = []alert.Suppression{prior, updated}
			want := []*model.Alert{probes[1], probes[2], probes[4]}
			assertSuppressionDotState(t, root, path, manager, rules, want, probes)
			assertSuppressionDotReload(t, root, handler, manager, 2, probes)
			assertSuppressionDotState(t, root, path, manager, rules, want, probes)
			rec = suppressionReloadRequest(handler, http.MethodDelete, target, "", true)
			if rec.Code != 204 || rec.Body.Len() != 0 {
				t.Fatalf("delete = %d/%s", rec.Code, rec.Body.String())
			}
			assertSuppressionDotState(t, root, path, manager, []alert.Suppression{prior}, probes[1:], probes)
		})
	}
}

func TestHTTPFileLoadedDotSuppressionIDsRetainManagerCompatibility(t *testing.T) {
	root, path, manager := suppressionDotFixture(t, "legacy-dots")
	probes := suppressionReloadProbes()
	seeds := manager.List()
	caller := suppressionReloadBody(t, seeds[1]) + suppressionReloadBody(t, seeds[2])
	handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	assertSuppressionDotReload(t, root, handler, manager, 3, probes)
	assertSuppressionDotState(t, root, path, manager, seeds, probes[3:], probes)
	edited := manager.List()
	edited[1].AnyCIDRs, edited[2].AnyCIDRs = edited[2].AnyCIDRs, edited[1].AnyCIDRs
	if err := alert.SaveSuppressionsToFile(path, edited); err != nil {
		t.Fatal(err)
	}
	before := suppressionReloadTree(t, root)
	rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions/reload", "", true)
	var reloaded struct{ Reloaded int }
	if err := json.Unmarshal(rec.Body.Bytes(), &reloaded); err != nil || rec.Code != 200 || reloaded.Reloaded != 3 || !reflect.DeepEqual(suppressionReloadTree(t, root), before) {
		t.Fatalf("edited legacy reload = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
	}
	assertSuppressionDotState(t, root, path, manager, edited, []*model.Alert{probes[1], probes[2], probes[4]}, probes)
	if suppressionReloadBody(t, seeds[1])+suppressionReloadBody(t, seeds[2]) != caller {
		t.Fatal("file edit changed original callers")
	}
	if err := manager.Delete("."); err != nil {
		t.Fatal(err)
	}
	assertSuppressionDotState(t, root, path, manager, []alert.Suppression{seeds[0], edited[2]}, []*model.Alert{probes[1], probes[2], probes[4]}, probes)
	if err := manager.Delete(".."); err != nil {
		t.Fatal(err)
	}
	assertSuppressionDotState(t, root, path, manager, seeds[:1], probes[1:], probes)
	for _, id := range []string{".", ".."} {
		candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
		body := suppressionReloadBody(t, candidate)
		if err := manager.Add(candidate); err != nil {
			t.Fatal(err)
		}
		assertSuppressionDotState(t, root, path, manager, []alert.Suppression{seeds[0], candidate}, []*model.Alert{probes[1], probes[3], probes[4]}, probes)
		updated := suppressionReloadRule(id, "candidate-rule", "198.51.100.0/24")
		updatedBody := suppressionReloadBody(t, updated)
		if err := manager.Update(id, updated); err != nil {
			t.Fatal(err)
		}
		assertSuppressionDotState(t, root, path, manager, []alert.Suppression{seeds[0], updated}, []*model.Alert{probes[1], probes[2], probes[4]}, probes)
		if err := manager.Delete(id); err != nil {
			t.Fatal(err)
		}
		assertSuppressionDotState(t, root, path, manager, seeds[:1], probes[1:], probes)
		if suppressionReloadBody(t, candidate) != body || suppressionReloadBody(t, updated) != updatedBody {
			t.Fatal("direct manager changed caller")
		}
	}
}
