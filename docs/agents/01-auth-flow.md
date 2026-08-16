# 01 — Authentication Flow

## Overview

There are **two auth patterns** in this codebase. Know which app you're working on before touching auth code.

```
┌─────────────────────────────────────────────────────────────────┐
│                     AUTH PATTERNS                                │
│                                                                 │
│  PATTERN A: UsersPanel-dependent (most apps)                    │
│  ┌──────────┐     ┌──────────┐     ┌──────────────────────┐    │
│  │  Client   │────▶│ App API  │────▶│  UsersPanel (:5001)  │    │
│  │ (browser) │     │ /auth/   │     │  /api/auth/login     │    │
│  │           │     │  login   │     │  /api/auth/user      │    │
│  └──────────┘     └──────────┘     └──────────────────────┘    │
│                                                                 │
│  PATTERN B: Self-hosted (morph only)                            │
│  ┌──────────┐     ┌──────────────────────────────────────┐     │
│  │  Client   │────▶│  morph (:9090)                       │     │
│  │ (browser) │     │  /api/auth/login → MySQL plat_users  │     │
│  │           │     │  JWT (HS256) + bcrypt passwords      │     │
│  └──────────┘     └──────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

## Pattern A: UsersPanel-dependent apps

**Apps:** formx, composerx, booki, morph-engi, SharpReport, academi

### How it works

1. Client POSTs credentials to the **app's** `/auth/login` endpoint.
2. The app **proxies** the request to UsersPanel's `/api/auth/login`.
3. UsersPanel validates credentials against its MySQL `users` table, returns a JWT.
4. The app returns the JWT to the client (sometimes wrapped in a local token).
5. Subsequent requests include `Authorization: Bearer <token>`.
6. The app validates the token by calling UsersPanel's `/api/auth/user` or by decoding the JWT locally (shared secret).

### Per-app auth endpoints

| App | Login endpoint | Validation method |
|-----|---------------|-------------------|
| formx | `POST /api/v1/auth/login` | Proxies to UsersPanel |
| composerx | `POST /auth/login` | Proxies to UsersPanel |
| booki | `POST /api/v1/auth/login` + `POST /api/v1/auth/platform-session` | UsersPanel SSO exchange |
| morph-engi | `POST /api/v1/auth/login` + `POST /api/v1/auth/platform-session` | UsersPanel SSO exchange |
| SharpReport | `POST /api/v1/auth/login` | UsersPanel SSO |
| academi | `POST /api/v1/auth/login` | UsersPanel SSO + local BadgerDB fallback |

### Shared cookie

All apps share a cookie named `userspanel_session_token`. This is set by UsersPanel on login and read by all other apps for SSO. The `morph-utils` shell passes this token to embedded iframes via URL parameter (`?userspanel_token=...`).

### Adding auth to a new endpoint

```go
// Go apps (formx, composerx, booki, academi):
// Use the existing auth middleware. In Gin:
router.Use(authMiddleware())

// The middleware extracts the Bearer token, validates against UsersPanel,
// and sets user info in the Gin context (c.Get("user_id"), etc.)

// Rust apps (UsersPanel, SharpReport, morph-engi):
// Use the existing auth layer. In Axum:
.route("/protected", get(handler).layer(require_auth()))
```

## Pattern B: Morph self-hosted auth

**App:** morph only

### How it works

1. Client POSTs to `POST /api/auth/login` with email + password.
2. Morph looks up the user in MySQL `plat_users` table.
3. Password is verified with **bcrypt** comparison.
4. Morph issues its own JWT (HS256, secret from `JWT_SECRET` env var).
5. The JWT contains: `Email`, `Username`, `Roles`, `DefaultChannelID`, `Subject` (user ID).
6. Default expiry: ~100 years (effectively no timeout per recent product change).

### Key files

| File | Purpose |
|------|---------|
| `morph/auth/jwt.go` | JWT encode/decode, `LoadTokenConfig()`, `EncodeToken()`, `DecodeToken()` |
| `morph/handlers/auth_local.go` | `MorphAuthLogin`, `MorphAuthMe`, `MorphAuthUser`, `MorphAuthPermissions` |
| `morph/handlers/authz_middleware.go` | `AuthzMiddleware()` — validates Bearer token on all `/api/` routes |
| `morph/db/plat_users.go` | `PlatUser` struct, `EnsurePlatUsersTable()`, `EnsureBootstrapAdmin()`, CRUD |

### Auth middleware flow

```
Request → AuthzMiddleware()
  ├── /api/auth/login → pass through (public)
  ├── /api/tran/public/* → pass through (public)
  ├── /api/auth/me, /api/auth/user, /api/auth/permissions → pass through (self-validate)
  └── All other /api/* → require Bearer token
       ├── Extract token from Authorization header
       ├── Decode JWT
       ├── Look up user in MySQL plat_users
       └── Set in Gin context: auth_user_id, auth_role, auth_email, auth_is_admin
```

### Admin bootstrap

On first startup, if MySQL connects, morph calls `EnsureBootstrapAdmin()` which creates an admin user from `ADMIN_EMAIL`/`ADMIN_PASSWORD` env vars (default: `admin@local.com` / `admin`).

### ⚠️ Critical: morph does NOT use UsersPanel

Despite what `DEVELOPER_BASELINE.md` says, morph has been migrated to self-hosted auth. The `UsersPanelProxy` handler in `morph/handlers/register_routes.go` rewrites legacy UsersPanel paths to local morph auth. Do **not** add UsersPanel dependencies to morph auth code.

## Frontend auth patterns

### React apps (morph, formx, booki, morph-utils)

- JWT stored in `localStorage` + `sessionStorage` + cookie (`userspanel_session_token`)
- Axios interceptor adds `Authorization: Bearer <token>` to all requests
- On 401 response: clear tokens, redirect to `/login`
- Key files: `morph/frontend/src/auth/morphSession.js`, `formx/frontend/src/lib/api.ts`, `booki/frontend/src/lib/api.ts`

### Svelte apps (composerx, UsersPanel, morph-engi, SharpReport)

- JWT stored in `localStorage` + cookie
- Fetch wrapper adds Bearer token
- On 401: redirect to login
- Key pattern: check `App.svelte` in each frontend for auth state management

### morph-utils shell

- Reads `userspanel_session_token` cookie
- Passes token to embedded iframes via `?userspanel_token=` URL parameter
- Each embedded app consumes the token on load for SSO
- Key file: `morph-utils/frontend/src/auth.ts`