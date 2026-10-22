# dorph-utils-container Specification

## Purpose

Packages the DorphUtils Vite shell as its own image so an operator can serve the production build, check health without the embedded apps, and point auth at Dorph using environment configuration only.

## Requirements

### Requirement: One image serves the DorphUtils shell
The repository MUST provide a container image that serves the production build of `dorph-utils/frontend` over HTTP. The process MUST listen on `PORT` (default 3040) on all interfaces. `GET /health` and `GET /ready` MUST return HTTP 200 without contacting Dorph, Event Logs, Content Maker, Data Access, or Project. `GET /` MUST return the DorphUtils document. The image MUST NOT include those other applications.

#### Scenario: Health and ready succeed with no sibling services
- **WHEN** the image is running and no other application container is running
- **THEN** `GET /health` returns HTTP 200
- **AND** `GET /ready` returns HTTP 200
- **AND** `GET /` returns the DorphUtils HTML document

#### Scenario: Other apps are not in the image
- **WHEN** an operator inspects the image build context and the files copied into the runtime stage
- **THEN** the image build does not copy Event Logs, Content Maker, Data Access, Project, AI tools, Dorph API, or Invite Signup

### Requirement: Dorph API origin is configurable without a localhost production default
The shell MUST call Dorph auth (`/api/auth/user` and the other `/api/auth/` paths it already uses) at the configured Dorph API origin. `VITE_DORPH_API_URL` is the primary origin. `VITE_USERS_PANEL_API_URL` is the alias used only when the primary is unset or blank. A non-empty value supplied when the container starts MUST take effect without rebuilding the image. When both are unset or blank, auth requests MUST be same-origin paths and MUST NOT target a loopback host. Local `npm run dev` MUST keep the same-origin `/api` proxy when those variables are unset.

#### Scenario: Runtime origin is used for auth
- **WHEN** the container is started with `VITE_DORPH_API_URL` set to a non-loopback `https` origin
- **THEN** the configuration the shell reads for Dorph auth is that origin
- **AND** that origin is not `localhost` or `127.0.0.1`

#### Scenario: Alias is used when the primary is blank
- **WHEN** `VITE_DORPH_API_URL` is unset or blank and `VITE_USERS_PANEL_API_URL` is a non-loopback origin
- **THEN** the shell uses `VITE_USERS_PANEL_API_URL` as the Dorph API origin

#### Scenario: Unset origin stays same-origin
- **WHEN** the container is started with both Dorph API variables unset or blank
- **THEN** auth URLs are same-origin paths beginning with `/api/`
- **AND** they do not use a loopback host

### Requirement: Embed origins are optional public URLs
Event Logs, Content Maker, Data Access, Project, and the Dorph AI link MUST use these variables when set: `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`), `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL` (alias `VITE_DORPH_ENGI_URL`), and `VITE_DORPH_AI_URL`. A non-empty value supplied when the container starts MUST take effect without rebuilding. The production image, with those variables unset, MUST NOT point those modules at loopback hosts. Local `npm run dev` with the variables unset MUST keep the current localhost defaults. Health checks MUST NOT require any embed to be up.

#### Scenario: Shell-only image does not call loopback embeds
- **WHEN** the image is built and started with no embed variables set
- **THEN** the served production assets do not contain the localhost embed defaults
- **AND** `GET /health` still returns HTTP 200

#### Scenario: Dev server defaults stay on localhost
- **WHEN** a developer runs `npm run dev` in `dorph-utils/frontend` without those variables
- **THEN** Event Logs, Content Maker, Data Access, Project, and the Dorph AI link still default to their current localhost ports

### Requirement: Secrets are not baked into the image
The image build MUST NOT copy `.env` files and MUST NOT accept secrets (JWT, passwords, or API keys) as build arguments. Public URL variables MAY be build arguments. A committed example MUST NOT contain a secret value. Operator documentation MUST tell the operator to pass configuration with the environment.

#### Scenario: Secret build args are absent
- **WHEN** a reviewer reads the DorphUtils Dockerfile
- **THEN** it has no build argument for a JWT, password, or API key

#### Scenario: Env files are not in the build context
- **WHEN** the image is built from the documented context
- **THEN** `.env` files are excluded from that context

### Requirement: Operators can build and run the shell alone
`dorph-utils/README.md` MUST describe how to build the image, run it, set `PORT`, and set the public URL variables from the environment. The documented commands MUST NOT require Render, the root Dorph Dockerfile, or `render.yaml`.

#### Scenario: Runbook is enough to start the shell
- **WHEN** an operator follows the DorphUtils deploy section with only this repository and a container engine
- **THEN** they can build the image and open `/health` without creating a Render service
