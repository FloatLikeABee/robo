# morph-engi-render Specification

## Purpose

Gives the product owner a Blueprint entry for Project and the steps to copy its public HTTPS URL, without this repository calling Render.

## Requirements

### Requirement: Project is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `morph-engi` after `morph` and `morph-utils`, with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./morph-engi/Dockerfile`, `dockerContext: .`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The file MUST NOT declare a `projects` block, `numInstances`, or `generateValue`. The `morph` and `morph-utils` services MUST remain, with Morph `PORT` `9090` and MorphUtils `PORT` `3040`.

#### Scenario: Blueprint matches the Project image
- **WHEN** a reviewer reads the `morph-engi` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `morph-engi/Dockerfile` with repo-root context
- **AND** its health check path is `/health` and it auto-deploys only after checks pass
- **AND** `morph` and `morph-utils` are still present with ports `9090` and `3040`

#### Scenario: The Blueprint stays attached to the existing project
- **WHEN** a reviewer reads the top-level keys of `render.yaml`
- **THEN** there is no `projects` key
- **AND** there is no `numInstances` and no `generateValue`

### Requirement: Project disk and port are pinned
The `morph-engi` service MUST declare disk `morph-engi-data` mounted at `/data` with `sizeGB: 1`, and MUST set `maxShutdownDelaySeconds: 120`. It MUST set `PORT` to `9096`. It MUST NOT set `MORPH_ENGI_PORT`.

#### Scenario: Port and disk match the image
- **WHEN** a reviewer reads the `morph-engi` service
- **THEN** `PORT` is `9096` and `MORPH_ENGI_PORT` is absent
- **AND** disk `morph-engi-data` is mounted at `/data` with size 1 GB

### Requirement: Morph API base, storage, and secrets are prompted
The `morph-engi` service MUST set `APP_ENV` to `production`, `STATIC_DIR` to `/app/frontend/dist`, `MORPH_ENGI_DATABASE_URL` to `sqlite:///data/morph_engi.db`, and `MORPH_ENGI_UPLOAD_DIR` to `/data/uploads`. It MUST list `USERS_PANEL_BASE_URL`, `JWT_SECRET`, and `MORPH_AI_API_KEY` with `sync: false` and MUST NOT give those keys a value. The committed Blueprint MUST NOT contain a JWT, password, or API key value.

#### Scenario: Storage is explicit
- **WHEN** a reviewer reads the plain env on `morph-engi`
- **THEN** the database URL is `sqlite:///data/morph_engi.db` and the upload directory is `/data/uploads`
- **AND** `APP_ENV` is `production` and `STATIC_DIR` is `/app/frontend/dist`

#### Scenario: Prompts have no values
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL`, `JWT_SECRET`, and `MORPH_AI_API_KEY` on `morph-engi`
- **THEN** each has `sync: false` and no value

### Requirement: MorphUtils embed URLs are not set
`render.yaml` MUST NOT list `VITE_PROJECTS_URL` or `VITE_MORPH_ENGI_URL` as env keys. The `morph-engi` `buildFilter.paths` MUST include `morph-engi/**`, `pkg/morphai-rs/**`, `morph-engi/Dockerfile`, `morph-engi/Dockerfile.dockerignore`, `morph-engi/deploy/docker-entrypoint.sh`, and `render.yaml`.

#### Scenario: Embed URLs stay unset
- **WHEN** a reviewer lists env keys in `render.yaml`
- **THEN** `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` are not among them

#### Scenario: Build filter covers the image inputs
- **WHEN** a reviewer reads the `morph-engi` build filter
- **THEN** it includes the Project tree, the MorphAI Rust crate, the Dockerfile, its dockerignore, the entrypoint, and `render.yaml`

### Requirement: The runbook names env and the public URL placeholder
`morph-engi/README.md` and `deploy/README.md` MUST tell the product owner to create the `morph-engi` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST list `USERS_PANEL_BASE_URL` as the Morph API origin `https://<morph public host>` and say it is not a secret. They MUST list `JWT_SECRET` (the same value as Morph, at least 32 characters) and `MORPH_AI_API_KEY` as dashboard prompts with no committed value. They MUST list `MORPH_ENGI_DATABASE_URL` and `MORPH_ENGI_UPLOAD_DIR` under `/data`, and the disk name `morph-engi-data`, and say the disk is a single instance. They MUST name `https://<morph-engi public host>` as the placeholder story #114 sets as `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` on MorphUtils. They MUST NOT set those variables. They MUST say `GET /health` is the health check and that the module id is `projects`.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Project Render section
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<morph-engi public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `VITE_PROJECTS_URL` or `VITE_MORPH_ENGI_URL`

#### Scenario: Required env is listed without a secret
- **WHEN** a product owner reads the Project env list
- **THEN** they see the Morph API base, the database URL, and the upload directory
- **AND** `JWT_SECRET` and `MORPH_AI_API_KEY` have no example value
