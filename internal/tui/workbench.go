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
		railWidth = 28
	}
	inspectorWidth := 36
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
	lines = append(lines, sectionTitle("files"))
	seen := map[string]bool{}
	for index, step := range session.Plan.ReviewOrder {
		if seen[step.FilePath] {
			continue
		}
		seen[step.FilePath] = true
		prefix := "  "
		if index == session.Cursor.StepIndex {
			prefix = "> "
		}
		lines = append(lines, prefix+step.FilePath)
	}
	lines = append(lines, "", sectionTitle("steps"))
	for index, step := range session.Plan.ReviewOrder {
		prefix := "  "
		if index == session.Cursor.StepIndex {
			prefix = "> "
		}
		lines = append(lines, truncateCells(fmt.Sprintf("%s%s", prefix, step.Title), width))
	}
	return pane(FocusRail, FocusRail, "review map", strings.Join(lines, "\n"), width, height)
}

func renderInspector(session review.ReviewSession, step review.ReviewStep, selectedSuggestion int, width int) string {
	var lines []string
	lines = append(lines, sectionTitle(step.Title))
	if step.Summary != "" {
		lines = append(lines, step.Summary)
	}
	if step.Why != "" {
		lines = append(lines, "", sectionTitle("why"), step.Why)
	}
	if len(step.Focus) > 0 {
		lines = append(lines, "", sectionTitle("focus"))
		for _, item := range step.Focus {
			lines = append(lines, "- "+item)
		}
	}
	if suggestion, ok := selectedReviewComment(step.Suggestions, selectedSuggestion); ok {
		lines = append(lines, "", sectionTitle("Suggestions"))
		lines = append(lines, fmt.Sprintf("%s / %s / %s", suggestion.Priority, suggestion.Category, suggestion.Status))
		lines = append(lines, targetText(suggestion))
		lines = append(lines, "", suggestion.Body)
	}
	approved := len(session.ApprovedComments())
	if approved > 0 {
		lines = append(lines, "", sectionTitle("queue"), fmt.Sprintf("%d approved", approved))
	}
	return wrapLines(lines, width)
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
		for lipgloss.Width(line) > width {
			wrapped = append(wrapped, truncateCells(line, width))
			line = strings.TrimSpace(line[minInt(len(line), maxInt(1, minInt(len(line), width/2))):])
		}
		wrapped = append(wrapped, line)
	}
	return strings.Join(wrapped, "\n")
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
