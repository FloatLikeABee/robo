## 1. Contract checks first

- [x] 1.1 Extend `dorph-utils/deploy/check-container-contract.sh` so it fails unless `render.yaml` has a `dorph-utils` Docker web service (Singapore, starter, `main`, `./dorph-utils/Dockerfile`, context `./dorph-utils`, `/health`, `checksPass`), `PORT` `3040`, `VITE_DORPH_API_URL` with `sync: false` and no value, the image `buildFilter` paths, no disk on that service, and service names only `dorph` and `dorph-utils`. It must also fail unless `deploy/README.md` and `dorph-utils/README.md` name project `prj-dahc33dbedkc73a1v8n0`, `https://<dorph public host>`, `https://<dorph-utils public host>`, `REACT_APP_DORPH_UTILS_URL`, the alias and embed variables, and that CORS already allows `Authorization` and is not edited. Run it and confirm it fails because the service is absent.
- [x] 1.2 Scope `deploy/check-container-contract.sh` to the `dorph` service block for env keys, `dockerContext: .`, `dockerfilePath: ./Dockerfile`, and that service's build-filter paths, so `dorph-utils` lines cannot satisfy them. Run it and confirm it still prints `container contract ok`.

## 2. Blueprint

- [x] 2.1 Add the `dorph-utils` service to `render.yaml` per `design.md`. Do not add a disk, a `projects` block, embed services, or `REACT_APP_DORPH_UTILS_URL`. Re-run `sh deploy/check-container-contract.sh` and confirm it still prints `container contract ok`.

## 3. Runbook

- [x] 3.1 Add the DorphUtils Render section to `deploy/README.md` and `dorph-utils/README.md` per `design.md`. Re-run `sh dorph-utils/deploy/check-container-contract.sh` and confirm it prints `dorph-utils container contract ok`.

## 4. Schema

- [x] 4.1 Validate `render.yaml` with `check-jsonschema` against `https://render.com/schema/render.yaml.json`. Do not edit `ci.yml` or add a DorphUtils image build to the Dorph image workflow.
