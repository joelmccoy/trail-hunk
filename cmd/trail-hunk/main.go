package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (model) Init() tea.Cmd {
	return nil
}

func (model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model{}, tea.Quit
		}
	}

	return model{}, nil
}

func (model) View() string {
	return "trail-hunk\n\nAI-assisted GitHub PR review TUI.\n\nPress q to quit.\n"
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "trail-hunk: %v\n", err)
		os.Exit(1)
	}
}
