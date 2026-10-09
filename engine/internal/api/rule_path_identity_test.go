package api_test

import (
	"bytes"
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

func rulePathBodyWithoutID(t *testing.T, candidate *model.Rule) string {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(reloadIDBody(t, candidate)), &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, "id")
	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte(`"id":`)) {
		t.Fatal("absent-ID fixture still emits an id member")
	}
	return string(data)
}

func rulePathFixture(t *testing.T, fixture string) (string, string, *rule.Engine) {
	t.Helper()
	base := fixture
	if fixture == "existing-slash" {
		base = "healthy"
	}
	root, path, engine := reloadIDFixture(t, base)
	if fixture == "existing-slash" {
		seeds := append(engine.Rules(), reloadIDRule("/prior/", 81))
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

func assertRulePathPreserved(t *testing.T, root string, engine *rule.Engine, tree map[string]reloadIDArtifact, want []*model.Rule) {
	t.Helper()
	if !reflect.DeepEqual(reloadIDTree(t, root), tree) || engine.RuleCount() != len(want) || !reflect.DeepEqual(engine.Rules(), want) {
		t.Fatal("rejected path changed tree membership/modes/bytes or full rule snapshot/count")
	}
	for _, port := range []uint16{80, 81, 82} {
		var expected []*model.Alert
		for _, candidate := range want {
			var cfg model.PortBlacklistConfig
			if err := json.Unmarshal(candidate.Config, &cfg); err != nil {
				t.Fatal(err)
			}
			for _, p := range cfg.Ports {
				if candidate.Enabled && p == int(port) {
					expected = append(expected, &model.Alert{RuleID: candidate.ID, RuleName: candidate.Name, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: port, Protocol: "TCP", Severity: candidate.Severity, MatchedKeyword: fmt.Sprintf("port_blacklist: %d", port)})
				}
			}
		}
		packet := reloadIDPacket(port)
		before := *packet
		if got := engine.Match(packet); !reflect.DeepEqual(got, expected) || *packet != before {
			t.Fatalf("port %d matching/caller changed: got %+v, want %+v", port, got, expected)
		}
	}
}

func TestHTTPRuleSlashManagementPathsDoNotAlias(t *testing.T) {
	paths := []string{"/api/rules/prior/", "/api/rules/%2Fprior", "/api/rules/prior%2F", "/api/rules/%2Fprior%2F", "/api/rules/prior%2Fchild", "/api/rules/%2F", "/api/rules/%2F%2F", "/api/rules/%2Fprior/"}
	for _, fixture := range []string{"absent", "healthy", "existing-slash", "file-parent"} {
		for _, auth := range []bool{false, true} {
			for _, method := range []string{http.MethodPut, http.MethodDelete} {
				for _, target := range paths {
					t.Run(fmt.Sprintf("%s/auth=%v/%s/%s", fixture, auth, method, target), func(t *testing.T) {
						root, path, engine := rulePathFixture(t, fixture)
						candidate := reloadIDRule("/prior/", 82)
						caller := reloadIDBody(t, candidate)
						body := rulePathBodyWithoutID(t, candidate)
						before, prior := reloadIDTree(t, root), engine.Rules()
						handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
						rec := reloadIDRequest(handler, method, target, body, auth)
						assertReloadIDError(t, rec, 404, "NOT_FOUND", "Rule not found", nil)
						assertRulePathPreserved(t, root, engine, before, prior)
						if reloadIDBody(t, candidate) != caller {
							t.Fatal("path rejection modified request candidate")
						}
					})
				}
			}
		}
	}
}

func TestHTTPRulePathIdentityDiagnosticsAndRawRedirects(t *testing.T) {
	for _, name := range []string{"empty-id", "slash-before-auth", "auth", "file", "malformed", "unknown-field", "mismatch", "unknown-put", "unknown-delete", "method", "raw-leading", "raw-middle"} {
		t.Run(name, func(t *testing.T) {
			root, path, engine := rulePathFixture(t, "existing-slash")
			opts := api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}
			method, target, authorized := http.MethodPut, "/api/rules/prior", true
			body := rulePathBodyWithoutID(t, reloadIDRule("prior", 82))
			status, code, message, details := 404, "NOT_FOUND", "Rule not found", []string(nil)
			location := ""
			switch name {
			case "empty-id":
				target = "/api/rules/"
			case "slash-before-auth":
				target, authorized, body = "/api/rules/%2Fprior%2F", false, "{malformed"
			case "auth":
				authorized, body = false, "{malformed"
				status, code, message = 401, "UNAUTHORIZED", "Valid bearer token required"
			case "file":
				opts.RulesSeedFile, body = "", "{malformed"
				status, code, message = 409, "RULES_WRITE_UNAVAILABLE", "Rules seed file is not configured"
			case "malformed":
				body = "{"
				status, code, message, details = 400, "VALIDATION_ERROR", "Invalid rule request", []string{"unexpected EOF"}
			case "unknown-field":
				body = `{"unknown":true}`
				status, code, message, details = 400, "VALIDATION_ERROR", "Invalid rule request", []string{`json: unknown field "unknown"`}
			case "mismatch":
				body = reloadIDBody(t, reloadIDRule("other", 82))
				status, code, message = 400, "VALIDATION_ERROR", "Rule ID in path and body must match"
			case "unknown-put":
				target = "/api/rules/missing"
			case "unknown-delete":
				method, target = http.MethodDelete, "/api/rules/missing"
			case "method":
				method = http.MethodGet
				status, code, message = 405, "METHOD_NOT_ALLOWED", "Method not allowed"
			case "raw-leading":
				target, location = "/api/rules//prior", "/api/rules/prior"
			case "raw-middle":
				method, target, location = http.MethodDelete, "/api/rules/prior//child", "/api/rules/prior/child"
			}
			before, prior := reloadIDTree(t, root), engine.Rules()
			rec := reloadIDRequest(api.NewServerWithOptions(nil, nil, engine, nil, opts).Handler(), method, target, body, authorized)
			if location != "" {
				// Pinned Go 1.26.8 ServeMux cleans raw escaped paths before handler entry.
				if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != location {
					t.Fatalf("raw cleanup = %d/%q, want 307/%q", rec.Code, rec.Header().Get("Location"), location)
				}
			} else {
				assertReloadIDError(t, rec, status, code, message, details)
				if name == "method" && rec.Header().Get("Allow") != "PUT, DELETE" {
					t.Fatal("ordinary rule method policy changed")
				}
			}
			assertRulePathPreserved(t, root, engine, before, prior)
		})
	}
}

func TestHTTPRuleEncodedSlashFreeIDsRemainManageable(t *testing.T) {
	for _, id := range []string{"path-ordinary", "prior%2F", "Prior", "PRIOR", "prior-extra"} {
		t.Run(id, func(t *testing.T) {
			_, path, engine := rulePathFixture(t, "healthy")
			prior := engine.Rules()[0]
			handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			candidate := reloadIDRule(id, 81)
			rec := reloadIDRequest(handler, http.MethodPost, "/api/rules", reloadIDBody(t, candidate), true)
			var created model.Rule
			if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || rec.Code != 201 {
				t.Fatalf("create = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
			}
			assertReloadIDRules(t, []*model.Rule{&created}, []*model.Rule{candidate})
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior, candidate})
			list := reloadIDRequest(handler, http.MethodGet, "/api/rules", "", false)
			var listed struct{ Data []*model.Rule }
			if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || list.Code != 200 {
				t.Fatalf("list = %d/%s, error = %v", list.Code, list.Body.String(), err)
			}
			assertReloadIDRules(t, listed.Data, []*model.Rule{prior, candidate})
			updated := reloadIDRule(id, 82)
			updated.Name = "updated " + id
			caller := reloadIDBody(t, updated)
			target := "/api/rules/" + url.PathEscape(id)
			if id == "path-ordinary" {
				target = "/api/rules/%70ath-ordinary"
			}
			rec = reloadIDRequest(handler, http.MethodPut, target, rulePathBodyWithoutID(t, updated), true)
			var response model.Rule
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || reloadIDBody(t, updated) != caller {
				t.Fatalf("update = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
			}
			assertReloadIDRules(t, []*model.Rule{&response}, []*model.Rule{updated})
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior, updated})
			reloaded := reloadIDRequest(handler, http.MethodPost, "/api/rules/reload", "", true)
			var reload struct{ Reloaded int }
			if err := json.Unmarshal(reloaded.Body.Bytes(), &reload); err != nil || reloaded.Code != 200 || reload.Reloaded != 2 {
				t.Fatalf("reload = %d/%s, error = %v", reloaded.Code, reloaded.Body.String(), err)
			}
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior, updated})
			deleted := reloadIDRequest(handler, http.MethodDelete, target, "", true)
			if deleted.Code != 204 || deleted.Body.Len() != 0 {
				t.Fatalf("delete = %d/%s", deleted.Code, deleted.Body.String())
			}
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior})
		})
	}
}

func TestHTTPFileLoadedSlashRuleIDsRetainEngineCompatibility(t *testing.T) {
	root, path, engine := rulePathFixture(t, "existing-slash")
	prior := engine.Rules()
	caller := reloadIDBody(t, prior[1])
	before := reloadIDTree(t, root)
	handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	for _, seeds := range [][]*model.Rule{prior, {prior[0], reloadIDRule("/prior/", 82)}} {
		if err := rule.SaveToFile(path, seeds); err != nil {
			t.Fatal(err)
		}
		beforeReload := reloadIDTree(t, root)
		rec := reloadIDRequest(handler, http.MethodPost, "/api/rules/reload", "", true)
		var response struct{ Reloaded int }
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || response.Reloaded != 2 {
			t.Fatalf("legacy reload = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
		}
		assertReloadIDPublishedAndPersisted(t, engine, path, seeds)
		assertRulePathPreserved(t, root, engine, beforeReload, engine.Rules())
	}
	if reloadIDBody(t, prior[1]) != caller || reflect.DeepEqual(reloadIDTree(t, root), before) {
		t.Fatal("legacy file edit failed to change persisted identity data or modified caller")
	}
}
