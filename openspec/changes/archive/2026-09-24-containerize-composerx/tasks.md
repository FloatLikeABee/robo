## 1. API listen port and UI mount

- [x] 1.1 Add a failing Go test: `COMPOSERX_PORT` wins over `PORT`, empty `COMPOSERX_PORT` uses `PORT`, both empty uses `8043`. Run it and confirm it fails.
- [x] 1.2 Add a failing Go test: with `COMPOSERX_UI_DIR` set to a fixture that has `index.html`, `favicon.svg`, and `assets/app.js`, `GET /` and `GET /assets/app.js` return 200 without a token, and `GET /templates` without a token is 401. `GET /health` stays 200 and does not change. Run it and confirm it fails.
- [x] 1.3 Implement the listen fallback and the UI mount in `composerx/backend` until those tests pass. Do not add a catch-all to `index.html`. Leave checkout defaults (`./data`, `./storage`, no UI dir) unchanged.

## 2. Image

- [x] 2.1 Add `composerx/deploy/check-container-contract.sh` that fails unless the Dockerfile uses `CGO_ENABLED=0`, health-checks `http://127.0.0.1:${PORT}/health` with no `$$`, has no secret `ARG`, sets `VITE_API_BASE` empty, and does not copy `.env` or `ai.config.json`. Run it and confirm it fails because the Dockerfile is absent.
- [x] 2.2 Add `composerx/Dockerfile`, `composerx/deploy/docker-entrypoint.sh` (chown `/data` to uid 65532, then `su-exec`), `composerx/docker-compose.yml`, and `composerx/deploy/.env.production.example` with empty `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, and `TRAN_OPENAI_API_KEY`. Narrow root `.dockerignore` so `composerx` source is in a context of `.` while `composerx/backend/storage`, frontend `node_modules`, env files, and `ai.config.json` stay out. The ignore file must still contain the word `composerx`. Extend `deploy/check-container-contract.sh` so the root Dockerfile must not `COPY` `composerx`. Re-run both contract scripts.

## 3. Blueprint

- [x] 3.1 Extend `morph-utils/deploy/check-container-contract.sh` so service names must be `morph`, `morph-utils`, `composerx` in that order, and the `composerx` block is not required to satisfy MorphUtils fields. Run it and confirm it fails because `composerx` is absent. Confirm `sh deploy/check-container-contract.sh` still prints `container contract ok`.
- [x] 3.2 Append the `composerx` service to `render.yaml` per `design.md`: Singapore, starter, `main`, `./composerx/Dockerfile`, context `.`, `/health`, `checksPass`, disk `composerx-data` at `/data` size 1, `maxShutdownDelaySeconds: 120`, `PORT` and `COMPOSERX_PORT` `8043`, store paths under `/data`, prompted `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, `TRAN_QWEN_API_KEY`, and `TRAN_OPENAI_API_KEY` with `sync: false` and no value. No `projects`, no `numInstances`, no `VITE_COMPOSERX_URL`. Extend `composerx/deploy/check-container-contract.sh` to assert that block and the runbook strings, then re-run the Morph, MorphUtils, and Content Maker contract scripts.

## 4. Docs

- [x] 4.1 Update `deploy/README.md`, `composerx/backend/README.md`, `docs/agents/05-composerx.md`, and `docs/agents/12-build-deploy.md` with the image run, the env list, `USERS_PANEL_BASE_URL` as `https://<morph public host>` (not a secret), project `prj-dahc33dbedkc73a1v8n0`, and placeholder `https://<composerx public host>` for #114. Do not set `VITE_COMPOSERX_URL`.
- [x] 4.2 Update the MorphUtils Render sentences in `deploy/README.md` and `morph-utils/README.md` so Content Maker is a Blueprint service and `VITE_COMPOSERX_URL` is still not set on `morph-utils`. Keep the existing MorphUtils contract needles (`prj-dahc33dbedkc73a1v8n0`, `https://<morph public host>`, `https://<morph-utils public host>`, embed variable names, CORS wording).

## 5. Verify

- [x] 5.1 `go test` in `composerx/backend`. Production `npm run build` in `composerx/frontend`. `sh deploy/check-container-contract.sh`, `sh morph-utils/deploy/check-container-contract.sh`, and `sh composerx/deploy/check-container-contract.sh`. Validate `render.yaml` against `https://render.com/schema/render.yaml.json`.
- [x] 5.2 `docker build` the Content Maker image and run it with a scratch `/data` volume and no Morph. `GET /health` returns 200 and `"status":"ok"`. `GET /` returns the UI document. No Rust crates are touched, so no `cargo test`.
