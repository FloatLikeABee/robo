## Why

The Data Access image on Render exits before it listens. The process working directory is `/app`, and startup reads `./migrations`, but the runtime image only copies `config`. Render reports `update_failed` with `Failed to read migrations directory: No such file or directory`.

## What Changes

- Copy `SharpReport/backend/migrations` into the image at `/app/migrations`, the path the binary already reads relative to its working directory.
- Make the Data Access container contract fail when that copy, or the source SQL files, is missing.
- Leave the migration runner, listen port, and Blueprint env unchanged. Do not put `USERS_PANEL_BASE_URL` in git.

## Capabilities

### New Capabilities

### Modified Capabilities

- `data-access-container`: The image must include the SQL migrations at the relative path startup reads, so a missing directory cannot pass the contract.

## Impact

- `SharpReport/Dockerfile`
- `SharpReport/deploy/check-container-contract.sh`
- `openspec/specs/data-access-container/spec.md`
- No Render API calls. No other product Blueprints.
