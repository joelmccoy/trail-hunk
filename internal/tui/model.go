package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type ReviewStarter func(ctx context.Context) (review.ReviewSession, error)

type Model struct {
	Screen       Screen
	Session      review.ReviewSession
	FocusedPane  string
	Width        int
	Height       int
	ShowFileTree bool
	ShowAskPane  bool
	Loading      bool
	Err          error
	starter      ReviewStarter
}

func NewModel(session review.ReviewSession) Model {
	return NewModelWithStarter(session, nil)
}

func NewModelWithStarter(session review.ReviewSession, starter ReviewStarter) Model {
	return Model{
		Screen:      ScreenStartup,
		Session:     session,
		FocusedPane: "diff",
		starter:     starter,
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
		case keyStartReview:
			if m.Loading {
				return m, nil
			}
			if m.starter == nil {
				m.Err = fmt.Errorf("review startup is not configured")
				return m, nil
			}
			m.Loading = true
			m.Err = nil
			m.Screen = ScreenStartup
			return m, startReviewCmd(m.starter)
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
	case reviewStartedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Screen = ScreenStartup
			return m, nil
		}
		m.Err = nil
		m.Session = msg.Session
		if len(m.Session.Plan.ReviewOrder) > 0 {
			m.Screen = ScreenWalkthrough
		} else {
			m.Screen = ScreenOverview
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
		b.WriteString(renderStartup(m))
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

	b.WriteString("\nkeys: R review | q quit | n/p step | f files | t ask\n")
	return b.String()
}

type reviewStartedMsg struct {
	Session review.ReviewSession
	Err     error
}

func startReviewCmd(starter ReviewStarter) tea.Cmd {
	return func() tea.Msg {
		session, err := starter(context.Background())
		return reviewStartedMsg{Session: session, Err: err}
	}
}

func renderStartup(m Model) string {
	var b strings.Builder
	if m.Loading {
		b.WriteString("Generating guided review...\n")
		b.WriteString("Resolving git, GitHub PR context, diff, and AI walkthrough.\n")
		return b.String()
	}

	b.WriteString("Press R to initiate a guided review for the current GitHub pull request.\n")
	b.WriteString("Provider is configured with TRAIL_HUNK_PROVIDER and TRAIL_HUNK_MODEL.\n")
	if m.Err != nil {
		b.WriteString("\nerror: ")
		b.WriteString(m.Err.Error())
		b.WriteByte('\n')
	}
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
