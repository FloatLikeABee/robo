# Morph Engi — Projects

AI **project documents** from files or paste, plus a simple **files library**. Rust API + Svelte UI. Embedded as **Project** in Morph Utils.

## Quick start

```bash
cp .env.example .env

# From repo root (with morph-api running):
./start-all.sh start morph-api morph-engi-api morph-engi-ui

# Or manually:
cd backend && cargo run
cd frontend && npm install && npm run dev
```

| Service | URL |
|---------|-----|
| API | http://127.0.0.1:9096/health |
| UI | http://localhost:5179 |

## What it covers

- **Projects** — upload requirements/specs and/or paste content; AI organizes into markdown + HTML you can publish
- **Files** — retained uploads and paste-origin content (list, open, delete)

Primary navigation is Projects and Files only.

## AI assistant

`POST /api/v1/assistant/chat` — MorphAI tool loop for project documents and the files library. Set `MORPH_AI_API_KEY` in `.env`.

## AI project documents

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/projects/generate-document` | Multipart: `files` and/or `paste`; optional `title`. Requires at least one source |
| `POST /api/v1/projects/:id/publish` | Publish HTML to a public path |
| `GET /api/v1/public/projects/:slug` | Unauthenticated published HTML |
| `DELETE /api/v1/projects/:id` | Delete a project document |
| `GET/POST /api/v1/resource-files` | Files library |
| `POST /api/v1/resource-files/upload` | Store an uploaded file |
| `DELETE /api/v1/resource-files/:id` | Remove a files-library entry |

Limits: up to **5 files** per generate, each up to 8 MB. Accepted types: **PDF, TXT, CSV, MD**. Paste-only and file-only both work; an empty request is rejected before any AI call. Paste content is also saved into Files. Without `MORPH_AI_API_KEY`, generate returns HTTP 503 with an "AI not configured" message.

## Environment

| Variable | Default |
|----------|---------|
| `DATABASE_URL` | `sqlite://morph_engi.db` |
| `APP_PORT` | `9096` |
| `CORS_ORIGIN` | `http://localhost:5179` |
| `USERS_PANEL_BASE_URL` | `http://127.0.0.1:9090` (Morph auth) |
| `MORPH_AI_API_KEY` | _(empty — deterministic help only)_ |
