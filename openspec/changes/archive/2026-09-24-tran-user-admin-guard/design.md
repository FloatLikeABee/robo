## Context

See proposal.md for why this is a deploy blocker. The code on `main` matches the report.

`AuthzMiddleware` already requires a Morph session for `/api/tran/*` and returns 401 when `ResolveUserScope` fails (`morph/handlers/authz_middleware.go`). `ResolveUserScope` (lines 38–68) loads `plat_users` with `GetPlatUserByID` and sets `IsAdmin` from `PlatUser.IsAdmin()` (`morph/db/plat_users.go` lines 29–35, role `Admin`). `applyAuthScope` stores that boolean as `auth_is_admin`. `requireAdmin` (lines 216–225) reads only that boolean and returns 403 otherwise. `/api/admin/users` calls `requireAuthDB` then `requireAdmin` (`morph/handlers/admin_users.go` line 26). `requireAuthDB` only checks that the SQL store exists.

`CreateTranUser`, `UpdateTranUser`, and `DeleteTranUser` do not call `requireAdmin`. `allowedTranUserWrite` includes `email`, `administrator`, and `deactivated`. `allowedTranUserSelfWrite` includes `email`. `tranUserByEmail` uses `LIMIT 1` with no order, then the same pattern on `LoginID`. `tranUserIDFromContext` (`morph/handlers/tran_tools.go`) repeats the email `LIMIT 1` and, on a miss, falls through to `X-User-ID` and then user id 1. `ensureTranUserByAuthEmail` inserts a row only when the lookup returns `sql.ErrNoRows`.

The profile form (`morph/frontend/src/pages/admin/UserSettings.js`) is the only caller of `PUT /api/tran/users/me`. Nothing in the Morph SPA calls `POST`, `PUT /:id`, or `DELETE` on `/api/tran/users`. SQLite's `User` table has no `DeactivatedDate` column, but delete and deactivate already write that column (MySQL has it).

## Goals / Non-Goals

**Goals:**

- Reuse `requireAdmin` for the three administration routes.
- Stop self-service email, role, and deactivated edits.
- Fail closed on more than one active email or LoginID match.
- Keep one-match and zero-match profile behavior.
- Keep the diff off PR #87's owner-scoping work except the email branch inside `tranUserIDFromContext`.

**Non-Goals:**

- Linking Tran users to `plat_users.id`.
- Changing GET user routes, JWT shape, login, or environment variables.
- The empty-id `admin` fallbacks in forms, hybrid context, registration, and chat upload (issue #88). Do not copy `forms.go` lines 81–84 (`userID == ""` becomes `"admin"`).
- Replacing `?user_id=` or the user-id-1 miss fallback in `tranUserIDFromContext`. That is PR #87's area.
- A new admin helper or a second plat-user query inside the Tran handlers.

## Decisions

### 1. Handler gate, same helper as `/api/admin/users`

Call `requireAuthDB` then `requireAdmin` at the start of `CreateTranUser`, `UpdateTranUser`, and `DeleteTranUser`.

Admin means `auth_is_admin`, which middleware set from the current `plat_users` row. A missing or false flag is 403, never admin. 401 stays in the middleware: no token, invalid token, or unknown user (including an empty subject, because `GetPlatUserByID` misses and `ResolveUserScope` returns false). Do not add an empty-id branch in these handlers.

**Rejected:** special-case these paths inside `AuthzMiddleware` the way `/api/admin/` is gated. `GET /api/tran/users`, `GET /api/tran/users/:id`, and `/users/me` must stay open to every signed-in user. A path rule that also has to spare `me` is easier to get wrong than the three handler lines `/api/admin/users` already uses.

**Rejected:** a new helper that re-reads `plat_users` inside the handler. That duplicates `ResolveUserScope`. Trusting the JWT `roles` claim without the re-read is also rejected: `ResolveUserScope` already ignores the claim and loads the row.

**Rejected:** treating Tran `User.Administrator` as the platform admin flag. That column is one of the fields this bug can write.

### 2. `/users/me` drops `email` from the allowlist

Remove `email` from `allowedTranUserSelfWrite`. Unknown keys are already skipped. A body that still has `last_name` or `phone` saves those fields and leaves email alone. A body whose only recognized field was `email` hits the existing `no fields to update` 400 and writes nothing. `administrator` and `deactivated` are not in the map, so they stay ignored.

**Rejected:** return 400 whenever `email` is present. That is a new special case, and the profile form today always sends email. Ignore-plus-400-when-nothing-else-remains is the behavior already used for every other unknown key.

**Rejected:** reject only when the new email belongs to someone else. The account email is the login email. A user should not point the Tran row at a different address at all.

### 3. Ambiguous match is an error, HTTP 409 on the profile routes

Replace both `LIMIT 1` lookups in `tranUserByEmail` with a shared column match that reads active rows and stops at the second hit.

- One row: return it.
- Zero rows on `Email`: try `LoginID` the same way. One LoginID row returns that row. Zero LoginID rows returns `sql.ErrNoRows`, and `ensureTranUserByAuthEmail` still inserts. That preserves auto-create.
- Two or more rows on the stage that is actually used: return a sentinel error. `ensureTranUserByAuthEmail` must not insert, because it already inserts only on `sql.ErrNoRows`. `GET` and `PUT /api/tran/users/me` map the sentinel to 409.

Email is tried first. A single email hit does not also consult LoginID, which is what the code does today. Two email hits do not fall through to LoginID.

`tranUserIDFromContext` calls the email-column match only. It does not gain the LoginID fallback. On the sentinel it returns 0 and does not continue to `X-User-ID` or user id 1. On zero email rows it falls through exactly as it does now. Callers keep the current `int` signature so this does not rewrite Notes handlers.

**Rejected:** `ORDER BY UserID LIMIT 1`. That is deterministic and still binds the session to one of the two people. The takeover only needs the victim to be the chosen row.

**Rejected:** 403 for the ambiguous profile. 403 is the non-admin response. 409 says the identity is in conflict and the server will not guess. Both were allowed; 409 is the one that does not look like a permissions bug.

**Rejected:** `AbortWithStatusJSON(409)` inside `tranUserIDFromContext`, or a new error return on every Notes handler. The current handler keeps running after `Abort` and writes a second body. Changing every caller overlaps PR #87. Returning 0 means Notes queries match neither row and create nothing. Profile routes, which are the resolution API, return 409.

### 4. SQLite gets the column the deactivate SQL already writes

Add nullable `DeactivatedDate` to the SQLite `User` table and to the existing `sqliteAddColumnIfMissing` startup pass. `DELETE /api/tran/users/:id` and a deactivate `PUT` already set that column. Without it, an admin on the supported SQLite stack gets 500, so "admin succeeds" is false. MySQL already has the column. No backfill.

### 5. Profile email control is read-only

`UserSettings.js` shows email as read-only and stops putting `email` in the save body. No other frontend change. The user-management grid does not call these write routes.

## Risks / Trade-offs

- [Notes list stays HTTP 200 when the email is ambiguous, with user id 0] → No victim row is selected and no row is created. Profile GET/PUT return 409. A later owner-scoping change can surface 409 without this PR editing those handlers.
- [`?user_id=` on Notes still selects an id before the email lookup] → Left in place so this diff does not fight PR #87. The email takeover path is closed.
- [A miss in `tranUserIDFromContext` can still become user id 1] → Unchanged on purpose. Ambiguous matches must not take that path; a clean miss still does.
- [Existing duplicate active emails start failing `/users/me` with 409] → That is the fail-closed behavior. An admin deactivates or rewrites the extra row through the now-protected routes. Rollback is reverting the commit. The new column is nullable and unused by older binaries.
- [Handler tests that skip the middleware see 403 from `requireAdmin`, not 401] → Tests mount `AuthzMiddleware` the way production does. Do not teach `requireAdmin` to return 401; `/api/admin/users` shares it.
- [An old profile client still sends `email`] → The server ignores it. The updated form does not send it.

## Migration Plan

Ship the handler change and the nullable SQLite column together. No JWT or env change. Duplicate active emails are not auto-merged and are not auto-deleted. To roll back, revert the commit; leftover `DeactivatedDate` values can stay.

## Open Questions

None. The HTTP status for an ambiguous profile (409), the ignore rule for `email` on `/users/me`, and the notes resolver returning no id are fixed above.

## Review

Proposer and reviewer passed the same design. The challenges that had to be answered before tasks:

- Returning 0 from the notes resolver is not an HTTP 409. Aborting inside that helper double-writes, and a new return value edits every Notes handler that PR #87 is likely to touch. 409 stays on `/users/me`. User id 0 selects neither row.
- `?user_id=` still runs before the email match. Tests for ambiguity must not send that query. Removing the parameter is out of scope.
- A clean email miss may still become user id 1. Only the multi-match branch is forbidden from taking that fallthrough.
- `requireAdmin` returns 403 when the middleware never ran. The 401 tests have to go through `AuthzMiddleware`, which is how `/api/tran/*` is mounted.
- SQLite without `DeactivatedDate` makes admin delete return 500. Adding the nullable column is part of "admin succeeds" on this repo's database, not a new deactivate feature.
