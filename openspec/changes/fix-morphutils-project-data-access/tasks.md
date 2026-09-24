## 1. Project compiles without platform-chat

- [x] 1.1 Delete `morph-engi/frontend/src/components/PlatformAssistantDrawer.svelte`
- [x] 1.2 Remove the AI Assistant button, drawer mount, and `@robo/platform-chat` types from `AppLayout.svelte`
- [x] 1.3 Stop passing `getStateExtra` from `App.svelte`; delete `getAiStateExtra` if unused; remove `lib/aiContext.ts` if nothing else imports it
- [x] 1.4 Drop `@robo/platform-chat` from `morph-engi/frontend/package.json` and refresh the lockfile
- [x] 1.5 Set Project Vite `server.host` / `preview.host` to `true` (keep port 5179)
- [x] 1.6 Confirm `npm run build` in `morph-engi/frontend` succeeds with no `@robo/platform-chat` resolve error

## 2. MorphUtils Data Access embed probe

- [x] 2.1 Add a `no-cors` fetch probe for an embed origin; treat network failure as down
- [x] 2.2 When Data Access is down, render an in-shell hint naming Data Access and `./start-all.sh start sharpreport-ui` (Retry); do not set iframe `src` to the dead origin
- [x] 2.3 When the probe succeeds, mount the existing session-token iframe as today
- [x] 2.4 Apply the same probe to Project if the helper is shared (optional vs spec)

## 3. Launcher

- [x] 3.1 Change `resolve_services` so `morph-utils` / `utils` starts `morph-utils-ui sharpreport-api sharpreport-ui`
- [x] 3.2 Update MorphUtils README / `start-all.sh list` copy if it still says MorphUtils starts only the shell

## 4. Verify

- [x] 4.1 Restart `morph-engi-ui`; open MorphUtils Project — no Vite `@robo/platform-chat` overlay; no Projects AI drawer
- [x] 4.2 With Data Access UI stopped, open MorphUtils Data Access — in-shell hint, not “localhost refused to connect”
- [x] 4.3 Run `./start-all.sh start morph-utils` (or start the three services); open Data Access — iframe loads
