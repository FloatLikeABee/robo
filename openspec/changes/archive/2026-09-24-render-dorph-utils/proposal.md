## Why

The DorphUtils image from the container story can be built locally, but the product owner still has no Blueprint entry or public URL to create the hosted shell. Story #106 needs that HTTPS origin as `REACT_APP_DORPH_UTILS_URL` on Dorph, and this repository must not call Render to get it.

## What Changes

- Add one Docker web service, `dorph-utils`, to the existing flat root `render.yaml`, built from `dorph-utils/Dockerfile` with context `dorph-utils/`.
- Document how the product owner creates that shell in the existing Render project and copies its public HTTPS URL. The placeholder for #106 is `https://<dorph-utils public host>`. This change does not set `REACT_APP_DORPH_UTILS_URL`.
- Document the shell's runtime env: pinned `PORT`, required `VITE_DORPH_API_URL` (public Dorph origin, not a secret), the optional alias, and the optional embed URLs. No new secrets and no CORS code change.
- Keep Event Logs, Content Maker, Data Access, and Project out of the Blueprint.

## Capabilities

### New Capabilities

- `dorph-utils-render`: The DorphUtils shell service in the root Blueprint, its env, and the product-owner steps to create it and record the public URL.

### Modified Capabilities

- `dorph-render`: The flat `services` list is no longer a single service. The existing `dorph` service requirements stay; the "only service" wording does not.

## Impact

- `render.yaml`, `deploy/README.md`, `dorph-utils/README.md`.
- `deploy/check-container-contract.sh` must keep validating `dorph` when a second service has its own `PORT`.
- `dorph-utils/deploy/check-container-contract.sh` gains the Blueprint assertions. Schema validation stays the existing Dorph image workflow step. No new `ci.yml` check name.
- No Render API calls, no disk, no embed services, no Dorph header wiring, no Dorph CORS edit.
