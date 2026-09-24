# Morph image

One container runs Morph API and the Morph AI UI. Other apps are not in the image. No host is selected here.

## Build

From the repo root:

```bash
docker build -t morph:local .
```

That build leaves `REACT_APP_MORPH_UTILS_URL` unset, so the production UI omits the MorphUtils header chip. To inline a public origin, pass it as a build arg and rebuild. There is no default host:

```bash
docker build --build-arg REACT_APP_MORPH_UTILS_URL=https://<morph-utils public host> -t morph:local .
```

The context is the repo root so `pkg/` (Go `replace` directives) and `scripts/with-root-env.cjs` are available. `.dockerignore` keeps env files, `node_modules`, local `data/`, and the other apps out. There are no secret build args.

`sh deploy/check-container-contract.sh` checks the Dockerfile, compose file, and CI job names without a daemon.

## Run

```bash
cp deploy/.env.production.example deploy/.env.production
# fill JWT_SECRET (32+ random characters) and ADMIN_PASSWORD (12+)
docker compose -f deploy/docker-compose.yml up --build
```

`deploy/.env.production` is gitignored. Compose loads it with `env_file`. It is not the compose interpolation file (that would be `deploy/.env`).

The process listens on port 9090 inside the container. The container `PORT` must stay 9090. Leave it unset so the image default is used. The healthcheck follows `PORT`, and compose always maps the host port (`MORPH_PUBLISH_PORT`, default 9090) to container port 9090.

The `tls` profile still publishes 9090 on the host as well as 80 and 443. Bind that host port to `127.0.0.1`, or firewall it. Render terminates TLS for the hosted service. The compose TLS profile is not that deploy.

`GET /health` is the container healthcheck. `GET /` is the Morph AI UI. The API and the UI are the same origin.

`docker compose up` does not start Caddy. TLS is the `tls` profile (`deploy/Caddyfile`, issue #53):

```bash
docker compose -f deploy/docker-compose.yml --profile tls up
```

## Volume

Named volume `morph-data` is mounted at `/data`. The image defaults are:

| Variable | Path |
|----------|------|
| `DB_PATH` | `/data/badger` |
| `TRAN_SQLITE_PATH` | `/data/tran.sqlite` |
| `ENTITY_DETAILS_BADGER` | `/data/entity_details` |
| `MORPH_KNOWLEDGE_DIR` | `/data/knowledge` |
| `TRAN_ENTITY_ATTACHMENT_DIR` | `/data/uploads/entity_attachments` |

The entrypoint starts as root, gives `/data` to uid 65532 (`morph`), and execs the server as that user. Run one replica. SQLite and Badger are single-writer.

A checkout started with `start-all.sh` still uses `./data` under the working directory. Those defaults are not changed.

## Env

Required in `deploy/.env.production` when `MORPH_ENV=production`:

- `JWT_SECRET` — unique, at least 32 characters, not the development default
- `ADMIN_PASSWORD` — unique, at least 12 characters, not `admin123`

Leave `JWT_EXPIRY_HOURS` unset (24 hours). `876000` is refused. `MORPH_AI_API_KEY` is optional; the process starts without it.

A development secret or password makes the process exit before it listens. `restart: unless-stopped` will then retry. Fix the env file. Do not commit it.

## Backup

Stop the container so SQLite is not mid-write, then copy the volume:

```bash
docker compose -f deploy/docker-compose.yml stop morph
docker run --rm -v deploy_morph-data:/data:ro -v "$PWD":/backup alpine \
  tar -C /data -czf /backup/morph-data.tgz .
docker compose -f deploy/docker-compose.yml start morph
```

The volume name is `<compose project>_morph-data`. From `deploy/` the project defaults to `deploy`, so the volume is `deploy_morph-data`. `docker volume ls` shows the real name.

## Upgrade

Build the new image and start it on the same volume:

```bash
docker compose -f deploy/docker-compose.yml up --build -d
```

`docker compose down` keeps the volume. `docker compose down -v` deletes it.

## Deploy on Render

The product owner creates the service. This repo does not call Render. After merge, in the existing Render project, create a Blueprint and point it at `render.yaml` on `main`. Render builds the root `Dockerfile` (context `.`) as the web service `morph` in Singapore on the starter plan. Deploys from `main` run only after CI checks pass (`autoDeployTrigger: checksPass`). Branch protection should require the five platform checks and `Build image`, so a red image build does not go out.

### Secrets

The Blueprint lists every secret with `sync: false` and no value. Fill these in the dashboard before the first boot. Startup in `MORPH_ENV=production` exits before it listens when a required secret is missing or is a development default.

| Key | Rule |
|-----|------|
| `JWT_SECRET` | At least 32 random characters. Not `morph-dev-jwt-secret-change-me`, not a repeated character, and not a `change-me` / `replace-me` placeholder. |
| `ADMIN_PASSWORD` | At least 12 characters. Not `admin123`. |
| `MORPH_AI_API_KEY` | DashScope key. The service sets `MORPH_AI_PROVIDER=dashscope`, which reads this key. The process starts without it; chat fails until it is set. |
| Other `sync: false` keys | Leave blank unless you use that integration (`GEMINI_API_KEY`, `SMTP_PASS`, `NEO4J_PASSWORD`, and the rest). Do not paste them into git. |

`PORT` is `9090` in the Blueprint. Render would otherwise inject its own port, and the image healthcheck calls `http://127.0.0.1:${PORT}/health`.

Leave `JWT_EXPIRY_HOURS` unset. Production uses 24 hours. `876000` is refused.

### Disk

`morph-data` is mounted at `/data` (1 GB). A persistent disk attaches to one instance, so the service cannot do a zero-downtime deploy and must stay a single instance. Do not scale it out. During a deploy Render stops the old instance before the new one can mount the disk.

The image entrypoint starts as root, gives `/data` to uid 65532, and then runs the server as that user. Render disks mount as root. That chown is what makes the first boot writable.

### Backup

Use Render disk snapshots of `morph-data`. Take a snapshot before an upgrade you may need to undo. A snapshot is a point-in-time copy of `/data` (Badger, `tran.sqlite` and its WAL, knowledge, and uploads). Restoring a snapshot replaces the disk contents. Do not `docker compose down -v` against this disk; that command is only for the local compose volume.

### After a deploy

Open `https://<the service host>/health`. It must return HTTP 200 and a JSON body with `"status": "healthy"`. The same path is the container healthcheck. `GET /` is the Morph AI UI.

## MorphUtils stack on Render

Forge must not create services on Render. The product owner creates the services in Render project `prj-dahc33dbedkc73a1v8n0` (Prod). This repository does not call Render.

`morph` is already the Morph panel. Example already live (example, do not recreate): `https://morph-gjmb.onrender.com`. Copy the origin the dashboard shows. Do not guess an `onrender.com` host from the service name.

### 1. Sync the Blueprint

After this file is on `main`, sync `render.yaml` in that project. One sync creates any missing service:

| Service | Product |
|---------|---------|
| `morph-utils` | MorphUtils |
| `formx` | Event Logs |
| `composerx` | Content Maker |
| `sharpreport` | Data Access |
| `morph-engi` | Project |

If a service is already there, the sync updates it. Do not create a second copy.

### 2. Copy each public HTTPS origin

After the service is live, open the URL the dashboard shows. `GET /health` must succeed (`morph-utils` returns a body of `ok`; the others return HTTP 200). Write the origin down outside git. These hosts are examples the product owner already created. They are not required names.

| Service | Product | Health | Example already live (example, do not recreate) |
|---------|---------|--------|--------------------------------------------------|
| `morph-utils` | MorphUtils | `GET /health` body `ok` | `https://morph-utils.onrender.com` |
| `formx` | Event Logs | `GET /health` | `https://formx-vucj.onrender.com` |
| `composerx` | Content Maker | `GET /health` | `https://composerx.onrender.com` |
| `sharpreport` | Data Access | `GET /health` | `https://sharpreport.onrender.com` |
| `morph-engi` | Project | `GET /health` | `https://morph-engi.onrender.com` |

### 3. Dashboard env the create step can drop

A Blueprint create can drop nested env vars. After create, open each service. If `USERS_PANEL_BASE_URL` is missing or blank on `formx`, `composerx`, `sharpreport`, or `morph-engi`, set it in the dashboard to `https://<morph public host>` with no path. The Morph panel example above is one such origin. It is not a secret. Save so that service starts again. `GET /health` does not call Morph, so a green health check does not prove this key is set.

On `morph-utils`, set `VITE_MORPH_API_URL` to `https://<morph public host>` when it is missing or blank, then restart. The entrypoint rewrites `/config.js`. A MorphUtils image rebuild is not required for that key.

Do not put a JWT, password, or API key in git. Fill those prompts in the dashboard only.

### 4. Rebuild Morph after the MorphUtils origin exists

Only after step 2 has copied the `morph-utils` origin: on the `morph` service, set `REACT_APP_MORPH_UTILS_URL` to `https://<morph-utils public host>` (no path) and rebuild the Morph image. The Blueprint already lists that key with `sync: false` and no value. Fill the dashboard prompt. Do not put a value in `render.yaml`. Morph image rebuild is required. The root Dockerfile declares that name as `ARG` and the UI build inlines it. Render passes service env vars into the Docker build. A restart without a rebuild does not set the header link. An empty or loopback value omits it. If the link is still missing after the deploy, clear the build cache and deploy again. Do not commit the URL.

### 5. Env matrix for story #114

Leave these `VITE_*` keys unset on `morph-utils` in this change. Story #114 sets them. They are read when the MorphUtils container starts, so a later change does not require a MorphUtils image rebuild. Fill the recorded origin outside git.

| Key | Product | Placeholder | Example already live (example, do not recreate) | Recorded origin |
|-----|---------|-------------|--------------------------------------------------|-----------------|
| `REACT_APP_MORPH_UTILS_URL` (on `morph`; Morph image rebuild is required) | MorphUtils | `https://<morph-utils public host>` | `https://morph-utils.onrender.com` | |
| `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) | Event Logs | `https://<event-logs public host>` | `https://formx-vucj.onrender.com` | |
| `VITE_COMPOSERX_URL` | Content Maker | `https://<composerx public host>` | `https://composerx.onrender.com` | |
| `VITE_DATAX_URL` | Data Access | `https://<sharpreport public host>` | `https://sharpreport.onrender.com` | |
| `VITE_PROJECTS_URL` (alias `VITE_MORPH_ENGI_URL`) | Project | `https://<morph-engi public host>` | `https://morph-engi.onrender.com` | |

The sections below keep the per-service port, disk, and secret tables. Those sections still do not put a value for `REACT_APP_MORPH_UTILS_URL` in the Blueprint. The `morph` service lists the key as a dashboard prompt.

## MorphUtils on Render

The product owner creates the shell. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync the Blueprint from `render.yaml` on `main`. That adds the web service `morph-utils` (Singapore, starter) beside `morph`. Render builds `morph-utils/Dockerfile` with context `morph-utils/`. Deploys from `main` run only after CI checks pass. There is no disk. Event Logs is the `formx` service. Content Maker is the `composerx` service. Data Access is the `sharpreport` service. Project is the `morph-engi` service. Do not set `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, or `VITE_DATAX_URL` on `morph-utils`.

### Env

| Key | Required | Value |
|-----|----------|--------|
| `PORT` | yes | `3040`. Render would otherwise inject its own port. The image healthcheck calls `http://127.0.0.1:${PORT}/health`. |
| `VITE_MORPH_API_URL` | yes | `https://<morph public host>`, no path. This is the public Morph origin. It is not a secret. The Blueprint prompts for it (`sync: false`) and stores no value in git. |
| `VITE_USERS_PANEL_API_URL` | no | Alias used only when `VITE_MORPH_API_URL` is unset or blank. Leave it unset. |
| `VITE_SHEETX_URL` | no | Event Logs origin. Alias: `VITE_FORMSX_URL`. Do not set this on `morph-utils`. Story #114 uses `https://<event-logs public host>`. |
| `VITE_FORMSX_URL` | no | Legacy alias for `VITE_SHEETX_URL`. Do not set this on `morph-utils`. |
| `VITE_COMPOSERX_URL` | no | Content Maker origin. This change does not set it on `morph-utils`. #114 uses `https://<composerx public host>`. |
| `VITE_DATAX_URL` | no | Data Access origin. Do not set this on `morph-utils`. The placeholder for story #114 is `https://<sharpreport public host>`. |
| `VITE_PROJECTS_URL` | no | Leave unset. Story #114 sets this to `https://<morph-engi public host>`. Alias: `VITE_MORPH_ENGI_URL`. |
| `VITE_MORPH_ENGI_URL` | no | Legacy alias for `VITE_PROJECTS_URL`. Leave unset. |
| `VITE_MORPH_AI_URL` | no | Morph AI origin, if the header link in the shell should leave MorphUtils. |

Set `VITE_MORPH_API_URL` in the dashboard, then restart the service so the entrypoint rewrites `/config.js`. A rebuild is not required for a runtime value. With the embed variables unset, those modules stay blank. `GET /health` does not call them.

Morph already allows cross-origin `Authorization` for non-loopback origins (`Access-Control-Allow-Origin: *`, credentials false). The shell sends `Authorization: Bearer` and does not send cookies cross-origin. This change does not edit CORS.

No JWT, password, or API key is set on this service.

### Public URL

After the first deploy is live, open the URL Render shows for `morph-utils`. `GET /health` must return HTTP 200 and a body of `ok`. Copy that origin. The placeholder is `https://<morph-utils public host>`. The `morph` service lists `REACT_APP_MORPH_UTILS_URL` as a dashboard prompt with no value. Fill that prompt and rebuild Morph. This MorphUtils section does not set `REACT_APP_MORPH_UTILS_URL` on `morph-utils`. Do not guess an `onrender.com` host from the service name.

## Event Logs on Render

The product owner creates the service. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync the Blueprint from `render.yaml` on `main`. That adds the web service `formx` (Singapore, starter) beside `morph` and `morph-utils`. Render builds `formx/Dockerfile` with context `.` (the repo root, so `pkg/` is available). Deploys from `main` run only after CI checks pass. Do not set `VITE_SHEETX_URL` or `VITE_FORMSX_URL` in this Blueprint. Do not set `REACT_APP_MORPH_UTILS_URL` on `formx`. The `morph` service lists that key as a dashboard prompt with no value.

### Env

| Key | Required | Value |
|-----|----------|--------|
| `PORT` | yes | `29909`. Render would otherwise inject its own port. The image healthcheck calls `http://127.0.0.1:${PORT}/health`, and the entrypoint copies `PORT` onto `SERVER_PORT`. |
| `FORMSX_SQLITE_PATH` | yes | `/data/formsx.sqlite` |
| `FORMSX_BADGER_PATH` | yes | `/data/formsx_badger` |
| `UPLOAD_DIR` | yes | `/data/uploads` |
| `USERS_PANEL_BASE_URL` | yes | `https://<morph public host>`, no path. This is the public Morph origin. It is not a secret. The Blueprint prompts for it (`sync: false`) and stores no value in git. Login fails until it is set. `GET /health` does not call Morph or MorphUtils. |
| `PUBLIC_FORM_BASE_URL` | no | `https://<event-logs public host>`, no path. Used in broadcast mail. Prompted, no value in git. |
| `MORPH_AI_API_KEY` | no | Optional model key. The process starts without it. |
| `SMTP_PASSWORD` | no | Optional. Leave blank unless broadcast mail needs SMTP. Do not commit it. |

`SMTP_HOST`, `SMTP_USER`, and `SMTP_FROM` stay unset unless you send mail. Do not put a password in git.

### Disk

`formx-data` is mounted at `/data` (1 GB). A persistent disk is a single instance, so the service cannot do a zero-downtime deploy and must not be scaled out. During a deploy Render stops the old instance before the new one can mount the disk.

The entrypoint starts as root, gives `/data` to uid 65532, and then runs the server as that user. SQLite, Badger, and uploads all live on that mount.

Use Render disk snapshots of `formx-data`. Take a snapshot before an upgrade you may need to undo. Restoring a snapshot replaces the disk contents.

### Public URL

After the first deploy is live, open the URL Render shows for `formx`. `GET /health` must return HTTP 200 and a JSON body with `"status": "healthy"`. `GET /events-info` is the UI. Copy that origin. The placeholder for story #114 is `https://<event-logs public host>`. #114 sets that value as `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) on MorphUtils. This change does not set `VITE_SHEETX_URL` on MorphUtils. Do not guess an `onrender.com` host from the service name.

## Content Maker on Render

The product owner creates the service. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync the Blueprint from `render.yaml` on `main`. That adds the web service `composerx` (Singapore, starter) beside `morph`, `morph-utils`, and `formx`. Render builds `composerx/Dockerfile` with context `.` (the repo root, so `pkg/` replace directives resolve). Deploys from `main` run only after CI checks pass. The image runbook for a local build is [`composerx/backend/README.md`](../composerx/backend/README.md).

`GET /health` returns HTTP 200 and `"status":"ok"` without calling Morph or MorphUtils. `GET /` is the Content Maker UI on the same port. Local Vite stays on 8044; the image does not publish 8044.

### Env

| Key | Required | Value |
|-----|----------|--------|
| `PORT` | yes | `8043`. Render would otherwise inject its own port. The image healthcheck calls `http://127.0.0.1:${PORT}/health`. `COMPOSERX_PORT` is also `8043`. The process listens on `COMPOSERX_PORT` when that is set. |
| `COMPOSERX_PORT` | yes | `8043` |
| `GIN_MODE` | yes | `release` |
| `COMPOSERX_SQLITE_PATH` | yes | `/data/composerx.sqlite` |
| `COMPOSERX_BADGER_PATH` | yes | `/data/composerx_badger` |
| `TRAN_FILE_STORAGE_PATH` | yes | `/data/storage` |
| `USERS_PANEL_BASE_URL` | yes for sign-in | `https://<morph public host>`, no path. This is the Morph auth base URL. It is not a secret. The Blueprint prompts for it (`sync: false`) and stores no value in git. `/health` does not call it. |
| `MORPH_AI_API_KEY` | no | DashScope key for compose chat. The process starts without it. |
| `TRAN_QWEN_API_KEY` | no | Legacy alias used only when `MORPH_AI_API_KEY` is unset. |
| `TRAN_OPENAI_API_KEY` | no | Optional reference-library embeddings. Leave blank unless you use that path. |

Do not put key values in git. This change does not set `VITE_COMPOSERX_URL` on MorphUtils.

### Disk

`composerx-data` is mounted at `/data` (1 GB). A persistent disk attaches to one instance, so the service cannot do a zero-downtime deploy and must stay a single instance. SQLite is a single writer. Do not scale it out. The entrypoint starts as root, gives `/data` to uid 65532, and then runs the server as that user.

Use Render disk snapshots of `composerx-data` before an upgrade you may need to undo.

### Public URL

After the first deploy is live, open the URL Render shows for `composerx`. `GET /health` must return HTTP 200. Copy that origin. The placeholder for story #114 is `https://<composerx public host>`. #114 sets that value as `VITE_COMPOSERX_URL` on MorphUtils. This change does not set `VITE_COMPOSERX_URL`. Do not guess an `onrender.com` host from the service name.

## Data Access on Render

The product owner creates the service. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync the Blueprint from `render.yaml` on `main`. That adds the web service `sharpreport` (Singapore, starter) beside `morph`, `morph-utils`, `formx`, and `composerx`. Render builds `SharpReport/Dockerfile` with context `.` (the image needs `pkg/morphai-rs`). Deploys from `main` run only after CI checks pass.

### Local image

From the repo root:

```bash
docker build -t sharpreport:local -f SharpReport/Dockerfile .
docker run --rm -p 127.0.0.1:3050:3050 sharpreport:local
```

`GET /health` and `GET /ready` return HTTP 200 and `{"status":"ok"}` without Morph. `GET /` is the Data Access UI. The image does not publish Vite port 5178. `sh SharpReport/deploy/check-container-contract.sh` checks the contract without a daemon.

`docker compose -f SharpReport/docker-compose.yml up --build` uses `SharpReport/deploy/.env.production` when that file exists (copy from `SharpReport/deploy/.env.production.example`). Data is the named volume at `/data`.

### Env

| Key | Required | Value |
|-----|----------|--------|
| `PORT` | yes | `3050`. Render would otherwise inject its own port. The image healthcheck calls `http://127.0.0.1:${PORT}/health`. `SHARPREPORT_PORT` is also `3050`. The process listens on `SHARPREPORT_PORT` when that is set. |
| `SHARPREPORT_PORT` | yes | `3050` |
| `SHARPREPORT_DATABASE_URL` | yes | `sqlite:///data/datapulse.db` |
| `SHARPREPORT_UI_DIR` | yes | `/app/ui` |
| `USERS_PANEL_BASE_URL` | yes for sign-in | `https://<morph public host>`, no path. This is the Morph auth base URL. It is not a secret. The Blueprint prompts for it (`sync: false`) and stores no value in git. `GET /health` does not call it. |
| `JWT_SECRET` | before storing connection passwords | Dashboard prompt. Empty still lets `/health` succeed. The development file secret is used until this is set. |
| `MORPH_AI_API_KEY` | no | Data AI stays off until this is set. |

Do not pass secrets as `docker build` arguments. Do not set `RUN_ENV=production` on this service. That file selects Postgres and placeholder passwords. Metabase is not in the image.

### Disk

`sharpreport-data` is mounted at `/data` (1 GB). A persistent disk attaches to one instance, so the service cannot do a zero-downtime deploy and must stay a single instance. SQLite is a single writer. Do not scale it out. The entrypoint starts as root, gives `/data` to uid 65532, and then runs the server as that user.

Use Render disk snapshots of `sharpreport-data` before an upgrade you may need to undo.

### Public URL

After the first deploy is live, open the URL Render shows for `sharpreport`. `GET /health` must return HTTP 200. Copy that origin. The placeholder for story #114 is `https://<sharpreport public host>`. #114 sets that origin as `VITE_DATAX_URL` on MorphUtils. This change does not set `VITE_DATAX_URL`. Do not guess an `onrender.com` host from the service name.

## Project on Render

The product owner creates the service. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync the Blueprint from `render.yaml` on `main`. That adds the web service `morph-engi` (Singapore, starter) after `morph-utils`. Render builds `morph-engi/Dockerfile` with context `.`. Deploys from `main` run only after CI checks pass. The MorphUtils module id is `projects`. This Blueprint does not set `VITE_PROJECTS_URL` or `VITE_MORPH_ENGI_URL`.

### Env

| Key | Required | Value |
|-----|----------|--------|
| `PORT` | yes | `9096`. Do not set `MORPH_ENGI_PORT`; it overrides `PORT`, and the image healthcheck calls `http://127.0.0.1:${PORT}/health`. |
| `APP_ENV` | yes | `production`. A development `JWT_SECRET` or a loopback Morph API base then refuses to listen. |
| `STATIC_DIR` | yes | `/app/frontend/dist`. The API serves the UI on the same origin. |
| `MORPH_ENGI_DATABASE_URL` | yes | `sqlite:///data/morph_engi.db` |
| `MORPH_ENGI_UPLOAD_DIR` | yes | `/data/uploads` |
| `USERS_PANEL_BASE_URL` | yes | `https://<morph public host>`, no path. This is the public Morph API origin. It is not a secret. The Blueprint prompts for it (`sync: false`) and stores no value in git. |
| `JWT_SECRET` | yes | The same value as Morph. At least 32 characters. Dashboard prompt, no value in git. |
| `MORPH_AI_API_KEY` | no | Dashboard prompt, no value in git. `/health` succeeds without it. AI project documents do not. |

### Disk

`morph-engi-data` is mounted at `/data` (1 GB). A persistent disk attaches to one instance, so the service is a single instance. Do not scale it out. The entrypoint starts as root, creates `/data/uploads`, gives `/data` to uid 65532, and then runs the server as that user.

### Public URL

`GET /health` must return HTTP 200. Copy the HTTPS origin Render shows. The placeholder for story #114 is `https://<morph-engi public host>`. #114 sets that origin as `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` on MorphUtils. This change does not set those variables.

Local image build, without Render: `DOCKER_BUILDKIT=1 docker build -f morph-engi/Dockerfile -t morph-engi:local .` from the repo root. BuildKit is required so `morph-engi/Dockerfile.dockerignore` is used instead of the root ignore file. Fill `morph-engi/deploy/.env.production.example` into a gitignored env file before `docker run`.
