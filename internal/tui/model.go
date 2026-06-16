package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type ReviewStarter func(ctx context.Context) (review.ReviewSession, error)
type StartupLoader func(ctx context.Context) (review.StartupContext, error)

type Options struct {
	ReviewStarter ReviewStarter
	StartupLoader StartupLoader
}

type Model struct {
	Screen             Screen
	Session            review.ReviewSession
	Startup            review.StartupContext
	StartupLoading     bool
	StartupErr         error
	FocusedPane        string
	Width              int
	Height             int
	ShowFileTree       bool
	ShowAskPane        bool
	SelectedSuggestion int
	Loading            bool
	Err                error
	starter            ReviewStarter
	startupLoader      StartupLoader
}

func NewModel(session review.ReviewSession) Model {
	return NewModelWithStarter(session, nil)
}

func NewModelWithStarter(session review.ReviewSession, starter ReviewStarter) Model {
	return NewModelWithOptions(session, Options{ReviewStarter: starter})
}

func NewModelWithOptions(session review.ReviewSession, opts Options) Model {
	return Model{
		Screen:         ScreenStartup,
		Session:        session,
		FocusedPane:    "diff",
		starter:        opts.ReviewStarter,
		startupLoader:  opts.StartupLoader,
		StartupLoading: opts.StartupLoader != nil,
	}
}

func (m Model) Init() tea.Cmd {
	if m.startupLoader == nil {
		return nil
	}
	return loadStartupCmd(m.startupLoader)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case startupLoadedMsg:
		m.StartupLoading = false
		if msg.Err != nil {
			m.StartupErr = msg.Err
			return m, nil
		}
		m.StartupErr = nil
		m.Startup = msg.Startup
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
		case keySelectNext:
			m.selectSuggestion(1)
		case keySelectPrevious:
			m.selectSuggestion(-1)
		case keyAcceptComment:
			m.updateSelectedSuggestion(review.StatusApproved)
		case keyDismissComment:
			m.updateSelectedSuggestion(review.StatusDismissed)
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
		m.SelectedSuggestion = 0
		if len(m.Session.Plan.ReviewOrder) > 0 {
			m.Screen = ScreenWalkthrough
		} else {
			m.Screen = ScreenOverview
		}
	}

	return m, nil
}

func (m Model) View() string {
	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	if m.Width <= 0 {
		return strings.Join([]string{header, body, footer}, "\n")
	}

	header = lipgloss.NewStyle().
		Width(m.Width).
		Padding(0, 1).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Render(header)

	footer = lipgloss.NewStyle().
		Width(m.Width).
		Padding(0, 1).
		Foreground(lipgloss.Color("244")).
		Render(footer)

	bodyHeight := m.Height - lipgloss.Height(header) - lipgloss.Height(footer) - 2
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	body = lipgloss.NewStyle().
		Width(m.Width).
		Height(bodyHeight).
		Padding(1, 2).
		Render(body)

	return strings.Join([]string{header, body, footer}, "\n")
}

func (m Model) renderHeader() string {
	var b strings.Builder
	b.WriteString("trail-hunk")
	b.WriteString(fmt.Sprintf("  %s", m.Screen))
	if len(m.Session.Plan.ReviewOrder) > 0 {
		step := m.Session.Plan.ReviewOrder[m.Session.Cursor.StepIndex]
		b.WriteString(fmt.Sprintf("  step:%s", step.ID))
	}
	if m.Startup.Repo.Owner != "" {
		b.WriteString(fmt.Sprintf("  %s/%s:%s", m.Startup.Repo.Owner, m.Startup.Repo.Name, m.Startup.Repo.Branch))
	}
	return b.String()
}

func (m Model) renderBody() string {
	var b strings.Builder
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
	return b.String()
}

func (m Model) renderFooter() string {
	return "R review | q quit | n/p step | j/k select | a accept | d dismiss | f files | t ask"
}

type reviewStartedMsg struct {
	Session review.ReviewSession
	Err     error
}

type startupLoadedMsg struct {
	Startup review.StartupContext
	Err     error
}

func loadStartupCmd(loader StartupLoader) tea.Cmd {
	return func() tea.Msg {
		startup, err := loader(context.Background())
		return startupLoadedMsg{Startup: startup, Err: err}
	}
}

func startReviewCmd(starter ReviewStarter) tea.Cmd {
	return func() tea.Msg {
		session, err := starter(context.Background())
		return reviewStartedMsg{Session: session, Err: err}
	}
}

func renderStartup(m Model) string {
	var b strings.Builder
	if m.StartupLoading {
		b.WriteString("Detecting current GitHub pull request...\n\n")
	}
	if m.Startup.Repo.Owner != "" {
		b.WriteString("Repository\n")
		b.WriteString(fmt.Sprintf("  %s/%s\n", m.Startup.Repo.Owner, m.Startup.Repo.Name))
		b.WriteString(fmt.Sprintf("  branch: %s\n\n", m.Startup.Repo.Branch))
	}
	if m.Startup.PR != nil {
		b.WriteString("Current PR\n")
		b.WriteString(fmt.Sprintf("  #%d %s\n", m.Startup.PR.Number, m.Startup.PR.Title))
		if m.Startup.PR.State != "" {
			b.WriteString(fmt.Sprintf("  state: %s\n", m.Startup.PR.State))
		}
		if m.Startup.PR.URL != "" {
			b.WriteString(fmt.Sprintf("  %s\n", m.Startup.PR.URL))
		}
		b.WriteByte('\n')
	} else if m.Startup.Message != "" {
		b.WriteString(m.Startup.Message)
		b.WriteString("\n\n")
	}
	if m.Loading {
		b.WriteString("Generating guided review...\n")
		b.WriteString("Resolving git, GitHub PR context, diff, and AI walkthrough.\n")
		return b.String()
	}

	b.WriteString("Press R to initiate a guided review for the current GitHub pull request.\n")
	b.WriteString("Provider is configured with TRAIL_HUNK_PROVIDER and TRAIL_HUNK_MODEL.\n")
	if m.StartupErr != nil {
		b.WriteString("\nstartup error: ")
		b.WriteString(m.StartupErr.Error())
		b.WriteByte('\n')
	}
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
	if len(step.Suggestions) > 0 {
		b.WriteString("\nsuggestions:\n")
		for i, suggestion := range step.Suggestions {
			prefix := "  "
			if i == m.SelectedSuggestion {
				prefix = "> "
			}
			b.WriteString(prefix)
			b.WriteString("[")
			b.WriteString(string(suggestion.Status))
			b.WriteString("] ")
			b.WriteString(suggestion.Body)
			b.WriteByte('\n')
		}
	}
	if m.ShowAskPane {
		b.WriteString("[ask pane]\n")
	}
	return b.String()
}

func renderComments(m Model) string {
	approved := m.Session.ApprovedComments()
	return fmt.Sprintf("%d approved comments\n", len(approved))
}

func (m *Model) selectSuggestion(delta int) {
	suggestions := m.currentSuggestions()
	if len(suggestions) == 0 {
		m.SelectedSuggestion = 0
		return
	}

	m.SelectedSuggestion += delta
	if m.SelectedSuggestion < 0 {
		m.SelectedSuggestion = len(suggestions) - 1
	}
	if m.SelectedSuggestion >= len(suggestions) {
		m.SelectedSuggestion = 0
	}
}

func (m *Model) updateSelectedSuggestion(status review.CommentStatus) {
	suggestions := m.currentSuggestions()
	if len(suggestions) == 0 || m.SelectedSuggestion >= len(suggestions) {
		return
	}

	commentID := suggestions[m.SelectedSuggestion].ID
	var err error
	switch status {
	case review.StatusApproved:
		err = m.Session.AcceptSuggestion(commentID)
	case review.StatusDismissed:
		err = m.Session.DismissSuggestion(commentID)
	default:
		err = fmt.Errorf("unsupported suggestion status %q", status)
	}
	if err != nil {
		m.Err = err
		return
	}

	m.syncStepSuggestionStatus(commentID, status)
}

func (m Model) currentSuggestions() []review.ReviewComment {
	if len(m.Session.Plan.ReviewOrder) == 0 {
		return nil
	}
	stepIndex := m.Session.Cursor.StepIndex
	if stepIndex < 0 || stepIndex >= len(m.Session.Plan.ReviewOrder) {
		return nil
	}
	return m.Session.Plan.ReviewOrder[stepIndex].Suggestions
}

func (m *Model) syncStepSuggestionStatus(commentID string, status review.CommentStatus) {
	stepIndex := m.Session.Cursor.StepIndex
	if stepIndex < 0 || stepIndex >= len(m.Session.Plan.ReviewOrder) {
		return
	}
	for i := range m.Session.Plan.ReviewOrder[stepIndex].Suggestions {
		if m.Session.Plan.ReviewOrder[stepIndex].Suggestions[i].ID == commentID {
			m.Session.Plan.ReviewOrder[stepIndex].Suggestions[i].Status = status
			return
		}
	}
}
