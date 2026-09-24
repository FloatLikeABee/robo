# morph-utils-render Specification

## Purpose

Gives the product owner a Blueprint entry for the MorphUtils shell and the steps to copy its public HTTPS URL, without this repository calling Render.

## Requirements

### Requirement: MorphUtils is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `morph-utils` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./morph-utils/Dockerfile`, `dockerContext: ./morph-utils`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The service names in that file MUST be only `morph` and `morph-utils`.

#### Scenario: Blueprint matches the MorphUtils image
- **WHEN** a reviewer reads the `morph-utils` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `morph-utils/Dockerfile` with context `morph-utils/`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

#### Scenario: No other product services are declared
- **WHEN** a reviewer lists service names in `render.yaml`
- **THEN** the names are `morph` and `morph-utils`
- **AND** Event Logs, Content Maker, Data Access, and Project are not services in that file

### Requirement: The shell has no disk and a pinned port
The `morph-utils` service MUST NOT declare a `disk`. It MUST set `PORT` to `3040`. It MUST NOT set `numInstances`.

#### Scenario: Port matches the image default
- **WHEN** a reviewer reads the `PORT` entry on `morph-utils`
- **THEN** its value is `3040`
- **AND** that service has no disk

### Requirement: The Morph origin is prompted and not invented
The `morph-utils` service MUST list `VITE_MORPH_API_URL` with `sync: false` and MUST NOT give that key a `value`. The committed Blueprint MUST NOT contain a JWT, password, or API key for this service. It MUST NOT list `REACT_APP_MORPH_UTILS_URL` as an env key.

#### Scenario: Morph origin is a dashboard prompt
- **WHEN** a reviewer reads `VITE_MORPH_API_URL` on `morph-utils`
- **THEN** it has `sync: false` and no value

#### Scenario: Morph header URL is not set here
- **WHEN** a reviewer reads env keys in `render.yaml`
- **THEN** `REACT_APP_MORPH_UTILS_URL` is not one of them

### Requirement: Build filter tracks the MorphUtils image inputs
The `morph-utils` `buildFilter.paths` MUST include `morph-utils/frontend/**`, `morph-utils/Dockerfile`, `morph-utils/.dockerignore`, `morph-utils/deploy/docker-entrypoint.sh`, `morph-utils/deploy/nginx.conf.template`, and `render.yaml`.

#### Scenario: A Dockerfile input is listed
- **WHEN** a reviewer compares the `morph-utils` build filter to the MorphUtils Dockerfile `COPY` lines
- **THEN** the frontend tree, the Dockerfile, the entrypoint, and the nginx template are covered
- **AND** `render.yaml` and `morph-utils/.dockerignore` are listed as well

### Requirement: The runbook creates the shell and records the public URL
`deploy/README.md` and `morph-utils/README.md` MUST tell the product owner to create the `morph-utils` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `VITE_MORPH_API_URL` to `https://<morph public host>` and that this value is not a secret. They MUST document `VITE_USERS_PANEL_API_URL` as the optional alias, and `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, `VITE_MORPH_ENGI_URL`, and `VITE_MORPH_AI_URL` as optional embed origins that are not services in this Blueprint. They MUST state that Morph already allows cross-origin `Authorization` for non-loopback origins and that this change does not edit CORS. They MUST tell the product owner to copy the service's public HTTPS URL and MUST use the placeholder `https://<morph-utils public host>` for the value story #106 sets as `REACT_APP_MORPH_UTILS_URL` on Morph. They MUST NOT set that variable in this change.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the MorphUtils Render section in `deploy/README.md`
- **THEN** they can create the shell service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<morph-utils public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `REACT_APP_MORPH_UTILS_URL` on Morph

#### Scenario: Auth env is listed without a secret
- **WHEN** a product owner reads the MorphUtils env list
- **THEN** they see `PORT` `3040`, the required Morph origin, the optional alias, and the optional embed variables
- **AND** they see that Morph already allows cross-origin `Authorization` and that CORS is unchanged
