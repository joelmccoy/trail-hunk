# ADR 0014: Diff-first review surface

## Status

Accepted

## Context

The previous walkthrough screen kept evolving from a three-column text report: navigation, diff, and a prose-heavy assistant panel. That layout made the diff too cramped, duplicated AI comment text, and forced the reviewer to scan competing columns instead of reviewing code.

The desired product is an AI-assisted PR review workbench. The code diff must be the primary reading surface; AI context and suggested comments should support review decisions without taking over the screen.

## Decision

The walkthrough screen is now diff-first:

- The left rail contains the review path, files, and queue status.
- The main area starts with a compact assistant insight strip.
- The diff viewport owns the majority of the screen width.
- Inline AI comments are compact markers only.
- The selected AI comment renders in a bottom review drawer with target, body, and actions.
- The prior right-side prose panel is removed from the primary layout.

## Consequences

This makes the first local TUI flow closer to a code-review workbench than a generated report. Future work should continue this direction by replacing remaining custom render helpers with focused components for the review path, diff gutter, assistant insight strip, and review drawer.
