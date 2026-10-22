## Why

DorphUtils is a Vite shell that only runs from a checkout (`npm run dev` on port 3040). Production needs an image that serves the built shell, answers a health check without the sibling apps, and calls Dorph auth at a configured origin instead of a localhost fallback baked into the bundle.

## What Changes

- Add a DorphUtils-only image (Dockerfile, ignore file, optional compose) that serves the production Vite build and returns success on `/health` and `/ready`.
- Make the Dorph API base and embed origins configurable from the environment at container start, with optional build-args for the same public URLs. The production bundle must not fall back to localhost.
- Document how an operator builds and runs this shell alone. Secrets stay out of image layers.
- Leave the Dorph root image, `render.yaml`, and `REACT_APP_DORPH_UTILS_URL` unchanged.

## Capabilities

### New Capabilities

- `dorph-utils-container`: Production image and env surface for the DorphUtils Vite shell.

### Modified Capabilities

- (none — `dorph-container` still excludes DorphUtils from the Dorph image)

## Impact

- `dorph-utils/frontend` URL resolution (`auth.ts`, `config.ts`) and `index.html`.
- New files under `dorph-utils/` (image, compose, runbook section in `dorph-utils/README.md`).
- No Render API calls, no Blueprint edit, no other app images, no Dorph header chip URL.
