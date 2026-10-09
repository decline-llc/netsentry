package api_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/pkg/model"
)

const suppressionReloadIDDetail = `id "reload" is reserved for the suppressions reload endpoint`

func suppressionReloadRule(id, ruleID, cidr string) alert.Suppression {
	return alert.Suppression{ID: id, Enabled: true, RuleIDs: []string{ruleID}, AnyCIDRs: []string{cidr}}
}

func suppressionReloadBody(t *testing.T, candidate alert.Suppression) string {
	t.Helper()
	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func suppressionReloadRequest(handler http.Handler, method, endpoint, body string, authorized bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, endpoint, strings.NewReader(body))
	req.Header.Set("X-Request-ID", "suppression-reload-id-fixture")
	if authorized {
		req.Header.Set("Authorization", "Bearer fixture-auth-value")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func assertSuppressionReloadError(t *testing.T, rec *httptest.ResponseRecorder, status int, code, message string, details []string) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code, Message string
			Details       []string
			RequestID     string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if rec.Code != status || rec.Header().Get("Content-Type") != "application/json" || envelope.Error.Code != code || envelope.Error.Message != message || envelope.Error.RequestID != "suppression-reload-id-fixture" || !reflect.DeepEqual(envelope.Error.Details, details) {
		t.Fatalf("error = %d/%s, want %d/%s/%s/%v", rec.Code, rec.Body.String(), status, code, message, details)
	}
}

type suppressionReloadArtifact struct {
	Mode os.FileMode
	Data string
}

func suppressionReloadTree(t *testing.T, root string) map[string]suppressionReloadArtifact {
	t.Helper()
	result := make(map[string]suppressionReloadArtifact)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		artifact := suppressionReloadArtifact{Mode: info.Mode()}
		if !entry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			artifact.Data = string(data)
		}
		result[name] = artifact
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func suppressionReloadFixture(t *testing.T, fixture string) (string, string, *alert.SuppressionManager) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "seed directory with spaces", "suppressions.json")
	seeds := []alert.Suppression{suppressionReloadRule("prior", "old-rule", "192.0.2.0/24")}
	if fixture == "existing-reserved" {
		seeds = append(seeds, suppressionReloadRule("reload", "legacy-rule", "198.51.100.0/24"))
	}
	if fixture == "file-parent" {
		if err := os.WriteFile(filepath.Dir(path), []byte("retained parent occupant"), 0o600); err != nil {
			t.Fatal(err)
		}
	} else if fixture != "absent" {
		if err := alert.SaveSuppressionsToFile(path, seeds); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manager, err := alert.NewSuppressionManagerWithFile(seeds, path)
	if err != nil {
		t.Fatal(err)
	}
	return root, path, manager
}

func suppressionReloadProbes() []*model.Alert {
	return []*model.Alert{
		{RuleID: "old-rule", SrcIP: "192.0.2.9"},
		{RuleID: "legacy-rule", SrcIP: "198.51.100.9"},
		{RuleID: "candidate-rule", SrcIP: "203.0.113.9"},
		{RuleID: "candidate-rule", SrcIP: "198.51.100.9"},
		{RuleID: "other-rule", SrcIP: "203.0.113.9"},
	}
}

func TestHTTPSuppressionCreateRejectsReservedReloadIDWithoutMutation(t *testing.T) {
	for _, fixture := range []string{"absent", "healthy", "existing-reserved", "file-parent"} {
		for _, auth := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/auth=%v/enabled=%v", fixture, auth, enabled), func(t *testing.T) {
					root, _, manager := suppressionReloadFixture(t, fixture)
					candidate := suppressionReloadRule("reload", "candidate-rule", "203.0.113.0/24")
					candidate.Enabled = enabled
					beforeBody := suppressionReloadBody(t, candidate)
					beforeTree, beforeList := suppressionReloadTree(t, root), manager.List()
					probes := suppressionReloadProbes()
					beforeProbes := make([]model.Alert, len(probes))
					for i, probe := range probes {
						beforeProbes[i] = *probe
					}
					want := probes[1:]
					if fixture == "existing-reserved" {
						want = probes[2:]
					}
					if !reflect.DeepEqual(manager.Filter(probes), want) {
						t.Fatal("initial suppression filter fixture invalid")
					}
					handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
					rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions", beforeBody, auth)
					assertSuppressionReloadError(t, rec, 400, "VALIDATION_ERROR", "Invalid suppression request", []string{suppressionReloadIDDetail})
					if !reflect.DeepEqual(suppressionReloadTree(t, root), beforeTree) || !reflect.DeepEqual(manager.List(), beforeList) || !reflect.DeepEqual(manager.Filter(probes), want) || suppressionReloadBody(t, candidate) != beforeBody {
						t.Fatal("reserved-ID rejection changed tree, List, Filter or input")
					}
					for i, probe := range probes {
						if !reflect.DeepEqual(*probe, beforeProbes[i]) {
							t.Fatal("filter mutated caller alert")
						}
					}
				})
			}
		}
	}
}

func TestHTTPSuppressionReloadIDDiagnosticPrecedence(t *testing.T) {
	for _, name := range []string{"auth", "manager", "malformed", "unknown-field", "duplicate", "before-cidr", "before-compiler"} {
		t.Run(name, func(t *testing.T) {
			fixture := "healthy"
			if name == "duplicate" {
				fixture = "existing-reserved"
			}
			root, _, manager := suppressionReloadFixture(t, fixture)
			candidate := suppressionReloadRule("reload", "candidate-rule", "203.0.113.0/24")
			opts := api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}
			status, code, message, details, authorized := 400, "VALIDATION_ERROR", "Invalid suppression request", []string{suppressionReloadIDDetail}, true
			switch name {
			case "auth":
				opts.Suppressions = nil
				authorized, status, code, message, details = false, 401, "UNAUTHORIZED", "Valid bearer token required", nil
			case "manager":
				opts.Suppressions = nil
				status, code, message, details = 409, "SUPPRESSIONS_UNAVAILABLE", "Suppressions manager is not configured", nil
			case "malformed":
				details = []string{"unexpected EOF"}
			case "unknown-field":
				details = []string{`json: unknown field "unknown"`}
			case "before-cidr":
				candidate.AnyCIDRs = nil
			case "before-compiler":
				candidate.AnyCIDRs = []string{"invalid-prefix"}
			}
			body := suppressionReloadBody(t, candidate)
			if name == "auth" || name == "manager" || name == "malformed" {
				body = `{"id":"reload"`
			} else if name == "unknown-field" {
				body = `{"id":"reload","unknown":true}`
			}
			beforeTree, beforeList := suppressionReloadTree(t, root), manager.List()
			probes := suppressionReloadProbes()
			beforeFilter := manager.Filter(probes)
			rec := suppressionReloadRequest(api.NewServerWithOptions(nil, nil, nil, nil, opts).Handler(), http.MethodPost, "/api/suppressions", body, authorized)
			assertSuppressionReloadError(t, rec, status, code, message, details)
			if name == "auth" && rec.Header().Get("WWW-Authenticate") != `Bearer realm="netsentry"` {
				t.Fatal("auth challenge changed")
			}
			if !reflect.DeepEqual(suppressionReloadTree(t, root), beforeTree) || !reflect.DeepEqual(manager.List(), beforeList) || !reflect.DeepEqual(manager.Filter(probes), beforeFilter) {
				t.Fatal("diagnostic control changed persistent or published state")
			}
		})
	}
}

func assertSuppressionReloadState(t *testing.T, manager *alert.SuppressionManager, path string, rules []alert.Suppression, want []*model.Alert, probes []*model.Alert) {
	t.Helper()
	loaded, err := alert.LoadSuppressionsFromFile(path)
	if err != nil || !reflect.DeepEqual(loaded, rules) || !reflect.DeepEqual(manager.List(), rules) || !reflect.DeepEqual(manager.Filter(probes), want) {
		t.Fatalf("state = loaded %+v / list %+v / filter %+v / error %v", loaded, manager.List(), manager.Filter(probes), err)
	}
	observed, err := alert.NewSuppressionManager(loaded)
	if err != nil || !reflect.DeepEqual(observed.List(), rules) || !reflect.DeepEqual(observed.Filter(probes), want) {
		t.Fatalf("independent manager = %v / %v", observed, err)
	}
}

func TestHTTPSuppressionReloadCaseVariantIDsRemainManageable(t *testing.T) {
	for _, item := range []struct{ id, route string }{{"Reload", "%52eload"}, {"RELOAD", "%52ELOAD"}, {"reload-extra", "%72eload-extra"}} {
		t.Run(item.id, func(t *testing.T) {
			_, path, manager := suppressionReloadFixture(t, "healthy")
			prior := manager.List()[0]
			probes := suppressionReloadProbes()
			handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			candidate := suppressionReloadRule(item.id, "candidate-rule", "203.0.113.0/24")
			rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions", suppressionReloadBody(t, candidate), true)
			var response alert.Suppression
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 201 || !reflect.DeepEqual(response, candidate) {
				t.Fatalf("create = %d/%s / %v", rec.Code, rec.Body.String(), err)
			}
			rules, want := []alert.Suppression{prior, candidate}, []*model.Alert{probes[1], probes[3], probes[4]}
			assertSuppressionReloadState(t, manager, path, rules, want, probes)
			list := suppressionReloadRequest(handler, http.MethodGet, "/api/suppressions", "", false)
			var listed struct{ Data []alert.Suppression }
			if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || list.Code != 200 || !reflect.DeepEqual(listed.Data, rules) {
				t.Fatalf("list = %d/%s / %v", list.Code, list.Body.String(), err)
			}
			updated := suppressionReloadRule(item.id, "candidate-rule", "198.51.100.0/24")
			body := updated
			body.ID = ""
			for _, route := range []string{item.id, item.route} {
				rec = suppressionReloadRequest(handler, http.MethodPut, "/api/suppressions/"+route, suppressionReloadBody(t, body), true)
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || !reflect.DeepEqual(response, updated) || body.ID != "" {
					t.Fatalf("update = %d/%s / %v", rec.Code, rec.Body.String(), err)
				}
				rules, want = []alert.Suppression{prior, updated}, []*model.Alert{probes[1], probes[2], probes[4]}
				assertSuppressionReloadState(t, manager, path, rules, want, probes)
			}
			beforeReload := suppressionReloadTree(t, filepath.Dir(path))
			reloaded := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions/reload", "", true)
			var reload struct{ Reloaded int }
			if err := json.Unmarshal(reloaded.Body.Bytes(), &reload); err != nil || reloaded.Code != 200 || reload.Reloaded != 2 || !reflect.DeepEqual(suppressionReloadTree(t, filepath.Dir(path)), beforeReload) {
				t.Fatalf("reload = %d/%s / %v", reloaded.Code, reloaded.Body.String(), err)
			}
			assertSuppressionReloadState(t, manager, path, rules, want, probes)
			deleted := suppressionReloadRequest(handler, http.MethodDelete, "/api/suppressions/"+item.route, "", true)
			if deleted.Code != 204 || deleted.Body.Len() != 0 {
				t.Fatalf("delete = %d/%s", deleted.Code, deleted.Body.String())
			}
			assertSuppressionReloadState(t, manager, path, []alert.Suppression{prior}, probes[1:], probes)
		})
	}
}

func TestHTTPFileLoadedSuppressionReloadIDRetainsCompatibility(t *testing.T) {
	root, path, manager := suppressionReloadFixture(t, "existing-reserved")
	seeds, err := alert.LoadSuppressionsFromFile(path)
	if err != nil || !reflect.DeepEqual(seeds, manager.List()) {
		t.Fatalf("legacy load = %+v / %v", seeds, err)
	}
	probes := suppressionReloadProbes()
	assertSuppressionReloadState(t, manager, path, seeds, probes[2:], probes)
	before := suppressionReloadTree(t, root)
	handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	for _, endpoint := range []string{"/api/suppressions/reload", "/api/suppressions/%72eload"} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			rec := suppressionReloadRequest(handler, method, endpoint, suppressionReloadBody(t, suppressionReloadRule("reload", "candidate-rule", "203.0.113.0/24")), true)
			assertSuppressionReloadError(t, rec, 405, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
			if rec.Header().Get("Allow") != "POST" {
				t.Fatalf("Allow = %q", rec.Header().Get("Allow"))
			}
			assertSuppressionReloadState(t, manager, path, seeds, probes[2:], probes)
			if !reflect.DeepEqual(suppressionReloadTree(t, root), before) {
				t.Fatal("legacy rejected HTTP management changed tree")
			}
		}
		rec := suppressionReloadRequest(handler, http.MethodPost, endpoint, "", true)
		var response struct{ Reloaded int }
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || response.Reloaded != 2 {
			t.Fatalf("legacy reload = %d/%s / %v", rec.Code, rec.Body.String(), err)
		}
		assertSuppressionReloadState(t, manager, path, seeds, probes[2:], probes)
		if !reflect.DeepEqual(suppressionReloadTree(t, root), before) {
			t.Fatal("legacy reload changed tree")
		}
	}
	// HTTP reservation does not migrate the public manager/file identity policy.
	if err := manager.Delete("reload"); err != nil {
		t.Fatal(err)
	}
	assertSuppressionReloadState(t, manager, path, seeds[:1], probes[1:], probes)
	candidate := suppressionReloadRule("reload", "candidate-rule", "203.0.113.0/24")
	if err := manager.Add(candidate); err != nil {
		t.Fatal(err)
	}
	assertSuppressionReloadState(t, manager, path, []alert.Suppression{seeds[0], candidate}, []*model.Alert{probes[1], probes[3], probes[4]}, probes)
	candidate.AnyCIDRs = []string{"198.51.100.0/24"}
	if err := manager.Update("reload", candidate); err != nil {
		t.Fatal(err)
	}
	assertSuppressionReloadState(t, manager, path, []alert.Suppression{seeds[0], candidate}, []*model.Alert{probes[1], probes[2], probes[4]}, probes)
}
