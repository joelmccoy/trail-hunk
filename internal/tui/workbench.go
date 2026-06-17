package tui

import (
	"fmt"
	"strconv"
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

func (w *WorkbenchModel) Sync(session review.ReviewSession, selectedSuggestion int, showFiles bool, viewedFiles map[string]bool, focusMode bool) {
	step, ok := currentStep(session)
	if !ok {
		return
	}

	layout := workbenchLayout(w.Width, showFiles)
	if focusMode {
		layout = layoutSpec{DiffWidth: w.Width}
	}
	drawerHeight := reviewDrawerHeight(step, w.Height)
	insightHeight := assistantInsightHeight(step, layout.DiffWidth)
	if focusMode {
		insightHeight = 0
	}
	diffHeight := maxInt(1, w.Height-drawerHeight-insightHeight)
	w.Diff.Width = layout.DiffWidth
	w.Diff.Height = diffHeight
	w.Diff.SetContent(renderDiffRows(step, selectedSuggestion, layout.DiffWidth))

	drawerWidth := w.Width
	if drawerWidth > 4 {
		drawerWidth -= 4
	}
	w.Ask.SetWidth(drawerWidth)
	w.Ask.SetHeight(3)
}

func (w WorkbenchModel) View(session review.ReviewSession, selectedSuggestion int, showFiles bool, showAsk bool, viewedFiles map[string]bool, focusMode bool) string {
	step, ok := currentStep(session)
	if !ok {
		return fillBlock("No review steps loaded.", w.Width, w.Height)
	}

	layout := workbenchLayout(w.Width, showFiles)
	if focusMode {
		layout = layoutSpec{DiffWidth: w.Width}
	}
	drawerHeight := reviewDrawerHeight(step, w.Height)
	insightHeight := assistantInsightHeight(step, layout.DiffWidth)
	if focusMode {
		insightHeight = 0
	}
	diffHeight := maxInt(1, w.Height-drawerHeight-insightHeight)
	segments := make([]string, 0, 5)
	if layout.ShowRail {
		segments = append(segments, renderRail(session, layout.RailWidth, w.Height, viewedFiles, w.Focus))
		segments = append(segments, verticalSeparator(w.Height))
	}

	diff := w.Diff.View()
	if diff == "" {
		diff = renderDiffRows(step, selectedSuggestion, layout.DiffWidth)
	}
	rightWidth := layout.DiffWidth
	rightTop := lipgloss.JoinVertical(lipgloss.Left,
		renderAssistantInsight(session, step, rightWidth, insightHeight),
		pane(FocusDiff, w.Focus, "DIFF "+compactPath(step.FilePath, maxInt(1, rightWidth-5)), diff, rightWidth, diffHeight),
	)
	rightStack := rightTop
	if drawerHeight > 0 {
		rightStack = lipgloss.JoinVertical(lipgloss.Left, rightTop, renderReviewDrawer(session, step, selectedSuggestion, rightWidth, drawerHeight))
	}
	segments = append(segments, forceBlock(rightStack, rightWidth, w.Height))

	body := lipgloss.JoinHorizontal(lipgloss.Top, segments...)
	if showAsk {
		body = overlayAskDrawer(body, w.Width, w.Height, w.Ask)
	}
	return forceBlock(body, w.Width, w.Height)
}

type layoutSpec struct {
	ShowRail  bool
	RailWidth int
	DiffWidth int
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
		railWidth = clampInt(width/6, 26, 32)
	}
	separators := 0
	if railWidth > 0 {
		separators = 1
	}
	diffWidth := width - railWidth - separators
	if diffWidth < 40 {
		return layoutSpec{DiffWidth: width}
	}
	return layoutSpec{
		ShowRail:  railWidth > 0,
		RailWidth: railWidth,
		DiffWidth: diffWidth,
	}
}

func reviewDrawerHeight(step review.ReviewStep, totalHeight int) int {
	if len(step.Suggestions) == 0 || totalHeight < 18 {
		return 0
	}
	return 6
}

func assistantInsightHeight(step review.ReviewStep, width int) int {
	if width < 90 {
		return 2
	}
	return 3
}

func renderAssistantInsight(session review.ReviewSession, step review.ReviewStep, width int, height int) string {
	title := walkthroughStepTitle(step)
	progress := "step " + itoa(session.Cursor.StepIndex+1) + "/" + itoa(len(session.Plan.ReviewOrder))
	meta := progress + " · " + countLabel(len(step.Suggestions), "suggestion") + " · " + compactPath(step.FilePath, maxInt(1, width-lipgloss.Width(progress)-18))
	lines := []string{
		sectionTitle(title),
		mutedText(meta),
	}
	if height > 2 && step.Why != "" {
		lines = append(lines, navSection("WHY")+" "+truncateCells(step.Why, maxInt(1, width-5)))
	}
	return forceBlock(strings.Join(lines, "\n"), width, height)
}

func renderReviewDrawer(session review.ReviewSession, step review.ReviewStep, selectedSuggestion int, width int, height int) string {
	comment, ok := selectedReviewComment(step.Suggestions, selectedSuggestion)
	if !ok {
		return forceBlock("", width, height)
	}
	bodyWidth := maxInt(1, width-4)
	meta := lipgloss.JoinHorizontal(lipgloss.Top,
		statusPill(strings.ToUpper(string(comment.Priority)), lipgloss.Color("178")),
		statusPill(string(comment.Category), lipgloss.Color("63")),
		statusPill(string(comment.Status), lipgloss.Color("70")),
	)
	lines := []string{
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Render("AI suggestion"),
			strings.Repeat(" ", 2),
			meta,
			mutedText(compactPath(targetText(comment), maxInt(1, bodyWidth-lipgloss.Width(meta)-18))),
		),
	}
	body := wordWrap(comment.Body, bodyWidth)
	if body != "" {
		lines = append(lines, splitBlock(body)...)
	}
	lines = append(lines, mutedText("a approve  d dismiss  e edit  r reword  C queue"))
	if approved := len(session.ApprovedComments()); approved > 0 {
		lines = append(lines, mutedText("queue: "+itoa(approved)+" approved"))
	}
	content := strings.Join(lines, "\n")
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)
	return forceBlock(style.Render(content), width, height)
}

func statusPill(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Padding(0, 1).
		MarginRight(1).
		Render(text)
}

func itoa(value int) string {
	return strconv.Itoa(value)
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
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

func renderRail(session review.ReviewSession, width int, height int, viewedFiles map[string]bool, focused FocusPane) string {
	var lines []string
	files := orderedStepFiles(session)
	grouped := orderedGroups(session)
	lines = append(lines, mutedText(countLabel(len(session.Plan.ReviewOrder), "step")+" · "+countLabel(len(files), "file")))
	renderedSteps := map[int]bool{}
	for _, group := range grouped {
		lines = append(lines, "", sectionTitle(group.Title))
		for _, stepIndex := range stepIndexesForGroup(session, group.ID) {
			lines = appendStepRailEntry(lines, session, stepIndex, width)
			renderedSteps[stepIndex] = true
		}
	}
	if len(grouped) == 0 {
		for stepIndex := range session.Plan.ReviewOrder {
			lines = appendStepRailEntry(lines, session, stepIndex, width)
		}
	} else {
		var otherSteps []int
		for stepIndex := range session.Plan.ReviewOrder {
			if !renderedSteps[stepIndex] {
				otherSteps = append(otherSteps, stepIndex)
			}
		}
		if len(otherSteps) > 0 {
			lines = append(lines, "", sectionTitle("Other steps"))
			for _, stepIndex := range otherSteps {
				lines = appendStepRailEntry(lines, session, stepIndex, width)
			}
		}
	}
	current, _ := currentStep(session)
	lines = append(lines, "", navSection("FILES"))
	for _, file := range files {
		lines = appendFileRailEntry(lines, current, file, width, viewedFiles)
	}
	lines = append(lines, "")
	lines = append(lines, mutedText("step "+itoa(session.Cursor.StepIndex+1)+"/"+itoa(len(session.Plan.ReviewOrder))+" · "+itoa(viewedFileCount(viewedFiles))+"/"+itoa(len(files))+" viewed"))
	lines = append(lines, mutedText(countLabel(len(session.ApprovedComments()), "approved comment")))
	return pane(FocusRail, focused, "WALKTHROUGH", strings.Join(lines, "\n"), width, height)
}

func appendStepRailEntry(lines []string, session review.ReviewSession, stepIndex int, width int) []string {
	step := session.Plan.ReviewOrder[stepIndex]
	prefix := "  "
	style := lipgloss.NewStyle().Width(width)
	if stepIndex == session.Cursor.StepIndex {
		prefix = "▶ "
		style = style.Background(lipgloss.Color("235")).Foreground(lipgloss.Color("230")).Bold(true)
	}
	title := walkthroughStepTitle(step)
	row := fmt.Sprintf("%s%02d %s", prefix, stepIndex+1, title)
	lines = append(lines, style.Render(truncateCells(row, width)))
	return lines
}

func appendFileRailEntry(lines []string, current review.ReviewStep, file string, width int, viewedFiles map[string]bool) []string {
	viewedPrefix := "  "
	if viewedFiles[file] {
		viewedPrefix = "✓ "
	}
	filePrefix := viewedPrefix
	style := lipgloss.NewStyle().Width(width)
	if file == current.FilePath {
		filePrefix = "▶ "
		style = style.Background(lipgloss.Color("235")).Foreground(lipgloss.Color("230"))
		if viewedFiles[file] {
			filePrefix = "▶ ✓ "
		}
	}
	lines = append(lines, style.Render(truncateCells(filePrefix+compactPath(file, width-2), width)))
	return lines
}

type railGroup struct {
	ID    string
	Title string
}

func orderedGroups(session review.ReviewSession) []railGroup {
	seen := map[string]bool{}
	var groups []railGroup
	for _, step := range session.Plan.ReviewOrder {
		id := step.GroupID
		title := step.GroupTitle
		if id == "" {
			id = title
		}
		if title == "" {
			continue
		}
		if id == "" {
			id = title
		}
		titleKey := "title:" + title
		idKey := "id:" + id
		if seen[idKey] || seen[titleKey] {
			continue
		}
		seen[idKey] = true
		seen[titleKey] = true
		groups = append(groups, railGroup{ID: title, Title: title})
	}
	return groups
}

func filesForGroup(session review.ReviewSession, groupID string) []string {
	seen := map[string]bool{}
	var files []string
	for _, step := range session.Plan.ReviewOrder {
		id := step.GroupID
		if id == "" {
			id = step.GroupTitle
		}
		if id != groupID && step.GroupTitle != groupID {
			continue
		}
		if step.FilePath == "" || seen[step.FilePath] {
			continue
		}
		seen[step.FilePath] = true
		files = append(files, step.FilePath)
	}
	return files
}

func viewedFileCount(viewedFiles map[string]bool) int {
	count := 0
	for _, viewed := range viewedFiles {
		if viewed {
			count++
		}
	}
	return count
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

func stepIndexesForGroup(session review.ReviewSession, groupID string) []int {
	var indexes []int
	for i, step := range session.Plan.ReviewOrder {
		id := step.GroupID
		if id == "" {
			id = step.GroupTitle
		}
		if id == groupID || step.GroupTitle == groupID {
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

func countLabel(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return itoa(count) + " " + singular + "s"
}

func walkthroughStepTitle(step review.ReviewStep) string {
	if step.LayerTitle != "" {
		return step.LayerTitle
	}
	if step.Title != "" {
		return strings.TrimPrefix(step.Title, "Review ")
	}
	return "Review step"
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
	layerLabel := fmt.Sprintf("step %d/%d", session.Cursor.StepIndex+1, len(session.Plan.ReviewOrder))
	if step.GroupTitle != "" {
		layerLabel += " · " + step.GroupTitle
	}
	lines = append(lines, mutedText(layerLabel))
	title := walkthroughStepTitle(step)
	lines = append(lines, sectionTitle(title))
	lines = append(lines, mutedText(compactPath(step.FilePath, width)))
	lines = append(lines, mutedText(itoa(len(step.Suggestions))+" suggested comments"))
	lines = append(lines, navSection("PURPOSE"))
	if step.Summary != "" {
		lines = append(lines, step.Summary)
	} else {
		lines = append(lines, "This walkthrough step focuses on the displayed diff chunk.")
	}
	if step.Why != "" {
		lines = append(lines, navSection("WHY"), step.Why)
	}
	lines = append(lines, navSection("IMPACT"))
	lines = append(lines, impactSummary(step))
	lines = append(lines, navSection("REVIEW FOCUS"))
	if len(step.Focus) > 0 {
		for _, item := range step.Focus {
			lines = append(lines, "• "+item)
		}
	} else {
		lines = append(lines, "• Read the changed lines in order.")
		lines = append(lines, "• Check whether the behavior matches the PR intent.")
	}
	approved := len(session.ApprovedComments())
	lines = append(lines, navSection("CONFIDENCE"))
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
		return "Medium confidence. No suggested comments are attached to this step, but behavior still needs review."
	}
	return fmt.Sprintf("Medium confidence. %d suggested comment(s) need reviewer judgment before this step is considered clear.", len(step.Suggestions))
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

func navSection(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("178")).
		Bold(true).
		Render(text)
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
