## Context

See proposal.md for motivation. Spec: `generic-data-list-complete`.

`GET /api/tran/generic-data` loops `rows.Next()` then calls `querySingleRowMapFromRows`, which calls `Next()` again before Scan. That skips every other SQL row. With an even number of rows the newest (first in `ORDER BY last_updated DESC, id DESC`) is skipped, so Save looks like it failed and an older row stays visible. The detail drawer can still open the real saved id from the create response.

## Goals / Non-Goals

**Goals:**
- List JSON length equals stored row count (within the existing 500 cap).
- Newest insert/update is in the payload and therefore in the DataGrid after reload.

**Non-Goals:**
- Cursor pagination or raising LIMIT 500.
- Changing import extract/save (already review-then-save).

## Decisions

### 1. Scan the current row; do not Next twice
- **Choice:** Keep `for rows.Next()` in `ListGenericData`. Change the helper to Scan the **current** row only (no inner `Next`). Same pattern as other Morph list handlers.
- **Rationale:** `querySingleRowMap` still needs `Next` for a one-shot query. The list helper was copied from that and used inside an existing Next loop.
- **Alternatives:** Remove the loop Next and only Next inside the helper — easy to break `rows.Err()`. Scanning in-place is the usual `database/sql` list loop.

### 2. No frontend workaround as the fix
- **Choice:** Do not prepend the create response onto stale/partial list state as the real fix. After the API returns all ids, existing `load()` after Save is enough. Optionally keep merging the saved record if load fails — not required.
- **Rationale:** A client merge would still hide other skipped historical rows.

## Risks / Trade-offs

- [Grid shows ~2× as many rows as today] → Expected; operators who thought they had three records may have six.
- [LIMIT 500 still truncates huge tables] → Unchanged; only mention if a follow-up is needed.

## Migration Plan

1. Ship the list scan fix; reload Generic data.
2. Rollback: revert the helper (would restore skipping).

## Open Questions

None.
