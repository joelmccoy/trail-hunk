# Atlas Change Stack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Atlas-inspired change stack concepts to Trail Hunk: groups, layers, focus mode, and viewed file progress.

**Architecture:** Preserve the flat executable `ReviewOrder`, but enrich each step with optional group/layer metadata. Render those fields in the TUI rail and guide pane, and keep keyboard-driven state in the root TUI model.

**Tech Stack:** Go 1.26.2, Bubble Tea, Bubbles, Lip Gloss.

---

### Task 1: Change Stack Metadata

**Files:**
- Modify: `internal/ai/provider.go`
- Modify: `internal/review/session.go`
- Modify: `internal/app/orchestration.go`
- Modify: `internal/ai/prompts.go`
- Modify: `internal/ai/fixture.go`
- Tests: `internal/ai/schema_test.go`, `internal/app/orchestration_test.go`

- [x] Add `group_id`, `group_title`, `layer_index`, and `layer_title` to AI review steps.
- [x] Carry those fields into review-domain steps.
- [x] Prompt providers to organize review order as change groups and layers.
- [x] Update fixture review output with group/layer metadata.
- [x] Verify with focused AI and app tests.

### Task 2: TUI Change Stack Rail

**Files:**
- Modify: `internal/tui/workbench.go`
- Tests: `internal/tui/model_test.go`

- [x] Rename the rail to `Change Stack`.
- [x] Group steps by group title when metadata exists.
- [x] Keep ungrouped files visible under `Other changes`.
- [x] Show layer number/title and original step title when they differ.
- [x] Show viewed file progress.

### Task 3: Focus And Viewed State

**Files:**
- Modify: `internal/tui/keys.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/workbench.go`
- Tests: `internal/tui/model_test.go`

- [x] Add `z` focus mode.
- [x] Add `v` viewed toggle for the current file.
- [x] Hide side panes in focus mode.
- [x] Render viewed status in the rail.

### Task 4: Verification

**Commands:**

- [x] `go test ./...`
- [x] `mise run check`
- [x] `mise run dev`, press `R`, smoke focus/viewed keys, quit.
