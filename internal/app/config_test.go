package app

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Provider != "codex" {
		t.Fatalf("Provider = %q, want codex", cfg.Provider)
	}
	if cfg.Model != "" {
		t.Fatalf("Model = %q, want empty default", cfg.Model)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("TRAIL_HUNK_PROVIDER", "claude")
	t.Setenv("TRAIL_HUNK_MODEL", "sonnet")

	cfg := ConfigFromEnv()
	if cfg.Provider != "claude" {
		t.Fatalf("Provider = %q, want claude", cfg.Provider)
	}
	if cfg.Model != "sonnet" {
		t.Fatalf("Model = %q, want sonnet", cfg.Model)
	}
}
