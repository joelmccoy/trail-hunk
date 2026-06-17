# ADR 0015: Code-stream diff and UX journey fixtures

## Status

Accepted

## Context

The walkthrough screen still looked like a generated report even after the first diff-first pass. The diff used table-like visual structure, the assistant context was split into competing columns, and the local fixture was too small to reveal realistic navigation problems.

Trail Hunk needs to feel like a code review workbench. The reviewer should read code first, see AI suggestions as line-level affordances, and use the rail to understand review order and changed files.

## Decision

The walkthrough diff now renders as a single code stream:

- One narrow gutter carries the AI marker and line number.
- The diff no longer exposes old/new/code column labels.
- The current step context is a compact stacked header, not a split report panel.
- The rail deduplicates groups by title and avoids repeating the active file under the selected step.
- A multi-step, multi-file fixture journey test renders snapshots for realistic review paths.

## Consequences

The main walkthrough is easier to scan because code is the dominant object and AI suggestions live in one selected drawer. The fixture journey gives future UI work a fast feedback loop for crowding, duplicate navigation labels, and line-width regressions before testing against a real GitHub PR.
