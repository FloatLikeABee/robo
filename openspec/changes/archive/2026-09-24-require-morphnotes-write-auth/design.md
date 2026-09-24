## Context

See proposal.md for why. The gate is `AuthzMiddleware` in `morph/handlers/authz_middleware.go`. Before this change, `isOpenMorphDataAPI` returned true for every path under `/api/tran/`, `/api/forms/`, `/api/knowledge/`, and `/api/graph/`, and the middleware called `Next` even when `resolveUserScope` failed. A separate prefix check let every method through on `/api/tran/public/`.

`resolveUserScope` accepts a Morph JWT when `TranMySQL` is set, and otherwise falls back to `X-User-ID` / `X-User-Role`. A non-empty bearer token that fails to decode returns false and does not consult the headers.

The MorphNotes SPA (`morph/frontend/src/appRouter.js`) does not wrap `/morphdata` in `ProtectedLayout`. The comment there says login lives only on Morph AI and Morph Data stays open. `tranClient.js` attaches `Authorization` only when `getMorphToken()` finds `userspanel_session_token`, and it deliberately does not redirect `/morphdata` to `/login` on 401. MorphUtils puts that same token on iframe URLs as `?userspanel_token=`. `execManagementAPI` copies `Authorization` onto internal requests and, if `X-User-ID` is empty, sets it to `admin`.

Published routes that exist today are GET only: `/api/tran/public/big-notes/:slug`, `/api/tran/public/timelines/:slug`, `/api/tran/public/research/:slug`. The published Big note HTML form does not POST back to the API.

CORS `OPTIONS` is aborted with 204 in `morph/main.go` before this middleware runs.

## Goals / Non-Goals

**Goals:**

- Anonymous mutating calls on the four prefixes return 401, including Research create and publish.
- A valid Morph JWT still reaches the existing handlers.
- Public published pages are an explicit GET/HEAD allowlist, not a method-blind prefix.
- Tests fail on today's middleware and pass after the change.
- Docs name the public routes and say writes need a JWT.

**Non-Goals:**

- Removing `X-User-ID` impersonation (issue #23).
- Forcing MorphNotes list/detail GETs behind login, or changing the SPA login redirect.
- BK / AI tools auth, `.github/`, or `docs/agents/12-build-deploy.md`.

## Decisions

### 1. Require a session for unsafe methods; keep private GET/HEAD open

Unsafe methods are `POST`, `PUT`, `PATCH`, `DELETE`, and any method that is not `GET` or `HEAD`. On the four prefixes those methods 401 when `resolveUserScope` is false. `GET` and `HEAD` on those prefixes still proceed without a session, and a present session is still attached via `applyAuthScope`.

This is the decision that survived review. The first code pass denied every method except the published-page allowlist. That draft was rejected. See alternatives below.

### 2. Public pages are an allowlist, not a prefix

`isPublicMorphRead` allows only `GET` and `HEAD` of `/api/tran/public/{big-notes|timelines|research}/{single-segment-slug}`. `POST /api/tran/public/research/:slug` and `GET /api/tran/public/other/:slug` are not public. Because those paths still start with `/api/tran/`, the unsafe-method rule returns 401 instead of falling through to a 404 after an open prefix.

New published kinds must be added to `publicMorphReadKinds` and registered as GET. The allowlist is what "default to deny" means here: nothing is public unless it is named. It does not mean every private GET becomes 401.

### 3. Sibling prefixes get the same write rule

`/api/forms/`, `/api/knowledge/`, and `/api/graph/` share `isOpenMorphDataAPI` today, so they share the unsafe-method rule. `POST /api/graph/search` is a query, but it is `POST`, so it requires a session. The management tool loop already forwards the caller's JWT, which is how a signed-in Morph AI user keeps calling it.

### 4. Do not change `resolveUserScope`

JWT success, invalid-JWT failure, and header fallback stay as they are. An invalid bearer on a write is 401 and does not become an anonymous write. A header-only caller still mutates. That keeps #23 as a separate change and keeps `execManagementAPI` working for both JWT callers and legacy header callers.

### 5. Docs describe the split

`docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, and `morph/README.md` list the public GET/HEAD routes and state that writes on the four prefixes need a Morph JWT. They also state that private reads on those prefixes stay unauthenticated in this story.

## Alternatives considered

Played as proposer and reviewer against the code, not against the issue text alone.

### A. Unsafe methods only, private reads stay open (chosen)

Matches the issue acceptance criteria and the instruction to give `/api/forms`, `/api/knowledge`, and `/api/graph` the same treatment for writes. `appRouter.js` and `tranClient.js` already treat MorphNotes as usable without login, and the 401 interceptor will not send those users to `/login`. Closing reads would turn every anonymous list into an error with no recovery path, and this story did not accept a frontend change (the Morph production build is already failing on mermaid, issue #10).

Failure mode that remains: anonymous `GET /api/tran/research` and attachment downloads still return data. That is a real shared-host leak. It is documented as remaining, not described as fixed.

### B. Default deny for every method except the published-page allowlist (rejected)

This was the first implementation. It matches a strict reading of "default to deny" and the sentence in `docs/agents/01-auth-flow.md` that other `/api/*` routes require a bearer token. It also closes the read leak.

Rejected because it contradicts the SPA: Morph Data is outside `ProtectedLayout`, and `tranClient.js` comments that Morph Data is usable without login and must not bounce to login on 401. The issue title and acceptance criteria name mutating APIs. The sibling-prefix note says "the same treatment for writes." Shipping B without a login redirect would break a path the code calls legitimate. A frontend redirect was considered and rejected because it touches the Morph frontend, whose production build is out of scope to repair.

### C. Require a valid JWT and ignore `X-User-ID` on these routes (rejected)

The acceptance line "with no JWT" can be read this way. It is issue #23. `execManagementAPI` still sets `X-User-ID` to `admin` when the outer request has none. Header-authenticated chat (still allowed until #23) would then 401 on inner writes. Out of scope.

### D. Leave `isOpenMorphDataAPI` and only special-case Research create/publish (rejected)

Smallest patch, and it would satisfy a narrow reading of the audit example. It leaves every other `POST`/`PATCH`/`DELETE` on `/api/tran` open, which fails the acceptance criteria. A method-blind `/api/tran/public/` prefix would also keep accepting writes if a route were added later.

### Failure modes checked in code

- Published questionnaire HTML does not POST to the API (`onsubmit="return false"`, no fetch). No anonymous write is required for public pages.
- `OPTIONS` never reaches this middleware.
- `GET` handlers on these prefixes are list, get, download, health, and public HTML. Allowing anonymous GET does not invoke create/publish. If a future GET mutates, that handler is a separate bug; this middleware will not catch it.
- Invalid JWT does not fall through to `X-User-ID` when `TranMySQL` is set. Writes with a bad token stay 401 even though the tool loop also sets `X-User-ID`.
- When the outer chat request has a valid JWT, `execManagementAPI` forwards it, so inner Research create still hits the JWT branch.

## Risks / Trade-offs

- [Anonymous clients can still read MorphNotes, forms, knowledge, and graph health] → Called out in the docs and the PR. A later story can deny private GET/HEAD without changing the public allowlist.
- [Header impersonation still authorizes writes] → Left for #23. A test locks the current behavior so this story does not silently remove it.
- [Signed-out MorphNotes can list data and then fail on Save] → Same as the SPA's current split, except Save now 401s instead of writing. No frontend change; the existing error path surfaces the 401.
- [`POST /api/graph/search` from a logged-out page 401s] → Morph AI chat already requires a session before the tool loop runs.

## Migration Plan

Deploy the Morph API. No data migration. Rollback is reverting the middleware: anonymous writes work again. No cookie or token format change.

## Open Questions

None that change the spec. Whether private reads should later require a JWT is a follow-up, not this story.
