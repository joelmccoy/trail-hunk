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
	if m.Width <= 0 {
		return strings.Join([]string{m.renderHeader(), m.renderBody(80), m.renderFooter(80)}, "\n")
	}

	header := renderHeaderBar(m.renderHeader(), m.Width)
	footer := m.renderFooter(m.Width)

	bodyHeight := m.Height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	bodyStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(bodyHeight).
		MaxHeight(bodyHeight).
		Padding(1, 2)
	if bodyHeight < 4 {
		bodyStyle = bodyStyle.Padding(0, 1)
	}

	body := bodyStyle.Render(m.renderBody(contentWidthForStyle(m.Width, bodyStyle)))

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func renderHeaderBar(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Render(truncateCells(" "+text, width))
}

func (m Model) renderFooter(width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("244")).
		Render(truncateCells(" "+footerText(width-1), width))
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

func (m Model) renderBody(width int) string {
	var b strings.Builder
	switch m.Screen {
	case ScreenStartup:
		b.WriteString(renderStartup(m, width))
	case ScreenOverview:
		b.WriteString(m.Session.Plan.Overview)
		b.WriteByte('\n')
	case ScreenWalkthrough:
		b.WriteString(renderWalkthrough(m, width))
	case ScreenComments:
		b.WriteString(renderComments(m, width))
	case ScreenSubmit:
		b.WriteString("Submit approved review comments\n")
	}
	return b.String()
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

func renderStartup(m Model, width int) string {
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
		return renderPanelGrid(width, panels)
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
	return renderPanelGrid(width, panels)
}

func renderWalkthrough(m Model, width int) string {
	if len(m.Session.Plan.ReviewOrder) == 0 {
		return "No review steps loaded.\n"
	}

	step := m.Session.Plan.ReviewOrder[m.Session.Cursor.StepIndex]
	var panels []func(int) string
	if m.ShowFileTree {
		panels = append(panels, func(width int) string {
			return infoPanel("Files", []string{step.FilePath}, width)
		})
	}

	overviewLines := []string{step.Title, "", step.Summary, "", "why: " + step.Why}
	if len(step.Focus) > 0 {
		overviewLines = append(overviewLines, "", "focus:")
		for _, item := range step.Focus {
			overviewLines = append(overviewLines, "- "+item)
		}
	}
	panels = append(panels, func(width int) string {
		return infoPanel("Step", overviewLines, width)
	})

	panels = append(panels, func(width int) string {
		return diffPanel(step, m.SelectedSuggestion, width)
	})

	if len(step.Suggestions) > 0 {
		panels = append(panels, func(width int) string {
			return suggestionsPanel(step.Suggestions, m.SelectedSuggestion, width)
		})
	}
	if m.ShowAskPane {
		panels = append(panels, func(width int) string {
			return statusPanel("Ask", "Ask pane placeholder for current step context.", width, lipgloss.Color("36"))
		})
	}
	return renderPanelGrid(width, panels)
}

func diffPanel(step review.ReviewStep, selectedSuggestion int, width int) string {
	var lines []string
	if step.FilePath != "" {
		lines = append(lines, step.FilePath)
	}
	if len(step.DiffLines) == 0 {
		lines = append(lines, "No diff lines mapped for this step.")
		return statusPanel("Diff", strings.Join(lines, "\n"), width, lipgloss.Color("178"))
	}
	lines = append(lines, diffHeader())
	targets := suggestionTargets(step.Suggestions, selectedSuggestion)
	for _, line := range focusedDiffLines(step.DiffLines, targets, 0) {
		if line.Text == omittedDiffText {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("..."))
			continue
		}
		lines = append(lines, renderDiffLine(line, targets))
	}
	return statusPanel("Diff", strings.Join(lines, "\n"), width, lipgloss.Color("63"))
}

const omittedDiffText = "__trail_hunk_omitted_diff_context__"

func focusedDiffLines(lines []review.DiffLine, targets map[string]diffTarget, contextRadius int) []review.DiffLine {
	if len(lines) == 0 || len(targets) == 0 {
		return lines
	}

	first := len(lines)
	last := -1
	for i, line := range lines {
		target := targetForDiffLine(line, targets)
		if target.Label == "" {
			continue
		}
		if i < first {
			first = i
		}
		if i > last {
			last = i
		}
	}
	if last == -1 {
		return lines
	}

	start := maxInt(0, first-contextRadius)
	end := last + contextRadius
	if end >= len(lines) {
		end = len(lines) - 1
	}

	var focused []review.DiffLine
	if start > 0 {
		focused = append(focused, omittedDiffLine())
	}
	focused = append(focused, lines[start:end+1]...)
	if end < len(lines)-1 {
		focused = append(focused, omittedDiffLine())
	}
	return focused
}

func omittedDiffLine() review.DiffLine {
	return review.DiffLine{Kind: review.DiffLineContext, Text: omittedDiffText}
}

func suggestionsPanel(suggestions []review.ReviewComment, selected int, width int) string {
	var lines []string
	for i, suggestion := range suggestions {
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		target := suggestion.FilePath
		if suggestion.Line > 0 {
			target = fmt.Sprintf("%s:%d", suggestion.FilePath, suggestion.Line)
		}
		lines = append(lines, fmt.Sprintf("%s[%s/%s] %s", prefix, suggestion.Priority, suggestion.Status, target))
		lines = append(lines, "  "+suggestion.Body)
	}
	return infoPanel("Suggestions", lines, width)
}

type diffTarget struct {
	Label   string
	Current bool
}

func suggestionTargets(suggestions []review.ReviewComment, selected int) map[string]diffTarget {
	targets := make(map[string]diffTarget)
	for i, suggestion := range suggestions {
		if suggestion.Line <= 0 {
			continue
		}
		side := suggestion.Side
		if side == "" {
			side = "RIGHT"
		}
		label := "note"
		if i == selected {
			label = "focus"
		}
		targets[diffTargetKey(side, suggestion.Line)] = diffTarget{
			Label:   label,
			Current: i == selected,
		}
	}
	return targets
}

func diffHeader() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("mark    old    new  code")
}

func renderDiffLine(line review.DiffLine, targets map[string]diffTarget) string {
	marker := " "
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	switch line.Kind {
	case review.DiffLineAdded:
		marker = "+"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case review.DiffLineDeleted:
		marker = "-"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	}

	target := targetForDiffLine(line, targets)
	row := fmt.Sprintf("%-5s %s  %s%s", target.Label, renderLineNumbers(line), marker, line.Text)
	if target.Current {
		style = style.Bold(true).Background(lipgloss.Color("236"))
	}
	return style.Render(row)
}

func renderLineNumbers(line review.DiffLine) string {
	oldLine := " "
	newLine := " "
	if line.OldLine != nil {
		oldLine = fmt.Sprintf("%d", *line.OldLine)
	}
	if line.NewLine != nil {
		newLine = fmt.Sprintf("%d", *line.NewLine)
	}
	return fmt.Sprintf("%4s %4s", oldLine, newLine)
}

func targetForDiffLine(line review.DiffLine, targets map[string]diffTarget) diffTarget {
	switch line.Kind {
	case review.DiffLineDeleted:
		if line.OldLine == nil {
			return diffTarget{}
		}
		return targets[diffTargetKey("LEFT", *line.OldLine)]
	default:
		if line.NewLine == nil {
			return diffTarget{}
		}
		return targets[diffTargetKey("RIGHT", *line.NewLine)]
	}
}

func diffTargetKey(side string, line int) string {
	return fmt.Sprintf("%s:%d", side, line)
}

func renderComments(m Model, width int) string {
	approved := m.Session.ApprovedComments()
	panelWidth := panelWidth(width)

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

func contentWidthForStyle(totalWidth int, style lipgloss.Style) int {
	width := totalWidth - style.GetHorizontalFrameSize()
	if width < 1 {
		return 1
	}
	return width
}

func panelWidth(totalWidth int) int {
	if totalWidth < 1 {
		return 1
	}
	return totalWidth
}

func renderPanelGrid(totalWidth int, panels []func(int) string) string {
	if len(panels) == 0 {
		return ""
	}

	if totalWidth >= 96 && len(panels) > 1 {
		availableWidth := totalWidth
		gap := 4
		leftWidth := (availableWidth - gap) / 2
		rightWidth := availableWidth - gap - leftWidth

		var rows []string
		for i := 0; i < len(panels); i += 2 {
			left := panels[i](leftWidth)
			if i+1 >= len(panels) {
				rows = append(rows, left)
				continue
			}
			right := panels[i+1](rightWidth)
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
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1)
	return style.Width(maxInt(1, width-style.GetHorizontalBorderSize()))
}

func footerText(width int) string {
	long := "R review C queue S submit q quit n/p step j/k pick a accept d dismiss f files t ask"
	if lipgloss.Width(long) <= width {
		return long
	}

	compact := "R review C queue S submit q quit n/p j/k a ok d drop f files t ask"
	if lipgloss.Width(compact) <= width {
		return compact
	}

	minimal := "R rev C q S sub q quit n/p j/k a d f t"
	if lipgloss.Width(minimal) <= width {
		return minimal
	}

	essential := "R rev q quit n/p j/k t"
	return truncateCells(essential, width)
}

func truncateCells(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= width {
		return text
	}
	if width <= 3 {
		var b strings.Builder
		for _, r := range text {
			if lipgloss.Width(b.String()+string(r)) > width {
				break
			}
			b.WriteRune(r)
		}
		return b.String()
	}

	limit := width - 3
	var b strings.Builder
	for _, r := range text {
		next := b.String() + string(r)
		if lipgloss.Width(next) > limit {
			break
		}
		b.WriteRune(r)
	}
	return b.String() + "..."
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func panelContent(title string, body string) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230"))
	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	return titleStyle.Render(title) + "\n" + bodyStyle.Render(body)
}
