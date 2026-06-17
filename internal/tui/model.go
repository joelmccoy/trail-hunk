package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
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
	Workbench          WorkbenchModel
	Loading            bool
	Submitting         bool
	Err                error
	keys               keyMap
	help               help.Model
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
	keys := defaultKeyMap()
	return Model{
		Screen:         ScreenStartup,
		Session:        session,
		FocusedPane:    "diff",
		ShowFileTree:   true,
		Workbench:      NewWorkbenchModel(),
		keys:           keys,
		help:           newHelpModel(),
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
		case keyToggleAskPane:
			m.ShowAskPane = !m.ShowAskPane
			if m.ShowAskPane {
				cmd := m.Workbench.Ask.Focus()
				return m, cmd
			}
			return m, nil
		}
		if m.ShowAskPane && m.Screen == ScreenWalkthrough {
			var cmd tea.Cmd
			m.Workbench, cmd = m.Workbench.UpdateAsk(msg)
			return m, cmd
		}
		switch msg.String() {
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
			m.SelectedSuggestion = 0
			m.Screen = ScreenWalkthrough
		case keyPreviousStep:
			m.Session.PreviousStep()
			m.SelectedSuggestion = 0
			m.Screen = ScreenWalkthrough
		case keyNextFile:
			m.selectFile(1)
			m.Screen = ScreenWalkthrough
		case keyPreviousFile:
			m.selectFile(-1)
			m.Screen = ScreenWalkthrough
		case keyToggleFiles:
			m.ShowFileTree = !m.ShowFileTree
		case keyMoveDown, "down", keyMoveUp, "up":
			if m.Screen == ScreenWalkthrough {
				m.syncWorkbench()
				var cmd tea.Cmd
				m.Workbench, cmd = m.Workbench.Update(msg)
				return m, cmd
			}
		case keyNextSuggestion:
			m.selectSuggestion(1)
		case keyPreviousSuggestion:
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

func (m *Model) syncWorkbench() {
	if m.Screen != ScreenWalkthrough {
		return
	}
	if m.Workbench.Width <= 0 || m.Workbench.Height <= 0 {
		width := maxInt(1, m.Width-4)
		height := maxInt(1, m.Height-4)
		m.Workbench.SetSize(width, height)
	}
	m.Workbench.Sync(m.Session, m.SelectedSuggestion, m.ShowFileTree)
}

func (m Model) View() string {
	if m.Width <= 0 {
		return strings.Join([]string{m.renderHeader(), m.renderBody(80, 20), m.renderFooter(80)}, "\n")
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

	bodyWidth := contentWidthForStyle(m.Width, bodyStyle)
	bodyHeight = contentHeightForStyle(bodyHeight, bodyStyle)
	body := bodyStyle.Render(m.renderBody(bodyWidth, bodyHeight))

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
	helpText := m.help.View(m.keys)
	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("244")).
		Render(truncateCells(" "+helpText, width))
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

func (m Model) renderBody(width int, height int) string {
	var b strings.Builder
	switch m.Screen {
	case ScreenStartup:
		b.WriteString(renderStartup(m, width))
	case ScreenOverview:
		b.WriteString(m.Session.Plan.Overview)
		b.WriteByte('\n')
	case ScreenWalkthrough:
		b.WriteString(renderWalkthrough(m, width, height))
	case ScreenComments:
		b.WriteString(renderComments(m, width, height))
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
	var lines []string
	if m.StartupLoading {
		lines = append(lines, startupSection("Detecting PR", []string{"Checking local git context and GitHub pull requests."})...)
	}
	if m.Startup.Repo.Owner != "" {
		lines = append(lines, startupSection("Repository", []string{
			fmt.Sprintf("%s/%s", m.Startup.Repo.Owner, m.Startup.Repo.Name),
			fmt.Sprintf("branch: %s", m.Startup.Repo.Branch),
		})...)
	}
	if m.Startup.PR != nil {
		prLines := []string{fmt.Sprintf("#%d %s", m.Startup.PR.Number, m.Startup.PR.Title)}
		if m.Startup.PR.State != "" {
			prLines = append(prLines, fmt.Sprintf("state: %s", m.Startup.PR.State))
		}
		if m.Startup.PR.URL != "" {
			prLines = append(prLines, m.Startup.PR.URL)
		}
		lines = append(lines, startupSection("Current PR", prLines)...)
	} else if m.Startup.Message != "" {
		lines = append(lines, startupSection("No PR Found", []string{m.Startup.Message})...)
	}
	if m.Loading {
		lines = append(lines, startupSection("Generating Review", []string{"Resolving git, GitHub PR context, diff, and AI walkthrough."})...)
		return wrapLines(lines, width)
	}

	lines = append(lines, startupSection("Next", []string{
		"Press R to initiate a guided review.",
		"Configure provider with TRAIL_HUNK_PROVIDER and TRAIL_HUNK_MODEL.",
	})...)
	if m.StartupErr != nil {
		lines = append(lines, startupSection("Startup Error", []string{m.StartupErr.Error()})...)
	}
	if m.Err != nil {
		lines = append(lines, startupSection("Review Error", []string{m.Err.Error()})...)
	}
	return wrapLines(lines, width)
}

func startupSection(title string, body []string) []string {
	lines := []string{sectionTitle(title)}
	lines = append(lines, body...)
	lines = append(lines, "")
	return lines
}

func renderWalkthrough(m Model, width int, height int) string {
	if len(m.Session.Plan.ReviewOrder) == 0 {
		return "No review steps loaded.\n"
	}
	workbench := m.Workbench
	workbench.SetSize(width, height)
	workbench.Sync(m.Session, m.SelectedSuggestion, m.ShowFileTree)
	return workbench.View(m.Session, m.SelectedSuggestion, m.ShowFileTree, m.ShowAskPane)
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

func renderComments(m Model, width int, height int) string {
	approved := m.Session.ApprovedComments()

	var lines []string
	lines = append(lines, sectionTitle("Comment Queue"))
	if m.Submitting {
		lines = append(lines, "Submitting approved comments to GitHub...")
	}
	if len(approved) == 0 {
		lines = append(lines, "No approved comments yet.")
	} else {
		lines = append(lines, fmt.Sprintf("%d approved comments ready.", len(approved)))
		for _, comment := range approved {
			lines = append(lines, fmt.Sprintf("- [%s/%s] %s - %s", comment.Priority, comment.Status, targetText(comment), comment.Body))
		}
	}
	if m.Err != nil {
		lines = append(lines, "", "error: "+m.Err.Error())
	}
	lines = append(lines, "", "Press S to submit approved comments.")
	return forceBlock(wrapLines(lines, width), width, height)
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

func (m *Model) selectFile(delta int) {
	if len(m.Session.Plan.ReviewOrder) == 0 {
		return
	}
	currentFile := m.Session.Plan.ReviewOrder[m.Session.Cursor.StepIndex].FilePath
	files := orderedStepFiles(m.Session)
	if len(files) == 0 {
		return
	}

	currentFileIndex := 0
	for i, file := range files {
		if file == currentFile {
			currentFileIndex = i
			break
		}
	}
	nextFileIndex := currentFileIndex + delta
	if nextFileIndex < 0 {
		nextFileIndex = len(files) - 1
	}
	if nextFileIndex >= len(files) {
		nextFileIndex = 0
	}
	targetFile := files[nextFileIndex]
	for i, step := range m.Session.Plan.ReviewOrder {
		if step.FilePath == targetFile {
			m.Session.Cursor.StepIndex = i
			m.SelectedSuggestion = 0
			return
		}
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

func contentHeightForStyle(totalHeight int, style lipgloss.Style) int {
	height := totalHeight - style.GetVerticalFrameSize()
	if height < 1 {
		return 1
	}
	return height
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
