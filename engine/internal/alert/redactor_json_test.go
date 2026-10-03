package alert

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestRedactJSONCredentialValuesWithEscapes(t *testing.T) {
	values := []string{
		"plain-canary", `quote"suffix-canary`, `backslash\suffix-canary`,
		`backslash\"suffix-canary`, "line\ncontrol\tcanary", "nul\x00canary",
		"unicode-π-canary", "",
	}
	for _, key := range []string{"password", "token", "PaSsWoRd", "TOKEN"} {
		for i, value := range values {
			t.Run(fmt.Sprintf("%s/%d", key, i), func(t *testing.T) {
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				input := `{"` + key + `":` + string(encoded) + `,"public":"keep"}`
				want := `{"` + key + `":"[REDACTED]","public":"keep"}`
				if !json.Valid([]byte(input)) {
					t.Fatalf("invalid source fixture: %q", input)
				}
				got := RedactSensitivePayload(input)
				if got != want || !json.Valid([]byte(got)) || strings.Contains(got, "canary") {
					t.Fatalf("full-value redaction = %q, want %q", got, want)
				}
				var decoded map[string]string
				if err := json.Unmarshal([]byte(got), &decoded); err != nil || decoded[key] != redactedValue || decoded["public"] != "keep" {
					t.Fatalf("decoded redaction = %+v, error=%v", decoded, err)
				}
				if twice := RedactSensitivePayload(got); twice != got {
					t.Fatalf("redaction is not idempotent: %q then %q", got, twice)
				}
			})
		}
	}
}

func TestRedactJSONValueQuoteParityFormattingAndMultipleFields(t *testing.T) {
	cases := []struct {
		name, input, want string
		completeJSON      bool
	}{
		{"escaped_quote", `{"password":"prefix\"suffix-canary","public":"keep"}`, `{"password":"[REDACTED]","public":"keep"}`, true},
		{"escaped_backslash_then_quote", `{"token":"prefix\\\"suffix-canary","public":"keep"}`, `{"token":"[REDACTED]","public":"keep"}`, true},
		{"even_backslashes_before_terminator", `{"password":"prefix\\\\","token":"second\"suffix-canary","public":"keep"}`, `{"password":"[REDACTED]","token":"[REDACTED]","public":"keep"}`, true},
		{"unicode_escape", `{"token":"prefix\u0022suffix-canary","public":"keep"}`, `{"token":"[REDACTED]","public":"keep"}`, true},
		{"escaped_slash", `{"password":"prefix\/suffix-canary","public":"keep"}`, `{"password":"[REDACTED]","public":"keep"}`, true},
		{"nested", `{"outer":[{"PaSsWoRd":"a\"suffix-canary"}],"public":"keep"}`, `{"outer":[{"PaSsWoRd":"[REDACTED]"}],"public":"keep"}`, true},
		{"repeated_key", `{"token":"a\"suffix-canary","token":"b\\end-canary","public":"keep"}`, `{"token":"[REDACTED]","token":"[REDACTED]","public":"keep"}`, true},
		{"whitespace", "{\r\n  \"password\" \t: \"prefix\\\"suffix-canary\",\r\n  \"public\": \"keep\"\r\n}", "{\r\n  \"password\" \t: \"[REDACTED]\",\r\n  \"public\": \"keep\"\r\n}", true},
		{"complete_value_in_fragment", "POST / HTTP/1.1\r\n\r\n" + `{"token":"a\"suffix-canary"`, "POST / HTTP/1.1\r\n\r\n" + `{"token":"[REDACTED]"`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.completeJSON && !json.Valid([]byte(tc.input)) {
				t.Fatalf("invalid quote-parity fixture: %q", tc.input)
			}
			got := RedactSensitivePayload(tc.input)
			if got != tc.want || strings.Contains(got, "canary") {
				t.Fatalf("redacted preview = %q, want %q", got, tc.want)
			}
			if tc.completeJSON && !json.Valid([]byte(got)) {
				t.Fatalf("redaction broke JSON syntax: %q", got)
			}
			if twice := RedactSensitivePayload(got); twice != got {
				t.Fatalf("redaction changed on repeat: %q", twice)
			}
		})
	}
}

func TestRedactEscapedJSONBatchPreservesAlertMetadataAndNilEntries(t *testing.T) {
	stamp := time.Unix(1719300000, 123456000).UTC()
	first := &model.Alert{ID: "alert-1", EventID: "event-1", RuleID: "rule-1", RuleName: "fixture", Timestamp: stamp, Severity: model.SeverityHigh, SrcIP: "192.0.2.1", DstIP: "198.51.100.1", DstPort: 443, PayloadPreview: `{"password":"a\"suffix-canary"}`, MatchedKeyword: "match-marker", RawPayload: "raw-marker"}
	second := &model.Alert{ID: "alert-2", RuleID: "rule-2", PayloadPreview: `{"token":"b\\\"suffix-canary"}`}
	empty := &model.Alert{RuleID: "empty"}
	wantFirst, wantSecond := *first, *second
	wantFirst.PayloadPreview, wantSecond.PayloadPreview = `{"password":"[REDACTED]"}`, `{"token":"[REDACTED]"}`
	alerts := []*model.Alert{nil, empty, first, second}
	RedactSensitivePayloads(alerts)
	if alerts[0] != nil || alerts[1] != empty || *empty != (model.Alert{RuleID: "empty"}) || *first != wantFirst || *second != wantSecond {
		t.Fatalf("batch redaction changed metadata/order or left suffixes: %+v %+v", first, second)
	}
	RedactSensitivePayloads(alerts)
	if *first != wantFirst || *second != wantSecond {
		t.Fatal("batch redaction changed on repeat")
	}
}

func TestJSONRedactionPreservesExistingHeaderPairAndUnrelatedValues(t *testing.T) {
	input := "Authorization: Bearer fixture\r\nCookie: a=fixture\r\nSet-Cookie: b=fixture\r\n\r\npassword=fixture&token=fixture " + `{"password":"plain"}`
	want := "Authorization: [REDACTED]\r\nCookie: [REDACTED]\r\nSet-Cookie: [REDACTED]\r\n\r\npassword=[REDACTED]&token=[REDACTED] " + `{"password":"[REDACTED]"}`
	if got := RedactSensitivePayload(input); got != want {
		t.Fatalf("header/pair compatibility = %q, want %q", got, want)
	}
	for _, unchanged := range []string{
		`{"public":"keep\"suffix","password":123,"token":null,"auth":"keep"}`,
		`{"pass\u0077ord":"a\"suffix-canary"}`,
		`{"password":"open-canary`,
		`{"password":"a\"suffix-canary`,
		"{\"token\":\"line\r\nsuffix-canary\"}",
	} {
		if got := RedactSensitivePayload(unchanged); got != unchanged {
			t.Fatalf("out-of-scope fixture changed: %q to %q", unchanged, got)
		}
	}
}
