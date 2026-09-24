## Context

MorphUtils already resolves embed origins from runtime `/config.js`, then build-time `VITE_*`, and uses localhost only when `import.meta.env.DEV` is true (`publicUrl.ts`). A blank production origin stays blank, so Event Logs does not become a truthy `/events-info` on the shell itself. The entrypoint already writes those keys into `/config.js`. What still blocks production is the Blueprint and the runbooks, which say not to set the keys, plus auth that cannot cross separate `*.onrender.com` sites.

`auth.ts` stores `userspanel_session_token` as a host-only `SameSite=Lax` cookie and appends `?userspanel_token=` on iframe `src`. Morph's Apps link already appends that query when opening MorphUtils. Each embed copies the query into its own host cookie and then calls Morph with `Authorization: Bearer` at `USERS_PANEL_BASE_URL`. The shell's unauthenticated control is a sidebar link to `VITE_MORPH_AI_URL`, which is `/` when that variable is blank, so it does not leave MorphUtils.

## Goals / Non-Goals

**Goals:**

- A configured production shell iframes the four live origins, not localhost.
- One Morph sign-in is enough for the shell and the embeds it mounts.
- An unauthenticated shell shows a Morph sign-in link and does not mount embeds.
- Docs list the env vars and the cookie failure modes.

**Non-Goals:**

- Render API calls, new services, or committed `onrender.com` hosts.
- A Morph SPA return-URL flow.
- `bk` and invite-signup.
- A shared cookie `Domain` on `onrender.com`.

## Decisions

### 1. Token handoff stays the session strategy

The shell keeps passing `userspanel_token`. The cookie stays host-only: `Path=/; SameSite=Lax`, no `Domain`. Embeds already accept that query and send the bearer to Morph.

**Rejected: parent-domain cookie (`Domain=.onrender.com`, `SameSite=None; Secure`).** `onrender.com` is a public suffix. Browsers refuse a `Domain` on it. Each service is its own site, so a cookie set on Morph is not sent to MorphUtils or to an embed. `SameSite=None` does not create a parent the operator does not control.

**Rejected: one reverse proxy for every app.** The services are already separate origins. A gateway is a new deploy shape and is out of scope.

**Rejected: Morph SPA `?next=` redirect.** That needs a Morph SPA change. After the visitor signs in on Morph, the existing Apps link appends the bearer. The extra click is the documented path. Quill is not required.

### 2. Embed origins are dashboard prompts, not source defaults

`render.yaml` lists `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, `VITE_MORPH_ENGI_URL`, and `VITE_MORPH_AI_URL` on `morph-utils` only, each `sync: false` and without a `value`. No `onrender.com` host is committed. The entrypoint already emits them. Contract checks that forbid the keys anywhere in the file change to: allowed on `morph-utils` as empty prompts, forbidden on the other services.

Example hosts in the runbook stay labeled examples. Application source does not default to them.

**Rejected: hardcode the live hosts in `config.ts`.** The issue forbids required production hosts in app source, and a later dashboard URL would be wrong until the next image build.

### 3. Unauthenticated shell does not mount embeds

`morphLoginHref` uses `VITE_MORPH_AI_URL`, then `VITE_MORPH_API_URL`. A loopback value is ignored outside dev. The main pane shows that link (`target="_top"`). Iframes mount only after a bearer is present, so an unsigned visitor is not dropped into a second signup form inside the frame.

Direct embed visits still cannot see Morph's host cookie. That is a documented failure mode. Data Access and Project MUST NOT point a production sign-in link at `localhost` when no Morph origin is configured. Their dev default stays `http://localhost:3031`. Event Logs and Content Maker already sign in with the Morph account through their own API; this story does not add a second form and does not inject `USERS_PANEL_BASE_URL` into those HTML documents.

**Rejected: inject the Morph origin into all four SPA servers.** `ServeDir` / static file routes would each need an HTML rewrite, and the shell already withholds the iframe until the bearer exists. The direct-visit gap is documented instead of a four-server injection.

### 4. Remote embed down hint is not a local start command

Data Access and Project are still probed. A loopback miss keeps the `./start-all.sh` hint. A non-loopback miss says the origin is unreachable and offers Retry. It does not tell a production operator to start a laptop process.

## Risks / Trade-offs

- [Bearer in the first iframe URL] → Each app strips `userspanel_token` with `replaceState`. The first request can still be logged. Docs say so. A parent cookie cannot replace this on `*.onrender.com`.
- [Sign-out is per host] → Clearing the shell cookie does not clear embed cookies. Docs say so. A global logout endpoint is out of scope.
- [Bookmarking MorphUtils while already signed in on Morph] → The shell does not see Morph's cookie. The sign-in link sends the visitor to Morph; the Apps link brings the bearer back. No silent SSO.
- [Prompt keys with no value] → Unset origins leave those iframes unmounted. The runbook says to fill the prompts and restart, not rebuild.
- [`SameSite=Lax` without `Secure`] → Dev `http://localhost` can store the cookie. Production is HTTPS. Do not add `Secure` in this change or local sign-in stops sticking.

## Migration Plan

The product owner fills the new `morph-utils` prompts with the dashboard origins and restarts that service. `/config.js` updates without an image rebuild. Rollback is clearing those prompts and restarting; the shell then mounts no embeds. No data migration.

## Open Questions

None. The session strategy, the prompt shape, and the login path are fixed above.
