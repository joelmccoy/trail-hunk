# trail-hunk MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the local-first guided GitHub PR review TUI described in `docs/superpowers/specs/2026-06-16-trail-hunk-mvp-design.md`.

**Architecture:** Keep orchestration in `internal/app`, API boundaries in `internal/git`, `internal/github`, and `internal/ai`, review state in `internal/review`, and Bubble Tea rendering in `internal/tui`. The TUI sends intent into the review session model and never calls git, GitHub, or AI providers directly.

**Tech Stack:** Go 1.26.2, Bubble Tea, native GitHub REST/GraphQL over `net/http`, local git commands, `gh auth token` for auth fallback, shell-out `codex` and `claude` providers, `mise`, and `hk`.

---

## File Structure

- Create `internal/git/repo.go`: repository root, branch, and GitHub remote discovery.
- Create `internal/git/repo_test.go`: repo discovery tests using temporary git repositories.
- Create `internal/github/auth.go`: token lookup from environment and `gh`.
- Create `internal/github/client.go`: authenticated HTTP client and typed API helpers.
- Create `internal/github/pr.go`: PR lookup and context fetch.
- Create `internal/github/review.go`: GitHub pull request review submission.
- Create `internal/github/diffmap.go`: unified diff parser and review target validator.
- Create `internal/github/diffmap_test.go`: diff parsing and line mapping tests.
- Create `internal/ai/provider.go`: provider interfaces and request/response types.
- Create `internal/ai/shell.go`: command runner with timeout and stderr capture.
- Create `internal/ai/codex.go`: Codex CLI adapter.
- Create `internal/ai/claude.go`: Claude CLI adapter.
- Create `internal/ai/schema.go`: AI JSON validation.
- Create `internal/ai/schema_test.go`: validation tests.
- Create `internal/review/session.go`: review session state and commands.
- Create `internal/review/comments.go`: comment state transitions.
- Create `internal/review/session_test.go`: accept, edit, reject, and submit queue tests.
- Create `internal/tui/model.go`: root Bubble Tea model.
- Create `internal/tui/keys.go`: Vim-like keymap.
- Create `internal/tui/screens.go`: screen and mode routing.
- Create `internal/tui/model_test.go`: key workflow update tests.
- Modify `cmd/trail-hunk/main.go`: wire `internal/app` into the CLI.
- Modify `README.md`: document setup, auth, provider selection, and first review flow.

## Task 1: App Configuration and Entrypoint

**Files:**

- Create: `internal/app/config.go`
- Create: `internal/app/app.go`
- Modify: `cmd/trail-hunk/main.go`
- Test: `internal/app/config_test.go`

- [ ] **Step 1: Write config tests**

```go
package app

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Provider != "codex" {
		t.Fatalf("Provider = %q, want codex", cfg.Provider)
	}
	if cfg.Model != "" {
		t.Fatalf("Model = %q, want empty default", cfg.Model)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("TRAIL_HUNK_PROVIDER", "claude")
	t.Setenv("TRAIL_HUNK_MODEL", "sonnet")

	cfg := ConfigFromEnv()
	if cfg.Provider != "claude" {
		t.Fatalf("Provider = %q, want claude", cfg.Provider)
	}
	if cfg.Model != "sonnet" {
		t.Fatalf("Model = %q, want sonnet", cfg.Model)
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./internal/app`

Expected: fails because `internal/app` does not exist.

- [ ] **Step 3: Add configuration and app runner**

```go
package app

import "os"

type Config struct {
	Provider string
	Model    string
}

func DefaultConfig() Config {
	return Config{Provider: "codex"}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	if provider := os.Getenv("TRAIL_HUNK_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}
	if model := os.Getenv("TRAIL_HUNK_MODEL"); model != "" {
		cfg.Model = model
	}
	return cfg
}
```

```go
package app

import (
	"fmt"
	"io"
)

func Run(stdout io.Writer, cfg Config) error {
	_, err := fmt.Fprintf(stdout, "trail-hunk provider=%s model=%s\n", cfg.Provider, cfg.Model)
	return err
}
```

- [ ] **Step 4: Wire `cmd/trail-hunk/main.go` to `app.Run`**

Keep the Bubble Tea placeholder until Task 7 replaces it. Add config loading in
`main` so the command has a stable startup seam:

```go
cfg := app.ConfigFromEnv()
_ = cfg
```

- [ ] **Step 5: Verify and commit**

Run:

```sh
go test ./...
go vet ./...
hk check
git add cmd internal go.mod go.sum
git commit -m "feat: add app configuration"
```

## Task 2: Git Repository Discovery

**Files:**

- Create: `internal/git/repo.go`
- Create: `internal/git/repo_test.go`

- [ ] **Step 1: Write repo parsing tests**

```go
package git

import "testing"

func TestParseGitHubRemote(t *testing.T) {
	tests := map[string]struct {
		remote string
		owner  string
		repo   string
	}{
		"ssh":   {"git@github.com:joelmccoy/trail-hunk.git", "joelmccoy", "trail-hunk"},
		"https": {"https://github.com/joelmccoy/trail-hunk.git", "joelmccoy", "trail-hunk"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseGitHubRemote(tt.remote)
			if err != nil {
				t.Fatal(err)
			}
			if got.Owner != tt.owner || got.Name != tt.repo {
				t.Fatalf("got %s/%s, want %s/%s", got.Owner, got.Name, tt.owner, tt.repo)
			}
		})
	}
}
```

- [ ] **Step 2: Implement remote parsing and git command helpers**

Create `RepoRef`, `Repository`, `ParseGitHubRemote`, and `Discover`.
`Discover` should run `git rev-parse --show-toplevel`, `git branch
--show-current`, and `git remote get-url origin` with `exec.CommandContext`.

- [ ] **Step 3: Verify and commit**

Run:

```sh
go test ./internal/git ./...
go vet ./...
hk check
git add internal/git
git commit -m "feat: discover git repository context"
```

## Task 3: GitHub Auth and PR Lookup

**Files:**

- Create: `internal/github/auth.go`
- Create: `internal/github/client.go`
- Create: `internal/github/pr.go`
- Create: `internal/github/auth_test.go`

- [ ] **Step 1: Write auth tests**

```go
package github

import "testing"

func TestTokenFromEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "abc123")
	token, err := TokenFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if token != "abc123" {
		t.Fatalf("token = %q, want abc123", token)
	}
}
```

- [ ] **Step 2: Implement auth lookup**

Implement `TokenFromEnv`, `TokenFromGH(ctx)`, and `DiscoverToken(ctx)`.
`DiscoverToken` must prefer `GITHUB_TOKEN`, fall back to `gh auth token`, and
return a setup-oriented error when both fail.

- [ ] **Step 3: Implement GitHub client boundary**

Add:

```go
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Token   string
}
```

Implement `DoJSON(ctx, method, path string, in, out any) error` with
authorization, JSON encoding, JSON decoding, and non-2xx error bodies.

- [ ] **Step 4: Implement PR lookup**

Implement `FindPullRequestsForBranch(ctx, owner, repo, branch string)` using
the GitHub pulls API with `state=open` and `head=owner:branch`.

- [ ] **Step 5: Verify and commit**

Run:

```sh
go test ./internal/github ./...
go vet ./...
hk check
git add internal/github
git commit -m "feat: look up github pull requests"
```

## Task 4: Diff Parser and Review Target Mapping

**Files:**

- Create: `internal/github/diffmap.go`
- Create: `internal/github/diffmap_test.go`

- [ ] **Step 1: Write diff mapping tests**

Use a compact unified diff fixture with one added line, one deleted line, and
one context line. Assert that added lines map to `RIGHT`, deleted lines map to
`LEFT`, and invalid target lookups return an error.

- [ ] **Step 2: Implement parser types**

Add `PullRequestDiff`, `DiffFile`, `DiffHunk`, `DiffLine`,
`DiffLineKind`, and `ReviewTarget`.

- [ ] **Step 3: Implement parsing**

Parse `diff --git`, `+++ b/path`, `@@ -old,count +new,count @@`, and line
prefixes ` `, `+`, and `-`. Assign stable hunk IDs as
`<file-path>:<hunk-index>`.

- [ ] **Step 4: Implement target validation**

Add `FindTarget(path string, side string, line int) (ReviewTarget, error)`.
Reject targets not present in the parsed diff.

- [ ] **Step 5: Verify and commit**

Run:

```sh
go test ./internal/github ./...
go vet ./...
hk check
git add internal/github/diffmap.go internal/github/diffmap_test.go
git commit -m "feat: map github diff review targets"
```

## Task 5: Shell AI Provider Boundary

**Files:**

- Create: `internal/ai/provider.go`
- Create: `internal/ai/shell.go`
- Create: `internal/ai/codex.go`
- Create: `internal/ai/claude.go`
- Create: `internal/ai/schema.go`
- Create: `internal/ai/schema_test.go`

- [ ] **Step 1: Write schema validation tests**

Test that valid JSON produces a `ReviewResponse`, missing overview fails, and
invalid priorities fail.

- [ ] **Step 2: Define provider contracts**

Add provider request/response structs for `Review`, `Ask`, and `Reword`.
Include `Provider`, `ModelRef`, `ReviewRequest`, `ReviewResponse`,
`SuggestedComment`, and `Risk`.

- [ ] **Step 3: Implement shell runner**

Create `Runner` with command path, args, stdin, timeout, stdout, stderr, and
exit-code handling. Default timeout should be five minutes for review calls and
one minute for reword/follow-up calls.

- [ ] **Step 4: Implement Codex and Claude adapters**

Adapters should build a prompt, invoke the configured executable, and parse
strict JSON from stdout. Keep command names configurable with defaults `codex`
and `claude`.

- [ ] **Step 5: Verify and commit**

Run:

```sh
go test ./internal/ai ./...
go vet ./...
hk check
git add internal/ai
git commit -m "feat: add shell ai providers"
```

## Task 6: Review Session and Comment Workflow

**Files:**

- Create: `internal/review/session.go`
- Create: `internal/review/comments.go`
- Create: `internal/review/priority.go`
- Create: `internal/review/session_test.go`

- [ ] **Step 1: Write state transition tests**

Cover accepting a suggested comment, editing an approved comment, dismissing a
suggestion, adding a manual comment, and listing only approved comments for
submission.

- [ ] **Step 2: Implement review state**

Add `ReviewSession`, `WalkthroughPlan`, `ReviewStep`, `ReviewCursor`,
`ReviewComment`, `Priority`, `CommentCategory`, `CommentStatus`, and
`CommentSource`.

- [ ] **Step 3: Implement commands**

Add methods `AcceptSuggestion`, `DismissSuggestion`, `EditComment`,
`AddManualComment`, `ApprovedComments`, `NextStep`, and `PreviousStep`.

- [ ] **Step 4: Verify and commit**

Run:

```sh
go test ./internal/review ./...
go vet ./...
hk check
git add internal/review
git commit -m "feat: manage review session state"
```

## Task 7: Bubble Tea TUI Skeleton

**Files:**

- Create: `internal/tui/model.go`
- Create: `internal/tui/keys.go`
- Create: `internal/tui/screens.go`
- Create: `internal/tui/model_test.go`
- Modify: `cmd/trail-hunk/main.go`

- [ ] **Step 1: Write key handling tests**

Assert `q` quits, `n` advances a review step, `p` moves back, `f` toggles the
file tree flag, and `t` toggles the ask pane flag.

- [ ] **Step 2: Implement root model**

Create a `Model` with screen, focused pane, dimensions, session, and transient
error fields. Implement `Init`, `Update`, and `View`.

- [ ] **Step 3: Implement screen rendering**

Render startup, overview, walkthrough, comment queue, and submit screens as
plain text first. Keep layout simple and deterministic so tests can assert
screen content.

- [ ] **Step 4: Wire main**

Replace the placeholder model in `cmd/trail-hunk/main.go` with
`tui.NewModel(...)` and `tea.NewProgram`.

- [ ] **Step 5: Verify and commit**

Run:

```sh
go test ./internal/tui ./...
go vet ./...
hk check
git add cmd internal/tui
git commit -m "feat: add tui skeleton"
```

## Task 8: GitHub Review Submission

**Files:**

- Create: `internal/github/review.go`
- Create: `internal/github/review_test.go`
- Modify: `internal/review/comments.go`

- [ ] **Step 1: Write payload tests**

Build two approved comments and assert the generated GitHub payload contains
`event: COMMENT`, `path`, `line`, `side`, and `body`.

- [ ] **Step 2: Implement payload builder**

Add a pure function that converts approved review comments into the GitHub
pull request review request body.

- [ ] **Step 3: Implement submit call**

Add `SubmitReview(ctx, owner, repo string, number int, req ReviewRequest)`.
Use `POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews`.

- [ ] **Step 4: Verify and commit**

Run:

```sh
go test ./internal/github ./internal/review ./...
go vet ./...
hk check
git add internal/github internal/review
git commit -m "feat: submit github pull request reviews"
```

## Task 9: End-to-End Orchestration

**Files:**

- Modify: `internal/app/app.go`
- Modify: `internal/app/config.go`
- Modify: `README.md`

- [ ] **Step 1: Write app orchestration test**

Use fake git, GitHub, and AI interfaces to assert that `RunReview` discovers a
PR, fetches context, parses diff, asks AI for a plan, and returns an initialized
review session.

- [ ] **Step 2: Implement orchestration interfaces**

Define small interfaces in `internal/app` for repo discovery, GitHub client,
AI provider, and TUI launcher. Production constructors should wire real
implementations.

- [ ] **Step 3: Implement startup flow**

Run discovery, auth, PR lookup, context fetch, diff parse, provider selection,
AI review generation, session creation, and TUI launch in order. Return clear
errors for missing repo, missing auth, missing PR, and missing AI executable.

- [ ] **Step 4: Document usage**

Update `README.md` with:

```sh
gh auth login
TRAIL_HUNK_PROVIDER=codex go run ./cmd/trail-hunk
TRAIL_HUNK_PROVIDER=claude go run ./cmd/trail-hunk
```

- [ ] **Step 5: Verify and commit**

Run:

```sh
mise run check
git status --short
git add internal README.md cmd
git commit -m "feat: orchestrate guided pr review"
```

## Task 10: Manual MVP Verification

**Files:**

- Modify: `README.md`
- Create: `docs/manual-test.md`

- [ ] **Step 1: Add manual test checklist**

Document a real repository test that covers PR discovery, context fetch, AI
walkthrough generation, accepting one comment, rejecting one comment, adding a
manual comment, and submitting a comment review.

- [ ] **Step 2: Run local checks**

Run:

```sh
mise run check
go run ./cmd/trail-hunk
```

Expected: checks pass and the TUI starts.

- [ ] **Step 3: Commit**

```sh
git add README.md docs/manual-test.md
git commit -m "docs: add manual mvp verification"
```

