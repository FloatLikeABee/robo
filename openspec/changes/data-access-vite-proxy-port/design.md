## Context

See proposal.md — Why. SharpReport API already reads `SHARPREPORT_PORT` in `config.rs`. Vite `envDir` is the repo root, but proxy targets and `hooks.server.ts` still hardcode 3050. Nested `SharpReport/.env` `PORT=` is ignored by the backend; do not revive it.

## Goals / Non-Goals

**Goals:**

- One local source of truth: `SHARPREPORT_PORT` (default 3050) for API bind, Vite proxy, SSR rewrite, and `start-all.sh` status URL.
- Vite must restart to pick up a port change (same as today for other env).

**Non-Goals:**

- Changing auth, data-tables handlers, MorphUtils iframe URL (`VITE_DATAX_URL` stays 5178).
- Syncing `SHARPREPORT_BASE_URL` (Morph’s server-to-server Data Access URL is a separate mismatch).
- Binding Vite to IPv4 vs IPv6.

## Decisions

### 1. Follow `SHARPREPORT_PORT` instead of forcing 3050

- **Choice:** Read `SHARPREPORT_PORT` in Vite (`loadEnv` from repo root, prefix `''`) and in `hooks.server.ts` (`process.env` / `$env/dynamic/private`). Default `"3050"`.
- **Why:** Local `.env` already uses 3888; resetting the API to 3050 would fight that override and still leave the next port change broken.
- **Alternative:** Set `SHARPREPORT_PORT=3050` only. Rejected — the UI would stay hardcoded and drift again.

### 2. Keep `/api/messages` on Morph `:9090`

- Unchanged. Only `/api`, `/public`, `/metabase` use the SharpReport origin.

### 3. `start-all.sh` status uses the same env

- `service_url sharpreport-api` → `http://127.0.0.1:${SHARPREPORT_PORT:-3050}` after `apply_dotenv`.

## Risks / Trade-offs

- [Vite already running] → Restart `sharpreport-ui` after the proxy change; env is read at config load.
- [`SHARPREPORT_BASE_URL` still 3050] → Morph calling Data Access API directly can still miss. Out of scope; operators can align that env later.
- [Invalid port string] → Fall back to 3050 if parse fails.

## Migration Plan

- No data migration. After code lands: restart `sharpreport-ui` (and keep API on current `SHARPREPORT_PORT`). Rollback: revert proxy/hooks/start-all to 3050.

## Open Questions

- None.
