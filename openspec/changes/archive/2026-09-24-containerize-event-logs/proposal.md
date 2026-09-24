## Why

MorphUtils can iframe Event Logs only after that app has a production origin. The shell image and Blueprint are already on main. Event Logs (`formx/`) still runs only from `start-all.sh`, so the product owner has nothing to deploy and no public URL to copy for the later embed wire-up.

## What Changes

- Add one Event Logs image that serves the Go API and the Vite UI on the same origin, with `/health` and env-only secrets.
- Add an Event Logs service to the existing flat `render.yaml` (Singapore, starter, `checksPass`, pinned port, one data disk) without calling Render and without changing the Morph or MorphUtils service settings that are already correct.
- Document the Morph auth base URL, the disk paths, and the public URL placeholder `https://<event-logs public host>` for story #114 (`VITE_SHEETX_URL` / `VITE_FORMSX_URL`). Do not set those variables on Morph or MorphUtils.

## Capabilities

### New Capabilities

- `formx-container`: One image runs the Event Logs API and UI, keeps SQLite, Badger, and uploads on `/data`, and answers `/health` without MorphUtils or sibling embeds.
- `formx-render`: The root Blueprint declares the Event Logs service for the product owner, and the runbook names the env and the public URL placeholder.

### Modified Capabilities

- `morph-utils-render`: The flat Blueprint may include the Event Logs service beside `morph` and `morph-utils`. Event Logs is no longer "not a service." Its origin is still not written onto the MorphUtils service.

## Impact

- New `formx/Dockerfile` and a Dockerfile-specific ignore file. The root `.dockerignore` still excludes `formx` from the Morph image context.
- `formx/backend` gains `/health` and static UI serving when a Vite `dist` is present. Local `start-all.sh` ports and cwd-relative data paths stay.
- `render.yaml`, `deploy/README.md`, `formx/README.md`. The MorphUtils runbook sentence that says Event Logs is absent from the Blueprint is updated so it stays true. No `REACT_APP_MORPH_UTILS_URL`, no `VITE_SHEETX_URL` on the `morph-utils` service, no Render API calls, no new required name in `ci.yml`.
