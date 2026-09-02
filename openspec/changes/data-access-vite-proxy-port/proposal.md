## Why

Data Access in MorphUtils shows HTTP 500 on `/api/v1/auth/me` and `/api/v1/data-tables` when `SHARPREPORT_PORT` is not 3050. The API listens on that env port; the SvelteKit Vite proxy and SSR fetch rewrite stay hardcoded to 3050, so the UI talks to a dead port. Operators already set `SHARPREPORT_PORT=3888` locally; the tables page still looks like an app error.

## What Changes

- Dev Data Access UI (`localhost:5178`) MUST proxy `/api` (and related backend paths) to `127.0.0.1:${SHARPREPORT_PORT}`, defaulting to 3050 when the env is unset.
- SvelteKit server `fetch` rewrite MUST use the same port (SSR does not use the Vite proxy).
- `start-all.sh` status URL for `sharpreport-api` MUST show that same port.
- Default documented port stays **3050**. No API route or auth behavior change.

## Capabilities

### New Capabilities

- `data-access-dev-api-proxy`: Local Data Access UI reaches the SharpReport API on `SHARPREPORT_PORT` instead of a hardcoded 3050.

### Modified Capabilities

- (none)

## Impact

- `SharpReport/frontend/vite.config.ts`, `SharpReport/frontend/src/hooks.server.ts`
- `start-all.sh` `service_url` for `sharpreport-api`
- Root `.env.example` comment that Vite/start-all follow `SHARPREPORT_PORT`
- MorphUtils Data Access iframe (`VITE_DATAX_URL` remains 5178; only the API hop changes)
