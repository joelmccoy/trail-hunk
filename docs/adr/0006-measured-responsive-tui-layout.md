# ADR 0006: Use measured responsive TUI regions

## Status

Accepted

## Context

The terminal UI needs to feel like a full-screen application across different terminal sizes. Earlier layout code mixed raw terminal width, manual padding subtraction, and fixed panel caps. That made footer and panel rows wrap or leave awkward unused space at wider sizes.

Bubble Tea reports terminal size through `tea.WindowSizeMsg`, and Charm examples size full-screen views by reserving header and footer height, then assigning the remaining height to the body. Lip Gloss style widths are rendered block widths before margins, and borders/padding must be accounted for when composing child panels.

## Decision

The root TUI model owns terminal dimensions and renders three measured regions:

- a one-line header sized to the terminal width
- a body sized to the remaining terminal height
- a one-line footer sized to the terminal width

Child screens receive the measured body content width instead of reading raw terminal width. Panel helpers accept a desired rendered panel width and subtract border width internally before applying Lip Gloss `Width`.

The footer uses adaptive key labels that collapse from long to compact to minimal forms based on measured width. Footer text is truncated explicitly when needed so it never wraps.

## Consequences

Responsive behavior is centralized in the root renderer, which makes screen-specific views simpler and easier to test. Tests now verify that startup views render full-width rows across narrow, normal, and wide terminal sizes, and that the footer remains a single full-width row.

Future panes should accept a measured content width from the root layout instead of using `Model.Width` directly.
