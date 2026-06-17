# ADR 0008: Review steps store mapped diff lines

## Status

Accepted

## Context

The walkthrough needs to show code alongside AI explanations and suggested
comments. GitHub diff parsing currently lives in `internal/github`, and the TUI
should not need to understand GitHub hunk parsing, side semantics, or raw patch
format.

The app orchestration already parses the pull request diff so it can validate
AI-suggested comment targets. That is the right boundary to attach display-ready
hunk data to each review step.

## Decision

Store a small `review.DiffLine` slice on each `review.ReviewStep`. The app
orchestrator maps parsed GitHub hunk lines into this review-domain shape while
building the `ReviewSession`.

The TUI renders only review-domain data: step explanation, mapped diff lines,
and suggested comments. It does not parse raw GitHub diff text.

## Consequences

Walkthrough rendering is deterministic and testable without live GitHub calls.
The review session now carries more display data, but the extra state is scoped
to the hunk lines needed by the guided walkthrough.

If later views need whole-file context, add that as a separate review-domain
projection instead of passing raw GitHub diff internals into the TUI.
