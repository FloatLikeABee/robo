## 1. Contract checks first

- [x] 1.1 Extend `morph-utils/deploy/check-container-contract.sh` so it fails unless `render.yaml` has a `morph-utils` Docker web service (Singapore, starter, `main`, `./morph-utils/Dockerfile`, context `./morph-utils`, `/health`, `checksPass`), `PORT` `3040`, `VITE_MORPH_API_URL` with `sync: false` and no value, the image `buildFilter` paths, no disk on that service, and service names only `morph` and `morph-utils`. It must also fail unless `deploy/README.md` and `morph-utils/README.md` name project `prj-dahc33dbedkc73a1v8n0`, `https://<morph public host>`, `https://<morph-utils public host>`, `REACT_APP_MORPH_UTILS_URL`, the alias and embed variables, and that CORS already allows `Authorization` and is not edited. Run it and confirm it fails because the service is absent.
- [x] 1.2 Scope `deploy/check-container-contract.sh` to the `morph` service block for env keys, `dockerContext: .`, `dockerfilePath: ./Dockerfile`, and that service's build-filter paths, so `morph-utils` lines cannot satisfy them. Run it and confirm it still prints `container contract ok`.

## 2. Blueprint

- [x] 2.1 Add the `morph-utils` service to `render.yaml` per `design.md`. Do not add a disk, a `projects` block, embed services, or `REACT_APP_MORPH_UTILS_URL`. Re-run `sh deploy/check-container-contract.sh` and confirm it still prints `container contract ok`.

## 3. Runbook

- [x] 3.1 Add the MorphUtils Render section to `deploy/README.md` and `morph-utils/README.md` per `design.md`. Re-run `sh morph-utils/deploy/check-container-contract.sh` and confirm it prints `morph-utils container contract ok`.

## 4. Schema

- [x] 4.1 Validate `render.yaml` with `check-jsonschema` against `https://render.com/schema/render.yaml.json`. Do not edit `ci.yml` or add a MorphUtils image build to the Morph image workflow.
