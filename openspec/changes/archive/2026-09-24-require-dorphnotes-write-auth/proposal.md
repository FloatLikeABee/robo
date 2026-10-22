## Why

An unauthenticated `POST` that creates a DorphNotes Research job returns 200. `isOpenDorphDataAPI` in `dorph/handlers/authz_middleware.go` skips the session check for every method on `/api/tran/`, `/api/forms/`, `/api/knowledge/`, and `/api/graph/`. On a shared host, anyone can create, update, or delete that data. Issue #22 (epic #4) requires those writes to need a Dorph JWT while published HTML pages stay readable.

## What Changes

- **BREAKING** for anonymous clients: `POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*` (including Research create and publish), `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` return 401 when no session is present.
- A valid Dorph JWT still runs those mutations as they do today. Dorph AI's management tool loop keeps working because it forwards the caller's `Authorization` header.
- Published HTML stays public, but only as an explicit GET/HEAD allowlist: `/api/tran/public/big-notes/:slug`, `/api/tran/public/timelines/:slug`, `/api/tran/public/research/:slug`. A `/api/tran/public/` prefix is not enough; other methods on that prefix are 401.
- Private GET/HEAD on those four prefixes stay unauthenticated so the DorphNotes UI, which is not behind login, can still list data. That read exposure is called out as remaining, not closed here.
- `X-User-ID` / `X-User-Role` header fallback stays. Removing it is issue #23.
- Go tests cover anonymous reject, JWT allow, public GET, and the internal tool-loop forward.
- Security notes in `docs/agents/01-auth-flow.md`, `docs/agents/03-dorph.md`, and `dorph/README.md` list the public routes and state that writes need a JWT.

## Capabilities

### New Capabilities

- `dorph-data-api-auth`: Session rules for DorphNotes and the sibling Dorph data APIs (`/api/tran`, `/api/forms`, `/api/knowledge`, `/api/graph`), including the published-page allowlist.

### Modified Capabilities

- (none — `openspec/specs/` has no existing capability for this behavior)

## Impact

- `dorph/handlers/authz_middleware.go` and new handler tests.
- Docs listed above. No `.github/` or `docs/agents/12-build-deploy.md` edits.
- No frontend change. The SPA already sends the Dorph JWT on `tranApi` when a cookie is present, and DorphUtils already passes `?userspanel_token=` into iframes.
- Callers that mutated these APIs with no session and no `X-User-ID` header will get 401.
