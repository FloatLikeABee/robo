## Context

See proposal.md for motivation. MorphUtils always mounts hidden iframes for every module (`App.tsx` `ModulePanel`). Data Access `src` is `VITE_DATAX_URL` (default `http://localhost:5178`). Project is `VITE_PROJECTS_URL` (`:5179`). `start-all.sh` `morph-utils` currently starts only `morph-utils-ui`. Data Access Vite already uses `server.host: true` on 5178; Project Vite does not set `host`. Platform-chat was removed as a product; Project still imports it. Event Logs and Content Maker already dropped the drawer in source.

## Goals / Non-Goals

**Goals:**
- Project Vite transform succeeds without `@robo/platform-chat`.
- MorphUtils Data Access never shows a connection-refused iframe document.
- `start morph-utils` brings Data Access UI+API up with the shell.
- Project iframe can use IPv4/IPv6 localhost when `morph-engi-ui` is running.

**Non-Goals:**
- Deleting the leftover `platform-chat/` tree or cleaning other apps’ unused `package.json` entries.
- Starting Event Logs / Content Maker / Project UIs from `start morph-utils` (only Data Access is required).
- Changing Data Access API auth, proxy port, or Morph SSO cookies.
- Restoring an in-app Projects AI drawer.

## Decisions

### 1. Delete the Project drawer instead of restoring the package

**Choice:** Remove `PlatformAssistantDrawer.svelte`, the header button, `getStateExtra` / `buildAiStateExtra` wiring if unused, and the `file:../../platform-chat` dependency. Point operators to Morph AI (`:3031`).

**Why:** Matches `remove-platform-chat` and Morph AI-only chat. Restoring the package would bring back a satellite drawer the rest of MorphUtils already dropped.

**Alternative:** Keep the package and fix the Vite alias — rejected.

### 2. Probe with `fetch` + `no-cors`, then mount iframe

**Choice:** Before setting Data Access iframe `src`, `fetch(origin, { method: 'GET', mode: 'no-cors', cache: 'no-store' })`. Network failure → in-shell hint with `./start-all.sh start sharpreport-ui`. Success (opaque) → existing `withSessionToken` iframe. Retry control on the hint. Reuse the same helper for Project if cheap (not required by spec).

**Why:** Cross-origin Vite pages usually lack CORS for a MorphUtils `fetch`; `no-cors` still fails when TCP refuses, which is this bug. Iframe `onerror` is unreliable for connection refused.

**Alternative:** Always iframe and overlay CSS — still flashes the browser error page. **Alternative:** MorphUtils Vite proxy to 5178 — changes embed origin and SSO; rejected.

### 3. Expand `resolve_services morph-utils`

**Choice:** `morph-utils` / `utils` starts `morph-utils-ui sharpreport-api sharpreport-ui`.

**Why:** Spec requires API as well as UI so tables/auth work, not only a Vite process. Operators who only start MorphUtils today get a refused Data Access iframe.

**Alternative:** Document-only — rejected; they already hit refused connect.

### 4. Project `server.host: true`

**Choice:** Match Data Access (`host: true`, keep port 5179). Do not change port.

**Why:** Same IPv6/`localhost` iframe class as the Data Access lesson, once Project compiles.

## Risks / Trade-offs

- [False “down” if fetch is blocked by mixed content / browser] → Probe only same-http localhost origins already used as iframe src; hint still tells them to start `sharpreport-ui`.
- [False “up” then iframe still fails] → Retry stays on the hint; operator can hard-refresh.
- [`start morph-utils` now compiles/starts Rust Data Access API] → Slower first start; acceptable vs a dead iframe.
- [Other apps still list `@robo/platform-chat`] → Out of scope; Project no longer depends on it.

## Migration Plan

1. Drop Project platform-chat usage and dependency; `npm install` in `morph-engi/frontend`.
2. Ship MorphUtils probe + hint; expand `start-all.sh`.
3. Restart `morph-engi-ui`, `morph-utils-ui`, Data Access via `start morph-utils`.
4. Rollback: revert those three trees; launcher again starts only the shell.

## Open Questions

None.
