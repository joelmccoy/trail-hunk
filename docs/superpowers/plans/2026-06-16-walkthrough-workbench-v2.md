# Walkthrough Workbench V2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the real Trail Hunk TUI match the approved v3 walkthrough mockup: changed-file rail, focused code chunk, inline findings, and chunk explainer.

**Architecture:** Keep the existing Bubble Tea root model and workbench components. Improve `internal/tui/diff_view.go` so findings render as annotation rows, `internal/tui/workbench.go` so the rail and inspector match the walkthrough model, and `internal/tui/model.go` so navigation can move by file and step.

**Tech Stack:** Go 1.26.2, Bubble Tea, Bubbles viewport/textarea/help/key, Lip Gloss.

---

### Task 1: Lock The Desired UI With Tests

**Files:**
- Modify: `internal/tui/model_test.go`

- [ ] Add tests proving: rail visible by default on wide screens, current file/step highlighted, code rows do not contain finding prose, findings render as annotation blocks, right pane explains chunks, and `]/[` navigates by changed file.
- [ ] Run `go test ./internal/tui -run 'TestWorkbench' -count=1` and confirm the new tests fail against the current UI.

### Task 2: Inline Finding Annotation Rows

**Files:**
- Modify: `internal/tui/diff_view.go`
- Modify: `internal/tui/model_test.go`

- [ ] Render code rows without appended finding body text.
- [ ] Render finding annotation rows below the target line, with priority/category/status/body/actions.
- [ ] Keep selected finding highlighting and approval/dismissal status visible.
- [ ] Run `go test ./internal/tui -count=1`.

### Task 3: Default Changed-File Rail And Review Map

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/workbench.go`
- Modify: `internal/tui/model_test.go`

- [ ] Show the file rail by default on wide terminals.
- [ ] Group review steps by file path.
- [ ] Highlight the current file and current step.
- [ ] Show suggestion counts and review progress.
- [ ] Run `go test ./internal/tui -count=1`.

### Task 4: Chunk Explainer Right Pane

**Files:**
- Modify: `internal/tui/workbench.go`
- Modify: `internal/tui/model_test.go`

- [ ] Replace finding-detail inspector content with current chunk explanation.
- [ ] Render sections for what the chunk does, why it matters, used by/impact, review guidance, and confidence summary.
- [ ] Keep finding text out of the right pane's primary content.
- [ ] Run `go test ./internal/tui -count=1`.

### Task 5: File Navigation

**Files:**
- Modify: `internal/tui/keys.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/model_test.go`

- [ ] Add `]` and `[` bindings for next/previous changed file.
- [ ] Move to the first walkthrough step for the target file.
- [ ] Keep `n/p` as step navigation and `J/K` as finding navigation.
- [ ] Run `go test ./internal/tui -count=1`.

### Task 6: Verification And Commit

**Files:**
- Modify if needed: `docs/manual-test.md`

- [ ] Run `mise run check`.
- [ ] Smoke `mise run dev`, press `R`, inspect the workbench, and quit.
- [ ] Commit implementation with `git commit -m "feat: refine walkthrough workbench ui"`.
