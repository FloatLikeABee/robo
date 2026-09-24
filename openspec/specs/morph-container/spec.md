# morph-container Specification

## Purpose

Packages Morph API and the Morph AI UI as one runnable image so an operator can start that stack from compose, keep its data on a single volume, and supply secrets only through the environment.

## Requirements

### Requirement: One image serves Morph API and the Morph AI UI
The repository MUST provide a container image that contains the Morph API binary and the production build of the Morph AI UI in `morph/frontend`. That image MUST be the only application process for this capability. It MUST NOT include Event Logs, Content Maker, Data Access, Project, AI tools, MorphUtils, or Invite Signup. The process MUST listen on `PORT` (default 9090) on all interfaces. `GET /health` MUST return HTTP 200. `GET /` MUST return the UI `index.html`.

#### Scenario: Health and UI from the same process
- **WHEN** the image is running with a reachable `PORT`
- **THEN** `GET /health` returns HTTP 200
- **AND** `GET /` returns the Morph AI UI document

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
- **WHEN** Morph is started from a checkout without the image env defaults
- **THEN** the database paths still default under `./data`

### Requirement: Secrets come from the environment only
The image build MUST NOT copy `.env` files or accept secrets as build arguments. Production mode (`MORPH_ENV=production`) MUST still refuse to serve HTTP when `JWT_SECRET` or `ADMIN_PASSWORD` is a development default. Compose MUST load configuration from a gitignored env file. A committed example MUST list the required variable names and MUST NOT contain a usable secret.

#### Scenario: Default secrets do not boot in production
- **WHEN** the container starts with `MORPH_ENV=production` and the development JWT secret or development admin password
- **THEN** the process exits before it serves HTTP
- **AND** the log names the variable to set and does not print the secret

#### Scenario: Example env has no real secret
- **WHEN** a reviewer reads the committed production env example
- **THEN** `JWT_SECRET` and `ADMIN_PASSWORD` are empty or placeholders that production startup rejects

### Requirement: One command brings the stack up
Compose MUST start the Morph image with a named volume on `/data` and an `env_file` for the gitignored production env. `docker compose up` without an extra profile MUST NOT start a TLS proxy. TLS termination MAY be present as an optional compose profile so a later change can finish it. The compose file MUST NOT hardcode a production hostname.

#### Scenario: Default compose is the app only
- **WHEN** an operator runs `docker compose up` with the documented env file and no profile
- **THEN** the Morph process starts
- **AND** no TLS reverse proxy starts

### Requirement: Operators have a short runbook
`deploy/README.md` MUST describe how to build, run, set the env file, mount `/data`, back up `/data`, and upgrade the image. `docs/agents/12-build-deploy.md` MUST point at that runbook. The runbook MUST NOT tell the operator to deploy to a specific host.

#### Scenario: Reader follows the pointer
- **WHEN** a reader opens the production section of `docs/agents/12-build-deploy.md`
- **THEN** it links to `deploy/README.md` for the Morph image

### Requirement: Image build is checked without a new required check name
CI MUST build the image on pull requests and on pushes to `main`. That build MUST NOT add or rename any of the five required check names (`Go / Morph API`, `Go / Event Logs`, `Go / Content Maker`, `Go / morphai`, `Morph frontend`). The platform CI workflow MUST still report exactly those five names.

#### Scenario: Image workflow is separate
- **WHEN** a pull request is opened
- **THEN** an image build runs
- **AND** the platform CI workflow still reports only the five existing check names
