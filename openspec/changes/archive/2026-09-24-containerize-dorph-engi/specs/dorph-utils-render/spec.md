## MODIFIED Requirements

### Requirement: DorphUtils is a Docker web service in the flat Blueprint
`render.yaml` MUST declare a service named `dorph-utils` with `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, and `branch: main`. That service MUST set `dockerfilePath: ./dorph-utils/Dockerfile`, `dockerContext: ./dorph-utils`, `healthCheckPath: /health`, and `autoDeployTrigger: checksPass`. The Blueprint MUST NOT declare a `projects` block. The service list MUST begin with `dorph` then `dorph-utils`. Further services MAY follow, including `dorph-engi`. The `dorph-utils` service itself MUST stay a shell with no product embed URL set.

#### Scenario: Blueprint matches the DorphUtils image
- **WHEN** a reviewer reads the `dorph-utils` service in `render.yaml`
- **THEN** it is a Docker web service in `singapore` on the `starter` plan, tracking `main`, built from `dorph-utils/Dockerfile` with context `dorph-utils/`
- **AND** its health check path is `/health` and it auto-deploys only after checks pass

#### Scenario: Further services may follow the shell
- **WHEN** a reviewer lists service names in `render.yaml`
- **THEN** the first two names are `dorph` and `dorph-utils`
- **AND** further names are allowed, including `dorph-engi`
- **AND** the file still has no `projects` block

### Requirement: The runbook creates the shell and records the public URL
`deploy/README.md` and `dorph-utils/README.md` MUST tell the product owner to create the `dorph-utils` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `VITE_DORPH_API_URL` to `https://<dorph public host>` and that this value is not a secret. They MUST document `VITE_USERS_PANEL_API_URL` as the optional alias, and `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, and `VITE_DORPH_AI_URL` as optional embed origins. `VITE_PROJECTS_URL` and `VITE_DORPH_ENGI_URL` MUST remain unset on the `dorph-utils` service. The docs MUST name `https://<dorph-engi public host>` as the placeholder story #114 sets on those two variables, and MUST NOT set them in this Blueprint. They MUST state that Dorph already allows cross-origin `Authorization` for non-loopback origins and that this change does not edit CORS. They MUST tell the product owner to copy the service's public HTTPS URL and MUST use the placeholder `https://<dorph-utils public host>` for the value story #106 sets as `REACT_APP_DORPH_UTILS_URL` on Dorph. They MUST NOT set that variable in this change.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the DorphUtils Render section in `deploy/README.md`
- **THEN** they can create the shell service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<dorph-utils public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `REACT_APP_DORPH_UTILS_URL` on Dorph

#### Scenario: Auth env is listed without a secret
- **WHEN** a product owner reads the DorphUtils env list
- **THEN** they see `PORT` `3040`, the required Dorph origin, the optional alias, and the optional embed variables
- **AND** they see that Dorph already allows cross-origin `Authorization` and that CORS is unchanged

#### Scenario: Project embed URL is a later placeholder
- **WHEN** a product owner reads the DorphUtils embed list
- **THEN** `VITE_PROJECTS_URL` and `VITE_DORPH_ENGI_URL` are unset
- **AND** the documented placeholder for story #114 is `https://<dorph-engi public host>`
