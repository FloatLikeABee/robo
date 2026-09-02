## 1. Dev proxy and SSR rewrite

- [x] 1.1 `vite.config.ts`: load `SHARPREPORT_PORT` from repo-root env (default 3050); proxy `/api`, `/public`, `/metabase` to that origin; keep `/api/messages` on Morph 9090
- [x] 1.2 `hooks.server.ts`: rewrite same-origin `/api/*` fetch to the same SharpReport origin (not hardcoded 3050)

## 2. Operator status and docs

- [x] 2.1 `start-all.sh` `sharpreport-api` URL uses `${SHARPREPORT_PORT:-3050}`
- [x] 2.2 Note in root `.env.example` that Data Access Vite/SSR follow `SHARPREPORT_PORT`

## 3. Verify

- [x] 3.1 Restart `sharpreport-ui` if it is running; `GET` `/api/v1/auth/me` and `/api/v1/data-tables` via `localhost:5178` reach the API (not empty 500 to 3050)
