package alert

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/decline-llc/netsentry/pkg/model"
)

const redactedValue = "[REDACTED]"

var (
	sensitiveHeaderRe       = regexp.MustCompile(`(?i)\b(authorization|cookie|set-cookie)\s*:\s*[^\r\n]*`)
	sensitiveHeaderPrefixRe = regexp.MustCompile(`(?i)^\s*(authorization|cookie|set-cookie)\s*:`)
	sensitivePairRe         = regexp.MustCompile(`(?i)\b(password|token)\b\s*([=:])\s*[^&\s;\r\n]+`)
	// A bounded preview can end within a value, including after an escape backslash.
	sensitiveJSONRe = regexp.MustCompile(`(("(?:\\[^\r\n]|[^"\\\r\n])*")\s*:\s*")(?:\\[^\r\n]|[^"\\\r\n])*(?:(")|\\?$)`)
)

// RedactSensitivePayloads removes common credentials from alert payload previews.
func RedactSensitivePayloads(alerts []*model.Alert) {
	for _, alert := range alerts {
		if alert == nil || alert.PayloadPreview == "" {
			continue
		}
		alert.PayloadPreview = RedactSensitivePayload(alert.PayloadPreview)
	}
}

// RedactSensitivePayload redacts HTTP auth/cookie headers and common password/token fields.
func RedactSensitivePayload(payload string) string {
	payload = sensitiveHeaderRe.ReplaceAllStringFunc(payload, func(match string) string {
		prefix := sensitiveHeaderPrefixRe.FindString(match)
		if prefix == "" {
			return match
		}
		return prefix + " " + redactedValue
	})
	payload = sensitiveJSONRe.ReplaceAllStringFunc(payload, func(match string) string {
		parts := sensitiveJSONRe.FindStringSubmatch(match)
		// Decode only the complete key; bounded previews need not be valid JSON.
		var key string
		if err := json.Unmarshal([]byte(parts[2]), &key); err != nil ||
			(!strings.EqualFold(key, "password") && !strings.EqualFold(key, "token")) {
			return match
		}
		return parts[1] + redactedValue + parts[3]
	})
	payload = sensitivePairRe.ReplaceAllString(payload, `${1}${2}`+redactedValue)
	return payload
}
