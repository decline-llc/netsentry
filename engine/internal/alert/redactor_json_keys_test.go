package alert

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestRedactJSONDecodedCredentialNames(t *testing.T) {
	for _, key := range []string{"password", "token", "PaSsWoRd", "TOKEN"} {
		spellings := []string{key}
		var full strings.Builder
		for i, c := range key {
			escaped := fmt.Sprintf(`\u%04x`, c)
			spellings = append(spellings, key[:i]+escaped+key[i+1:])
			full.WriteString(escaped)
		}
		spellings = append(spellings, full.String())
		for i, spelling := range spellings {
			for j, value := range []string{"", "secret-canary", `quote"tail-canary`, `slash\tail-canary`, "line\ncanary", "π-canary"} {
				t.Run(fmt.Sprintf("%s/%d/%d", key, i, j), func(t *testing.T) {
					encoded, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					prefix := `{"public":"keep","outer":[{"` + spelling + `" : `
					input := prefix + string(encoded) + `}],"public2":"stay"}`
					want := prefix + `"[REDACTED]"}],"public2":"stay"}`
					got := RedactSensitivePayload(input)
					if !json.Valid([]byte(input)) || got != want || !json.Valid([]byte(got)) || strings.Contains(got, "canary") {
						t.Fatalf("decoded-name redaction = %q, want %q", got, want)
					}
					var decoded struct {
						Public  string              `json:"public"`
						Outer   []map[string]string `json:"outer"`
						Public2 string              `json:"public2"`
					}
					if err := json.Unmarshal([]byte(got), &decoded); err != nil || len(decoded.Outer) != 1 || decoded.Outer[0][key] != "[REDACTED]" || decoded.Public != "keep" || decoded.Public2 != "stay" {
						t.Fatalf("decoded fields = %+v, error = %v", decoded, err)
					}
					if twice := RedactSensitivePayload(got); twice != got {
						t.Fatalf("redaction not idempotent: %q", twice)
					}
				})
			}
		}
	}
}

func TestRedactJSONEscapedNamesAtPreviewEndAndRepeatedFields(t *testing.T) {
	for _, key := range []string{`pass\u0077ord`, `\u0074\u006f\u006b\u0065\u006e`, `\u0054OKEN`} {
		for i, tail := range []string{"", "secret-canary", `quote\"tail-canary`, `slash\\tail-canary`, "canary\\", `canary\u00`} {
			t.Run(fmt.Sprintf("%s/%d", key, i), func(t *testing.T) {
				prefix := "POST / HTTP/1.1\r\n\r\n" + `{"token":"earlier-canary","` + key + `" : "`
				input := prefix + tail
				want := "POST / HTTP/1.1\r\n\r\n" + `{"token":"[REDACTED]","` + key + `" : "[REDACTED]`
				got := RedactSensitivePayload(input)
				if got != want || strings.Contains(got, "canary") || RedactSensitivePayload(got) != got {
					t.Fatalf("preview-end redaction = %q, want %q", got, want)
				}
			})
		}
	}
	input := "{\r\n  \"pass\\u0077ord\"\t : \"one\\\"canary\",\r\n  \"pass\\u0077ord\":\"two\\\\canary\",\r\n  \"public\":\"keep\"\r\n}"
	want := "{\r\n  \"pass\\u0077ord\"\t : \"[REDACTED]\",\r\n  \"pass\\u0077ord\":\"[REDACTED]\",\r\n  \"public\":\"keep\"\r\n}"
	if got := RedactSensitivePayload(input); !json.Valid([]byte(input)) || got != want || !json.Valid([]byte(got)) || RedactSensitivePayload(got) != got {
		t.Fatalf("repeated formatted names = %q, want %q", got, want)
	}
}

func TestRedactJSONKeyControlsAndBatchMetadata(t *testing.T) {
	for _, input := range []string{
		`{"pass\\u0077ord":"keep"}`, `{"to\"ken":"keep"}`, `{"to\/ken":"keep"}`,
		`{"pass\u0077ord_suffix":"keep","prefix_token":"keep"}`,
		`{"pass\u007word":"keep"}`, `{"to\x6ben":"keep"}`, `{"to\u006":"keep"}`,
		`{"public":"keep\\tail","other":"keep\"tail"}`,
		`{"pass\u0077ord":123,"to\u006ben":null,"token":false}`,
		`{"pass\u0077ord`,
	} {
		if got := RedactSensitivePayload(input); got != input {
			t.Fatalf("noncredential/unsupported input changed: %q to %q", input, got)
		}
	}
	stamp := time.Unix(1719300000, 123456000).UTC()
	first := &model.Alert{ID: "row", EventID: "event", RuleID: "rule", RuleName: "fixture", Timestamp: stamp, Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, PayloadPreview: `{"pass\u0077ord":"quote\"canary"}`, RawPayload: "raw-marker", MatchedKeyword: "keep"}
	second := &model.Alert{ID: "second", PayloadPreview: `{"to\u006ben":"canary\`}
	empty := &model.Alert{ID: "empty"}
	wantFirst, wantSecond := *first, *second
	wantFirst.PayloadPreview, wantSecond.PayloadPreview = `{"pass\u0077ord":"[REDACTED]"}`, `{"to\u006ben":"[REDACTED]`
	batch := []*model.Alert{nil, empty, first, second}
	for i := 0; i < 2; i++ {
		RedactSensitivePayloads(batch)
		if len(batch) != 4 || batch[0] != nil || batch[1] != empty || batch[2] != first || batch[3] != second || *empty != (model.Alert{ID: "empty"}) || *first != wantFirst || *second != wantSecond {
			t.Fatalf("batch metadata/order/content changed: %+v %+v", first, second)
		}
	}
}
