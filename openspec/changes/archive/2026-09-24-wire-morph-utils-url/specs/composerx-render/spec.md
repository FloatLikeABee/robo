# Spec Delta

## MODIFIED Requirements

### Requirement: Morph auth URL and secrets are prompted and not invented
The `composerx` service MUST list `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, `TRAN_QWEN_API_KEY`, and `TRAN_OPENAI_API_KEY` with `sync: false` and MUST NOT give those keys a `value`. The committed Blueprint MUST NOT contain a JWT, password, API key, or a guessed public host. It MUST NOT list `VITE_COMPOSERX_URL` as an env key. It MUST NOT list `REACT_APP_MORPH_UTILS_URL` on the `composerx` service.

#### Scenario: Morph auth base URL is a dashboard prompt
- **WHEN** a reviewer reads `USERS_PANEL_BASE_URL` on `composerx`
- **THEN** it has `sync: false` and no value

#### Scenario: Embed URL is not set here
- **WHEN** a reviewer reads env keys on `composerx`
- **THEN** `VITE_COMPOSERX_URL` and `REACT_APP_MORPH_UTILS_URL` are not set
