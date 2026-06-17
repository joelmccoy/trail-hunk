# ADR 0010: Render findings as inline diff annotations

## Status

Accepted

## Context

The first hybrid workbench rendered AI finding metadata by appending it to the
target code row. That made code difficult to read and caused important text to
truncate in the middle of the source line. A code review tool must preserve the
integrity of the code view.

The product flow is a sequential walkthrough. Findings are attached to lines in
the current chunk, while the right pane explains the chunk as a whole.

## Decision

Render AI findings as separate annotation rows immediately below their target
diff line. Code rows show only code. Annotation rows show priority, category,
status, comment body, and actions.

Keep the right pane focused on explaining the current walkthrough chunk: what it
does, why it matters, likely impact, review guidance, and confidence.

## Consequences

The diff pane becomes taller because comments consume their own rows, but code
readability is much better. The selected finding can be highlighted without
mutating the code row. The right pane no longer needs to duplicate full finding
text, which leaves room for useful review guidance.
