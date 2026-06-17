#!/usr/bin/env bash
set -euo pipefail

branch="trail-hunk-dev/dummy-pr"
base_branch="main"
worktree_name="dummy-pr"
mode="open"
dry_run=0

usage() {
  cat <<'USAGE'
Usage: scripts/dev-dummy-pr.sh [open|setup|reset] [--dry-run]

Creates or refreshes a real draft GitHub pull request from an isolated worktree
and runs trail-hunk with deterministic fixture AI data.

Commands:
  open       Create/update the dummy PR worktree and launch trail-hunk. This is the default.
  setup      Create/update the dummy PR worktree, branch, commit, push, and draft PR.
  reset      Remove the dummy worktree and recreate the dummy PR branch from origin/main.

Options:
  --dry-run  Print the planned workflow without mutating git or GitHub.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    open|setup|reset)
      mode="$1"
      ;;
    --dry-run)
      dry_run=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n\n' "$1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

repo_root="$(git rev-parse --show-toplevel)"
worktree="$repo_root/.dev/worktrees/$worktree_name"
fixture_dir="$worktree/dev/fixtures/dummy-pr"
fixture_file="$fixture_dir/review_target.go"
remote_url="$(git -C "$repo_root" config --get remote.origin.url || true)"
source_head="$(git -C "$repo_root" rev-parse --short HEAD)"
source_ref="$(git -C "$repo_root" rev-parse HEAD)"
source_branch="$(git -C "$repo_root" branch --show-current)"
if [[ -z "$source_branch" ]]; then
  source_branch="detached-$source_head"
fi

run() {
  if [[ "$dry_run" -eq 1 ]]; then
    printf '+ %q' "$1"
    shift
    for arg in "$@"; do
      printf ' %q' "$arg"
    done
    printf '\n'
    return 0
  fi
  "$@"
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

print_summary() {
  cat <<SUMMARY
trail-hunk dummy PR workflow
mode: $mode
branch: $branch
base: $base_branch
worktree: $worktree
remote: ${remote_url:-unknown}
provider: fixture
source: $source_branch@$source_head
snapshot: current checkout plus dummy fixture

next:
  cd $worktree
  TRAIL_HUNK_PROVIDER=fixture go run ./cmd/trail-hunk
SUMMARY
}

ensure_clean_repo() {
  if [[ "$dry_run" -eq 1 ]]; then
    return 0
  fi
  if [[ "$mode" != "reset" ]]; then
    return 0
  fi

  if [[ -n "$(git -C "$repo_root" status --porcelain)" ]]; then
    cat >&2 <<'ERROR'
The main checkout has uncommitted changes.
Commit or stash them before creating the dummy PR worktree.
ERROR
    exit 1
  fi
}

remove_worktree() {
  if git -C "$repo_root" worktree list --porcelain | grep -Fqx "worktree $worktree"; then
    run git -C "$repo_root" worktree remove --force "$worktree"
  elif [[ -d "$worktree" ]]; then
    run rm -rf "$worktree"
  fi
}

delete_local_branch() {
  if git -C "$repo_root" show-ref --verify --quiet "refs/heads/$branch"; then
    run git -C "$repo_root" branch -D "$branch"
  fi
}

create_worktree() {
  run git -C "$repo_root" fetch origin "$base_branch"
  run git -C "$repo_root" worktree prune
  remove_worktree
  delete_local_branch

  run git -C "$repo_root" worktree add -B "$branch" "$worktree" "$source_ref"
}

apply_tracked_diff() {
  if git -C "$repo_root" diff --quiet HEAD --; then
    return 0
  fi

  if [[ "$dry_run" -eq 1 ]]; then
    run git -C "$repo_root" diff --binary HEAD
    run git -C "$worktree" apply --binary
    return 0
  fi

  git -C "$repo_root" diff --binary HEAD | git -C "$worktree" apply --binary
}

copy_untracked_files() {
  if [[ "$dry_run" -eq 1 ]]; then
    run git -C "$repo_root" ls-files --others --exclude-standard
    run cp "<untracked non-ignored files>" "$worktree"
    return 0
  fi

  while IFS= read -r -d '' path; do
    mkdir -p "$worktree/$(dirname "$path")"
    cp -R "$repo_root/$path" "$worktree/$path"
  done < <(git -C "$repo_root" ls-files --others --exclude-standard -z)
}

seed_fixture() {
  run mkdir -p "$fixture_dir"
  if [[ "$dry_run" -eq 1 ]]; then
    run tee "$fixture_file"
    return 0
  fi

  cat >"$fixture_file" <<'GO'
package dummypr

import "strings"

type Account struct {
	ID       string
	Role     string
	IsActive bool
}

// CanAccessBilling is intentionally imperfect fixture code for trail-hunk reviews.
func CanAccessBilling(account Account, requestedAccountID string) bool {
	if strings.TrimSpace(requestedAccountID) == "" {
		return true
	}

	if account.Role == "admin" {
		return true
	}

	return account.IsActive && account.ID == requestedAccountID
}

func NormalizeDisplayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) > 24 {
		return trimmed[:24]
	}
	return trimmed
}
GO
}

commit_snapshot() {
  run git -C "$worktree" add -A
  if [[ "$dry_run" -eq 1 ]]; then
    run git -C "$worktree" commit -m "dev: snapshot dummy review workflow"
    return 0
  fi

  if git -C "$worktree" diff --cached --quiet; then
    printf 'no changes to snapshot\n'
    return 0
  fi
  run git -C "$worktree" commit -m "dev: snapshot dummy review workflow"
}

push_and_open_pr() {
  run git -C "$worktree" push --force-with-lease -u origin "$branch"

  if [[ "$dry_run" -eq 1 ]]; then
    run gh pr view "$branch"
    run gh pr create --draft --base "$base_branch" --head "$branch" --title "Dummy PR for trail-hunk workflow testing"
    return 0
  fi

  local repo
  repo="$(gh_repo_slug)"
  if gh pr view "$branch" --repo "$repo" >/dev/null 2>&1; then
    printf 'reusing existing draft PR for %s\n' "$branch"
    gh pr view "$branch" --repo "$repo" --json url --jq .url
    return 0
  fi

  gh pr create \
    --draft \
    --base "$base_branch" \
    --head "$branch" \
    --title "Dummy PR for trail-hunk workflow testing" \
    --body "$(pr_body)"
}

launch_tui() {
  if [[ "$mode" != "open" ]]; then
    return 0
  fi

  if [[ "$dry_run" -eq 1 ]]; then
    run bash -lc "cd '$worktree' && TRAIL_HUNK_PROVIDER=fixture go run ./cmd/trail-hunk"
    return 0
  fi

  cd "$worktree"
  export TRAIL_HUNK_PROVIDER=fixture
  exec go run ./cmd/trail-hunk
}

gh_repo_slug() {
  local slug
  slug="$(gh repo view --json nameWithOwner --jq .nameWithOwner)"
  printf '%s' "$slug"
}

pr_body() {
  cat <<'BODY'
This draft PR is a local development fixture for trail-hunk.

It intentionally contains reviewable issues so the TUI can exercise:

- PR startup detection
- diff fetching
- AI review overview generation
- hunk walkthrough
- suggested review comment approval/rejection
- final pending review submission

Close or overwrite this PR whenever you want; `mise run dev:dummy-pr:reset` recreates the branch.
BODY
}

main() {
  require_command git
  require_command gh
  ensure_clean_repo

  print_summary
  printf '\n'

  if [[ "$dry_run" -eq 1 ]]; then
    printf 'planned actions:\n'
  fi

  if [[ "$mode" == "reset" ]]; then
    remove_worktree
    delete_local_branch
  fi

  create_worktree
  apply_tracked_diff
  copy_untracked_files
  seed_fixture
  commit_snapshot
  push_and_open_pr

  printf '\nready:\n'
  printf '  cd %s\n' "$worktree"
  printf '  TRAIL_HUNK_PROVIDER=fixture go run ./cmd/trail-hunk\n'

  launch_tui
}

main
