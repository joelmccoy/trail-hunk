# Atlas-Inspired Trail Hunk TUI Design

## Context

The target product direction is "CodeRabbit Atlas as a local, free, OSS TUI."
The public CodeRabbit Review material describes a PR review interface that
reorganizes pull requests from flat alphabetical file lists into guided
walkthroughs with logical cohorts and layers. Each layer anchors to exact diff
ranges, has its own summary, and can include diagrams when useful. The interface
uses three panels: cohort/layer navigation on the left, the active diff in the
center, and per-range context on the right. It also supports keyboard navigation,
focus mode, viewed state, native review submission, and stale snapshot
protection.

Trail Hunk should adopt the workflow ideas while staying terminal-native and
local-first.

References:

- CodeRabbit blog, "Introducing CodeRabbit Review: The first AI-native code
  review interface", May 13, 2026.
- CodeRabbit docs, "CodeRabbit Review".
- YouTube video link from the blog was reachable, but transcript content was
  not directly available from YouTube in this environment. Search snippets and
  the blog/docs provided the useful product details.

## Decision

Trail Hunk will model the guided review as a "change stack":

- change groups: conceptual cohorts of related work
- layers: ordered review steps inside a group
- layer-scoped diff chunks: focused code ranges for the current layer
- right-rail context: explanation for the currently visible layer

This builds on the existing `ReviewStep` concept by adding optional group and
layer metadata. When AI providers do not provide metadata, Trail Hunk derives a
reasonable fallback from file path and review order.

## Near-Term Implementation

Add these Atlas-inspired TUI behaviors now:

- prompt AI providers to return group/layer metadata
- carry group/layer metadata through `ai`, `app`, and `review` models
- show "Change Stack" in the left rail, grouped by change group and file
- show current layer and group in the header/rail
- add `z` focus mode to hide rail and right pane for full-width diff review
- add `v` viewed-state toggle for the current file, reflected in the rail

## Later Implementation

These should follow after the core review stack is stable:

- generated diagrams for steps where diagrams help
- snapshot generation ID and stale-state warnings
- code peek for definitions/usages
- right-pane range summary sync while scrolling within a long layer
- true all-files searchable view

## Acceptance Criteria

- `mise run check` passes.
- Fixture review data includes change group and layer metadata.
- The TUI left rail uses "Change Stack" language and groups layers.
- `z` toggles focus mode, hiding side panes and widening the diff pane.
- `v` marks the current file viewed and the rail shows that state.
- Existing review/comment workflows continue to work.
