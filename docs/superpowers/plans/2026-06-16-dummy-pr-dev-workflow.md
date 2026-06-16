# Dummy PR Dev Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a repo-local development command that creates an isolated dummy GitHub draft PR for testing trail-hunk workflows.

**Architecture:** A checked-in shell script owns git worktree, branch, dummy fixture commit, push, and draft PR creation. Mise tasks call the script for setup, dry-run validation, and reset. The workflow uses the developer's existing `gh` authentication and never mutates the current `main` checkout.

**Tech Stack:** Go project, mise tasks, POSIX-ish Bash, GitHub CLI.

---

### Task 1: Script Contract

**Files:**
- Create: `scripts/dev-dummy-pr.sh`
- Create: `scripts/test-dev-dummy-pr.sh`

- [ ] Add `scripts/test-dev-dummy-pr.sh` with assertions for `scripts/dev-dummy-pr.sh --dry-run`.
- [ ] Run `bash scripts/test-dev-dummy-pr.sh` and confirm it fails because the script does not exist yet.
- [ ] Add `scripts/dev-dummy-pr.sh` with `setup`, `reset`, and `--dry-run` support.
- [ ] Run `bash scripts/test-dev-dummy-pr.sh` and confirm it passes.

### Task 2: Mise Tasks and Docs

**Files:**
- Modify: `mise.toml`
- Modify: `README.md`
- Add: `docs/adr/0007-real-draft-pr-dev-workflow.md`

- [ ] Add `dev:dummy-pr`, `dev:dummy-pr:reset`, and `dev:dummy-pr:dry-run` mise tasks.
- [ ] Document the workflow in README.
- [ ] Record the decision to use real draft PR worktrees in an ADR.
- [ ] Run `mise run check`.
