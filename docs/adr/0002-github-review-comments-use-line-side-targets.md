# ADR 0002: GitHub Review Comments Use Line and Side Targets

## Status

Accepted

## Context

GitHub pull request review comments can only target lines that exist in the PR
diff. AI suggestions may describe useful review comments but cannot be trusted
to produce valid GitHub line targets.

## Decision

Parse the unified PR diff into files, hunks, and diff lines. Use modern GitHub
review target fields:

- `line`
- `side`
- `start_line`
- `start_side`

Avoid legacy `position` for new review comments.

## Consequences

The app validates AI suggestions against a local diff map before comments enter
the review queue. Invalid suggestions are skipped or converted to non-submittable
notes later. This reduces failed review submissions and makes target validation
testable.

