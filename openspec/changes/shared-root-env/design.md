## Context

See proposal.md for motivation. Specs: `specs/shared-root-env/spec.md`.

Today `start-all.sh` `load_app_env` applies the parent directory `.env` then the app working-directory `.env`. App files win, including empty `MORPH_AI_API_KEY=`. Binaries also call `godotenv.Load()` / `dotenvy` against cwd or `../.env`, so a nested empty file still wins when running without the launcher. Production already uses one file: `deploy/.env.production`.

Relative paths such as `TRAN_SQLITE_PATH=./data/tran.sqlite` are resolved from each service’s working directory. That must stay true after the root file is introduced.

## Goals / Non-Goals

**Goals:**

- One local source of truth: repo-root `.env`.
- Launcher and direct `go`/`cargo`/`npm`/`python` starts all see it.
- Nested leftover `.env` files cannot blank shared keys.
- Document keys once in root `.env.example`.

**Non-Goals:**

- Changing production to read the developer’s local `.env`.
- Unifying database engines or port numbers.
- Deleting operators’ existing per-app `.env` files from disk.
- A secrets manager / vault.
- Renaming `USERS_PANEL_BASE_URL`.

## Decisions

### 1. Root file wins; nested `.env` is not loaded
- **Choice**: `start-all.sh` loads only `${ROOT}/.env`. Go/Rust/Python loaders walk up from cwd until they find a directory that contains `start-all.sh`, then load that directory’s `.env` if present. They MUST NOT load a nearer nested `.env`.
- **Rationale**: Nested empty keys are what hid Morph AI from FormsX. The user asked for one file, not per-app files with overrides.
- **Alternatives**: Load nested files but skip empty values — still two sources of truth. Load root then nested non-empty overrides — useful later, not requested.

### 2. Shared Go helper for walk-up load
- **Choice**: Add a small helper (e.g. `pkg/repoenv`) used by Morph, FormsX, ComposerX, Booki, Academi, morphgraph-worker. Rust (Morph Engi, SharpReport, morphai-rs) and Python (bk) get the same walk-up rule in their existing config loaders.
- **Rationale**: One Go implementation avoids copy-paste mistakes. Rust/Python cannot import that package.
- **Alternatives**: Only fix `start-all.sh` — fails `go run` from `formx/backend`.

### 3. Vite `envDir` = repo root; CRA inherits process env
- **Choice**: Vite apps set `envDir` to the repo root (from each `vite.config`). Morph’s CRA/webpack frontend relies on the shell env from `start-all.sh`, plus the same walk-up if a Node start script can load dotenv from root (keep this thin: `dotenv` pointing at root, or document that Morph UI is started via `start-all.sh`).
- **Rationale**: Vite does not read a parent `.env` unless `envDir` is set. Process env from `start-all.sh` is enough when using the launcher; `envDir` covers `npm run dev` in the frontend folder.
- **Alternatives**: Symlink each `frontend/.env` to root — still many files.

### 4. Root `.env.example` is the catalog; per-app examples shrink
- **Choice**: Commit `.env.example` at repo root with shared MorphAI/auth keys first, then commented sections per app (ports, sqlite/badger paths). Each existing per-app `.env.example` becomes a short pointer to `../../.env.example` (or equivalent) plus any key that is truly unique and easy to miss.
- **Rationale**: Spec requires one place to see required keys. Keeping a one-line pointer avoids stale duplicate lists.
- **Alternatives**: Delete all per-app examples — worse for someone who only opens `formx/README.md`.

### 5. Optional one-shot merge for existing machines
- **Choice**: A small `scripts/merge-local-env.sh` (not run by default) copies non-empty keys from known per-app `.env` files into a new root `.env` if the root file is missing. Last non-empty value wins; never overwrite an existing root `.env`.
- **Rationale**: Operators already have keys scattered; they should not retype them. Do not commit the merged file.
- **Alternatives**: Manual copy only — easy to miss FormsX.

## Risks / Trade-offs

- [Wrong working directory for relative paths] → Keep `start-all.sh` `cd` into each app dir as today; document that `./data/...` is per-app cwd.
- [Two `.env` files confuse people] → Docs: ignore nested files; launcher does not load them.
- [Vite picks up unrelated `VITE_*` from a huge root file] → Prefix remains app-specific; unused vars are harmless.
- [Merge script picks a stale empty-then-filled order] → Only copy non-empty; prefer Morph’s key when several are set (document order: morph, then others).

## Migration Plan

1. Add root `.env.example` and helper/loaders.
2. Switch `start-all.sh` to root-only load.
3. Point per-app examples and READMEs at the root file.
4. Operators: `cp .env.example .env` or run `scripts/merge-local-env.sh`, then `./start-all.sh restart`.
5. Rollback: restore previous `load_app_env` and godotenv cwd loads; nested `.env` files remain on disk.

## Open Questions

None that change specs or tasks.
