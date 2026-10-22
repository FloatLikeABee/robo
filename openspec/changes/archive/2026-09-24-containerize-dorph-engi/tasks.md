## 1. Production guard and upload directory

- [x] 1.1 Add failing tests for the upload directory default, a configured upload directory, and the production block (dev JWT, short JWT, loopback Dorph base, storage outside `/data`, development defaults allowed)
- [x] 1.2 Implement `DORPH_ENGI_UPLOAD_DIR` and the production startup check, and point every upload read and write at that directory

## 2. Project image

- [x] 2.1 Add `dorph-engi/Dockerfile`, `dorph-engi/Dockerfile.dockerignore`, and `dorph-engi/deploy/docker-entrypoint.sh` (Debian runtime, `${PORT}` healthcheck, no secret build args, uid 65532)
- [x] 2.2 Add `dorph-engi/deploy/.env.production.example` and `dorph-engi/deploy/check-container-contract.sh`
- [x] 2.3 Add `.github/workflows/dorph-engi-image.yml` to run the contract, schema-validate `render.yaml`, and build the image without adding a `ci.yml` job

## 3. Blueprint and docs

- [x] 3.1 Append the `dorph-engi` service to `render.yaml` without changing the `dorph` or `dorph-utils` env, and without setting `VITE_PROJECTS_URL` or `VITE_DORPH_ENGI_URL`
- [x] 3.2 Relax the DorphUtils contract so service names may continue after `dorph` and `dorph-utils`
- [x] 3.3 Document the Dorph API base, storage, disk, healthcheck, and the `https://<dorph-engi public host>` placeholder for #114 in `dorph-engi/README.md`, `deploy/README.md`, and `dorph-utils/README.md`

## 4. Verification

- [x] 4.1 Run `cargo test` in `dorph-engi/backend`, `npm run build` in `dorph-engi/frontend`, the Project and DorphUtils contract scripts, and `render.yaml` schema validation
- [x] 4.2 Build the image and confirm `GET /health` returns HTTP 200
