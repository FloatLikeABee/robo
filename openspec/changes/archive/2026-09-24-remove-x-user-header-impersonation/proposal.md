## Why

`resolveUserScope` in `morph/handlers/authz_middleware.go` treats `X-User-ID` / `X-User-Role` / `X-User-Email` as a Morph session when the request has no valid JWT. Any caller can claim to be any user, including admin, and reach chat, admin, and the MorphNotes write routes that issue #22 now gates on a session. Issue #23 (epic #4) removes that impersonation. A valid token is the only session.

## What Changes

- **BREAKING** for header-only clients: `X-User-ID`, `X-User-Role`, `X-User-Roles`, `X-User-Email`, and `X-User-Permissions` no longer authenticate. Chat, admin, and mutating `/api/tran`, `/api/forms`, `/api/knowledge`, and `/api/graph` routes return 401 when those headers are the only identity.
- A valid Morph JWT still identifies the caller. The user id and role come from that token's `plat_users` row. Spoofed identity headers sent with the token do not replace the user or the role.
- An invalid bearer token plus `X-User-ID` stays 401 and does not create a Research row.
- Morph AI's management tool loop forwards the caller's `Authorization` value and stops injecting `X-User-ID` (including the `admin` default).
- Published HTML stays public only for `GET`/`HEAD` of exactly `/api/tran/public/{kind}/{slug}` with one non-empty slug segment. `research//slug`, a trailing slash, `%2F`, dot segments, and a different kind case are not public.
- `OPTIONS` is not an authenticated method. CORS allows `Authorization`, `Content-Type`, and `Accept` by name and does not allow `X-User-*`.
- Dev docs say the header fallback is gone. There is no silent substitute.
- Token-to-user resolution is a plain Go function on `*http.Request`, so a later MCP owner check can call the same rules. This change does not add MCP tools.

## Capabilities

### New Capabilities

- `morph-session-identity`: Who counts as a Morph session on protected routes (chat, admin, and any non-public `/api` route), including header rejection, token-only identity, `OPTIONS`, and the CORS allow list.

### Modified Capabilities

- `morph-data-api-auth`: A legacy `X-User-ID` header is no longer a session for MorphNotes writes. Published pages require exactly `{kind}/{slug}`. The management tool loop must forward the exact `Authorization` header and must not authenticate with `X-User-*`.

## Impact

- `morph/handlers/authz_middleware.go`, `morph/handlers/internal_api.go`, handler tests, and the CORS block in `morph/main.go` only.
- Dev docs: `docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, `morph/README.md`.
- No frontend change. The Morph SPA already sends `Authorization: Bearer` from `userspanel_session_token` and does not send `X-User-*`.
- Local curl or scripts that called protected routes with only `X-User-ID` must send a Morph JWT. `./start-all.sh start morph` and the default login are unchanged.
- SheetX, Content Maker, Data Access, Project, AI tools, and invite-signup do not send `X-User-*` to Morph. Their own header handling is unchanged. No shared-secret service credential is added.
- Out of scope: requiring auth on anonymous GETs (#69), renaming `userspanel_session_token`, `/api/graph/health` schema side effects, the SPA sign-in prompt, `morph/auth/jwt.go`, and `morph/config`.
