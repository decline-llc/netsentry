package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/pkg/model"
)

func ruleDotFixture(t *testing.T, fixture string) (string, string, *rule.Engine) {
	t.Helper()
	base := fixture
	if fixture == "legacy-dots" {
		base = "healthy"
	}
	root, path, engine := reloadIDFixture(t, base)
	if fixture == "legacy-dots" {
		dot := reloadIDRule(".", 81)
		dot.Priority = 110
		seeds := append(engine.Rules(), dot, reloadIDRule("..", 82))
		if err := rule.SaveToFile(path, seeds); err != nil {
			t.Fatal(err)
		}
		loaded, err := rule.LoadFromFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := engine.Reload(loaded); err != nil {
			t.Fatal(err)
		}
		assertReloadIDPublishedAndPersisted(t, engine, path, seeds)
	}
	return root, path, engine
}

func TestHTTPRuleCreateRejectsDotIDsWithoutMutation(t *testing.T) {
	for _, id := range []string{".", ".."} {
		for _, fixture := range []string{"absent", "healthy", "legacy-dots", "file-parent"} {
			for _, enabled := range []bool{false, true} {
				for _, auth := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/enabled=%v/auth=%v", id, fixture, enabled, auth), func(t *testing.T) {
						root, path, engine := ruleDotFixture(t, fixture)
						candidate := reloadIDRule(id, 82)
						candidate.Enabled = enabled
						body := reloadIDBody(t, candidate)
						before, prior := reloadIDTree(t, root), engine.Rules()
						handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
						rec := reloadIDRequest(handler, http.MethodPost, "/api/rules", body, auth)
						assertReloadIDError(t, rec, 400, "VALIDATION_ERROR", "Invalid rule request", []string{"id cannot be . or .."})
						assertRulePathPreserved(t, root, engine, before, prior)
						if reloadIDBody(t, candidate) != body {
							t.Fatal("dot rejection changed caller data")
						}
					})
				}
			}
		}
	}
}

func TestHTTPRuleDotManagementAndRawRedirectBoundaries(t *testing.T) {
	for _, target := range []string{"/api/rules/%2e", "/api/rules/%2E", "/api/rules/%2e%2e", "/api/rules/%2E%2E"} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			for _, mode := range []string{"no-auth", "authorized", "missing-auth"} {
				t.Run(target+"/"+method+"/"+mode, func(t *testing.T) {
					root, path, engine := ruleDotFixture(t, "legacy-dots")
					before, prior := reloadIDTree(t, root), engine.Rules()
					candidate := reloadIDRule(".", 82)
					body := rulePathBodyWithoutID(t, candidate)
					if mode == "missing-auth" {
						body = "{malformed"
					}
					handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: mode != "no-auth", AuthToken: "fixture-auth-value"}).Handler()
					rec := reloadIDRequest(handler, method, target, body, mode == "authorized")
					assertReloadIDError(t, rec, 404, "NOT_FOUND", "Rule not found", nil)
					assertRulePathPreserved(t, root, engine, before, prior)
				})
			}
		}
	}
	for _, id := range []string{".", ".."} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			t.Run("raw/"+id+"/"+method, func(t *testing.T) {
				root, path, engine := ruleDotFixture(t, "legacy-dots")
				before, prior := reloadIDTree(t, root), engine.Rules()
				handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
				rec := reloadIDRequest(handler, method, "/api/rules/"+id, "{malformed", false)
				location := "/api/rules"
				if id == ".." {
					location = "/api"
				}
				if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != location {
					t.Fatalf("raw dot redirect = %d/%q, want 307/%q", rec.Code, rec.Header().Get("Location"), location)
				}
				assertRulePathPreserved(t, root, engine, before, prior)
			})
		}
	}
}

func TestHTTPRuleDotCreationDiagnosticPrecedence(t *testing.T) {
	for _, name := range []string{"auth", "file", "malformed", "unknown-field", "slash", "reload"} {
		t.Run(name, func(t *testing.T) {
			root, path, engine := ruleDotFixture(t, "legacy-dots")
			before, prior := reloadIDTree(t, root), engine.Rules()
			opts := api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}
			body, authorized := reloadIDBody(t, reloadIDRule(".", 82)), true
			status, code, message := 400, "VALIDATION_ERROR", "Invalid rule request"
			var details []string
			switch name {
			case "auth":
				authorized = false
				status, code, message = 401, "UNAUTHORIZED", "Valid bearer token required"
			case "file":
				opts.RulesSeedFile = ""
				status, code, message = 409, "RULES_WRITE_UNAVAILABLE", "Rules seed file is not configured"
			case "malformed":
				body, details = "{", []string{"unexpected EOF"}
			case "unknown-field":
				body, details = `{"id":".","unknown":true}`, []string{`json: unknown field "unknown"`}
			case "slash":
				body, details = reloadIDBody(t, reloadIDRule("./", 82)), []string{"id cannot contain /, ?, or #"}
			case "reload":
				body, details = reloadIDBody(t, reloadIDRule("reload", 82)), []string{reloadIDDetail}
			}
			handler := api.NewServerWithOptions(nil, nil, engine, nil, opts).Handler()
			rec := reloadIDRequest(handler, http.MethodPost, "/api/rules", body, authorized)
			assertReloadIDError(t, rec, status, code, message, details)
			assertRulePathPreserved(t, root, engine, before, prior)
		})
	}
}

func TestHTTPRuleOtherDotAndPercentIDsRetainExactCRUD(t *testing.T) {
	for _, id := range []string{"...", ".prior", "prior.", "prior..id", "%2e", "%2E"} {
		t.Run(id, func(t *testing.T) {
			_, path, engine := ruleDotFixture(t, "healthy")
			prior := engine.Rules()[0]
			handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			candidate := reloadIDRule(id, 81)
			body := reloadIDBody(t, candidate)
			rec := reloadIDRequest(handler, http.MethodPost, "/api/rules", body, true)
			var response model.Rule
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 201 || reloadIDBody(t, candidate) != body {
				t.Fatalf("create = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
			}
			assertReloadIDRules(t, []*model.Rule{&response}, []*model.Rule{candidate})
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior, candidate})
			list := reloadIDRequest(handler, http.MethodGet, "/api/rules", "", false)
			var listed struct{ Data []*model.Rule }
			if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || list.Code != 200 {
				t.Fatalf("list = %d/%s, error = %v", list.Code, list.Body.String(), err)
			}
			assertReloadIDRules(t, listed.Data, []*model.Rule{prior, candidate})
			updated := reloadIDRule(id, 82)
			updated.Name = "updated " + id
			body = reloadIDBody(t, updated)
			target := "/api/rules/" + url.PathEscape(id)
			rec = reloadIDRequest(handler, http.MethodPut, target, rulePathBodyWithoutID(t, updated), true)
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || reloadIDBody(t, updated) != body {
				t.Fatalf("update = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
			}
			assertReloadIDRules(t, []*model.Rule{&response}, []*model.Rule{updated})
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior, updated})
			assertRuleDotReload(t, handler, 2)
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior, updated})
			rec = reloadIDRequest(handler, http.MethodDelete, target, "", true)
			if rec.Code != 204 || rec.Body.Len() != 0 {
				t.Fatalf("delete = %d/%s", rec.Code, rec.Body.String())
			}
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior})
		})
	}
}

func assertRuleDotReload(t *testing.T, handler http.Handler, count int) {
	t.Helper()
	rec := reloadIDRequest(handler, http.MethodPost, "/api/rules/reload", "", true)
	var response struct{ Reloaded int }
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || response.Reloaded != count {
		t.Fatalf("reload = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
	}
}

func TestHTTPFileLoadedDotRuleIDsRetainCompatibility(t *testing.T) {
	root, path, engine := ruleDotFixture(t, "legacy-dots")
	prior := engine.Rules()
	caller := reloadIDBody(t, prior[1]) + reloadIDBody(t, prior[2])
	handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	for _, edit := range []bool{false, true} {
		seeds := engine.Rules()
		if edit {
			seeds[1].Config, seeds[2].Config = seeds[2].Config, seeds[1].Config
		}
		if err := rule.SaveToFile(path, seeds); err != nil {
			t.Fatal(err)
		}
		before := reloadIDTree(t, root)
		assertRuleDotReload(t, handler, 3)
		assertReloadIDPublishedAndPersisted(t, engine, path, seeds)
		assertRulePathPreserved(t, root, engine, before, engine.Rules())
	}
	if reloadIDBody(t, prior[1])+reloadIDBody(t, prior[2]) != caller || reflect.DeepEqual(engine.Rules(), prior) {
		t.Fatal("legacy file edit did not change rules or changed callers")
	}
}
