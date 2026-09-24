## Why

On a hosted Morph, anyone who can reach the API can list tasks, research, knowledge files, people, and graph state without logging in. Issue #22 closed anonymous writes. Issue #69 (epic #4, blocks the v1 deploy #52) closes private reads. Only content the user explicitly publishes stays public, and a signed-out MorphNotes visitor is sent to login and returned to the page they opened.

## What Changes

- **BREAKING:** `GET` and `HEAD` on `/api/tran`, `/api/forms`, `/api/knowledge`, and `/api/graph` return 401 with no Morph session. That includes lists, details, downloads, graph health and search, and personal-data lists (`/api/tran/users`, `/members`, `/employees`, `/contacts`).
- Published HTML stays public: `GET` and `HEAD` of `/api/tran/public/{big-notes,timelines,research}/:slug`, and only when that record's published slug is set. An unpublished item is not reachable by slug.
- Signed-out visitors who open MorphNotes, or who get a 401 from these APIs in the Morph UI, go to login with a same-origin return path and land back after login.
- `GET /api/graph/health` no longer runs schema setup and no longer returns the Neo4j URI or a raw connection error.
- Production JWT parsing rejects a token whose `iat` is in the future, with a small clock-skew leeway.
- The published-only check is a plain Go function. The read-only MCP server calls it, so an unauthenticated caller cannot see a private record through that path.
- Chat stops treating a missing user id as `admin`. The unused `IsAdminRoles` helper is removed.

## Capabilities

### New Capabilities

- `published-record-visibility`: In-process rule for "this record is published," shared by public page handlers and `morph/mcp`.
- `morphnotes-login-return`: Signed-out MorphNotes and Morph API 401s redirect to login with a same-origin return path.

### Modified Capabilities

- `morph-data-api-auth`: Private reads on the four prefixes require a session. Published pages stay public. Graph health does not write or leak. The old requirement that private reads stay anonymous is removed.
- `morph-session-identity`: Production tokens with a future `iat` are rejected. An empty user id on chat is not an admin session.

## Impact

- Morph API middleware and public page handlers (`morph/handlers`), graph health (`morph/handlers/knowledge.go`), JWT decode (`morph/auth`), MCP (`morph/mcp`).
- Morph SPA: `/morphdata` requires a token; `tranClient.js` redirects on 401 with a validated return path.
- Operator docs that currently say private reads stay public (`docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, `morph/README.md`, `docs/agents/14-morph-mcp.md`, hosting checklist).
- MorphUtils, Event Logs, Content Maker, Data Access, and Project do not call these GETs. MorphUtils already passes `?userspanel_token=` into iframes. The Morph SPA already sends the bearer token when a session cookie is present.
- Out of scope: `/u/:handle`, SheetX/ComposerX header hardening, API tokens, sharing/ACLs.
