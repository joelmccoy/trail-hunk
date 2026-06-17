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
	if marker.Body != "" {
		row += renderInlineMarker(marker)
	}
	if marker.IsSelected {
		style = style.Bold(true).Background(lipgloss.Color("236"))
	}
	return style.Render(truncateCells(row, width))
}

func renderInlineMarker(marker diffMarker) string {
	var parts []string
	if marker.Priority != "" {
		parts = append(parts, string(marker.Priority))
	}
	if marker.Category != "" {
		parts = append(parts, string(marker.Category))
	}
	if marker.Status != "" {
		parts = append(parts, string(marker.Status))
	}
	parts = append(parts, marker.Body)
	return "  | " + strings.Join(parts, " ")
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
