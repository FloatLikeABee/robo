# Spec Delta

## MODIFIED Requirements

### Requirement: The Morph origin is prompted and not invented
The `morph-utils` service MUST list `VITE_MORPH_API_URL` with `sync: false` and MUST NOT give that key a `value`. The committed Blueprint MUST NOT contain a JWT, password, or API key for this service. The `morph-utils` service MUST NOT list `REACT_APP_MORPH_UTILS_URL` as an env key.

#### Scenario: Morph origin is a dashboard prompt
- **WHEN** a reviewer reads `VITE_MORPH_API_URL` on `morph-utils`
- **THEN** it has `sync: false` and no value

#### Scenario: Morph header URL is not set here
- **WHEN** a reviewer reads env keys on `morph-utils`
- **THEN** `REACT_APP_MORPH_UTILS_URL` is not one of them

### Requirement: The runbook creates the shell and records the public URL
`deploy/README.md` and `morph-utils/README.md` MUST tell the product owner to create the `morph-utils` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `VITE_MORPH_API_URL` to `https://<morph public host>` and that this value is not a secret. They MUST document `VITE_USERS_PANEL_API_URL` as the optional alias. They MUST document `VITE_SHEETX_URL` and `VITE_FORMSX_URL` as the Event Logs origin variables and MUST NOT set them on the `morph-utils` service. They MUST document `VITE_COMPOSERX_URL` as the Content Maker embed origin and MUST state that this change does not set it on `morph-utils`. They MUST document `VITE_DATAX_URL` as the Data Access origin and MUST NOT set it on `morph-utils`. They MUST use the placeholder `https://<sharpreport public host>` for the value story #114 sets as `VITE_DATAX_URL`. They MUST document `VITE_MORPH_AI_URL` as an optional embed origin. `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` MUST remain unset on the `morph-utils` service. The docs MUST name `https://<morph-engi public host>` as the placeholder story #114 sets on those two variables, and MUST NOT set them in this Blueprint. They MUST state that Morph already allows cross-origin `Authorization` for non-loopback origins and that this change does not edit CORS. They MUST tell the product owner to copy the shell's public HTTPS URL and MUST use the placeholder `https://<morph-utils public host>` for the value set as `REACT_APP_MORPH_UTILS_URL` on Morph. They MUST NOT set that variable on the `morph-utils` service. They MUST say the `morph` service lists that key as a dashboard prompt with no value, and that the product owner fills the prompt and rebuilds Morph.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the MorphUtils Render section in `deploy/README.md`
- **THEN** they can create the shell service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<morph-utils public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `REACT_APP_MORPH_UTILS_URL` on the `morph-utils` service
- **AND** they do not set `VITE_DATAX_URL` on MorphUtils

#### Scenario: Auth env is listed without a secret
- **WHEN** a product owner reads the MorphUtils env list
- **THEN** they see `PORT` `3040`, the required Morph origin, the optional alias, and the optional embed variables
- **AND** they see that Morph already allows cross-origin `Authorization` and that CORS is unchanged
- **AND** they see that `VITE_COMPOSERX_URL` is not set on `morph-utils` in this change

#### Scenario: Project embed URL is a later placeholder
- **WHEN** a product owner reads the MorphUtils embed list
- **THEN** `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` are unset
- **AND** the documented placeholder for story #114 is `https://<morph-engi public host>`
