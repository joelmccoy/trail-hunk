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

Run the placeholder TUI:

```sh
go run ./cmd/trail-hunk
```

Inside the TUI, press `R` to initiate review generation for the current
branch's GitHub pull request. Press `q` to quit.

## Auth And Providers

GitHub auth is local-first. The app will use `GITHUB_TOKEN` when set and will
otherwise fall back to `gh auth token`.

```sh
gh auth login
TRAIL_HUNK_PROVIDER=codex go run ./cmd/trail-hunk
TRAIL_HUNK_PROVIDER=claude TRAIL_HUNK_MODEL=sonnet go run ./cmd/trail-hunk
```

The review flow requires an open GitHub pull request for the current branch.
When you press `R`, the app resolves the local repository, finds the matching
PR, fetches its raw diff, asks the selected AI provider for a structured review,
and opens the walkthrough screen.

## Architecture Decisions

Important architecture decisions are recorded as lightweight ADRs in
`docs/adr/`. Add a new ADR when changing package boundaries, provider behavior,
GitHub API strategy, diff/comment mapping, or TUI state architecture.
