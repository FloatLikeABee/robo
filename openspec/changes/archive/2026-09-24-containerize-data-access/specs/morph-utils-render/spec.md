## MODIFIED Requirements

### Requirement: MorphUtils is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `morph-utils` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./morph-utils/Dockerfile`, `dockerContext: ./morph-utils`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The service list MUST include `morph`, then `morph-utils`, then `sharpreport`. Other services MAY appear when a sibling change adds them. The `morph-utils` fields above stay on the `morph-utils` service.

#### Scenario: Blueprint matches the MorphUtils image
- **WHEN** a reviewer reads the `morph-utils` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `morph-utils/Dockerfile` with context `morph-utils/`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

#### Scenario: Data Access is declared beside the shell
- **WHEN** a reviewer lists service names in `render.yaml`
- **THEN** `morph` and `morph-utils` are present in that order
- **AND** `sharpreport` is present
- **AND** `morph-utils` is still the shell service

### Requirement: The runbook creates the shell and records the public URL
`deploy/README.md` and `morph-utils/README.md` MUST tell the product owner to create the `morph-utils` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `VITE_MORPH_API_URL` to `https://<morph public host>` and that this value is not a secret. They MUST document `VITE_USERS_PANEL_API_URL` as the optional alias, and `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_PROJECTS_URL`, `VITE_MORPH_ENGI_URL`, and `VITE_MORPH_AI_URL` as optional embed origins. They MUST document `VITE_DATAX_URL` as the MorphUtils variable for the Data Access origin and MUST NOT set it on `morph-utils`. They MUST state that Morph already allows cross-origin `Authorization` for non-loopback origins and that this change does not edit CORS. They MUST tell the product owner to copy the shell's public HTTPS URL and MUST use the placeholder `https://<morph-utils public host>` for the value story #106 sets as `REACT_APP_MORPH_UTILS_URL` on Morph. They MUST NOT set that variable in this change. They MUST use the placeholder `https://<sharpreport public host>` for the value story #114 sets as `VITE_DATAX_URL`.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the MorphUtils Render section in `deploy/README.md`
- **THEN** they can create the shell service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<morph-utils public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `REACT_APP_MORPH_UTILS_URL` on Morph
- **AND** they do not set `VITE_DATAX_URL` on MorphUtils

#### Scenario: Auth env is listed without a secret
- **WHEN** a product owner reads the MorphUtils env list
- **THEN** they see `PORT` `3040`, the required Morph origin, the optional alias, and the optional embed variables
- **AND** they see that Morph already allows cross-origin `Authorization` and that CORS is unchanged
