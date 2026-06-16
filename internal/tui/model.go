package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type Model struct {
	Screen       Screen
	Session      review.ReviewSession
	FocusedPane  string
	Width        int
	Height       int
	ShowFileTree bool
	ShowAskPane  bool
	Err          error
}

func NewModel(session review.ReviewSession) Model {
	return Model{
		Screen:      ScreenStartup,
		Session:     session,
		FocusedPane: "diff",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case keyQuit, "ctrl+c":
			return m, tea.Quit
		case keyNextStep:
			m.Session.NextStep()
			m.Screen = ScreenWalkthrough
		case keyPreviousStep:
			m.Session.PreviousStep()
			m.Screen = ScreenWalkthrough
		case keyToggleFiles:
			m.ShowFileTree = !m.ShowFileTree
		case keyToggleAskPane:
			m.ShowAskPane = !m.ShowAskPane
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("trail-hunk\n")
	b.WriteString(fmt.Sprintf("screen: %s", m.Screen))
	if len(m.Session.Plan.ReviewOrder) > 0 {
		step := m.Session.Plan.ReviewOrder[m.Session.Cursor.StepIndex]
		b.WriteString(fmt.Sprintf(" | step: %s", step.ID))
	}
	b.WriteString("\n\n")

	switch m.Screen {
	case ScreenStartup:
		b.WriteString("Resolving repository and pull request context...\n")
	case ScreenOverview:
		b.WriteString(m.Session.Plan.Overview)
		b.WriteByte('\n')
	case ScreenWalkthrough:
		b.WriteString(renderWalkthrough(m))
	case ScreenComments:
		b.WriteString(renderComments(m))
	case ScreenSubmit:
		b.WriteString("Submit approved review comments\n")
	}

	b.WriteString("\nkeys: q quit | n/p step | f files | t ask\n")
	return b.String()
}

func renderWalkthrough(m Model) string {
	if len(m.Session.Plan.ReviewOrder) == 0 {
		return "No review steps loaded.\n"
	}

	step := m.Session.Plan.ReviewOrder[m.Session.Cursor.StepIndex]
	var b strings.Builder
	if m.ShowFileTree {
		b.WriteString("[files]\n")
	}
	b.WriteString(step.Title)
	b.WriteByte('\n')
	b.WriteString(step.Summary)
	b.WriteByte('\n')
	b.WriteString("why: ")
	b.WriteString(step.Why)
	b.WriteByte('\n')
	if m.ShowAskPane {
		b.WriteString("[ask pane]\n")
	}
	return b.String()
}

func renderComments(m Model) string {
	approved := m.Session.ApprovedComments()
	return fmt.Sprintf("%d approved comments\n", len(approved))
}
