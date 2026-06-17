# Trail Hunk Hybrid Review Workbench Design

## Context

The current walkthrough screen is useful as a prototype, but it is not the right
interaction model for sustained pull request review. It renders static panel
strings, uses boxes as the primary layout primitive, and shows only a narrow
slice of hunk data. That makes the app feel like a dashboard instead of a code
review tool.

This redesign keeps the AI guide, but makes the diff the persistent center of
the workflow. The AI review order becomes a navigation layer over a real
scrollable diff workspace.

Research inputs:

- Charm Bubbles provides production-ready Bubble Tea components for scrolling
  viewports, lists, text areas, help, and key bindings.
- The Bubbles viewport component is intended for vertically scrolling content,
  pager keys, mouse wheel input, and alternate-screen performance mode.
- Lazygit validates the terminal workbench pattern: persistent panes, focused
  navigation, and keyboard-driven git workflows.
- tuicr validates a code-review-first terminal workflow: review state, comments,
  and diff context should stay close together.

## Goals

- Show a real PR diff as the primary surface.
- Keep AI guidance visible without replacing the diff.
- Make every major region scrollable when content exceeds available space.
- Support Vim-style keyboard movement across panes, steps, suggestions, and
  comments.
- Preserve the current local-first architecture and fixture-backed dev command.
- Make the layout responsive from narrow laptop terminals to wide fullscreen
  terminals without hard-coded terminal assumptions.

## Non-Goals

- Do not build a hosted service.
- Do not implement perfect syntax highlighting in this pass.
- Do not replace the existing GitHub comment mapping or review submission
  boundaries.
- Do not add a complex plugin system for AI providers yet.

## Product Direction

Use a hybrid review workbench.

The main screen should feel like a code review session, not a slideshow. The
center pane is always the diff. The AI-generated walkthrough provides anchors
inside that diff: current step, current file, current hunk, and current suggested
comment. Moving through the walkthrough jumps and highlights the matching diff
lines instead of replacing the screen.

## Layout

The root model still owns the measured terminal size and reserves:

- a one-line status bar at the top
- a full-height workbench body
- a one-line command/help bar at the bottom

The workbench body has three logical panes on wide terminals:

- left rail: files and review steps
- center diff viewport: unified diff with old/new gutters
- right inspector: AI summary, why, risks, current suggestion, and actions

On narrower terminals, the rail and inspector collapse into toggleable drawers
or tabs while the diff remains the main surface. The user should never see a
half-width diff if the terminal is too narrow for useful code review.

No decorative card grid is used for the workbench. Borders are reserved for
pane boundaries and focus indication only. The visual hierarchy comes from
spacing, color, line gutters, and selected-row highlighting.

## Pane Behavior

### Diff Viewport

The diff viewport renders the current file/hunk from review-domain
`review.DiffLine` data. Each row includes:

- old line number
- new line number
- diff marker
- code text
- optional AI/comment marker

Added, deleted, and context lines receive distinct but restrained styling.
The selected review step highlights its relevant line range. The selected
suggested comment highlights its exact GitHub comment target line. Approved,
dismissed, and queued comments use different markers so review state is visible
without opening another screen.

The viewport owns vertical scroll state. `j/k`, arrow keys, page keys, `g`, and
`G` move inside the focused diff pane. `n/p` jumps between AI review anchors.

### File And Step Rail

The rail lists changed files and nested AI review steps. It should show review
progress at a glance:

- current file
- number of suggestions
- approved and dismissed counts
- risk badges by priority

The rail uses a scrollable list model. `f` toggles visibility. On wide screens it
is persistent; on narrow screens it opens as an overlay drawer.

### Inspector

The inspector shows information for the current anchor:

- brief step title
- one- or two-sentence summary of what changed
- why it matters
- priority/category for the selected suggestion
- suggested comment body
- available actions

The inspector is also scrollable. It is not a box of instructions; it is a
context panel for the currently selected diff line or AI step.

### Ask Drawer

The ask drawer opens over the bottom portion of the workbench with `t`. It uses a
text area for follow-up questions about the current PR, file, hunk, or selected
line. It should preserve the underlying workbench state and submit questions
with current review context attached.

In the first implementation pass, the drawer can be interactive UI plumbing with
a fixture response. Live provider follow-up can come after the layout is solid.

### Comment Queue

The queue view remains reachable with `C`, but it should share the workbench
visual language. It lists approved comments with target file/line, priority,
category, and body. `S` submits the pending GitHub review.

## Keyboard Model

Use `bubbles/key` for key bindings and `bubbles/help` for the footer. The footer
should derive from active key bindings, not a hand-written string.

Baseline keys:

- `q`: quit
- `R`: start or refresh AI review
- `tab`: cycle focused pane
- `h/l`: move focus left/right when panes are visible
- `j/k`: move within the focused pane
- `n/p`: next/previous AI review anchor
- `J/K`: next/previous suggested comment
- `a`: approve selected suggestion
- `d`: dismiss selected suggestion
- `e`: edit selected suggestion
- `r`: ask AI to reword selected suggestion
- `c`: create a manual comment on the selected diff line
- `C`: open comment queue
- `S`: submit approved review
- `f`: toggle file/step rail
- `t`: toggle ask drawer
- `/`: search diff text
- `?`: expand help

## State Model

Introduce focused workbench state instead of relying on one flat `Model` with
string flags.

Core additions:

- `WorkbenchModel`: owns pane focus, dimensions, selected file, selected step,
  selected suggestion, and child component models.
- `DiffViewport`: wraps `viewport.Model` and renders diff rows from
  review-domain data.
- `RailModel`: wraps a list-like model for files and AI anchors.
- `InspectorModel`: wraps `viewport.Model` for contextual AI and comment detail.
- `AskModel`: wraps `textarea.Model` for follow-up prompts.
- `KeyMap`: centralizes key bindings and help text.

Derived indexes:

- files by path
- steps by file and hunk
- suggestions by target line
- diff row index by GitHub comment side/line
- review anchor index by step ID

The review-domain `ReviewSession` remains the source of truth for plan,
comments, and statuses. TUI child models may cache rendered rows and scroll
positions, but they should not own canonical review state.

## API Boundaries

Do not move GitHub diff parsing into the TUI. `internal/app` should continue to
map GitHub hunks into review-domain diff data, consistent with ADR 0008.

The TUI can request:

- start review
- submit review
- ask follow-up question
- reword suggestion

Those stay as injected functions on the root model or a small orchestration
interface so tests can use deterministic fixtures.

## Error And Loading States

Startup, analysis, follow-up, and submission should use inline status regions,
not modal boxes unless the user must make a choice. The workbench should remain
visible while background work runs when possible.

Errors should appear in the status bar and an expandable detail area. They must
not corrupt the measured layout or cause footer wrapping.

## Implementation Slices

1. Add Bubbles components and central key map.
2. Split TUI rendering into root shell plus workbench child models.
3. Build the scrollable diff viewport from existing `review.DiffLine` data.
4. Add anchor navigation and line highlighting for current step/suggestion.
5. Add the file/step rail and responsive collapse behavior.
6. Add the inspector pane.
7. Add the ask drawer UI plumbing with fixture response support.
8. Update tests to verify measured layout, scrolling, anchor jumps, and comment
   status rendering.

## Acceptance Criteria

- `mise run check` passes.
- `mise run dev` opens a fullscreen TUI with a persistent diff surface.
- The diff pane shows real added/deleted/context lines from fixture PR data.
- `n/p` jumps between AI review anchors and highlights the relevant diff lines.
- `J/K` moves between suggestions and highlights exact target lines.
- `a/d` visibly changes suggestion state in the diff and inspector.
- The footer remains one line and is generated from active key bindings.
- Wide terminals show rail, diff, and inspector. Narrow terminals prioritize the
  diff and use toggles/drawers for secondary panes.
