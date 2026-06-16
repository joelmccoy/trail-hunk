# Manual MVP Verification

Use this checklist when testing `trail-hunk` against a real GitHub pull request.

## Local Setup

```sh
mise install
mise run install-hooks
gh auth login
```

Optional provider selection:

```sh
export TRAIL_HUNK_PROVIDER=codex
export TRAIL_HUNK_MODEL=gpt-5
```

or:

```sh
export TRAIL_HUNK_PROVIDER=claude
export TRAIL_HUNK_MODEL=sonnet
```

## Current Automated Coverage

- Git repository remote parsing and branch discovery.
- GitHub token discovery from `GITHUB_TOKEN` and `gh auth token`.
- GitHub pull request lookup by current branch.
- Raw GitHub PR diff fetch using the diff media type.
- Unified diff parsing into files, hunks, and commentable line targets.
- Shell runner and strict AI review JSON validation.
- Review session comment accept, edit, dismiss, manual-add, and approved queue.
- Bubble Tea key handling for quit, step navigation, file tree toggle, and ask
  pane toggle.
- GitHub pull request review payload construction and submission call.
- App orchestration with fake git, GitHub, and AI dependencies.

## Manual Smoke Test

From this repository:

```sh
mise run check
go run ./cmd/trail-hunk
```

Expected today:

- The TUI starts.
- The TUI uses the full-screen alternate terminal screen and restores the shell
  after exit.
- Startup detection displays repository information and PR information when a
  PR exists for the current branch.
- Pressing `R` initiates review generation.
- If no PR exists for the current branch, the startup screen shows an actionable
  error.
- Pressing `q` exits.
- No panic or terminal corruption occurs.

## Real PR Test Checklist

Use a repository with an open PR on the current branch.

- [ ] Start `trail-hunk` from the repository root.
- [ ] Press `R` to initiate review generation.
- [ ] Confirm the app resolves owner, repository, branch, and PR.
- [ ] Confirm GitHub context fetches PR metadata and raw diff.
- [ ] Confirm selected AI provider generates a structured overview.
- [ ] Confirm walkthrough order appears in the TUI.
- [ ] Accept one AI-suggested comment.
- [ ] Edit one accepted AI comment.
- [ ] Reject one AI-suggested comment.
- [ ] Add one manual line comment.
- [ ] Submit a `COMMENT` review to GitHub.

The current implementation has the underlying API boundaries, session state,
review startup command, and walkthrough loading path in place. The next focus is
making the walkthrough panes richer and adding in-TUI editing/submission flows.
