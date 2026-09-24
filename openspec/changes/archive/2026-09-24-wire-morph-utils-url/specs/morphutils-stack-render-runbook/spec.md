# Spec Delta

## MODIFIED Requirements

### Requirement: Create order records each stack URL before the Morph rebuild
The create order MUST name the Blueprint file `render.yaml` on branch `main` and the services `morph-utils` (MorphUtils), `formx` (Event Logs), `composerx` (Content Maker), `sharpreport` (Data Access), and `morph-engi` (Project). It MUST tell the product owner to copy each service's public HTTPS origin from the dashboard after that service is live. It MUST place the Morph rebuild after the MorphUtils origin has been copied. That rebuild MUST set `REACT_APP_MORPH_UTILS_URL` to `https://<morph-utils public host>` and MUST say the value is a Docker build argument inlined when the Morph image is built, so a restart without a rebuild does not add the header link. The `morph` service in `render.yaml` MUST list `REACT_APP_MORPH_UTILS_URL` with `sync: false` and MUST NOT give that key a `value`. The runbook MUST tell the product owner to fill that dashboard prompt with the copied origin and then rebuild the Morph image. The runbook MUST NOT put a value for `REACT_APP_MORPH_UTILS_URL` in `render.yaml`.

#### Scenario: Product owner follows the numbered order
- **WHEN** a product owner follows the create order
- **THEN** they can create MorphUtils, Event Logs, Content Maker, Data Access, and Project from the Blueprint
- **AND** they record each public HTTPS origin before rebuilding Morph
- **AND** the Morph rebuild uses `REACT_APP_MORPH_UTILS_URL` set to `https://<morph-utils public host>`
- **AND** that value is entered on the Morph dashboard prompt, not committed in `render.yaml`

#### Scenario: Restart is not treated as the Morph wire-up
- **WHEN** a reader looks up how `REACT_APP_MORPH_UTILS_URL` takes effect
- **THEN** the runbook says the Morph image must be rebuilt
- **AND** it says a restart without that rebuild leaves the header link unset

#### Scenario: Blueprint prompt has no host
- **WHEN** a reviewer reads `REACT_APP_MORPH_UTILS_URL` in `render.yaml`
- **THEN** it is on the `morph` service with `sync: false` and no value
- **AND** the file does not contain an `onrender.com` host as that value
