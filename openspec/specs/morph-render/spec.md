# morph-render Specification

## Purpose

Gives the product owner a Blueprint that runs the existing Morph image on Render, with data on one disk and secrets filled in the dashboard, without this repository calling Render.

## Requirements

### Requirement: One Docker web service in a flat Blueprint
The repository MUST contain `render.yaml` at the repo root. The file MUST declare a top-level `services` list and MUST NOT declare a `projects` block. That list MUST contain one service with `type: web`, `runtime: docker`, `name: morph`, `region: singapore`, `plan: starter`, and `branch: main`. The service MUST build `dockerfilePath: ./Dockerfile` with `dockerContext: .`. It MUST set `healthCheckPath: /health` and `autoDeployTrigger: checksPass`.

#### Scenario: Blueprint matches the image
- **WHEN** a reviewer reads the only service in `render.yaml`
- **THEN** it is a Docker web service named `morph` in `singapore` on the `starter` plan, tracking `main`, built from the root Dockerfile and context
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

### Requirement: Data disk is a single instance
The service MUST attach one disk named `morph-data`, mounted at `/data`, with `sizeGB: 1`. The Blueprint MUST NOT set an instance count above one.

#### Scenario: Disk mount matches the image
- **WHEN** a reviewer reads the service disk
- **THEN** the name is `morph-data`, the mount path is `/data`, and the size is 1 GB

### Requirement: Shutdown window covers store close
The service MUST set `maxShutdownDelaySeconds` to a whole number from 60 through 300 inclusive.

#### Scenario: Delay is inside the schema range and long enough to flush
- **WHEN** a reviewer reads `maxShutdownDelaySeconds`
- **THEN** the value is at least 60 and at most 300

### Requirement: Build filter tracks the image inputs
`buildFilter.paths` MUST include `morph/**`, `pkg/**`, `Dockerfile`, `.dockerignore`, `scripts/with-root-env.cjs`, `deploy/docker-entrypoint.sh`, and `render.yaml`.

#### Scenario: A Dockerfile input is listed
- **WHEN** a reviewer compares `buildFilter.paths` to the Dockerfile `COPY` lines
- **THEN** each copied path is covered by a listed pattern or exact path
- **AND** `render.yaml` and `.dockerignore` are listed as well

### Requirement: Production env is explicit and secrets are empty
The service env list MUST set `MORPH_ENV` to `production`, `PORT` to `9090`, `GIN_MODE` to `release`, and `MORPH_AI_PROVIDER` to `dashscope`. It MUST set `ADMIN_USERNAME` and `ADMIN_EMAIL` to plain values. It MUST set `DB_PATH`, `TRAN_SQLITE_PATH`, `ENTITY_DETAILS_BADGER`, `MORPH_KNOWLEDGE_DIR`, and `TRAN_ENTITY_ATTACHMENT_DIR` to the image defaults under `/data`. Every secret the Morph API can read, including `JWT_SECRET`, `ADMIN_PASSWORD`, and `MORPH_AI_API_KEY`, MUST appear as `sync: false` and MUST NOT have a `value`. The committed file MUST NOT contain a usable JWT secret, admin password, or API key.

#### Scenario: Port is pinned
- **WHEN** a reviewer reads the `PORT` entry
- **THEN** its value is `9090`

#### Scenario: Secrets are dashboard-only
- **WHEN** a reviewer reads `JWT_SECRET`, `ADMIN_PASSWORD`, and `MORPH_AI_API_KEY`
- **THEN** each has `sync: false` and no value

### Requirement: Render runbook names the dashboard secrets
`deploy/README.md` MUST include a Deploy on Render section that tells the product owner to create the service from the Blueprint, to set `JWT_SECRET` to at least 32 random characters, and to set `ADMIN_PASSWORD` to at least 12 characters and not `admin123`. It MUST state that the disk makes deploys single-instance and not zero-downtime, that `/data` is backed up with Render disk snapshots, and that `GET /health` is how to check the service after a deploy.

#### Scenario: Operator fills secrets from the runbook
- **WHEN** an operator reads the Deploy on Render section
- **THEN** they see the JWT secret length, the admin password length, and that `admin123` is refused
- **AND** they see the disk constraint, snapshot backup, and `/health`
