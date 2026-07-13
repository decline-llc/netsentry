package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRepositoryConfigFile(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "configs", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Capture.UDSSocketPath == "" || cfg.Engine.UDSSocketPath == "" {
		t.Fatalf("expected capture and engine UDS paths, got capture=%q engine=%q", cfg.Capture.UDSSocketPath, cfg.Engine.UDSSocketPath)
	}
	if cfg.Engine.RulesSeedFile == "" {
		t.Fatal("expected engine.rules_seed_file to be configured")
	}
	if cfg.Engine.SuppressionsFile == "" {
		t.Fatal("expected engine.suppressions_file to be configured")
	}
	if cfg.Engine.AlertRecoveryLogPath != "" {
		t.Fatalf("repository config should use default recovery log path, got %q", cfg.Engine.AlertRecoveryLogPath)
	}
	if cfg.Engine.PprofEnabled {
		t.Fatal("repository config should not enable pprof by default")
	}
	if cfg.Engine.APIListenHost != "127.0.0.1" {
		t.Fatalf("repository API must default to loopback, got %q", cfg.Engine.APIListenHost)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		field    string
	}{
		{
			name:     "top level",
			contents: "unexpected: true\n",
			field:    "unexpected",
		},
		{
			name: "nested engine",
			contents: `engine:
  worker_count: 2
  worker_cout: 4
`,
			field: "worker_cout",
		},
		{
			name:     "multiple documents",
			contents: "{}\n---\n{}\n",
			field:    "multiple YAML documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.contents), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if err == nil || !strings.Contains(err.Error(), tt.field) {
				t.Fatalf("expected unknown field %q error, got %v", tt.field, err)
			}
		})
	}
}

func TestValidateRequiresAuthForNonLoopbackAPI(t *testing.T) {
	cfg := defaults()
	cfg.Engine.APIListenHost = "0.0.0.0"
	if err := validate(cfg); err == nil || !strings.Contains(err.Error(), "api_auth_enabled") {
		t.Fatalf("expected non-loopback auth validation error, got %v", err)
	}

	cfg.Engine.APIAuthEnabled = true
	cfg.Engine.APIAuthToken = "test-only-token"
	if err := validate(cfg); err != nil {
		t.Fatalf("authenticated non-loopback API should validate: %v", err)
	}
}

func TestValidateRejectsInvalidAPIPort(t *testing.T) {
	for _, port := range []int{0, -1, 65536} {
		cfg := defaults()
		cfg.Engine.APIPort = port
		if err := validate(cfg); err == nil || !strings.Contains(err.Error(), "api_port") {
			t.Fatalf("api_port=%d: expected validation error, got %v", port, err)
		}
	}
}

func TestValidateAllowsLoopbackPprofAddress(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:6060", "localhost:6060", "[::1]:6060"} {
		t.Run(addr, func(t *testing.T) {
			cfg := defaults()
			cfg.Engine.PprofEnabled = true
			cfg.Engine.PprofAddr = addr
			if err := validate(cfg); err != nil {
				t.Fatalf("validate(%q): %v", addr, err)
			}
		})
	}
}

func TestValidateRejectsPublicPprofAddress(t *testing.T) {
	cfg := defaults()
	cfg.Engine.PprofEnabled = true
	cfg.Engine.PprofAddr = "0.0.0.0:6060"
	err := validate(cfg)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "engine.pprof_addr") {
		t.Fatalf("unexpected error: %v", err)
	}
}
