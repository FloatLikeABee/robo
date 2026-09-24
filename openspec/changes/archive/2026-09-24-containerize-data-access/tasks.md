## 1. Listen port, database path, and health

- [x] 1.1 Add failing tests for `SHARPREPORT_PORT` over `PORT`, an absolute `sqlite:///data/datapulse.db` path, and `GET /health` plus `GET /ready` returning `{"status":"ok"}` without a Morph call.
- [x] 1.2 Make those tests pass. Keep the debug `GET /` redirect when the UI directory is unset. Do not change `start-all.sh` ports.

## 2. Production UI in the API process

- [x] 2.1 Build the SvelteKit app as a static fallback (`index.html`) and serve it when `SHARPREPORT_UI_DIR` is set. `GET /` returns that document. Unmatched `/api` paths stay non-HTML. Leave `PUBLIC_API_URL` unset.
- [x] 2.2 Confirm `npm run build` in `SharpReport/frontend` succeeds.

## 3. Image and local compose

- [x] 3.1 Add `SharpReport/Dockerfile` (context repo root), entrypoint (chown `/data` to uid 65532), and compose that uses an env file. Healthcheck uses `${PORT}`. No secret build args. No Java. Remove the compose file that commits a JWT and a database password.
- [x] 3.2 Add `SharpReport/deploy/.env.production.example` with empty `JWT_SECRET` and `USERS_PANEL_BASE_URL`. Add `SharpReport/deploy/check-container-contract.sh`. Narrow the root `.dockerignore` so the Data Access image can copy `SharpReport/` while the word `SharpReport` remains.

## 4. Blueprint and docs

- [x] 4.1 Append the `sharpreport` service to `render.yaml`. Keep `morph` and `morph-utils`. Pin ports, disk, and empty prompts. Do not set `VITE_DATAX_URL`.
- [x] 4.2 Update the MorphUtils contract so `sharpreport` is required and extra sibling services are allowed. Document build/runtime env, the Morph auth base URL, the volume, and `https://<sharpreport public host>` in `deploy/README.md`, `SharpReport/README.md`, `docs/agents/07-rust-apps.md`, and `docs/agents/12-build-deploy.md`. Run the Data Access contract from the existing image workflow without a new check name.

## 5. Verify

- [x] 5.1 Run `cargo test` for `SharpReport/backend`, `npm run build` for `SharpReport/frontend`, both container contracts, and `check-jsonschema` on `render.yaml`.
