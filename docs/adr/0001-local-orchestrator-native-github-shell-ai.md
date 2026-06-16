# ADR 0001: Local Orchestrator with Native GitHub APIs and Shell AI Providers

## Status

Accepted

## Context

`trail-hunk` should run from a local git repository and provide a high-quality
terminal workflow without requiring a hosted service. The user already has local
AI tools available, specifically Codex and Claude Code, and wants GitHub auth to
respect local developer setup.

## Decision

Build `trail-hunk` as a local Go orchestrator:

- Discover repository and PR context locally.
- Use native GitHub API calls from Go.
- Resolve GitHub auth from `GITHUB_TOKEN` or `gh auth token`.
- Treat AI providers as local shell commands behind an interface.

## Consequences

This keeps v1 local-first, avoids provider API key management, and makes GitHub
behavior testable in Go. AI command behavior remains adapter-specific, so the
provider interface must validate structured responses before the rest of the app
trusts them.

