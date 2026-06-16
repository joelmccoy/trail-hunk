#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script="$repo_root/scripts/dev-dummy-pr.sh"

output="$("$script" --dry-run)"

assert_contains() {
  local needle="$1"
  if [[ "$output" != *"$needle"* ]]; then
    printf 'expected dry-run output to contain: %s\n\nactual output:\n%s\n' "$needle" "$output" >&2
    exit 1
  fi
}

assert_contains "trail-hunk dummy PR workflow"
assert_contains "mode: setup"
assert_contains "branch: trail-hunk-dev/dummy-pr"
assert_contains "worktree: $repo_root/.dev/worktrees/dummy-pr"
assert_contains "base: main"
assert_contains "next:"
assert_contains "cd $repo_root/.dev/worktrees/dummy-pr"
assert_contains "go run ./cmd/trail-hunk"
