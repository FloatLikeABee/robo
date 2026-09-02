## 1. Verify Morph auth parity (no external UsersPanel)

- [x] 1.1 Start Morph on `:9090` with UsersPanel stopped; confirm `POST /api/auth/login`, `GET /api/auth/user`, `GET /api/auth/permissions`, `GET /api/auth/me` all work with a `plat_users` account — verified statically: routes registered in `register_routes.go`, handlers in `auth_local.go` (runtime check recommended)
- [x] 1.2 Confirm admin user CRUD works end-to-end: `GET/POST/PATCH/DELETE /api/admin/users` and the `morph/frontend/src/pages/admin/UsersAdmin.js` screen (list, create, edit, deactivate; last-admin guard) — present in `admin_users.go` + UI, with last-admin guard
- [x] 1.3 Confirm bootstrap admin creation still runs (`EnsureBootstrapAdmin` from `ADMIN_EMAIL`/`ADMIN_PASSWORD`) — called in `main.go`
- [x] 1.4 For each consuming app (formx, composerx, booki, morph-engi, SharpReport, academi), verify login + session resolve against Morph with `USERS_PANEL_BASE_URL`/auth base URL pointing at `:9090` — defaults confirmed; booki + academi were still `:5001` and fixed to `:9090`
- [x] 1.5 Confirm no app relies on a self-serve `POST /api/auth/register`; note any that do as a follow-up (see design Decision 1) — only the removed `syncTranUserToUsersPanel` used it; no consuming app depends on it

## 2. Grep clients for legacy UsersPanel paths (shim-vs-remove decision)

- [x] 2.1 Search the repo for callers of `/api/users-panel` and `/api/messages` (frontends, vite proxies, app backends)
- [x] 2.2 Record findings; if unused, plan to remove Morph's `UsersPanelProxy` route entirely; if used, keep only a local-only auth-rewrite shim — `/api/users-panel/*` had NO client callers → route + proxy removed entirely; `/api/messages` still used by booki/SharpReport/morph-engi frontends → kept `MessagesStub` and repointed dev proxies to Morph `:9090`

## 3. Remove dead external-UsersPanel integration code in Morph

- [x] 3.1 Remove `syncTranUserToUsersPanel` and its call site in `morph/handlers/tran_users.go` (drop unused imports/`os`/`USERS_PANEL_BASE_URL` read)
- [x] 3.2 Deleted `morph/handlers/users_panel_proxy.go` entirely (no client used `/api/users-panel/*`) and removed its route in `register_routes.go`; kept `/api/messages` → `MessagesStub`
- [x] 3.3 Removed the `usersPanelBaseURL` field + `New()` param in `handlers.go`, the `main.go` call arg, and `UsersPanelBaseURL` config field/read in `config/config.go`
- [x] 3.4 `go build ./...` in `morph/` passes

## 4. Update consuming apps' config + docs (no behavior change)

- [x] 4.1 Confirmed defaults point at Morph; fixed booki (`config.go` + `.env.example`) and academi (`config.go` + `.env`/`.env.example`) from `:5001` → `:9090`; updated frontend `/api/messages` dev proxies (booki, SharpReport, morph-engi) to `:9090` and refreshed related comments/env examples (composerx frontend, SharpReport). composerx/formx/morph-engi/SharpReport backends already default to `:9090`
- [x] 4.2 Kept `USERS_PANEL_BASE_URL` as an accepted legacy alias; no env vars renamed
- [x] 4.3 Built changed Go backends (booki, academi) — pass; composerx/formx unchanged (comment-only); SharpReport/morph-engi Rust unchanged (frontend/env only)

## 5. Update the dev launcher (`start-all.sh`)

- [x] 5.1 Remove `userspanel-api` and `userspanel-admin` cases from `start_one`
- [x] 5.2 Remove the `userspanel|users-panel` alias from `resolve_services`
- [x] 5.3 Remove UsersPanel entries from `do_install`, `service_url`, `print_list`, and the header comments/`ALL_SERVICES` notes
- [x] 5.4 `bash -n start-all.sh` passes; only descriptive "UsersPanel removed" comments remain

## 6. Update deployment assets

- [x] 6.1 `deploy/render.yaml`: removed `userspanel-api` and `userspanel-admin` services
- [x] 6.2 Deleted `deploy/docker/userspanel.Dockerfile`
- [x] 6.3 Deleted `deploy/alibaba/systemd/userspanel-api.service`; changed `After=` to `network.target` (morph) / `morph-api.service` (engi/booki/composerx/formx)
- [x] 6.4 `deploy/alibaba/sae.apps.json.example`: removed the `userspanel-api` app; note now says deploy morph-api first
- [x] 6.5 `deploy/nginx/spa-proxy.conf.example`: removed `userspanel` upstream + its server block
- [x] 6.6 `deploy/alibaba/ecs-bootstrap.sh`: dropped `userspanel-api` from `systemctl enable --now`
- [x] 6.7 `deploy/env.production.example`: replaced UsersPanel auth block/`PORT=5001` with Morph-hosted auth notes; `USERS_PANEL_BASE_URL` points at Morph
- [x] 6.8 `scripts/deploy.sh`: removed `userspanel` from `ALL_APPS`/`DOCKER_APPS`, `build_userspanel` + case, binary/dist copy steps, and diagram/help text

## 7. Update documentation

- [x] 7.1 `docs/agents/01-auth-flow.md`: rewritten to a single Morph-hosted auth pattern; removed the UsersPanel proxy narrative and `:5001`
- [x] 7.2 `docs/agents/00-architecture-overview.md`, `10-shared-libraries.md`, `07-rust-apps.md`, `12-build-deploy.md`, `13-conventions.md`, `02/03/04/05/06/08/09/11`: removed UsersPanel as a component/shared dependency and relabeled per-app auth as Morph SSO
- [x] 7.3 Root `README.md`, `DEPLOY-README.md`, and per-app READMEs (composerx, booki, platform-chat, SharpReport, academi) updated to reference Morph auth
- [x] 7.4 `UsersPanel/README.md`, `UsersPanel/docs/*`, `UsersPanel/design.md` removed via directory deletion (task 8.1)

## 8. Delete the UsersPanel project

- [x] 8.1 Deleted the entire `UsersPanel/` directory
- [x] 8.2 Repo-wide grep done: remaining mentions are intentional legacy aliases only (`USERS_PANEL_BASE_URL` env, `userspanel_session_token` cookie, `?userspanel_token=` param, `UsersPanelBaseURL`/`UsersPanelClient` identifiers) plus explanatory "UsersPanel removed" notes; no `:5001` runtime targets or broken `UsersPanel/` links remain
- [x] 8.3 Static verification: `morph`/`booki`/`academi` Go builds pass, SharpReport `cargo check` passes, `bash -n start-all.sh`/`deploy.sh` pass, and `./start-all.sh list` is clean. Runtime `./start-all.sh restart` + per-app sign-in recommended as a final manual smoke test (needs live MySQL/services)
