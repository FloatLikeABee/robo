## Context

See proposal.md — Why. Today:

- MorphNotes **Settings → Users** (`UsersAdmin`) lets admins CRUD `plat_users` via `/api/admin/users`.
- MorphUtils has no account UI; it only checks the shared `userspanel_session_token` cookie.
- Auth login supports username or email (`MorphAuthLogin`); `plat_users` already has unique `username` and `email`.
- There is no self-service password/username API and no invite-code flow.

## Goals / Non-Goals

**Goals:**

- Remove user admin UI from MorphNotes.
- MorphUtils: user icon → modal for own username/password only.
- Standalone Invite Signup SPA (separate port, minimal dark UI).
- Morph API: `PATCH /api/auth/me`, invite-code table + admin/redeem endpoints.
- Simple generated credentials on redeem (memorable charset, uniqueness enforced server-side).

**Non-Goals:**

- Email verification, password reset email, OAuth.
- MorphNotes tran-user profile fields (`/api/tran/users/me` first/last name) — out of scope unless reused for display name later.
- Removing `/api/admin/users` HTTP handlers entirely (keep for API/bootstrap; hide UI).
- Multi-use invite codes or role assignment via code (all redeem → standard non-admin user).
- New database engine — reuse tran SQLite `plat_users` + new `plat_invite_codes` table.

## Decisions

### 1. Backend stays on Morph (`morph/`)

**Choice:** Add `plat_invite_codes` to tran SQLite schema and new handlers in `morph/handlers/`. Invite Signup SPA calls Morph API (same as other apps use `USERS_PANEL_BASE_URL`).

**Why:** `plat_users` already lives in Morph; a second user DB would break SSO.

**Alternative:** Separate microservice — rejected (YAGNI, duplicate auth).

### 2. Invite Signup as new top-level app `invite-signup/`

**Choice:** Vite + React (match morph-utils stack), port **3050**, proxy `/api` → Morph `:9090`. Two routes: `/` redeem, `/admin` codes. Not listed in MorphUtils module nav.

**Why:** User asked for app “not connected to anything morph” in UX terms — standalone URL/bookmark.

### 3. Credential generation on redeem

**Choice:** Server generates:

- Username: `user` + 4 random lowercase letters (retry on collision), e.g. `userbkfm`.
- Password: 8 chars from `abcdefghjkmnpqrstuvwxyz23456789` (no ambiguous `i/l/o/0/1`).

Store placeholder email `{username}@invite.local` (internal, not shown) to satisfy NOT NULL email column.

**Why:** Simple, typeable, unique username without user input.

**Alternative:** Let user pick username at redeem — more UI; defer.

### 4. Invite codes

**Choice:** 10-char uppercase alphanumeric code (crypto random), single-use, no expiry in v1 (optional `expires_at` column nullable for later).

Admin list endpoint `GET /api/admin/invite-codes` shows code, created_at, redeemed_at, redeemed_by (no passwords).

### 5. MorphUtils profile modal

**Choice:** Add `Person` icon button in `morph-utils-sidebar-footer` when `authed`. Modal component calls `GET /api/auth/me` + `PATCH /api/auth/me`. Match existing MorphUtils dark tokens.

Password change requires current password field.

### 6. Remove MorphNotes Users admin

**Choice:** Delete route `configuration/users`, remove AppDrawer nav item, delete or orphan `UsersAdmin.js`. Update `platformUiDefaults` label `nav_user_settings` if referenced.

Keep `/api/admin/users` for emergency/scripts; Invite Signup admin uses invite endpoints only.

## Risks / Trade-offs

- [Invite code leakage] → Single-use; show “treat like password” copy on admin create.
- [Generated password shown once] → Clear UI warning; user must copy before navigating away.
- [Placeholder invite emails] → Not used for login; document that login is username-only for invite users.
- [Admin API still exists] → MorphNotes UI removed; invite app is the supported provisioning path.

## Migration Plan

1. Ship Morph API + DB migration (`plat_invite_codes`).
2. Ship Invite Signup app + `start-all.sh` service `invite-signup-ui`.
3. Ship MorphUtils profile modal.
4. Remove MorphNotes Users nav/page.
5. Existing users unchanged; new users via codes only going forward.

Rollback: restore MorphNotes Users route; invite codes table can remain unused.

## Open Questions

None — expiry policy and multi-use codes can be added later without breaking v1.
