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

