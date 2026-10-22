# 12 — Build and deploy

## Local development (supported)

Source of truth: **`start-all.sh`**. Remaining services: dorph, formx, composerx, dorph-engi, bk, SharpReport, dorph-utils, invite-signup (`invite-signup-ui` at http://localhost:3051; no separate API process and no alias).

### Prerequisites

| Tool | Used by |
|------|---------|
| Go 1.21+ | dorph, formx, composerx |
| Node.js 18+ (20+ recommended) | All frontends |
| Rust (stable) + cargo | Data Access, Project |
| Java 17+ | Data Access (Metabase) |
| Neo4j (optional) | GraphRAG — `start-all.sh` may run `neo4j start` |
| Python 3.11+ | AI tools |

Do **not** install MySQL, MongoDB, or Redis for the default stack. Data is SQLite + Badger under `./data/`.

### Quick start

```bash
cp .env.example .env
./start-all.sh --install
./start-all.sh
```

```bash
./start-all.sh start dorph-api
./start-all.sh restart formx
./start-all.sh logs sharpreport-api
./start-all.sh status
./start-all.sh list
```

### Service names

| API | UI | Alias |
|-----|----|--------|
| `dorph-api` | `dorph-ui` | `dorph` |
| `formx-api` | `formx-ui` | `formx` |
| `composerx-api` | `composerx-ui` | `composerx` |
| `dorph-engi-api` | `dorph-engi-ui` | `dorph-engi` / `engi` |
| `sharpreport-api` | `sharpreport-ui` | `sharpreport` |
| `bk-api` | `bk-ui` | `bk` |
| — | `dorph-utils-ui` | `dorph-utils` |
| — | `invite-signup-ui` | — |

Logs: `.robo-dev/logs/<service>.log`. macOS Dorph API is built before run.

`stop` and `restart` kill the process group recorded for that service (the recorded process plus `go run`, cargo, or npm children), then wait until its port is free before the next start. Bash job control creates the group, so macOS does not need a `setsid` binary. The launcher never signals PID 0, PID 1, or an empty PID, and it group-kills only PIDs it recorded. A listener it did not start is stopped as that process tree only. `restart dorph-api` waits until `/health` succeeds. Service names match exactly. If `lsof` is missing, the launcher warns that it cannot prove the port is free.

Start **Dorph API** first when bringing apps up one by one (auth hub).

### Environment

- Local: one gitignored `.env` at the **repo root**
- Nested leftover `.env` files are ignored
- Hosted Dorph: copy `deploy/.env.production.example` to gitignored `deploy/.env.production`. Runbook: [`deploy/README.md`](../../deploy/README.md).

## Per-app build (without the launcher)

```bash
cd dorph && go build -o dorph-server main.go
cd dorph && go build -o dorph-mcp ./cmd/dorph-mcp   # stdio MCP; see docs/agents/14-dorph-mcp.md
cd formx/backend && go build -o formsx-server ./cmd/server
cd composerx/backend && go build -o composerx-server .
cd SharpReport/backend && cargo build --release
cd dorph-engi/backend && cargo build --release

cd dorph/frontend && npm run build
cd formx/frontend && npm run build
cd dorph-utils/frontend && npm run build
cd composerx/frontend && npm run build
cd dorph-engi/frontend && npm run build
cd SharpReport/frontend && npm run build
cd bk/frontend && npm run build
```

Tests: `go vet ./...` and `go test ./...` in dorph, formx/backend, composerx/backend, and pkg/dorphai (see CI below); `cargo test` in the Rust backends.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on every pull request to `main` and on every push to `main`. Jobs run in parallel. A newer run on the same pull request cancels the one it replaces. Runs for different pull requests do not cancel each other.

Every check below runs on every pull request. There are no path filters: `dorph`, `formx/backend`, and `composerx/backend` `replace` sibling packages, so a `pkg/dorphai` edit can break an app whose own directory did not change. The five check names are stable (they are what branch protection should require). Do not rename them in the workflow without updating that protection.

There is no `go.work`. Each Go module is tested from its own directory, using the `go` version in that module's `go.mod`. `actions/setup-go` caches modules and build outputs, keyed by that module's `go.sum` (`go.mod` for `pkg/dorphai`, which has no third-party requirements). `DORPH_AI_API_KEY` is unset; nothing in these jobs should call an AI provider.

| Check | Directory | Commands |
|-------|-----------|----------|
| Go / Dorph API | `dorph` | `go vet ./...` then `go test ./...` |
| Go / Event Logs | `formx/backend` | `go vet ./...` then `go test ./...` |
| Go / Content Maker | `composerx/backend` | `go vet ./...` then `go test ./...` |
| Go / dorphai | `pkg/dorphai` | `go vet ./...` then `go test ./...` |
| Dorph frontend | `dorph/frontend` | `python3 openspec/check_files_workspace_archive.py` once from the repo root, then `npm ci`, then `CI=true npm test -- --watchAll=false`, then `CI=true npm run build` |

The frontend job uses Node.js 22 (the repo `engines.node` is `>=20`) and caches npm from `dorph/frontend/package-lock.json`.

Reproduce locally from the repo root. Leave `DORPH_AI_API_KEY` unset (do not export a key, and do not rely on a root `.env` for these commands):

```bash
unset DORPH_AI_API_KEY GEMINI_API_KEY TRAN_OPENAI_API_KEY OPENAI_API_KEY

( cd dorph && go vet ./... && go test ./... )
( cd formx/backend && go vet ./... && go test ./... )
( cd composerx/backend && go vet ./... && go test ./... )
( cd pkg/dorphai && go vet ./... && go test ./... )

python3 openspec/check_files_workspace_archive.py
( cd dorph/frontend && npm ci && CI=true npm test -- --watchAll=false && CI=true npm run build )
```

The Dorph image build is a separate workflow (`.github/workflows/docker-image.yml`). It is not one of the five check names above. `sh deploy/check-container-contract.sh` is the fast local check; `docker build -t dorph:local .` from the repo root builds the image. That workflow also validates `render.yaml` against the Render Blueprint schema and runs `sh SharpReport/deploy/check-container-contract.sh`. It does not add a check name to `ci.yml`.

## Production / cloud

Dorph API and the Dorph AI UI ship as one image. Content Maker ships as `composerx/Dockerfile` (API and UI on port 8043). The runbook is [`deploy/README.md`](../../deploy/README.md): local compose, and **Deploy on Render** for the hosted services (`render.yaml` at the repo root). Content Maker image details are in [`composerx/backend/README.md`](../../composerx/backend/README.md). The product owner creates services from the Blueprint after merge. The ordered create, URL record, and Dorph image rebuild are under `DorphUtils stack on Render` in [`deploy/README.md`](../../deploy/README.md). `render.yaml` still lists `REACT_APP_DORPH_UTILS_URL` on `dorph` with `sync: false` and no value (leftover image contract). Dorph AI does not show a DorphUtils header chip. `scripts/deploy.sh` still talks about Alibaba and removed apps. Do not follow that script.

### What does exist

- **Local:** `start-all.sh` (this file + root README).
- **Project UI static preview on Vercel:** [`dorph-engi/README.md`](../../dorph-engi/README.md#vercel-static-preview) (`npm run vercel-build` → `dorph-engi/frontend/dist`). No Dorph API, no login, data in localStorage.
- **Data Access image:** [`SharpReport/README.md`](../../SharpReport/README.md) and the Data Access section of [`deploy/README.md`](../../deploy/README.md). `docker build -t sharpreport:local -f SharpReport/Dockerfile .` from the repo root. Do not follow `scripts/deploy.sh`.

Optional GraphRAG: [`docs/DORPH_GRAPH_OPS.md`](../DORPH_GRAPH_OPS.md).
