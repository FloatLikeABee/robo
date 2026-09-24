# Proposal

## Why

Morph API and the Morph AI UI already build as one image, but nothing in the repo tells Render how to run that image. Issue #53 is the production deploy. The product owner will create the service from a Blueprint after merge. This change adds that Blueprint and the runbook, and does not call Render.

## What Changes

- Add a root `render.yaml` with one Docker web service for the existing root `Dockerfile`.
- Validate that file against Render's Blueprint schema inside the existing Morph image workflow, without a new check name and without editing `ci.yml`.
- Document how to create the service, which secrets to fill in, the disk's single-instance limit, how to snapshot `/data`, and how to check `/health`.
- Leave `deploy/docker-compose.yml`, `deploy/Caddyfile`, and the Dockerfile's runtime behavior unchanged.

## Capabilities

### New Capabilities

- `morph-render`: the Blueprint contract for the Morph image (one web service, disk, env inventory, schema check, Render runbook).

### Modified Capabilities

- `morph-container`: the operator runbook may name Render as the host the product owner creates from the Blueprint. It still must not send the operator through `scripts/deploy.sh`. The image workflow must also schema-check `render.yaml`.

## Impact

- New file: `render.yaml`.
- Edited: `.github/workflows/docker-image.yml`, `deploy/check-container-contract.sh`, `deploy/README.md`, `docs/agents/12-build-deploy.md`.
- No Go, frontend, or Rust behavior change. No Render API calls. Compose and the Dockerfile stay as they are.
