package config

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type EngineConfig struct {
	UDSSocketPath               string `yaml:"uds_socket_path"`
	UDSSocketMode               string `yaml:"uds_socket_mode"`
	ChannelBufferSize           int    `yaml:"channel_buffer_size"`
	WorkerCount                 int    `yaml:"worker_count"`
	DBDir                       string `yaml:"db_dir"`
	DBPath                      string `yaml:"db_path"`
	DBShardDaily                bool   `yaml:"db_shard_daily"`
	DBJournalMode               string `yaml:"db_journal_mode"`
	DBBusyTimeout               int    `yaml:"db_busy_timeout"`
	AlertRecoveryLogPath        string `yaml:"alert_recovery_log_path"`
	RulesSeedFile               string `yaml:"rules_seed_file"`
	SuppressionsFile            string `yaml:"suppressions_file"`
	APIListenHost               string `yaml:"api_listen_host"`
	APIPort                     int    `yaml:"api_port"`
	AlertAggregationWindow      int    `yaml:"alert_aggregation_window"`
	AlertRetentionDays          int    `yaml:"alert_retention_days"`
	APIAuthToken                string `yaml:"api_auth_token"`
	APIAuthEnabled              bool   `yaml:"api_auth_enabled"`
	RedactSensitiveFields       bool   `yaml:"redact_sensitive_fields"`
	HealthFreshnessLimitSeconds int    `yaml:"health_freshness_limit_seconds"`
	PprofEnabled                bool   `yaml:"pprof_enabled"`
	PprofAddr                   string `yaml:"pprof_addr"`
}

type LoggingConfig struct {
	Format string `yaml:"format"`
}

type Config struct {
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
	decoder := yaml.NewDecoder(bytes.NewReader(expanded))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("parse config %s: multiple YAML documents are not supported", path)
		}
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
		Engine: EngineConfig{
			UDSSocketPath:               "/tmp/netsentry.sock",
			UDSSocketMode:               "0600",
			ChannelBufferSize:           10000,
			WorkerCount:                 1,
			DBDir:                       "data/",
			DBPath:                      "data/netsentry.db",
			DBShardDaily:                false,
			DBJournalMode:               "WAL",
			DBBusyTimeout:               5000,
			AlertRecoveryLogPath:        "",
			RulesSeedFile:               "configs/rules.json",
			SuppressionsFile:            "configs/suppressions.json",
			APIListenHost:               "127.0.0.1",
			APIPort:                     8080,
			AlertAggregationWindow:      60,
			AlertRetentionDays:          7,
			APIAuthEnabled:              false,
			RedactSensitiveFields:       true,
			HealthFreshnessLimitSeconds: 30,
			PprofEnabled:                false,
			PprofAddr:                   "127.0.0.1:6060",
		},
		Logging: LoggingConfig{
			Format: "json",
		},
	}
}

func validate(cfg *Config) error {
	var errs []string
	if mode, err := strconv.ParseUint(cfg.Engine.UDSSocketMode, 8, 32); err != nil || mode == 0 || mode > 0o777 {
		errs = append(errs, "engine.uds_socket_mode must be a non-zero octal permission mode no greater than 0777")
	}
	if cfg.Logging.Format != "json" && cfg.Logging.Format != "console" {
		errs = append(errs, "logging.format must be json or console")
	}
	if cfg.Engine.ChannelBufferSize < 1 || cfg.Engine.ChannelBufferSize > 1_000_000 {
		errs = append(errs, "engine.channel_buffer_size must be between 1 and 1000000")
	}
	if cfg.Engine.WorkerCount < 1 || cfg.Engine.WorkerCount > 64 {
		errs = append(errs, "engine.worker_count must be between 1 and 64")
	}
	if cfg.Engine.APIPort < 1 || cfg.Engine.APIPort > 65535 {
		errs = append(errs, "engine.api_port must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.Engine.APIListenHost) == "" {
		errs = append(errs, "engine.api_listen_host must be set")
	} else if !isLoopbackHost(cfg.Engine.APIListenHost) && !cfg.Engine.APIAuthEnabled {
		errs = append(errs, "engine.api_auth_enabled must be true when api_listen_host is not loopback")
	}
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

func isLoopbackHost(host string) bool {
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip, err := netip.ParseAddr(host)
	return err == nil && ip.IsLoopback()
}

func isLoopbackListenAddr(addr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil || strings.TrimSpace(host) == "" {
		return false
	}
	return isLoopbackHost(host)
}
