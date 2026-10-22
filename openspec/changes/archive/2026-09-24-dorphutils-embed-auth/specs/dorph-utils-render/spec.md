## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Embed origins are prompts on the shell
The `dorph-utils` service MUST list `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, `VITE_DORPH_ENGI_URL`, and `VITE_DORPH_AI_URL` with `sync: false` and MUST NOT give those keys a `value`. The committed Blueprint MUST NOT contain an `onrender.com` host. Those keys MUST NOT appear on `dorph`, `formx`, `composerx`, `sharpreport`, or `dorph-engi`.

#### Scenario: Embed prompts have no host
- **WHEN** a reviewer reads the embed origin keys on `dorph-utils`
- **THEN** each has `sync: false` and no value
- **AND** `render.yaml` does not contain an `onrender.com` host
