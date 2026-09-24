## Context

See proposal.md for why. The gate is `AuthzMiddleware` in `morph/handlers/authz_middleware.go`. `isPublicMorphRead` already allowlists `GET`/`HEAD` of `/api/tran/public/{big-notes,timelines,research}/{slug}`. Everything else under `/api/tran/`, `/api/forms/`, `/api/knowledge/`, and `/api/graph/` is `isMorphDataAPI`. Today that branch lets `GET` and `HEAD` through when `resolveUserScope` fails.

Publish is a non-empty `published_slug` on `research`, `big_note`, and `timeline`. There is no second published flag. `ServePublicResearch`, `ServePublicBigNote`, and `ServePublicTimeline` select `WHERE published_slug = ?`. No other Morph API route serves a published page. Project's `GET /api/v1/public/projects/:slug` is a different process and is not in this change.

`NewTranSQL` (what `morph/main.go` opens) creates graph tables in `ensureTranSQLiteSchema`. `NewTranMySQLLegacy` calls `EnsureGraphKnowledgeSchema`. `GraphHealth` calls it again and returns `neo4j_uri` plus `neo4j_error`.

`DecodeToken` enforces production lifetime (`iat` and `exp` present, window ≤ 168h) and does not reject a future `iat`. `jwt.WithIssuedAt` is not set. Local mode still accepts a token with no `iat` (`morph/auth/jwt_lifetime_test.go`).

MorphNotes is `/morphdata` in `appRouter.js`, outside `ProtectedLayout`. The comment says Morph Data stays open. `tranClient.js` skips the login redirect when the path is `/morphdata`, `/forms`, or `/transfinderx`, and `window.location.assign('/login')` drops the page the user was on. `LoginPage` reads `location.state.from.pathname` only.

`morph/mcp` has `whoami` only and refuses to start without a verified JWT. It does not import `handlers`. `IsAdminRoles` has no callers. `chat.go` and `chat_sessions.go` set user id `admin` when `X-User-ID` is empty.

Clients of the private GETs: the Morph SPA (`tranApi` adds `Authorization` when `userspanel_session_token` is set). MorphUtils passes that token into iframes as `?userspanel_token=` and calls only `/api/auth/*`. Event Logs, Content Maker, Data Access, Project, and Invite Signup do not call `/api/tran`, `/api/forms`, `/api/knowledge`, or `/api/graph`.

## Goals / Non-Goals

**Goals:**

- Anonymous `GET`/`HEAD` on the four prefixes is 401, including personal-data lists and downloads.
- The three published page routes stay public, and only for a non-empty published slug.
- One plain Go function decides "published," callable from handlers and `morph/mcp` without Gin.
- Signed-out `/morphdata` and a 401 in the Morph UI go to login and back, with no open redirect.
- Graph health does not write schema and does not return the URI or a raw error.
- Production decode rejects a future `iat`, with one minute of leeway.
- Chat fails closed when the user id is empty. Dead `IsAdminRoles` is removed.

**Non-Goals:**

- `/u/:handle`, API tokens, sharing/ACLs, SheetX or Content Maker header hardening.
- A Morph MCP tool that reads SQLite or Badger.
- Changing `forms.go`, `hybrid_context_handlers.go`, `registration.go`, or `chat_file.go` admin fallbacks. Those are the same smell on other handlers; this story's review note is chat, and the sibling is `chat_sessions.go`.
- Reading the session cookie inside the API. Clients keep sending `Authorization: Bearer`.

## Decisions

### 1. Require a session for every morph-data method except the published-page allowlist

After the existing `isPublicMorphRead` early return, `isMorphDataAPI` with no session is 401 for every method, including `GET` and `HEAD`. A valid session still calls `applyAuthScope` and `Next`. `OPTIONS`, `/api/auth/login`, `/api/invite/redeem`, and the auth handlers that check the bearer themselves stay as they are.

This replaces the safe-method exception added for issue #22. That exception is the bug #69 names.

### 2. Published means a trimmed non-empty slug, in package `publish`

`idongivaflyinfa/publish` has no Gin import:

- `PageRoute(method, path string) bool` — the current `isPublicMorphRead` rules, moved so the allowlist is not trapped in the middleware file.
- `Visible(publishedSlug string) bool` — true only when `strings.TrimSpace(publishedSlug) != ""`.

Middleware calls `PageRoute`. Each `ServePublic*` handler returns 404 when `!Visible(slug)` before it would serve HTML. The SQL lookup stays. A row with a null slug cannot match, and the function is what MCP calls.

`morph/mcp.ExposeRecord(caller Identity, publishedSlug string) bool` returns `publish.Visible(publishedSlug)` when `caller.UserID` is empty, and true when the caller has a verified user id. Unauthenticated callers therefore cannot receive a private record. No new MCP tool and no store open. `whoami` is unchanged and still requires a user id at `NewServer`.

### 3. Graph health is read-only and quiet

Delete the `EnsureGraphKnowledgeSchema` call in `GraphHealth`. Do not add another startup call; both store constructors already ensure the tables. Drop JSON fields `neo4j_uri` and `neo4j_error`. Keep `enabled`, `neo4j_ok`, `embeddings`, `outbox_pending`, and `knowledge_library`. Do not put `err.Error()` in the body or in a log line (driver errors include the URI).

### 4. Future `iat` only in production, one minute of leeway

Production `jwt.ParseWithClaims` options add `jwt.WithIssuedAt()` and `jwt.WithLeeway(time.Minute)`. Local parsing does not, so a token with no `iat` still decodes in local mode. One minute also softens `exp` and `nbf` by the same amount. That is the clock skew the review asked for, not a longer lifetime.

### 5. Login return is a query param checked by one function

`safeReturnPath` in `morph/frontend/src/auth/returnTo.js` accepts only a string that starts with a single `/`, has no backslash, no scheme, and still has the same origin after `new URL(value, 'http://morph.local')`. Anything else becomes `''`, and login then goes to `/`.

`ProtectedLayout` wraps `/morphdata` as well as Morph AI. It navigates to `/login?returnTo=` plus the encoded pathname, search, and hash. `tranClient` does the same on 401, including when the path is `/morphdata`, and does nothing when the path is already `/login` or the request URL is `/api/auth/login`. `LoginPage` prefers `returnTo`, then router state (pathname + search + hash), then `/`.

`/forms/*` already redirects to `/morphdata/big-notes`, so it lands on the same gate.

### 6. Chat user id

`sessionUserID` returns the trimmed `X-User-ID` or false. `chat.go` and every handler in `chat_sessions.go` respond 401 when it is false. They do not substitute `admin`. Delete `IsAdminRoles`.

### 7. Docs name the public set

`docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, `morph/README.md`, and `docs/agents/14-morph-mcp.md` say private reads need a session, list the three published page routes as the only anonymous data reads, and say MCP uses `publish.Visible` for any later record. The hosting checklist notes that production rejects a future `iat` within the one-minute leeway. `scripts/check-operator-product-docs.py` still requires `GET /api/tran/public/research/:slug` in `03-morph.md`; that line stays.

## Alternatives considered

Proposer and reviewer, against the code.

### A. Middleware 401 on every morph-data method, published pages stay the existing allowlist (chosen)

Matches the acceptance list and the SPA change the issue now requires. The #22 design rejected this because MorphNotes had no login redirect. That redirect is in scope here, so the earlier rejection does not hold.

Failure mode: a new `GET` under `/api/tran` is private unless someone adds it to `publish.PageRoute` and registers it. That is the intended default.

### B. Check auth inside each handler (rejected)

`register_routes.go` has dozens of GETs. A missed handler stays public. One branch in the middleware covers lists, details, and downloads, including routes added later under those prefixes.

### C. Treat the whole `/api/tran/public/` prefix as public and filter only in SQL (rejected)

Unknown kinds and `POST` under that prefix would skip the session again. The allowlist already closed that. Keep it.

### D. New `published` column (rejected)

Publish already writes `published_slug`. A second flag can disagree with the slug the public URL uses. `Visible` is the slug.

### E. Put `Visible` in `handlers` and import that from `mcp` (rejected)

`handlers` imports Gin, the database, and the AI client. `mcp` is a stdio process that must not open those stores. A file with two functions does not.

### F. Add an MCP tool that reads published rows (rejected)

The v1 server is `whoami` only, and it must not open SQLite while the API holds the file. `ExposeRecord` is the gate a later tool calls. Building the tool is a different story.

### G. Return 404 for anonymous private reads (rejected)

The acceptance text says 401, and the SPA uses 401 to send the user to login. The public slug route stays 404 when nothing is published under that slug, so a guess does not confirm a private id.

### H. `WithIssuedAt` in every environment (rejected)

Local tests and existing dev tokens omit `iat` and must keep working. Production already has a separate lifetime path. The future-`iat` check joins that path.

### I. React Router state only, no query param (rejected)

`tranClient` uses `window.location.assign`. Router state does not survive that. The query param is the return path both redirects can set. The validator runs before `navigate` after login.

### J. Fix every `userID = "admin"` fallback in this change (rejected)

`chat.go` and `chat_sessions.go` are one user-id path. Forms, hybrid context, registration, and chat file uploads are separate handlers. Changing them without a failing test per handler is a drive-by, and some tests call those handlers with an explicit `X-User-ID`.

## Risks / Trade-offs

- [Open redirect via `returnTo=//evil.example` or `https://evil.example`] → `safeReturnPath` rejects those shapes and a unit test locks the cases.
- [Login 401 loop] → no redirect when the path is `/login` or the failing URL is `/api/auth/login`.
- [Published HTML requested as an SPA route] → those URLs are `/api/tran/public/...`, proxied to the API, not `/morphdata`.
- [`WithLeeway` also accepts `exp` up to one minute past expiry] → same one-minute skew; the 168-hour cap is unchanged.
- [Graph health used to create MySQL tables on first GET] → those tables are created when the legacy MySQL store opens. SQLite already creates them at `NewTranSQL`. A process that called `GraphHealth` without opening through those constructors would not create tables; `main` does not do that.
- [Signed-in data scoping] → handlers are unchanged. Middleware still attaches the token user. Anonymous access is what changes.
- [Local MorphNotes now asks for login] → same default account as today. No new environment variable.

## Migration Plan

No data migration and no new configuration. Deploy the API and the Morph SPA together. Existing JWTs keep working unless `iat` is in the future (production). Rollback is reverting the middleware exception and the SPA gate; published pages do not change shape.

## Open Questions

None. The public route list is the three slug routes plus login, invite redeem, and the auth handlers that check the bearer themselves. No other Morph API published-page route exists.
