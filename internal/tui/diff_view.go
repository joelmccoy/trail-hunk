package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type diffMarker struct {
	Label      string
	Body       string
	Priority   review.Priority
	Category   review.CommentCategory
	Status     review.CommentStatus
	IsSelected bool
}

func renderDiffRows(step review.ReviewStep, selectedSuggestion int, width int) string {
	if width < 1 {
		width = 1
	}

	rows := []string{
		diffHeaderRow(width),
	}
	markers := diffMarkers(step.Suggestions, selectedSuggestion)
	for _, line := range step.DiffLines {
		rows = append(rows, renderWorkbenchDiffLine(line, markers, width))
		marker := markerForLine(line, markers)
		if marker.Body != "" {
			rows = append(rows, renderAnnotationRows(marker, width)...)
		}
	}
	if len(step.DiffLines) == 0 {
		rows = append(rows, mutedLine("no diff lines mapped for this step", width))
	}
	return strings.Join(rows, "\n")
}

func diffMarkers(suggestions []review.ReviewComment, selected int) map[string]diffMarker {
	markers := make(map[string]diffMarker)
	for i, suggestion := range suggestions {
		if suggestion.Line <= 0 {
			continue
		}
		side := suggestion.Side
		if side == "" {
			side = "RIGHT"
		}
		marker := diffMarker{
			Label:      "note",
			Body:       suggestion.Body,
			Priority:   suggestion.Priority,
			Category:   suggestion.Category,
			Status:     suggestion.Status,
			IsSelected: i == selected,
		}
		if marker.IsSelected {
			marker.Label = ">>"
		}
		markers[diffTargetKey(side, suggestion.Line)] = marker
	}
	return markers
}

func diffHeaderRow(width int) string {
	return mutedLine(fmt.Sprintf("%-6s %5s %5s  %s", "mark", "old", "new", "code"), width)
}

func renderWorkbenchDiffLine(line review.DiffLine, markers map[string]diffMarker, width int) string {
	marker := markerForLine(line, markers)
	prefix := " "
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	switch line.Kind {
	case review.DiffLineAdded:
		prefix = "+"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case review.DiffLineDeleted:
		prefix = "-"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	}

	label := marker.Label
	if label == "" {
		label = " "
	}
	row := fmt.Sprintf("%-6s %5s %5s  %s%s", label, lineNumber(line.OldLine), lineNumber(line.NewLine), prefix, line.Text)
	if marker.IsSelected {
		style = style.Bold(true).Background(lipgloss.Color("236"))
	}
	return style.Render(truncateCells(row, width))
}

func renderAnnotationRows(marker diffMarker, width int) []string {
	if width < 1 {
		width = 1
	}
	label := "suggested comment"
	if marker.Status != "" {
		label = fmt.Sprintf("%s comment", marker.Status)
	}
	meta := strings.TrimSpace(strings.Join(nonEmptyStrings(
		strings.ToUpper(string(marker.Priority)),
		string(marker.Category),
		label,
	), " · "))
	header := fmt.Sprintf("       %s", meta)
	bodyWidth := maxInt(1, width-9)
	body := wordWrap(marker.Body, bodyWidth)
	actions := "a approve  d dismiss  e edit  r reword"
	if marker.Status == review.StatusApproved {
		actions = "approved  d dismiss  e edit"
	}
	if marker.Status == review.StatusDismissed {
		actions = "dismissed  a approve"
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	if marker.IsSelected {
		style = style.Background(lipgloss.Color("236"))
	}

	rows := []string{
		style.Render(truncateCells(header, width)),
	}
	for _, line := range splitBlock(body) {
		rows = append(rows, style.Render(truncateCells("       "+line, width)))
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(truncateCells("       "+actions, width)))
	return rows
}

func markerForLine(line review.DiffLine, markers map[string]diffMarker) diffMarker {
	switch line.Kind {
	case review.DiffLineDeleted:
		if line.OldLine == nil {
			return diffMarker{}
		}
		return markers[diffTargetKey("LEFT", *line.OldLine)]
	default:
		if line.NewLine == nil {
			return diffMarker{}
		}
		return markers[diffTargetKey("RIGHT", *line.NewLine)]
	}
}

func lineNumber(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

func mutedLine(text string, width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(truncateCells(text, width))
}

func nonEmptyStrings(values ...string) []string {
	var out []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}
