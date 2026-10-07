package api_test

import (
	"bytes"
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

	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/internal/rule"
	"github.com/decline-llc/netsentry/pkg/model"
)

const reloadIDDetail = `id "reload" is reserved for the rules reload endpoint`

func reloadIDRule(id string, port int) *model.Rule {
	return &model.Rule{ID: id, Name: "fixture " + id, Type: model.RuleTypePortBlacklist, Severity: model.SeverityHigh, Priority: 100, Enabled: true, Config: json.RawMessage(fmt.Sprintf(`{"ports":[%d]}`, port))}
}

func reloadIDPacket(port uint16) *model.PacketInfo {
	return &model.PacketInfo{SrcIP: "192.0.2.1", DstIP: "198.51.100.1", SrcPort: 40000, DstPort: port, Protocol: 6}
}

func reloadIDRequest(handler http.Handler, method, path, body string, authorized bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("X-Request-ID", "reload-id-fixture")
	if authorized {
		req.Header.Set("Authorization", "Bearer fixture-auth-value")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func reloadIDBody(t *testing.T, candidate *model.Rule) string {
	t.Helper()
	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertReloadIDError(t *testing.T, rec *httptest.ResponseRecorder, status int, code, message string, details []string) {
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
	if rec.Code != status || rec.Header().Get("Content-Type") != "application/json" || envelope.Error.Code != code || envelope.Error.Message != message || envelope.Error.RequestID != "reload-id-fixture" || !reflect.DeepEqual(envelope.Error.Details, details) {
		t.Fatalf("error envelope = %d/%s, want %d/%s/%s/%v", rec.Code, rec.Body.String(), status, code, message, details)
	}
}

type reloadIDArtifact struct {
	Mode os.FileMode
	Data string
}

func reloadIDTree(t *testing.T, root string) map[string]reloadIDArtifact {
	t.Helper()
	result := make(map[string]reloadIDArtifact)
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
		artifact := reloadIDArtifact{Mode: info.Mode()}
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

func reloadIDFixture(t *testing.T, fixture string) (string, string, *rule.Engine) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "seed directory with spaces", "rules.json")
	prior := reloadIDRule("prior", 80)
	prior.Priority = 200
	seeds := []*model.Rule{prior}
	if fixture == "existing-reserved" {
		seeds = append(seeds, reloadIDRule("reload", 81))
	}
	engine := rule.NewEngine()
	if err := engine.Reload(seeds); err != nil {
		t.Fatal(err)
	}
	if fixture == "file-parent" {
		if err := os.WriteFile(filepath.Dir(path), []byte("retained parent occupant"), 0o600); err != nil {
			t.Fatal(err)
		}
	} else if fixture != "absent" {
		if err := os.Mkdir(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := rule.SaveToFile(path, seeds); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root, path, engine
}

func TestHTTPRuleCreateRejectsReservedReloadIDWithoutMutation(t *testing.T) {
	for _, fixture := range []string{"absent", "healthy", "existing-reserved", "file-parent"} {
		for _, auth := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/auth=%v", fixture, auth), func(t *testing.T) {
				root, path, engine := reloadIDFixture(t, fixture)
				candidate := reloadIDRule("reload", 82)
				candidateBefore := *candidate
				candidateBefore.Config = append(json.RawMessage(nil), candidate.Config...)
				priorRules := engine.Rules()
				packet80, packet81 := reloadIDPacket(80), reloadIDPacket(81)
				before80, before81 := *packet80, *packet81
				alerts80, alerts81 := engine.Match(packet80), engine.Match(packet81)
				before := reloadIDTree(t, root)
				handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: auth, AuthToken: "fixture-auth-value"}).Handler()
				rec := reloadIDRequest(handler, http.MethodPost, "/api/rules", reloadIDBody(t, candidate), auth)
				assertReloadIDError(t, rec, 400, "VALIDATION_ERROR", "Invalid rule request", []string{reloadIDDetail})
				if !reflect.DeepEqual(reloadIDTree(t, root), before) || engine.RuleCount() != len(priorRules) || !reflect.DeepEqual(engine.Rules(), priorRules) || !reflect.DeepEqual(engine.Match(packet80), alerts80) || !reflect.DeepEqual(engine.Match(packet81), alerts81) || *packet80 != before80 || *packet81 != before81 || !reflect.DeepEqual(*candidate, candidateBefore) {
					t.Fatal("reserved-ID rejection changed tree, snapshot, matching or caller data")
				}
			})
		}
	}
}

func TestHTTPRuleReloadIDDiagnosticPrecedence(t *testing.T) {
	for _, name := range []string{"auth", "file", "decode", "required-id", "forbidden-id", "before-name", "before-config"} {
		t.Run(name, func(t *testing.T) {
			root, path, engine := reloadIDFixture(t, "healthy")
			candidate := reloadIDRule("reload", 82)
			opts := api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}
			status, code, message, details, authorized := 400, "VALIDATION_ERROR", "Invalid rule request", []string{reloadIDDetail}, true
			switch name {
			case "auth":
				authorized, status, code, message, details = false, 401, "UNAUTHORIZED", "Valid bearer token required", nil
			case "file":
				opts.RulesSeedFile = ""
				status, code, message, details = 409, "RULES_WRITE_UNAVAILABLE", "Rules seed file is not configured", nil
			case "decode":
				details = []string{`json: unknown field "unknown"`}
			case "required-id":
				candidate.ID, candidate.Name, details = "", "", []string{"id is required"}
			case "forbidden-id":
				candidate.ID, candidate.Name, details = "reload/", "", []string{"id cannot contain /, ?, or #"}
			case "before-name":
				candidate.Name = ""
			case "before-config":
				candidate.Config = json.RawMessage(`{"ports":[]}`)
			}
			body := reloadIDBody(t, candidate)
			if name == "auth" || name == "file" {
				body = "{malformed"
			} else if name == "decode" {
				body = `{"id":"reload","unknown":true}`
			}
			before, prior := reloadIDTree(t, root), engine.Rules()
			packet := reloadIDPacket(80)
			matched := engine.Match(packet)
			rec := reloadIDRequest(api.NewServerWithOptions(nil, nil, engine, nil, opts).Handler(), http.MethodPost, "/api/rules", body, authorized)
			assertReloadIDError(t, rec, status, code, message, details)
			if !reflect.DeepEqual(reloadIDTree(t, root), before) || engine.RuleCount() != 1 || !reflect.DeepEqual(engine.Rules(), prior) || !reflect.DeepEqual(engine.Match(packet), matched) {
				t.Fatal("diagnostic precedence changed persisted or published state")
			}
		})
	}
}

func assertReloadIDRules(t *testing.T, got, want []*model.Rule) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("rule count = %d, want %d", len(got), len(want))
	}
	for i := range got {
		a, b := *got[i], *want[i]
		var ac, bc bytes.Buffer
		if err := json.Compact(&ac, a.Config); err != nil {
			t.Fatal(err)
		}
		if err := json.Compact(&bc, b.Config); err != nil {
			t.Fatal(err)
		}
		a.Config, b.Config = ac.Bytes(), bc.Bytes()
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("full rule %d = %+v, want %+v", i, a, b)
		}
	}
}

func assertReloadIDPublishedAndPersisted(t *testing.T, engine *rule.Engine, path string, want []*model.Rule) {
	t.Helper()
	assertReloadIDRules(t, engine.Rules(), want)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := rule.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertReloadIDRules(t, loaded, want)
	observer := rule.NewEngine()
	if err := observer.Reload(loaded); err != nil {
		t.Fatal(err)
	}
	assertReloadIDRules(t, observer.Rules(), want)
	for _, port := range []uint16{80, 81, 82} {
		var expected []*model.Alert
		for _, candidate := range want {
			var cfg model.PortBlacklistConfig
			if err := json.Unmarshal(candidate.Config, &cfg); err != nil {
				t.Fatal(err)
			}
			for _, p := range cfg.Ports {
				if p == int(port) {
					expected = append(expected, &model.Alert{RuleID: candidate.ID, RuleName: candidate.Name, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: port, Protocol: "TCP", Severity: model.SeverityHigh, MatchedKeyword: fmt.Sprintf("port_blacklist: %d", port)})
				}
			}
		}
		for _, observed := range []*rule.Engine{engine, observer} {
			packet := reloadIDPacket(port)
			prior := *packet
			if got := observed.Match(packet); !reflect.DeepEqual(got, expected) || *packet != prior || observed.RuleCount() != len(want) {
				t.Fatalf("published/rebuilt port %d = %+v, want %+v", port, got, expected)
			}
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("read-only loader/rebuild modified file: %v", err)
	}
}

func TestHTTPRuleReloadCaseVariantIDsRemainManageable(t *testing.T) {
	for _, id := range []string{"Reload", "RELOAD", "reload-extra"} {
		t.Run(id, func(t *testing.T) {
			_, path, engine := reloadIDFixture(t, "healthy")
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
			body := *updated
			body.ID = "" // Public update fills the exact ID from its path.
			rec = reloadIDRequest(handler, http.MethodPut, "/api/rules/"+id, reloadIDBody(t, &body), true)
			var response model.Rule
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || body.ID != "" {
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
			deleted := reloadIDRequest(handler, http.MethodDelete, "/api/rules/"+id, "", true)
			if deleted.Code != 204 || deleted.Body.Len() != 0 {
				t.Fatalf("delete = %d/%s", deleted.Code, deleted.Body.String())
			}
			assertReloadIDPublishedAndPersisted(t, engine, path, []*model.Rule{prior})
		})
	}
}

func TestHTTPFileLoadedReloadIDRetainsCompatibility(t *testing.T) {
	root, path, engine := reloadIDFixture(t, "existing-reserved")
	loaded, err := rule.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Reload(loaded); err != nil {
		t.Fatal(err)
	}
	prior, before := engine.Rules(), reloadIDTree(t, root)
	handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path, AuthEnabled: true, AuthToken: "fixture-auth-value"}).Handler()
	for _, endpoint := range []string{"/api/rules/reload", "/api/rules/%72eload"} {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			rec := reloadIDRequest(handler, method, endpoint, reloadIDBody(t, reloadIDRule("reload", 82)), true)
			assertReloadIDError(t, rec, 405, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
			if rec.Header().Get("Allow") != "POST" {
				t.Fatalf("reserved endpoint Allow = %q", rec.Header().Get("Allow"))
			}
		}
		rec := reloadIDRequest(handler, http.MethodPost, endpoint, "", true)
		var response struct{ Reloaded int }
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || rec.Code != 200 || response.Reloaded != 2 {
			t.Fatalf("existing reserved reload = %d/%s, error = %v", rec.Code, rec.Body.String(), err)
		}
		assertReloadIDPublishedAndPersisted(t, engine, path, prior)
		if !reflect.DeepEqual(reloadIDTree(t, root), before) {
			t.Fatal("existing reserved-ID routing/reload changed file bytes/modes/membership")
		}
	}
}
