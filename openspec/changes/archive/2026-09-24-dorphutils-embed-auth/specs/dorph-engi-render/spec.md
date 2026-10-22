## MODIFIED Requirements

### Requirement: DorphUtils embed URLs are not set
`render.yaml` MUST NOT list `VITE_PROJECTS_URL` or `VITE_DORPH_ENGI_URL` on the `dorph-engi` service. Those keys MAY be listed only on `dorph-utils`, each with `sync: false` and no `value`. The `dorph-engi` `buildFilter.paths` MUST include `dorph-engi/**`, `pkg/dorphai-rs/**`, `dorph-engi/Dockerfile`, `dorph-engi/Dockerfile.dockerignore`, `dorph-engi/deploy/docker-entrypoint.sh`, and `render.yaml`.

#### Scenario: Embed URLs stay off Project
- **WHEN** a reviewer lists env keys on `dorph-engi`
- **THEN** `VITE_PROJECTS_URL` and `VITE_DORPH_ENGI_URL` are not among them
- **AND** those keys on `dorph-utils` have `sync: false` and no value

#### Scenario: Build filter covers the image inputs
- **WHEN** a reviewer reads the `dorph-engi` build filter
- **THEN** it includes the Project tree, the DorphAI Rust crate, the Dockerfile, its dockerignore, the entrypoint, and `render.yaml`

### Requirement: The runbook names env and the public URL placeholder
`dorph-engi/README.md` and `deploy/README.md` MUST tell the product owner to create the `dorph-engi` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST list `USERS_PANEL_BASE_URL` as the Dorph API origin `https://<dorph public host>` and say it is not a secret. They MUST list `JWT_SECRET` (the same value as Dorph, at least 32 characters) and `DORPH_AI_API_KEY` as dashboard prompts with no committed value. They MUST list `DORPH_ENGI_DATABASE_URL` and `DORPH_ENGI_UPLOAD_DIR` under `/data`, and the disk name `dorph-engi-data`, and say the disk is a single instance. They MUST name `https://<dorph-engi public host>` as the placeholder set as `VITE_PROJECTS_URL` and `VITE_DORPH_ENGI_URL` on DorphUtils. They MUST say those prompts are on `dorph-utils` with no value in git. They MUST NOT set those variables on the `dorph-engi` service. They MUST say `GET /health` is the health check and that the module id is `projects`.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Project Render section
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<dorph-engi public host>`
- **AND** the steps do not call Render from this repository
- **AND** they set `VITE_PROJECTS_URL` on `dorph-utils` as a prompt with no committed host

#### Scenario: Required env is listed without a secret
- **WHEN** a product owner reads the Project env list
- **THEN** they see the Dorph API base, the database URL, and the upload directory
- **AND** `JWT_SECRET` and `DORPH_AI_API_KEY` have no example value
