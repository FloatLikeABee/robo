## Context

See proposal.md for why. The gate is `AuthzMiddleware` in `morph/handlers/authz_middleware.go`. `resolveUserScope` decodes `Authorization: Bearer` when `TranMySQL` is set and the token is non-empty. If that token fails, it returns no session and does not read headers. If the token is missing, or `TranMySQL` is nil, it accepts `X-User-ID` and treats `X-User-Role` / `X-User-Roles` containing `admin` as an admin. `applyAuthScope` then writes `X-User-ID`, `X-User-Email`, and `X-User-Role` for handlers that still read those names.

The Morph SPA (`tranClient.js`, `morphSession.js`) sends `Authorization: Bearer` from the `userspanel_session_token` cookie. It does not send `X-User-*`. The API never reads that cookie itself. `execManagementAPI` in `morph/handlers/internal_api.go` copies `Authorization` and also sets `X-User-ID`, defaulting to `admin` when the header is empty, then copies `X-User-Role`, `X-User-Roles`, and `X-User-Permissions`.

A repo search for `X-User-ID`, `X-User-Role`, `X-User-Email`, `X-User-Roles`, and `X-User-Permissions` found one sender toward Morph: `execManagementAPI`. SheetX (`formx`) and Content Maker (`composerx`) read those headers on their own servers and, when a bearer token is present, resolve the role from Morph with `Authorization`. Data Access stamps `x-user-*` on its own request after a Morph session lookup. Project, AI tools, invite-signup, and the Morph SPA do not send the headers to Morph. Handler unit tests that call `CreateUserNoteTodo` directly set `X-User-ID` and do not mount `AuthzMiddleware`.

CORS in `morph/main.go` aborts `OPTIONS` with 204 before auth and sets `Access-Control-Allow-Headers: *`. `isSafeMethod` is only GET and HEAD, so `OPTIONS` would be treated as an authenticated method if it reached the middleware. `isPublicMorphRead` trims slashes on the slug, so `research//slug` and `research/slug/` match a published page. Gin decodes `%2F` into `URL.Path` (`UseRawPath` is false). Query strings are not part of `URL.Path`.

PR #70 (issue #24, not merged) edits Morph startup and `morph/auth/jwt.go`. This change does not edit `morph/auth/jwt.go` or `morph/config`. `morph/main.go` changes stay in the CORS helper and the `r.Use` CORS registration.

## Goals / Non-Goals

**Goals:**

- No request is a session unless `Authorization` carries a valid Morph JWT for an existing `plat_users` row.
- Client identity headers cannot select the user or the role, with or without a token.
- The management tool loop proves it forwarded the exact `Authorization` value.
- Published pages require exactly `{kind}/{slug}`.
- `OPTIONS` is not 401. CORS names the headers the SPA sends.
- Dev docs describe the removal. Local `./start-all.sh start morph` stays the same.

**Non-Goals:**

- Requiring a session on anonymous GET/HEAD (#69), including `?user_id=` on `tranUserIDFromContext`.
- Renaming `userspanel_session_token`, teaching the API to read that cookie, or changing SPA login redirects.
- `/api/graph/health` schema work and the MorphNotes sign-in prompt.
- A shared-secret or `X-Internal-Token` for service calls.
- Changing SheetX, Content Maker, or Data Access auth. They do not authenticate to Morph with these headers.
- Wiring `morph-mcp` to this function. The stdio server must not open SQLite. It already verifies claims with `auth.DecodeToken`.

## Decisions

### 1. JWT plus the `plat_users` row is the only session

`ResolveUserScope(r *http.Request) (UserScope, bool)` reads the bearer token, decodes it with the existing `auth.DecodeToken`, and loads `GetPlatUserByID` for `claims.Subject`. The role is that row's current role (`IsAdmin`), not `X-User-*` and not the role list inside the token by itself. The Gin middleware calls this function. Missing token, invalid token, unknown user, or a nil store is not a session.

The issue's "identity comes from the token only" is the contrast with client headers. The token chooses the account. The account row supplies the live role, which is what the JWT branch already does. Switching to raw claims would keep a demoted user as admin until the token expired and would skip the existence check.

**Alternative A — delete only the header fallback.** Smaller diff. Rejected: `execManagementAPI` would still inject `X-User-ID: admin`, and `applyAuthScope` does not clear `X-User-Roles` or `X-User-Permissions`, so a spoofed header would remain visible to any handler that reads it.

**Alternative B — trust JWT claims and skip the database.** Easier for MCP, which cannot open SQLite. Rejected for HTTP: a deleted user or a stale admin claim would still pass. MCP keeps its existing claims check.

**Alternative C — replace the header with a shared service secret.** Rejected. The only in-repo caller already has the user's JWT. The issue says not to invent a shared secret when no service-to-service gap exists.

### 2. Strip client identity headers, then stamp the resolved session

On every `/api` request the middleware deletes `X-User-ID`, `X-User-Role`, `X-User-Roles`, `X-User-Email`, and `X-User-Permissions` before handlers run. `applyAuthScope` writes `X-User-ID`, `X-User-Email`, and `X-User-Role` only after a session resolves. Handlers that still read `X-User-ID` (chat, skills, forms, knowledge, notes) keep working for a signed-in caller. They do not see the client's values.

Anonymous GET/HEAD still reach those handlers (#69). After the strip, a client-supplied `X-User-ID` on those reads is ignored. The SPA never sends it. Direct unit tests that skip the middleware still set the header; that is in-process, not an HTTP trust path.

`execManagementAPI` forwards `Authorization` and stops setting identity headers, including the `admin` default. The inner `ServeHTTP` runs the same middleware, which resolves the forwarded token and stamps headers for the inner handler.

### 3. Published paths are one kind and one slug

`isPublicMorphRead` accepts only `GET` and `HEAD` of `/api/tran/public/{kind}/{slug}` where `kind` is exactly `big-notes`, `timelines`, or `research`, and `slug` has no `/` and is not empty, `.`, or `..`. No `Trim` of extra slashes. Case-sensitive. `%2F` is rejected because the decoded path has an extra segment. A query string is not part of `URL.Path`, so it does not affect the match. Shapes that fail the allowlist and still reach the middleware under `/api/tran/public/` are 401 rather than an open read.

Gin checks trailing-slash redirects before the handler chain. `GET /api/tran/public/research/:slug/` therefore 301s to the exact slug and never enters `AuthzMiddleware`. That redirect is not the published HTML. Turning off `RedirectTrailingSlash` would change every other route, and `main.go` edits stay in the CORS block, so the redirect stays. The helper still rejects the slashed path, which is what a caller would see if the redirect were disabled.

### 4. OPTIONS pass-through and an explicit CORS list

`AuthzMiddleware` returns `c.Next()` for `OPTIONS` before the session check. CORS in `main.go` still answers preflight with 204; the pass-through is the second safeguard if that abort is reordered. `Access-Control-Allow-Headers` becomes `Authorization, Content-Type, Accept`. Those are the headers the Morph SPA sets. `X-User-*` is not on the list. `Access-Control-Allow-Origin` behavior is unchanged.

### 5. Tests assert the behavior the review called out

- Header-only admin on chat, admin, and `POST /api/tran/research` is 401, and the research row count does not change.
- Invalid bearer plus `X-User-ID` is 401, and the row count does not change.
- The management-call spy is a handler mounted without `AuthzMiddleware`, so it sees the executor's request rather than headers the middleware would strip and rewrite. It requires `Authorization` to equal the outer value and requires `X-User-ID` to be empty. A following test still runs the real Research create through the middleware so a forwarded token still returns 201. A 201 alone is not the proof: today's executor also defaults `X-User-ID` to `admin`, which is enough to pass a status check while the bearer copy is missing.
- A valid employee JWT plus `X-User-Role: admin` and a foreign `X-User-ID` is 403 on `/api/admin/users`. A chat stub returns the token user's id and role.
- Table cases for `%2F`, trailing slash, `//`, dot segments, case, and query strings.

### 6. Docs state the removal

`docs/agents/01-auth-flow.md` currently says the headers still satisfy the session check. Replace that with the removal. `morph/README.md` and `docs/agents/03-morph.md` already say writes require a JWT; add that identity headers are not a session and that chat and admin require a JWT too, so those sentences stay true. Swagger comments on chat must not tell clients to send `X-User-ID`.

## Risks / Trade-offs

- [Header-only scripts get 401] → Dev docs say so. No in-repo caller depends on the fallback. Rollback is reverting the commit.
- [Anonymous reads ignore `X-User-ID`] → Same result the SPA already gets, because it does not send the header. `?user_id=` is unchanged and is not a session.
- [CORS list is too short] → Limited to headers the SPA sets. A new browser header needs a list edit and a test.
- [Handlers still read `X-User-ID`] → Safe on the HTTP path because the middleware stamps it after the strip. A route mounted without `AuthzMiddleware` would trust a caller-set header again. Production mounts the middleware globally in `main.go`.
- [`ResolveUserScope` needs `Handlers` and SQLite] → MCP does not call it in this change. A later owner check can.
- [PR #70 touches `main.go`] → Diff limited to the CORS helper and its `r.Use` line.

## Migration Plan

Deploy the Morph API binary. Clients of protected routes send `Authorization: Bearer <Morph JWT>`. No database migration. No new env var. `./start-all.sh start morph` and the default login are unchanged.

## Open Questions

None. Caller search, the JWT-versus-database role choice, and the decision not to add a service secret are settled above.
