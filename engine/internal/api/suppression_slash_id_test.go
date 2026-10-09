package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/pkg/model"
)

const suppressionSlashIDDetail = "id cannot contain /"

func suppressionSlashBodyWithoutID(t *testing.T, candidate alert.Suppression) string {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(suppressionReloadBody(t, candidate)), &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, "id")
	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"id":`) {
		t.Fatal("absent-ID fixture still emits an id member")
	}
	return string(data)
}

func suppressionSlashFixture(t *testing.T, fixture, id string) (string, string, *alert.SuppressionManager) {
	t.Helper()
	base := fixture
	if fixture == "existing-slash" {
		base = "healthy"
	}
	root, path, manager := suppressionReloadFixture(t, base)
	probes := suppressionReloadProbes()
	want := probes[1:]
	if fixture == "existing-slash" {
		if err := manager.Add(suppressionReloadRule(id, "legacy-rule", "198.51.100.0/24")); err != nil {
			t.Fatal(err)
		}
		want = probes[2:]
	}
	if !reflect.DeepEqual(manager.Filter(probes), want) {
		t.Fatal("initial suppression filter fixture invalid")
	}
	return root, path, manager
}

type suppressionSlashObservation struct {
	Tree             map[string]suppressionReloadArtifact
	Rules            []alert.Suppression
	Probes, Filtered string
}

func suppressionSlashSnapshot(t *testing.T, root string, manager *alert.SuppressionManager, probes []*model.Alert) suppressionSlashObservation {
	t.Helper()
	encode := func(alerts []*model.Alert) string {
		t.Helper()
		data, err := json.Marshal(alerts)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	return suppressionSlashObservation{Tree: suppressionReloadTree(t, root), Rules: manager.List(), Probes: encode(probes), Filtered: encode(manager.Filter(probes))}
}

func TestHTTPSuppressionCreateRejectsSlashIDsWithoutMutation(t *testing.T) {
	for _, id := range []string{"/", "/prior", "prior/", "/prior/", "prior/child", "/reload", "reload/"} {
		for _, fixture := range []string{"absent", "healthy", "existing-slash", "file-parent"} {
			for _, auth := range []bool{false, true} {
				for _, enabled := range []bool{false, true} {
					t.Run(fmt.Sprintf("id=%q/%s/auth=%v/enabled=%v", id, fixture, auth, enabled), func(t *testing.T) {
						root, _, manager := suppressionSlashFixture(t, fixture, id)
						candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
						candidate.Enabled = enabled
						body := suppressionReloadBody(t, candidate)
						probes := suppressionReloadProbes()
						before := suppressionSlashSnapshot(t, root, manager, probes)
						handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
						rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions", body, auth)
						assertSuppressionReloadError(t, rec, 400, "VALIDATION_ERROR", "Invalid suppression request", []string{suppressionSlashIDDetail})
						if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) || suppressionReloadBody(t, candidate) != body {
							t.Fatal("slash creation rejection changed tree, rules, filter or caller data")
						}
					})
				}
			}
		}
	}
}

func TestHTTPSuppressionSlashIDDiagnosticPrecedence(t *testing.T) {
	for _, name := range []string{"auth", "manager", "malformed", "unknown-field", "required-id", "reserved-id", "duplicate", "before-cidr", "before-compiler"} {
		t.Run(name, func(t *testing.T) {
			fixture := "healthy"
			if name == "duplicate" {
				fixture = "existing-slash"
			}
			root, _, manager := suppressionSlashFixture(t, fixture, "/prior")
			candidate := suppressionReloadRule("/prior", "candidate-rule", "203.0.113.0/24")
			opts := api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}
			status, code, message, details, authorized := 400, "VALIDATION_ERROR", "Invalid suppression request", []string{suppressionSlashIDDetail}, true
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
			case "required-id":
				candidate.ID, details = "", []string{"id is required"}
			case "reserved-id":
				candidate.ID, details = "reload", []string{suppressionReloadIDDetail}
			case "before-cidr":
				candidate.AnyCIDRs = nil
			case "before-compiler":
				candidate.AnyCIDRs = []string{"invalid-prefix"}
			}
			body := suppressionReloadBody(t, candidate)
			if name == "auth" || name == "manager" || name == "malformed" {
				body = `{"id":"/prior"`
			} else if name == "unknown-field" {
				body = `{"id":"/prior","unknown":true}`
			}
			probes := suppressionReloadProbes()
			before := suppressionSlashSnapshot(t, root, manager, probes)
			rec := suppressionReloadRequest(api.NewServerWithOptions(nil, nil, nil, nil, opts).Handler(), http.MethodPost, "/api/suppressions", body, authorized)
			assertSuppressionReloadError(t, rec, status, code, message, details)
			if name == "auth" && rec.Header().Get("WWW-Authenticate") != `Bearer realm="netsentry"` {
				t.Fatal("auth challenge changed")
			}
			if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
				t.Fatal("diagnostic control changed persistent or published state")
			}
		})
	}
}

func TestHTTPSuppressionSlashManagementPathsDoNotAlias(t *testing.T) {
	root, _, manager := suppressionSlashFixture(t, "existing-slash", "/prior/")
	probes := suppressionReloadProbes()
	before := suppressionSlashSnapshot(t, root, manager, probes)
	handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	candidate := suppressionReloadRule("", "candidate-rule", "203.0.113.0/24")
	body := suppressionSlashBodyWithoutID(t, candidate) // An absent identity would otherwise select the trimmed neighbor.
	for _, route := range []string{"prior/", "%2Fprior", "prior%2f", "%2fprior%2F", "prior%2Fchild", "%2F", "reload/", ""} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			for _, authorized := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/auth=%v", route, method, authorized), func(t *testing.T) {
					rec := suppressionReloadRequest(handler, method, "/api/suppressions/"+route, body, authorized)
					assertSuppressionReloadError(t, rec, 404, "NOT_FOUND", "Suppression not found", nil)
					if !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) || suppressionSlashBodyWithoutID(t, candidate) != body {
						t.Fatal("decoded slash path changed requested or neighboring identities")
					}
				})
			}
		}
	}
	// ServeMux can normalize raw paths before this handler; redirects are a separate boundary.
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		rec := suppressionReloadRequest(handler, method, "/api/suppressions//prior", body, true)
		if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != "/api/suppressions/prior" || !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
			t.Fatalf("raw cleanup = %d/%s, location %q", rec.Code, rec.Body.String(), rec.Header().Get("Location"))
		}
	}
}

func TestHTTPSuppressionEncodedSlashFreeIDsRemainManageable(t *testing.T) {
	for _, id := range []string{"ordinary", "prior%2Fchild", "prior?child", "prior#child", "Reload"} {
		t.Run(id, func(t *testing.T) {
			root, path, manager := suppressionSlashFixture(t, "healthy", "")
			prior := manager.List()[0]
			probes := suppressionReloadProbes()
			handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
			candidate := suppressionReloadRule(id, "candidate-rule", "203.0.113.0/24")
			rec := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions", suppressionReloadBody(t, candidate), true)
			var response alert.Suppression
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 201 || !reflect.DeepEqual(response, candidate) {
				t.Fatalf("create = %d/%s / %v", rec.Code, rec.Body.String(), err)
			}
			rules := []alert.Suppression{prior, candidate}
			assertSuppressionReloadState(t, manager, path, rules, []*model.Alert{probes[1], probes[3], probes[4]}, probes)
			list := suppressionReloadRequest(handler, http.MethodGet, "/api/suppressions", "", false)
			var listed struct{ Data []alert.Suppression }
			if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || list.Code != 200 || !reflect.DeepEqual(listed.Data, rules) {
				t.Fatalf("list = %d/%s / %v", list.Code, list.Body.String(), err)
			}
			endpoint := "/api/suppressions/" + url.PathEscape(id)
			updated := suppressionReloadRule(id, "candidate-rule", "198.51.100.0/24")
			body := updated
			body.ID = ""
			rec = suppressionReloadRequest(handler, http.MethodPut, endpoint, suppressionSlashBodyWithoutID(t, body), true)
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || !reflect.DeepEqual(response, updated) || body.ID != "" {
				t.Fatalf("update = %d/%s / %v", rec.Code, rec.Body.String(), err)
			}
			rules = []alert.Suppression{prior, updated}
			want := []*model.Alert{probes[1], probes[2], probes[4]}
			assertSuppressionReloadState(t, manager, path, rules, want, probes)
			beforeReload := suppressionReloadTree(t, root)
			reloaded := suppressionReloadRequest(handler, http.MethodPost, "/api/suppressions/reload", "", true)
			var reload struct{ Reloaded int }
			if err := json.Unmarshal(reloaded.Body.Bytes(), &reload); err != nil || reloaded.Code != 200 || reload.Reloaded != 2 || !reflect.DeepEqual(suppressionReloadTree(t, root), beforeReload) {
				t.Fatalf("reload = %d/%s / %v", reloaded.Code, reloaded.Body.String(), err)
			}
			assertSuppressionReloadState(t, manager, path, rules, want, probes)
			deleted := suppressionReloadRequest(handler, http.MethodDelete, endpoint, "", true)
			if deleted.Code != 204 || deleted.Body.Len() != 0 {
				t.Fatalf("delete = %d/%s", deleted.Code, deleted.Body.String())
			}
			assertSuppressionReloadState(t, manager, path, []alert.Suppression{prior}, probes[1:], probes)
		})
	}
}

func TestHTTPFileLoadedSlashSuppressionIDsRetainManagerCompatibility(t *testing.T) {
	root, path, manager := suppressionSlashFixture(t, "existing-slash", "/prior/")
	seeds, err := alert.LoadSuppressionsFromFile(path)
	if err != nil || !reflect.DeepEqual(seeds, manager.List()) {
		t.Fatalf("legacy load = %+v / %v", seeds, err)
	}
	probes := suppressionReloadProbes()
	assertSuppressionReloadState(t, manager, path, seeds, probes[2:], probes)
	before := suppressionSlashSnapshot(t, root, manager, probes)
	handler := api.NewServerWithOptions(nil, nil, nil, nil, api.Options{Suppressions: manager, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	for _, endpoint := range []string{"/api/suppressions/reload", "/api/suppressions/%72eload"} {
		rec := suppressionReloadRequest(handler, http.MethodPost, endpoint, "", true)
		var response struct{ Reloaded int }
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || response.Reloaded != 2 || !reflect.DeepEqual(suppressionSlashSnapshot(t, root, manager, probes), before) {
			t.Fatalf("legacy reload = %d/%s / %v", rec.Code, rec.Body.String(), err)
		}
		assertSuppressionReloadState(t, manager, path, seeds, probes[2:], probes)
	}
	updated := suppressionReloadRule("/prior/", "candidate-rule", "203.0.113.0/24")
	if err := manager.Update("/prior/", updated); err != nil {
		t.Fatal(err)
	}
	assertSuppressionReloadState(t, manager, path, []alert.Suppression{seeds[0], updated}, []*model.Alert{probes[1], probes[3], probes[4]}, probes)
	if err := manager.Delete("/prior/"); err != nil {
		t.Fatal(err)
	}
	assertSuppressionReloadState(t, manager, path, seeds[:1], probes[1:], probes)
	if err := manager.Add(updated); err != nil {
		t.Fatal(err)
	}
	assertSuppressionReloadState(t, manager, path, []alert.Suppression{seeds[0], updated}, []*model.Alert{probes[1], probes[3], probes[4]}, probes)
}
