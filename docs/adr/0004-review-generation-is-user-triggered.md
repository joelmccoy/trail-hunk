# ADR 0004: Review Generation Is User-Triggered

## Status

Accepted

## Context

Generating a guided review requires local git discovery, GitHub API calls, raw
diff parsing, and an AI provider invocation. Running that work immediately on
TUI launch can make startup feel stuck and may spend AI calls before the user is
ready.

## Decision

Start the TUI quickly and require the user to press `R` to initiate review
generation. While review generation runs, show a loading state. When it
finishes, load the generated review session into the walkthrough screen or show
an actionable error.

## Consequences

The initial app experience is responsive and predictable. The TUI needs an
asynchronous review-start command and clear error states. This also creates a
natural place to add provider/model selection before review generation.

