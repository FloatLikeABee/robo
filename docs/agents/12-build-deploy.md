# 12 — Build and deploy

## Local development (supported)

Source of truth: **`start-all.sh`**. Remaining services: morph, formx, composerx, morph-engi, bk, SharpReport, morph-utils, invite-signup (`invite-signup-ui` at http://localhost:3051; no separate API process and no alias).

### Prerequisites

| Tool | Used by |
|------|---------|
| Go 1.21+ | morph, formx, composerx |
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
./start-all.sh start morph-api
./start-all.sh restart formx
./start-all.sh logs sharpreport-api
./start-all.sh status
./start-all.sh list
```

### Service names

| API | UI | Alias |
|-----|----|--------|
| `morph-api` | `morph-ui` | `morph` |
| `formx-api` | `formx-ui` | `formx` |
| `composerx-api` | `composerx-ui` | `composerx` |
| `morph-engi-api` | `morph-engi-ui` | `morph-engi` / `engi` |
| `sharpreport-api` | `sharpreport-ui` | `sharpreport` |
| `bk-api` | `bk-ui` | `bk` |
| — | `morph-utils-ui` | `morph-utils` |
| — | `invite-signup-ui` | — |

Logs: `.robo-dev/logs/<service>.log`. macOS Morph API is built before run.

Start **Morph API** first when bringing apps up one by one (auth hub).

### Environment

- Local: one gitignored `.env` at the **repo root**
- Nested leftover `.env` files are ignored
- Production path remains `deploy/.env.production` **when that tree exists** (it is currently absent)
- Sealed secrets: `MORPH_SECRETS_KEY`, generated with `openssl rand -base64 32`. Optional `MORPH_SECRETS_KEY_PREVIOUS` is decrypt-only during rotation and is loaded when set. The key is not `JWT_SECRET`. Local dev (`MORPH_ENV` unset, or `development`/`dev`/`local`/`test`) can leave the current key unset; startup then uses or creates `morph-secrets.key` (mode `0600`) next to the Tran database and logs a path-only warning. `MORPH_ENV=production` or `prod` refuses a missing or invalid key. The error names the variable and does not print the value. See [`14-secrets-key.md`](14-secrets-key.md).

## Per-app build (without the launcher)

```bash
cd morph && go build -o morph-server main.go
cd morph && go build -o morph-mcp ./cmd/morph-mcp   # stdio MCP; see docs/agents/14-morph-mcp.md
cd formx/backend && go build -o formsx-server ./cmd/server
cd composerx/backend && go build -o composerx-server .
cd SharpReport/backend && cargo build --release
cd morph-engi/backend && cargo build --release

cd morph/frontend && npm run build
cd formx/frontend && npm run build
cd morph-utils/frontend && npm run build
cd composerx/frontend && npm run build
cd morph-engi/frontend && npm run build
cd SharpReport/frontend && npm run build
cd bk/frontend && npm run build
```

Tests: `go vet ./...` and `go test ./...` in morph, formx/backend, composerx/backend, and pkg/morphai (see CI below); `cargo test` in the Rust backends.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on every pull request to `main` and on every push to `main`. Jobs run in parallel. A newer run on the same pull request cancels the one it replaces. Runs for different pull requests do not cancel each other.

Every check below runs on every pull request. There are no path filters: `morph`, `formx/backend`, and `composerx/backend` `replace` sibling packages, so a `pkg/morphai` edit can break an app whose own directory did not change. The five check names are stable (they are what branch protection should require). Do not rename them in the workflow without updating that protection.

There is no `go.work`. Each Go module is tested from its own directory, using the `go` version in that module's `go.mod`. `actions/setup-go` caches modules and build outputs, keyed by that module's `go.sum` (`go.mod` for `pkg/morphai`, which has no third-party requirements). `MORPH_AI_API_KEY` is unset; nothing in these jobs should call an AI provider.

| Check | Directory | Commands |
|-------|-----------|----------|
| Go / Morph API | `morph` | `go vet ./...` then `go test ./...` |
| Go / Event Logs | `formx/backend` | `go vet ./...` then `go test ./...` |
| Go / Content Maker | `composerx/backend` | `go vet ./...` then `go test ./...` |
| Go / morphai | `pkg/morphai` | `go vet ./...` then `go test ./...` |
| Morph frontend | `morph/frontend` | `npm ci`, then `CI=true npm test -- --watchAll=false`, then `CI=true npm run build` |

The frontend job uses Node.js 22 (the repo `engines.node` is `>=20`) and caches npm from `morph/frontend/package-lock.json`.

Reproduce locally from the repo root. Leave `MORPH_AI_API_KEY` unset (do not export a key, and do not rely on a root `.env` for these commands):

```bash
unset MORPH_AI_API_KEY GEMINI_API_KEY TRAN_OPENAI_API_KEY OPENAI_API_KEY

( cd morph && go vet ./... && go test ./... )
( cd formx/backend && go vet ./... && go test ./... )
( cd composerx/backend && go vet ./... && go test ./... )
( cd pkg/morphai && go vet ./... && go test ./... )

( cd morph/frontend && npm ci && CI=true npm test -- --watchAll=false && CI=true npm run build )
```

## Production / cloud

**Not currently documented as a complete path.** `scripts/deploy.sh` still talks about Render and Alibaba and still lists removed apps, but **`deploy/` is missing** (no `render.yaml`, no `DEPLOY-README.md`). Do not follow that script as an operator runbook until `deploy/` is restored.

### What does exist

- **Local:** `start-all.sh` (this file + root README).
- **Project UI static preview on Vercel:** [`morph-engi/README.md`](../../morph-engi/README.md#vercel-static-preview) (`npm run vercel-build` → `morph-engi/frontend/dist`). No Morph API, no login, data in localStorage.
- **Data Access Docker notes:** [`SharpReport/docs/DEPLOYMENT.md`](../../SharpReport/docs/DEPLOYMENT.md) if you are packaging that app alone.

Optional GraphRAG: [`docs/MORPH_GRAPH_OPS.md`](../MORPH_GRAPH_OPS.md).
