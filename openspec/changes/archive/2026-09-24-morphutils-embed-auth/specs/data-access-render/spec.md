## MODIFIED Requirements

### Requirement: Morph auth base URL is prompted and not invented
The `sharpreport` service MUST list `USERS_PANEL_BASE_URL` with `sync: false` and MUST NOT give that key a `value`. It MUST list `JWT_SECRET` and `MORPH_AI_API_KEY` with `sync: false` and no `value`. The committed Blueprint MUST NOT contain a usable JWT, password, or API key for this service. It MUST NOT set `VITE_DATAX_URL` on the `sharpreport` service. That key MAY be listed only on `morph-utils` with `sync: false` and no `value`.

#### Scenario: Morph auth URL is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `sharpreport`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not on Data Access
- **WHEN** a reviewer reads env keys on `sharpreport`
- **THEN** `VITE_DATAX_URL` is not one of them
- **AND** `VITE_DATAX_URL` on `morph-utils` has `sync: false` and no value

### Requirement: The runbook records the public URL placeholder
`deploy/README.md` MUST tell the product owner to create the `sharpreport` service from the root Blueprint in Render project `prj-dahc33dbedkc73a1v8n0` without this repository calling Render. It MUST say to set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and that this value is not a secret. It MUST tell the product owner to copy the service's public HTTPS URL and MUST use the placeholder `https://<sharpreport public host>` for the value set as `VITE_DATAX_URL` on MorphUtils. It MUST say that prompt is on `morph-utils` with no value in git. It MUST NOT set that variable on the `sharpreport` service. It MUST document the `/data` disk.

#### Scenario: Product owner can obtain the public URL
- **WHEN** a product owner follows the Data Access Render section in `deploy/README.md`
- **THEN** they can create the service in project `prj-dahc33dbedkc73a1v8n0` and copy `https://<sharpreport public host>`
- **AND** the steps do not call Render from this repository
- **AND** they set `VITE_DATAX_URL` on `morph-utils` as a prompt with no committed host
