## Why

Production MorphUtils still cannot be used end to end. The shell can read embed origins from `/config.js`, but the Blueprint and runbooks still tell the product owner not to set them, and a signed-in Morph session does not cross `*.onrender.com` hosts. Story #114 wires those origins and one shared Morph sign-in without a second signup, without baking live hosts into source.

## What Changes

- List the four embed origin variables on the `morph-utils` service as dashboard prompts with no committed value, so a restart rewrites `/config.js` and the iframe `src` values are those HTTPS origins.
- Keep localhost embed defaults on `npm run dev` only. A production shell with a blank origin does not mount that iframe.
- Authenticate embeds with the existing Morph bearer handoff (`?userspanel_token=`), not a parent-domain cookie. Document why `Domain` and `SameSite` cannot share a session across `*.onrender.com`.
- When MorphUtils has no session, show a sign-in link to the configured Morph origin and do not mount embeds. Direct embed visits must not send the user to a production `localhost` Morph URL.
- Leave `bk` and invite-signup out of this story.

## Capabilities

### New Capabilities

- `morphutils-embed-auth`: Production MorphUtils iframe origins, Morph token handoff, and the unauthenticated path back to Morph.

### Modified Capabilities

- `morph-utils-render`: Embed origin keys become prompts on `morph-utils` (no value in git) instead of staying unset.
- `morphutils-stack-render-runbook`: The env matrix is filled on `morph-utils` by this story, including cookie `Domain` / `SameSite` notes.
- `formx-render`: `VITE_SHEETX_URL` / `VITE_FORMSX_URL` may be prompts on `morph-utils` only.
- `composerx-render`: `VITE_COMPOSERX_URL` may be a prompt on `morph-utils` only.
- `data-access-render`: `VITE_DATAX_URL` may be a prompt on `morph-utils` only.
- `morph-engi-render`: `VITE_PROJECTS_URL` / `VITE_MORPH_ENGI_URL` may be prompts on `morph-utils` only.

## Impact

- `render.yaml` `morph-utils` env prompts, `deploy/README.md`, `morph-utils/README.md`, and the sibling READMEs that still say story #114 has not set the embed variables.
- Container contract checks that currently reject those keys anywhere in `render.yaml`.
- `morph-utils/frontend` URL and session helpers, plus the unauthenticated shell. Event Logs, Content Maker, Data Access, and Project keep validating bearers at `USERS_PANEL_BASE_URL`. Their production login links must not target loopback.
- No Render API calls, no new services, no secrets, no Morph SPA redesign, no `bk` or invite-signup deploy.
