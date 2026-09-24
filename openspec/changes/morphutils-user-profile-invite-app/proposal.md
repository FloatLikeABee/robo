## Why

MorphNotes **Settings → Users** exposes full admin user CRUD inside the notes admin shell. That belongs in MorphUtils (where operators already land for cross-app utilities) and does not match how we want to onboard people: admins should issue **invitation codes**, not hand-type accounts. End users need a dead-simple self-service path to get a username and password, and any signed-in user should be able to update **only their own** login from MorphUtils without opening MorphNotes configuration.

## What Changes

- **Remove** MorphNotes admin **Users** page (`configuration/users`, `UsersAdmin`) and the Settings nav entry; stop creating users from MorphNotes.
- **Add** MorphUtils **account icon** in the shell sidebar/footer that opens a modal for the **current session user** to view and change **username** and **password** (requires current password to change password).
- **Add** a **standalone Invite Signup** mini-app (separate Vite SPA + port, not embedded in MorphUtils/MorphNotes):
  - **Admin page**: root admin signs in with existing Morph credentials → create one-time (or limited-use) **invitation codes** to share out-of-band.
  - **Redeem page**: anyone with a code redeems it once → receives an auto-generated **simple unique username** and **simple password** (shown once; copy-friendly).
- **Extend Morph auth backend** (`plat_users` in tran SQLite): invite-code table + APIs, public redeem endpoint, self-service `PATCH /api/auth/me` for username/password.
- **Deprecate** direct admin user create UI in MorphNotes; root admin uses Invite Signup app for provisioning instead.

## Capabilities

### New Capabilities

- `morphutils-user-profile`: MorphUtils shell exposes a user icon and in-app modal so the signed-in user can read and update their own Morph login username and password.
- `invite-signup`: Standalone invite-signup SPA (admin login + code creation; public code redemption with generated credentials) backed by Morph auth APIs.

### Modified Capabilities

- `platform-auth`: Morph-hosted auth gains self-service account update (username/password for bearer owner) and invitation-code lifecycle (admin create, public redeem → `plat_users` row).

## Impact

- **MorphNotes** (`morph/frontend`): remove `UsersAdmin` route/nav; optional cleanup of unused `UserSettings.js` tran-profile page if still orphaned.
- **MorphUtils** (`morph-utils/frontend`): account icon, profile modal, calls new auth APIs.
- **Morph API** (`morph/handlers`, `morph/db`): `plat_invite_codes` table, invite + self-service handlers, route registration.
- **New app** (`invite-signup/` or similar): Vite React SPA, own port (~3050), wired in `start-all.sh` as optional service.
- **Auth**: existing JWT/cookie flow unchanged for other apps; new users created via redeem use same `plat_users` + login as today.
- **Bootstrap admin**: unchanged env-based root admin; Invite Signup admin UI requires `is_admin` Morph session.
