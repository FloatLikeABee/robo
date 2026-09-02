## Context

See proposal.md for motivation. Spec: `case-task-unassigned-create`.

Morph Data Tasks save (`POST /api/tran/case-tasks`) builds a dynamic `INSERT INTO CaseTask` from allowed columns (`title`, `description`, `start_at`, `end_at`, `location`). Assignees live in `CaseTaskAssignee`; the UI omits them on write. Legacy columns `assignee_type` and `assignee_id` remain on `CaseTask`.

SQLite `CREATE TABLE IF NOT EXISTS` does not upgrade existing files. Older Morph Data DBs have `assignee_type TEXT NOT NULL` **without** a DEFAULT. Omitting the column on INSERT stores NULL and SQLite returns `constraint failed: NOT NULL constraint failed: CaseTask.assignee_type (1299)`. Current `sqlite_schema.go` uses `NOT NULL DEFAULT ''`, which only applies to newly created tables. MySQL migration 035 already made `assignee_type` nullable.

## Goals / Non-Goals

**Goals:**
- Unassigned create works on existing SQLite files (NOT NULL, no DEFAULT) and on new schema.
- Insert always supplies non-NULL legacy assignee columns (`''` / `0`) when no assignee is chosen; `replaceCaseTaskAssignees` still copies the first assignee into those columns when present.
- A regression test covers title-only create against a NOT NULL `assignee_type` table.

**Non-Goals:**
- Rebuilding every existing `CaseTask` table via copy/rename unless INSERT defaults are insufficient.
- Bringing back assignee editing in the Tasks drawer.
- Changing MySQL ENUM values.

## Decisions

### 1. Always bind legacy columns on INSERT
- **Choice:** `CreateCaseTask` always includes `assignee_type` and `assignee_id` in the INSERT. Empty assignees → `''` and `0`. If assignees were parsed, use the first key’s kind/id (same as `replaceCaseTaskAssignees` today).
- **Rationale:** SQLite DEFAULT is skipped when the column is omitted **and** older tables have no DEFAULT. Explicit values fix both old and new files without a table rebuild.
- **Alternatives:** Only add `DEFAULT ''` in `CREATE TABLE` — does not migrate existing files. Make the column nullable via table rebuild — heavier, easy to get wrong with indexes.

### 2. Keep schema DEFAULT for new files; no required ALTER
- **Choice:** Leave `sqlite_schema.go` `DEFAULT ''` as-is. Do not require an ALTER of NOT NULL (SQLite cannot change column nullability cheaply). Optional: document that INSERT is the compatibility layer.
- **Rationale:** One code path in the handler is enough for the reported bug.
- **Alternatives:** Recreate `CaseTask` to match MySQL NULL — only if INSERT-alone is not enough in tests.

### 3. Test against a NOT NULL, no-DEFAULT table
- **Choice:** Unit/integration test opens a temp SQLite `CaseTask` with `assignee_type TEXT NOT NULL` (no DEFAULT) and posts create with title only; expect 200 and a row.
- **Rationale:** Matches production files created before DEFAULT was added.

## Risks / Trade-offs

- [Old rows with empty string vs NULL] → Reads already treat empty/NULL as “no legacy assignee”; `''` + `0` matches that.
- [Someone sends `assignee_type: null` in JSON] → Continue deleting client keys then set server-side defaults so NULL never reaches SQLite.

## Migration Plan

1. Ship handler INSERT defaults; no operator DB step.
2. Rollback: revert handler; constraint error returns on unassigned create.

## Open Questions

None.
