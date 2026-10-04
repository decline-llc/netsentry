package config_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestLoadBusyTimeoutRepresentabilityAndDefaults(t *testing.T) {
	baseline, err := loadDurationConfigUnchanged(t, writeDurationConfig(t, "{}\n"))
	if err != nil || baseline == nil || baseline.Engine.DBBusyTimeout != 5000 {
		t.Fatalf("default config = %+v, error = %v", baseline, err)
	}
	for _, value := range []int64{-9223372036854775808, -1, 0, 1, 5000, 2147483647, 2147483648, 4294967295, 9223372036854775807} {
		t.Run(strconv.FormatInt(value, 10), func(t *testing.T) {
			path := writeDurationConfig(t, fmt.Sprintf("engine:\n  db_busy_timeout: %d\n", value))
			got, err := loadDurationConfigUnchanged(t, path)
			if durationValueOutsideNativeInt(value) {
				assertDurationParseRejection(t, got, err)
				return
			}
			if value > 2147483647 {
				want := "config validation failed:\n  - engine.db_busy_timeout must not exceed 2147483647 milliseconds"
				if got != nil || err == nil || err.Error() != want {
					t.Fatalf("Load = %+v, error = %v, want nil and %q", got, err, want)
				}
				return
			}
			want := *baseline
			want.Path, want.Engine.DBBusyTimeout = path, int(value)
			if err != nil || got == nil || *got != want {
				t.Fatalf("Load = %+v, error = %v, want complete config %+v", got, err, want)
			}
		})
	}
}

func TestLoadExpandedBusyTimeoutUsesTheSameBoundary(t *testing.T) {
	for _, value := range []int64{-1, 0, 2147483647, 2147483648, 4294967295} {
		t.Run(strconv.FormatInt(value, 10), func(t *testing.T) {
			t.Setenv("NETSENTRY_TEST_BUSY_TIMEOUT_MS", strconv.FormatInt(value, 10))
			path := writeDurationConfig(t, "engine:\n  db_busy_timeout: ${NETSENTRY_TEST_BUSY_TIMEOUT_MS}\n")
			got, err := loadDurationConfigUnchanged(t, path)
			if durationValueOutsideNativeInt(value) {
				assertDurationParseRejection(t, got, err)
				return
			}
			if value > 2147483647 {
				want := "config validation failed:\n  - engine.db_busy_timeout must not exceed 2147483647 milliseconds"
				if got != nil || err == nil || err.Error() != want {
					t.Fatalf("expanded Load = %+v, error = %v, want nil and %q", got, err, want)
				}
			} else if err != nil || got == nil || int64(got.Engine.DBBusyTimeout) != value {
				t.Fatalf("expanded Load = %+v, error = %v, want %d", got, err, value)
			}
		})
	}
}

func TestLoadBusyTimeoutDiagnosticOrdering(t *testing.T) {
	path := writeDurationConfig(t, "engine:\n  db_busy_timeout: 2147483648\n  alert_aggregation_window: 9223372037\n  api_port: 0\n")
	got, err := loadDurationConfigUnchanged(t, path)
	if strconv.IntSize == 32 {
		assertDurationParseRejection(t, got, err)
		return
	}
	want := "config validation failed:\n  - " + strings.Join([]string{
		"engine.db_busy_timeout must not exceed 2147483647 milliseconds",
		durationBoundsDiagnostic("alert_aggregation_window"),
		"engine.api_port must be between 1 and 65535",
	}, "\n  - ")
	if got != nil || err == nil || err.Error() != want {
		t.Fatalf("combined Load = %+v, error = %v, want nil and %q", got, err, want)
	}
}
