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

The `tls` profile still publishes 9090 on the host as well as 80 and 443. Bind that host port to `127.0.0.1`, or firewall it. Issue #53 will finalize TLS and that publish.

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
