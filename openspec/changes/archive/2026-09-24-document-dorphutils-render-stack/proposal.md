# Proposal

## Why

Per-service Render sections tell the product owner how to add one image, but they never give one ordered create path for the DorphUtils stack. The Dorph header link is a Docker build arg (`REACT_APP_DORPH_UTILS_URL` in the root `Dockerfile`), so a rebuild before that URL exists omits the link. Forge must not create the services.

## What Changes

- Add one create-order section to `deploy/README.md` for project `prj-dahc33dbedkc73a1v8n0`: sync `render.yaml` on `main`, record each public HTTPS origin, then rebuild Dorph only after the DorphUtils origin exists.
- State that Forge must not create services on Render and that this repository does not call Render.
- Add an env matrix with placeholders for the DorphUtils URL and each embed origin the wire-up story uses. Live hosts already created by the product owner are examples, not hosts to recreate.
- Note that a Blueprint create can drop nested `envVars`, so `USERS_PANEL_BASE_URL` may need a dashboard update after create. No secret values enter git. `render.yaml` is unchanged.

## Capabilities

### New Capabilities

- `dorphutils-stack-render-runbook`: The product-owner create and record order for DorphUtils, Event Logs, Content Maker, Data Access, and Project, including the Dorph rebuild step and the embed-origin placeholders.

### Modified Capabilities

- None. Existing Blueprint contracts still forbid committing `REACT_APP_DORPH_UTILS_URL` and the embed URL keys. This change documents the dashboard order; it does not put those values in `render.yaml`.

## Impact

- `deploy/README.md` and a short pointer from `docs/agents/12-build-deploy.md`.
- A small doc contract check, same style as the existing container contract scripts.
- No Render API calls, no new services, no Dockerfiles, no secrets.
