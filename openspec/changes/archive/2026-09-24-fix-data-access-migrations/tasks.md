## 1. Contract first

- [x] 1.1 Add the Data Access container checks from design.md (exact `COPY` line, last `WORKDIR /app`, at least one source `.sql` file, dockerignore does not drop the migrations tree or `*.sql`) and run the contract so it fails before the Dockerfile changes.

## 2. Image

- [x] 2.1 Copy `SharpReport/backend/migrations` to `/app/migrations` in the runtime stage. Do not add a migrations env var, do not change the runner, and do not set `USERS_PANEL_BASE_URL`.

## 3. Verify

- [x] 3.1 Re-run `sh SharpReport/deploy/check-container-contract.sh` and confirm it passes. Do not edit `render.yaml`.
