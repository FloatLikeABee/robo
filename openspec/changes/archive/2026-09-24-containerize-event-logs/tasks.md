## 1. Failing checks

- [x] 1.1 Add a Go test that `GET /health` returns 200 and `{"status":"healthy"}` with no outbound call, that a present UI build serves `/events-info` as the UI document, and that `/api/` is not that document
- [x] 1.2 Add a Go test that the listener port stays `SERVER_PORT` (default 29909) when `PORT` is 9090, and that image path env is not the checkout default
- [x] 1.3 Add `formx/deploy/check-container-contract.sh` for the image, Blueprint, and runbook strings in the specs, and run it so it fails before those files exist

## 2. Image

- [x] 2.1 Implement `/health`, the SPA mount (skipped when `index.html` is absent), and the existing `SERVER_PORT` listener so the Go tests pass
- [x] 2.2 Add `formx/Dockerfile`, `formx/Dockerfile.dockerignore`, `formx/deploy/docker-entrypoint.sh`, `formx/docker-compose.yml`, and `formx/deploy/.env.production.example` per design.md
- [x] 2.3 Re-run the Go tests and the contract script

## 3. Blueprint and docs

- [x] 3.1 Append the `formx` service to `render.yaml` without changing the `morph` or `morph-utils` pins and without setting embed URLs
- [x] 3.2 Document env, disk, and `https://<event-logs public host>` in `formx/README.md` and `deploy/README.md`, and correct the MorphUtils runbook so it no longer says Event Logs is absent
- [x] 3.3 Update `morph-utils/deploy/check-container-contract.sh` so the service names are `morph`, `morph-utils`, and `formx`
- [x] 3.4 Add `.github/workflows/formx-image.yml` to run the Event Logs contract and `docker build -f formx/Dockerfile .` without editing `ci.yml`

## 4. Verification

- [x] 4.1 `go test ./...` in `formx/backend`
- [x] 4.2 `npm run build` in `formx/frontend`
- [x] 4.3 Build the image and confirm `GET /health` and `GET /events-info` from the running container
- [x] 4.4 Validate `render.yaml` against the Render schema and run the Morph and MorphUtils contract scripts
