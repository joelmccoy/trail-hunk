package tui

import (
	"context"
	"errors"
	"os"
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
	for _, want := range []string{"rename", "DIFF", "◆", "func newName() {}", "func helper() {}", "AI suggestion", "Confirm callers"} {
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

	for _, want := range []string{"◆", "func helper() {}"} {
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
	if !strings.Contains(view, "◆") {
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
	model.Workbench.Sync(model.Session, model.SelectedSuggestion, false, model.ViewedFiles, model.FocusMode)

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

	for _, want := range []string{"WALKTHROUGH", "FILES", "review_target.go", "orchestration.go", "Billing access guard", "Map diff lines"} {
		if !strings.Contains(view, want) {
			t.Fatalf("workbench missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, "▶ review_target.go") {
		t.Fatalf("current file is not highlighted:\n%s", view)
	}
	if !strings.Contains(view, "▶ 01 Billing access guard") {
		t.Fatalf("current step is not highlighted:\n%s", view)
	}
	for _, bad := range []string{"Change Stack", "layer 1", "findings"} {
		if strings.Contains(view, bad) {
			t.Fatalf("walkthrough rail should not be organized around %q:\n%s", bad, view)
		}
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
}

func TestWorkbenchDiffFocusesAroundSuggestionTargets(t *testing.T) {
	model := walkthroughModelWithLongDiff()
	model.Session.Plan.ReviewOrder[0].Suggestions = []review.ReviewComment{
		{
			ID:       "c1",
			FilePath: "dev/fixtures/dummy-pr/review_target.go",
			Side:     "RIGHT",
			Line:     28,
			Body:     "Review the target line.",
			Priority: review.PriorityHigh,
			Category: review.CategoryBug,
			Status:   review.StatusSuggested,
		},
	}
	step := model.Session.Plan.ReviewOrder[0]

	rendered := renderDiffRows(step, 0, 100)

	if !strings.Contains(rendered, "28") {
		t.Fatalf("focused diff did not include target line 28:\n%s", rendered)
	}
	if strings.Contains(rendered, " 1  +line body") {
		t.Fatalf("focused diff should elide unrelated top context:\n%s", rendered)
	}
	if !strings.Contains(rendered, "unchanged context") {
		t.Fatalf("focused diff should show omitted context marker:\n%s", rendered)
	}
}

func TestWorkbenchFindingAnnotationsStayCompact(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Session.Plan.ReviewOrder[0].Suggestions[0].Body = "This branch fails open for a blank requested account ID. Consider returning false here, or validating the request before calling this helper, so malformed input cannot grant billing access."
	step := model.Session.Plan.ReviewOrder[0]

	rendered := renderDiffRows(step, 0, 84)
	annotationLines := 0
	for _, line := range strings.Split(rendered, "\n") {
		if strings.Contains(line, "HIGH") || strings.Contains(line, "approve") || strings.Contains(line, "fails open") || strings.Contains(line, "malformed input") {
			annotationLines++
		}
	}

	if annotationLines > 3 {
		t.Fatalf("selected finding should render as a compact component, got %d annotation lines:\n%s", annotationLines, rendered)
	}
	if strings.Contains(rendered, "malformed input cannot grant billing access") {
		t.Fatalf("diff pane should not render the full finding body inline:\n%s", rendered)
	}
}

func TestSelectedSuggestionRendersInReviewDrawerNotDiffBody(t *testing.T) {
	model := walkthroughModelWithDummyFixtureDiff()
	model.Width = 200
	model.Height = 36

	view := model.View()

	if !strings.Contains(view, "AI suggestion") {
		t.Fatalf("review drawer missing selected suggestion title:\n%s", view)
	}
	if !strings.Contains(view, "returning false here") {
		t.Fatalf("review drawer missing selected suggestion body:\n%s", view)
	}
	diffLine := lineContaining(view, "◆    14")
	if strings.Contains(diffLine, "returning false here") {
		t.Fatalf("diff annotation should stay compact; full body belongs in drawer: %q", diffLine)
	}
}

func TestWorkbenchUsesProductChromeNotRawTextColumns(t *testing.T) {
	model := walkthroughModelWithDummyFixtureDiff()
	model.Width = 200
	model.Height = 36

	view := model.View()

	for _, want := range []string{"WALKTHROUGH", "FILES", "WHY", "DIFF", "AI suggestion"} {
		if !strings.Contains(view, want) {
			t.Fatalf("workbench missing product chrome %q:\n%s", want, view)
		}
	}
	for _, bad := range []string{"review   old   new  change", "Step context"} {
		if strings.Contains(view, bad) {
			t.Fatalf("workbench still exposes raw report label %q:\n%s", bad, view)
		}
	}
}

func TestWorkbenchAssistantInsightExplainsCurrentChunk(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34

	view := model.View()

	for _, want := range []string{"Billing access guard", "WHY", "Callers may need updates."} {
		if !strings.Contains(view, want) {
			t.Fatalf("assistant insight missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "selected comment") {
		t.Fatalf("assistant insight should not be a finding detail panel:\n%s", view)
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

func TestWorkbenchShowsChangeStackMetadata(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34
	model.Session.Plan.ReviewOrder[0].GroupTitle = "Fixture account helpers"
	model.Session.Plan.ReviewOrder[0].LayerTitle = "Billing access guard"
	model.Session.Plan.ReviewOrder[0].LayerIndex = 1

	view := model.View()

	for _, want := range []string{"WALKTHROUGH", "Fixture account helpers", "01 Billing access guard"} {
		if !strings.Contains(view, want) {
			t.Fatalf("workbench missing %q:\n%s", want, view)
		}
	}
}

func TestWorkbenchRailDoesNotDuplicateLayerAndStepRows(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34
	model.Session.Plan.ReviewOrder[0].Title = "Review billing access guard"
	model.Session.Plan.ReviewOrder[0].LayerTitle = "Billing access guard"

	view := model.View()

	if strings.Contains(view, "Review billing access guard") {
		t.Fatalf("rail should not duplicate layer and step titles:\n%s", view)
	}
	if !strings.Contains(view, "▶ 01 Billing access guard") {
		t.Fatalf("rail should keep the current layer row:\n%s", view)
	}
}

func TestFocusModeHidesSidePanes(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34

	updated, _ := model.Update(key("z"))
	model = updated.(Model)

	view := model.View()
	if strings.Contains(view, "WALKTHROUGH") || strings.Contains(view, "ASSISTANT") {
		t.Fatalf("focus mode should hide side panes:\n%s", view)
	}
	if !strings.Contains(view, "DIFF") || !strings.Contains(view, "func newName()") {
		t.Fatalf("focus mode should keep diff visible:\n%s", view)
	}
}

func TestViewedToggleMarksCurrentFileInRail(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34

	updated, _ := model.Update(key("v"))
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "✓ review_target.go") {
		t.Fatalf("viewed file marker missing:\n%s", view)
	}
}

func TestTabCyclesWorkbenchFocusPanes(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34
	model.Screen = ScreenWalkthrough

	updated, _ := model.Update(key("tab"))
	model = updated.(Model)
	if model.Workbench.Focus != FocusRail {
		t.Fatalf("focus after first tab = %q, want %q", model.Workbench.Focus, FocusRail)
	}

	updated, _ = model.Update(key("tab"))
	model = updated.(Model)
	if model.Workbench.Focus != FocusDiff {
		t.Fatalf("focus after second tab = %q, want %q", model.Workbench.Focus, FocusDiff)
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

func TestWalkthroughFooterIsCompactAndNotClipped(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 160
	model.Height = 34

	view := model.View()
	lines := strings.Split(view, "\n")
	footer := lines[len(lines)-1]

	if strings.Contains(footer, "...") {
		t.Fatalf("walkthrough footer should not clip at normal width: %q", footer)
	}
	for _, want := range []string{"n/p step", "]/[ file", "a approve", "z focus", "q quit"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("walkthrough footer missing %q: %q", want, footer)
		}
	}
}

func TestWorkbenchWideRenderAvoidsKnownRoughEdges(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 200
	model.Height = 36
	model.Session.Plan.ReviewOrder[0].Suggestions[0].Body = "This branch fails open for a blank requested account ID. Consider returning false here, or validating the request before calling this helper, so malformed input cannot grant billing access."

	view := model.View()
	logViewSnapshot(t, view)

	for _, bad := range []string{
		"cohorts · layers · files",
		"Change Stack",
		"findings",
		"layer 1",
		"Review billing access guard",
	} {
		if strings.Contains(view, bad) {
			t.Fatalf("wide workbench contains rough edge %q:\n%s", bad, view)
		}
	}
	for _, want := range []string{
		"WALKTHROUGH",
		"Billing access guard",
		"DIFF dev/fixtures/dummy-pr/review_target.go",
		"a approve",
		"n/p step",
		"]/[ file",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("wide workbench missing %q:\n%s", want, view)
		}
	}
}

func TestWorkbenchDummyFixtureRenderFocusesOnActiveChunk(t *testing.T) {
	model := walkthroughModelWithDummyFixtureDiff()
	model.Width = 200
	model.Height = 36

	view := model.View()
	logViewSnapshot(t, view)

	if strings.Contains(view, "+package dummypr") {
		t.Fatalf("dummy fixture render should not start at unrelated file preamble:\n%s", view)
	}
	for _, want := range []string{
		"unchanged context",
		"return true",
		"AI suggestion",
		"HIGH",
		"security",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("dummy fixture render missing %q:\n%s", want, view)
		}
	}
}

func TestWorkbenchComplexFixtureJourneySnapshots(t *testing.T) {
	model := walkthroughModelWithComplexFixture()
	model.Width = 200
	model.Height = 36

	cases := []struct {
		stepIndex int
		title     string
		file      string
		code      string
	}{
		{stepIndex: 0, title: "Billing access guard", file: "review_target.go", code: "return true"},
		{stepIndex: 1, title: "Display-name normalization", file: "review_target.go", code: "return trimmed[:24]"},
		{stepIndex: 2, title: "Permission check callsite", file: "billing_handler.go", code: "CanAccessBilling"},
		{stepIndex: 3, title: "Unicode display-name test", file: "review_target_test.go", code: "NormalizeDisplayName"},
	}

	for _, tc := range cases {
		model.Session.Cursor.StepIndex = tc.stepIndex
		model.SelectedSuggestion = 0
		view := model.View()
		logViewSnapshot(t, view)
		for _, want := range []string{tc.title, tc.file, tc.code, "WALKTHROUGH", "FILES"} {
			if !strings.Contains(view, want) {
				t.Fatalf("step %d render missing %q:\n%s", tc.stepIndex+1, want, view)
			}
		}
		for _, bad := range []string{"old    new", "mark", "cohorts", "findings"} {
			if strings.Contains(view, bad) {
				t.Fatalf("step %d render contains rough edge %q:\n%s", tc.stepIndex+1, bad, view)
			}
		}
		for i, line := range strings.Split(view, "\n") {
			if width := lipgloss.Width(line); width != model.Width {
				t.Fatalf("step %d line %d width = %d, want %d: %q", tc.stepIndex+1, i, width, model.Width, line)
			}
		}
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
					ID:         "step-1",
					FilePath:   "dev/fixtures/dummy-pr/review_target.go",
					Title:      "Access guard",
					GroupID:    "fixture-account",
					GroupTitle: "Fixture account helpers",
					LayerIndex: 1,
					LayerTitle: "Billing access guard",
					Summary:    "A function was renamed.",
					Why:        "Callers may need updates.",
					Focus:      []string{"Confirm callers are updated."},
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

func walkthroughModelWithDummyFixtureDiff() Model {
	model := walkthroughModelWithDiff()
	text := []string{
		"package dummypr",
		"",
		"import \"strings\"",
		"",
		"type Account struct {",
		"ID       string",
		"Role     string",
		"IsActive bool",
		"}",
		"",
		"// CanAccessBilling is intentionally imperfect fixture code for trail-hunk reviews.",
		"func CanAccessBilling(account Account, requestedAccountID string) bool {",
		"if strings.TrimSpace(requestedAccountID) == \"\" {",
		"return true",
		"}",
		"",
		"if account.Role == \"admin\" {",
		"return true",
		"}",
		"",
		"return account.IsActive && account.ID == requestedAccountID",
		"}",
		"",
		"func NormalizeDisplayName(name string) string {",
		"trimmed := strings.TrimSpace(name)",
		"if len(trimmed) > 24 {",
		"return trimmed[:24]",
		"}",
		"return trimmed",
		"}",
	}
	var lines []review.DiffLine
	for i, line := range text {
		lines = append(lines, review.DiffLine{
			Kind:    review.DiffLineAdded,
			NewLine: intPtr(i + 1),
			Text:    line,
		})
	}
	model.Session.Plan.ReviewOrder[0].Title = "Review billing access guard"
	model.Session.Plan.ReviewOrder[0].Summary = "The helper grants access when the requested account ID is blank."
	model.Session.Plan.ReviewOrder[0].Why = "Empty or malformed resource identifiers should fail closed so callers cannot accidentally bypass authorization."
	model.Session.Plan.ReviewOrder[0].Focus = []string{
		"Check whether blank requestedAccountID should ever be valid.",
		"Confirm admin bypass behavior is intentional and audited.",
	}
	model.Session.Plan.ReviewOrder[0].DiffLines = lines
	model.Session.Plan.ReviewOrder[0].Suggestions = []review.ReviewComment{
		{ID: "c1", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: 14, Body: "This branch fails open for a blank requested account ID. Consider returning false here, or validating the request before calling this helper, so malformed input cannot grant billing access.", Priority: review.PriorityHigh, Category: review.CategorySecurity, Status: review.StatusSuggested},
		{ID: "c2", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: 18, Body: "The admin bypass may be intended, but it would be safer to make that policy explicit in the function name, documentation, or a caller-side authorization check.", Priority: review.PriorityMedium, Category: review.CategoryQuestion, Status: review.StatusSuggested},
	}
	return model
}

func walkthroughModelWithComplexFixture() Model {
	model := walkthroughModelWithDummyFixtureDiff()
	line14 := 14
	line18 := 18
	line27 := 27
	line42 := 42
	line12 := 12
	model.Session.Plan.ReviewOrder = []review.ReviewStep{
		model.Session.Plan.ReviewOrder[0],
		{
			ID:         "fixture-display-name",
			FilePath:   "dev/fixtures/dummy-pr/review_target.go",
			Title:      "Review display-name normalization",
			GroupID:    "fixture-account-helpers",
			GroupTitle: "Fixture account helpers",
			LayerIndex: 2,
			LayerTitle: "Display-name normalization",
			Summary:    "The display-name helper trims whitespace and truncates long names.",
			Why:        "User-visible strings often contain multi-byte characters, so byte slicing can produce invalid UTF-8.",
			Focus:      []string{"Check byte/rune handling.", "Look for Unicode tests."},
			DiffLines: []review.DiffLine{
				{Kind: review.DiffLineAdded, NewLine: intPtr(24), Text: "func NormalizeDisplayName(name string) string {"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(25), Text: "trimmed := strings.TrimSpace(name)"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(26), Text: "if len(trimmed) > 24 {"},
				{Kind: review.DiffLineAdded, NewLine: &line27, Text: "return trimmed[:24]"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(28), Text: "}"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(29), Text: "return trimmed"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(30), Text: "}"},
			},
			Suggestions: []review.ReviewComment{
				{ID: "c3", FilePath: "dev/fixtures/dummy-pr/review_target.go", Side: "RIGHT", Line: line27, Body: "This truncates by byte index and can split multi-byte characters.", Priority: review.PriorityMedium, Category: review.CategoryCorrectness, Status: review.StatusSuggested},
			},
		},
		{
			ID:         "fixture-handler-callsite",
			FilePath:   "internal/billing/billing_handler.go",
			Title:      "Review permission check callsite",
			GroupID:    "billing-flow",
			GroupTitle: "Billing request flow",
			LayerIndex: 3,
			LayerTitle: "Permission check callsite",
			Summary:    "The handler delegates billing access to the new helper.",
			Why:        "The helper now sits on the request path and controls access for downstream billing operations.",
			Focus:      []string{"Verify empty account IDs cannot reach the helper.", "Check caller-side auditing."},
			DiffLines: []review.DiffLine{
				{Kind: review.DiffLineContext, OldLine: intPtr(39), NewLine: intPtr(39), Text: "func HandleBillingRequest(account Account, request Request) error {"},
				{Kind: review.DiffLineAdded, NewLine: &line42, Text: "if !CanAccessBilling(account, request.AccountID) {"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(43), Text: "return ErrForbidden"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(44), Text: "}"},
				{Kind: review.DiffLineContext, OldLine: intPtr(45), NewLine: intPtr(45), Text: "return createBillingSession(request)"},
			},
			Suggestions: []review.ReviewComment{
				{ID: "c4", FilePath: "internal/billing/billing_handler.go", Side: "RIGHT", Line: line42, Body: "Validate request.AccountID before calling the helper so malformed IDs fail closed at the boundary.", Priority: review.PriorityHigh, Category: review.CategorySecurity, Status: review.StatusSuggested},
			},
		},
		{
			ID:         "fixture-display-test",
			FilePath:   "internal/billing/review_target_test.go",
			Title:      "Review Unicode display-name test",
			GroupID:    "billing-flow",
			GroupTitle: "Billing request flow",
			LayerIndex: 4,
			LayerTitle: "Unicode display-name test",
			Summary:    "A regression test documents display-name behavior.",
			Why:        "Tests should describe whether truncation is byte-based, rune-based, or display-width-based.",
			Focus:      []string{"Confirm the expected string is valid UTF-8.", "Add a boundary case for exactly 24 display cells."},
			DiffLines: []review.DiffLine{
				{Kind: review.DiffLineAdded, NewLine: intPtr(10), Text: "func TestNormalizeDisplayNameUnicode(t *testing.T) {"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(11), Text: "got := NormalizeDisplayName(\"José 🚀 customer account\")"},
				{Kind: review.DiffLineAdded, NewLine: &line12, Text: "if got == \"\" { t.Fatal(\"expected display name\") }"},
				{Kind: review.DiffLineAdded, NewLine: intPtr(13), Text: "}"},
			},
			Suggestions: []review.ReviewComment{
				{ID: "c5", FilePath: "internal/billing/review_target_test.go", Side: "RIGHT", Line: line12, Body: "This assertion does not prove truncation preserves Unicode boundaries; assert the exact expected value.", Priority: review.PriorityMedium, Category: review.CategoryCorrectness, Status: review.StatusSuggested},
			},
		},
	}
	model.Session.Plan.ReviewOrder[0].Suggestions[0].Line = line14
	model.Session.Plan.ReviewOrder[0].Suggestions[1].Line = line18
	model.Session.Comments = nil
	for _, step := range model.Session.Plan.ReviewOrder {
		model.Session.Comments = append(model.Session.Comments, step.Suggestions...)
	}
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

func logViewSnapshot(t *testing.T, view string) {
	t.Helper()
	if os.Getenv("TRAIL_HUNK_TEST_RENDER") == "" {
		return
	}
	t.Logf("\n%s", view)
}

func contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}
