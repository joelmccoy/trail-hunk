# trail-hunk

`trail-hunk` is an AI-assisted guided code review TUI for GitHub pull
requests. It runs inside a local git checkout, discovers the current PR,
fetches PR context from GitHub, shells out to local AI tools, and guides the
review file by file and hunk by hunk.

The first version is intentionally local-first:

- Go and Bubble Tea for the terminal UI.
- Native GitHub API calls using local auth from `GITHUB_TOKEN` or `gh auth token`.
- Shell-out AI providers for local `codex` and `claude` integrations.
- A Vim-like review workflow for accepting, editing, rejecting, and submitting
  review comments.

## Development

Install tools and hooks:

```sh
mise install
mise run install-hooks
```

Useful tasks:

```sh
mise run fmt
mise run tidy
mise run test
mise run vet
mise run check
```

Run the TUI directly:

```sh
go run ./cmd/trail-hunk
```

Inside the TUI, press `R` to initiate review generation for the current
branch's GitHub pull request. Press `q` to quit.

The TUI runs in Bubble Tea's alternate screen mode, so it behaves like a
full-screen terminal application and restores your shell when it exits.

### Local Dev Workflow

For local feature iteration, run one command:

```sh
mise run dev
```

The task creates or reuses a real draft GitHub PR from
`.dev/worktrees/dummy-pr`, launches trail-hunk from that PR worktree, and sets
`TRAIL_HUNK_PROVIDER=fixture`. Press `R` in the TUI to generate a deterministic
dummy review with overview text, walkthrough steps, risks, and suggested review
comments that map to real lines in the dummy PR diff.

Use this command to iterate on walkthrough, comment approval/rejection, the
comment queue, and final GitHub review submission without depending on live AI
output.

Support tasks:

```sh
mise run dev:dummy-pr:dry-run
mise run dev:dummy-pr
mise run dev:dummy-pr:reset
```

`dev:dummy-pr` prepares the PR without launching the TUI. The dry run prints the
planned git and GitHub actions. The reset task removes and recreates the local
dummy worktree and force-updates the dummy branch.

When you specifically need to test the current trail-hunk checkout as a PR, use:

```sh
mise run dev:review-current:setup
```

That helper creates `.dev/worktrees/review-current` from your current checkout,
but it is not the default workflow for iterating on dummy review data.

## Auth And Providers

GitHub auth is local-first. The app will use `GITHUB_TOKEN` when set and will
otherwise fall back to `gh auth token`.

```sh
gh auth login
TRAIL_HUNK_PROVIDER=codex go run ./cmd/trail-hunk
TRAIL_HUNK_PROVIDER=claude TRAIL_HUNK_MODEL=sonnet go run ./cmd/trail-hunk
TRAIL_HUNK_PROVIDER=fixture go run ./cmd/trail-hunk
```

The startup screen automatically detects the local repository and matching open
GitHub pull request for the current branch. When you press `R`, the app fetches
the raw diff, asks the selected AI provider for a structured review, and opens
the walkthrough screen.

## Architecture Decisions

Important architecture decisions are recorded as lightweight ADRs in
`docs/adr/`. Add a new ADR when changing package boundaries, provider behavior,
GitHub API strategy, diff/comment mapping, or TUI state architecture.
