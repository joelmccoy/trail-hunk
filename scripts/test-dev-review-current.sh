#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script="$repo_root/scripts/dev-review-current.sh"

output="$("$script" --dry-run)"

assert_contains() {
  local needle="$1"
  if [[ "$output" != *"$needle"* ]]; then
    printf 'expected dry-run output to contain: %s\n\nactual output:\n%s\n' "$needle" "$output" >&2
    exit 1
  fi
}

assert_contains "trail-hunk review-current workflow"
assert_contains "mode: open"
assert_contains "branch: trail-hunk-dev/review-current"
assert_contains "worktree: $repo_root/.dev/worktrees/review-current"
assert_contains "base: main"
assert_contains "snapshot: tracked diff plus untracked non-ignored files"
assert_contains "ensure reviewable diff"
assert_contains "next:"
assert_contains "cd $repo_root/.dev/worktrees/review-current"
assert_contains "go run ./cmd/trail-hunk"
