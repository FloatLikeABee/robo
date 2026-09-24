# data-access-container Specification

## Purpose

Packages the Data Access API and UI as one image an operator can start locally, with health checks that do not call Morph and with secrets only from the environment.

## Requirements

### Requirement: One image serves the Data Access API and UI
The repository MUST provide a container image that contains the Data Access API binary and the production build of the Data Access UI. That image MUST be the only application process for this capability. It MUST NOT include Event Logs, Content Maker, Project, AI tools, Morph, MorphUtils, or Invite Signup. The process MUST listen on all interfaces. The listen port MUST be `SHARPREPORT_PORT` when that variable is non-empty, otherwise `PORT`, otherwise `3050`. `GET /health` and `GET /ready` MUST return HTTP 200 and a JSON body whose `status` is `ok`, and MUST NOT call Morph or MorphUtils. When the UI directory is set and contains `index.html`, `GET /` MUST return that UI document without authentication. `GET /api/v1/health` MUST still return HTTP 200. A local checkout that does not set the UI directory MUST keep the debug redirect from `GET /` to the Vite dev server.

#### Scenario: Health and ready do not call Morph
- **WHEN** the image is running and Morph is not reachable
- **THEN** `GET /health` returns HTTP 200 and `{"status":"ok"}`
- **AND** `GET /ready` returns HTTP 200 and `{"status":"ok"}`

#### Scenario: UI and API share the listen port
- **WHEN** the image is running with the UI directory set
- **THEN** `GET /` returns the Data Access UI document
- **AND** the Vite dev port 5178 is not published by the image

### Requirement: Local port selection prefers SHARPREPORT_PORT
When both `SHARPREPORT_PORT` and `PORT` are set, the API MUST listen on `SHARPREPORT_PORT`. Local `start-all.sh` MUST keep the API on 3050 and the Vite UI on 5178.

#### Scenario: Root env PORT does not steal the API port
- **WHEN** `PORT` is `9090` and `SHARPREPORT_PORT` is `3050`
- **THEN** the API listens on `3050`

### Requirement: SQLite lives on one mount
The image MUST default the Data Access SQLite file to `/data/datapulse.db`. A named volume mounted at `/data` MUST be enough for that file to survive a container recreate. An absolute `sqlite://` URL MUST stay on that absolute path. Local `start-all.sh` defaults (`sqlite://./data/datapulse.db` relative to the working directory) MUST stay unchanged when the image env defaults are not set. Metabase MUST NOT be started by the image.

#### Scenario: Database file is on the volume
- **WHEN** the image starts with the default database URL and `/data` is a named volume
- **THEN** the SQLite file is created under `/data`

#### Scenario: Local checkout paths stay relative
- **WHEN** Data Access is started from a checkout without the image database URL
- **THEN** the database path still defaults under `./data`

### Requirement: Secrets come from the environment only
The image build MUST NOT copy `.env` files or accept secrets as build arguments. Compose MUST load configuration from a gitignored env file. A committed example MUST list `USERS_PANEL_BASE_URL` and `JWT_SECRET` with empty values. The example MUST NOT contain a usable JWT or database password. `GET /health` MUST succeed when those values are empty.

#### Scenario: Example env has no real secret
- **WHEN** a reviewer reads the committed production env example
- **THEN** `JWT_SECRET` and `USERS_PANEL_BASE_URL` are empty

#### Scenario: Health does not need the Morph origin
- **WHEN** the container starts without `USERS_PANEL_BASE_URL`
- **THEN** `GET /health` still returns HTTP 200

### Requirement: Operators have a short runbook
`SharpReport/README.md` and `docs/agents/07-rust-apps.md` MUST describe how to build and run the image, which env vars the process reads, that `USERS_PANEL_BASE_URL` is the Morph auth base URL, and that `GET /health` does not call Morph. `docs/agents/12-build-deploy.md` MUST point at that image. The runbook MUST NOT tell the operator to follow `scripts/deploy.sh`.

#### Scenario: Reader finds the Morph auth base URL
- **WHEN** a reader opens the Data Access deploy notes
- **THEN** they see `USERS_PANEL_BASE_URL` documented as the Morph auth base URL with no invented host
- **AND** the steps do not run `scripts/deploy.sh`

### Requirement: Startup can read SQL migrations from the working directory
The image MUST contain the Data Access SQL migrations at `migrations` relative to the process working directory `/app` (`/app/migrations`). Those files MUST be the SQL files from `SharpReport/backend/migrations`. The container contract MUST fail when the image does not copy that directory to `/app/migrations`, or when that source directory has no `.sql` file. Startup MUST keep reading that relative directory. The contract MUST NOT require a value for `USERS_PANEL_BASE_URL`.

#### Scenario: Image includes the migrations directory
- **WHEN** a reviewer reads the Data Access Dockerfile and `SharpReport/backend/migrations`
- **THEN** the Dockerfile copies that directory to `/app/migrations`
- **AND** the source directory contains at least one `.sql` file
- **AND** the container contract fails if either of those is missing
