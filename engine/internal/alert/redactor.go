package alert

import (
	"regexp"

	"github.com/decline-llc/netsentry/pkg/model"
)

const redactedValue = "[REDACTED]"

var (
	sensitiveHeaderRe       = regexp.MustCompile(`(?i)\b(authorization|cookie|set-cookie)\s*:\s*[^\r\n]*`)
	sensitiveHeaderPrefixRe = regexp.MustCompile(`(?i)^\s*(authorization|cookie|set-cookie)\s*:`)
	sensitivePairRe         = regexp.MustCompile(`(?i)\b(password|token)\b\s*([=:])\s*[^&\s;\r\n]+`)
	sensitiveJSONRe         = regexp.MustCompile(`(?i)("(?:password|token)"\s*:\s*")[^"\r\n]*(")`)
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
	payload = sensitiveJSONRe.ReplaceAllString(payload, `${1}`+redactedValue+`${2}`)
	payload = sensitivePairRe.ReplaceAllString(payload, `${1}${2}`+redactedValue)
	return payload
}
