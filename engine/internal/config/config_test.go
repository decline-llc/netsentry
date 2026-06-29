package config

import (
	"strings"
	"testing"
)

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
