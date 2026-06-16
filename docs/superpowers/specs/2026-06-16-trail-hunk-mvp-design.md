# trail-hunk MVP Design

## Goal

Build a local-first Go terminal UI that guides a user through reviewing the
current GitHub pull request with help from a local AI coding assistant.

The user runs `trail-hunk` from a git checkout. The app discovers the current
branch's GitHub PR, fetches PR metadata and diff context, asks a selected local
AI tool to produce a review overview and walkthrough plan, then presents a
Vim-like review workflow where the user can accept, edit, reject, or write
comments before submitting a GitHub PR review.

## MVP Scope

In scope:

- Existing GitHub PR for the current branch.
- Native GitHub API calls from Go.
- Local auth discovery from `GITHUB_TOKEN`, then `gh auth token`.
- Shell-out AI providers for `codex` and `claude`.
- Structured AI responses for overview, risks, review order, and suggested
  comments.
- Bubble Tea TUI with Vim-like navigation.
- Guided walkthrough by file and hunk.
- Comment queue with AI and user-authored comments.
- Final PR review submission with approved comments.

Out of scope for v1:

- Hosted service or background daemon.
- Direct OpenAI or Anthropic API integrations.
- Reviewing local-only or unpushed changes.
- Multi-user collaboration.
- Persistent cloud storage.

## Architecture

`trail-hunk` is a local orchestrator. It owns the review session state, calls
GitHub directly, and treats AI tools as external local commands behind a stable
provider interface.

The core dependency direction is:

```text
cmd -> app -> git/github/ai/review/tui/storage
```

The TUI should not call GitHub, git, or AI providers directly. It sends user
intent to the review session model, receives state updates, and renders the
current mode.

## Package Structure

```text
cmd/trail-hunk/
  main.go

internal/app/
  app.go
  config.go

internal/git/
  repo.go
  diff.go

internal/github/
  auth.go
  client.go
  pr.go
  review.go
  diffmap.go

internal/ai/
  provider.go
  shell.go
  codex.go
  claude.go
  prompts.go
  schema.go

internal/review/
  session.go
  plan.go
  comments.go
  priority.go

internal/tui/
  model.go
  keys.go
  screens.go
  diffview.go
  walkthrough.go
  comments.go
  ask.go
  files.go
  status.go

internal/storage/
  cache.go
```

## Data Model

The review session is the central application state.

```go
type ReviewSession struct {
    Repo     RepoRef
    PR       PullRequest
    Files    []ChangedFile
    Diff     PullRequestDiff
    Plan     WalkthroughPlan
    Cursor   ReviewCursor
    Comments []ReviewComment
    Provider AIProviderRef
}
```

The AI walkthrough plan is user-facing and should explain both what changed and
why each step matters.

```go
type WalkthroughPlan struct {
    Overview    string
    Risks       []Risk
    ReviewOrder []ReviewStep
}

type ReviewStep struct {
    ID          string
    FilePath    string
    HunkID      string
    Title       string
    Summary     string
    Why         string
    Focus       []string
    Suggestions []SuggestedComment
}
```

Comments move through explicit states:

```go
type ReviewComment struct {
    FilePath  string
    Side      string
    Line      int
    StartLine *int
    Body      string
    Priority  Priority
    Category  CommentCategory
    Status    CommentStatus
    Source    CommentSource
}
```

Priorities: `blocker`, `high`, `medium`, `low`, `note`.

Categories: `bug`, `security`, `correctness`, `maintainability`,
`performance`, `test`, `question`.

Statuses: `suggested`, `approved`, `edited`, `dismissed`, `submitted`.

## GitHub API Flow

1. Discover the local repo with git:
   - repository root
   - current branch
   - configured GitHub remote
   - owner and repository name

2. Discover auth:
   - prefer `GITHUB_TOKEN`
   - fall back to `gh auth token`
   - fail with a clear setup message if neither works

3. Discover the PR:
   - query open PRs for the current branch and repository
   - select the only match automatically
   - show a PR picker if multiple matches are possible
   - exit clearly if no PR exists

4. Fetch PR context:
   - PR title, body, author, base, head, and state
   - changed files
   - commits
   - existing review comments
   - existing issue comments
   - raw unified PR diff

5. Parse the diff:
   - files
   - hunks
   - added, deleted, and context lines
   - stable hunk IDs
   - valid GitHub review comment targets

6. Submit review:
   - convert approved comments to GitHub review comment payloads
   - submit one pull request review
   - use review event `COMMENT` for the MVP

## Diff Mapping

GitHub PR review comments must target lines present in the PR diff. AI output is
not trusted for final line mapping.

The app maintains a parsed diff table:

```go
type DiffLine struct {
    FilePath   string
    HunkID     string
    Kind       DiffLineKind
    OldLine    *int
    NewLine    *int
    DiffLineNo int
    Side       string
    CanComment bool
}
```

For review submissions, prefer the modern GitHub fields `line`, `side`,
`start_line`, and `start_side` instead of legacy `position`.

Rules:

- Added lines target `RIGHT` and `NewLine`.
- Deleted lines target `LEFT` and `OldLine`.
- Context lines may target `RIGHT` in the MVP unless GitHub validation requires
  side-specific handling for multi-line ranges.
- AI suggested comments are validated against the diff table before display.
- Invalid suggestions are surfaced as untargeted notes, not submitted comments.

## AI Provider Boundary

AI integrations shell out to local tools. The rest of the app depends only on a
provider interface.

```go
type Provider interface {
    Name() string
    Models(ctx context.Context) ([]ModelRef, error)
    Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error)
    Ask(ctx context.Context, req AskRequest) (AskResponse, error)
    Reword(ctx context.Context, req RewordRequest) (RewordResponse, error)
}
```

Provider adapters:

- `codex`: invokes the local Codex CLI.
- `claude`: invokes the local Claude Code CLI.

Provider responses must be structured JSON. The app validates JSON shape,
comment targets, categories, priorities, and maximum body sizes before placing
content into the review session.

## TUI

Main screens:

- Startup and PR resolve
- Provider and model selection
- Overview
- Walkthrough
- Comment queue
- Submit review

Walkthrough layout:

```text
+--------------------------------------------------------------+
| repo / PR / file / step / provider                           |
+---------------+--------------------------+-------------------+
| file tree     | diff panel               | AI notes/comments |
| toggle        |                          |                   |
+---------------+--------------------------+-------------------+
| ask pane toggle: current step/file/hunk follow-up question    |
+--------------------------------------------------------------+
| keys / mode / pending comments / errors                       |
+--------------------------------------------------------------+
```

Keyboard model:

- `j` and `k`: move within current pane
- `h` and `l`: switch panes
- `n` and `p`: next or previous review step
- `]f` and `[f`: next or previous file
- `]h` and `[h`: next or previous hunk
- `a`: accept suggestion
- `e`: edit comment
- `r`: reword selected comment with AI
- `d`: dismiss suggestion
- `c`: create manual comment
- `?`: ask follow-up question
- `t`: toggle ask pane
- `f`: toggle file tree
- `q`: quit or back
- `S`: submit review

Each walkthrough step should show a brief summary and a short "why this matters"
explanation at the top of the current pane.

## Error Handling

Errors should be recoverable whenever possible:

- Missing git repo: show a startup error.
- Unsupported remote: show expected GitHub remote formats.
- Missing auth: show `gh auth login` and `GITHUB_TOKEN` options.
- No PR found: explain that v1 requires an existing PR for the current branch.
- AI command missing: show provider installation guidance.
- Invalid AI JSON: show retry option and keep raw provider output in debug logs.
- Invalid comment target: convert to a non-submittable note.
- GitHub submission failure: keep approved comments in the queue.

## Testing Strategy

Unit tests:

- git remote parsing
- PR discovery query construction
- diff parser and diff line mapping
- GitHub review payload generation
- AI response validation
- review session state transitions

Integration-style tests:

- fake GitHub client for PR context and submission
- fake AI provider for walkthrough generation
- Bubble Tea model update tests for key workflows

Manual verification:

- run in a real repository with an open PR
- fetch context
- generate a walkthrough with `codex`
- generate a walkthrough with `claude`
- approve and submit a small review

## Milestones

1. Repository baseline with Go, Bubble Tea, `mise`, and `hk`.
2. GitHub auth, repo discovery, and PR lookup.
3. PR diff parser and comment target validator.
4. Shell AI provider interface and provider adapters.
5. Review session model and walkthrough state machine.
6. Bubble Tea overview and walkthrough screens.
7. Comment queue, edit, reject, reword, and manual comments.
8. GitHub review submission.
9. Polish, retries, debug logs, and local session cache.

