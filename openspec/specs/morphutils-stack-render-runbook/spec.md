# morphutils-stack-render-runbook Specification

## Purpose

Gives the product owner one ordered Render create path for the MorphUtils stack so each public HTTPS origin is recorded before Morph is rebuilt with that origin.

## Requirements

### Requirement: Forge does not create Render services
`deploy/README.md` MUST state that Forge must not create services on Render. It MUST state that the product owner creates the services and that this repository does not call Render. The runbook MUST name Render project `prj-dahc33dbedkc73a1v8n0`.

#### Scenario: A reader sees the ownership rule
- **WHEN** a reader opens the MorphUtils stack create order in `deploy/README.md`
- **THEN** they see that Forge must not create services on Render
- **AND** they see that the product owner creates the services in project `prj-dahc33dbedkc73a1v8n0`
- **AND** they see that this repository does not call Render

### Requirement: Create order records each stack URL before the Morph rebuild
The create order MUST name the Blueprint file `render.yaml` on branch `main` and the services `morph-utils` (MorphUtils), `formx` (Event Logs), `composerx` (Content Maker), `sharpreport` (Data Access), and `morph-engi` (Project). It MUST tell the product owner to copy each service's public HTTPS origin from the dashboard after that service is live. It MUST place the Morph rebuild after the MorphUtils origin has been copied. That rebuild MUST set `REACT_APP_MORPH_UTILS_URL` to `https://<morph-utils public host>` and MUST say the value is a Docker build argument inlined when the Morph image is built, so a restart without a rebuild does not add the header link. The runbook MUST NOT put a value for `REACT_APP_MORPH_UTILS_URL` in `render.yaml`.

#### Scenario: Product owner follows the numbered order
- **WHEN** a product owner follows the create order
- **THEN** they can create MorphUtils, Event Logs, Content Maker, Data Access, and Project from the Blueprint
- **AND** they record each public HTTPS origin before rebuilding Morph
- **AND** the Morph rebuild uses `REACT_APP_MORPH_UTILS_URL` set to `https://<morph-utils public host>`

#### Scenario: Restart is not treated as the Morph wire-up
- **WHEN** a reader looks up how `REACT_APP_MORPH_UTILS_URL` takes effect
- **THEN** the runbook says the Morph image must be rebuilt
- **AND** it says a restart without that rebuild leaves the header link unset

### Requirement: Env matrix has placeholders for the wire-up origins
The create order MUST include an env matrix the product owner can fill. The matrix MUST include a placeholder for the MorphUtils origin used as `REACT_APP_MORPH_UTILS_URL`, and a placeholder for each embed origin: `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) as `https://<event-logs public host>`, `VITE_COMPOSERX_URL` as `https://<composerx public host>`, `VITE_DATAX_URL` as `https://<sharpreport public host>`, and `VITE_PROJECTS_URL` (alias `VITE_MORPH_ENGI_URL`) as `https://<morph-engi public host>`. The matrix MUST say those `VITE_*` keys stay unset on `morph-utils` in this change and are applied by story #114. It MUST say those `VITE_*` keys are read when the MorphUtils container starts, so a later change of them does not require a MorphUtils image rebuild.

#### Scenario: Product owner fills the matrix
- **WHEN** a product owner fills the env matrix
- **THEN** there is a placeholder for the MorphUtils URL
- **AND** there is a placeholder for each embed origin used by the wire-up story

### Requirement: Live hosts are examples and secrets stay out of git
The runbook MUST list these hosts the product owner has already created, and MUST label each as an example that is not a requirement to recreate: `https://morph-utils.onrender.com`, `https://formx-vucj.onrender.com`, `https://composerx.onrender.com`, `https://sharpreport.onrender.com`, `https://morph-engi.onrender.com`, and Morph panel `https://morph-gjmb.onrender.com`. It MUST say to copy the origin the dashboard shows for the service being recorded. It MUST say not to guess an `onrender.com` host from the service name. It MUST NOT commit a JWT, password, or API key. It MUST say that a Blueprint create can drop nested env vars, and that the product owner sets `USERS_PANEL_BASE_URL` in the dashboard afterward when it is missing, to `https://<morph public host>` with no path.

#### Scenario: Examples are not recreate instructions
- **WHEN** a reader sees an `onrender.com` host in the create order
- **THEN** that host is marked as an example
- **AND** the recorded origin is the one the dashboard shows

#### Scenario: Auth base URL may be set after create
- **WHEN** `USERS_PANEL_BASE_URL` is missing after a Blueprint create
- **THEN** the runbook tells the product owner to set it in the dashboard to `https://<morph public host>`
- **AND** no secret value is written into the repository
