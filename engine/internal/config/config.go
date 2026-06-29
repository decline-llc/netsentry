package config

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type CaptureConfig struct {
	Mode              string `yaml:"mode"`
	OfflineFile       string `yaml:"offline_file"`
	PayloadPreviewLen int    `yaml:"payload_preview_len"`
	UDSSocketPath     string `yaml:"uds_socket_path"`
	UDSSocketMode     string `yaml:"uds_socket_mode"`
	HeartbeatInterval int    `yaml:"heartbeat_interval"`
}

type EngineConfig struct {
	UDSSocketPath               string   `yaml:"uds_socket_path"`
	ChannelBufferSize           int      `yaml:"channel_buffer_size"`
	WorkerCount                 int      `yaml:"worker_count"`
	DBDir                       string   `yaml:"db_dir"`
	DBPath                      string   `yaml:"db_path"`
	DBShardDaily                bool     `yaml:"db_shard_daily"`
	DBJournalMode               string   `yaml:"db_journal_mode"`
	DBBusyTimeout               int      `yaml:"db_busy_timeout"`
	RulesSeedFile               string   `yaml:"rules_seed_file"`
	APIPort                     int      `yaml:"api_port"`
	CORSAllowedOrigins          []string `yaml:"cors_allowed_origins"`
	AlertAggregationWindow      int      `yaml:"alert_aggregation_window"`
	AlertAggregationMaxCount    int      `yaml:"alert_aggregation_max_count"`
	AlertRetentionDays          int      `yaml:"alert_retention_days"`
	APIAuthToken                string   `yaml:"api_auth_token"`
	APIAuthEnabled              bool     `yaml:"api_auth_enabled"`
	RedactSensitiveFields       bool     `yaml:"redact_sensitive_fields"`
	HealthFreshnessLimitSeconds int      `yaml:"health_freshness_limit_seconds"`
	PprofEnabled                bool     `yaml:"pprof_enabled"`
	PprofAddr                   string   `yaml:"pprof_addr"`
}

type LoggingConfig struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	EngineLog string `yaml:"engine_log"`
}

type Config struct {
	Capture CaptureConfig `yaml:"capture"`
	Engine  EngineConfig  `yaml:"engine"`
	Logging LoggingConfig `yaml:"logging"`

	// Path is the file this config was loaded from.
	Path string `yaml:"-"`
}

var envVarRe = regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

func expandEnv(s string) string {
	return envVarRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := envVarRe.FindStringSubmatch(m)
		key, def := sub[1], sub[2]
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return def
	})
}

func expandConfigStrings(raw []byte) []byte {
	return []byte(expandEnv(string(raw)))
}

// Load reads and parses a YAML config file, expanding ${ENV_VAR} references.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	expanded := expandConfigStrings(raw)

	cfg := defaults()
	if err := yaml.Unmarshal(expanded, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.Path = path

	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Capture: CaptureConfig{
			Mode:              "offline",
			PayloadPreviewLen: 4096,
			UDSSocketPath:     "/tmp/netsentry.sock",
			UDSSocketMode:     "0600",
			HeartbeatInterval: 5,
		},
		Engine: EngineConfig{
			UDSSocketPath:               "/tmp/netsentry.sock",
			ChannelBufferSize:           10000,
			WorkerCount:                 1,
			DBDir:                       "data/",
			DBPath:                      "data/netsentry.db",
			DBShardDaily:                false,
			DBJournalMode:               "WAL",
			DBBusyTimeout:               5000,
			RulesSeedFile:               "configs/rules.json",
			APIPort:                     8080,
			CORSAllowedOrigins:          []string{"http://localhost:3000"},
			AlertAggregationWindow:      60,
			AlertAggregationMaxCount:    100,
			AlertRetentionDays:          7,
			APIAuthEnabled:              false,
			RedactSensitiveFields:       true,
			HealthFreshnessLimitSeconds: 30,
			PprofEnabled:                false,
			PprofAddr:                   "127.0.0.1:6060",
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "json",
			EngineLog: "logs/engine.log",
		},
	}
}

func validate(cfg *Config) error {
	var errs []string
	if cfg.Engine.APIAuthEnabled && strings.TrimSpace(cfg.Engine.APIAuthToken) == "" {
		errs = append(errs, "engine.api_auth_token must be set when api_auth_enabled is true")
	}
	if cfg.Engine.PprofEnabled && !isLoopbackListenAddr(cfg.Engine.PprofAddr) {
		errs = append(errs, "engine.pprof_addr must be a localhost or loopback address when pprof_enabled is true")
	}
	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func isLoopbackListenAddr(addr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil || strings.TrimSpace(host) == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return ip.IsLoopback()
}
