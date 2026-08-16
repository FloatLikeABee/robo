# Platform deployment guide

Deploy the **robo** workspace apps to **Render** or **Alibaba Cloud**. Local development stays on [`start-all.sh`](./start-all.sh); this guide is for staging and production.

**Auto-deploy entrypoint:** [`scripts/deploy.sh`](./scripts/deploy.sh)

```bash
chmod +x scripts/deploy.sh deploy/alibaba/ecs-bootstrap.sh
./scripts/deploy.sh help
```

---

## What gets deployed

| App | Folder | Runtime | Default internal port | Notes |
|-----|--------|---------|------------------------|-------|
| UsersPanel API | `UsersPanel/backend` | Rust | `5001` | Auth hub — deploy first |
| UsersPanel Admin | `UsersPanel/admin` | Static SPA | — | OSS/CDN or Render Static |
| MorphData / Morph AI | `morph` | Go (+ embedded React) | `9090` | Serves API + SPA in one process |
| FormsX API | `formx/backend` | Go | `29909` | Uploads need disk or OSS |
| FormsX UI | `formx/frontend` | Static SPA | — | |
| ComposerX API | `composerx/backend` | Go | `8043` | File storage path needed |
| ComposerX UI | `composerx/frontend` | Static SPA | — | |
| Booki API | `booki/backend` | Go | `9095` | |
| Booki UI | `booki/frontend` | Static SPA | — | |
| Morph Engi API | `morph-engi/backend` | Rust | `9096` | |
| Morph Engi UI | `morph-engi/frontend` | Static SPA | — | |
| SharpReport / DataX | `SharpReport` | Rust + Java/Metabase | `3050` | Prefer Alibaba ECS (RAM/Java) |
| Morph Utils shell | `morph-utils/frontend` | Static SPA | — | Hosts FormsX/ComposerX/DataX UI |
| Academi | `academi` | Go + RN | `8978` | Optional; mobile API only |
| platform-chat | `platform-chat` | npm package | — | Bundled into frontends |

**Shared data plane (required for Morph / FormsX / ComposerX):**

| Dependency | Typical product |
|------------|-----------------|
| MySQL 8 (`tran`) | Alibaba RDS MySQL · Render MySQL/Postgres only if you migrate |
| MongoDB (`athena`, `alterathena`) | ApsaraDB for MongoDB · Atlas |
| Redis | Alibaba Tair/Redis · Render Redis |
| Neo4j (GraphRAG, when enabled) | Native on ECS — see [`docs/MORPH_GRAPH_RAG_PLAN.md`](./docs/MORPH_GRAPH_RAG_PLAN.md) |
| DashScope / MorphAI | `MORPH_AI_API_KEY` (Alibaba DashScope recommended) |

---

## Recommended topologies

### A. Alibaba Cloud (production — preferred in CN / near DashScope)

```text
                    [ ALB / SLB + HTTPS ]
                              |
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   OSS + CDN              API group            SharpReport ECS
   (Vite SPAs)         (ECS or SAE)           (Compose / Java)
                              |
                    UsersPanel → Morph → FormsX → ComposerX → …
                              |
              RDS MySQL + MongoDB + Redis (+ OSS for uploads)
```

**Placement**

1. **Stateful:** RDS MySQL, ApsaraDB MongoDB, Tair Redis, OSS buckets for uploads/storage.
2. **UsersPanel** first (public HTTPS). All other apps set `USERS_PANEL_BASE_URL`.
3. **Morph** on ECS/SAE with a **persistent disk** for Badger (`DB_PATH`).
4. **FormsX / ComposerX / Booki / Engi** as separate web processes; SPAs on OSS+CDN.
5. **SharpReport** on a dedicated ECS (2+ vCPU, 4–8 GB) using [`SharpReport/deploy/docker-compose.yml`](./SharpReport/deploy/docker-compose.yml).
6. **Neo4j** (later): native install on ECS — not Docker (per GraphRAG plan).

### B. Render (staging / lighter Western deploy)

```text
Render Web Services (Docker)     Render Static Sites
  userspanel-api                   userspanel-admin
  morph-api                        formx-ui, composerx-ui, …
  formx-api, composerx-api, …
                                   (or Morph embeds its own SPA)
External / Alibaba RDS+Mongo+Redis  ← keep data near AI if using DashScope
```

**Render constraints**

- Bind `0.0.0.0` and listen on `$PORT` (Render injects it).
- Badger, ComposerX `storage/`, FormsX `uploads/` need a **persistent disk**.
- SharpReport + Metabase is a poor fit for free/small Render instances — use Alibaba ECS.
- Blueprint: [`deploy/render.yaml`](./deploy/render.yaml).

---

## Quick start (auto-deploy script)

From the repo root:

```bash
# 1) Check toolchains + optional CLIs
./scripts/deploy.sh doctor

# 2) Copy and fill production secrets
cp deploy/env.production.example deploy/.env.production
# edit deploy/.env.production

# 3) Build release artifacts locally (binaries + SPA dist)
./scripts/deploy.sh build --all
# or: ./scripts/deploy.sh build --app=morph,formx,userspanel

# 4a) Render: validate blueprint, then apply via Render Dashboard / CLI
./scripts/deploy.sh render validate
./scripts/deploy.sh render apply          # requires `render` CLI + login

# 4b) Alibaba: build Linux images/artifacts and sync to ECS
./scripts/deploy.sh alibaba package
./scripts/deploy.sh alibaba sync --host=user@ecs-ip --path=/opt/robo

# 5) On the ECS host (first time)
ssh user@ecs-ip 'sudo bash /opt/robo/deploy/alibaba/ecs-bootstrap.sh'
```

---

## Environment

Shared template: [`deploy/env.production.example`](./deploy/env.production.example).

Per-app examples remain in each project’s `.env.example`. Minimum cross-app keys:

| Variable | Purpose |
|----------|---------|
| `USERS_PANEL_BASE_URL` | Public UsersPanel API HTTPS URL |
| `MORPH_AI_API_KEY` | DashScope / compatible provider |
| `MORPH_AI_MODEL` | Default `qwen3-max` |
| `MORPH_AI_BASE_URL` | DashScope compatible-mode or SiliconFlow |
| `TRAN_MYSQL_DSN` / `DATABASE_URL` | MySQL |
| `TRAN_MONGO_URI` / `TRAN_MONGO_DB` | Mongo (`athena` / `alterathena`) |
| `TRAN_REDIS_ADDR` | Redis |
| `JWT_SECRET` | UsersPanel signing secret |
| `FRONTEND_ORIGIN` / CORS | Public SPA origins |
| `NEO4J_*` / `MORPH_GRAPH_ENABLED` | Optional GraphRAG |

**Never commit** `deploy/.env.production`. Set secrets in Render Dashboard or Alibaba KMS / SAE env.

---

## Build recipes (what the script runs)

| App | Build |
|-----|--------|
| UsersPanel API | `cargo build --release` → `users-panel-api` |
| UsersPanel Admin | `npm ci && npm run build` → `admin/dist` |
| Morph | `npm run build` in `frontend/` then `go build -o morph-server` |
| FormsX | `go build -o formsx-server ./cmd/server` + Vite `frontend/dist` |
| ComposerX | `go build -o composerx-server` + Vite `frontend/dist` |
| Booki | `go build -o booki-server ./cmd/server` + Vite `frontend/dist` |
| Morph Engi | `cargo build --release` + Vite `frontend/dist` |
| SharpReport | Prefer existing Compose image under `SharpReport/deploy/` |

Docker images (context = **repo root** because of `pkg/morphai` replace directives):

```bash
./scripts/deploy.sh docker build --all
# images tagged robo/<app>:<tag>
```

Dockerfiles live in [`deploy/docker/`](./deploy/docker/).

---

## Render

### Prerequisites

- GitHub/GitLab repo connected to Render
- [Render CLI](https://render.com/docs/cli) optional: `brew install render`
- External MySQL + Mongo + Redis (Alibaba RDS/Atlas/etc.) — do not rely on ephemeral Render disks for primary data

### Apply blueprint

1. Push this repo (or a deploy branch) to your remote.
2. Render Dashboard → **New** → **Blueprint** → select repo → file `deploy/render.yaml`.
3. Or: `./scripts/deploy.sh render apply`
4. Set secret env groups (`robo-secrets`) in the Dashboard to match `deploy/env.production.example`.
5. Attach **persistent disks** to Morph (`/data`), FormsX (`/uploads`), ComposerX (`/storage`).

### Port binding

All Docker `CMD` wrappers honor `PORT` when set by Render. Health paths:

| Service | Health |
|---------|--------|
| Morph | `GET /` or API route you expose |
| FormsX | Swagger `/swagger/index.html` or `/api/v1/...` |
| ComposerX | `GET /health` |
| Booki | `GET /health` |
| Morph Engi | `GET /health` |
| UsersPanel | TCP on `$PORT` (add `/health` later if desired) |
| SharpReport | `GET /api/v1/health` |

---

## Alibaba Cloud

### Option 1 — ECS (full control)

1. Provision Ubuntu 22.04+ ECS (2+ vCPU, 4+ GB). SharpReport: 4–8 GB.
2. Provision RDS MySQL, MongoDB, Redis in the same VPC.
3. Package and sync:

```bash
./scripts/deploy.sh alibaba package
./scripts/deploy.sh alibaba sync --host=root@<ecs> --path=/opt/robo
```

4. On the host:

```bash
sudo bash /opt/robo/deploy/alibaba/ecs-bootstrap.sh
# installs docker (optional), nginx, systemd unit templates
sudo cp /opt/robo/deploy/alibaba/systemd/*.service /etc/systemd/system/
# edit EnvironmentFile= paths, then:
sudo systemctl daemon-reload
sudo systemctl enable --now userspanel-api morph-api formx-api composerx-api
```

5. Point ALB/SLB + HTTPS certificates at the ECS (or Nginx on the box).
6. Upload SPA `dist/` folders to OSS and enable CDN; set `VITE_*` / `FRONTEND_ORIGIN` accordingly.

### Option 2 — SAE (Serverless App Engine)

- Build/push images to **ACR** (Alibaba Container Registry):

```bash
./scripts/deploy.sh docker build --all --tag=registry.cn-hangzhou.aliyuncs.com/<ns>/robo
./scripts/deploy.sh alibaba push-acr --registry=registry.cn-hangzhou.aliyuncs.com/<ns>
```

- Create one SAE application per API image; inject env from KMS or config.
- Mount NAS/OSS for Badger, uploads, storage.
- Put SPAs on OSS+CDN.

Example app list: [`deploy/alibaba/sae.apps.json.example`](./deploy/alibaba/sae.apps.json.example).

### Option 3 — ACK (Kubernetes)

Use the same images as SAE. Provide your own Deployments/Ingress; this repo does not yet ship full Helm charts (SharpReport K8s is documented as future in its own deploy guide).

---

## Deploy order (always)

1. MySQL schema migrations / init scripts for `tran`
2. Mongo databases `athena` / `alterathena`
3. Redis
4. **UsersPanel** API + Admin
5. Morph, FormsX, ComposerX (set satellite base URLs)
6. Booki, Morph Engi, Morph Utils
7. SharpReport (separate host)
8. (Later) Neo4j + morphgraph-worker

---

## Nginx / ALB path sketch

```nginx
# See deploy/nginx/spa-proxy.conf.example
# /api/auth, /api/users  → userspanel
# /morph/                → morph (or dedicated host)
# /formsx-api/           → formx
# /composerx-api/        → composerx
# Static hosts serve each SPA dist/ with try_files SPA fallback
```

Prefer **separate hostnames** per product (`panel.`, `morph.`, `forms.`, `mail.`) over deep path rewriting — JWT cookies and CORS stay simpler.

---

## Persistence checklist

| Path | App | Why |
|------|-----|-----|
| `DB_PATH` / `/data/badger` | Morph | Embedded Badger |
| `UPLOAD_DIR` | FormsX | Uploaded files |
| `TRAN_FILE_STORAGE_PATH` | ComposerX | Email/assets storage |
| SharpReport `data/` + Metabase | SharpReport | JVM + DB files |
| Neo4j data dir | GraphRAG | Native Neo4j |

---

## CI / automation ideas

```bash
# GitHub Actions sketch (not required)
./scripts/deploy.sh doctor
./scripts/deploy.sh build --all
./scripts/deploy.sh docker build --all --tag=ghcr.io/org/robo
# then push + render blueprint refresh or ACR + SAE update
```

Add CI workflows per app as needed; use `scripts/deploy.sh` for packaging.

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| 401 everywhere | `USERS_PANEL_BASE_URL`, JWT secret, `FRONTEND_ORIGIN` |
| Morph AI empty answers | `MORPH_AI_API_KEY`, `MORPH_AI_BASE_URL` region |
| FormsX/ComposerX DB errors | DSN, Mongo DB name (`athena` vs `alterathena`) |
| Render crash loop | App must listen on `$PORT`; check logs |
| SPA blank API calls | Rebuild frontend with production `VITE_*` / `REACT_APP_*` URLs |
| SharpReport OOM | Move to larger ECS; Metabase needs ≥2 GB |

---

## GraphRAG / Neo4j (optional)

See [`docs/MORPH_GRAPH_OPS.md`](./docs/MORPH_GRAPH_OPS.md).

```bash
cd morphgraph-worker && go build -o morphgraph-worker .
./morphgraph-worker bootstrap-schema
./morphgraph-worker backfill --all
# daily:
./morphgraph-worker sync --mode=daily
```

Morph Knowledge Library: chat drawer → **Knowledge Library** tab (md/json/csv/txt/pdf).

FormsX Survey Bot: sidebar **Survey Bot**, or chat `survey bot …`.

- Architecture: [`DEVELOPER_BASELINE.md`](./DEVELOPER_BASELINE.md)
- Assistant contract: [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](./AI_ASSISTANT_MORPHAI_CONTRACT.md)
- GraphRAG plan: [`docs/MORPH_GRAPH_RAG_PLAN.md`](./docs/MORPH_GRAPH_RAG_PLAN.md)
- SharpReport-only: [`SharpReport/docs/DEPLOYMENT.md`](./SharpReport/docs/DEPLOYMENT.md)
- UsersPanel production notes: [`UsersPanel/README.md`](./UsersPanel/README.md)

---

## GraphRAG / Neo4j (optional)

Full ops: [`docs/MORPH_GRAPH_OPS.md`](./docs/MORPH_GRAPH_OPS.md).

```bash
cd morphgraph-worker && go build -o morphgraph-worker .
./morphgraph-worker bootstrap-schema
./morphgraph-worker backfill --all
./morphgraph-worker sync --mode=daily   # cron at 03:00 recommended
```

Morph Knowledge Library lives in the chat context drawer (**Knowledge Library** tab). FormsX Survey Bot is under the FormsX sidebar.

## Script reference

```text
./scripts/deploy.sh doctor
./scripts/deploy.sh build [--all | --app=a,b]
./scripts/deploy.sh package                 # tarball under dist/deploy/
./scripts/deploy.sh docker build [--all | --app=…] [--tag=…]
./scripts/deploy.sh render validate|apply|status
./scripts/deploy.sh alibaba package|sync|push-acr|print-topology
./scripts/deploy.sh env check               # validate deploy/.env.production keys
./scripts/deploy.sh list
./scripts/deploy.sh help
```
