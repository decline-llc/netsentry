package api_test

import (
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

func TestHTTPRuleReloadPreservesLegacyDecodeFailureAndAllowsRepair(t *testing.T) {
	for _, field := range []string{"mitre_tactic", "mitre_technique_id", "mitre_technique_name"} {
		for _, nullPrefix := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/null=%v", field, nullPrefix), func(t *testing.T) {
				engine := rule.NewEngine()
				seed := &model.Rule{ID: "prior", Name: "prior snapshot", Type: model.RuleTypePortBlacklist, Severity: model.SeverityHigh, Priority: 100, Enabled: true, Config: json.RawMessage(`{"ports":[80]}`)}
				if err := engine.Reload([]*model.Rule{seed}); err != nil {
					t.Fatal(err)
				}
				packet := &model.PacketInfo{DstPort: 80, Protocol: 6, PayloadPreview: base64.StdEncoding.EncodeToString([]byte("needle"))}
				originalPacket := *packet
				priorRules, priorAlerts := engine.Rules(), engine.Match(packet)
				fixture := map[string]any{
					"id": "repaired", "name": "legacy repaired", "type": "payload_match", "severity": "high", "enabled": true,
					"config":       map[string]any{"keywords": []string{"needle"}},
					"mitre_tactic": "Initial Access", "mitre_technique_id": "T1190", "mitre_technique_name": "Exploit Public-Facing Application",
				}
				valid := fixture[field]
				fixture[field] = map[string]any{}
				items := []any{fixture}
				if nullPrefix {
					items = append([]any{nil}, items...)
				}
				dir := filepath.Join(t.TempDir(), "seed directory with spaces")
				if err := os.Mkdir(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, "rules.json")
				badData, err := json.Marshal(map[string]any{"rules": items})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, badData, 0o600); err != nil {
					t.Fatal(err)
				}
				handler := api.NewServerWithOptions(nil, nil, engine, nil, api.Options{RulesSeedFile: path}).Handler()
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil)
				req.Header.Set("X-Request-ID", "req-legacy-decode")
				handler.ServeHTTP(rec, req)
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
				if rec.Code != http.StatusInternalServerError || rec.Header().Get("Content-Type") != "application/json" || envelope.Error.Code != "INTERNAL_ERROR" || envelope.Error.Message != "Could not load rules" || envelope.Error.RequestID != "req-legacy-decode" || len(envelope.Error.Details) != 1 || !strings.Contains(envelope.Error.Details[0], field) || !strings.HasPrefix(envelope.Error.Details[0], "parse rules "+path+": ") {
					t.Fatalf("load-error envelope changed: status %d body %s", rec.Code, rec.Body.String())
				}
				if engine.RuleCount() != 1 || !reflect.DeepEqual(engine.Rules(), priorRules) || !reflect.DeepEqual(engine.Match(packet), priorAlerts) {
					t.Fatal("HTTP decode rejection replaced active snapshot")
				}
				if after, err := os.ReadFile(path); err != nil || string(after) != string(badData) {
					t.Fatalf("HTTP decode rejection changed seed bytes: %v", err)
				}
				// The operator repairs the field and removes the null entry before retry.
				fixture[field] = valid
				goodData, err := json.Marshal(map[string]any{"rules": []any{fixture}})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, goodData, 0o600); err != nil {
					t.Fatal(err)
				}
				retry := httptest.NewRecorder()
				handler.ServeHTTP(retry, httptest.NewRequest(http.MethodPost, "/api/rules/reload", nil))
				var response struct {
					Reloaded int `json:"reloaded"`
				}
				if err := json.Unmarshal(retry.Body.Bytes(), &response); err != nil || retry.Code != http.StatusOK || response.Reloaded != 1 {
					t.Fatalf("explicit repair retry failed: status %d body %s error %v", retry.Code, retry.Body.String(), err)
				}
				alerts := engine.Match(packet)
				if engine.RuleCount() != 1 || len(alerts) != 1 || alerts[0].RuleID != "repaired" || alerts[0].MatchedKeyword != "needle" || alerts[0].MitreTechniqueID != "T1190" || alerts[0].MitreTactic != "Initial Access" || alerts[0].MitreTechniqueName != "Exploit Public-Facing Application" || *packet != originalPacket {
					t.Fatalf("repair retry changed normalized tuple/match/input: %+v", alerts)
				}
				if after, err := os.ReadFile(path); err != nil || string(after) != string(goodData) {
					t.Fatalf("valid HTTP reload changed seed bytes: %v", err)
				}
			})
		}
	}
}
