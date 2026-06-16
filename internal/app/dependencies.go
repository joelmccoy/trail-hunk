package app

import (
	"context"
	"fmt"

	"github.com/joelmccoy/trail-hunk/internal/ai"
	"github.com/joelmccoy/trail-hunk/internal/git"
	"github.com/joelmccoy/trail-hunk/internal/github"
)

type GitRepoDiscoverer struct{}

func (GitRepoDiscoverer) Discover(ctx context.Context, dir string) (git.Repository, error) {
	return git.Discover(ctx, dir)
}

func NewAIProvider(cfg Config) (ai.Provider, error) {
	switch cfg.Provider {
	case "", "codex":
		return ai.NewCodexProvider(cfg.Model), nil
	case "claude":
		return ai.NewClaudeProvider(cfg.Model), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider %q", cfg.Provider)
	}
}

func NewReviewDependencies(cfg Config, githubToken string) (ReviewDependencies, error) {
	provider, err := NewAIProvider(cfg)
	if err != nil {
		return ReviewDependencies{}, err
	}

	return ReviewDependencies{
		Repo:   GitRepoDiscoverer{},
		GitHub: github.NewClient("", githubToken),
		AI:     provider,
	}, nil
}
