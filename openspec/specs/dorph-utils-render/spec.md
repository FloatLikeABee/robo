# dorph-utils-render Specification

## Purpose

Gives the product owner a Blueprint entry for the DorphUtils shell and the steps to copy its public HTTPS URL, without this repository calling Render.

## Requirements

### Requirement: DorphUtils is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `dorph-utils` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./dorph-utils/Dockerfile`, `dockerContext: ./dorph-utils`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The service names in that file MUST be `dorph`, `dorph-utils`, `formx`, `composerx`, `sharpreport`, and `dorph-engi`, in that order. The `dorph-utils` service itself MUST stay a shell with no product embed URL set. The `dorph-utils` fields above stay on the `dorph-utils` service.

#### Scenario: Blueprint matches the DorphUtils image
- **WHEN** a reviewer reads the `dorph-utils` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `dorph-utils/Dockerfile` with context `dorph-utils/`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

#### Scenario: Declared product services stay together
- **WHEN** a reviewer lists service names in `render.yaml`
- **THEN** the names are `dorph`, `dorph-utils`, `formx`, `composerx`, `sharpreport`, and `dorph-engi`
- **AND** the file still has no `projects` block
- **AND** `dorph-utils` is still the shell service

### Requirement: The shell has no disk and a pinned port
The `dorph-utils` service MUST NOT declare a `disk`. It MUST set `PORT` to `3040`. It MUST NOT set `numInstances`.

#### Scenario: Port matches the image default
- **WHEN** a reviewer reads the `PORT` entry on `dorph-utils`
- **THEN** its value is `3040`
- **AND** that service has no disk

### Requirement: The Dorph origin is prompted and not invented
The `dorph-utils` service MUST list `VITE_DORPH_API_URL` with `sync: false` and MUST NOT give that key a `value`. The committed Blueprint MUST NOT contain a JWT, password, or API key for this service. The `dorph-utils` service MUST NOT list `REACT_APP_DORPH_UTILS_URL` as an env key.

#### Scenario: Dorph origin is a dashboard prompt
- **WHEN** a reviewer reads `VITE_DORPH_API_URL` on `dorph-utils`
- **THEN** it has `sync: false` and no value

#### Scenario: Dorph header URL is not set here
- **WHEN** a reviewer reads env keys on `dorph-utils`
- **THEN** `REACT_APP_DORPH_UTILS_URL` is not one of them

### Requirement: Build filter tracks the DorphUtils image inputs
The `dorph-utils` `buildFilter.paths` MUST include `dorph-utils/frontend/**`, `dorph-utils/Dockerfile`, `dorph-utils/.dockerignore`, `dorph-utils/deploy/docker-entrypoint.sh`, `dorph-utils/deploy/nginx.conf.template`, and `render.yaml`.

#### Scenario: A Dockerfile input is listed
- **WHEN** a reviewer compares the `dorph-utils` build filter to the DorphUtils Dockerfile `COPY` lines
- **THEN** the frontend tree, the Dockerfile, the entrypoint, and the nginx template are covered
- **AND** `render.yaml` and `dorph-utils/.dockerignore` are listed as well

### Requirement: The runbook creates the shell and records the public URL
`deploy/README.md` and `dorph-utils/README.md` MUST tell the product owner to create the `dorph-utils` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `VITE_DORPH_API_URL` to `https://<dorph public host>` and that this value is not a secret. They MUST document `VITE_USERS_PANEL_API_URL` as the optional alias. They MUST document `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`), `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL` (alias `VITE_DORPH_ENGI_URL`), and `VITE_DORPH_AI_URL` as embed origins the product owner sets on `dorph-utils`. The placeholders MUST be `https://<event-logs public host>`, `https://<composerx public host>`, `https://<sharpreport public host>`, and `https://<dorph-engi public host>`. They MUST say those keys are dashboard prompts with no value in git, read when the container starts, so a restart rewrites `/config.js` and a DorphUtils image rebuild is not required. They MUST state that Dorph already allows cross-origin `Authorization` for non-loopback origins and that this change does not edit CORS. They MUST tell the product owner to copy the shell's public HTTPS URL and MUST use the placeholder `https://<dorph-utils public host>` for the value set as `REACT_APP_DORPH_UTILS_URL` on Dorph. They MUST NOT set that variable on the `dorph-utils` service. They MUST say the `dorph` service lists that key as a dashboard prompt with no value, and that the product owner fills the prompt and rebuilds Dorph.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the DorphUtils Render section in `deploy/README.md`
- **THEN** they can create the shell service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<dorph-utils public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `REACT_APP_DORPH_UTILS_URL` on the `dorph-utils` service
- **AND** they set `VITE_DATAX_URL` on `dorph-utils` as a prompt with no committed host

#### Scenario: Auth env is listed without a secret
- **WHEN** a product owner reads the DorphUtils env list
- **THEN** they see `PORT` `3040`, the required Dorph origin, the optional alias, and the embed origin prompts
- **AND** they see that Dorph already allows cross-origin `Authorization` and that CORS is unchanged
- **AND** they see that `VITE_COMPOSERX_URL` is a prompt on `dorph-utils` with no value in git

#### Scenario: Project embed URL is a prompt
- **WHEN** a product owner reads the DorphUtils embed list
- **THEN** `VITE_PROJECTS_URL` and `VITE_DORPH_ENGI_URL` are prompts with no committed host
- **AND** the documented placeholder is `https://<dorph-engi public host>`

### Requirement: Embed origins are prompts on the shell
The `dorph-utils` service MUST list `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, `VITE_DORPH_ENGI_URL`, and `VITE_DORPH_AI_URL` with `sync: false` and MUST NOT give those keys a `value`. The committed Blueprint MUST NOT contain an `onrender.com` host. Those keys MUST NOT appear on `dorph`, `formx`, `composerx`, `sharpreport`, or `dorph-engi`.

#### Scenario: Embed prompts have no host
- **WHEN** a reviewer reads the embed origin keys on `dorph-utils`
- **THEN** each has `sync: false` and no value
- **AND** `render.yaml` does not contain an `onrender.com` host
