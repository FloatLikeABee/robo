## Context

See proposal.md for why. Issue #58's acceptance criteria override the earlier brief where they differ: initialize over stdio, declare tools and resources capabilities (an empty resource list is acceptable), a ping-class tool is enough, Cursor `mcp.json` docs, and a pinned official Go SDK. The earlier brief still holds for everything #58 does not contradict: stdio only, read-only, no network listener, no secrets, do not touch `morph/frontend/` or `pkg/morphai`.

Code checked for this design:

- `morph/main.go` opens Badger (`db.New`, `db.NewBadgerEntityDetails`) and SQLite (`db.NewTranSQL`) in the API process.
- Badger v4 `acquireDirectoryLock` takes `LOCK_EX` for a normal open and `LOCK_SH` for `ReadOnly`. Either flock fails with "Cannot acquire directory lock" while the API holds `LOCK_EX` (`dir_unix.go`).
- `db.NewTranSQL` sets `journal_mode(WAL)` and runs schema ensure. Those are writes. WAL itself does allow a separate `mode=ro` reader.
- `morph/auth/jwt.go` issues and checks HS256 session JWTs (`sub`, email, username, roles). `LoadTokenConfig` reads `JWT_SECRET`. There is no API-token table.
- `resolveUserScope` in `authz_middleware.go` also accepts `X-User-ID` when no bearer is present. That header is not a credential.

## Goals / Non-Goals

**Goals:**

- A stdio binary clients can launch, that answers `initialize`, `tools/list`, `resources/list`, and `whoami`.
- Identity and data-access choices that do not break a running API and do not impersonate a user.
- A package boundary a later in-process HTTP transport can call.

**Non-Goals:**

- MorphNotes list/get tools, `morph://` resources, Streamable HTTP, per-user API tokens, write tools, Engi projects, BK TCP MCP.

## Decisions

### 1. Do not open Badger or SQLite in this process

The stdio server opens no Morph store. `whoami` uses JWT claims only. Startup cannot take or fail on the API's Badger lock, and it does not run SQLite schema writes.

Alternatives:

- Open the same Badger directories as the API (the design spike). Rejected. A second `badger.Open` requests `LOCK_EX` and fails while `morph-api` is up. It can also block the API from reopening the directory.
- Open Badger with `ReadOnly: true`. Rejected. Read-only still calls `acquireDirectoryLock` with `LOCK_SH`. A shared flock does not succeed against the API's exclusive flock, so the skeleton would still fail beside a running API. Badger's own read-only tests are skipped or require the writer to be closed.
- Open SQLite read-only now (`mode=ro`, `query_only=1`) and skip Badger. Rejected for this story. `whoami` does not need rows. A missing or not-yet-WAL file would make handshake startup depend on the database. `db.NewTranSQL` is the wrong opener: it writes `journal_mode` and runs migrations against the live file.
- Proxy tool calls to `http://127.0.0.1:9090`. Rejected. It adds a network client, depends on the API process, and rides the HTTP auth surface (including the `X-User-ID` fallback). The spike already rejected HTTP-to-self as the primary path.

Failure mode of the choice: a later change that "just calls `db.New`" will fight the API lock. Guard it with a test that `morph/mcp` and `morph/cmd/morph-mcp` do not import `idongivaflyinfa/db` or Badger. Story 2 reads SQLite only, with a read-only DSN, and still does not open Badger. The read-only opener belongs in that story, not in `db.NewTranSQL`.

### 2. Identity is a verified session JWT, not a user id

`MORPH_MCP_TOKEN` is a Morph session JWT. `auth.DecodeToken` checks HS256 with `auth.LoadTokenConfig` (`JWT_SECRET`, same default as the API when unset). Missing, invalid, expired, and subject-less tokens exit non-zero before any stdout byte. The error string and logs omit the token. `whoami` returns `sub`, email, username, and roles.

A raw user id is rejected. `MORPH_MCP_USER_ID` would be the stdio form of `X-User-ID` impersonation.

Alternatives:

- Env user id for local dev. Rejected. No signature, no expiry, and it trains the next story to skip auth.
- JWT plus a `plat_users` lookup before serving. Stronger, and it matches `resolveUserScope`, but it needs the read-only SQLite open that decision 1 defers. Rejected for this skeleton. Deleted users still resolve until the token expires. Default JWT expiry is very long (`JWT_EXPIRY_HOURS`, otherwise about 100 years). This story does not add revocation. Streamable HTTP waits on real API tokens.
- `GET /api/auth/me` with the bearer token. Rejected. The stdio process would not start unless the API is up, and it would trust HTTP instead of the shared auth helper.

The process loads the repo-root `.env` without overriding variables already set, so a desktop client should pass `JWT_SECRET` and `MORPH_MCP_TOKEN` in its MCP config. The token must not be committed.

### 3. Advertise tools and resources; register only whoami

Issue #58 requires both capabilities. The original brief advertised resources only after a resource was registered. #58 wins.

`initialize` sets tools (inferred from `whoami`) and resources (`listChanged: true`, no subscribe). `resources/list` returns an empty page. No `morph://` URIs in this story. `whoami` is the ping-class tool and returns the resolved user, which also satisfies the earlier brief. Logging capability stays off (the SDK would otherwise advertise it by default).

### 4. Layout and protocol

`morph/mcp` builds the server and checks identity. `morph/cmd/morph-mcp` is the stdio main: logs on stderr, stdout only via the SDK stdio transport, no listener.

SDK pin: `github.com/modelcontextprotocol/go-sdk` v1.8.0. Tests and the doc smoke use 2025-06-18. The server does not refuse a newer revision the SDK implements (including 2026-07-28). Restricting the server to only 2025-06-18 was rejected because #58 says "prefer" that revision for Cursor/Claude, and the original brief allows negotiating newer revisions.

### Grill notes that did not change the decisions

- Process listings can show `MORPH_MCP_TOKEN` because it is an environment variable. Stdio has no header channel. Mitigation is local-only use, no logging, and no committed token. Not a reason to accept an unsigned user id.
- Extra JSON fields the SDK adds on list results (cache hints) are ignored by 2025-06-18 clients. Not a reason to fork the SDK.
- `whoami` does not prove the Badger lock is free. The proof is the import and startup path: this process never calls `db.New`, `db.NewBadgerEntityDetails`, or `db.NewTranSQL`. A test that the handshake works with no data files covers the observable half.

## Risks / Trade-offs

- [Long-lived session JWT, no `plat_users` check] → Document it. Story 2 confirms `sub` on a read-only SQLite connection. Do not ship Streamable HTTP on this identity.
- [Empty resources capability teaches clients the server has resources] → `resources/list` is empty and honest. Real URIs are a later story.
- [Someone reuses `db.NewTranSQL` from the stdio process] → Package comment plus this decision. The helper writes.
- [Unset `JWT_SECRET` uses the API's development default] → Same as `auth.LoadTokenConfig`. Production must set the secret in the client env. Docs say the values must match and must not be committed.

## Migration Plan

New binary and docs only. No schema change and no data migration. Rollback is removing the command from client config; the API process is unchanged.

## Open Questions

- Exact read-only SQLite DSN flags for story 2, after a fixture proves `mode=ro` beside a WAL writer. That does not change this story's handshake.
