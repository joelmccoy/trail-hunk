package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type FocusPane string

const (
	FocusRail      FocusPane = "rail"
	FocusDiff      FocusPane = "diff"
	FocusInspector FocusPane = "inspector"
)

type WorkbenchModel struct {
	Focus     FocusPane
	Width     int
	Height    int
	Diff      viewport.Model
	Inspector viewport.Model
	Ask       textarea.Model
}

func NewWorkbenchModel() WorkbenchModel {
	ask := textarea.New()
	ask.Placeholder = "Ask about the current PR, file, hunk, or line"
	ask.Prompt = "> "
	ask.ShowLineNumbers = false
	ask.Focus()
	return WorkbenchModel{
		Focus:     FocusDiff,
		Diff:      viewport.New(0, 0),
		Inspector: viewport.New(0, 0),
		Ask:       ask,
	}
}

func (w WorkbenchModel) Update(msg tea.Msg) (WorkbenchModel, tea.Cmd) {
	var cmd tea.Cmd
	switch w.Focus {
	case FocusInspector:
		w.Inspector, cmd = w.Inspector.Update(msg)
	default:
		w.Diff, cmd = w.Diff.Update(msg)
	}
	return w, cmd
}

func (w WorkbenchModel) UpdateAsk(msg tea.Msg) (WorkbenchModel, tea.Cmd) {
	var cmd tea.Cmd
	w.Ask, cmd = w.Ask.Update(msg)
	return w, cmd
}

func (w *WorkbenchModel) SetSize(width int, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	w.Width = width
	w.Height = height
}

func (w *WorkbenchModel) Sync(session review.ReviewSession, selectedSuggestion int, showFiles bool) {
	step, ok := currentStep(session)
	if !ok {
		return
	}

	layout := workbenchLayout(w.Width, showFiles)
	w.Diff.Width = layout.DiffWidth
	w.Diff.Height = w.Height
	w.Diff.SetContent(renderDiffRows(step, selectedSuggestion, layout.DiffWidth))

	w.Inspector.Width = layout.InspectorWidth
	w.Inspector.Height = w.Height
	w.Inspector.SetContent(renderInspector(session, step, selectedSuggestion, layout.InspectorWidth))

	drawerWidth := w.Width
	if drawerWidth > 4 {
		drawerWidth -= 4
	}
	w.Ask.SetWidth(drawerWidth)
	w.Ask.SetHeight(3)
}

func (w WorkbenchModel) View(session review.ReviewSession, selectedSuggestion int, showFiles bool, showAsk bool) string {
	step, ok := currentStep(session)
	if !ok {
		return fillBlock("No review steps loaded.", w.Width, w.Height)
	}

	layout := workbenchLayout(w.Width, showFiles)
	segments := make([]string, 0, 5)
	if layout.ShowRail {
		segments = append(segments, renderRail(session, layout.RailWidth, w.Height))
		segments = append(segments, verticalSeparator(w.Height))
	}

	diff := w.Diff.View()
	if diff == "" {
		diff = renderDiffRows(step, selectedSuggestion, layout.DiffWidth)
	}
	segments = append(segments, pane(FocusDiff, w.Focus, "Diff "+step.FilePath, diff, layout.DiffWidth, w.Height))

	if layout.ShowInspector {
		segments = append(segments, verticalSeparator(w.Height))
		inspector := w.Inspector.View()
		if inspector == "" {
			inspector = renderInspector(session, step, selectedSuggestion, layout.InspectorWidth)
		}
		segments = append(segments, pane(FocusInspector, w.Focus, "guide", inspector, layout.InspectorWidth, w.Height))
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, segments...)
	if showAsk {
		body = overlayAskDrawer(body, w.Width, w.Height, w.Ask)
	}
	return forceBlock(body, w.Width, w.Height)
}

type layoutSpec struct {
	ShowRail       bool
	ShowInspector  bool
	RailWidth      int
	DiffWidth      int
	InspectorWidth int
}

func workbenchLayout(width int, showFiles bool) layoutSpec {
	if width < 1 {
		width = 1
	}
	if width < 96 {
		return layoutSpec{DiffWidth: width}
	}

	railWidth := 0
	if showFiles {
		railWidth = 32
	}
	inspectorWidth := 42
	separators := 1
	if railWidth > 0 {
		separators++
	}
	diffWidth := width - railWidth - inspectorWidth - separators
	if diffWidth < 40 {
		return layoutSpec{DiffWidth: width}
	}
	return layoutSpec{
		ShowRail:       railWidth > 0,
		ShowInspector:  true,
		RailWidth:      railWidth,
		DiffWidth:      diffWidth,
		InspectorWidth: inspectorWidth,
	}
}

func currentStep(session review.ReviewSession) (review.ReviewStep, bool) {
	if len(session.Plan.ReviewOrder) == 0 {
		return review.ReviewStep{}, false
	}
	index := session.Cursor.StepIndex
	if index < 0 || index >= len(session.Plan.ReviewOrder) {
		return review.ReviewStep{}, false
	}
	return session.Plan.ReviewOrder[index], true
}

func renderRail(session review.ReviewSession, width int, height int) string {
	var lines []string
	current, _ := currentStep(session)
	lines = append(lines, sectionTitle("Changed Files"))
	lines = append(lines, mutedText("f focus/toggle  j/k move  enter jump"))
	files := orderedStepFiles(session)
	for _, file := range files {
		filePrefix := "  "
		if file == current.FilePath {
			filePrefix = "▶ "
		}
		lines = append(lines, truncateCells(filePrefix+compactPath(file, width-2), width))
		lines = append(lines, mutedText(fmt.Sprintf("  %d findings", suggestionCountForFile(session, file))))
		for _, stepIndex := range stepIndexesForFile(session, file) {
			step := session.Plan.ReviewOrder[stepIndex]
			stepPrefix := "    "
			if stepIndex == session.Cursor.StepIndex {
				stepPrefix = "  ▶ "
			}
			lines = append(lines, truncateCells(stepPrefix+step.Title, width))
		}
		lines = append(lines, "")
	}
	lines = append(lines, sectionTitle("Review Progress"))
	lines = append(lines, fmt.Sprintf("%d/%d steps", session.Cursor.StepIndex+1, len(session.Plan.ReviewOrder)))
	lines = append(lines, fmt.Sprintf("%d approved", len(session.ApprovedComments())))
	return pane(FocusRail, FocusRail, "review map", strings.Join(lines, "\n"), width, height)
}

func orderedStepFiles(session review.ReviewSession) []string {
	seen := map[string]bool{}
	var files []string
	for index, step := range session.Plan.ReviewOrder {
		if step.FilePath == "" || seen[step.FilePath] {
			continue
		}
		seen[step.FilePath] = true
		files = append(files, session.Plan.ReviewOrder[index].FilePath)
	}
	return files
}

func stepIndexesForFile(session review.ReviewSession, file string) []int {
	var indexes []int
	for i, step := range session.Plan.ReviewOrder {
		if step.FilePath == file {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func suggestionCountForFile(session review.ReviewSession, file string) int {
	count := 0
	for _, step := range session.Plan.ReviewOrder {
		if step.FilePath == file {
			count += len(step.Suggestions)
		}
	}
	return count
}

func compactPath(path string, width int) string {
	if width < 1 || lipgloss.Width(path) <= width {
		return path
	}
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return truncateCells(path, width)
	}
	return truncateCells(parts[len(parts)-1], width)
}

func renderInspector(session review.ReviewSession, step review.ReviewStep, selectedSuggestion int, width int) string {
	var lines []string
	lines = append(lines, mutedText(fmt.Sprintf("step %d/%d", session.Cursor.StepIndex+1, len(session.Plan.ReviewOrder))))
	lines = append(lines, sectionTitle(step.Title))
	lines = append(lines, "", sectionTitle("What this chunk does"))
	if step.Summary != "" {
		lines = append(lines, step.Summary)
	} else {
		lines = append(lines, "This walkthrough step focuses on the displayed diff chunk.")
	}
	if step.Why != "" {
		lines = append(lines, "", sectionTitle("Why it matters"), step.Why)
	}
	lines = append(lines, "", sectionTitle("Used by / impact"))
	lines = append(lines, impactSummary(step))
	lines = append(lines, "", sectionTitle("How to review"))
	if len(step.Focus) > 0 {
		for _, item := range step.Focus {
			lines = append(lines, "- "+item)
		}
	} else {
		lines = append(lines, "- Read the changed lines in order.")
		lines = append(lines, "- Check whether the behavior matches the PR intent.")
	}
	approved := len(session.ApprovedComments())
	lines = append(lines, "", sectionTitle("Confidence"))
	lines = append(lines, confidenceSummary(step))
	if approved > 0 {
		lines = append(lines, "", sectionTitle("queue"), fmt.Sprintf("%d approved", approved))
	}
	return wrapLines(lines, width)
}

func impactSummary(step review.ReviewStep) string {
	if step.FilePath == "" {
		return "Impact depends on callers of this changed chunk."
	}
	if strings.Contains(step.FilePath, "github") {
		return "This code affects GitHub diff/comment mapping and can change review submission behavior."
	}
	if strings.Contains(step.FilePath, "app") {
		return "This code affects orchestration between git, GitHub context, AI output, and the TUI session."
	}
	return "Review callers that depend on this file; behavior here can affect every downstream use of the changed helper."
}

func confidenceSummary(step review.ReviewStep) string {
	if len(step.Suggestions) == 0 {
		return "Medium confidence. No inline findings are attached to this chunk, but behavior still needs review."
	}
	return fmt.Sprintf("Medium confidence. %d inline finding(s) need reviewer judgment before this chunk is considered clear.", len(step.Suggestions))
}

func selectedReviewComment(suggestions []review.ReviewComment, selected int) (review.ReviewComment, bool) {
	if len(suggestions) == 0 || selected < 0 || selected >= len(suggestions) {
		return review.ReviewComment{}, false
	}
	return suggestions[selected], true
}

func targetText(comment review.ReviewComment) string {
	if comment.Line <= 0 {
		return comment.FilePath
	}
	side := comment.Side
	if side == "" {
		side = "RIGHT"
	}
	return fmt.Sprintf("%s:%d %s", comment.FilePath, comment.Line, side)
}

func pane(kind FocusPane, focused FocusPane, title string, content string, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	if kind == focused {
		titleStyle = titleStyle.Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
	}
	header := titleStyle.Render(truncateCells(" "+title+" ", width))
	bodyHeight := height - 1
	if bodyHeight < 0 {
		bodyHeight = 0
	}
	body := fitContent(content, width, bodyHeight)
	return forceBlock(lipgloss.JoinVertical(lipgloss.Left, header, body), width, height)
}

func sectionTitle(text string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Render(text)
}

func mutedText(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(text)
}

func verticalSeparator(height int) string {
	if height < 1 {
		return ""
	}
	return strings.Repeat("│\n", height-1) + "│"
}

func overlayAskDrawer(body string, width int, height int, ask textarea.Model) string {
	lines := splitBlock(body)
	drawerHeight := minInt(6, maxInt(3, height/3))
	start := maxInt(0, height-drawerHeight)
	drawer := []string{truncateCells(sectionTitle("ask current context"), width)}
	drawer = append(drawer, splitBlock(ask.View())...)
	for len(drawer) < drawerHeight {
		drawer = append(drawer, "")
	}
	for i := 0; i < drawerHeight && start+i < len(lines); i++ {
		lines[start+i] = drawer[i]
	}
	return strings.Join(lines, "\n")
}

func wrapLines(lines []string, width int) string {
	var wrapped []string
	for _, line := range lines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, splitBlock(wordWrap(line, width))...)
	}
	return strings.Join(wrapped, "\n")
}

func wordWrap(text string, width int) string {
	if width < 1 {
		return ""
	}
	var lines []string
	for _, sourceLine := range strings.Split(text, "\n") {
		words := strings.Fields(sourceLine)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		current := ""
		for _, word := range words {
			if current == "" {
				current = word
				continue
			}
			if lipgloss.Width(current+" "+word) <= width {
				current += " " + word
				continue
			}
			lines = append(lines, truncateCells(current, width))
			current = word
		}
		if current != "" {
			lines = append(lines, truncateCells(current, width))
		}
	}
	return strings.Join(lines, "\n")
}

func fitContent(content string, width int, height int) string {
	lines := splitBlock(content)
	if height < 1 {
		return ""
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	for i := range lines {
		lines[i] = truncateCells(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func forceBlock(content string, width int, height int) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(fitContent(content, width, height))
}

func fillBlock(text string, width int, height int) string {
	return forceBlock(text, width, height)
}

func splitBlock(content string) []string {
	if content == "" {
		return []string{""}
	}
	return strings.Split(content, "\n")
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
