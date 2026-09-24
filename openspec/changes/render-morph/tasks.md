# Tasks

## 1. Blueprint contract

- [x] 1.1 Extend `deploy/check-container-contract.sh` so it fails unless `render.yaml` is a flat Blueprint (no `projects`), names one `morph` web service, pins `PORT` to `9090`, sets `MORPH_AI_PROVIDER` to `dashscope`, mounts `morph-data` at `/data`, lists the image `buildFilter` paths, and gives every secret key `sync: false` with no `value`. Run the script and confirm it fails because `render.yaml` is absent.
- [x] 1.2 Add root `render.yaml` per `design.md` (service fields, disk, `maxShutdownDelaySeconds: 120`, env inventory, secrets). Re-run `sh deploy/check-container-contract.sh` and confirm it prints `container contract ok`.

## 2. Schema check on the image workflow

- [x] 2.1 Add a step to `.github/workflows/docker-image.yml` that validates `render.yaml` with `check-jsonschema` against `https://render.com/schema/render.yaml.json`. Do not edit `ci.yml` or rename the `Build image` job. Run the validator locally and record a successful exit.

## 3. Runbook

- [x] 3.1 Add a Deploy on Render section to `deploy/README.md` (create from the Blueprint, `JWT_SECRET` 32+ random characters, `ADMIN_PASSWORD` 12+ and not `admin123`, disk is single-instance and not zero-downtime, snapshot `/data`, check `GET /health`). Point `docs/agents/12-build-deploy.md` at that section. Update the `openspec/config.yaml` context line that still says not to document Render. Confirm the new section names those rules by reading the file.

## 4. Image verification

- [ ] 4.1 Run the image with `PORT=9090`, `MORPH_ENV=production`, throwaway secrets, and a named volume at `/data` that starts root-owned. Confirm `GET /health` is 200, restart the container, and confirm the data directory is still present. Do not pass `--user`.
