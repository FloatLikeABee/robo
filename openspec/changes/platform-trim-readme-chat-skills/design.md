## Context

See proposal.md for motivation. Remaining apps: `morph/`, `morph-utils/`, `formx/`, `composerx/`, `morph-engi/`, `SharpReport/`, `bk/`, `pkg/`, `landing/`, `morphgraph-worker/`. Booki and Academi folders are gone but `start-all.sh` and README still list them. Local config is already root `.env` via `load_root_env`; README still mentions Booki/Academi AI config. `@robo/platform-chat` is used as a drawer in formx/composerx/SharpReport/morph-engi and as `aiProgress` in compose/publish. AI tools (`bk`) still ships Video stories. Morph AI header in `SkoolAiChat.js` has notes, knowledge, and JSON export icons. Skills modal is a ~720px stacked form (`SkillsModal.js`). Dark-only product UIs. MorphUtils left nav and ids `sheetx` / `/survey-bot` stay.

## Goals / Non-Goals

**Goals:**

- README + default launcher = remaining stack; one `.env` documented.
- Delete `platform-chat/` after drawers are gone and `aiProgress` is copied.
- Strip Video stories from `bk`.
- Slim Morph AI header and AI tools chrome.
- Skills split editor + md upload + AI improve (draft only).

**Non-Goals:**

- Removing MorphNotes Notes & TODOs or the agent workspace Notes & TODOs and Context & Knowledge tabs.
- Deleting Graphic Documents / image generation unless it is only used by Video stories.
- Changing production `deploy/.env.production`.
- Adding a light/dark switch.

## Decisions

### 1. Answer on platform-chat: delete it

**Choice:** Remove the package and every `PlatformChatDrawer` mount. Morph AI is the assistant. Copy `aiProgress` (from `morph/frontend/src/lib/aiProgress.js` or the current package) into formx/composerx/SharpReport as needed for progress labels only. Drop `PlatformAssistantDrawer` components that exist only to wrap the shared drawer.

**Why:** The package exists so satellites can chat; operators asked to keep system chat only. Keeping a package with one helper is not worth it.

**Alternative:** Keep `platform-chat` for `aiProgress` only — rejected (still a workspace package to install).

### 2. README and launcher in the same pass

**Choice:** Rewrite the README folders/URLs/env sections. Remove `booki-*` and `academi-*` from `ALL_SERVICES`, aliases, install, and list. Nested `.env.example` files stay one-line pointers (already true for morph/formx).

**Why:** Docs that list missing apps make `./start-all.sh` fail. Config is already one file; the gap is README (and leftover Booki/Academi paragraphs).

**Alternative:** README-only — rejected; launcher would still try to start missing folders.

### 3. Video stories: delete product surface and API

**Choice:** Remove `VideoStoryGenerator` page, Header item, routes, `api.js` video-story helpers, `_setup_video_story_routes`, and `video_story_manager` / `video_story_service` (and tests). Redirect `/video-stories` to `/assistants`. Keep Graphic Documents if it is independent of video stories.

**Why:** User asked to remove video generator entirely.

**Alternative:** Hide nav only — rejected (API would remain).

### 4. Skills AI improve is a Morph endpoint, not a silent save

**Choice:** `POST /api/skills/improve` with `{ name, description, instructions }` returns the same shape. UI button fills the form. Parse `.md`: first `# heading` → name, remainder → instructions; filename stem if no heading. Modal CSS: `width: min(1100px, 96vw)`, grid two columns, catalog `overflow-y: auto; min-height: 0`. Instructions textarea `min-height: 16rem`. Force dark tokens (ignore `html[data-chat-theme=light]` for this dialog).

**Why:** Save stays explicit; AI is a draft step.

**Alternative:** Improve on the client by calling generic `/api/chat` — extra session/message coupling; a small dedicated route is clearer.

### 5. Header removals only

**Choice:** Delete the three icon buttons and `handleExportSession` if unused. Keep `ChatNotesTodosDrawer` / hybrid drawer code for the agent workspace (Notes & TODOs and Context & Knowledge) and MorphNotes. Remove Open in tab from `AiToolsWorkspaceDrawer` (including error fallback link).

**Why:** User named the header bar, not MorphNotes.

## Risks / Trade-offs

- [Event Logs / Content Maker lose in-app assistant] → Operators use Morph AI; document in README.
- [Orphan video-story files on disk] → Stop serving routes; leftover data dirs can remain until an operator deletes them.
- [Skill improve needs Morph AI key] → Same as chat; show the API error on the form.
- [start-all aliases for booki/academi] → Remove so `start booki` fails closed instead of spawning missing paths.

## Migration Plan

- Deploy Morph frontend + API together for Skills improve and header.
- Deploy formx/composerx/SharpReport/morph-engi frontends after removing `@robo/platform-chat`.
- Deploy bk API+UI together when Video stories is gone.
- Rollback: revert those apps; `platform-chat` deletion is hard to roll back without git.

## Open Questions

None.
