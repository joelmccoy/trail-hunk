# ADR 0012: Componentized workbench rendering

## Status

Accepted

## Context

The walkthrough TUI was readable in tests but still felt sloppy in real terminal screenshots. The diff pane, change stack, inspector, inline findings, and footer were assembled as independent text blobs. That made the UI hard to reason about, easy to overfill, and prone to clipped help text or oversized inline comments.

trail-hunk needs to feel like a production review workbench: scrollable panes, compact review affordances, and stable responsive behavior built from Charm primitives.

## Decision

The walkthrough workbench will use small internal view components:

- A measured pane shell for rail, diff, and guide panes.
- Bubbles `viewport` components for scrollable diff and guide content.
- A compact command-bar component instead of raw generated help text on the walkthrough screen.
- A focused diff window that centers around AI finding targets and elides unrelated context.
- Inline finding components that render selected comments with actions and non-selected comments as single-line markers.

The left rail remains the PR walkthrough map. The right guide remains explanatory context for the selected chunk, not a findings panel.

## Consequences

This keeps the first version terminal-native and testable without adding a heavier rendering framework. Future TUI work should add behavior through these measured components instead of appending free-form raw text directly into panes.
