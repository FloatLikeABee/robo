# Morph Engi

Simple **project management** for civil / construction / field teams — **Rust** API + **Svelte** UI.

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

- **Projects** — create jobs, track status, daily site logs, or let AI draft them from description files
- **Files** — drawings/docs via URL or upload, per project
- **People** — contractors and contacts linked to a project
- **Settings** — organization basics

## AI assistant

`POST /api/v1/assistant/chat` — MorphAI tool loop for projects, site logs, files, and people. Set `MORPH_AI_API_KEY` in `.env`.

## AI project import

Upload description files and let AI propose projects, then review before anything is
created. Available as the **AI import** tab in the Projects module.

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/project-imports/analyze` | Multipart: up to 3 `files` parts plus optional `instruction`. Stores a draft plan; creates no records |
| `GET /api/v1/project-imports` | List import sessions for the organization |
| `GET /api/v1/project-imports/:id` | One session with its draft plan and per-file read results |
| `POST /api/v1/project-imports/:id/confirm` | Create the projects and nested records from an edited draft |

Limits and behavior:
- At most **3 files** per import, each up to 8 MB. Accepted types: **PDF, TXT, CSV, MD**.
  A fourth file or an unsupported type is rejected before any AI call.
- A draft can propose **several projects**, each with nested site logs, people, and
  flow-log entries. Confirm accepts an edited plan and can exclude any project or
  individual nested record.
- Confirm creates each project and its children in one transaction per project.
  Records that fail validation are reported and skipped while their valid siblings
  are still created. A session already completed refuses a second confirm, so no
  duplicates are created.
- Flow-log entries have no project column, so a created entry carries the project
  code as a tag.
- Sessions are scoped by organization; another org's session reads as not found.
- Without `MORPH_AI_API_KEY`, analyze returns HTTP 503 with an "AI not configured" message.

## Environment

| Variable | Default |
|----------|---------|
| `DATABASE_URL` | `sqlite://morph_engi.db` |
| `APP_PORT` | `9096` |
| `CORS_ORIGIN` | `http://localhost:5179` |
| `USERS_PANEL_BASE_URL` | `http://127.0.0.1:9090` (Morph auth) |
| `MORPH_AI_API_KEY` | _(empty — deterministic help only)_ |
