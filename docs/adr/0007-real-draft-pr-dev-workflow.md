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

Provide `mise run dev:dummy-pr` as a development workflow that creates an
isolated git worktree at `.dev/worktrees/dummy-pr`, checks out
`trail-hunk-dev/dummy-pr` from `origin/main`, seeds a small dummy review target,
pushes the branch, and opens or reuses a GitHub draft PR targeting `main`.

The task uses the developer's existing GitHub CLI authentication. A
`dev:dummy-pr:dry-run` task prints the planned actions without mutating git or
GitHub, and `dev:dummy-pr:reset` recreates the local worktree and branch.

## Consequences

The development workflow exercises the same branch and PR detection path as
real use. It requires network access and GitHub CLI auth, so it is intentionally
not part of the normal `mise run check` path.

Generated worktrees live under `.dev/`, which is ignored by git.
