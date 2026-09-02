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

## AI

[`pkg/morphai-rs`](../pkg/morphai-rs) for table/report help. Set `MORPH_AI_API_KEY` in the root `.env`. Operator chat is Morph AI, not a satellite drawer.

```bash
./start-all.sh restart sharpreport-api
```
