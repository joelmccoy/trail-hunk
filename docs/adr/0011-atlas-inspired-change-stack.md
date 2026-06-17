# ADR 0011: Model reviews as a change stack

## Status

Accepted

## Context

Trail Hunk aims to be a local, OSS, terminal-native version of the AI-native PR
review workflow described by CodeRabbit Review/Atlas. The important product idea
is not a web layout; it is the shift from flat file-order review to a guided
story of the change.

The current Trail Hunk model has a flat `ReviewOrder` slice. That is enough for
linear walkthroughs, but it does not expose logical cohorts/layers or progress
in the way reviewers understand larger changes.

## Decision

Keep `ReviewOrder` as the executable sequence, but enrich each review step with
optional change-stack metadata:

- group ID/title for the conceptual cohort
- layer title and layer index for ordered review within a group

The TUI will render these as the primary navigation structure. AI providers are
prompted to emit this metadata. When they do not, Trail Hunk falls back to file
path and step order.

## Consequences

This is backward compatible with existing providers and fixtures. The TUI can
start presenting a stronger guided-review model immediately, while future AI
provider improvements can produce richer grouping.

The model remains local and GitHub-native: comments still map to real PR diff
lines and submit through GitHub reviews.
