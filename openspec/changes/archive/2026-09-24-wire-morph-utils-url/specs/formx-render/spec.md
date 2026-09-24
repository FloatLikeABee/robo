# Spec Delta

## MODIFIED Requirements

### Requirement: Morph auth origin is prompted and secrets are empty
The `formx` service MUST list `USERS_PANEL_BASE_URL` with `sync: false` and MUST NOT give that key a `value`. `MORPH_AI_API_KEY`, `SMTP_PASSWORD`, and `PUBLIC_FORM_BASE_URL`, when listed, MUST use `sync: false` and MUST NOT have a `value`. The committed Blueprint MUST NOT contain a usable JWT, password, or API key. It MUST NOT set `VITE_SHEETX_URL` or `VITE_FORMSX_URL` on any service. It MUST NOT set `REACT_APP_MORPH_UTILS_URL` on the `formx` service.

#### Scenario: Morph auth origin is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `formx`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not wired here
- **WHEN** a reviewer reads env keys on `formx`
- **THEN** `VITE_SHEETX_URL` and `REACT_APP_MORPH_UTILS_URL` are not set
- **AND** `VITE_SHEETX_URL` is not set on any service
