# Tasks

## 1. Production bundle guard

- [x] 1.1 Add `morph/frontend/scripts/check-morph-utils-bundle.js` that reads `build/static/js/*.js` (not source maps). `unset` fails if `process.env.REACT_APP_MORPH_UTILS_URL` remains or if `https://utils.example.com` or `https://morph-utils.onrender.com` appears. `set <url>` fails unless that exact URL appears and the `process.env` read is gone. Verify by running it with no `build/` directory and confirming a non-zero exit.
- [x] 1.2 Confirm `morph/frontend/src/lib/headerAppLinks.test.js` still expects production to omit an unset or loopback URL and to keep a public URL. Verify with `CI=true npm test -- --watchAll=false --watchman=false src/lib/headerAppLinks.test.js` from `morph/frontend`.

## 2. Morph build arg in the Blueprint

- [x] 2.1 Extend `deploy/check-container-contract.sh` so the `morph` service must list `REACT_APP_MORPH_UTILS_URL` with `sync: false` and no `value:`. Verify the script fails on the current `render.yaml` before the key is added.
- [x] 2.2 Add that prompt to the `morph` env list in `render.yaml` with no URL. Update `deploy/check-stack-runbook.sh` to require that shape and still reject a value and example hosts. Narrow `formx/deploy/check-container-contract.sh` and `composerx/deploy/check-container-contract.sh` so they reject the key on their own service only. Verify `sh deploy/check-container-contract.sh`, `sh deploy/check-stack-runbook.sh`, `sh formx/deploy/check-container-contract.sh`, `sh composerx/deploy/check-container-contract.sh`, and `sh morph-utils/deploy/check-container-contract.sh` all succeed.

## 3. Local build path and runbook

- [x] 3.1 Pass `REACT_APP_MORPH_UTILS_URL` as a compose build arg defaulting to empty, with no host in `deploy/docker-compose.yml`, the Dockerfile default, or `deploy/.env.production.example`. Verify `sh deploy/check-container-contract.sh` still succeeds and those files do not contain `onrender.com`.
- [x] 3.2 Update `deploy/README.md`, `docs/agents/12-build-deploy.md`, `morph-utils/README.md`, and `formx/README.md` so the product owner fills the Morph dashboard prompt with `https://<morph-utils public host>`, rebuilds, and does not commit the URL. State that a restart without a rebuild leaves the chip unset. Verify `sh deploy/check-stack-runbook.sh` and `sh morph-utils/deploy/check-container-contract.sh`.

## 4. CI production builds

- [x] 4.1 In `.github/workflows/ci.yml`, after the unset production build, run the bundle check in `unset` mode, then build again with `REACT_APP_MORPH_UTILS_URL=https://utils.example.com` and run the check in `set` mode. Verify both builds and both checks locally from `morph/frontend`.
