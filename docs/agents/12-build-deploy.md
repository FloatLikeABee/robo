# 12 — Build and deploy

## Local development (supported)

Source of truth: **`start-all.sh`**. Remaining services: morph, formx, composerx, morph-engi, bk, SharpReport, morph-utils.

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

Logs: `.robo-dev/logs/<service>.log`. macOS Morph API is built before run.

Start **Morph API** first when bringing apps up one by one (auth hub).

### Environment

- Local: one gitignored `.env` at the **repo root**
- Nested leftover `.env` files are ignored
- Production path remains `deploy/.env.production` **when that tree exists** (it is currently absent)

## Per-app build (without the launcher)

```bash
cd morph && go build -o morph-server main.go
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

Tests: `go test ./...` in morph / formx/backend / composerx/backend; `cargo test` in the Rust backends.

## Production / cloud

**Not currently documented as a complete path.** `scripts/deploy.sh` still talks about Render and Alibaba and still lists removed apps, but **`deploy/` is missing** (no `render.yaml`, no `DEPLOY-README.md`). Do not follow that script as an operator runbook until `deploy/` is restored.

### What does exist

- **Local:** `start-all.sh` (this file + root README).
- **Project UI static preview on Vercel:** [`morph-engi/README.md`](../../morph-engi/README.md#vercel-static-preview) (`npm run vercel-build` → `morph-engi/frontend/dist`). No Morph API, no login, data in localStorage.
- **Data Access Docker notes:** [`SharpReport/docs/DEPLOYMENT.md`](../../SharpReport/docs/DEPLOYMENT.md) if you are packaging that app alone.

Optional GraphRAG: [`docs/MORPH_GRAPH_OPS.md`](../MORPH_GRAPH_OPS.md).
