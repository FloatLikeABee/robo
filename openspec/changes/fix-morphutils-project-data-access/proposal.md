## Why

MorphUtils **Project** fails to compile because leftover `@robo/platform-chat` imports cannot resolve (`PlatformAssistantDrawer.svelte`). MorphUtils **Data Access** shows the browser’s “localhost refused to connect” because the iframe always loads `VITE_DATAX_URL` (`localhost:5178`) even when `sharpreport-ui` is not listening. Operators cannot use those two modules until Project builds and Data Access either is up or explains how to start it.

## What Changes

- Remove Project’s satellite chat drawer, **AI Assistant** header button, and `@robo/platform-chat` dependency so Vite can serve `morph-engi-ui` (`:5179`).
- Morph AI remains the system chat; do not restore `platform-chat` drawers.
- MorphUtils Data Access MUST NOT leave operators on a blank refused iframe. Probe `VITE_DATAX_URL` and, if the origin is down, show an in-shell message to start Data Access (`./start-all.sh start sharpreport-ui`) instead of mounting a dead iframe.
- `start-all.sh start morph-utils` MUST also start `sharpreport-ui` (and `sharpreport-api` if that is required for a useful Data Access session) so opening MorphUtils Data Access does not depend on a separate remembered command.
- Bind Project Vite `server.host` the same way Data Access already does (`host: true`) so the MorphUtils iframe can reach `:5179` on IPv4 and IPv6 localhost.

## Capabilities

### New Capabilities

- `project-no-platform-chat`: Project (morph-engi) compiles and embeds in MorphUtils without `@robo/platform-chat` or a floating assistant overlay.
- `data-access-embed-available`: MorphUtils Data Access reaches a listening UI on `VITE_DATAX_URL`, or shows a start hint; the MorphUtils launcher starts Data Access UI with the shell.

### Modified Capabilities

- (none)

## Impact

- `morph-engi/frontend`: `AppLayout.svelte`, `App.svelte`, delete `PlatformAssistantDrawer.svelte`, drop `@robo/platform-chat` from `package.json` / lockfile, `vite.config.ts` host binding. `lib/aiContext.ts` only if nothing else needs it after the drawer is gone.
- `morph-utils/frontend`: Data Access (and optionally Project) embed preflight instead of an always-on iframe.
- `start-all.sh` `resolve_services` for `morph-utils`.
- Does not delete the leftover `platform-chat/` package tree in this change (other `package.json` files still list it). Does not change Data Access API routes or Morph SSO cookies.
