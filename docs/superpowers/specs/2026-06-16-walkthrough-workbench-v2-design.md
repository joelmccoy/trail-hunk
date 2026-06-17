# Trail Hunk Walkthrough Workbench V2 Design

## Context

The current hybrid workbench improved the previous card-grid layout, but it
still fails as a production review UI:

- suggested comments are appended to code rows, which corrupts the diff
- the changed-file rail is hidden by default
- the right pane acts like a finding detail panel instead of explaining the
  current walkthrough chunk
- clipped text makes the guide hard to read
- there is no clear sense of file position across a multi-file PR

The approved v3 mockup changes the product model: Trail Hunk is a sequential PR
walkthrough. The center pane shows one coherent code chunk at a time. Findings
render inline in the diff. The right pane explains the current code chunk. The
left pane is the PR map.

## Decision

Build the review screen around three persistent panes on wide terminals:

- left: changed files and nested walkthrough steps
- center: focused diff chunk with inline AI findings
- right: current chunk explanation

The current file and current step are highlighted in the left rail. Multiple
files can be visible in the rail, and Vim-like navigation can move by pane,
step, file, finding, and scroll position.

## Diff Pane

The diff pane must preserve code readability. Code rows contain only diff code:
line marker, old/new line numbers, diff sign, and source text.

AI suggested comments render as separate annotation blocks immediately below the
target line. Annotation blocks include priority, category, status, body, and
available actions. The selected finding receives stronger highlight treatment.

The center pane shows the current walkthrough step's mapped diff lines. Future
work can expand this into a continuous full-PR diff, but the current
implementation should still communicate file/hunk/step position clearly.

## Right Pane

The right pane is not a findings list. It explains the current walkthrough
chunk:

- what this chunk does
- why it changed or why it matters
- likely downstream usage or impact
- how to review the chunk
- confidence summary

Suggested comments remain inline in the diff pane. The right pane may mention
finding counts and queue counts, but it should not duplicate full finding text.

## Left Rail

The changed-file rail is visible by default on wide terminals. It groups
walkthrough steps by file and shows:

- changed file path
- current file highlight
- nested step titles
- current step highlight
- suggestion counts
- review progress counts

On narrow terminals the app may collapse the rail, but wide fullscreen terminals
should show it without requiring a toggle.

## Navigation

Baseline keys:

- `tab`: cycle focused pane
- `j/k`: move within focused pane
- `n/p`: next/previous walkthrough step
- `]/[`: next/previous changed file
- `J/K`: next/previous finding in the current chunk
- `a/d/e/r`: approve, dismiss, edit, reword selected finding
- `f`: toggle the file rail
- `t`: ask about current step
- `C`: comment queue

## Acceptance Criteria

- `mise run check` passes.
- `mise run dev` opens with changed-file rail visible on wide terminals.
- The current file and current walkthrough step are highlighted in the rail.
- The diff pane shows code rows without appended comment prose.
- AI findings render as separate inline annotation blocks below target lines.
- The right pane explains the current chunk and does not duplicate finding body
  text as its primary content.
- `]/[` moves to the next/previous file's first walkthrough step.
- Existing approval/dismissal flow still updates inline annotation status.
