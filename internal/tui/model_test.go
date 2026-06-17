package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

func TestKeyQQuits(t *testing.T) {
	model := NewModel(review.ReviewSession{})
	_, cmd := model.Update(key("q"))
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("command returned %T, want tea.QuitMsg", msg)
	}
}

func TestStepNavigationKeys(t *testing.T) {
	model := NewModel(review.ReviewSession{
		Plan: review.WalkthroughPlan{
			ReviewOrder: []review.ReviewStep{
				{ID: "step-1"},
				{ID: "step-2"},
			},
		},
	})

	updated, _ := model.Update(key("n"))
	model = updated.(Model)
	if model.Session.Cursor.StepIndex != 1 {
		t.Fatalf("StepIndex = %d, want 1", model.Session.Cursor.StepIndex)
	}

	updated, _ = model.Update(key("p"))
	model = updated.(Model)
	if model.Session.Cursor.StepIndex != 0 {
		t.Fatalf("StepIndex = %d, want 0", model.Session.Cursor.StepIndex)
	}
}

func TestToggleKeys(t *testing.T) {
	model := NewModel(review.ReviewSession{})

	updated, _ := model.Update(key("f"))
	model = updated.(Model)
	if model.ShowFileTree {
		t.Fatal("expected file tree hidden after first toggle")
	}

	updated, _ = model.Update(key("f"))
	model = updated.(Model)
	if !model.ShowFileTree {
		t.Fatal("expected file tree visible after second toggle")
	}

	updated, _ = model.Update(key("t"))
	model = updated.(Model)
	if !model.ShowAskPane {
		t.Fatal("expected ask pane visible")
	}
}

func TestSuggestionSelectionAndActions(t *testing.T) {
	model := NewModel(review.ReviewSession{
		Plan: review.WalkthroughPlan{
			ReviewOrder: []review.ReviewStep{
				{
					ID:      "step-1",
					Title:   "Review startup",
					Summary: "Startup initializes a review.",
					Why:     "The user needs a first step.",
					Suggestions: []review.ReviewComment{
						{ID: "c1", Body: "First suggestion", Status: review.StatusSuggested},
						{ID: "c2", Body: "Second suggestion", Status: review.StatusSuggested},
					},
				},
			},
		},
		Comments: []review.ReviewComment{
			{ID: "c1", Body: "First suggestion", Status: review.StatusSuggested},
			{ID: "c2", Body: "Second suggestion", Status: review.StatusSuggested},
		},
	})
	model.Screen = ScreenWalkthrough

	updated, _ := model.Update(key("J"))
	model = updated.(Model)
	if model.SelectedSuggestion != 1 {
		t.Fatalf("SelectedSuggestion = %d, want 1", model.SelectedSuggestion)
	}

	updated, _ = model.Update(key("a"))
	model = updated.(Model)
	if model.Session.Comments[1].Status != review.StatusApproved {
		t.Fatalf("Status = %q, want approved", model.Session.Comments[1].Status)
	}
	if model.Session.Plan.ReviewOrder[0].Suggestions[1].Status != review.StatusApproved {
		t.Fatalf("step suggestion status = %q, want approved", model.Session.Plan.ReviewOrder[0].Suggestions[1].Status)
	}

	updated, _ = model.Update(key("K"))
	model = updated.(Model)
	if model.SelectedSuggestion != 0 {
		t.Fatalf("SelectedSuggestion = %d, want 0", model.SelectedSuggestion)
	}

	updated, _ = model.Update(key("d"))
	model = updated.(Model)
	if model.Session.Comments[0].Status != review.StatusDismissed {
		t.Fatalf("Status = %q, want dismissed", model.Session.Comments[0].Status)
	}
}

func TestCommentQueueAndSubmit(t *testing.T) {
	submitted := false
	model := NewModelWithOptions(review.ReviewSession{
		Comments: []review.ReviewComment{
			{ID: "c1", Body: "Approved suggestion", Status: review.StatusApproved},
			{ID: "c2", Body: "Dismissed suggestion", Status: review.StatusDismissed},
		},
	}, Options{
		ReviewSubmitter: func(context.Context, review.ReviewSession) error {
			submitted = true
			return nil
		},
	})

	updated, _ := model.Update(key("C"))
	model = updated.(Model)
	if model.Screen != ScreenComments {
		t.Fatalf("Screen = %q, want comments", model.Screen)
	}
	if !strings.Contains(model.View(), "Approved suggestion") {
		t.Fatalf("View() = %q, want approved comment", model.View())
	}
	if strings.Contains(model.View(), "Dismissed suggestion") {
		t.Fatalf("View() = %q, did not want dismissed comment", model.View())
	}

	updated, cmd := model.Update(key("S"))
	model = updated.(Model)
	if !model.Submitting {
		t.Fatal("expected submitting state")
	}
	if cmd == nil {
		t.Fatal("expected submit command")
	}

	updated, _ = model.Update(cmd())
	model = updated.(Model)
	if !submitted {
		t.Fatal("submitter was not called")
	}
	if model.Submitting {
		t.Fatal("did not expect submitting after success")
	}
	if model.Session.Comments[0].Status != review.StatusSubmitted {
		t.Fatalf("Status = %q, want submitted", model.Session.Comments[0].Status)
	}
}

func TestStartReviewKeyRunsStarterAndLoadsWalkthrough(t *testing.T) {
	expected := review.ReviewSession{
		Plan: review.WalkthroughPlan{
			Overview: "Adds a guided review flow.",
			ReviewOrder: []review.ReviewStep{
				{
					ID:      "step-1",
					Title:   "Review startup",
					Summary: "Startup now initializes a review.",
					Why:     "The user needs a concrete first review step.",
				},
			},
		},
	}
	model := NewModelWithStarter(review.ReviewSession{}, func(context.Context) (review.ReviewSession, error) {
		return expected, nil
	})

	updated, cmd := model.Update(key("R"))
	model = updated.(Model)
	if !model.Loading {
		t.Fatal("expected loading after pressing R")
	}
	if model.Screen != ScreenStartup {
		t.Fatalf("Screen = %q, want startup while loading", model.Screen)
	}
	if cmd == nil {
		t.Fatal("expected start review command")
	}

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)
	if model.Loading {
		t.Fatal("did not expect loading after success")
	}
	if model.Screen != ScreenWalkthrough {
		t.Fatalf("Screen = %q, want walkthrough", model.Screen)
	}
	if model.Session.Plan.Overview != expected.Plan.Overview {
		t.Fatalf("Overview = %q", model.Session.Plan.Overview)
	}
}

func TestWalkthroughRendersDiffAndSuggestions(t *testing.T) {
	oldLine := 2
	newLine := 2
	anotherNewLine := 3
	model := NewModel(review.ReviewSession{
		Plan: review.WalkthroughPlan{
			ReviewOrder: []review.ReviewStep{
				{
					ID:       "step-1",
					FilePath: "dev/fixtures/dummy-pr/review_target.go",
					Title:    "Review rename",
					Summary:  "A function was renamed.",
					Why:      "Callers may need updates.",
					DiffLines: []review.DiffLine{
						{Kind: review.DiffLineContext, OldLine: intPtr(1), NewLine: intPtr(1), Text: "package main"},
						{Kind: review.DiffLineDeleted, OldLine: &oldLine, Text: "func oldName() {}"},
						{Kind: review.DiffLineAdded, NewLine: &newLine, Text: "func newName() {}"},
						{Kind: review.DiffLineAdded, NewLine: &anotherNewLine, Text: "func helper() {}"},
					},
					Suggestions: []review.ReviewComment{
						{ID: "c1", FilePath: "app.go", Side: "RIGHT", Line: 2, Body: "Confirm callers were updated.", Priority: review.PriorityMedium, Status: review.StatusSuggested},
						{ID: "c2", FilePath: "app.go", Side: "RIGHT", Line: 3, Body: "Check helper visibility.", Priority: review.PriorityLow, Status: review.StatusSuggested},
					},
				},
			},
		},
		Comments: []review.ReviewComment{
			{ID: "c1", Body: "Confirm callers were updated.", Priority: review.PriorityMedium, Status: review.StatusSuggested},
			{ID: "c2", Body: "Check helper visibility.", Priority: review.PriorityLow, Status: review.StatusSuggested},
		},
	})
	model.Screen = ScreenWalkthrough
	model.Width = 120
	model.Height = 32

	view := model.View()
	for _, want := range []string{"Review rename", "Diff", "old", "new", "note", ">>", "func newName() {}", "func helper() {}", "suggested comment", "Confirm callers"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
	for i, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width != model.Width {
			t.Fatalf("line %d width = %d, want %d: %q", i, width, model.Width, line)
		}
	}
}

func TestRenderDiffRowsShowsTargetAndStatusMarkers(t *testing.T) {
	model := walkthroughModelWithDiff()
	step := model.Session.Plan.ReviewOrder[0]

	rendered := renderDiffRows(step, 1, 100)

	for _, want := range []string{"old", "new", ">>", "note", "func helper() {}", "Check helper visibility."} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered diff missing %q:\n%s", want, rendered)
		}
	}
}

func TestWalkthroughUsesWorkbenchNotCardGrid(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 140
	model.Height = 34

	view := model.View()

	if strings.Contains(view, "╭") || strings.Contains(view, "╰") {
		t.Fatalf("workbench should not render decorative card borders:\n%s", view)
	}
	for i, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width != model.Width {
			t.Fatalf("line %d width = %d, want %d: %q", i, width, model.Width, line)
		}
	}
}

func TestSuggestionNavigationHighlightsTargetLine(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 120
	model.Height = 32

	updated, _ := model.Update(key("J"))
	model = updated.(Model)

	view := model.View()
	if model.SelectedSuggestion != 1 {
		t.Fatalf("SelectedSuggestion = %d, want 1", model.SelectedSuggestion)
	}
	if !strings.Contains(view, ">>") {
		t.Fatalf("selected suggestion target was not highlighted:\n%s", view)
	}
	if !strings.Contains(view, "Check helper visibility.") {
		t.Fatalf("inspector did not show selected suggestion:\n%s", view)
	}
}

func TestApproveUpdatesWorkbenchSuggestionState(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 120
	model.Height = 32

	updated, _ := model.Update(key("a"))
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "approved") {
		t.Fatalf("approved status missing from workbench:\n%s", view)
	}
}

func TestCommentQueueUsesPlainWorkbenchStyle(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 100
	model.Height = 24
	updated, _ := model.Update(key("a"))
	model = updated.(Model)
	updated, _ = model.Update(key("C"))
	model = updated.(Model)

	view := model.View()
	if strings.Contains(view, "╭") || strings.Contains(view, "╰") {
		t.Fatalf("comment queue should not render decorative card borders:\n%s", view)
	}
	if !strings.Contains(view, "approved") {
		t.Fatalf("comment queue missing approved comment:\n%s", view)
	}
}

func TestAskDrawerAcceptsTypedInput(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 120
	model.Height = 32

	updated, _ := model.Update(key("t"))
	model = updated.(Model)
	updated, _ = model.Update(key("w"))
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "ask current context") {
		t.Fatalf("ask drawer missing:\n%s", view)
	}
	if got := model.Workbench.Ask.Value(); got != "w" {
		t.Fatalf("ask drawer input = %q, want %q", got, "w")
	}
}

func TestMoveKeysScrollFocusedDiffViewport(t *testing.T) {
	model := walkthroughModelWithLongDiff()
	model.Width = 100
	model.Height = 12
	model.Workbench.SetSize(96, 8)
	model.Workbench.Sync(model.Session, model.SelectedSuggestion, false)

	updated, _ := model.Update(key("j"))
	model = updated.(Model)

	if model.Workbench.Diff.YOffset == 0 {
		t.Fatalf("expected j to scroll diff viewport")
	}
}

func TestWorkbenchShowsChangedFileRailByDefault(t *testing.T) {
	model := walkthroughModelWithMultipleFiles()
	model.Width = 160
	model.Height = 34

	view := model.View()

	for _, want := range []string{"Changed Files", "review_target.go", "orchestration.go", "Access guard", "Map diff lines"} {
		if !strings.Contains(view, want) {
			t.Fatalf("workbench missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, "▶ review_target.go") {
		t.Fatalf("current file is not highlighted:\n%s", view)
	}
	if !strings.Contains(view, "▶ Access guard") {
		t.Fatalf("current step is not highlighted:\n%s", view)
	}
}

func TestWorkbenchRendersFindingsAsInlineAnnotations(t *testing.T) {
	model := walkthroughModelWithDiff()
	step := model.Session.Plan.ReviewOrder[0]

	rendered := renderDiffRows(step, 0, 120)
	codeLine := lineContaining(rendered, "func newName()")
	if strings.Contains(codeLine, "Confirm callers were updated.") {
		t.Fatalf("code row contains finding prose: %q\n%s", codeLine, rendered)
	}
	if !strings.Contains(rendered, "suggested comment") {
		t.Fatalf("inline annotation label missing:\n%s", rendered)
	}
	if !strings.Contains(rendered, "a approve") || !strings.Contains(rendered, "d dismiss") {
		t.Fatalf("annotation actions missing:\n%s", rendered)
	}
}

func TestWorkbenchRightPaneExplainsCurrentChunk(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34

	view := model.View()

	for _, want := range []string{"What this chunk does", "Why it matters", "Used by / impact", "How to review", "Confidence"} {
		if !strings.Contains(view, want) {
			t.Fatalf("right pane missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "selected comment") {
		t.Fatalf("right pane should not be a finding detail panel:\n%s", view)
	}
}

func TestFileNavigationJumpsToNextChangedFile(t *testing.T) {
	model := walkthroughModelWithMultipleFiles()

	updated, _ := model.Update(key("]"))
	model = updated.(Model)

	if got := model.Session.Plan.ReviewOrder[model.Session.Cursor.StepIndex].FilePath; got != "internal/app/orchestration.go" {
		t.Fatalf("current file = %q, want internal/app/orchestration.go", got)
	}

	updated, _ = model.Update(key("["))
	model = updated.(Model)
	if got := model.Session.Plan.ReviewOrder[model.Session.Cursor.StepIndex].FilePath; got != "dev/fixtures/dummy-pr/review_target.go" {
		t.Fatalf("current file = %q, want dev/fixtures/dummy-pr/review_target.go", got)
	}
}

func TestDiffPanelCentersSuggestedLines(t *testing.T) {
	var lines []review.DiffLine
	for line := 1; line <= 20; line++ {
		lines = append(lines, review.DiffLine{
			Kind:    review.DiffLineAdded,
			NewLine: intPtr(line),
			Text:    "line body",
		})
	}
	panel := diffPanel(review.ReviewStep{
		FilePath:  "app.go",
		DiffLines: lines,
		Suggestions: []review.ReviewComment{
			{ID: "c1", Side: "RIGHT", Line: 14, Status: review.StatusSuggested},
		},
	}, 0, 80)

	if !strings.Contains(panel, "focus") {
		t.Fatalf("diff panel did not mark selected target:\n%s", panel)
	}
	if !strings.Contains(panel, "14") {
		t.Fatalf("diff panel did not include target line 14:\n%s", panel)
	}
	if strings.Contains(panel, " 1  +line body") {
		t.Fatalf("diff panel did not elide unrelated top context:\n%s", panel)
	}
	if !strings.Contains(panel, "...") {
		t.Fatalf("diff panel did not show omitted context marker:\n%s", panel)
	}
}

func TestStartupLoaderDisplaysCurrentPullRequest(t *testing.T) {
	model := NewModelWithOptions(review.ReviewSession{}, Options{
		StartupLoader: func(context.Context) (review.StartupContext, error) {
			return review.StartupContext{
				Repo: review.RepoRef{Owner: "joelmccoy", Name: "trail-hunk", Branch: "feature"},
				PR:   &review.PullRequest{Number: 12, Title: "Add guided review", State: "open"},
			}, nil
		},
	})

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("expected startup loader command")
	}

	updated, _ := model.Update(cmd())
	model = updated.(Model)
	view := model.View()
	if !strings.Contains(view, "joelmccoy/trail-hunk") {
		t.Fatalf("View() = %q, want repo", view)
	}
	if !strings.Contains(view, "#12 Add guided review") {
		t.Fatalf("View() = %q, want PR title", view)
	}
}

func TestStartupLoaderFailureDisplaysError(t *testing.T) {
	model := NewModelWithOptions(review.ReviewSession{}, Options{
		StartupLoader: func(context.Context) (review.StartupContext, error) {
			return review.StartupContext{}, errors.New("gh auth token failed")
		},
	})

	cmd := model.Init()
	updated, _ := model.Update(cmd())
	model = updated.(Model)
	if model.StartupErr == nil {
		t.Fatal("expected startup error")
	}
	if !strings.Contains(model.View(), "gh auth token failed") {
		t.Fatalf("View() = %q, want startup error", model.View())
	}
}

func TestViewLinesFitTerminalWidth(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{
		{width: 60, height: 16},
		{width: 80, height: 20},
		{width: 120, height: 24},
		{width: 200, height: 36},
	}

	for _, size := range sizes {
		model := NewModel(review.ReviewSession{})
		model.Width = size.width
		model.Height = size.height
		model.Startup = review.StartupContext{
			Repo:    review.RepoRef{Owner: "joelmccoy", Name: "trail-hunk", Branch: "main"},
			Message: `No open GitHub pull request found for branch "main".`,
		}

		lines := strings.Split(model.View(), "\n")
		if len(lines) > model.Height {
			t.Fatalf("%dx%d rendered %d lines, want <= %d:\n%s", size.width, size.height, len(lines), model.Height, model.View())
		}
		for i, line := range lines {
			if width := lipgloss.Width(line); width != model.Width {
				t.Fatalf("%dx%d line %d width = %d, want %d: %q", size.width, size.height, i, width, model.Width, line)
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "─╮" || trimmed == "─╯" {
				t.Fatalf("%dx%d line %d has orphaned border cap: %q", size.width, size.height, i, line)
			}
		}
	}
}

func TestStartupUsesPlainFullWidthLayout(t *testing.T) {
	model := NewModel(review.ReviewSession{})
	model.Width = 160
	model.Height = 32
	model.Startup = review.StartupContext{
		Repo:    review.RepoRef{Owner: "joelmccoy", Name: "trail-hunk", Branch: "main"},
		Message: `No open GitHub pull request found for branch "main".`,
	}

	view := model.View()
	if strings.Contains(view, "╭") || strings.Contains(view, "╰") {
		t.Fatalf("startup should not render decorative card borders:\n%s", view)
	}
	if !strings.Contains(view, "Repository") || !strings.Contains(view, "No PR Found") {
		t.Fatalf("startup missing expected sections:\n%s", view)
	}
}

func TestFooterMenuDoesNotWrapAtNormalWidth(t *testing.T) {
	widths := []int{40, 60, 80, 120}
	for _, width := range widths {
		model := NewModel(review.ReviewSession{})
		model.Width = width
		model.Height = 20

		lines := strings.Split(model.View(), "\n")
		footer := lines[len(lines)-1]
		if lipgloss.Height(footer) != 1 {
			t.Fatalf("footer height = %d, want 1: %q", lipgloss.Height(footer), footer)
		}
		if got := lipgloss.Width(footer); got != width {
			t.Fatalf("footer width = %d, want %d: %q", got, width, footer)
		}
		if strings.TrimSpace(footer) == "" {
			t.Fatalf("footer is empty at width %d", width)
		}
	}
}

func TestFooterUsesGeneratedKeyHelp(t *testing.T) {
	model := NewModel(review.ReviewSession{})
	model.Width = 120
	model.Height = 24

	view := model.View()

	if !strings.Contains(view, "R review") {
		t.Fatalf("footer missing review key help:\n%s", view)
	}
	if !strings.Contains(view, "tab focus") {
		t.Fatalf("footer missing focus key help:\n%s", view)
	}
}

func TestCommentsScreenLinesFitTerminalWidth(t *testing.T) {
	model := NewModel(review.ReviewSession{})
	model.Width = 80
	model.Height = 20
	model.Screen = ScreenComments

	for i, line := range strings.Split(model.View(), "\n") {
		if width := lipgloss.Width(line); width > model.Width {
			t.Fatalf("line %d width = %d, want <= %d: %q", i, width, model.Width, line)
		}
	}
}

func TestStartReviewFailureShowsError(t *testing.T) {
	model := NewModelWithStarter(review.ReviewSession{}, func(context.Context) (review.ReviewSession, error) {
		return review.ReviewSession{}, errors.New("no open GitHub pull request found")
	})

	updated, cmd := model.Update(key("R"))
	model = updated.(Model)
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if model.Loading {
		t.Fatal("did not expect loading after failure")
	}
	if model.Err == nil {
		t.Fatal("expected error")
	}
	if model.Screen != ScreenStartup {
		t.Fatalf("Screen = %q, want startup", model.Screen)
	}
	if got := model.View(); !contains(got, "no open GitHub pull request found") {
		t.Fatalf("View() = %q, want error text", got)
	}
}

func key(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func intPtr(value int) *int {
	return &value
}

func walkthroughModelWithDiff() Model {
	oldLine := 2
	newLine := 2
	anotherNewLine := 3
	model := NewModel(review.ReviewSession{
		Plan: review.WalkthroughPlan{
			Overview: "Adds a guided review flow.",
			ReviewOrder: []review.ReviewStep{
				{
					ID:       "step-1",
					FilePath: "dev/fixtures/dummy-pr/review_target.go",
					Title:    "Access guard",
					Summary:  "A function was renamed.",
					Why:      "Callers may need updates.",
					Focus:    []string{"Confirm callers are updated."},
					DiffLines: []review.DiffLine{
						{Kind: review.DiffLineContext, OldLine: intPtr(1), NewLine: intPtr(1), Text: "package main"},
						{Kind: review.DiffLineDeleted, OldLine: &oldLine, Text: "func oldName() {}"},
						{Kind: review.DiffLineAdded, NewLine: &newLine, Text: "func newName() {}"},
						{Kind: review.DiffLineAdded, NewLine: &anotherNewLine, Text: "func helper() {}"},
					},
					Suggestions: []review.ReviewComment{
						{ID: "c1", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: 2, Body: "Confirm callers were updated.", Priority: review.PriorityMedium, Category: review.CategoryBug, Status: review.StatusSuggested},
						{ID: "c2", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: 3, Body: "Check helper visibility.", Priority: review.PriorityLow, Category: review.CategoryMaintainability, Status: review.StatusSuggested},
					},
				},
			},
		},
		Comments: []review.ReviewComment{
			{ID: "c1", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: 2, Body: "Confirm callers were updated.", Priority: review.PriorityMedium, Category: review.CategoryBug, Status: review.StatusSuggested},
			{ID: "c2", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: 3, Body: "Check helper visibility.", Priority: review.PriorityLow, Category: review.CategoryMaintainability, Status: review.StatusSuggested},
		},
	})
	model.Screen = ScreenWalkthrough
	return model
}

func walkthroughModelWithMultipleFiles() Model {
	model := walkthroughModelWithDiff()
	line := 9
	model.Session.Plan.ReviewOrder = append(model.Session.Plan.ReviewOrder,
		review.ReviewStep{
			ID:       "step-2",
			FilePath: "dev/fixtures/dummy-pr/review_target.go",
			Title:    "Admin bypass",
			Summary:  "Admin accounts bypass ownership checks.",
			Why:      "The policy should be explicit.",
			DiffLines: []review.DiffLine{
				{Kind: review.DiffLineAdded, NewLine: intPtr(18), Text: "return true"},
			},
		},
		review.ReviewStep{
			ID:       "step-3",
			FilePath: "internal/app/orchestration.go",
			Title:    "Map diff lines",
			Summary:  "Review steps receive mapped diff lines.",
			Why:      "The TUI needs display-ready hunk data.",
			DiffLines: []review.DiffLine{
				{Kind: review.DiffLineAdded, NewLine: &line, Text: "step.DiffLines = mapped"},
			},
		},
	)
	return model
}

func walkthroughModelWithLongDiff() Model {
	model := walkthroughModelWithDiff()
	var lines []review.DiffLine
	for line := 1; line <= 40; line++ {
		lines = append(lines, review.DiffLine{
			Kind:    review.DiffLineAdded,
			NewLine: intPtr(line),
			Text:    "line body",
		})
	}
	model.Session.Plan.ReviewOrder[0].DiffLines = lines
	model.Session.Plan.ReviewOrder[0].Suggestions = nil
	model.Session.Comments = nil
	return model
}

func lineContaining(text string, needle string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}

func contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}
