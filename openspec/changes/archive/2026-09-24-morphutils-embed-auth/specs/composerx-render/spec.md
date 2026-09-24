## MODIFIED Requirements

### Requirement: Morph auth URL and secrets are prompted and not invented
The `composerx` service MUST list `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, `TRAN_QWEN_API_KEY`, and `TRAN_OPENAI_API_KEY` with `sync: false` and MUST NOT give those keys a `value`. The committed Blueprint MUST NOT contain a JWT, password, API key, or a guessed public host. It MUST NOT list `VITE_COMPOSERX_URL` on the `composerx` service. That key MAY be listed only on `morph-utils` with `sync: false` and no `value`. It MUST NOT list `REACT_APP_MORPH_UTILS_URL` on the `composerx` service.

#### Scenario: Morph auth base URL is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `composerx`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not on Content Maker
- **WHEN** a reviewer reads env keys on `composerx`
- **THEN** `VITE_COMPOSERX_URL` and `REACT_APP_MORPH_UTILS_URL` are not set
- **AND** `VITE_COMPOSERX_URL` on `morph-utils` has `sync: false` and no value

### Requirement: The runbook records the public URL placeholder
`deploy/README.md` and `composerx/backend/README.md` MUST tell the product owner to create the `composerx` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and that this value is not a secret. They MUST list `MORPH_AI_API_KEY` and `TRAN_OPENAI_API_KEY` as optional keys with no sample secret. They MUST tell the product owner to copy the service's public HTTPS URL and MUST use the placeholder `https://<composerx public host>` for the value set as `VITE_COMPOSERX_URL` on MorphUtils. They MUST say that prompt is on `morph-utils` with no value in git. They MUST NOT set `VITE_COMPOSERX_URL` on the `composerx` service. They MUST state that `GET /health` returns HTTP 200 without calling MorphUtils.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Content Maker Render section in `deploy/README.md`
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<composerx public host>`
- **AND** the steps do not call Render from this repository
- **AND** they set `VITE_COMPOSERX_URL` on `morph-utils` as a prompt with no committed host
