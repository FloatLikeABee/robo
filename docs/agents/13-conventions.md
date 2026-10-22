# 13 — Conventions

## General

1. **Minimal diffs.** Change only files required for the task. Match the app you are in.
2. **Scope to one app** unless asked to cross-cut.
3. **Verify** with that app’s build/test, not the whole monorepo by default.
4. **Environment:** one repo-root `.env`. Nested leftover `.env` files are ignored. Production, when it exists, is `deploy/.env.production`.
5. **Auth:** Dorph JWT. See `01-auth-flow.md`. Do not reintroduce a UsersPanel service.
6. **Chat:** Dorph AI is the system assistant. Do not add `@robo/platform-chat` drawers.
7. **Theme:** Product UIs stay **dark-only**. Do not add a light/dark switch.

## User-facing names vs ids

| User-facing | Implementation id (keep) |
|-------------|--------------------------|
| Dorph AI, DorphNotes | `dorph/`, `/dorphdata` |
| DorphUtils | `dorph-utils/` (no space: not “Dorph Utils”) |
| Event Logs | folder `formx/`, module id `sheetx`, embed `/events-info` (legacy `/survey-bot` may still exist) |
| Content Maker | `composerx/` |
| Data Access | `SharpReport/`, iframe `/datax` |
| Project | `dorph-engi/` |
| AI tools | `bk/` |

Do not label Event Logs as Survey Maker / SurveyX. Do not label Content Maker as ComposerX in UI copy. Do not label Data Access as DataX in UI copy. Env vars and routes may still say `sheetx`, `FORMSX_*`, `datax`, `USERS_PANEL_*`.

Booki and Academi are not current products. Do not add them to README tables or `start-all.sh`.

## Local storage

Default stack is **SQLite + Badger** under each app’s `./data/`. Go packages named `mysql` / `mongo` may wrap those files. Do not document installing MySQL, MongoDB, or Redis as required.

Relative `./data/...` paths are **cwd-relative** (the process working directory), not repo-root-relative.

## Go backends

**Dorph / Content Maker:** large `main.go` (or split files next to it) + Gin.

**Event Logs:** `formx/backend/cmd/server/main.go` + `internal/{config,handler,models,mysql,mongo}`.

Load env with `pkg/repoenv` (walk up to `start-all.sh`).

macOS Dorph API: `go build` then run (Badger `LC_UUID`); `start-all.sh` already does this.

## Rust backends

Axum + SQLx + SQLite. Data Access API port is `SHARPREPORT_PORT` — Vite/SSR on **5178** must proxy to that port, not a hardcoded 3050. Project API is `:9096`, UI `:5179`.

## Frontends

- Dorph AI / DorphNotes: React CRA on `:3031`. Extensionless imports must resolve to tracked `.js` files.
- DorphUtils / Event Logs: React Vite.
- Content Maker / Project: Svelte Vite.
- Data Access: SvelteKit.

## Config loading

Operators copy **only** the root template: `cp .env.example .env`. Empty nested `.env` files must not blank `DORPH_AI_API_KEY` or other shared keys.

## Agent docs index

| File | Topic |
|------|--------|
| `00-architecture-overview.md` | Map and ports |
| `01-auth-flow.md` | Dorph JWT SSO |
| `02-ai-integration.md` | Dorph AI + pkg/dorphai |
| `03-dorph.md` | Dorph AI + DorphNotes |
| `04-formx.md` | Event Logs |
| `05-composerx.md` | Content Maker |
| `06-ai-tools.md` | AI tools (`bk/`) |
| `07-rust-apps.md` | Data Access + Project |
| `09-dorph-utils.md` | DorphUtils shell |
| `10-shared-libraries.md` | `pkg/` |
| `12-build-deploy.md` | `start-all.sh` and deploy honesty |
