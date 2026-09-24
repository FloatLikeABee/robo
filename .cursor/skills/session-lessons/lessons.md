# Session Lessons

Durable rules from past chats. Newest first. Each lesson is 1–4 lines. Update via `learning-from-sessions`.

## 2026-09-24 — Content Maker identity headers are not a session

- Trigger: Content Maker `requireTranmailAccess`, `X-User-Role` / `X-User-Permissions` on `/templates`
- Rule: A Morph-validated `Authorization` bearer is the only session. Strip client identity headers. Header-only admin is 401. Call Morph `/api/auth/user` and `/api/auth/permissions`; do not decode the JWT in this process.
- Source: [PR 118 Lens](current)

## 2026-09-24 — Empty embed origin must not become a relative path

- Trigger: MorphUtils Event Logs `embedUrl` when `VITE_SHEETX_URL` / the resolved origin is empty
- Rule: Append `/events-info` only when the origin is non-empty. `` `${origin}/events-info` `` on `''` is the truthy path `/events-info`, and the shell iframes its own SPA. Leave `embedUrl` empty so the iframe is not mounted.
- Source: [containerize MorphUtils review](current)

## 2026-09-24 — Dockerfile HEALTHCHECK `$$` is the shell PID

- Trigger: Dockerfile `HEALTHCHECK` or other shell-form `CMD` using `$$` so a `$` reaches the shell
- Rule: `$$` is a Compose escape, not a Dockerfile one. Shell-form `HEALTHCHECK` is `/bin/sh -c`, so `$$PORT` requests `<pid>PORT`. Use `${PORT}` and let sh expand the runtime port. `deploy/check-container-contract.sh` rejects `$$` in the Dockerfile.
- Source: [containerize Morph](current)

## 2026-09-24 — morphai env copies are snapshots

- Trigger: `LoadFromEnv` then a named provider; `MORPH_AI_BASE_URL`, compatible `MORPH_AI_API_URL`, or a key that rotated after load
- Rule: Snapshot the key and base URL at load. The empty-provider path is the only sender. A named provider drops a field that still equals its snapshot and uses that provider's own env or default. Do not re-read legacy env to decide whether the field is caller-set. A base or key written on `Config` or the call is kept.
- Source: [morphai re-check base provenance](current)

## 2026-09-22 — Morph AI has no Files workspace

- Trigger: Morph AI agent shell Files tab, Open folder, recents, local-folder pins, composer Files chip, IndexedDB `morphai-files-workspace`
- Rule: Morph AI has no Files workspace. Context & Knowledge is the file/knowledge surface (HybridContext + Knowledge Library). Notes & TODOs stay. Do not restore a folder picker, recents, or pin-from-folder.
- Source: [morphai-drop-files-workspace](current)

## 2026-09-10 — Chat enlarge modal must fill the viewport

- Trigger: mermaid, pixel art, or images opening in the Morph AI / Event Logs / Content Maker enlarge overlay
- Rule: The overlay is a near-full viewport stage. Clone mermaid SVG without its bubble `width`/`height` so it scales to fit (`viewBox` + contain). Do not keep `width: auto` on the intrinsic pixel size — that leaves diagrams tiny.
- Source: [chat-visual-enlarge-modal](current)

## 2026-09-05 — MorphUtils Data Access down hint

- Trigger: MorphUtils `localhost:3040/datax` iframe “localhost refused to connect”
- Rule: MorphUtils probes `VITE_DATAX_URL` (`no-cors` fetch). If down, show an in-shell hint (`./start-all.sh start sharpreport-ui`; `start morph-utils` also starts Data Access API+UI) and do not mount the iframe. Refused still means `sharpreport-ui` is down — keep Vite `server.host: true`. API on `SHARPREPORT_PORT` can be up while 5178 is dead.
- Source: [fix-morphutils-project-data-access](current)

## 2026-09-02 — Morph AI webpack missing workspace modules

- Trigger: CRA `Can't resolve` `AiToolsWorkspaceDrawer`, `agentContext`, `AgentWorkspace`, `appliedAssistantChannel`, `ExtractJsonFromTextDialog`
- Rule: Those modules must exist as tracked `.js` files under `morph/frontend/src/` matching the extensionless imports in `SkoolAiChat.js` / `AdminDataGrid.js`. Do not leave them untracked `.jsx`. There is no `filesWorkspaceStore` / `AgentFilesTab` — do not restore them. Compile Morph AI after git sync.
- Source: [webpack missing modules](current)

## 2026-08-31 — MorphUtils /datax connection refused

- Trigger: MorphUtils `localhost:3040/datax` iframe “localhost refused to connect”
- Rule: See 2026-09-05 MorphUtils Data Access down hint (probe + start `sharpreport-ui`). Do not treat a down UI as a MorphUtils path/router bug.
- Source: [datax refused](current)

## 2026-08-30 — Dark only, no theme switch

- Trigger: light/dark toggle, theme switch, Morph AI / MorphNotes / MorphUtils chrome
- Rule: Product UIs stay dark. Do not add a light/dark switch. Ignore stored light preferences and force dark.
- Source: [dark-only](current)

## 2026-08-30 — Data Access 5178 empty 500s

- Trigger: MorphUtils Data tables, `GET /api/v1/auth/me` or `/data-tables` 500 with empty body
- Rule: Vite/SSR must proxy to `SHARPREPORT_PORT` (not hardcoded 3050). Empty 500 on 5178 means the proxy target is down, not a tables/auth handler bug.
- Source: [data-access-vite-proxy-port](current)

## 2026-08-30 — MorphUtils product name

- Trigger: sidebar, title bars, Apps menu, landing, login copy for the Utils shell
- Rule: User-facing name is MorphUtils (no space), same compound style as MorphNotes. Do not write "Morph Utils".
- Source: [MorphUtils naming](current)

## 2026-08-30 — Data Access uses Morph AI session

- Trigger: Data Access / DataX login, “Sign in with your platform account”, iframe SSO
- Rule: Reuse the Morph AI JWT (`userspanel_session_token` / `userspanel_token`). Do not show a second credential form. Do not clear the shared cookie when Data Access `/auth/me` is down or returns 502 — only a real 401 invalidates the session.
- Source: [Data Access Morph SSO](current)

## 2026-09-19 — Event Logs has no Info Sheets tab

- Trigger: Event Logs header tabs, MorphUtils Event Logs description, `/survey-bot`
- Rule: Event Logs is Events & Info only. Do not restore an Info Sheets / AI Surveys tab. Old `/survey-bot` paths redirect to `/events-info`.
- Source: [drop-info-sheets-graphs-skill-task-create](current)

## 2026-08-27 — Morph Utils Event Logs / Info Sheets

- Trigger: sidebar, title bars, browser titles, landing Utils list in Morph Utils and SheetX
- Rule: Module is Event Logs (not Survey Maker / SurveyX / SurveysX). Embed and default route `/events-info`. Keep id `sheetx`. Info Sheets tab is removed (see 2026-09-19).
- Source: [morphutils-event-logs-info-sheets](b61bf1f4-55f9-46bb-9ee0-3f82db385107)

## 2026-08-17 — Verify Morph UI before claiming done

- Trigger: Morph frontend or Morph Utils/SheetX layout changes
- Rule: Do not claim complete until the touched UI compiles. `await` only inside `async` functions. Event Logs / MorphUtils `Layout` must compile without `@robo/platform-chat` (that package is gone; Morph AI is the system chat).
- Source: [MorphAI session extract](b61bf1f4-55f9-46bb-9ee0-3f82db385107)

## 2026-08-09 — Morph Utils product labels

- Trigger: sidebar, title bars, login copy, browser titles in Morph Utils and embedded apps
- Rule: User-facing names are Event Logs (not Survey Maker/SurveyX/SurveysX), Content Maker (not ComposerX), Data Access (not DataX), Project, Docs. No subtitle or email under the app title.
- Source: [Utils branding renames](816ba854-be27-42b0-aed3-b430e3fd1f5b)

## 2026-08-03 — Morph Utils navigation scope

- Trigger: "left side menu" / tabs in Morph Utils
- Rule: Keep the outer Morph Utils left nav. Convert each **module app's** inner left menu to header tabs. Do not replace the outer shell nav.
- Source: [Utils inner tabs correction](816ba854-be27-42b0-aed3-b430e3fd1f5b)

## 2026-06-10 — Form page names stay empty

- Trigger: SheetX/FormsX question pages
- Rule: Default is one unnamed page. If the user does not name a page, leave the name empty and do not display a label. Do not invent "Page 1".
- Source: [FormsX question pages](0404bdb5-5fb3-4347-b5c6-7764f5d75fba)

## Search history with short queries

- Trigger: looking up past chats
- Rule: `SearchConversations` ANDs unquoted keywords onto one conversation. Use 1–2 keywords per call and several parallel searches. Prefer hits that have `[correction]` or `[error-report]` turns. Then run `extract-session.py <id>` instead of reading the raw JSONL.
