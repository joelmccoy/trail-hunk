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
type ReviewSubmitter func(ctx context.Context, session review.ReviewSession) error

type Options struct {
	ReviewStarter   ReviewStarter
	StartupLoader   StartupLoader
	ReviewSubmitter ReviewSubmitter
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
	Submitting         bool
	Err                error
	starter            ReviewStarter
	startupLoader      StartupLoader
	submitter          ReviewSubmitter
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
		submitter:      opts.ReviewSubmitter,
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
		case keyComments:
			m.Screen = ScreenComments
		case keySubmitReview:
			if m.Submitting {
				return m, nil
			}
			if m.submitter == nil {
				m.Err = fmt.Errorf("review submission is not configured")
				return m, nil
			}
			if len(m.Session.ApprovedComments()) == 0 {
				m.Err = fmt.Errorf("no approved comments to submit")
				return m, nil
			}
			m.Submitting = true
			m.Err = nil
			return m, submitReviewCmd(m.submitter, m.Session)
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
	case reviewSubmittedMsg:
		m.Submitting = false
		if msg.Err != nil {
			m.Err = msg.Err
			return m, nil
		}
		m.Err = nil
		m.Session.MarkApprovedSubmitted()
		m.Screen = ScreenComments
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
		Width(contentWidth(m.Width, 2)).
		Padding(0, 1).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Render(header)

	footer = lipgloss.NewStyle().
		Width(contentWidth(m.Width, 2)).
		Padding(0, 1).
		Foreground(lipgloss.Color("244")).
		Render(footer)

	bodyHeight := m.Height - lipgloss.Height(header) - lipgloss.Height(footer) - 2
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	body = lipgloss.NewStyle().
		Width(contentWidth(m.Width, 4)).
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
	return "R review C queue S submit q quit n/p step j/k pick a ok d drop f files t ask"
}

type reviewStartedMsg struct {
	Session review.ReviewSession
	Err     error
}

type startupLoadedMsg struct {
	Startup review.StartupContext
	Err     error
}

type reviewSubmittedMsg struct {
	Err error
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

func submitReviewCmd(submitter ReviewSubmitter, session review.ReviewSession) tea.Cmd {
	return func() tea.Msg {
		return reviewSubmittedMsg{Err: submitter(context.Background(), session)}
	}
}

func renderStartup(m Model) string {
	var panels []func(int) string
	if m.StartupLoading {
		panels = append(panels, func(width int) string {
			return statusPanel("Detecting PR", "Checking local git context and GitHub pull requests.", width, lipgloss.Color("63"))
		})
	}
	if m.Startup.Repo.Owner != "" {
		panels = append(panels, func(width int) string {
			return infoPanel("Repository", []string{
				fmt.Sprintf("%s/%s", m.Startup.Repo.Owner, m.Startup.Repo.Name),
				fmt.Sprintf("branch: %s", m.Startup.Repo.Branch),
			}, width)
		})
	}
	if m.Startup.PR != nil {
		lines := []string{fmt.Sprintf("#%d %s", m.Startup.PR.Number, m.Startup.PR.Title)}
		if m.Startup.PR.State != "" {
			lines = append(lines, fmt.Sprintf("state: %s", m.Startup.PR.State))
		}
		if m.Startup.PR.URL != "" {
			lines = append(lines, m.Startup.PR.URL)
		}
		body := strings.Join(lines, "\n")
		panels = append(panels, func(width int) string {
			return statusPanel("Current PR", body, width, lipgloss.Color("36"))
		})
	} else if m.Startup.Message != "" {
		panels = append(panels, func(width int) string {
			return statusPanel("No PR Found", m.Startup.Message, width, lipgloss.Color("178"))
		})
	}
	if m.Loading {
		panels = append(panels, func(width int) string {
			return statusPanel("Generating Review", "Resolving git, GitHub PR context, diff, and AI walkthrough.", width, lipgloss.Color("63"))
		})
		return renderPanelGrid(m.Width, panels)
	}

	panels = append(panels, func(width int) string {
		return infoPanel("Next", []string{
			"Press R to initiate a guided review.",
			"Configure provider with TRAIL_HUNK_PROVIDER and TRAIL_HUNK_MODEL.",
		}, width)
	})
	if m.StartupErr != nil {
		panels = append(panels, func(width int) string {
			return statusPanel("Startup Error", m.StartupErr.Error(), width, lipgloss.Color("203"))
		})
	}
	if m.Err != nil {
		panels = append(panels, func(width int) string {
			return statusPanel("Review Error", m.Err.Error(), width, lipgloss.Color("203"))
		})
	}
	return renderPanelGrid(m.Width, panels)
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
	panelWidth := panelWidth(m.Width)

	var lines []string
	if m.Submitting {
		lines = append(lines, "Submitting approved comments to GitHub...")
	}
	if len(approved) == 0 {
		lines = append(lines, "No approved comments yet.")
	} else {
		lines = append(lines, fmt.Sprintf("%d approved comments ready.", len(approved)))
		for _, comment := range approved {
			target := comment.FilePath
			if comment.Line > 0 {
				target = fmt.Sprintf("%s:%d", comment.FilePath, comment.Line)
			}
			lines = append(lines, fmt.Sprintf("- [%s] %s — %s", comment.Priority, target, comment.Body))
		}
	}
	if m.Err != nil {
		lines = append(lines, "", "error: "+m.Err.Error())
	}
	lines = append(lines, "", "Press S to submit approved comments.")
	return infoPanel("Comment Queue", lines, panelWidth)
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

func contentWidth(totalWidth int, horizontalPadding int) int {
	width := totalWidth - horizontalPadding
	if width < 1 {
		return 1
	}
	return width
}

func panelWidth(totalWidth int) int {
	width := contentWidth(totalWidth, 16)
	if width <= 0 {
		return 64
	}
	if width > 68 {
		return 68
	}
	return width
}

func renderPanelGrid(totalWidth int, panels []func(int) string) string {
	if len(panels) == 0 {
		return ""
	}

	if totalWidth >= 120 && len(panels) > 1 {
		availableWidth := contentWidth(totalWidth, 8)
		gap := 4
		columnTotalWidth := (availableWidth - gap) / 2
		panelContentWidth := contentWidth(columnTotalWidth, 4)
		if panelContentWidth < 24 {
			panelContentWidth = 24
		}

		var rows []string
		for i := 0; i < len(panels); i += 2 {
			left := panels[i](panelContentWidth)
			if i+1 >= len(panels) {
				rows = append(rows, left)
				continue
			}
			right := panels[i+1](panelContentWidth)
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right))
		}
		return strings.Join(rows, "\n\n")
	}

	panelWidth := panelWidth(totalWidth)
	var rendered []string
	for _, panel := range panels {
		rendered = append(rendered, panel(panelWidth))
	}
	return strings.Join(rendered, "\n\n")
}

func infoPanel(title string, lines []string, width int) string {
	return panelStyle(width, lipgloss.Color("62")).Render(panelContent(title, strings.Join(lines, "\n")))
}

func statusPanel(title string, body string, width int, accent lipgloss.Color) string {
	return panelStyle(width, accent).Render(panelContent(title, body))
}

func panelStyle(width int, accent lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1)
}

func panelContent(title string, body string) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230"))
	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	return titleStyle.Render(title) + "\n" + bodyStyle.Render(body)
}
