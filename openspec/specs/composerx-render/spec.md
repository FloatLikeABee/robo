# composerx-render Specification

## Purpose

Gives the product owner a Blueprint entry for Content Maker and the steps to copy its public HTTPS URL, without this repository calling Render or wiring the MorphUtils embed.

## Requirements

### Requirement: Content Maker is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `composerx` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./composerx/Dockerfile`, `dockerContext: .`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The service names MUST be `morph`, `morph-utils`, `formx`, `composerx`, and `morph-engi`, in that order. Data Access, AI tools, and Invite Signup MUST NOT be services in that file.

#### Scenario: Blueprint matches the Content Maker image
- **WHEN** a reviewer reads the `composerx` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `composerx/Dockerfile` with context `.`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

#### Scenario: Existing services stay
- **WHEN** a reviewer lists service names in `render.yaml`
- **THEN** the names are `morph`, `morph-utils`, `formx`, `composerx`, and `morph-engi`
- **AND** `morph` and `morph-utils` are still present

### Requirement: Data disk is a single instance with a pinned port
The `composerx` service MUST attach one disk named `composerx-data`, mounted at `/data`, with `sizeGB: 1`. It MUST set `maxShutdownDelaySeconds` to a whole number from 60 through 300 inclusive. It MUST set `PORT` and `COMPOSERX_PORT` to `8043`. It MUST NOT set `numInstances`. It MUST set `GIN_MODE` to `release` and MUST set `COMPOSERX_SQLITE_PATH`, `COMPOSERX_BADGER_PATH`, and `TRAN_FILE_STORAGE_PATH` to the image defaults under `/data`.

#### Scenario: Port matches the image default
- **WHEN** a reviewer reads `PORT` and `COMPOSERX_PORT` on `composerx`
- **THEN** both values are `8043`
- **AND** the disk name is `composerx-data`, the mount path is `/data`, and the size is 1 GB

### Requirement: Morph auth URL and secrets are prompted and not invented
The `composerx` service MUST list `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, `TRAN_QWEN_API_KEY`, and `TRAN_OPENAI_API_KEY` with `sync: false` and MUST NOT give those keys a `value`. The committed Blueprint MUST NOT contain a JWT, password, API key, or a guessed public host. It MUST NOT list `VITE_COMPOSERX_URL` or `REACT_APP_MORPH_UTILS_URL` as an env key.

#### Scenario: Morph auth base URL is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `composerx`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not set here
- **WHEN** a reviewer reads env keys in `render.yaml`
- **THEN** `VITE_COMPOSERX_URL` is not one of them

### Requirement: Build filter tracks the Content Maker image inputs
The `composerx` `buildFilter.paths` MUST include `composerx/**`, `pkg/**`, `composerx/Dockerfile`, `.dockerignore`, `composerx/deploy/docker-entrypoint.sh`, and `render.yaml`.

#### Scenario: A Dockerfile input is listed
- **WHEN** a reviewer compares the `composerx` build filter to the Content Maker Dockerfile `COPY` lines
- **THEN** the Content Maker tree, `pkg/`, the Dockerfile, and the entrypoint are covered
- **AND** `render.yaml` and `.dockerignore` are listed as well

### Requirement: The runbook records the public URL placeholder
`deploy/README.md` and `composerx/backend/README.md` MUST tell the product owner to create the `composerx` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and that this value is not a secret. They MUST list `MORPH_AI_API_KEY` and `TRAN_OPENAI_API_KEY` as optional keys with no sample secret. They MUST tell the product owner to copy the service's public HTTPS URL and MUST use the placeholder `https://<composerx public host>` for the value story #114 sets as `VITE_COMPOSERX_URL` on MorphUtils. They MUST NOT set `VITE_COMPOSERX_URL` in this change. They MUST state that `GET /health` returns HTTP 200 without calling MorphUtils.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Content Maker Render section in `deploy/README.md`
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<composerx public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `VITE_COMPOSERX_URL`
