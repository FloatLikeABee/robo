# 01 — Authentication Flow

## Overview

**Morph hosts authentication** for every remaining app. Login accounts live in Morph SQLite (`plat_users` in `TRAN_SQLITE_PATH`, default `./data/tran.sqlite`). Passwords are bcrypt. Morph issues an HS256 JWT (`JWT_SECRET`).

```
  Browser ──▶ Morph POST /api/auth/login ──▶ SQLite plat_users
                    │
                    ▼
         JWT in cookie userspanel_session_token
                    │
     MorphUtils iframes  ?userspanel_token=...
                    │
     Event Logs / Content Maker / Data Access / Project
     validate via Morph GET /api/auth/user  (USERS_PANEL_BASE_URL)
```

A standalone UsersPanel service is **not** part of the supported stack. Keep the legacy names `USERS_PANEL_BASE_URL` and `userspanel_session_token`.

## How it works

1. Operator signs in on Morph (`:3031`) or a module login that proxies to Morph.
2. Morph verifies credentials and returns a JWT.
3. The browser stores the token in `userspanel_session_token` (and often `localStorage`).
4. MorphUtils reads the cookie and passes `?userspanel_token=` into iframes.
5. Embedded apps send `Authorization: Bearer <token>` and call Morph `GET /api/auth/user` / `GET /api/auth/permissions`.

### Auth base URL

Every consuming app reads `USERS_PANEL_BASE_URL` (legacy name). Default: `http://127.0.0.1:9090`. Point it at the Morph API in every environment. Some apps also accept `MORPH_AUTH_BASE_URL` / `MORPH_API_BASE_URL`.

### Apps that consume Morph SSO

| App | Typical login / session |
|-----|-------------------------|
| Morph | `POST /api/auth/login`, `GET /api/auth/me` |
| Event Logs (`formx`) | `POST /api/v1/auth/login` → Morph |
| Content Maker (`composerx`) | `POST /auth/login` → Morph. API routes require that bearer; `X-User-Role` and `X-User-Permissions` are not a session |
| Project (`morph-engi`) | Morph SSO (`USERS_PANEL_BASE_URL`) |
| Data Access (`SharpReport`) | Morph SSO; reuse the Morph cookie — do not add a second credential form |
| MorphUtils | Validates cookie against Morph `/api/auth/user` |
| AI tools (`bk`) | Linked from Morph AI with the same session |

### Morph internals (key files)

| File | Purpose |
|------|---------|
| `morph/auth/jwt.go` | Encode/decode JWT |
| `morph/handlers/auth_local.go` | Login, me, user, permissions |
| `morph/handlers/admin_users.go` | Admin user CRUD |
| `morph/handlers/authz_middleware.go` | Bearer on `/api/*` |
| `morph/db/plat_users.go` | `plat_users` in SQLite (legacy MySQL helpers may still exist in code) |
| `morph/frontend/src/auth/morphSession.js` | Morph AI session helpers |

### Which `/api/*` routes are public

Most `/api/*` routes need a Morph session (`Authorization: Bearer`). The only anonymous data reads are published pages. A record is published when its published slug is non-empty. That check is `publish.Visible` in `morph/publish`, which the HTTP handlers and `morph/mcp` both call.

| Method | Path | Why |
|--------|------|-----|
| POST | `/api/auth/login` | Sign-in |
| POST | `/api/invite/redeem` | Invite signup |
| GET, HEAD | `/api/tran/public/big-notes/:slug` | Published Big note HTML |
| GET, HEAD | `/api/tran/public/timelines/:slug` | Published Timeline HTML |
| GET, HEAD | `/api/tran/public/research/:slug` | Published Research HTML |

`GET /api/auth/me`, `/api/auth/user`, and `/api/auth/permissions` are not rejected by the middleware; each handler checks the Bearer token itself. `OPTIONS` is not rejected for lack of a session. `/health` is not under `/api`.

`GET`, `HEAD`, `POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` return 401 with no session. That includes lists, details, downloads, `GET /api/graph/health`, `POST /api/graph/search`, and `/api/tran/users`, `/members`, `/employees`, and `/contacts`. An unpublished research, big note, or timeline is not returned by a guessed public slug. `GET /api/graph/health` does not create schema and does not return the Neo4j URI or a raw connection error. Schema is created when the Tran store opens.

There is no other published page route on the Morph API. Project's public project HTML is a different service.

The MorphNotes SPA (`/morphdata`) sends a signed-out browser to `/login?returnTo=` and lands back on that path after login. The return path must be a same-origin relative path. A 401 from `tranApi`, including on Morph Data, does the same redirect. The SPA and Morph AI attach the bearer token from `userspanel_session_token` when the user is signed in. MorphUtils copies that token into iframes as `?userspanel_token=`. Event Logs, Content Maker, Data Access, and Project do not call these private GETs. The API does not read the session cookie. Morph AI's management tool loop (`internal_api.go`) forwards the caller's `Authorization` onto internal `/api/tran` calls and does not send identity headers.

A new published page must be registered as GET and added to `pageKinds` in `morph/publish/publish.go`. The path must be exactly `/api/tran/public/{kind}/{slug}` with one non-empty slug segment.

`X-User-ID`, `X-User-Role`, `X-User-Roles`, `X-User-Email`, and `X-User-Permissions` are not a session. There is no header fallback. Chat (`/api/chat`), admin (`/api/admin`), and the Morph data routes above require `Authorization: Bearer` with a Morph JWT. Spoofed identity headers do not replace the user or role in that token. Chat does not treat a missing user id as `admin`.

### Admin bootstrap

On startup Morph calls `EnsureBootstrapAdmin()` from env (`ADMIN_EMAIL` / `ADMIN_USERNAME` / `ADMIN_PASSWORD`). That insert does not change an existing row. Default operator login for local/dev (`MORPH_ENV` unset): **`morphadmin`** / **`admin123`** (or `morphadmin@local.com`). `./start-all.sh` needs no extra auth config.

`MORPH_ENV=production` refuses the development JWT secret, the development admin password (including a hash already stored in `plat_users`), and a JWT lifetime outside 1–168 hours. Production default lifetime is 24 hours when `JWT_EXPIRY_HOURS` is unset. One start with `MORPH_ROTATE_DEFAULT_ADMIN=1` replaces stored development passwords and keeps account ids. Checklist: [`docs/security-hosting-checklist.md`](../security-hosting-checklist.md). Startup errors name the variable to set and do not print secret values.

### MorphUtils SSO

Key file: `morph-utils/frontend/src/auth.ts`.

1. Read `userspanel_session_token`.
2. Validate against Morph `GET /api/auth/user`.
3. On login, set the cookie and reload.
4. Each iframe gets `?userspanel_token=`.
5. Data Access must not clear the shared cookie on 502 / empty 500 from its own `/auth/me` — only a real 401 invalidates the Morph session.

Product UIs are **dark-only**. Do not add a light/dark switch on login chrome.
