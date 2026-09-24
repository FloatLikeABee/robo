## Why

MorphUtils is a Vite shell that only runs from a checkout (`npm run dev` on port 3040). Production needs an image that serves the built shell, answers a health check without the sibling apps, and calls Morph auth at a configured origin instead of a localhost fallback baked into the bundle.

## What Changes

- Add a MorphUtils-only image (Dockerfile, ignore file, optional compose) that serves the production Vite build and returns success on `/health` and `/ready`.
- Make the Morph API base and embed origins configurable from the environment at container start, with optional build-args for the same public URLs. The production bundle must not fall back to localhost.
- Document how an operator builds and runs this shell alone. Secrets stay out of image layers.
- Leave the Morph root image, `render.yaml`, and `REACT_APP_MORPH_UTILS_URL` unchanged.

## Capabilities

### New Capabilities

- `morph-utils-container`: Production image and env surface for the MorphUtils Vite shell.

### Modified Capabilities

- (none — `morph-container` still excludes MorphUtils from the Morph image)

## Impact

- `morph-utils/frontend` URL resolution (`auth.ts`, `config.ts`) and `index.html`.
- New files under `morph-utils/` (image, compose, runbook section in `morph-utils/README.md`).
- No Render API calls, no Blueprint edit, no other app images, no Morph header chip URL.
