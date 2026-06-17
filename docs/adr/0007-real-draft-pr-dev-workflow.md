# ADR 0007: Use a real draft PR for local workflow testing

## Status

Accepted

## Context

trail-hunk's core behavior depends on local git branch detection, GitHub PR
lookup, diff fetching, diff-position mapping, AI review generation, and pending
review submission. A local-only fake PR fixture would make the app easy to
start, but it would skip the GitHub behavior most likely to break.

Developers still need a fast, repeatable way to test workflows without editing
the main checkout directly.

## Decision

Provide two real-PR development workflows:

- `mise run dev:review-current` creates an isolated git worktree at
  `.dev/worktrees/review-current`, checks out
  `trail-hunk-dev/review-current` from the current checkout's `HEAD`, overlays
  the current tracked diff plus untracked non-ignored files, commits that
  snapshot only in the generated worktree, pushes the branch, opens or reuses a
  GitHub draft PR targeting `main`, and launches trail-hunk from that worktree.
- `mise run dev:dummy-pr` creates an isolated git worktree at
  `.dev/worktrees/dummy-pr`, checks out `trail-hunk-dev/dummy-pr` from
  `origin/main`, seeds a small dummy review target, pushes the branch, and
  opens or reuses a GitHub draft PR targeting `main`.

Both tasks use the developer's existing GitHub CLI authentication. Dry-run
tasks print planned actions without mutating git or GitHub. Reset tasks remove
the local generated worktree and branch.

## Consequences

The development workflows exercise the same branch and PR detection path as
real use. `dev:review-current` is the fastest way to test local trail-hunk code,
including uncommitted changes, without modifying the active checkout.

These workflows require network access and GitHub CLI auth, so they are
intentionally not part of the normal `mise run check` path.

Generated worktrees live under `.dev/`, which is ignored by git.
