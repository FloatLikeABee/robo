# 07 — Data Access and Project (Rust)

Two Rust APIs: **Axum**, **SQLx**, SQLite, `pkg/morphai-rs`, Morph SSO (`USERS_PANEL_BASE_URL` → Morph `:9090`). Neither ships a `platform-chat` drawer.

## Data Access (`SharpReport/`)

User-facing name: **Data Access** (not DataPulse / DataX in UI copy). Iframe path in MorphUtils: `/datax`.

| | |
|--|--|
| API | `SHARPREPORT_PORT` (often 3050; Vite on **5178** must proxy to this) |
| UI | `http://localhost:5178` |
| DB | SQLite |
| Extra | Embedded Metabase (Java 17+), dark-only UI |

If MorphUtils `/datax` says connection refused, start `sharpreport-ui` (`./start-all.sh start sharpreport-ui`) and keep Vite `server.host` reachable. API up on `SHARPREPORT_PORT` is not enough if 5178 is down.

Reuse the Morph JWT cookie. Do not show a second password form. Do not clear the shared cookie on 502 from Data Access `/auth/me`.

Key files: `SharpReport/backend/src/main.rs`, `config.rs`, `api/`. Frontend is SvelteKit. More: [`SharpReport/README.md`](../../SharpReport/README.md), [`SharpReport/docs/`](../../SharpReport/docs/).

## Project (`morph-engi/`)

User-facing name: **Project** inside MorphUtils.

| | |
|--|--|
| API | `9096` |
| UI | `5179` |
| DB | SQLite (`DATABASE_URL`, default `sqlite://morph_engi.db`) |

AI project documents from files or paste, plus a files library. `POST /api/v1/projects/generate-document` needs `MORPH_AI_API_KEY`. Chat for operators is Morph AI.

**Vercel static preview** (UI only, localStorage, no Rust API): see [`morph-engi/README.md`](../../morph-engi/README.md#vercel-static-preview).

## Shared auth

See `01-auth-flow.md`. Env: `USERS_PANEL_BASE_URL=http://127.0.0.1:9090`.
