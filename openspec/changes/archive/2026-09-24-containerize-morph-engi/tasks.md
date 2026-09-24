## 1. Production guard and upload directory

- [x] 1.1 Add failing tests for the upload directory default, a configured upload directory, and the production block (dev JWT, short JWT, loopback Morph base, storage outside `/data`, development defaults allowed)
- [x] 1.2 Implement `MORPH_ENGI_UPLOAD_DIR` and the production startup check, and point every upload read and write at that directory

## 2. Project image

- [x] 2.1 Add `morph-engi/Dockerfile`, `morph-engi/Dockerfile.dockerignore`, and `morph-engi/deploy/docker-entrypoint.sh` (Debian runtime, `${PORT}` healthcheck, no secret build args, uid 65532)
- [x] 2.2 Add `morph-engi/deploy/.env.production.example` and `morph-engi/deploy/check-container-contract.sh`
- [x] 2.3 Add `.github/workflows/morph-engi-image.yml` to run the contract, schema-validate `render.yaml`, and build the image without adding a `ci.yml` job

## 3. Blueprint and docs

- [x] 3.1 Append the `morph-engi` service to `render.yaml` without changing the `morph` or `morph-utils` env, and without setting `VITE_PROJECTS_URL` or `VITE_MORPH_ENGI_URL`
- [x] 3.2 Relax the MorphUtils contract so service names may continue after `morph` and `morph-utils`
- [x] 3.3 Document the Morph API base, storage, disk, healthcheck, and the `https://<morph-engi public host>` placeholder for #114 in `morph-engi/README.md`, `deploy/README.md`, and `morph-utils/README.md`

## 4. Verification

- [x] 4.1 Run `cargo test` in `morph-engi/backend`, `npm run build` in `morph-engi/frontend`, the Project and MorphUtils contract scripts, and `render.yaml` schema validation
- [x] 4.2 Build the image and confirm `GET /health` returns HTTP 200
