# MergeEmailX Frontend (Svelte + Vite)

Frontend UI for MergeEmailX: dashboard (default), saved emails, templates, **Contacts** (import CSV + API sync), merge data, attachments, and jobs. Compose via **New email**; add recipients from the contact picker in the sidebar.

## Prerequisites

- Node.js 18+ and npm
- MergeEmailX backend running (default on `http://localhost:8043`)

## Install dependencies

From the `frontend` directory:

```bash
npm install
```

## Configure API base (optional)

In **dev**, leave `VITE_API_BASE` unset — the Vite server on port **8044** proxies `/auth`, `/emails`, etc. to the backend on **8043** (same origin, no CORS).

For production builds or direct API access, set `VITE_API_BASE` (no trailing slash):

```bash
VITE_API_BASE="http://localhost:8043" npm run build
```

See `.env.example`.

**Login:** TranMail authenticates via **UsersPanel** (`USERS_PANEL_BASE_URL` on the backend, default `http://127.0.0.1:5001`). Ensure UsersPanel is running. Local dev users are in the `plat_users` table (e.g. `admin@local.com` if bootstrapped).

**AI assistant:** configured on the **backend** only — see [`../backend/README.md`](../backend/README.md#ai-provider-configuration) (`MORPH_AI_API_KEY`, optional `ai.config.json` and `TRAN_OPENAI_API_KEY` for reference RAG).

## Run the dev server

```bash
npm run dev
```

The app will be served on:

```text
http://localhost:8044
```

## Build for production

```bash
npm run build
```

To preview the production build:

```bash
npm run preview
```

This also serves on port **8044** by default (configurable in `vite.config.js`).
