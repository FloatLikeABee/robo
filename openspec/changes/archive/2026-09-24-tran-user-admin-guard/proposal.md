## Why

Any signed-in user can create, rewrite, or deactivate Tran `User` rows, including email and administrator. Notes and TODOs resolve "me" by matching the login email to an active Tran user with an unordered `LIMIT 1`, so rewriting a victim's email and deactivating the attacker's own row takes over the victim's notes. This is a v1 deploy blocker (issues #90 and #89).

## What Changes

- **BREAKING** for non-admins: `POST /api/tran/users`, `PUT /api/tran/users/:id`, and `DELETE /api/tran/users/:id` require a platform admin. No session returns 401. A signed-in non-admin returns 403. An admin succeeds.
- Admin is the platform role on the session's `plat_users` row, the same check `/api/admin/users` uses. It is not a client header, the request body, a JWT role claim by itself, or the Tran `User.Administrator` column. An empty user id is 401, not an admin.
- `PUT /api/tran/users/me` no longer writes `email`. Email follows the login account. `administrator` and `deactivated` stay off the self-write allowlist.
- When more than one active Tran user matches the login email, or the LoginID fallback, resolution fails closed. It does not pick a row and it does not create a new one. A single match and a complete miss (auto-create) stay as they are.
- The profile form shows email as read-only. GET `/api/tran/users` is unchanged.

## Capabilities

### New Capabilities

- `tran-user-admin`: Who may create, update, and delete Tran users, what a user may change on their own profile, and how a login email resolves to one active Tran user.

### Modified Capabilities

- `morph-data-api-auth`: A valid session still reaches Tran handlers, but a non-admin session must not succeed at creating, updating, or deleting Tran users.

## Impact

- `morph/handlers/tran_users.go` (admin gate, self-write allowlist, email lookup).
- `morph/handlers/tran_tools.go` (`tranUserIDFromContext` email match only; no owner-scoping rewrite).
- `morph/handlers/authz_middleware.go` is reused, not rewritten. `requireAdmin` stays the admin check.
- `morph/db/sqlite_schema.go` gains `User.DeactivatedDate` so the existing deactivate SQL succeeds on SQLite.
- `morph/frontend/src/pages/admin/UserSettings.js` makes email read-only.
- No JWT, login, or environment-variable changes. Forms, hybrid context, registration, and chat upload admin fallbacks stay for issue #88. Notes owner scoping stays for PR #87.
