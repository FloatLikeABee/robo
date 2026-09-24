## Why

MorphUtils can iframe Content Maker only after that app has a public origin. The API and UI still run from a checkout (`start-all.sh`, ports 8043 and 8044). Morph and the MorphUtils shell already have images and Blueprint services. Content Maker does not, so the product owner cannot create it in project `prj-dahc33dbedkc73a1v8n0`.

## What Changes

- Add one Content Maker image: the Go API and the production Svelte UI, health on `/health`, data on `/data`, secrets only from the environment.
- Add a `composerx` Docker web service to the root `render.yaml`, beside `morph` and `morph-utils`. Do not remove those services. Do not set `VITE_COMPOSERX_URL`.
- Document the env vars, the Morph auth base URL `USERS_PANEL_BASE_URL`, and the placeholder `https://<composerx public host>` for story #114.
- This repository does not call Render.

## Capabilities

### New Capabilities

- `composerx-container`: One image serves the Content Maker API and UI, with a data mount and env-only secrets.
- `composerx-render`: The flat Blueprint declares the Content Maker service and the runbook names its public URL placeholder.

### Modified Capabilities

- `morph-utils-render`: The Blueprint may include `composerx` after `morph` and `morph-utils`. Content Maker is a service. `VITE_COMPOSERX_URL` stays unset on the shell until #114. Event Logs, Data Access, and Project stay out.

## Impact

- New `composerx/Dockerfile`, entrypoint, contract check, and local compose file.
- `composerx/backend` listens on `COMPOSERX_PORT` when set, otherwise `PORT`, and serves the UI when `COMPOSERX_UI_DIR` is set. Local `start-all.sh` defaults stay `./data` and port 8043 for the API, 8044 for Vite.
- Root `.dockerignore` must keep Content Maker source in the shared build context. The Morph Dockerfile still must not copy that tree.
- `render.yaml`, `deploy/README.md`, `composerx/backend/README.md`, `docs/agents/05-composerx.md`, `docs/agents/12-build-deploy.md`.
- `morph-utils/deploy/check-container-contract.sh` and the MorphUtils Render sentences that said Content Maker is not a service.
- No formx, SharpReport, morph-engi, bk, or invite-signup behavior changes. No Morph header chip. No new `ci.yml` check name.
