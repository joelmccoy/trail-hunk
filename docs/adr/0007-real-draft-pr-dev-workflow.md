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

Provide `mise run dev` as the primary development workflow:

- It creates an isolated git worktree at `.dev/worktrees/dummy-pr`.
- It checks out `trail-hunk-dev/dummy-pr` from `origin/main`.
- It seeds a small dummy review target.
- It pushes the branch and opens or reuses a GitHub draft PR targeting `main`.
- It launches trail-hunk from the dummy PR worktree with
  `TRAIL_HUNK_PROVIDER=fixture`.
- The fixture AI provider returns deterministic overview text, risks,
  walkthrough steps, and suggested comments that map to real lines in the dummy
  PR diff.

Keep `dev:dummy-pr:*` as support tasks for setup, dry-run, and reset. Keep
`dev:review-current:*` as a secondary helper for testing the current checkout as
a PR, but do not make it the main local feature iteration path.

Both tasks use the developer's existing GitHub CLI authentication. Dry-run
tasks print planned actions without mutating git or GitHub. Reset tasks remove
the local generated worktree and branch.

## Consequences

The primary dev workflow exercises the same branch and PR detection path as real
use while making AI output deterministic. This is better for iterating on TUI
screens, walkthrough state, suggested comment decisions, queue behavior, and
review submission.

These workflows require network access and GitHub CLI auth, so they are
intentionally not part of the normal `mise run check` path.

Generated worktrees live under `.dev/`, which is ignored by git.
