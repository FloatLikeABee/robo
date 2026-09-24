# Proposal

## Why

Production Morph still hides the MorphUtils header chip. PR #93 omits the link when `REACT_APP_MORPH_UTILS_URL` is unset or loopback, and the root Dockerfile already accepts that name as a build arg, but the Morph Blueprint never declares it. A comment does not become a Docker build arg, so a Render rebuild inlines nothing and the chip stays off. MorphUtils now has a public HTTPS origin the product owner can copy. This change makes that build arg explicit without committing a host.

## What Changes

- Declare `REACT_APP_MORPH_UTILS_URL` on the `morph` service only, as a dashboard prompt (`sync: false`, no value). Render passes service env vars into the Docker build. The product owner sets `https://<morph-utils public host>` in the dashboard and rebuilds the Morph image.
- Keep the Dockerfile arg with no default host. Pass the same arg through local compose only when the shell sets it; an empty value still omits the chip.
- Check a production CRA bundle: a non-loopback URL is inlined when the arg is set, and the chip stays omitted when it is unset.
- Update the stack runbook so it points at that prompt. Do not commit `https://morph-utils.onrender.com` or any other host into app code or `render.yaml`.

## Capabilities

### New Capabilities

- `morph-utils-header-url`: Production Morph inlines a non-loopback MorphUtils base URL from the build arg and keeps the header chip omitted when that arg is unset or loopback.

### Modified Capabilities

- `morph-render`: The `morph` env list includes `REACT_APP_MORPH_UTILS_URL` as `sync: false` with no value.
- `morph-container`: The image build accepts that public URL as a non-secret build arg and does not default it to a host.
- `morphutils-stack-render-runbook`: The create order tells the product owner to fill the Morph dashboard prompt and rebuild. The Blueprint still has no value for the key.
- `morph-utils-render`: The `morph-utils` service still does not list the key. The MorphUtils section no longer claims the Morph service leaves it unset.
- `formx-render`: Event Logs still does not list the key. The Blueprint may list the Morph prompt.
- `composerx-render`: Content Maker still does not list the key. The Blueprint may list the Morph prompt.

## Impact

- `render.yaml` (`morph` env only), root `Dockerfile` (already has the arg), `deploy/docker-compose.yml`, `deploy/README.md`, `docs/agents/12-build-deploy.md`, `morph-utils/README.md`, and the container contract checks that currently forbid the key anywhere in the Blueprint.
- Morph frontend production build output. The header helper from PR #93 stays the gate. No Go or Rust behavior change. No Render API calls. No secrets.
