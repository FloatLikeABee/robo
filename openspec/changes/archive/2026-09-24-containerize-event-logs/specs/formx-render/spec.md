## Purpose

Gives the product owner a Blueprint entry that runs the Event Logs image on Render, with data on one disk and the Morph auth origin filled in the dashboard, without this repository calling Render.

## ADDED Requirements

### Requirement: Event Logs is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `formx` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./formx/Dockerfile`, `dockerContext: .`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The `morph` and `morph-utils` services MUST remain. The file MUST NOT set `numInstances`.

#### Scenario: Blueprint matches the Event Logs image
- **WHEN** a reviewer reads the `formx` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `formx/Dockerfile` with context `.`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

### Requirement: Data disk is a single instance
The `formx` service MUST attach one disk named `formx-data`, mounted at `/data`, with `sizeGB: 1`. It MUST set `maxShutdownDelaySeconds` to a whole number from 60 through 300 inclusive.

#### Scenario: Disk mount matches the image
- **WHEN** a reviewer reads the `formx` disk
- **THEN** the name is `formx-data`, the mount path is `/data`, and the size is 1 GB

### Requirement: Build filter tracks the Event Logs image inputs
The `formx` `buildFilter.paths` MUST include `formx/**`, `pkg/**`, `formx/Dockerfile`, `formx/Dockerfile.dockerignore`, `formx/deploy/docker-entrypoint.sh`, and `render.yaml`. The `morph` and `morph-utils` build filters MUST stay as they are.

#### Scenario: A Dockerfile input is listed
- **WHEN** a reviewer compares the `formx` build filter to the Event Logs Dockerfile `COPY` lines
- **THEN** the `formx` tree and `pkg` are covered
- **AND** `render.yaml` is listed
- **AND** the `morph` filter does not gain a `formx/**` path

### Requirement: Port and storage paths are pinned
The `formx` service MUST set `PORT` to `29909`. It MUST set `FORMSX_SQLITE_PATH`, `FORMSX_BADGER_PATH`, and `UPLOAD_DIR` to the image paths under `/data`.

#### Scenario: Port matches the image default
- **WHEN** a reviewer reads the `PORT` entry on `formx`
- **THEN** its value is `29909`

### Requirement: Morph auth origin is prompted and secrets are empty
The `formx` service MUST list `USERS_PANEL_BASE_URL` with `sync: false` and MUST NOT give that key a `value`. `MORPH_AI_API_KEY`, `SMTP_PASSWORD`, and `PUBLIC_FORM_BASE_URL`, when listed, MUST use `sync: false` and MUST NOT have a `value`. The committed Blueprint MUST NOT contain a usable JWT, password, or API key. It MUST NOT set `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, or `REACT_APP_MORPH_UTILS_URL` on any service.

#### Scenario: Morph auth origin is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `formx`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not wired here
- **WHEN** a reviewer reads env keys in `render.yaml`
- **THEN** `VITE_SHEETX_URL` and `REACT_APP_MORPH_UTILS_URL` are not set

### Requirement: The runbook names env, the disk, and the public URL placeholder
`deploy/README.md` and `formx/README.md` MUST tell the product owner to create the `formx` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and that this value is not a secret. They MUST document the `/data` disk (SQLite, Badger, uploads), that a disk is a single instance, and that `GET /health` does not call MorphUtils. They MUST use the placeholder `https://<event-logs public host>` for the origin story #114 sets as `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) on MorphUtils. They MUST NOT set that variable on Morph or MorphUtils in this change.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Event Logs Render section
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<event-logs public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `VITE_SHEETX_URL` on MorphUtils
