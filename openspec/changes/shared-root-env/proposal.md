## Why

Each app keeps its own `.env`, so shared keys like `MORPH_AI_API_KEY` drift: Morph can have a working key while Event Logs / FormsX has an empty one, and collect features silently skip AI. Operators should set secrets once.

## What Changes

- **BREAKING (local setup):** Local/dev configuration lives in one repo-root `.env` (gitignored). Apps MUST NOT require a per-app `.env` to start.
- Add a committed repo-root `.env.example` that lists shared and app-specific keys in one file.
- `start-all.sh` loads only the root `.env` for every service. It MUST NOT load per-app `.env` files (those currently override shared keys with empty values).
- Processes started without `start-all.sh` (direct `go run`, `cargo run`, `npm run dev`) MUST still find the root `.env`.
- Per-app `.env.example` files become pointers to the root template (or are removed after the root file covers their keys).
- Production remains `deploy/.env.production` (already one file). Do not merge prod secrets into the local root `.env`.
- Relative data paths (`./data/...`) stay relative to each app’s working directory as today.

## Capabilities

### New Capabilities

- `shared-root-env`: One local config file for all platform apps; shared keys apply everywhere; per-app `.env` is not used at runtime.

### Modified Capabilities

- (none — no existing `openspec/specs/` capabilities)

## Impact

- `start-all.sh` env loading
- Go/Rust/Python dotenv loaders (Morph, FormsX, ComposerX, Booki, Academi, Morph Engi, SharpReport, morphgraph-worker, DataX/bk)
- Vite/CRA frontends that read `VITE_*` / `REACT_APP_*`
- Root `.gitignore` already ignores `**/.env`; add root `.env.example`
- README, DEVELOPER_BASELINE, deploy notes, agent docs that say “copy each app’s `.env.example`”
- Existing per-app `.env` files on disk become unused (leave them in place locally; do not delete operator secrets)
