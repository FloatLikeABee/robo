# Design

## Context

See proposal.md for why. Code checked before this design:

- A human on local Morph opens `http://localhost:3031/`. Notes & TODOs (`NotesTodosContent`, also the Morph AI drawer) loads `GET /api/tran/notes-todos?type=note` and renders `user_note_todo` rows. The list primary text is the title (`noWrap` in a ~168px column). The editor shows the body. The CRA dev server proxies `/api` to the API on `:9090`.
- `CaseTask` is the shared MorphNotes Tasks board. It has no owner. `big_note` is Big notes (HTML, owner key, publish). `tool_note` is private messages and a public box. None of those is the Notes & TODOs list.
- `CreateUserNoteTodo` inserts `user_note_todo` with the Tran `User.UserID` from `tranUserIDFromContext`. That helper takes `?user_id=` first, then one active `User` by `auth_email`, then numeric `X-User-ID`, then **user id 1**. Two active email matches return 0 and fall through to user id 1. The SPA does not send `user_id`. `AuthzMiddleware` requires a session for `/api/tran/` and strips client identity headers, then sets `X-User-ID` to `plat_users.id` (text, not the Tran integer).
- MCP list/get already scope `user_note_todo` with `ownerClause`: `plat_users.id =` the token subject, joined to one active Tran `User` on email (`Deactivated = 0`, `LIMIT 1`). No match yields an empty list and not-found. They do not call `mcp.ExposeRecord`. That function returns true for any verified user on any record.
- The read pool is a `file:` URI with `mode=ro`, `_query_only=1`, and `_busy_timeout=5000`. It does not set `journal_mode`. `db.NewTranSQL` sets WAL and runs migrations; this process must not call it. A committed insert from another connection is already visible to the read pool.
- `MORPH_MCP_TOKEN` is a Morph session JWT. Startup and every tool call run `auth.DecodeToken`. There is no separate agent principal in the token.

## Goals / Non-Goals

**Goals:**

- `create_note` inserts one note the same human already sees under Notes & TODOs, scoped by `ownerClause`.
- List and get return that note. Read tools stay read-only.
- No token, a bad token, or no Tran user means no row.
- The human can see the note was agent-created without a schema change.

**Non-Goals:**

- Creating todos, editing, or deleting. Case tasks, Big notes, and tool notes.
- Changing `tranUserIDFromContext` or the SPA.
- A new column, a new HTTP route, Badger, or `db.NewTranSQL`.
- Render, skills, commands, and `pkg/morphai`.

## Decisions

### 1. The note is `user_note_todo` with item type `note`

That is the row Notes & TODOs on `:3031` lists for the signed-in user. `create_note` always inserts `ItemType = 'note'`, `Completed = 0`, and `DeadlineAt` NULL. `list_my_tasks` (`type=note`) and `get_task` stay the list and get. The create result is the same task object (id, item type, title, status, body, timestamps when present).

Alternatives:

- `CaseTask` / the MorphNotes Tasks board. Rejected. No owner column, so "user B cannot get user A's note" cannot be implemented, and the existing MCP design already refused this board.
- `big_note`. Rejected. Different screen, HTML body, publish flow. Not what the Notes list shows.
- `tool_note`. Rejected. Inbox and public messages, not Notes & TODOs.
- A new table. Rejected. The human SPA would not show it, and this story does not change the SPA.

Failure mode: a later change points `create_note` at `CaseTask`. The privacy test seeds another user's note and asserts B's get does not contain A's title.

### 2. Write with a second SQLite connection, using `ownerClause`

`create_note` inserts through `INSERT ... SELECT` whose user id is the same scalar subquery `ownerClause` already wraps (`ownerUserSelect`). The read pool stays `query_only`. The writer opens lazily on the first create: `file:` URI, `mode=rw` (the file must already exist; do not use `rwc`), `_busy_timeout=5000`, `MaxOpenConns(1)`. It does not set `journal_mode`, does not migrate, and does not import `idongivaflyinfa/db`. One statement, autocommit. Then the tool reads the row back with the owner predicate.

If the subquery matches nobody, zero rows are inserted and the tool errors. User id 1 is never a default. A request field cannot choose the owner.

Alternatives:

- (a) `POST /api/tran/notes-todos` on `:9090` or via the `:3031` proxy, with the session JWT. Rejected. `tranUserIDFromContext` writes as user id 1 when the email matches no active Tran user, and again when two active users share the email (it returns 0, then falls through). Passing `?user_id=` would hit the query override, which does not check that the id is the session's user. A successful HTTP create that we then fail to read back would already have stored a note on the wrong user. `:3031` only forwards `/api` while the CRA dev server is running. The handler's happy path also returns `{id}` only, so the tool would still read SQLite to return the body.
- Call `db.NewTranSQL` from this process. Rejected. It sets the journal mode and migrates the live file under `morph-api`.
- Import the gin handler and call it in-process. Rejected. Same user-id-1 fallback, plus gin and a writable store inside the stdio process.
- Change `tranUserIDFromContext` so HTTP is safe, then POST. Rejected for this story. The fallback is the demo-install behavior of every Tran route. Fixing it is a separate API change and still leaves MCP unable to create a note when the API process is down, while list/get already use the file.

Failure modes of the chosen path:

- API holds a write lock: the busy timeout elapses, the tool errors, the statement inserts nothing.
- Schema gains a NOT NULL column without a default: the insert errors, no partial row.
- Two active Tran users share the email: `LIMIT 1` picks the same row list/get already pick. The note is visible to those tools. It is not user id 1.
- The read pool is not the access check. A query that omits `ownerClause` is a bug. The insert selects the owner; it does not bind a client integer.

### 3. Attribution is text in the existing columns

Stored title is `[morph-mcp]` when the caller sends no title, otherwise `[morph-mcp] ` plus the trimmed title. An input that already starts with `[morph-mcp]` is not prefixed twice. Stored body is `source: morph-mcp`, then a blank line and the trimmed body when the caller sent one. The Notes list is narrow and uses `noWrap`, so the title prefix is what a human sees without opening the note. The body line remains if they later rename the title in the editor. `create_note` does not add a column.

Alternatives:

- A `source` column. Rejected. The story asks for a marker without a schema change unless one is required. The list and the editor already show title and body.
- Body line only. Rejected. The list would look like any other note until the human opens it.
- Title prefix only. Rejected. Saving a new title from the existing editor drops the only marker. The body line survives that save.

Ceiling: the human can delete both strings. A `source` column the editor does not rewrite is the upgrade if attribution must survive an edit.

### 4. Limits, one duplicate, and quiet errors

At least one of title or body must be non-empty after trim. Title maximum is 200 characters (Unicode code points). Body maximum is 32000 characters, before the source line is added. Over the limit, or both empty: tool error, no insert.

The same stored title and body for the same owner returns the existing id. The check is inside the insert (`NOT EXISTS`), so a sequential retry does not add a row. Two concurrent calls can both insert. ponytail: no unique index, because the SPA allows two notes with the same text; a client idempotency key would need a new column.

Errors say the input was rejected, the note could not be stored, or the session has no notes user. They do not include the token, the secret, another user's title or body, or the words "forbidden". SQLite details go to stderr only. A rejected create is a tool error (`IsError`), not a dropped stdio session, except when the process never starts because the token is missing or invalid.

## Review

Proposer: reuse `POST /api/tran/notes-todos` so validation stays in one place. Reviewer: the handler's owner resolution is not `ownerClause`. No Tran user, or two active users with the same email, stores the note as user id 1. Checking after the POST cannot undo that write. The HTTP option stays rejected.

Proposer: title prefix is enough. Reviewer: the Notes editor saves a new title and keeps the body. A title-only marker disappears on that save, and the list column is too narrow to be the only copy once the title is edited. Keep both the title prefix and the body line.

Proposer: open the writer at startup. Reviewer: a read-only file would then refuse list and get, which work today. Open `mode=rw` on the first create. `mode=rw` must not create a missing file.

Proposer: `list_my_tasks` always shows the new note. Reviewer: the list is capped. Require `get_task` always, and list when the note fits on the page. A new note sorts with `CreatedOn DESC` for `type=note`, so a short page includes it.

Three failure modes that stay tested: invalid token inserts nothing; no Tran user does not write user id 1; user B's get of user A's id is not-found and does not echo A's text.

## Risks / Trade-offs

- [Second writer beside `morph-api`] → WAL plus a 5s busy timeout and a single autocommit insert. The coexistence test holds `db.NewTranSQL` open, creates through MCP, and reads back on the read-only pool. Journal mode stays `wal`.
- [Lazy writer] → List and get still start when the process can read. `create_note` errors if `mode=rw` cannot open the existing file. `mode=rw` does not create a missing file.
- [Duplicate race] → Two identical concurrent creates can store two rows. Sequential retries return one id.
- [Marker is editable] → Document the ceiling. Do not add a column in this change.
- [Email `LIMIT 1`] → Same ponytail as list/get. The insert uses that subquery, so the created row stays on the user those tools read.
- [`ExposeRecord` is not an owner check] → Create, list, and get do not call it. User B's get of A's id is not-found.

## Migration Plan

No schema change. Rollback is removing `create_note` and the writer open. Existing notes stay. The API process is unchanged. Operators set an absolute `TRAN_SQLITE_PATH` in `mcp.json`. The human UI remains `http://localhost:3031/`. The token comes from `POST http://127.0.0.1:9090/api/auth/login` (the `:3031` dev server proxies `/api` to that port). Docs must not contain a real token.

## Open Questions

None. The entity, the write path, the marker, and the limits are decided above.
