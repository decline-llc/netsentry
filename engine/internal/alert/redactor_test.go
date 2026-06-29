package alert

import (
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestRedactSensitivePayload(t *testing.T) {
	payload := "GET /login?token=abc123 HTTP/1.1\r\n" +
		"Authorization: Bearer secret-token\r\n" +
		"Cookie: session=secret; theme=dark\r\n" +
		"Set-Cookie: auth=secret\r\n" +
		"\r\n" +
		`{"password":"supersecret","token":"jsonsecret"}` + " password=hunter2"

	got := RedactSensitivePayload(payload)
	for _, leaked := range []string{"abc123", "secret-token", "session=secret", "auth=secret", "supersecret", "jsonsecret", "hunter2"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted payload leaked %q: %s", leaked, got)
		}
	}
	for _, want := range []string{"Authorization: [REDACTED]", "Cookie: [REDACTED]", "Set-Cookie: [REDACTED]", "token=[REDACTED]", `"password":"[REDACTED]"`, `"token":"[REDACTED]"`, "password=[REDACTED]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted payload missing %q: %s", want, got)
		}
	}
}

func TestRedactSensitivePayloadsSkipsNilAlerts(t *testing.T) {
	alerts := []*model.Alert{nil, &model.Alert{PayloadPreview: "password=secret"}}
	RedactSensitivePayloads(alerts)
	if alerts[1].PayloadPreview != "password=[REDACTED]" {
		t.Fatalf("unexpected payload: %s", alerts[1].PayloadPreview)
	}
}
