package app

import "os"

type Config struct {
	Provider string
	Model    string
}

func DefaultConfig() Config {
	return Config{Provider: "codex"}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	if provider := os.Getenv("TRAIL_HUNK_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}
	if model := os.Getenv("TRAIL_HUNK_MODEL"); model != "" {
		cfg.Model = model
	}
	return cfg
}
