# Morph image

One container runs Morph API and the Morph AI UI. Other apps are not in the image. No host is selected here.

## Build

From the repo root:

```bash
docker build -t morph:local .
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

## MorphUtils on Render

The product owner creates the shell. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync the Blueprint from `render.yaml` on `main`. That adds the web service `morph-utils` (Singapore, starter) beside `morph`. Render builds `morph-utils/Dockerfile` with context `morph-utils/`. Deploys from `main` run only after CI checks pass. There is no disk. Do not set Event Logs, Content Maker, or Data Access on this service. Project is a separate service below.

### Env

| Key | Required | Value |
|-----|----------|--------|
| `PORT` | yes | `3040`. Render would otherwise inject its own port. The image healthcheck calls `http://127.0.0.1:${PORT}/health`. |
| `VITE_MORPH_API_URL` | yes | `https://<morph public host>`, no path. This is the public Morph origin. It is not a secret. The Blueprint prompts for it (`sync: false`) and stores no value in git. |
| `VITE_USERS_PANEL_API_URL` | no | Alias used only when `VITE_MORPH_API_URL` is unset or blank. Leave it unset. |
| `VITE_SHEETX_URL` | no | Event Logs origin. Alias: `VITE_FORMSX_URL`. Not a service in this Blueprint. |
| `VITE_FORMSX_URL` | no | Legacy alias for `VITE_SHEETX_URL`. |
| `VITE_COMPOSERX_URL` | no | Content Maker origin. Not a service in this Blueprint. |
| `VITE_DATAX_URL` | no | Data Access origin. Not a service in this Blueprint. |
| `VITE_PROJECTS_URL` | no | Leave unset. Story #114 sets this to `https://<morph-engi public host>`. Alias: `VITE_MORPH_ENGI_URL`. |
| `VITE_MORPH_ENGI_URL` | no | Legacy alias for `VITE_PROJECTS_URL`. Leave unset. |
| `VITE_MORPH_AI_URL` | no | Morph AI origin, if the header link in the shell should leave MorphUtils. |

Set `VITE_MORPH_API_URL` in the dashboard, then restart the service so the entrypoint rewrites `/config.js`. A rebuild is not required for a runtime value. With the embed variables unset, those modules stay blank. `GET /health` does not call them.

Morph already allows cross-origin `Authorization` for non-loopback origins (`Access-Control-Allow-Origin: *`, credentials false). The shell sends `Authorization: Bearer` and does not send cookies cross-origin. This change does not edit CORS.

No JWT, password, or API key is set on this service.

### Public URL

After the first deploy is live, open the URL Render shows for `morph-utils`. `GET /health` must return HTTP 200 and a body of `ok`. Copy that origin. The placeholder for story #106 is `https://<morph-utils public host>`. #106 sets that value as `REACT_APP_MORPH_UTILS_URL` on the Morph service. This change does not set `REACT_APP_MORPH_UTILS_URL`. Do not guess an `onrender.com` host from the service name.

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

Local image build, without Render: `docker build -f morph-engi/Dockerfile -t morph-engi:local .` from the repo root. Fill `morph-engi/deploy/.env.production.example` into a gitignored env file before `docker run`.
