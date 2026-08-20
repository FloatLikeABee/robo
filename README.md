# robo — local development

This workspace contains multiple independent platform apps. Use **`start-all.sh`** to run them together in dev.

For architecture and per-app deep docs, see [`DEVELOPER_BASELINE.md`](./DEVELOPER_BASELINE.md).

**Production / staging (Render or Alibaba Cloud):** see [`DEPLOY-README.md`](./DEPLOY-README.md) and run [`scripts/deploy.sh`](./scripts/deploy.sh).

**Vercel static preview (Projects UI only):** import this repo on Vercel with Framework **Other**, root `./`, then set Build Command to `npm run vercel-build` and Output Directory to `morph-engi/frontend/dist`. Details: [`morph-engi/README.md`](./morph-engi/README.md#vercel-static-preview).

---

## Project folders

| Folder | Product | Stack |
|--------|---------|-------|
| `morph/` | Morph AI / MorphData | Go + React |
| `morph-utils/` | Morph Utils (SheetX, ComposerX, DataX, Booki, Projects, Academi) | React (Vite) |
| `formx/` | FormsX / SheetX | Go + React (Vite) |
| `composerx/` | ComposerX (TranMail) | Go + Svelte (Vite) |
| `booki/` | Morph Booki | Go + React (Vite) |
| `morph-engi/` | Projects / Morph Engi (in Morph Utils) | Rust + Svelte (Vite) |
| `UsersPanel/` | Users, roles, auth | Rust + Svelte admin |
| `SharpReport/` | DataPulse / SharpReport | Rust + SvelteKit |
| `academi/` | Academi (study assistant; in Morph Utils) | Go API + static web (+ React Native) |
| `bk/` | BK / Ground Control (RAG agent workspace) | Python FastAPI + React |
| `pkg/` | Shared libraries (`morphai`, etc.) | — |
| `platform-chat/` | Shared chat drawer UI | — |
| `morph-broadcast/` | Shared broadcast email composer (legacy; removed from Utils) | — |

**Legacy names** (still referenced in some docs): `TranDemo` → `morph`, `tranform` → `formx`, `tranmail` → `composerx`.

---

## Prerequisites

Install these before running the stack:

| Tool | Used by |
|------|---------|
| **Go** 1.21+ | morph, formx, composerx, booki |
| **Node.js** 18+ (20+ recommended) | All frontends |
| **Rust** (stable) + `cargo` | UsersPanel, SharpReport, morph-engi |
| **Java** 17+ | SharpReport (embedded Metabase) |
| **Neo4j** (optional) | Morph GraphRAG / AI graph — `start-all.sh` tries `neo4j start` on full stack boot |
| **Python** 3.11+ (optional) | BK / Ground Control API |

**No MySQL, MongoDB, or Redis.** Morph AI, Morph Data, ComposerX, SheetX/FormsX, and related Utils backends use embedded SQLite + Badger + in-process cache under each app’s `./data/` directory.

Each app may have its own `.env` (copy from `.env.example` in that project). **Morph API (`:9090`) is the auth source** — other apps set `USERS_PANEL_BASE_URL=http://127.0.0.1:9090`.

Default Morph login: **`morphadmin`** / **`admin123`** (or `morphadmin@local.com`).

---

## Quick start

From the repo root:

```bash
chmod +x start-all.sh   # once

# Start every app (backends + frontends)
./start-all.sh

# Install deps first, then start
./start-all.sh --install
```

Press **Ctrl+C** or run `./start-all.sh --stop` to shut everything down.

### Standalone data (no MySQL / Mongo / Redis)

Clone → copy `.env.example` → `.env` → `./start-all.sh`. Relational data and documents live under each app’s `./data/` (SQLite + Badger). Optional **Neo4j** is only for AI graph / GraphRAG; uploads and skills enqueue async Neo4j ingest when Neo4j is configured. Morph AI skills UI: `/skills` after login.

**Smoke checklist (this change):** Morph login without ports 3306/27017/6379; ComposerX `/publishes/history` + FormsX `/api/v1/forms` with Morph JWT; `GET /api/skills` returns builtins; `GET /api/admin/neo4j-ingest/status` reachable for admin.

---

## URLs (default ports)

| Service | URL |
|---------|-----|
| UsersPanel API | http://127.0.0.1:5001/swagger-ui |
| UsersPanel Admin | http://localhost:5173 |
| Morph API | http://localhost:9090 |
| Morph UI | http://localhost:3031 |
| Morph Utils | http://localhost:3040 |
| BK API | http://localhost:8000/docs |
| BK UI | http://localhost:3000 |
| FormsX API | http://localhost:29909/swagger/index.html |
| FormsX UI | http://localhost:19909 |
| ComposerX API | http://localhost:8043/health |
| ComposerX UI | http://localhost:8044 |
| Booki API | http://127.0.0.1:9095/health |
| Morph Booki UI | http://localhost:5174 |
| Morph Engi API | http://127.0.0.1:9096/health |
| Morph Engi UI | http://localhost:5179 |
| SharpReport API | http://127.0.0.1:3050 |
| SharpReport UI | http://localhost:5178 |

---

## `start-all.sh` commands

### Start / stop / restart everything

Includes the default stack (academi excluded unless you start it explicitly).

```bash
./start-all.sh              # start all apps (foreground; Ctrl+C stops all)
./start-all.sh --install    # npm install / go mod download / cargo fetch, then start
./start-all.sh start        # start all (one-shot; skips already running)
./start-all.sh stop         # stop all  (same as --stop)
./start-all.sh restart      # stop all, then start all again
./start-all.sh status       # show running / stopped for each service
./start-all.sh list         # print service names and aliases
./start-all.sh help         # short usage summary
```

`start all` / `stop all` / `restart all` work the same as the no-arg forms above.

### One app at a time

```bash
./start-all.sh start <service>
./start-all.sh stop <service>
./start-all.sh restart <service>
./start-all.sh logs <service>    # tail -f that app's log
```

**Examples:**

```bash
# Restart only the Morph backend after a code change
./start-all.sh restart morph-api

# Restart FormsX frontend only
./start-all.sh restart formx-ui

# Restart both API + UI for an app (alias)
./start-all.sh restart morph
./start-all.sh restart composerx

# Watch logs for a failing service
./start-all.sh logs sharpreport-api
```

### Service names

| API | UI | Alias *(restarts both)* |
|-----|----|-------------------------|
| `userspanel-api` | `userspanel-admin` | `userspanel` |
| `morph-api` | `morph-ui` | `morph` |
| — | `morph-utils-ui` | `morph-utils` |
| `bk-api` | `bk-ui` | `bk` |
| `formx-api` | `formx-ui` | `formx` |
| `composerx-api` | `composerx-ui` | `composerx` |
| `booki-api` | `booki-ui` | `booki` |
| `morph-engi-api` | `morph-engi-ui` | `morph-engi`, `engi` |
| `academi-api` | `academi-ui` | `academi` |
| `sharpreport-api` | `sharpreport-ui` | `sharpreport` |

You can pass multiple services: `./start-all.sh restart morph-api formx-ui`.

---

## Logs and state

| Path | Purpose |
|------|---------|
| `.robo-dev/logs/<service>.log` | stdout/stderr for each service |
| `.robo-dev/pids` | PID file used by the script |
| `.robo-dev/morph-server` | Built Morph binary on macOS |

Logs are appended on each start. Use `./start-all.sh logs <service>` to follow them live.

---

## Notes

### macOS + Morph backend

On macOS, `morph-api` is **built** before run (`go build`) to avoid a known BadgerDB / `LC_UUID` issue with `go run`. Restarts rebuild automatically.

### Start order

When starting all apps, **UsersPanel API** starts first. Other backends and frontends follow. Individual `start` / `restart` does not enforce order — start `userspanel-api` before apps that need auth if you bring them up one by one.

### Neo4j (Morph GraphRAG)

On full-stack `start` / `restart` / default `./start-all.sh`, the launcher checks bolt port **7687** and runs `neo4j start` when the CLI is installed. Install with `brew install neo4j` on macOS. Graph features are optional — other apps start even if Neo4j is missing.

### academi

The **academi** app (React Native + Go API) is intentionally excluded. Run it from `academi/` per [`academi/README.md`](./academi/README.md).

### Port already in use

If a service fails to bind its port:

```bash
./start-all.sh status
./start-all.sh logs <service>
lsof -i :<port>    # find conflicting process
```

Then `./start-all.sh restart <service>` after freeing the port.

### Environment files

The script loads `.env` from each app's directory when present (e.g. `morph/.env`, `formx/backend/.env`). Configure embedded paths (`TRAN_SQLITE_PATH`, `COMPOSERX_SQLITE_PATH`, …), `MORPH_AI_API_KEY`, and `USERS_PANEL_BASE_URL=http://127.0.0.1:9090` — see each project's `.env.example`.

---

## AI provider configuration

All platform assistants (except **morph** and **academi**, which have their own stacks) use the shared **MorphAI** provider — [Alibaba DashScope](https://dashscope.aliyun.com/) with Qwen models by default.

| App | Config file | README section |
|-----|-------------|----------------|
| FormsX | `formx/backend/.env` | [`formx/README.md`](./formx/README.md#ai-provider-configuration) |
| ComposerX | `composerx/backend/.env` (+ optional `ai.config.json`) | [`composerx/backend/README.md`](./composerx/backend/README.md#ai-provider-configuration) |
| Booki | `booki/.env` | [`booki/README.md`](./booki/README.md#ai-provider-configuration) |
| UsersPanel | `UsersPanel/backend/.env` | [`UsersPanel/README.md`](./UsersPanel/README.md#ai-provider-configuration) |
| SharpReport | `SharpReport/.env` | [`SharpReport/README.md`](./SharpReport/README.md#ai-provider-configuration) |

**Minimum setup** (same key can be reused across apps in local dev):

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
```

After editing `.env`, restart that app's API: `./start-all.sh restart formx-api` (etc.).

**ComposerX only:** optional `TRAN_OPENAI_API_KEY` for reference-library embeddings (RAG).  
**Shared libraries:** [`pkg/morphai/`](./pkg/morphai/) (Go), [`pkg/morphai-rs/`](./pkg/morphai-rs/) (Rust).  
**Contract:** [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](./AI_ASSISTANT_MORPHAI_CONTRACT.md).

---

## Related docs

- [`DEVELOPER_BASELINE.md`](./DEVELOPER_BASELINE.md) — workspace map, shared MorphAI config, conventions
- [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](./AI_ASSISTANT_MORPHAI_CONTRACT.md) — assistant API contract across apps
- Per-app READMEs: `morph/`, `formx/`, `composerx/`, `booki/`, `UsersPanel/`, `SharpReport/`, `academi/`
