# Design

## Context

See proposal.md for why. Requirements are in `specs/morph-render/spec.md` and the `morph-container` delta.

The root `Dockerfile` (merged in #85) already builds Morph API and the Morph AI UI. Its final stage sets `PORT=9090`, the `/data` paths, and `GIN_MODE=release`. `HEALTHCHECK` calls `http://127.0.0.1:${PORT}/health`. The entrypoint starts as root, `chown`s `/data` to uid 65532, and `exec`s the server. `GET /health` is not under `/api/` and returns 200 without a session. `morph/main.go` does not install a SIGTERM handler; `r.Run` blocks until the process is killed.

This change does not edit `deploy/docker-compose.yml`, `deploy/Caddyfile`, or that runtime behavior, and it does not call Render. The product owner creates the service from the Blueprint after merge.

## Goals / Non-Goals

**Goals:**

- One flat Blueprint the product owner can attach to the existing Render project.
- Every env var the running API can read is either set or listed as deliberately unset.
- Secrets exist only as dashboard prompts (`sync: false`, no value).
- The Morph image workflow rejects a Blueprint that fails Render's schema, without a new check name.

**Non-Goals:**

- Creating or updating the Render service from this change.
- TLS, a custom domain, or Caddy. Compose stays the local path.
- Other apps, a Neo4j sidecar, or a second instance.
- A SIGTERM handler in `morph/main.go`.

## Decisions

### 1. Flat `services`, Docker, no `projects` and no `repo`

- **Choice:** Top-level `services` with one `type: web`, `runtime: docker` service. No `projects`, no `repo`, no `numInstances`.
- **Rationale:** The product owner attaches the file to the existing project, which already has the Git repo. The schema does not require `repo` on a service. A disk cannot move between instances, so an instance count above one cannot boot.
- **Rejected:** A `projects` block. That creates a project instead of joining the one that exists. A native Go or Node runtime. That would skip the entrypoint `chown` and the UI stage. Setting `numInstances: 1` is redundant with the disk and is omitted.

### 2. Pin `PORT=9090`

- **Choice:** Blueprint env `PORT=9090`, matching the image `ENV` and `EXPOSE`.
- **Rationale:** Render injects its own `PORT` for web services. If that value wins over the image `ENV`, the process binds whatever Render injected. The image healthcheck interpolates `${PORT}`, and the product owner was told the container port is 9090. Setting it in the Blueprint keeps the listen port, the Docker `HEALTHCHECK`, and Render's `healthCheckPath` on the same port. The Dockerfile is not edited.
- **Rejected:** Leave `PORT` unset and change the image to follow Render's default. That changes Dockerfile runtime behavior. Leave `PORT` unset and hope the image `ENV` wins. Render's injected value is the one that overrides image `ENV`.

### 3. `MORPH_AI_PROVIDER=dashscope`

- **Choice:** Set `dashscope`. Chat then works when the only key filled in is `MORPH_AI_API_KEY`.
- **Rationale:** `pkg/morphai/provider.go` lists `MORPH_AI_API_KEY` in DashScope `ExtraKeyEnvs`, after `DASHSCOPE_API_KEY`. `providers_test.go` ("dashscope uses morph key but not gemini key", "openai does not use morph key") shows the split. Valid ids are `openai`, `anthropic`, `xai`, `gemini`, `openrouter`, `ollama`, `dashscope`, `mistral`, `groq`, and `openai-compatible`. An empty provider also uses `MORPH_AI_API_KEY` (legacy path). `morph/ai.New` copies `config.GetConfig`'s key and model onto that client; the model default is `qwen3-max`, which is DashScope's default model. `ai.New` returns success when the key is empty, so the process still listens before the dashboard key is filled. Chat calls fail until the key is set.
- **Rejected:** `openai` or any other named id. Those read their own `*_API_KEY` and ignore `MORPH_AI_API_KEY`. Leaving the provider unset. That also works, and it hides the choice. The Blueprint sets the id that matches the key the product owner is asked to fill. `ollama`. It needs a local daemon this service does not run.

### 4. Env inventory

Source: `os.Getenv` / `LookupEnv` under `morph/` and `pkg/morphai/`, the keys passed into `morph/config`, plus `pkg/morphgraph.LoadFromEnv` because `morph/ai.New` and the Neo4j worker call those loaders at runtime. `cmd/seed_tran` and `cmd/morph-mcp` are not in the image. `MORPH_MCP_TOKEN` and `SEED_TRAN_MODE` are not slots on this service.

`gin` reads `GIN_MODE`. The image already sets `release`. The Blueprint sets it again so a dashboard edit cannot turn debug logs on without a diff.

#### Set to a plain value

| Key | Value | Read by |
|-----|--------|---------|
| `MORPH_ENV` | `production` | `config.ParseMorphEnv` |
| `PORT` | `9090` | `config.GetConfig` |
| `GIN_MODE` | `release` | gin |
| `MORPH_AI_PROVIDER` | `dashscope` | `morphai.LoadFromEnv` |
| `ADMIN_USERNAME` | `morphadmin` | `config.GetConfig` (`DefaultAdminUsername`) |
| `ADMIN_EMAIL` | `morphadmin@local.com` | `config.GetConfig` (`DefaultAdminEmail`) |
| `DB_PATH` | `/data/badger` | `config.GetConfig`; image `ENV` |
| `TRAN_SQLITE_PATH` | `/data/tran.sqlite` | `config.GetConfig`; image `ENV` |
| `ENTITY_DETAILS_BADGER` | `/data/entity_details` | `config.GetConfig`; image `ENV` |
| `MORPH_KNOWLEDGE_DIR` | `/data/knowledge` | `handlers/knowledge.go`; image `ENV` |
| `TRAN_ENTITY_ATTACHMENT_DIR` | `/data/uploads/entity_attachments` | `config.GetConfig`; image `ENV` |

#### Secrets: `sync: false`, no `value`

`JWT_SECRET`, `ADMIN_PASSWORD`, `BOOTSTRAP_ADMIN_PASSWORD`, `MORPH_AI_API_KEY`, `GEMINI_API_KEY`, `TRAN_QWEN_API_KEY`, `DASHSCOPE_API_KEY`, `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `XAI_API_KEY`, `OPENROUTER_API_KEY`, `MISTRAL_API_KEY`, `GROQ_API_KEY`, `OPENAI_COMPATIBLE_API_KEY`, `MORPH_IMAGE_API_KEY`, `POLLINATIONS_API_KEY`, `SMTP_PASS`, `TRAN_MYSQL_DSN`, `TRAN_MONGO_URI`, `NEO4J_PASSWORD`, `TRAN_OPENAI_API_KEY`.

`BOOTSTRAP_ADMIN_PASSWORD` is the fallback `config.GetConfig` reads when `ADMIN_PASSWORD` is empty. The DSN and Mongo URI can carry a password. `NEO4J_PASSWORD` and `TRAN_OPENAI_API_KEY` are read when the ingest worker calls `morphgraph.LoadFromEnv`. `generateValue` is not used: the product owner must type `JWT_SECRET` and `ADMIN_PASSWORD` to rules the process enforces, and a generated value would still be a value in the file's sibling field.

#### Left unset, and why

| Key | Why it stays unset |
|-----|--------------------|
| `JWT_EXPIRY_HOURS` | Production default is 24 hours. Allowed range is 1 through 168. The local example `876000` refuses to start. |
| `MORPH_ROTATE_DEFAULT_ADMIN` | One-shot flag. Left set, every boot tries to rotate the admin password. |
| `MORPH_AI_MODEL`, `TRAN_QWEN_MODEL`, `GEMINI_MODEL` | DashScope uses `qwen3-max` when the model is empty. `config.GetConfig` also defaults the model to `qwen3-max`. `GEMINI_MODEL` is not copied onto a named provider. |
| `MORPH_AI_API_URL` | Unset keeps compatible-mode DashScope. Setting the native generation URL turns vision off. |
| `MORPH_AI_BASE_URL`, `TRAN_QWEN_BASE_URL`, `DASHSCOPE_BASE_URL` | Default base is `https://dashscope.aliyuncs.com/compatible-mode/v1`. |
| `MORPH_AI_VISION_MODEL`, `TRAN_QWEN_VISION_MODEL` | Default `qwen-vl-max`. |
| `OPENAI_BASE_URL`, `ANTHROPIC_BASE_URL`, `XAI_BASE_URL`, `GEMINI_BASE_URL`, `OPENROUTER_BASE_URL`, `OLLAMA_BASE_URL`, `MISTRAL_BASE_URL`, `GROQ_BASE_URL`, `OPENAI_COMPATIBLE_BASE_URL` | Per-provider overrides. Unused while the provider is DashScope. A baked URL would follow a later provider change to the wrong host. |
| `EXTERNAL_API_BASE` | Go default is `http://localhost:8000`. AI tools are not in this image. Writing that default into the Blueprint would look like the tools are expected. |
| `STORAGE_BACKEND` | Default `embedded`. `legacy` needs MySQL and Mongo, which this service does not run. |
| `TRAN_MONGO_DB` | Legacy database name, default `athena`. Unused while storage is embedded. |
| `SEED_TRAN_CAP`, `SEED_TRAN_SKIP_PRUNE` | `GetConfig` reads them. The server does not seed. Defaults are 50 and false. |
| `BOOTSTRAP_ADMIN_EMAIL`, `BOOTSTRAP_ADMIN_USERNAME` | Aliases. `ADMIN_EMAIL` and `ADMIN_USERNAME` are set. |
| `SHARPREPORT_BASE_URL`, `TRANFORM_BASE_URL`, `TRANMAIL_BASE_URL`, `BOOKI_BASE_URL` | Empty disables each integration. Those apps are not this service. |
| `TRAN_ENTITY_ATTACHMENT_MAX` | Default 10. |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM`, `SMTP_USER` | Mail is off until an operator fills them. `SMTP_USER` is an account name, not a token; `SMTP_PASS` is the secret slot. |
| `MORPH_IMAGE_MODEL`, `MORPH_AI_IMAGE_MODEL` | Story images default to `wanx-v1`. |
| `MORPH_GRAPH_ENABLED` | Only `true` or `1` enables Neo4j. Unset stays off. This service has no Neo4j. |
| `NEO4J_URI`, `NEO4J_USER`, `NEO4J_DATABASE`, `MORPH_GRAPH_EMBEDDING_MODEL`, `TRAN_OPENAI_BASE_URL` | Graph defaults (`neo4j://127.0.0.1:7687`, user `neo4j`, database `neo4j`, embedding `text-embedding-3-small`). Unused while the graph is off. The password and `TRAN_OPENAI_API_KEY` are still secret slots. |
| `MORPH_MCP_TOKEN`, `SEED_TRAN_MODE`, `MORPH_MCP_STDIO_CHILD` | Read by `morph-mcp` or `cmd/seed_tran`, not by the image process. |

### 5. `maxShutdownDelaySeconds: 120`

- **Choice:** 120. The schema allows 1 through 300.
- **Rationale:** Badger `Close` syncs the value log and flushes the memtable. SQLite `Close` checkpoints the WAL. On a store that can grow to this 1 GB disk, that is usually a few seconds and can approach a minute when a checkpoint is already running. 120 seconds covers both with margin. The disk makes the deploy single-instance, so this wait is downtime. 300 seconds would hold that cutover for up to five minutes when the process is slow to exit. 120 is the shorter ceiling that still fits a clean flush.
- **What the process does today:** `morph/main.go` does not trap SIGTERM. `r.Run` does not return, so the `defer` closes do not run. The Go runtime exits, and the kernel releases the Badger directory lock and the SQLite WAL lock. The next open recovers. 120 seconds is the budget for that exit and for a later handler that does call `Close`. It is not a claim that `Close` runs today.
- **Rejected:** 30, as too tight for a value-log sync plus a WAL checkpoint on a full disk. 300, as extra downtime on every single-instance deploy. Adding a SIGTERM handler in this change. The story does not change process lifetime, and both stores already recover from an immediate exit.

### 6. `buildFilter` matches the Dockerfile inputs

Dockerfile `COPY` lines, and the pattern that covers each:

| Copied path | `buildFilter` entry |
|-------------|---------------------|
| `scripts/with-root-env.cjs` | `scripts/with-root-env.cjs` |
| `morph/frontend/**`, `morph/go.mod`, the rest of `morph/` | `morph/**` |
| `pkg/` | `pkg/**` |
| `deploy/docker-entrypoint.sh` | `deploy/docker-entrypoint.sh` |

Also listed, though not `COPY`d: `Dockerfile` (the build definition), `.dockerignore` (what the context includes), and `render.yaml` (env and disk changes must roll out). `deploy/docker-compose.yml`, `deploy/Caddyfile`, and `deploy/README.md` do not affect the image and are not listed.

`morph/**` and `pkg/**` are the patterns the story names. They cover files nested under those directories, which is everything the Dockerfile copies from them.

### 7. Schema check stays inside the Morph image workflow

- **Choice:** A step in `.github/workflows/docker-image.yml`, job name still `Build image`, runs `check-jsonschema` against `https://render.com/schema/render.yaml.json`. `deploy/check-container-contract.sh` asserts the Morph-specific fields offline (port, secrets have no value, no `projects` key, disk, filter paths). `ci.yml` is not edited.
- **Rationale:** `platform-ci` allows exactly five check names. The image workflow is already not one of them. The contract script fails before a daemon is needed; the schema step fails when the file is not a Blueprint.
- **Rejected:** A new workflow or a job in `ci.yml`. Either adds a check name the protection rules do not expect, or hides a Blueprint failure under "Go / Morph API".

### 8. Runbook, not a Render API call

- **Choice:** `deploy/README.md` gains "Deploy on Render". `docs/agents/12-build-deploy.md` points at it. `openspec/config.yaml` context stops saying Render must not be documented.
- **Rationale:** The product owner creates the service in the dashboard from the Blueprint. Disk snapshots are Render's backup for a persistent disk. `scripts/deploy.sh` stays unused.
- **Rejected:** Calling Render from this change to create the service. The story forbids it. Documenting only compose. That leaves #53 without an operator path.

## Risks / Trade-offs

- [Render injects `PORT`] → Blueprint sets `9090`.
- [Dashboard secret missing on first boot] → Process exits before listen when `JWT_SECRET` or `ADMIN_PASSWORD` is empty or a development default. The runbook states the rules (32+ random characters; 12+ and not `admin123`). `MORPH_AI_API_KEY` may be empty at boot; chat fails until it is set.
- [Disk is root-owned] → Existing entrypoint `chown`. Local verification starts the container as root against a fresh volume.
- [SIGTERM skips `Close`] → WAL recovery on next open. Shutdown budget is 120 seconds for a later handler. Not fixed here.
- [`checksPass` and an unrequired image workflow] → Document that branch protection should require `Build image` along with the five platform checks. This change cannot edit that protection.
- [Schema fetched from render.com] → The image job fails closed if the schema cannot be fetched. The contract script still checks the Morph rules offline.
- [`chown -R` on every start] → Existing ceiling, unchanged. A large `/data` makes boot slower. Upgrade path stays in the entrypoint comment.

## Migration Plan

1. Merge the Blueprint. Do not apply it from CI.
2. Product owner creates the `morph` service from `render.yaml` in the existing project and fills the secret slots.
3. After the first healthy deploy, confirm `GET /health`.
4. Rollback is a Render rollback to the previous image. The disk is not deleted by a rollback. Do not detach the disk to "fix" a bad boot.

## Open Questions

None. The host, plan, region, and port are fixed by the story.

## Grill

Proposer and reviewer passed the decisions above against the code, not against the story text alone.

Alternatives considered:

1. Compose-only, no Blueprint. Rejected. #53 is the Render deploy, and the product owner needs a file to apply.
2. `projects` wrapper. Rejected. It would create a second project.
3. Native buildpack instead of the root image. Rejected. The entrypoint and the UI stage are the deploy unit.
4. Follow Render's injected `PORT`. Rejected. It fights the image healthcheck unless the Dockerfile changes.
5. Empty `MORPH_AI_PROVIDER` (legacy DashScope). Works with `MORPH_AI_API_KEY`, and was the closest alternative. Rejected in favor of the explicit id `dashscope`, which is the only catalog id whose key fallback is `MORPH_AI_API_KEY`.
6. `generateValue` for `JWT_SECRET`. Rejected. The story requires no value, and the product owner has to choose a secret that passes the 32-character check.
7. Schema check as a new required check. Rejected. Five names stay in `ci.yml`.
8. SIGTERM handler so 120 seconds actually runs `Close`. Rejected for this change. Stores recover without it, and the process lifetime is out of scope.

Failure modes the review tried to break:

- A provider id other than `dashscope` with only `MORPH_AI_API_KEY` does not resolve a key. The Blueprint uses `dashscope`.
- A committed secret or `sync: true` with a value would ship in git. The contract script forbids a `value` on the secret keys.
- `JWT_EXPIRY_HOURS=876000` copied from the local `.env` example aborts production startup. The key is absent.
- Two instances and one disk: the second never mounts. No instance count is set, and the runbook says deploys are not zero-downtime.
- `morph/**` missing a `COPY`. The table in decision 6 is the check. The contract script requires each listed pattern.
- Volume mounted as root, server uid 65532, no `USER` in the Dockerfile. The entrypoint already handles that. Verification uses a fresh volume and does not pass `--user`.
