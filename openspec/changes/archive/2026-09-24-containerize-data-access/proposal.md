## Why

MorphUtils can embed Data Access only after that app has a container and a Blueprint entry the product owner can create. The shell is already a Render service. Data Access (`SharpReport/`) is still a local Rust API plus a Vite UI, and the old `SharpReport/deploy` image hardcodes a port, a JWT, and a Metabase database password.

## What Changes

- Add one Data Access image: the Rust API and the production UI, `GET /health` and `GET /ready`, SQLite on `/data`, secrets only from the environment.
- Add a `sharpreport` web service to `render.yaml` beside `morph` and `morph-utils`. Pin `PORT` and `SHARPREPORT_PORT` to `3050`. Prompt for the Morph auth base URL. Do not set `VITE_DATAX_URL`.
- Document build and runtime env, the Morph auth base URL, the `/data` volume, and the public URL placeholder `https://<sharpreport public host>` for story #114.
- Replace the old SharpReport compose file that commits a JWT and a database password.

## Capabilities

### New Capabilities

- `data-access-container`: One image starts locally, serves health and ready, and keeps secrets out of the build.
- `data-access-render`: The root Blueprint declares the Data Access service, disk, and env prompts for the product owner.

### Modified Capabilities

- `morph-utils-render`: The Blueprint may include `sharpreport` after `morph` and `morph-utils`. `VITE_DATAX_URL` stays unset on the shell. The Data Access public URL is a placeholder for #114, not a value in this change.

## Impact

- `SharpReport/` (Dockerfile, entrypoint, listen port, static UI, SQLite path), root `.dockerignore`, `render.yaml`.
- `deploy/README.md`, `SharpReport/README.md`, `docs/agents/07-rust-apps.md`, `docs/agents/12-build-deploy.md`.
- `morph-utils/deploy/check-container-contract.sh` service-name check, because it currently requires the names to be only `morph` and `morph-utils`.
- No Render API calls. No MorphUtils embed URL. No edits to Event Logs, Content Maker, Project, AI tools, or Invite Signup application code.
