# Session Lessons

Durable rules from past chats. Newest first. Each lesson is 1–4 lines. Update via `learning-from-sessions`.

## 2026-09-02 — Morph AI webpack missing workspace modules

- Trigger: CRA `Can't resolve` `AiToolsWorkspaceDrawer`, `agentContext`, `filesWorkspaceStore`, `AgentWorkspace`, `appliedAssistantChannel`, `ExtractJsonFromTextDialog`
- Rule: Those modules must exist as tracked `.js` files under `morph/frontend/src/` matching the extensionless imports in `SkoolAiChat.js` / `AdminDataGrid.js`. Do not leave them untracked `.jsx` — `git stash -u` / reset drops them and webpack breaks. Compile Morph AI after git sync.
- Source: [webpack missing modules](current)

## 2026-08-31 — MorphUtils /datax connection refused

- Trigger: MorphUtils `localhost:3040/datax` iframe “localhost refused to connect”
- Rule: The shell is MorphUtils; Data Access is the iframe at `VITE_DATAX_URL` (`localhost:5178`). Refused means `sharpreport-ui` is down or bound IPv6-only — start it (`./start-all.sh start sharpreport-ui`) and keep Vite `server.host: true`. Do not treat it as a MorphUtils route bug. API on `SHARPREPORT_PORT` (e.g. 3888) can be up while 5178 is dead.
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

## 2026-08-27 — Morph Utils Event Logs / Info Sheets

- Trigger: sidebar, title bars, browser titles, landing Utils list in Morph Utils and SheetX
- Rule: Module is Event Logs (not Survey Maker / SurveyX / SurveysX). Inner tabs: Events & Info then Info Sheets (not AI Surveys). Embed and default route `/events-info`. Keep ids `sheetx` and `/survey-bot`.
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
