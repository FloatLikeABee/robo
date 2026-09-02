# Content Maker frontend

Svelte + Vite UI for **Content Maker** (`composerx/`). API default `http://localhost:8043`. UI port **8044**.

## Prerequisites

- Node.js 18+ and npm
- Content Maker backend running
- Morph API for login (`USERS_PANEL_BASE_URL` on the backend)

## Install

```bash
npm install
```

In **dev**, leave `VITE_API_BASE` unset — Vite on **8044** proxies to **8043**.

For production builds:

```bash
VITE_API_BASE="http://localhost:8043" npm run build
```

AI keys are backend-only (`MORPH_AI_API_KEY`, optional `TRAN_OPENAI_API_KEY`). See [`../backend/README.md`](../backend/README.md).

```bash
npm run dev
```

http://localhost:8044

Or from repo root: `./start-all.sh start composerx-ui`.
