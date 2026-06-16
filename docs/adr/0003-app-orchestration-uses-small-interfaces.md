# ADR 0003: App Orchestration Uses Small Interfaces

## Status

Accepted

## Context

The startup flow touches several external boundaries: git, GitHub, AI commands,
diff parsing, and the TUI. Tests should not need real repositories, network
calls, or AI providers.

## Decision

Keep orchestration in `internal/app` and depend on small interfaces for:

- repository discovery
- GitHub PR context
- AI review generation

Concrete git, GitHub, and AI adapters are wired at the edge. The orchestration
layer produces a `review.ReviewSession` for the TUI.

## Consequences

End-to-end startup behavior can be tested with fakes while preserving concrete
package boundaries. The TUI stays isolated from git, GitHub, and AI provider
details.

