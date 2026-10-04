package api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func TestRuleHTTPProtocolWildcardPersistsRejectsAndPermitsRetry(t *testing.T) {
	for _, kind := range []model.RuleType{model.RuleTypePayloadMatch, model.RuleTypeIPBlacklist, model.RuleTypePortBlacklist} {
		for _, operation := range []string{"create", "update", "reload"} {
			for _, invalidProtocols := range [][]string{{"sctp", "any"}, {"any", "sctp"}} {
				t.Run(fmt.Sprintf("%s/%s/%v", kind, operation, invalidProtocols), func(t *testing.T) {
					fixture := func(id string, protocols []string) *model.Rule {
						t.Helper()
						config := map[string]any{"protocols": protocols, "direction": "dest"}
						switch kind {
						case model.RuleTypePayloadMatch:
							config["keywords"], config["ports"] = []string{"needle"}, []int{80}
						case model.RuleTypeIPBlacklist:
							config["ips"] = []string{"198.51.100.9"}
						case model.RuleTypePortBlacklist:
							config["ports"] = []int{80}
						}
						data, err := json.Marshal(config)
						if err != nil {
							t.Fatal(err)
						}
						return &model.Rule{ID: id, Name: "Protocol fixture", Type: kind, Severity: model.SeverityHigh, Enabled: true, Priority: 100, Config: data}
					}
					path := filepath.Join(t.TempDir(), "rules with spaces.json")
					seed := fixture("original", []string{"TCP"})
					if err := rule.SaveToFile(path, []*model.Rule{seed}); err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(path, 0o600); err != nil {
						t.Fatal(err)
					}
					engine := rule.NewEngine()
					if err := engine.Reload([]*model.Rule{seed}); err != nil {
						t.Fatal(err)
					}
					packet := &model.PacketInfo{SrcIP: "192.0.2.1", DstIP: "198.51.100.9", SrcPort: 40000, DstPort: 80, Protocol: 6, PayloadPreview: base64.StdEncoding.EncodeToString([]byte("needle"))}
					packetBefore := *packet
					priorRules, priorAlerts := engine.Rules(), engine.Match(packet)
					udp := *packet
					udp.Protocol = 17
					if len(priorAlerts) != 1 || len(engine.Match(&udp)) != 0 {
						t.Fatal("named-only seed fixture invalid")
					}
					handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path}).Handler()
					id := "original"
					if operation == "create" {
						id = "created"
					}
					invoke := func(candidate *model.Rule) *httptest.ResponseRecorder {
						t.Helper()
						method, target := http.MethodPost, "/api/rules"
						var body []byte
						var err error
						switch operation {
						case "update":
							method, target = http.MethodPut, "/api/rules/original"
						case "reload":
							target = "/api/rules/reload"
						}
						if operation != "reload" {
							body, err = json.Marshal(candidate)
							if err != nil {
								t.Fatal(err)
							}
						}
						req := httptest.NewRequest(method, target, bytes.NewReader(body))
						req.Header.Set("X-Request-ID", "req-protocol-wildcard")
						rec := httptest.NewRecorder()
						handler.ServeHTTP(rec, req)
						return rec
					}
					invalid := fixture(id, invalidProtocols)
					if operation == "reload" {
						if err := rule.SaveToFile(path, []*model.Rule{invalid}); err != nil {
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
					bytesBefore := readFile()
					infoBefore, err := os.Stat(path)
					if err != nil {
						t.Fatal(err)
					}
					membership := func() []string {
						t.Helper()
						entries, err := os.ReadDir(filepath.Dir(path))
						if err != nil {
							t.Fatal(err)
						}
						out := make([]string, len(entries))
						for i, entry := range entries {
							out[i] = entry.Name()
						}
						return out
					}
					entriesBefore := membership()
					rec := invoke(invalid)
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
					status, code, message, detailPrefix := http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rule request", "validate rules: "
					if operation == "reload" {
						status, code, message, detailPrefix = http.StatusBadRequest, "VALIDATION_ERROR", "Could not reload rules", ""
					}
					wantDetail := detailPrefix + fmt.Sprintf("rule %s: unsupported protocol %q", id, "sctp")
					if rec.Code != status || rec.Header().Get("Content-Type") != "application/json" || envelope.Error.Code != code || envelope.Error.Message != message || envelope.Error.RequestID != "req-protocol-wildcard" || !reflect.DeepEqual(envelope.Error.Details, []string{wantDetail}) {
						t.Fatalf("rejection = %d %s; want %d %s %q", rec.Code, rec.Body.String(), status, code, wantDetail)
					}
					infoAfter, err := os.Stat(path)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(readFile(), bytesBefore) || infoAfter.Mode() != infoBefore.Mode() || !reflect.DeepEqual(membership(), entriesBefore) || !reflect.DeepEqual(engine.Rules(), priorRules) || !reflect.DeepEqual(engine.Match(packet), priorAlerts) || engine.RuleCount() != 1 || len(engine.Match(&udp)) != 0 {
						t.Fatal("rejected request changed file bytes/mode/membership or published rules/matching/count")
					}
					valid := fixture(id, []string{"TCP", " AnY ", "", "UDP", "any"})
					if operation == "reload" {
						if err := rule.SaveToFile(path, []*model.Rule{valid}); err != nil {
							t.Fatal(err)
						}
					}
					validFileBefore := readFile()
					retry := invoke(valid)
					wantStatus, wantCount := http.StatusOK, 1
					if operation == "create" {
						wantStatus, wantCount = http.StatusCreated, 2
					}
					if retry.Code != wantStatus || engine.RuleCount() != wantCount {
						t.Fatalf("retry = %d %s; count=%d", retry.Code, retry.Body.String(), engine.RuleCount())
					}
					if operation == "reload" {
						var response struct {
							Reloaded int `json:"reloaded"`
						}
						if err := json.Unmarshal(retry.Body.Bytes(), &response); err != nil || response.Reloaded != 1 || !bytes.Equal(readFile(), validFileBefore) {
							t.Fatal("reload retry response or file changed")
						}
					} else {
						var response model.Rule
						if err := json.Unmarshal(retry.Body.Bytes(), &response); err != nil || response.ID != id || response.Type != kind || !strings.Contains(string(response.Config), `" AnY "`) {
							t.Fatal("mutation retry response lost original protocol list")
						}
					}
					loaded, err := rule.LoadFromFile(path)
					if err != nil || !reflect.DeepEqual(loaded, engine.Rules()) {
						t.Fatalf("canonical persistence differs from active rules: %v", err)
					}
					for _, loadedRule := range loaded {
						if loadedRule.ID == id {
							var config struct {
								Protocols []string `json:"protocols"`
							}
							if err := json.Unmarshal(loadedRule.Config, &config); err != nil || !reflect.DeepEqual(config.Protocols, []string{"TCP", " AnY ", "", "UDP", "any"}) {
								t.Fatal("persisted protocol list was normalized, reordered or narrowed")
							}
						}
					}
					for _, protocol := range []uint8{0, 1, 6, 17, 255} {
						probe := *packet
						probe.Protocol = protocol
						probeBefore := probe
						got := engine.Match(&probe)
						found := false
						for _, hit := range got {
							if hit.RuleID == id && hit.Protocol == model.ProtocolName(protocol) {
								found = true
								wantPreview, wantKeyword := "", "port_blacklist: 80"
								switch kind {
								case model.RuleTypePayloadMatch:
									wantPreview, wantKeyword = "needle", "needle"
								case model.RuleTypeIPBlacklist:
									wantKeyword = "ip_blacklist: 198.51.100.9"
								}
								if hit.MatchedKeyword != wantKeyword || hit.PayloadPreview != wantPreview || hit.SrcIP != probe.SrcIP || hit.DstIP != probe.DstIP || hit.DstPort != probe.DstPort || hit.Severity != model.SeverityHigh {
									t.Fatalf("wildcard changed alert contents: %+v", hit)
								}
							}
						}
						if !found || probe != probeBefore {
							t.Fatalf("protocol %d lost wildcard matching or changed probe: %+v", protocol, got)
						}
					}
					if *packet != packetBefore {
						t.Fatal("HTTP operations changed caller packet")
					}
				})
			}
		}
	}
}
