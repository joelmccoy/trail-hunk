# Hybrid Review Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the static walkthrough card grid with a responsive, scrollable hybrid review workbench centered on a real PR diff.

**Architecture:** Keep the root Bubble Tea model as the app shell, but split review rendering into focused child helpers for key bindings, diff rows, workbench layout, and contextual panes. The review-domain `ReviewSession` remains the source of truth; the TUI caches only viewport/focus state.

**Tech Stack:** Go 1.26.2, Bubble Tea, Lip Gloss, Charm Bubbles `viewport`, `textarea`, `key`, and `help`.

---

### Task 1: Add Component Dependencies And Key Map

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Create: `internal/tui/keymap.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing footer/key-map tests**

Add tests that assert the rendered footer contains generated help labels and stays one line at common widths:

```go
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
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `go test ./internal/tui -run TestFooterUsesGeneratedKeyHelp -count=1`

Expected: FAIL because the current footer is hand-written and does not include `tab focus`.

- [ ] **Step 3: Add Bubbles and central key map**

Run: `go get github.com/charmbracelet/bubbles@latest`

Create `internal/tui/keymap.go` with a `keyMap` type using `bubbles/key` and `bubbles/help`. Include short help for review, focus, movement, step navigation, suggestion navigation, approve, dismiss, queue, submit, files, ask, search, help, and quit.

- [ ] **Step 4: Render footer from key map**

Replace the hand-written footer text with `help.Model.View(keys)` output and constrain it to the measured footer width.

- [ ] **Step 5: Run tests**

Run: `go test ./internal/tui -count=1`

Expected: PASS.

### Task 2: Build Diff Row Rendering For A Persistent Viewport

**Files:**
- Create: `internal/tui/diff_view.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing diff rendering tests**

Add tests for rendering a unified diff surface from `review.DiffLine` data:

```go
func TestWorkbenchRendersUnifiedDiffRows(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 120
	model.Height = 32

	view := model.View()

	for _, want := range []string{"old", "new", "+", "-", "func newName() {}", "func oldName() {}"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `go test ./internal/tui -run TestWorkbenchRendersUnifiedDiffRows -count=1`

Expected: FAIL if the static diff panel is still the only implementation.

- [ ] **Step 3: Implement diff row helpers**

Create a `diffRow` projection with old line, new line, marker, text, row kind, and target marker. Add rendering helpers that style added, deleted, context, selected step, and selected suggestion rows.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/tui -count=1`

Expected: PASS.

### Task 3: Introduce The Workbench Model And Scrollable Diff Viewport

**Files:**
- Create: `internal/tui/workbench.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing workbench layout tests**

Add tests that assert walkthrough rendering no longer uses rounded card borders, fills measured width, and keeps within terminal height:

```go
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
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `go test ./internal/tui -run TestWalkthroughUsesWorkbenchNotCardGrid -count=1`

Expected: FAIL because the current walkthrough uses panel borders.

- [ ] **Step 3: Implement workbench layout**

Create `WorkbenchModel` with focus, width, height, `viewport.Model` for diff and inspector, and methods to sync from the current `ReviewSession`. Render wide terminals as rail, diff, and inspector columns. Render narrow terminals as diff-first with optional rail/inspector drawers.

- [ ] **Step 4: Wire root model to workbench**

Add a `Workbench WorkbenchModel` field to `Model`. On review start, step changes, suggestion changes, and window resize, sync workbench dimensions and content before rendering `ScreenWalkthrough`.

- [ ] **Step 5: Run tests**

Run: `go test ./internal/tui -count=1`

Expected: PASS.

### Task 4: Add Anchor Navigation And Line Highlighting

**Files:**
- Modify: `internal/tui/workbench.go`
- Modify: `internal/tui/diff_view.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing anchor tests**

Add tests that `n/p` changes selected step, `J/K` changes selected suggestion, and both updates appear as highlighted diff markers:

```go
func TestSuggestionNavigationHighlightsTargetLine(t *testing.T) {
	model := walkthroughModelWithDiff()
	model.Width = 120
	model.Height = 32

	updated, _ := model.Update(key("J"))
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, ">>") {
		t.Fatalf("selected suggestion target was not highlighted:\n%s", view)
	}
	if !strings.Contains(view, "Check helper visibility.") {
		t.Fatalf("inspector did not show selected suggestion:\n%s", view)
	}
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `go test ./internal/tui -run TestSuggestionNavigationHighlightsTargetLine -count=1`

Expected: FAIL because `J/K` are not implemented as suggestion navigation and the diff marker is not `>>`.

- [ ] **Step 3: Implement anchor and suggestion navigation**

Add `J/K` handling, keep `j/k` for focused-pane scrolling, and make `n/p` jump the diff viewport to the current step's first relevant row. Render selected suggestion target with `>>` and current step rows with a subtler focus style.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/tui -count=1`

Expected: PASS.

### Task 5: Add Inspector, Queue State, And Ask Drawer Shell

**Files:**
- Modify: `internal/tui/workbench.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing interaction tests**

Add tests that `a/d` updates status in both inspector and diff markers, `C` renders the queue without decorative cards, and `t` opens an ask drawer over the workbench:

```go
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
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `go test ./internal/tui -run TestApproveUpdatesWorkbenchSuggestionState -count=1`

Expected: FAIL because the current walkthrough does not expose approved state in the diff workbench.

- [ ] **Step 3: Implement inspector and drawer rendering**

Render the current step summary, why, selected suggestion body, priority, category, and status in the inspector viewport. Render the ask drawer as a bottom overlay using `textarea.Model` when `t` is active. Keep queue rendering in the same plain workbench visual language.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/tui -count=1`

Expected: PASS.

### Task 6: Verify Full Dev Workflow

**Files:**
- Modify if needed: `README.md`
- Modify if needed: `docs/manual-test.md`

- [ ] **Step 1: Run full checks**

Run: `mise run check`

Expected: PASS.

- [ ] **Step 2: Run the fixture PR workflow**

Run: `mise run dev`

Expected: The app opens in fullscreen with a persistent central diff, file/step rail on wide terminals, inspector pane, one-line generated footer, working `n/p`, `J/K`, `a/d`, `f`, `t`, and `C`.

- [ ] **Step 3: Commit implementation**

Run:

```bash
git add go.mod go.sum internal/tui docs README.md
git commit -m "feat: add hybrid review workbench"
```
