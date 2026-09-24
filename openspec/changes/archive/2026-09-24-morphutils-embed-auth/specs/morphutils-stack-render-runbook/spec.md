## MODIFIED Requirements

### Requirement: Env matrix has placeholders for the wire-up origins
The create order MUST include an env matrix the product owner can fill. The matrix MUST include a placeholder for the MorphUtils origin used as `REACT_APP_MORPH_UTILS_URL`, and a placeholder for each embed origin: `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) as `https://<event-logs public host>`, `VITE_COMPOSERX_URL` as `https://<composerx public host>`, `VITE_DATAX_URL` as `https://<sharpreport public host>`, and `VITE_PROJECTS_URL` (alias `VITE_MORPH_ENGI_URL`) as `https://<morph-engi public host>`. The matrix MUST say the product owner sets those `VITE_*` keys on `morph-utils` as dashboard prompts with no value in git. It MUST say those keys are read when the MorphUtils container starts, so a later change of them does not require a MorphUtils image rebuild. It MUST name the session cookie `userspanel_session_token`, `SameSite=Lax`, and that no cookie `Domain` is set. It MUST say `bk` and invite-signup are not part of this stack.

#### Scenario: Product owner fills the matrix
- **WHEN** a product owner fills the env matrix
- **THEN** there is a placeholder for the MorphUtils URL
- **AND** there is a placeholder for each embed origin set on `morph-utils`
- **AND** the cookie notes name `userspanel_session_token` and `SameSite=Lax`
