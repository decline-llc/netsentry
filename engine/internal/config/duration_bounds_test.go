package config_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/internal/config"
)

var durationFields = []string{"alert_aggregation_window", "health_freshness_limit_seconds"}

func TestLoadDurationSettingsPreserveRepresentableValuesAndRejectOverflow(t *testing.T) {
	baseline, err := loadDurationConfigUnchanged(t, writeDurationConfig(t, "{}\n"))
	if err != nil {
		t.Fatalf("load default baseline: %v", err)
	}
	cases := []struct {
		name   string
		value  int64
		accept bool
	}{
		{"minimum_seconds", -9223372036, true},
		{"negative_default_request", -1, true},
		{"zero_default_request", 0, true},
		{"one_second", 1, true},
		{"maximum_seconds", 9223372036, true},
		{"first_negative_overflow", -9223372037, false},
		{"first_positive_overflow", 9223372037, false},
		{"minimum_int64", -9223372036854775808, false},
		{"maximum_int64", 9223372036854775807, false},
	}
	for _, field := range durationFields {
		for _, tc := range cases {
			t.Run(field+"/"+tc.name, func(t *testing.T) {
				path := writeDurationConfig(t, fmt.Sprintf("engine:\n  %s: %d\n", field, tc.value))
				got, err := loadDurationConfigUnchanged(t, path)
				if durationValueOutsideNativeInt(tc.value) {
					assertDurationParseRejection(t, got, err)
					return
				}
				if !tc.accept {
					want := "config validation failed:\n  - " + durationBoundsDiagnostic(field)
					if got != nil || err == nil || err.Error() != want {
						t.Fatalf("overflow Load=%+v err=%v, want nil and %q", got, err, want)
					}
					return
				}
				if err != nil || got == nil {
					t.Fatalf("representable Load=%+v err=%v", got, err)
				}
				want := *baseline
				want.Path = path
				switch field {
				case "alert_aggregation_window":
					want.Engine.AlertAggregationWindow = int(tc.value)
				case "health_freshness_limit_seconds":
					want.Engine.HealthFreshnessLimitSeconds = int(tc.value)
				}
				if *got != want || durationSetting(got, field) != tc.value {
					t.Fatalf("Load=%+v, want complete config %+v", got, want)
				}
				duration := time.Duration(durationSetting(got, field)) * time.Second
				if int64(duration/time.Second) != tc.value || (tc.value > 0 && duration <= 0) || (tc.value < 0 && duration >= 0) {
					t.Fatalf("seconds %d lost representability: duration=%v", tc.value, duration)
				}
			})
		}
	}
}

func TestLoadDurationDefaultsAndCombinedDiagnostics(t *testing.T) {
	path := writeDurationConfig(t, "{}\n")
	got, err := loadDurationConfigUnchanged(t, path)
	if err != nil || got == nil || got.Engine.AlertAggregationWindow != 60 || got.Engine.HealthFreshnessLimitSeconds != 30 {
		t.Fatalf("default Load=%+v err=%v, want aggregation=60 freshness=30", got, err)
	}
	path = writeDurationConfig(t, "engine:\n  alert_aggregation_window: 9223372037\n  health_freshness_limit_seconds: -9223372037\n  api_port: 0\n")
	got, err = loadDurationConfigUnchanged(t, path)
	if strconv.IntSize == 32 {
		assertDurationParseRejection(t, got, err)
		return
	}
	want := "config validation failed:\n  - " + durationBoundsDiagnostic(durationFields[0]) +
		"\n  - " + durationBoundsDiagnostic(durationFields[1]) +
		"\n  - engine.api_port must be between 1 and 65535"
	if got != nil || err == nil || err.Error() != want {
		t.Fatalf("combined Load=%+v err=%v, want nil and %q", got, err, want)
	}
}

func TestLoadExpandedDurationSettingsUseTheSameBounds(t *testing.T) {
	for _, field := range durationFields {
		for _, value := range []int64{60, -9223372037, 9223372037} {
			t.Run(fmt.Sprintf("%s/%d", field, value), func(t *testing.T) {
				t.Setenv("NETSENTRY_TEST_DURATION_SECONDS", strconv.FormatInt(value, 10))
				path := writeDurationConfig(t, fmt.Sprintf("engine:\n  %s: ${NETSENTRY_TEST_DURATION_SECONDS}\n", field))
				got, err := loadDurationConfigUnchanged(t, path)
				if durationValueOutsideNativeInt(value) {
					assertDurationParseRejection(t, got, err)
					return
				}
				if value == 60 {
					if err != nil || got == nil || durationSetting(got, field) != value {
						t.Fatalf("expanded Load=%+v err=%v, want %d", got, err, value)
					}
					return
				}
				want := "config validation failed:\n  - " + durationBoundsDiagnostic(field)
				if got != nil || err == nil || err.Error() != want {
					t.Fatalf("expanded overflow Load=%+v err=%v, want nil and %q", got, err, want)
				}
			})
		}
	}
}

func writeDurationConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "duration config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write duration fixture: %v", err)
	}
	return path
}

func loadDurationConfigUnchanged(t *testing.T, path string) (*config.Config, error) {
	t.Helper()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read input before Load: %v", err)
	}
	got, loadErr := config.Load(path)
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("Load modified input bytes: read error=%v", err)
	}
	return got, loadErr
}

func durationBoundsDiagnostic(field string) string {
	return "engine." + field + " must be between -9223372036 and 9223372036 seconds"
}

func durationSetting(cfg *config.Config, field string) int64 {
	if field == "alert_aggregation_window" {
		return int64(cfg.Engine.AlertAggregationWindow)
	}
	return int64(cfg.Engine.HealthFreshnessLimitSeconds)
}

func durationValueOutsideNativeInt(value int64) bool {
	return strconv.IntSize == 32 && (value < -2147483648 || value > 2147483647)
}

func assertDurationParseRejection(t *testing.T, got *config.Config, err error) {
	t.Helper()
	if got != nil || err == nil || !strings.HasPrefix(err.Error(), "parse config ") {
		t.Fatalf("out-of-native-int Load=%+v err=%v, want nil and parse error", got, err)
	}
}
