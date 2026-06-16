package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joelmccoy/trail-hunk/internal/app"
	"github.com/joelmccoy/trail-hunk/internal/github"
	"github.com/joelmccoy/trail-hunk/internal/review"
	"github.com/joelmccoy/trail-hunk/internal/tui"
)

func main() {
	cfg := app.ConfigFromEnv()
	startupLoader := func(ctx context.Context) (review.StartupContext, error) {
		token, err := github.DiscoverToken(ctx)
		if err != nil {
			return review.StartupContext{}, err
		}

		deps, err := app.NewReviewDependencies(cfg, token)
		if err != nil {
			return review.StartupContext{}, err
		}

		return app.ResolveStartupContext(ctx, ".", app.StartupDependencies{
			Repo:   deps.Repo,
			GitHub: deps.GitHub,
		})
	}
	starter := func(ctx context.Context) (review.ReviewSession, error) {
		token, err := github.DiscoverToken(ctx)
		if err != nil {
			return review.ReviewSession{}, err
		}

		deps, err := app.NewReviewDependencies(cfg, token)
		if err != nil {
			return review.ReviewSession{}, err
		}

		return app.RunReview(ctx, ".", deps)
	}
	submitter := func(ctx context.Context, session review.ReviewSession) error {
		token, err := github.DiscoverToken(ctx)
		if err != nil {
			return err
		}
		return app.SubmitApprovedReview(ctx, session, github.NewClient("", token))
	}

	model := tui.NewModelWithOptions(review.ReviewSession{}, tui.Options{
		ReviewStarter:   starter,
		StartupLoader:   startupLoader,
		ReviewSubmitter: submitter,
	})
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "trail-hunk: %v\n", err)
		os.Exit(1)
	}
}
