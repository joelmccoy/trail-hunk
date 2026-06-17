# ADR 0009: Use a hybrid scrollable review workbench

## Status

Accepted

## Context

Trail Hunk needs to support sustained pull request review in a terminal. The
current TUI renders static string panels in a grid. That was enough to validate
startup detection, fixture AI output, comment approval, and review submission,
but it does not feel like a code review tool.

The main user workflow is reading code. AI guidance should orient and annotate
the review, not replace the diff as the primary surface.

Charm's Bubbles project provides Bubble Tea components for scrolling viewports,
lists, text areas, key bindings, and help. Those are better primitives for this
workflow than hand-rendered panel strings.

## Decision

Replace the card-grid walkthrough with a hybrid workbench:

- a persistent central diff viewport
- a file/step rail for navigation and review progress
- an inspector pane for AI explanation and selected comment details
- a toggleable ask drawer for follow-up questions
- key bindings and footer help generated from a central key map

The AI review plan becomes an anchor map over the diff. Moving through review
steps jumps the diff viewport to relevant hunks and highlights the affected
lines.

## Consequences

The TUI will gain more internal structure: root shell, workbench model, diff
viewport, rail, inspector, ask drawer, and key map. This is intentional because
the current monolithic renderer has reached its limit.

The review-domain model remains the source of truth. GitHub diff parsing and
comment target validation stay outside the TUI.

The first implementation pass should improve interaction quality before adding
more AI behavior. Follow-up questions, edit/reword flows, and richer provider
integration can build on the workbench once navigation and rendering are solid.
