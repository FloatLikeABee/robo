## Context

See proposal.md for why. Code checked before this design:

- `CaseTask` columns are title, description, dates, location, and assignee fields. There is no creator or owner column. `ListCaseTasks` returns every row (`ORDER BY start_at DESC LIMIT 500`). `GetCaseTask` loads any id. Assignees are `member`, `employee`, or `contact` (`caseTaskAssigneeName`), not `plat_users` and not Tran `User`.
- The Tasks page (`/morphdata/case-tasks`) renders that unfiltered list. `GET /api/tran/*` is still allowed without a session (`docs/agents/03-morph.md`). Copying that list into MCP would show every task to every caller.
- Extra case-task JSON lives in the entity-details Badger store (`attachEntityDetail` on the full/write handlers). `description` is already a SQLite column. Badger takes an exclusive directory lock; this process must not open it.
- `user_note_todo` is keyed by Tran `User.UserID`. `ListUserNotesTodos` / `GetUserNoteTodo` filter `WHERE UserID = ?`, and get returns 404 (not 403) when the id is missing or owned by someone else. The body is a SQLite text column.
- `tranUserIDFromContext` resolves that integer by JWT email (`User.Email`, `Deactivated = 0`, `LIMIT 1`), then falls through to `?user_id=`, `X-User-ID`, and finally user id `1`. `ensureTranUserByAuthEmail` inserts a `User` row when the email is unknown. MCP must not take those write or fallback paths.
- JWT `sub` is `plat_users.id` (text), set in `auth.EncodeToken` from `auth_local.go`. It is not `User.UserID`.
- `auth.LoadTokenConfig` replaces an empty `JWT_SECRET` with `config.DefaultJWTSecret` (`morph-dev-jwt-secret-change-me`). Production startup refuses that value (`config.jwtSecretProblem`). MCP today calls `LoadTokenConfig` and would accept the default.
- `db.NewTranSQL` sets `journal_mode(WAL)` and runs `ensureTranSQLiteSchema`. `morph/cmd/migrate_embedded` already opens the file with `?mode=ro` via `modernc.org/sqlite`. WAL allows a second `mode=ro` reader.
- `startLineReader` in `morph/cmd/morph-mcp/stdio_test.go` sends on an unbuffered channel. A `t.Fatal` or a drain timeout stops receiving, and the goroutine blocks forever on the next send.

## Goals / Non-Goals

**Goals:**

- Two read-only tools that return only the caller's `user_note_todo` rows, beside a live `morph-api`.
- Startup and per-call identity checks that fail closed without logging the token or the secret.
- The stdio line-reader goroutine can leave when the test stops reading.

**Non-Goals:**

- Case-task tools, Badger entity details, resources, prompts, HTTP, write tools, creating Tran users, or the handler's user-id `1` fallback.
- Changing `ListCaseTasks` or the Tasks page. They stay a shared board.

## Decisions

### 1. "Mine" is `user_note_todo`, not `CaseTask`

Resolve the caller the way the notes list does for a real session: JWT email → `User.UserID` where `Deactivated = 0`. Scope every read with that integer. `get_task` uses `ID = ? AND UserID = ?`, so a foreign id and a missing id are the same not-found tool error. Do not mention the other row and do not say forbidden.

If the email matches no active Tran user, list is empty and get is not-found. Do not insert a user and do not fall back to user `1`.

Tool names stay `list_my_tasks` and `get_task` (issue #83's list/get). The description says they are the signed-in user's Notes & TODOs (`item_type` `note` or `todo`), not the shared MorphNotes Tasks board. Optional `type` (`all`|`note`|`todo`, default `all`) and `status` (`all`|`open`|`done`, default `all`, mapped from `Completed`) follow that list. Ordering matches `ListUserNotesTodos` for the chosen type. Default limit 50, applied limit at most 100 (clamp, and report the applied limit). Fields: `id`, `item_type`, `title`, `status`, `body`, and `deadline_at` / `created_on` / `last_updated` when the SQLite text is non-empty. No `user_id` (it is a different namespace from the JWT subject).

`whoami` stays. All three tools set `readOnlyHint`.

Alternatives:

- Return every `CaseTask`, matching the Tasks page. Rejected. There is no owner predicate, so "another user's task → not-found" cannot be implemented, and decision #69 forbids handing the shared board to an MCP client. The page's unauthenticated GET is the bug #69 is closing, not the model to copy.
- Treat "mine" as `CaseTaskAssignee` whose member, employee, or contact email equals the JWT email. Rejected. Those rows are not the plat user. Tasks have no creator, so work the user entered but did not assign to a contact with their email disappears. A shared contact email would return someone else's case. The Tasks page does not filter this way either.
- Proxy `GET /api/tran/case-tasks`. Rejected in the skeleton design, and that route is not per-user.
- Read `big_note` by `user_id`. Rejected. That module is Big notes, the body is large HTML, and the later catalog (#37) owns it.

Failure mode: a later change points these tools at `CaseTask` and the privacy tests go red, because the fixture seeds another user's note and a `CaseTask` and asserts neither foreign title appears.

### 2. Read-only SQLite in this process, never Badger and never `db.NewTranSQL`

`morph/mcp` opens `TRAN_SQLITE_PATH` with `modernc.org/sqlite` as a `file:` URI: `mode=ro`, `_query_only=1`, `_busy_timeout=5000`. A bare `path?mode=ro` is not a URI to this driver; it creates the file. `_query_only` is the driver's checked pragma shorthand (applied last). The opener does not set `journal_mode` and does not run migrations. The import test still forbids `idongivaflyinfa/db` and Badger in non-test files. Tests seed with `db.NewTranSQL` (the real `ensureTranSQLiteSchema`), leave that writer open, then open the read-only connection and show that an insert on it fails while the writer can still commit.

Missing file, missing `plat_users`, or a failed open fails startup. Desktop configs should set an absolute `TRAN_SQLITE_PATH`; the API default `./data/tran.sqlite` applies only when the variable is unset, and a missing file still fails.

`user_note_todo.Body` is the text body. Nothing is read from Badger, so the tools do not add a "detail lives in Badger" caveat. Case-task detail JSON stays in Badger and stays out of this response.

Alternatives:

- Call `db.NewTranSQL` from the stdio process. Rejected. It writes the journal mode and migrates the live file out from under `morph-api`.
- Open Badger read-only for case-task detail. Rejected. Shared flock still fails against the API's exclusive lock (skeleton design).

### 3. Identity hardening around the existing JWT

Startup order: refuse an empty or development-default `JWT_SECRET` by reading the environment before `LoadTokenConfig` can substitute the default (case-insensitive, trimmed, error does not echo the secret); verify `MORPH_MCP_TOKEN`; open the read-only DB; `SELECT` `plat_users.id = sub`. Secret is first so an unset secret cannot fall through to the development key. The existing missing-token test must set a non-default secret; otherwise it would now fail on `JWT_SECRET` and stop proving the token path. `get_task` takes a JSON integer id; a non-integer or non-positive id is a tool error `invalid id`, which does not reveal whether a row exists. An unknown `type` or `status` is a tool error, not a silent "all". Each tool handler calls `DecodeToken` again and returns a tool error (`IsError`) when that fails. The process stays up so the client sees a tool error rather than a dead stdio pipe. Startup still refuses an already-expired token before any protocol byte.

Other production secret rules (length, placeholder fragments, repeated characters) stay on the API. MCP tests and local non-default secrets such as `stdio-test-secret` must keep working. The only built-in value `LoadTokenConfig` substitutes is `config.DefaultJWTSecret`.

Alternatives:

- Trust claims for the whole process lifetime, as the skeleton does. Rejected. Review follow-up from #65, and a deleted user would keep reading notes.
- Re-query `plat_users` on every call. Not required. Startup is the existence check; expiry is the per-call check. A user deleted after launch works until the process restarts. Document that, same operational note as secret rotation.

### 4. Line reader exit

`startLineReader` takes a `context.Context`. Sends select on `ctx.Done()` and return, including the scan-error send. Callers pass `t.Context()`, which is canceled when the test finishes, so a `t.Fatal` unblocks the goroutine. An unbuffered channel stays; a buffer only postpones the stuck send.

### 5. Bundled OpenSpec sentences

`bk/openspec/changes/remove-tools-merge-adviser-customization/proposal.md` must not say BK still serves MCP `tools/list` / `tools/call`. Those routes left in #40 / PR #73. `bk/openspec/changes/archive/2026-09-24-remove-bk-tcp-mcp/tasks.md` must record the actual check: pytest 9 passed and 1 pre-existing Gemini-key failure, and `bk/frontend` `npm run build` failed on a missing module already on main.

## Risks / Trade-offs

- [Notes & TODOs are not the MorphNotes Tasks board] → Tool description says so. Case tasks wait until they have a real owner column. Shipping the board now violates #69.
- [Email `LIMIT 1` if two Tran users share an email] → Same predicate as the notes handler. ponytail: first active row wins; a unique email index would be the upgrade.
- [Deleted user after startup still reads until restart] → Expiry is re-checked; existence is startup-only. Restart the process to drop a deleted account.
- [A huge note body makes a large tool result] → Return the SQLite text as stored. ponytail: no size cap; if a body becomes a problem, truncate and set a flag.
- [Read-only URI rejected when no WAL shm exists yet] → The API opens WAL before MCP is pointed at the file. The coexistence test holds the writer open.
- [`.env` still contains the development JWT secret] → MCP refuses it. The client config must set a non-default `JWT_SECRET` that matches the API.

## Migration Plan

No schema change. Rollback is removing the two tools. The API process is unchanged. Operators add `TRAN_SQLITE_PATH` to `mcp.json`.

## Open Questions

None. Ownership, the read-only DSN, and the secret rule are decided above.
