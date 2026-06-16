# ADR 0005: Startup Detects PR Context Without AI

## Status

Accepted

## Context

The startup screen should tell the user whether `trail-hunk` found a GitHub pull
request for the current branch. Full guided review generation is heavier because
it fetches the raw diff and invokes an AI provider.

## Decision

Run lightweight startup detection asynchronously when the TUI starts. This
detects the local repository and matching open GitHub PR, then displays PR
metadata on the startup screen. Keep AI review generation user-triggered with
`R`.

## Consequences

The user gets immediate context without spending an AI call. Startup detection
still requires GitHub auth, so the startup screen needs a visible error state
when auth or PR lookup fails. The heavier review generation path remains
separate and explicit.

