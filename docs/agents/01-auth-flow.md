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
| Content Maker (`composerx`) | `POST /auth/login` → Morph |
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

Public routes include `POST /api/auth/login` and selected `/api/tran/public/*`. Other `/api/*` routes require a Bearer token.

### Admin bootstrap

On startup Morph calls `EnsureBootstrapAdmin()` from env (`ADMIN_EMAIL` / `ADMIN_USERNAME` / `ADMIN_PASSWORD`). Default operator login in the root README: **`morphadmin`** / **`admin123`** (or `morphadmin@local.com`).

### MorphUtils SSO

Key file: `morph-utils/frontend/src/auth.ts`.

1. Read `userspanel_session_token`.
2. Validate against Morph `GET /api/auth/user`.
3. On login, set the cookie and reload.
4. Each iframe gets `?userspanel_token=`.
5. Data Access must not clear the shared cookie on 502 / empty 500 from its own `/auth/me` — only a real 401 invalidates the Morph session.

Product UIs are **dark-only**. Do not add a light/dark switch on login chrome.
