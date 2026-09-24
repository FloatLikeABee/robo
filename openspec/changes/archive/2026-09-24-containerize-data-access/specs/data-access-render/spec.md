## Purpose

Gives the product owner a Blueprint entry for Data Access and the placeholder for its public URL, without this repository calling Render or setting the MorphUtils embed variable.

## ADDED Requirements

### Requirement: Data Access is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `sharpreport` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./SharpReport/Dockerfile`, `dockerContext: .`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The existing `morph` and `morph-utils` services MUST remain.

#### Scenario: Blueprint matches the Data Access image
- **WHEN** a reviewer reads the `sharpreport` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `SharpReport/Dockerfile` with context `.`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

#### Scenario: Existing services stay
- **WHEN** a reviewer lists service names in `render.yaml`
- **THEN** `morph` and `morph-utils` are still present
- **AND** `sharpreport` is present after them

### Requirement: SQLite disk and pinned port
The `sharpreport` service MUST attach one disk named `sharpreport-data`, mounted at `/data`, with `sizeGB: 1`. It MUST set `maxShutdownDelaySeconds` to `120`. It MUST set `PORT` and `SHARPREPORT_PORT` to `3050`. It MUST set `SHARPREPORT_DATABASE_URL` to `sqlite:///data/datapulse.db`. It MUST NOT set `numInstances`.

#### Scenario: Port matches the image default
- **WHEN** a reviewer reads the `PORT` and `SHARPREPORT_PORT` entries on `sharpreport`
- **THEN** both values are `3050`
- **AND** the disk name is `sharpreport-data` at `/data` with size 1 GB

### Requirement: Morph auth base URL is prompted and not invented
The `sharpreport` service MUST list `USERS_PANEL_BASE_URL` with `sync: false` and MUST NOT give that key a `value`. It MUST list `JWT_SECRET` and `MORPH_AI_API_KEY` with `sync: false` and no `value`. The committed Blueprint MUST NOT contain a usable JWT, password, or API key for this service. It MUST NOT set `VITE_DATAX_URL` on any service.

#### Scenario: Morph auth URL is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `sharpreport`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not set
- **WHEN** a reviewer reads env keys in `render.yaml`
- **THEN** `VITE_DATAX_URL` is not one of them

### Requirement: Build filter tracks the Data Access image inputs
The `sharpreport` `buildFilter.paths` MUST include `SharpReport/**`, `pkg/morphai-rs/**`, `SharpReport/Dockerfile`, `SharpReport/deploy/docker-entrypoint.sh`, `.dockerignore`, and `render.yaml`.

#### Scenario: A Dockerfile input is listed
- **WHEN** a reviewer compares the `sharpreport` build filter to the Data Access Dockerfile `COPY` lines
- **THEN** the SharpReport tree and `pkg/morphai-rs` are covered
- **AND** `render.yaml` and `.dockerignore` are listed as well

### Requirement: The runbook records the public URL placeholder
`deploy/README.md` MUST tell the product owner to create the `sharpreport` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. It MUST say to set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and that this value is not a secret. It MUST tell the product owner to copy the service's public HTTPS URL and MUST use the placeholder `https://<sharpreport public host>` for the value story #114 sets as `VITE_DATAX_URL` on MorphUtils. It MUST NOT set that variable in this change. It MUST document the `/data` disk.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Data Access Render section in `deploy/README.md`
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<sharpreport public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `VITE_DATAX_URL` on MorphUtils
