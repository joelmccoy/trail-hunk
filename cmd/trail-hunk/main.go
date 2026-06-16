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

	if _, err := tea.NewProgram(tui.NewModelWithStarter(review.ReviewSession{}, starter)).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "trail-hunk: %v\n", err)
		os.Exit(1)
	}
}
