## Why

MorphUtils can embed Project only after Project has a production origin. The app still runs as a local Rust API plus a Vite dev server, and the Blueprint attached to the product owner's Render project does not declare it. This change packages that app and adds the Blueprint entry the owner applies. It does not create anything on Render.

## What Changes

- Add a Project image: Rust API and the Svelte UI in one container, `GET /health` succeeds, secrets come from the environment only.
- Add a `morph-engi` web service to the root `render.yaml`, with a persistent disk for SQLite and uploads, without removing `morph` or `morph-utils`.
- Document the Morph API base URL, storage paths, and the public-origin placeholder `https://<morph-engi public host>` for `VITE_PROJECTS_URL` / `VITE_MORPH_ENGI_URL` (story #114). Do not set those variables.
- Allow the flat Blueprint to list services after `morph` and `morph-utils`. The MorphUtils shell itself stays diskless on port 3040.

## Capabilities

### New Capabilities
- `morph-engi-container`: One image serves Project, checks `/health`, keeps secrets out of the build, and stores SQLite and uploads on a data directory.
- `morph-engi-render`: The root Blueprint declares the `morph-engi` service for project `prj-dahc33dbedkc73a1v8n0`, and the runbook names the env and the #114 placeholder.

### Modified Capabilities
- `morph-utils-render`: The Blueprint may contain services besides `morph` and `morph-utils`. Project may be one of them. `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` stay unset on the shell and are documented as the #114 placeholder.

## Impact

- `morph-engi/` image, entrypoint, contract check, and README.
- `render.yaml` gains one service. `morph` and `morph-utils` stay.
- `morph-utils/deploy/check-container-contract.sh` stops requiring an exclusive two-service list.
- `deploy/README.md` and `morph-utils/README.md` gain the Project placeholder without setting embed URLs.
- Project reads `MORPH_ENGI_UPLOAD_DIR` and refuses a development JWT or a loopback Morph API base when `APP_ENV=production`.
- Local `./start-all.sh` ports stay 9096 and 5179. No MorphUtils embed wiring, no Render API calls, no edits to Event Logs, Content Maker, Data Access, AI tools, or invite-signup.
