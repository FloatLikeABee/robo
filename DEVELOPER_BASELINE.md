# Developer baseline (AI + human onboarding)

This file is **baseline documentation** for the `robo` workspace: independent app projects spanning web, mobile-style, and backends. Treat it as the first stop when onboarding or when an AI assistant needs architectural context **before** making changes.

**Related:** Shared assistant API contract → [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](./AI_ASSISTANT_MORPHAI_CONTRACT.md) (FormsX, TranMail, SharpReport/DataPulse, Booki, UsersPanel).

---

## Workspace map

The root folder may contain multiple **separate repositories** or checkouts side by side. Prefer changing only the project that matches the ticket; do not refactor across apps unless explicitly asked.

| Directory | Product / purpose | Backend | Frontend / client |
|-----------|-------------------|---------|-------------------|
| `TranDemo/` | MorphData / Morph AI, chat, Tran admin APIs, forms | Go (Gin), BadgerDB, optional MySQL/Mongo/Redis | React (`frontend/`) |
| `tranform/` | FormsX (forms platform) | Go (Gin, GORM, MySQL, Mongo) — `backend/` | Vite + React (`frontend/`) |
| `tranmail/` | TranMail (templates, mail flows) | Go (Gin, MySQL, Mongo, Redis) — `backend/` | Vite + Svelte (`frontend/`) |
| `booki/` | Booki product | Go (Gin, MySQL, Redis, JWT) — `backend/` | Vite + React + Tailwind (`frontend/`) |
| `UsersPanel/` | Users, roles, auth admin | Rust (Axum, sqlx, MySQL) — `backend/` | Svelte 5 (`admin/`) |
| `SharpReport/` | DataPulse / SharpReport analytics (Metabase-backed) | Rust (Axum, sqlx) — `backend/` | Svelte 5 Kit (`frontend/`) |
| `academi/` | Academi (mobile app + API) | Go (Gin, Badger, JWT) — `backend/` | React Native (`package.json` at root) |

Always read each project’s own `README.md` (or `design.md`) when making non-trivial changes; this baseline does not replace deep docs.

---

## Conventions for AI-assisted development

1. **Minimal diffs:** Change only files required for the task; match existing naming, formatting, and patterns in that repo.
2. **Verify build for the touched project:** e.g. `go test ./...` or `cargo check` — scope to edited packages where full builds are slow.
3. **Environment:** Each app uses `.env` or shell exports differently; grep for `godotenv`, `dotenv`, or `DATABASE_URL` in that repo.
4. **Auth:** Morph / Tran integrations often assume UsersPanel JWT at `USERS_PANEL_BASE_URL`; see TranDemo README and UsersPanel README.
5. **Cross-project assistant behavior:** Align with `AI_ASSISTANT_MORPHAI_CONTRACT.md` when touching assistant endpoints or conversational state.

---

## Shared MorphAI configuration (Phase 0)

All platform assistants use the **same model stack** as Morph Data (TranDemo):

| Variable | Default | Purpose |
|----------|---------|---------|
| `MORPH_AI_API_KEY` | _(required for LLM)_ | DashScope API key |
| `MORPH_AI_MODEL` | `qwen3-max` | Chat/completion model |
| `MORPH_AI_API_URL` | DashScope text-generation URL | Native endpoint (Go/Rust clients) |
| `MORPH_AI_BASE_URL` | DashScope compatible-mode `/v1` | OpenAI-compatible clients (TranMail composer) |

Legacy fallbacks: `GEMINI_API_KEY`, `GEMINI_MODEL`, `TRAN_QWEN_*`.

**Libraries:**

| Path | Language | Use |
|------|----------|-----|
| `pkg/morphai/` | Go | DashScope client + `LoadFromEnv()` |
| `pkg/morphai-rs/` | Rust | DashScope client for SharpReport / UsersPanel |
| `platform-chat/` | TS/Svelte/React | Shared 720px chat drawer UI |

**Docs:** [`AI_ASSISTANT_SESSIONS_API.md`](./AI_ASSISTANT_SESSIONS_API.md) (session persistence spec).

Copy `TranDemo/.env.example` or each app’s `.env.example` and set `MORPH_AI_API_KEY` before enabling LLM features.

---

## TranDemo — MorphData / Morph AI

- **Role:** Primary “School Manager” style stack: AI chat, sheet/form helpers, **`/api/tran/*`** admin-style REST, forms APIs, Morph AI agents.
- **Backend:** `go.mod` module `idongivaflyinfa`; entry typically `main.go`; routes wired through `handlers/`, `handlers/register_routes.go`.
- **Frontend:** React under `TranDemo/frontend/` (proxies/API base via env; see TranDemo `README.md`).
- **Persistence:** BadgerDB (embedded), optional **MySQL** (`TRAN_MYSQL_*`), **Mongo**, **Redis** for Tran features.
- **Docs:** [`TranDemo/README.md`](./TranDemo/README.md) — ports, env table, MorphData URLs (`/morphdata`, legacy redirects).

---

## tranform — FormsX

- **Backend:** Module `github.com/formsx/backend`; server commonly under `backend/cmd/server/`.
- **Stack:** Gin, Swagger, GORM + MySQL, Mongo driver for documents.
- **Frontend:** React + Vite in `tranform/frontend/`.
- **When editing:** Respect existing handler layout under `backend/internal/` or similar project structure.

---

## tranmail — TranMail

- **Backend:** Module `mergeemailx-backend`; Gin + MySQL + Mongo + Redis; OpenAI client present for assistant-style features.
- **Frontend:** Svelte + Vite in `tranmail/frontend/`.

---

## booki — Booki

- **Backend:** Module `github.com/academi/booki`; Gin, MySQL, Redis, JWT.
- **Frontend:** React 19 + Vite + Tailwind v4 (`booki/frontend/`).
- **Design notes:** Optionally `booki/design.md`, `booki/design-db-ui.md`.

---

## UsersPanel

- **Purpose:** JWT auth, roles, OAuth; API consumed by Morph / Tran setups.
- **Backend:** Rust Axum crate `users-panel-api` in `UsersPanel/backend/`. Default listen **`127.0.0.1:5001`**; Swagger **`/swagger-ui`**.
- **Admin UI:** `UsersPanel/admin/` (Vite + Svelte).
- **Docs:** [`UsersPanel/README.md`](./UsersPanel/README.md), [`UsersPanel/docs/microservices-and-api.md`](./UsersPanel/docs/microservices-and-api.md).
- **`DATABASE_URL`:** MySQL URL form for sqlx (not Go DSN); details in README.

---

## SharpReport (DataPulse)

- **Purpose:** Analytics / Metabase integration; branded as DataPulse in README.
- **Backend:** Rust (`datapulse` crate), Axum + SQLx — `SharpReport/backend/`.
- **Frontend:** SvelteKit + Tailwind + ECharts — `SharpReport/frontend/`.
- **Prereqs:** Java for Metabase, Node 22+, Rust 1.84+ per project README.
- **Docs:** [`SharpReport/README.md`](./SharpReport/README.md), `design.md`, `docs/`.

---

## academi

- **Backend:** `academi/backend/` — Gin, BadgerDB, JWT (`github.com/academi/backend`).
- **Client:** React Native app at `academi/` (`npm start`, `react-native`).
- **Docs:** [`academi/README.md`](./academi/README.md).

---

## Quick dependency matrix

| Capability | Typical projects |
|------------|------------------|
| MySQL | TranDemo, tranform, tranmail, booki, UsersPanel, SharpReport |
| MongoDB | TranDemo, tranform, tranmail |
| Redis | TranDemo, tranmail, booki |
| Badger embedded | TranDemo, academi |
| Rust / Axum | UsersPanel, SharpReport |

---

## Suggested edits to this baseline

When you add a major feature, dependency, default port, or env var that future tickets will assume:

1. Update this file **or** the project README and add one line here pointing to it.
2. If behavior affects multiple assistants, consider updating **`AI_ASSISTANT_MORPHAI_CONTRACT.md`**.

---

_Last updated for developer baseline / AI onboarding (**ClickUp baseline doc task**)._ 
