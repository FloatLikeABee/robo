## Why

The standalone **UsersPanel** project (Rust API on `:5001` + Svelte admin SPA) is no longer needed. Morph already self-hosts the full "simple user mechanism" the platform relies on — UsersPanel-compatible auth endpoints (`/api/auth/login`, `/api/auth/user`, `/api/auth/permissions`, `/api/auth/me`), admin user CRUD (`/api/admin/users`) over the MySQL `plat_users` table, and a Users admin UI (`morph/frontend/src/pages/admin/UsersAdmin.js`). Every consuming app already defaults its auth base URL to Morph (`http://127.0.0.1:9090`), and the dev launcher already excludes UsersPanel from the default stack. Keeping UsersPanel around is dead weight: duplicated auth code, an extra service to build/deploy/secure, and stale docs that still describe it as the central auth provider.

## What Changes

- **Consolidate the simple user mechanism in Morph** as the single source of truth for authentication and user management: login/session/permissions, and admin user CRUD (create/list/update/deactivate) with bcrypt passwords + HS256 JWT. Confirm and close any gaps so Morph is a complete drop-in for what UsersPanel provided.
- **Remove the entire `UsersPanel/` project** (Rust backend, Svelte admin, docs, mock-data). **BREAKING** for anyone still running UsersPanel on `:5001`.
- **Retire the external-UsersPanel compatibility shims / dead code** in Morph and consuming apps: the reverse-proxy passthrough to an external UsersPanel base URL, and the `syncTranUserToUsersPanel` register call. The `/api/users-panel/*` compatibility routes on Morph are kept only if still consumed by a client; otherwise they are removed.
- **Standardize the auth base-URL config** across apps (formx, composerx, booki, morph-engi, SharpReport, academi) to point at Morph. Keep the legacy env name `USERS_PANEL_BASE_URL` as an accepted alias (already defaulting to Morph) to minimize churn, documented as "Morph auth".
- **Update the dev launcher** (`start-all.sh`): remove `userspanel-api` / `userspanel-admin` service cases, the `userspanel` alias, install steps, and status/URL entries.
- **Update deployment assets**: delete `deploy/docker/userspanel.Dockerfile`, remove `userspanel-api` / `userspanel-admin` from `deploy/render.yaml`, `deploy/alibaba/sae.apps.json.example`, `deploy/alibaba/systemd/*` (drop the `After=userspanel-api.service` ordering and delete `userspanel-api.service`), `deploy/nginx/spa-proxy.conf.example`, `deploy/alibaba/ecs-bootstrap.sh`, `deploy/env.production.example`, and `scripts/deploy.sh`.
- **Update documentation** to describe a single Morph-hosted auth pattern: `docs/agents/01-auth-flow.md`, `docs/agents/00-architecture-overview.md`, `docs/agents/10-shared-libraries.md`, root `README.md`, and per-app READMEs that reference UsersPanel.

## Capabilities

### New Capabilities
- `platform-auth`: Morph-hosted authentication and simple user management for the whole platform — credential login issuing a shared JWT, session/user and permissions lookup, and admin-only user CRUD. Defines the endpoint contract that every other app depends on for single sign-on, replacing the standalone UsersPanel service.

### Modified Capabilities
<!-- No pre-existing OpenSpec specs describe UsersPanel; this change introduces the consolidated capability rather than modifying an existing spec. -->

## Impact

- **Removed:** `UsersPanel/` (entire directory), `deploy/docker/userspanel.Dockerfile`, `deploy/alibaba/systemd/userspanel-api.service`.
- **Morph code:** `handlers/users_panel_proxy.go` (proxy/stub cleanup), `handlers/tran_users.go` (`syncTranUserToUsersPanel` removal + call site), `handlers/register_routes.go` (route cleanup if `/api/users-panel/*` unused). No change to `plat_users` schema, JWT, or the `/api/auth/*` and `/api/admin/users` contracts.
- **Consuming apps (config only):** formx, composerx, booki, morph-engi, SharpReport, academi — auth base URL keeps pointing at Morph; env docs updated. No auth-flow behavior change for end users.
- **Tooling/deploy:** `start-all.sh`, `deploy/render.yaml`, `deploy/alibaba/*`, `deploy/nginx/spa-proxy.conf.example`, `deploy/env.production.example`, `scripts/deploy.sh`.
- **Docs:** auth-flow, architecture overview, shared-libraries, root and per-app READMEs.
- **Ports:** `:5001` (UsersPanel) is retired; `:9090` (Morph) hosts auth.
