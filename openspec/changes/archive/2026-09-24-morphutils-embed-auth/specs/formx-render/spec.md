## MODIFIED Requirements

### Requirement: Morph auth origin is prompted and secrets are empty
The `formx` service MUST list `USERS_PANEL_BASE_URL` with `sync: false` and MUST NOT give that key a `value`. `MORPH_AI_API_KEY`, `SMTP_PASSWORD`, and `PUBLIC_FORM_BASE_URL`, when listed, MUST use `sync: false` and MUST NOT have a `value`. The committed Blueprint MUST NOT contain a usable JWT, password, or API key. It MUST NOT set `VITE_SHEETX_URL` or `VITE_FORMSX_URL` on the `formx` service. Those keys MAY be listed only on `morph-utils`, each with `sync: false` and no `value`. It MUST NOT set `REACT_APP_MORPH_UTILS_URL` on the `formx` service.

#### Scenario: Morph auth origin is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `formx`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not on Event Logs
- **WHEN** a reviewer reads env keys on `formx`
- **THEN** `VITE_SHEETX_URL` and `REACT_APP_MORPH_UTILS_URL` are not set
- **AND** `VITE_SHEETX_URL` on `morph-utils` has `sync: false` and no value

### Requirement: The runbook names env, the disk, and the public URL placeholder
`deploy/README.md` and `formx/README.md` MUST tell the product owner to create the `formx` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. They MUST say to set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and that this value is not a secret. They MUST document the `/data` disk (SQLite, Badger, uploads), that a disk is a single instance, and that `GET /health` does not call MorphUtils. They MUST use the placeholder `https://<event-logs public host>` for the origin set as `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) on MorphUtils. They MUST say that prompt is on `morph-utils` with no value in git. They MUST NOT set that variable on the `formx` or `morph` service.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Event Logs Render section
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<event-logs public host>`
- **AND** the steps do not call Render from this repository
- **AND** they set `VITE_SHEETX_URL` on `morph-utils` as a prompt with no committed host
