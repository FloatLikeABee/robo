# dorph-container Specification

## Purpose

Packages Dorph API and the Dorph AI UI as one runnable image so an operator can start that stack from compose, keep its data on a single volume, and supply secrets only through the environment.

## Requirements

### Requirement: One image serves Dorph API and the Dorph AI UI
The repository MUST provide a container image that contains the Dorph API binary and the production build of the Dorph AI UI in `dorph/frontend`. That image MUST be the only application process for this capability. It MUST NOT include Event Logs, Content Maker, Data Access, Project, AI tools, DorphUtils, or Invite Signup. The process MUST listen on `PORT` (default 9090) on all interfaces. `GET /health` MUST return HTTP 200. `GET /` MUST return the UI `index.html`.

#### Scenario: Health and UI from the same process
- **WHEN** the image is running with a reachable `PORT`
- **THEN** `GET /health` returns HTTP 200
- **AND** `GET /` returns the Dorph AI UI document

#### Scenario: Other apps are not in the image
- **WHEN** an operator inspects the image contents
- **THEN** the image does not contain the other application trees listed above

### Requirement: Data lives on one mount
The image MUST default these paths under `/data`: the Badger app database, the Tran SQLite file, the entity-details Badger store, the knowledge directory, and entity-attachment uploads. A named volume mounted at `/data` MUST be enough for those stores to survive a container recreate. Local `start-all.sh` defaults (`./data` relative to the working directory) MUST stay unchanged when the image env defaults are not set.

#### Scenario: Note survives recreate
- **WHEN** an authenticated client creates a note while `/data` is a named volume
- **AND** the container is removed and started again with the same volume and the same admin password
- **THEN** that note is still readable

#### Scenario: Local checkout paths stay relative
- **WHEN** Dorph is started from a checkout without the image env defaults
- **THEN** the database paths still default under `./data`

### Requirement: Secrets come from the environment only
The image build MUST NOT copy `.env` files or accept secrets as build arguments. Production mode (`DORPH_ENV=production`) MUST still refuse to serve HTTP when `JWT_SECRET` or `ADMIN_PASSWORD` is a development default. Compose MUST load configuration from a gitignored env file. A committed example MUST list the required variable names and MUST NOT contain a usable secret.

#### Scenario: Default secrets do not boot in production
- **WHEN** the container starts with `DORPH_ENV=production` and the development JWT secret or development admin password
- **THEN** the process exits before it serves HTTP
- **AND** the log names the variable to set and does not print the secret

#### Scenario: Example env has no real secret
- **WHEN** a reviewer reads the committed production env example
- **THEN** `JWT_SECRET` and `ADMIN_PASSWORD` are empty or placeholders that production startup rejects

### Requirement: One command brings the stack up
Compose MUST start the Dorph image with a named volume on `/data` and an `env_file` for the gitignored production env. `docker compose up` without an extra profile MUST NOT start a TLS proxy. TLS termination MAY be present as an optional compose profile so a later change can finish it. The compose file MUST NOT hardcode a production hostname.

#### Scenario: Default compose is the app only
- **WHEN** an operator runs `docker compose up` with the documented env file and no profile
- **THEN** the Dorph process starts
- **AND** no TLS reverse proxy starts

### Requirement: Operators have a short runbook
`deploy/README.md` MUST describe how to build, run, set the env file, mount `/data`, back up `/data`, and upgrade the image. `docs/agents/12-build-deploy.md` MUST point at that runbook. The runbook MUST document creating the Dorph service on Render from the root Blueprint, and MUST NOT tell the operator to follow `scripts/deploy.sh`.

#### Scenario: Reader follows the pointer
- **WHEN** a reader opens the production section of `docs/agents/12-build-deploy.md`
- **THEN** it links to `deploy/README.md` for the Dorph image and for the Render Blueprint

#### Scenario: Deploy script is not the runbook
- **WHEN** a reader follows `deploy/README.md` to host Dorph
- **THEN** the steps create the service from `render.yaml`
- **AND** they do not run `scripts/deploy.sh`

### Requirement: Image build is checked without a new required check name
CI MUST build the image on pull requests and on pushes to `main`. That build MUST NOT add or rename any of the five required check names (`Go / Dorph API`, `Go / Event Logs`, `Go / Content Maker`, `Go / dorphai`, `Dorph frontend`). The platform CI workflow MUST still report exactly those five names. The Dorph image workflow MUST validate `render.yaml` against the Render Blueprint schema in a step of that same workflow.

#### Scenario: Image workflow is separate
- **WHEN** a pull request is opened
- **THEN** an image build runs
- **AND** the platform CI workflow still reports only the five existing check names

#### Scenario: Blueprint schema is part of the image workflow
- **WHEN** the Dorph image workflow runs
- **THEN** it validates `render.yaml` against the Render Blueprint schema
- **AND** it does not add a job or check name to the platform CI workflow

### Requirement: DorphUtils URL is a non-secret build argument
The Dorph image build MUST accept `REACT_APP_DORPH_UTILS_URL` as a build argument. That argument MUST NOT have a default production host. The committed Dockerfile, compose file, and production env example MUST NOT contain a production DorphUtils origin. Compose MUST pass the argument from the shell, and an empty value MUST be allowed. The Dorph AI UI MUST NOT render a DorphUtils header chip from that argument.

#### Scenario: Unset build argument has no host
- **WHEN** a reviewer reads the Dorph Dockerfile build argument `REACT_APP_DORPH_UTILS_URL`
- **THEN** it has no default production host

#### Scenario: Compose does not bake a host
- **WHEN** a reviewer reads the Dorph service build in `deploy/docker-compose.yml`
- **THEN** `REACT_APP_DORPH_UTILS_URL` is a build argument
- **AND** the file does not hardcode a production hostname
