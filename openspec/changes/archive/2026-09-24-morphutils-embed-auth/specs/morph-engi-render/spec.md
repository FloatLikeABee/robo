## MODIFIED Requirements

### Requirement: MorphUtils embed URLs are not set
`render.yaml` MUST NOT list `VITE_PROJECTS_URL` or `VITE_MORPH_ENGI_URL` on the `morph-engi` service. Those keys MAY be listed only on `morph-utils`, each with `sync: false` and no `value`. The `morph-engi` `buildFilter.paths` MUST include `morph-engi/**`, `pkg/morphai-rs/**`, `morph-engi/Dockerfile`, `morph-engi/Dockerfile.dockerignore`, `morph-engi/deploy/docker-entrypoint.sh`, and `render.yaml`.

#### Scenario: Embed URLs stay off Project
- **WHEN** a reviewer lists env keys on `morph-engi`
- **THEN** `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` are not among them
- **AND** those keys on `morph-utils` have `sync: false` and no value

#### Scenario: Build filter covers the image inputs
- **WHEN** a reviewer reads the `morph-engi` build filter
- **THEN** it includes the Project tree, the MorphAI Rust crate, the Dockerfile, its dockerignore, the entrypoint, and `render.yaml`

### Requirement: The runbook names env and the public URL placeholder
`morph-engi/README.md` and `deploy/README.md` MUST tell the product owner to create the `morph-engi` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST list `USERS_PANEL_BASE_URL` as the Morph API origin `https://<morph public host>` and say it is not a secret. They MUST list `JWT_SECRET` (the same value as Morph, at least 32 characters) and `MORPH_AI_API_KEY` as dashboard prompts with no committed value. They MUST list `MORPH_ENGI_DATABASE_URL` and `MORPH_ENGI_UPLOAD_DIR` under `/data`, and the disk name `morph-engi-data`, and say the disk is a single instance. They MUST name `https://<morph-engi public host>` as the placeholder set as `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` on MorphUtils. They MUST say those prompts are on `morph-utils` with no value in git. They MUST NOT set those variables on the `morph-engi` service. They MUST say `GET /health` is the health check and that the module id is `projects`.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Project Render section
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<morph-engi public host>`
- **AND** the steps do not call Render from this repository
- **AND** they set `VITE_PROJECTS_URL` on `morph-utils` as a prompt with no committed host

#### Scenario: Required env is listed without a secret
- **WHEN** a product owner reads the Project env list
- **THEN** they see the Morph API base, the database URL, and the upload directory
- **AND** `JWT_SECRET` and `MORPH_AI_API_KEY` have no example value
