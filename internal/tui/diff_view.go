package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type diffMarker struct {
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

	var rows []string
	markers := diffMarkers(step.Suggestions, selectedSuggestion)
	targets := suggestionTargets(step.Suggestions, selectedSuggestion)
	lines := focusedDiffLines(step.DiffLines, targets, 3)
	for _, line := range lines {
		if line.Text == omittedDiffText {
			rows = append(rows, omittedDiffRow(width))
			continue
		}
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
			Body:       suggestion.Body,
			Priority:   suggestion.Priority,
			Category:   suggestion.Category,
			Status:     suggestion.Status,
			IsSelected: i == selected,
		}
		markers[diffTargetKey(side, suggestion.Line)] = marker
	}
	return markers
}

func omittedDiffRow(width int) string {
	return mutedLine("     ⋯ unchanged context", width)
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

	gutterWidth := 8
	codeWidth := maxInt(1, width-gutterWidth)
	row := lipgloss.JoinHorizontal(lipgloss.Top,
		diffGutter(line, marker),
		style.MaxWidth(codeWidth).Render(truncateCells(prefix+" "+line.Text, codeWidth)),
	)
	if marker.IsSelected {
		row = lipgloss.NewStyle().Width(width).Background(lipgloss.Color("235")).Render(truncateCells(row, width))
	}
	return truncateCells(row, width)
}

func diffGutter(line review.DiffLine, marker diffMarker) string {
	markerText := "  "
	if marker.Body != "" {
		markerText = "◆"
	}
	markerStyle := lipgloss.NewStyle().Width(2).Foreground(lipgloss.Color("244"))
	if marker.Body != "" {
		markerStyle = markerStyle.Foreground(lipgloss.Color("78")).Bold(marker.IsSelected)
	}
	lineStyle := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(lipgloss.Color("244"))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		markerStyle.Render(markerText),
		lineStyle.Render(displayLineNumber(line)),
		lipgloss.NewStyle().Width(1).Render(" "),
	)
}

func compactFindingBody(body string, width int) string {
	body = strings.Join(strings.Fields(body), " ")
	if body == "" {
		return "Review this line before approving the chunk."
	}
	return truncateCells(body, width)
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
	return strconv.Itoa(*value)
}

func displayLineNumber(line review.DiffLine) string {
	if line.Kind == review.DiffLineDeleted && line.OldLine != nil {
		return lineNumber(line.OldLine)
	}
	if line.NewLine != nil {
		return lineNumber(line.NewLine)
	}
	return lineNumber(line.OldLine)
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
