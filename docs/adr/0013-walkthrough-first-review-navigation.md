# ADR 0013: Walkthrough-first review navigation

## Status

Accepted

## Context

trail-hunk is an AI-assisted PR walkthrough tool, not only a findings browser. The prior workbench rail grouped files, layers, and finding counts together. In practice this made the UI feel like an internal data dump: the reviewer saw `layers` and `findings` before they saw the intended review path through the PR.

The product goal is different: the AI should create a sequential walkthrough of the PR, explain each step, and attach suggested comments to the relevant diff lines.

## Decision

The primary left navigation is the PR walkthrough path:

- Ordered review steps are first-class navigation items.
- Changed files remain visible, but secondary.
- Suggested comments/findings render inline in the diff and in the review queue, not as the primary navigation hierarchy.
- User-facing terminology uses `step`, `file`, `suggested comment`, and `review queue`; it avoids internal `layer` and `finding` labels in the main workbench chrome.

## Consequences

The TUI now optimizes for guided understanding before comment management. Future AI schema improvements should produce better walkthrough steps, not just more findings. Review comment state can still be summarized, but it should not replace the PR walkthrough as the main interaction model.
