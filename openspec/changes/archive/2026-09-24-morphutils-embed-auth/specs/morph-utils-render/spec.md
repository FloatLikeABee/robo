## MODIFIED Requirements

### Requirement: The runbook creates the shell and records the public URL
`deploy/README.md` and `morph-utils/README.md` MUST tell the product owner to create the `morph-utils` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `VITE_MORPH_API_URL` to `https://<morph public host>` and that this value is not a secret. They MUST document `VITE_USERS_PANEL_API_URL` as the optional alias. They MUST document `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`), `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL` (alias `VITE_MORPH_ENGI_URL`), and `VITE_MORPH_AI_URL` as embed origins the product owner sets on `morph-utils`. The placeholders MUST be `https://<event-logs public host>`, `https://<composerx public host>`, `https://<sharpreport public host>`, and `https://<morph-engi public host>`. They MUST say those keys are dashboard prompts with no value in git, read when the container starts, so a restart rewrites `/config.js` and a MorphUtils image rebuild is not required. They MUST state that Morph already allows cross-origin `Authorization` for non-loopback origins and that this change does not edit CORS. They MUST tell the product owner to copy the shell's public HTTPS URL and MUST use the placeholder `https://<morph-utils public host>` for the value set as `REACT_APP_MORPH_UTILS_URL` on Morph. They MUST NOT set that variable on the `morph-utils` service. They MUST say the `morph` service lists that key as a dashboard prompt with no value, and that the product owner fills the prompt and rebuilds Morph.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the MorphUtils Render section in `deploy/README.md`
- **THEN** they can create the shell service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<morph-utils public host>`
- **AND** the steps do not call Render from this repository
- **AND** they do not set `REACT_APP_MORPH_UTILS_URL` on the `morph-utils` service
- **AND** they set `VITE_DATAX_URL` on `morph-utils` as a prompt with no committed host

#### Scenario: Auth env is listed without a secret
- **WHEN** a product owner reads the MorphUtils env list
- **THEN** they see `PORT` `3040`, the required Morph origin, the optional alias, and the embed origin prompts
- **AND** they see that Morph already allows cross-origin `Authorization` and that CORS is unchanged
- **AND** they see that `VITE_COMPOSERX_URL` is a prompt on `morph-utils` with no value in git

#### Scenario: Project embed URL is a prompt
- **WHEN** a product owner reads the MorphUtils embed list
- **THEN** `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` are prompts with no committed host
- **AND** the documented placeholder is `https://<morph-engi public host>`

## ADDED Requirements

### Requirement: Embed origins are prompts on the shell
The `morph-utils` service MUST list `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, `VITE_MORPH_ENGI_URL`, and `VITE_MORPH_AI_URL` with `sync: false` and MUST NOT give those keys a `value`. The committed Blueprint MUST NOT contain an `onrender.com` host. Those keys MUST NOT appear on `morph`, `formx`, `composerx`, `sharpreport`, or `morph-engi`.

#### Scenario: Embed prompts have no host
- **WHEN** a reviewer reads the embed origin keys on `morph-utils`
- **THEN** each has `sync: false` and no value
- **AND** `render.yaml` does not contain an `onrender.com` host
