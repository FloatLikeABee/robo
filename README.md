# robo — local development

[![CI](https://github.com/FloatLikeABee/robo/actions/workflows/ci.yml/badge.svg)](https://github.com/FloatLikeABee/robo/actions/workflows/ci.yml)

This workspace contains the Morph platform apps. Use **`start-all.sh`** to run them together in dev.

Architecture and per-app notes: [`docs/agents/00-architecture-overview.md`](./docs/agents/00-architecture-overview.md). Local run and build: [`docs/agents/12-build-deploy.md`](./docs/agents/12-build-deploy.md).

**Project static preview (Vercel, UI only):** import this repo with Framework **Other**, root `./`, Build Command `npm run vercel-build`, Output Directory `morph-engi/frontend/dist`. Details: [`morph-engi/README.md`](./morph-engi/README.md#vercel-static-preview).

The supported path is **local `start-all.sh`**. `scripts/deploy.sh` still mentions Render/Alibaba, but the `deploy/` tree is not in this repo — do not treat that script as a complete production runbook.

---

## Project folders

| Folder | Product | Stack |
|--------|---------|-------|
| `morph/` | Morph AI and MorphNotes (Tasks, Timelines, Big notes, Research, Generic data) | Go + React (CRA) |
| `morph-utils/` | MorphUtils shell (Event Logs, Content Maker, Data Access, Project) | React (Vite) |
| `invite-signup/` | Invite Signup (admin codes and user redeem) | React (Vite) |
| `formx/` | Event Logs | Go + React (Vite) |
| `composerx/` | Content Maker | Go + Svelte (Vite) |
| `morph-engi/` | Project (in MorphUtils) | Rust + Svelte (Vite) |
| `SharpReport/` | Data Access | Rust + SvelteKit |
| `bk/` | AI tools (Assistants, RAG, Documents, System) | Python FastAPI + React |
| `pkg/` | Shared libraries (`morphai`, `morphai-rs`, …) | — |
| `morphgraph-worker/` | Optional GraphRAG worker | — |

**Auth:** Morph hosts platform authentication. Other apps point `USERS_PANEL_BASE_URL` (legacy env name) at the Morph API (`http://127.0.0.1:9090`).

Chat lives in **Morph AI**. MorphUtils modules do not ship a separate assistant drawer.

Booki, Academi, and a standalone UsersPanel app are **not** part of the supported stack (even if leftover folders remain on disk).

---

## Prerequisites

Install these before running the stack:

| Tool | Used by |
|------|---------|
| **Go** 1.21+ | morph, formx, composerx |
| **Node.js** 18+ (20+ recommended) | All frontends |
| **Rust** (stable) + `cargo` | Data Access, Project |
| **Java** 17+ | Data Access (embedded Metabase) |
| **Neo4j** (optional) | Morph GraphRAG / AI graph — `start-all.sh` tries `neo4j start` on full stack boot |
| **Python** 3.11+ | AI tools API |

**No MySQL, MongoDB, or Redis.** Apps use embedded SQLite + Badger + in-process cache under each app’s `./data/` directory.

---

## One config file

Local/dev configuration is a **single repo-root `.env`**. Nested leftover `.env` files are ignored (they must not blank shared keys). Production stays `deploy/.env.production` when that tree exists.

```bash
cp .env.example .env
```

Set `MORPH_AI_API_KEY`, `USERS_PANEL_BASE_URL=http://127.0.0.1:9090`, and per-app data paths (`TRAN_SQLITE_PATH`, `COMPOSERX_SQLITE_PATH`, …) in that file. Relative `./data/...` paths stay relative to each app’s working directory.

Default Morph login for local/dev (`MORPH_ENV` unset): **`morphadmin`** / **`admin123`** (or `morphadmin@local.com`). `./start-all.sh` needs no extra auth config. Hosting sets `MORPH_ENV=production` and the overrides in [`docs/security-hosting-checklist.md`](./docs/security-hosting-checklist.md). In that mode Morph refuses to start while the development JWT secret or admin password is still in use.

**Minimum AI setup** (DashScope / Qwen by default):

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
```

After editing `.env`, restart the affected API: `./start-all.sh restart morph-api` (etc.).

**Content Maker only:** optional `TRAN_OPENAI_API_KEY` for reference-library embeddings.  
**Shared libraries:** [`pkg/morphai/`](./pkg/morphai/) (Go), [`pkg/morphai-rs/`](./pkg/morphai-rs/) (Rust). Auth and AI details: [`docs/agents/01-auth-flow.md`](./docs/agents/01-auth-flow.md), [`docs/agents/02-ai-integration.md`](./docs/agents/02-ai-integration.md).

---

## Quick start

From the repo root:

```bash
chmod +x start-all.sh   # once

./start-all.sh              # start every remaining app
./start-all.sh --install    # install deps, then start
```

Press **Ctrl+C** or run `./start-all.sh --stop` to shut everything down.

Clone → `cp .env.example .env` → `./start-all.sh`. Data lives under each app’s `./data/` (SQLite + Badger). Optional **Neo4j** is only for AI graph / GraphRAG. Morph AI Skills: header **Skills** after login.

---

## URLs (default ports)

| Service | URL |
|---------|-----|
| Morph API | http://localhost:9090 |
| Morph AI / MorphNotes UI | http://localhost:3031 |
| MorphUtils | http://localhost:3040 |
| Invite Signup | http://localhost:3051 |
| AI tools API | http://localhost:8000/docs |
| AI tools UI | http://localhost:3000 |
| Event Logs API | http://localhost:29909/swagger/index.html |
| Event Logs UI | http://localhost:19909 |
| Content Maker API | http://localhost:8043/health |
| Content Maker UI | http://localhost:8044 |
| Project API | http://127.0.0.1:9096/health |
| Project UI | http://localhost:5179 |
| Data Access API | http://127.0.0.1:3050 |
| Data Access UI | http://localhost:5178 |

Data Access API port follows `SHARPREPORT_PORT` in `.env` (Vite on 5178 proxies to it).

---

## `start-all.sh` commands

```bash
./start-all.sh              # start all apps (foreground; Ctrl+C stops all)
./start-all.sh --install    # npm install / go mod download / cargo fetch, then start
./start-all.sh start        # start all (one-shot; skips already running)
./start-all.sh stop         # stop all
./start-all.sh restart      # stop all, then start all again
./start-all.sh status       # show running / stopped for each service
./start-all.sh list         # print service names and aliases
./start-all.sh help         # short usage summary
```

### One app at a time

```bash
./start-all.sh start <service>
./start-all.sh stop <service>
./start-all.sh restart <service>
./start-all.sh logs <service>
```

**Examples:**

```bash
./start-all.sh restart morph-api
./start-all.sh restart formx-ui
./start-all.sh restart morph
./start-all.sh restart composerx
./start-all.sh logs sharpreport-api
```

### Service names

| API | UI | Alias *(restarts both)* |
|-----|----|-------------------------|
| `morph-api` | `morph-ui` | `morph` |
| — | `morph-utils-ui` | `morph-utils` |
| — | `invite-signup-ui` | — |
| `bk-api` | `bk-ui` | `bk` |
| `formx-api` | `formx-ui` | `formx` |
| `composerx-api` | `composerx-ui` | `composerx` |
| `morph-engi-api` | `morph-engi-ui` | `morph-engi`, `engi` |
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

When starting all apps, bring up **Morph API** (auth) first so other apps can resolve sessions. Individual `start` / `restart` does not enforce order — start `morph-api` before apps that need auth if you bring them up one by one.

### Neo4j (Morph GraphRAG)

On full-stack `start` / `restart` / default `./start-all.sh`, the launcher checks bolt port **7687** and runs `neo4j start` when the CLI is installed. Install with `brew install neo4j` on macOS. Graph features are optional — other apps start even if Neo4j is missing. Ops notes: [`docs/MORPH_GRAPH_OPS.md`](./docs/MORPH_GRAPH_OPS.md).

### Port already in use

```bash
./start-all.sh status
./start-all.sh logs <service>
lsof -i :<port>
```

Then `./start-all.sh restart <service>` after freeing the port.

---

## Related docs

- [`docs/agents/00-architecture-overview.md`](./docs/agents/00-architecture-overview.md) — live app map
- [`docs/agents/12-build-deploy.md`](./docs/agents/12-build-deploy.md) — run, build, honest deploy status
- [`docs/agents/13-conventions.md`](./docs/agents/13-conventions.md) — names, dark-only, one `.env`
- Per-app READMEs: [`morph/`](./morph/README.md), [`morph-utils/`](./morph-utils/README.md), [`formx/`](./formx/README.md), [`composerx/backend/`](./composerx/backend/README.md), [`SharpReport/`](./SharpReport/README.md), [`morph-engi/`](./morph-engi/README.md), [`bk/`](./bk/README.md)
