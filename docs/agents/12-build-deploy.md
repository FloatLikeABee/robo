# 12 — Build & Deploy

## Local development

### Prerequisites

| Tool | Used by |
|------|---------|
| Go 1.21+ | morph, formx, composerx, booki, academi |
| Node.js 18+ (20+ recommended) | All frontends |
| Rust (stable) + cargo | UsersPanel, SharpReport, morph-engi |
| MySQL 8 | Most backends |
| MongoDB | morph, formx, composerx |
| Redis | morph, composerx, booki |
| Java 17+ | SharpReport (embedded Metabase) |
| Neo4j (optional) | morph GraphRAG — started by `start-all.sh` when CLI is available |
| Python 3.11+ | bk (Ground Control API) |

### Quick start

```bash
# Start all apps (foreground; Ctrl+C stops all)
./start-all.sh

# Install dependencies first, then start
./start-all.sh --install

# Start/stop/restart individual services
./start-all.sh start morph-api
./start-all.sh stop formx-ui
./start-all.sh restart composerx        # alias restarts both API + UI
./start-all.sh logs sharpreport-api     # tail -f logs

# Status and listing
./start-all.sh status
./start-all.sh list
```

### Service names

| API | UI | Alias (restarts both) |
|-----|----|-----------------------|
| `morph-api` | `morph-ui` | `morph` |
| `formx-api` | `formx-ui` | `formx` |
| `composerx-api` | `composerx-ui` | `composerx` |
| `booki-api` | `booki-ui` | `booki` |
| `morph-engi-api` | `morph-engi-ui` | `morph-engi` / `engi` |
| `userspanel-api` | `userspanel-admin` | `userspanel` |
| `academi-api` | `academi-ui` | `academi` |
| `sharpreport-api` | `sharpreport-ui` | `sharpreport` |
| `bk-api` | `bk-ui` | `bk` |
| — | `morph-utils-ui` | `morph-utils` |

### Logs and state

| Path | Purpose |
|------|---------|
| `.robo-dev/logs/<service>.log` | stdout/stderr per service (appended) |
| `.robo-dev/pids` | PID file |
| `.robo-dev/morph-server` | Built Morph binary (macOS) |

### macOS: Morph backend

On macOS, `morph-api` is **built before run** (`go build`) to avoid BadgerDB `LC_UUID` issue with `go run`. `start-all.sh` handles this automatically.

### Start order

When starting all apps, **UsersPanel API** starts first. Other backends and frontends follow. Individual `start`/`restart` does not enforce order.

**Neo4j:** full-stack start/restart calls `ensure_neo4j` — checks port 7687 and runs `neo4j start` when the CLI exists (non-fatal if missing).

### Environment files

- `.env` files are gitignored (except `.env.example`)
- `start-all.sh` loads `.env` from each app's directory (parent first, then app dir)
- Copy `.env.example` to `.env` and configure MySQL DSNs, `MORPH_AI_API_KEY`, `USERS_PANEL_BASE_URL`

---

## Per-app build commands

### Go backends

```bash
# morph
cd morph && go build -o morph-server main.go

# formx
cd formx/backend && go build -o formsx-server ./cmd/server

# composerx
cd composerx/backend && go build -o composerx-server .

# booki
cd booki/backend && go build -o booki-server ./cmd/server

# academi
cd academi/backend && go build -o academi-server ./cmd/main.go
```

### Rust backends

```bash
# UsersPanel
cd UsersPanel/backend && cargo build --release

# SharpReport
cd SharpReport/backend && cargo build --release

# morph-engi
cd morph-engi/backend && cargo build --release
```

### Frontends

```bash
# React (Vite)
cd formx/frontend && npm run build     # → dist/
cd booki/frontend && npm run build     # → dist/
cd morph-utils/frontend && npm run build  # → dist/

# React (CRA)
cd morph/frontend && npm run build     # → build/

# Svelte (Vite)
cd composerx/frontend && npm run build # → dist/
cd morph-engi/frontend && npm run build  # → dist/
cd UsersPanel/admin && npm run build   # → dist/

# SvelteKit
cd SharpReport/frontend && npm run build
```

---

## Testing

```bash
# Go
cd morph && go test ./...
cd formx/backend && go test ./...
cd composerx/backend && go test ./...
cd booki/backend && go test ./...

# Rust
cd UsersPanel/backend && cargo test
cd SharpReport/backend && cargo test
cd morph-engi/backend && cargo test

# Frontend (varies by app — check package.json scripts)
cd formx/frontend && npm test
```

---

## Deployment

### Entry point

```bash
./scripts/deploy.sh help
```

### Key commands

```bash
./scripts/deploy.sh doctor              # Check toolchains
./scripts/deploy.sh build --all         # Build all release artifacts
./scripts/deploy.sh build --app=morph,formx  # Build specific apps
./scripts/deploy.sh package             # Create tarball under dist/deploy/
./scripts/deploy.sh docker build --all  # Build all Docker images
./scripts/deploy.sh render validate     # Validate Render blueprint
./scripts/deploy.sh render apply        # Apply to Render
./scripts/deploy.sh alibaba package     # Package for Alibaba Cloud
./scripts/deploy.sh alibaba sync --host=user@ecs --path=/opt/robo
```

### Docker images

Dockerfiles in `deploy/docker/`:

| Dockerfile | App |
|-----------|-----|
| `morph.Dockerfile` | Morph |
| `formx.Dockerfile` | FormsX |
| `composerx.Dockerfile` | ComposerX |
| `booki.Dockerfile` | Booki |
| `morph-engi.Dockerfile` | Morph Engi |
| `userspanel.Dockerfile` | UsersPanel |
| `sharpreport.Dockerfile` | SharpReport |

Build context = **repo root** (because of `pkg/morphai` replace directives).

### Render

Blueprint: `deploy/render.yaml`. Each app has a web service (Docker) + static site (SPA). Requires external MySQL, MongoDB, Redis.

### Alibaba Cloud

Options: ECS (full control), SAE (serverless), ACK (Kubernetes). See `DEPLOY-README.md` for topology.

---

## Port map

| Service | API Port | UI Port |
|---------|----------|---------|
| UsersPanel | 5001 | 5173 |
| Morph | 9090 | 3031 |
| Morph Utils | — | 3040 |
| FormsX | 29909 | 19909 |
| ComposerX | 8043 | 8044 |
| Booki | 9095 | 5174 |
| Morph Engi | 9096 | 5179 |
| Academi | 8978 | 8765 |
| SharpReport | 3050 | 5178 |
| BK (Ground Control) | 8000 | 3000 |

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Port already in use | `./start-all.sh status`, `lsof -i :<port>` |
| Morph macOS LC_UUID error | Use `go build` not `go run` (handled by start-all.sh) |
| 401 everywhere | `USERS_PANEL_BASE_URL`, JWT secret, `FRONTEND_ORIGIN` |
| Morph AI empty answers | `MORPH_AI_API_KEY`, `MORPH_AI_BASE_URL` region |
| SPA blank API calls | Rebuild frontend with production `VITE_*` / `REACT_APP_*` URLs |