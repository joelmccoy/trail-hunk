package app

import "testing"

func TestNewAIProviderSelectsCodex(t *testing.T) {
	provider, err := NewAIProvider(Config{Provider: "codex", Model: "gpt-5"})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Name() != "codex" {
		t.Fatalf("Name = %q, want codex", provider.Name())
	}
}

func TestNewAIProviderSelectsClaude(t *testing.T) {
	provider, err := NewAIProvider(Config{Provider: "claude", Model: "sonnet"})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Name() != "claude" {
		t.Fatalf("Name = %q, want claude", provider.Name())
	}
}

func TestNewAIProviderSelectsFixture(t *testing.T) {
	provider, err := NewAIProvider(Config{Provider: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Name() != "fixture" {
		t.Fatalf("Name = %q, want fixture", provider.Name())
	}
}

func TestNewAIProviderRejectsUnknownProvider(t *testing.T) {
	_, err := NewAIProvider(Config{Provider: "unknown"})
	if err == nil {
		t.Fatal("expected unknown provider error")
	}
}

func TestNewReviewDependenciesWiresConcreteBoundaries(t *testing.T) {
	deps, err := NewReviewDependencies(Config{Provider: "codex"}, "token")
	if err != nil {
		t.Fatal(err)
	}
	if deps.Repo == nil {
		t.Fatal("Repo dependency is nil")
	}
	if deps.GitHub == nil {
		t.Fatal("GitHub dependency is nil")
	}
	if deps.AI == nil {
		t.Fatal("AI dependency is nil")
	}
}
