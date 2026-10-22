## Context

See proposal. Pool is MCP-only. dorph-mcp still does not run migrations or import `idongivaflyinfa/db`.

## Goals / Non-Goals

**Goals:** two-step consent, alias-only attribution, scanner that blocks, FTS index, save-copy, rate limits, three-report hide, stick shareable.

**Non-Goals:** subscriptions, admin UI, pool webpage, MCP Apps, argon2, cursor pagination, LLM scan.

## Decisions

### 1. Snapshot share, confirm_token in memory

`preview_share` copies title/body, strips `[dorph-mcp]` / `source: dorph-mcp`, scans, and stores a 5-minute token in process. `share_note` requires that token. Stdio is one process; a restart drops unused tokens.

**Rejected:** persisting preview rows (extra table for a 5-minute token).

### 2. Block on secrets or PII

User asked for a block, not optional redaction. Findings of type `secret` or `pii` refuse `share_note`. Preview still returns the findings.

### 3. CanReadShare, not ExposeRecord

`ExposeRecord` is a published-slug check. Pool uses `status=live` plus `pool:read`. Search lists `visibility=pool` only. `link` shares are get-by-id. Hidden and revoked are not found.

### 4. Stick and task may be shared

User override of the plan's v1 ban. Event logs stay unshareable (no owner snapshot in SQLite).

### 5. JWT empty scopes stay notes-only

Existing `HasScope` test: empty scopes must not grant `pool:publish`. Share requires `pool:publish` on the PAT. Session JWT users mint a PAT with that scope.

### 6. Rate limits in process

Counters keyed by user and action. Defaults: share 20/day, search 60/min, get 300/hour, report 20/day. Over → `rate_limited`.

## Risks / Trade-offs

- [In-memory token/limit lost on restart] → next preview/share is a new token; limits reset. Upgrade: SQLite rows.
- [Admin UI still shows every CaseTask] → unchanged. MCP is the isolation boundary.
