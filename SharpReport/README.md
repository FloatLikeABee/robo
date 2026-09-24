# Data Access (`SharpReport/`)

Rust API + SvelteKit UI for **Data Access**: data tables, file-based reports, embedded Metabase. MorphUtils iframe: `/datax` on **5178**. Product UI is **dark-only** (no theme switch).

Agent notes: [`docs/agents/07-rust-apps.md`](../docs/agents/07-rust-apps.md). More detail: [`docs/API.md`](./docs/API.md), [`docs/DEVELOPMENT.md`](./docs/DEVELOPMENT.md), [`docs/DEPLOYMENT.md`](./docs/DEPLOYMENT.md).

## Stack

- **Backend:** Rust, Axum, SQLx (SQLite), Tokio, Metabase subprocess
- **Frontend:** SvelteKit + Svelte 5, TailwindCSS, ECharts
- **Auth:** Morph JWT (`USERS_PANEL_BASE_URL` → Morph `:9090`). Reuse `userspanel_session_token`; do not add a second login form.

## Ports

| | |
|--|--|
| UI | http://localhost:5178 |
| API | `SHARPREPORT_PORT` in the repo-root `.env` (often 3050). Vite **must** proxy 5178 to that port. |

## Prerequisites

- Rust (stable)
- Node.js 18+
- Java 17+ (Metabase)

No MySQL/Mongo/Redis required for local SQLite.

## Run

```bash
cp .env.example .env   # repo root
./start-all.sh start morph-api sharpreport
```

Manual:

```bash
cd backend && cargo run
cd frontend && npm install && npm run dev
```

If MorphUtils `/datax` refuses to connect, `sharpreport-ui` on 5178 is down or bound oddly — start it even if the API is already up.

## Image

One container runs the API and the production UI. Build context is the repo root so `pkg/morphai-rs` is available. No secret build args.

```bash
docker build -t sharpreport:local -f SharpReport/Dockerfile .
docker run --rm -p 127.0.0.1:3050:3050 sharpreport:local
```

`GET /health` and `GET /ready` return `{"status":"ok"}` without calling Morph. `GET /` is the UI. Local `start-all.sh` still uses API port 3050 and Vite on 5178. The image does not publish 5178.

| Variable | Image default |
|----------|----------------|
| `PORT` / `SHARPREPORT_PORT` | `3050` (`SHARPREPORT_PORT` wins when both are set) |
| `SHARPREPORT_DATABASE_URL` | `sqlite:///data/datapulse.db` |
| `SHARPREPORT_UI_DIR` | `/app/ui` |
| `USERS_PANEL_BASE_URL` | Morph auth base URL. Not a secret. Empty still passes `GET /health`. Set `https://<morph public host>` before expecting sign-in. |
| `JWT_SECRET` | Empty in the example. Set it before storing database connection passwords. |
| `MORPH_AI_API_KEY` | Optional. |

`docker compose -f SharpReport/docker-compose.yml up --build` reads `SharpReport/deploy/.env.production` when that file exists. Copy it from `SharpReport/deploy/.env.production.example`. The named volume is mounted at `/data`.

`sh SharpReport/deploy/check-container-contract.sh` checks the image contract without a daemon.

The product owner syncs `render.yaml` in project `prj-dahc33dbedkc73a1v8n0`. The service is `sharpreport`. The disk is `sharpreport-data` at `/data`. See [`deploy/README.md`](../deploy/README.md). The placeholder for story #114 is `https://<sharpreport public host>`. This change does not set `VITE_DATAX_URL`. Do not follow `scripts/deploy.sh`.

## AI

[`pkg/morphai-rs`](../pkg/morphai-rs) for table/report help. Set `MORPH_AI_API_KEY` in the root `.env`. Operator chat is Morph AI, not a satellite drawer.

```bash
./start-all.sh restart sharpreport-api
```
