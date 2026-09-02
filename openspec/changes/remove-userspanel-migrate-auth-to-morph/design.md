## Context

See `proposal.md` — Why. The important starting fact: the auth migration to Morph has effectively already happened at runtime.

- Morph serves the auth contract: `POST /api/auth/login`, `GET /api/auth/user`, `GET /api/auth/permissions`, `GET /api/auth/me` (`morph/handlers/auth_local.go`, `register_routes.go`), plus admin CRUD `GET|POST|PATCH|DELETE /api/admin/users` over MySQL `plat_users` (`morph/handlers/admin_users.go`, `db/plat_users.go`) and a Users admin UI (`morph/frontend/src/pages/admin/UsersAdmin.js`).
- Consuming apps call `/api/auth/login|user|permissions` against a configurable base URL that already defaults to Morph `http://127.0.0.1:9090`: composerx (`USERS_PANEL_BASE_URL`), booki (`usersPanelBaseURL`), SharpReport (`users_panel.base_url`, also honoring `MORPH_API_BASE_URL`), morph-engi, academi, formx.
- The dev launcher default stack already excludes UsersPanel; it survives only as a deprecated `userspanel` alias.
- Remaining coupling to the *external* UsersPanel service: Morph's `UsersPanelProxy` (forwards `/api/users-panel/*` to `usersPanelBaseURL` when set, stubs `/api/messages`), `syncTranUserToUsersPanel` in `tran_users.go` (POSTs new users to an external `/api/auth/register`), and the deploy/build/docs assets that still provision `:5001`.

This change is therefore mostly **verification + deletion + cleanup**, not new feature work.

## Goals / Non-Goals

**Goals:**
- Guarantee Morph is a complete, self-contained replacement for the "simple user mechanism" (login, session, permissions, admin user CRUD) with no dependency on an external UsersPanel.
- Physically remove the `UsersPanel/` project and every launcher/deploy/build/doc reference to it.
- Remove now-dead external-UsersPanel integration code paths without changing the stable `/api/auth/*` and `/api/admin/users` contracts other apps depend on.

**Non-Goals:**
- No changes to the JWT scheme, `plat_users` schema, password hashing, or the request/response shapes of `/api/auth/*` and `/api/admin/users`.
- No port-forwarding or advanced RBAC beyond the existing admin flag / full-permissions model (UsersPanel's roles/permissions admin, messaging, data-collector, and AI assistant are intentionally dropped — "simple" mechanism only).
- No rename of app-level auth env vars beyond documentation (kept backward compatible).

## Decisions

### 1. Morph stays the auth source of truth; verify parity before deleting
Rather than porting UsersPanel code, we confirm Morph already covers the needed surface and only fill gaps. Parity checklist to verify during apply: login, `/api/auth/user`, `/api/auth/permissions`, `/api/auth/me`, admin user CRUD, bootstrap admin, and JWT accepted by every consuming app.
- **Gap to resolve:** self-serve `POST /api/auth/register` does not exist on Morph. It is currently only invoked by `syncTranUserToUsersPanel` (removed here) and is not part of the "simple" login flow (admins create users via `/api/admin/users`). Decision: **do not add** a public register endpoint; drop the sync instead. If any consuming app truly relies on self-registration, that surfaces in the parity check and is handled as a follow-up.
- *Alternative considered:* extract a shared auth crate from UsersPanel into `pkg/`. Rejected — duplicates what Morph already ships and expands scope.

### 2. Reduce `UsersPanelProxy` to a local-only compatibility shim
Keep the rewrite that maps legacy `/api/users-panel/api/auth/*` to Morph's local auth handlers (cheap safety net for any client still using the old path), but **remove the external reverse-proxy branch, the `usersPanelBaseURL` handler field, and the `/api/messages` stub** so there is no path that reaches an external service. During apply, grep clients (`morph-utils`, app frontends, vite proxies) for `/api/users-panel` and `/api/messages`; if genuinely unused, remove the route entirely instead of keeping the shim.
- *Alternative considered:* delete the proxy immediately. Deferred to the usage check to avoid breaking an overlooked caller.

### 3. Keep `USERS_PANEL_BASE_URL` as a legacy alias for the Morph auth base URL
Consuming apps already default it to Morph. Renaming across six apps is churn with no functional gain and risks breaking existing `.env` files. Decision: keep the env var name working (SharpReport already also accepts `MORPH_API_BASE_URL`), and update docs/examples to describe it as "Morph auth base URL". Defaults point at `http://127.0.0.1:9090`.
- *Alternative considered:* rename to `MORPH_AUTH_BASE_URL` everywhere. Rejected for churn; can be a later cosmetic pass.

### 4. Delete UsersPanel wholesale, including deploy assets
Remove `UsersPanel/` and all provisioning: `deploy/docker/userspanel.Dockerfile`, `deploy/alibaba/systemd/userspanel-api.service`, and the `userspanel-*` entries/ordering in `render.yaml`, `sae.apps.json.example`, `spa-proxy.conf.example`, `ecs-bootstrap.sh`, `env.production.example`, `scripts/deploy.sh`, and `start-all.sh`. Update the systemd `After=userspanel-api.service` ordering to `After=network.target` (or `morph-api.service`).

## Risks / Trade-offs

- **A consuming app or external `.env` still points `USERS_PANEL_BASE_URL` at `:5001`.** → Env still works because the name is retained; the value simply must be Morph. Docs/examples updated; defaults already Morph. Deleting UsersPanel makes any lingering `:5001` value fail loudly (connection refused) rather than silently — acceptable and discoverable.
- **A forgotten client calls `/api/users-panel/*` or `/api/messages`.** → Mitigated by the grep-before-remove step in Decision 2; if found, keep the local-only shim. `/api/messages` stub returning empty is harmless if retained.
- **Removing the sync leaves admin-created Morph users unmirrored.** → No longer meaningful once UsersPanel is gone; Morph `plat_users` is the only store.
- **Production still running the old UsersPanel binary/service.** → Migration plan below decommissions it after Morph is confirmed serving auth.

## Migration Plan

1. Verify parity (Decision 1) against a running Morph with UsersPanel stopped; confirm each app logs in and resolves sessions.
2. Grep for `/api/users-panel` and `/api/messages` clients; decide shim-vs-remove (Decision 2).
3. Remove dead integration code (`syncTranUserToUsersPanel` + call site, external proxy branch/field/stub). Build Morph and affected Go/Rust apps.
4. Update `start-all.sh`, deploy manifests, and docs.
5. Delete the `UsersPanel/` directory and `deploy/docker/userspanel.Dockerfile` + `userspanel-api.service`.
6. In each environment, stop and deregister the `userspanel-api`/`userspanel-admin` services and free port `:5001`.

**Rollback:** restore the `UsersPanel/` directory and its deploy assets from version control and re-point `USERS_PANEL_BASE_URL` at the restored `:5001`; Morph auth continues to work regardless, so rollback is only needed if an unforeseen dependency on the standalone service appears.
