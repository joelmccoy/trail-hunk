package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joelmccoy/trail-hunk/internal/app"
	"github.com/joelmccoy/trail-hunk/internal/review"
	"github.com/joelmccoy/trail-hunk/internal/tui"
)

func main() {
	cfg := app.ConfigFromEnv()
	_ = cfg

	if _, err := tea.NewProgram(tui.NewModel(review.ReviewSession{})).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "trail-hunk: %v\n", err)
		os.Exit(1)
	}
}
