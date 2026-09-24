## Why

The MorphUtils image from the container story can be built locally, but the product owner still has no Blueprint entry or public URL to create the hosted shell. Story #106 needs that HTTPS origin as `REACT_APP_MORPH_UTILS_URL` on Morph, and this repository must not call Render to get it.

## What Changes

- Add one Docker web service, `morph-utils`, to the existing flat root `render.yaml`, built from `morph-utils/Dockerfile` with context `morph-utils/`.
- Document how the product owner creates that shell in the existing Render project and copies its public HTTPS URL. The placeholder for #106 is `https://<morph-utils public host>`. This change does not set `REACT_APP_MORPH_UTILS_URL`.
- Document the shell's runtime env: pinned `PORT`, required `VITE_MORPH_API_URL` (public Morph origin, not a secret), the optional alias, and the optional embed URLs. No new secrets and no CORS code change.
- Keep Event Logs, Content Maker, Data Access, and Project out of the Blueprint.

## Capabilities

### New Capabilities

- `morph-utils-render`: The MorphUtils shell service in the root Blueprint, its env, and the product-owner steps to create it and record the public URL.

### Modified Capabilities

- `morph-render`: The flat `services` list is no longer a single service. The existing `morph` service requirements stay; the "only service" wording does not.

## Impact

- `render.yaml`, `deploy/README.md`, `morph-utils/README.md`.
- `deploy/check-container-contract.sh` must keep validating `morph` when a second service has its own `PORT`.
- `morph-utils/deploy/check-container-contract.sh` gains the Blueprint assertions. Schema validation stays the existing Morph image workflow step. No new `ci.yml` check name.
- No Render API calls, no disk, no embed services, no Morph header wiring, no Morph CORS edit.
