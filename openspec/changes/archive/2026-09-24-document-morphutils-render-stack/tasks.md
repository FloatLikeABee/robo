# Tasks

## 1. Contract check

- [x] 1.1 Add `deploy/check-stack-runbook.sh` that fails unless `deploy/README.md` has a `MorphUtils stack on Render` section stating `Forge must not create services on Render`, project `prj-dahc33dbedkc73a1v8n0`, `does not call Render`, services `morph-utils`, `formx`, `composerx`, `sharpreport`, and `morph-engi`, placeholders `https://<morph-utils public host>`, `https://<event-logs public host>`, `https://<composerx public host>`, `https://<sharpreport public host>`, and `https://<morph-engi public host>`, keys `REACT_APP_MORPH_UTILS_URL`, `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, `VITE_MORPH_ENGI_URL`, and `USERS_PANEL_BASE_URL`, example hosts `https://morph-utils.onrender.com`, `https://formx-vucj.onrender.com`, `https://composerx.onrender.com`, `https://sharpreport.onrender.com`, `https://morph-engi.onrender.com`, and `https://morph-gjmb.onrender.com`, story `#114`, and that a Morph image rebuild is required while a restart without a rebuild does not set the header link. The MorphUtils origin record must appear before the rebuild step. The script must also fail if `render.yaml` contains `key: REACT_APP_MORPH_UTILS_URL` or any of those example hosts. Run it and confirm it fails because the section is absent.

## 2. Runbook

- [x] 2.1 Add `## MorphUtils stack on Render` to `deploy/README.md` per `design.md`, after the Morph "After a deploy" section and before `## MorphUtils on Render`. Verify with `sh deploy/check-stack-runbook.sh` printing success.
- [x] 2.2 Point `docs/agents/12-build-deploy.md` at that heading, and add one sentence to `morph-utils/README.md` that this Blueprint section does not set `REACT_APP_MORPH_UTILS_URL` and that the ordered rebuild is in `deploy/README.md`. Verify both files contain `MorphUtils stack on Render` and `sh morph-utils/deploy/check-container-contract.sh` still prints `morph-utils container contract ok`.

## 3. CI hook

- [x] 3.1 Run `deploy/check-stack-runbook.sh` from `.github/workflows/docker-image.yml` next to `deploy/check-container-contract.sh`. Verify the workflow file contains the script name and `sh deploy/check-stack-runbook.sh` exits 0.
