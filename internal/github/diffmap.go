package github

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	SideLeft  = "LEFT"
	SideRight = "RIGHT"
)

type DiffLineKind string

const (
	DiffLineContext DiffLineKind = "context"
	DiffLineAdded   DiffLineKind = "added"
	DiffLineDeleted DiffLineKind = "deleted"
)

type PullRequestDiff struct {
	Files []DiffFile
}

type DiffFile struct {
	Path  string
	Hunks []DiffHunk
}

type DiffHunk struct {
	ID         string
	Header     string
	OldStart   int
	NewStart   int
	Lines      []DiffLine
	DiffLineNo int
}

type DiffLine struct {
	FilePath   string
	HunkID     string
	Kind       DiffLineKind
	OldLine    *int
	NewLine    *int
	DiffLineNo int
	Side       string
	CanComment bool
	Text       string
}

type ReviewTarget struct {
	Path   string
	Side   string
	Line   int
	HunkID string
}

var hunkHeaderPattern = regexp.MustCompile(`^@@ -([0-9]+)(?:,[0-9]+)? \+([0-9]+)(?:,[0-9]+)? @@`)

func ParsePullRequestDiff(raw string) (PullRequestDiff, error) {
	var diff PullRequestDiff
	var currentFile *DiffFile
	var currentHunk *DiffHunk
	var oldLine int
	var newLine int
	var diffLineNo int

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		diffLineNo++

		switch {
		case strings.HasPrefix(line, "diff --git "):
			diff.Files = append(diff.Files, DiffFile{})
			currentFile = &diff.Files[len(diff.Files)-1]
			currentHunk = nil
		case currentFile != nil && strings.HasPrefix(line, "+++ "):
			currentFile.Path = normalizeDiffPath(strings.TrimPrefix(line, "+++ "))
		case currentFile != nil && strings.HasPrefix(line, "@@ "):
			oldStart, newStart, err := parseHunkHeader(line)
			if err != nil {
				return PullRequestDiff{}, err
			}
			oldLine = oldStart
			newLine = newStart
			hunkID := fmt.Sprintf("%s:%d", currentFile.Path, len(currentFile.Hunks)+1)
			currentFile.Hunks = append(currentFile.Hunks, DiffHunk{
				ID:         hunkID,
				Header:     line,
				OldStart:   oldStart,
				NewStart:   newStart,
				DiffLineNo: diffLineNo,
			})
			currentHunk = &currentFile.Hunks[len(currentFile.Hunks)-1]
		case currentFile != nil && currentHunk != nil:
			parsed, consumeOld, consumeNew, ok := parseDiffLine(currentFile.Path, currentHunk.ID, line, oldLine, newLine, diffLineNo)
			if !ok {
				continue
			}
			currentHunk.Lines = append(currentHunk.Lines, parsed)
			if consumeOld {
				oldLine++
			}
			if consumeNew {
				newLine++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return PullRequestDiff{}, fmt.Errorf("scan diff: %w", err)
	}

	return diff, nil
}

func (d PullRequestDiff) FindTarget(path string, side string, line int) (ReviewTarget, error) {
	for _, file := range d.Files {
		if file.Path != path {
			continue
		}
		for _, hunk := range file.Hunks {
			for _, diffLine := range hunk.Lines {
				if !diffLine.CanComment || diffLine.Side != side {
					continue
				}
				if diffLine.reviewLine() == line {
					return ReviewTarget{
						Path:   path,
						Side:   side,
						Line:   line,
						HunkID: diffLine.HunkID,
					}, nil
				}
			}
		}
	}

	return ReviewTarget{}, fmt.Errorf("no commentable diff target for %s %s line %d", path, side, line)
}

func parseHunkHeader(header string) (int, int, error) {
	matches := hunkHeaderPattern.FindStringSubmatch(header)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("parse hunk header %q", header)
	}

	oldStart, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse old hunk start: %w", err)
	}
	newStart, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, fmt.Errorf("parse new hunk start: %w", err)
	}
	return oldStart, newStart, nil
}

func parseDiffLine(filePath string, hunkID string, raw string, oldLine int, newLine int, diffLineNo int) (DiffLine, bool, bool, bool) {
	if raw == "" {
		return DiffLine{}, false, false, false
	}

	text := raw[1:]
	switch raw[0] {
	case ' ':
		oldLineCopy := oldLine
		newLineCopy := newLine
		return DiffLine{
			FilePath:   filePath,
			HunkID:     hunkID,
			Kind:       DiffLineContext,
			OldLine:    &oldLineCopy,
			NewLine:    &newLineCopy,
			DiffLineNo: diffLineNo,
			Side:       SideRight,
			CanComment: true,
			Text:       text,
		}, true, true, true
	case '+':
		newLineCopy := newLine
		return DiffLine{
			FilePath:   filePath,
			HunkID:     hunkID,
			Kind:       DiffLineAdded,
			NewLine:    &newLineCopy,
			DiffLineNo: diffLineNo,
			Side:       SideRight,
			CanComment: true,
			Text:       text,
		}, false, true, true
	case '-':
		oldLineCopy := oldLine
		return DiffLine{
			FilePath:   filePath,
			HunkID:     hunkID,
			Kind:       DiffLineDeleted,
			OldLine:    &oldLineCopy,
			DiffLineNo: diffLineNo,
			Side:       SideLeft,
			CanComment: true,
			Text:       text,
		}, true, false, true
	default:
		return DiffLine{}, false, false, false
	}
}

func (l DiffLine) reviewLine() int {
	if l.Side == SideLeft && l.OldLine != nil {
		return *l.OldLine
	}
	if l.NewLine != nil {
		return *l.NewLine
	}
	return 0
}

func normalizeDiffPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "b/")
	return path
}
