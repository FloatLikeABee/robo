## 1. Root template and launcher

- [x] 1.1 Add repo-root `.env.example` with MorphAI/auth keys first, then per-app port and path sections (no secrets)
- [x] 1.2 Change `start-all.sh` to load only `${ROOT}/.env`; stop `load_app_env` nested/parent `.env` (empty nested keys must not win)
- [x] 1.3 Add `scripts/merge-local-env.sh` that creates root `.env` from non-empty keys in known per-app `.env` files and never overwrites an existing root file

## 2. Process loaders (walk up to `start-all.sh`)

- [x] 2.1 Add Go `pkg/repoenv` (or equivalent) that loads repo-root `.env` and does not load nested `.env`; unit test walk-up + nested empty key ignored
- [x] 2.2 Call the helper from Morph, FormsX, ComposerX, Booki, Academi, morphgraph-worker (replace cwd/`../.env` godotenv)
- [x] 2.3 Apply the same walk-up in Morph Engi, SharpReport, and `pkg/morphai-rs` dotenv loaders
- [x] 2.4 Apply the same walk-up in bk/DataX Python env load

## 3. Frontends

- [x] 3.1 Set Vite `envDir` to the repo root in formx, morph-utils, composerx, booki, morph-engi, SharpReport (and DataX/bk if Vite)
- [x] 3.2 Make Morph CRA (`morph/frontend`) read `REACT_APP_*` from the root `.env` when started without `start-all.sh`

## 4. Docs and per-app examples

- [x] 4.1 Shrink per-app `.env.example` files to a pointer at the root template (keep only a unique-key hint if needed)
- [x] 4.2 Update README, DEVELOPER_BASELINE, DEPLOY-README, and agent docs: copy root `.env.example`, production stays `deploy/.env.production`
- [x] 4.3 Smoke: root `.env` with `MORPH_AI_API_KEY` set and leftover empty `formx/backend/.env` — FormsX still sees the key; `deploy/.env.production` path unchanged
