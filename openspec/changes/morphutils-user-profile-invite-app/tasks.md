## 1. Morph API — data model

- [x] 1.1 Add `plat_invite_codes` table to SQLite schema (`morph/db/sqlite_schema.go` + `Ensure*` in new `plat_invite_codes.go`)
- [x] 1.2 Add `CreatePlatUserFromInvite` helper: generated username/password, placeholder email, non-admin role
- [x] 1.3 Add username uniqueness + `UpdatePlatUserCredentials` for self-service PATCH

## 2. Morph API — handlers & routes

- [x] 2.1 `PATCH /api/auth/me` — update own username and/or password (current password required for password change)
- [x] 2.2 `POST /api/admin/invite-codes` + `GET /api/admin/invite-codes` (admin only)
- [x] 2.3 `POST /api/invite/redeem` (public, single-use) — returns `{ username, password }` once
- [x] 2.4 Register routes in `register_routes.go`; add handler tests for redeem + self-service update

## 3. Invite Signup app (`invite-signup/`)

- [x] 3.1 Scaffold Vite React app on port 3050 with `/api` proxy to Morph `:9090`
- [x] 3.2 Redeem page: code input → call redeem API → show credentials with copy buttons
- [x] 3.3 Admin page: Morph login → create code → display new code; list existing codes (used/unused)
- [x] 3.4 Add `invite-signup-ui` to `start-all.sh` (optional service, URL in status)

## 4. MorphUtils — account modal

- [x] 4.1 Add user icon in sidebar footer when session is valid
- [x] 4.2 `UserProfileModal` component: load `/api/auth/me`, edit username, change password (current + new)
- [x] 4.3 Wire `PATCH /api/auth/me` with inline errors; dark-theme styles matching shell

## 5. MorphNotes — remove Users admin

- [x] 5.1 Remove `configuration/users` route and AppDrawer **Users** nav item
- [x] 5.2 Remove or stop importing `UsersAdmin.js`; clean `platformUiDefaults` nav label if orphaned
- [x] 5.3 Grep MorphNotes for remaining links to user admin page

## 6. Verification

- [x] 6.1 Admin creates code in Invite Signup → redeem → login on Morph AI with generated credentials
- [x] 6.2 Signed-in user updates username/password from MorphUtils modal
- [x] 6.3 MorphNotes Settings has no Users page; non-admin cannot create codes
