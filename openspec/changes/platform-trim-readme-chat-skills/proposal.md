## Why

The root README and launcher still describe apps that are gone (Booki, Academi, `platform-chat` as a product). Satellite chat drawers, Video stories, Morph AI header shortcuts, and a cramped Skills form no longer match how operators actually work: Morph AI is the system chat, config is one root `.env`, and Skills need a proper dark editor.

## What Changes

- Rewrite root `README.md` so it lists only remaining products and URLs (Morph AI / MorphNotes, MorphUtils modules, Event Logs, Content Maker, Data Access, Project, AI tools). Drop Booki, Academi, UsersPanel, `morph-broadcast`, and other removed folders. Align `start-all.sh` service lists with those folders.
- Document that **one repo-root `.env`** is the only local config (`cp .env.example .env`). Nested leftover `.env` files stay ignored; production stays `deploy/.env.production`.
- **BREAKING (satellite UIs):** Delete `platform-chat/`. Remove the shared assistant drawer from Event Logs, Content Maker, Data Access, and Project. Morph AI chat remains the platform assistant. Copy `aiProgress` locally where compose/progress UI still needs it.
- **BREAKING (AI tools):** Remove Video stories / video generator end to end (nav, routes, pages, API, services).
- Remove AI tools **Open in tab** (and the error-state new-tab link).
- Morph AI header: remove the Notes & TODOs, Context & Knowledge, and Export session (JSON) icon buttons. Keep Skills, AI tools, app links, clear chat, sign out. Notes/knowledge remain in MorphNotes and the Files workspace tabs.
- Skills modal: upload `.md` as well as the form; wider dark dialog; upload form (left) and catalog (right) side by side; catalog scrolls vertically; longer instructions field; **Improve with AI** reads Name, Description, and Instructions and fills a draft the operator can save.

## Capabilities

### New Capabilities

- `current-stack-docs`: README (and launcher names) match the apps that still exist; one root `.env` is documented as the local config.
- `remove-platform-chat`: The shared `platform-chat` package and satellite assistant drawers are gone; Morph AI is the system chat.
- `remove-video-stories`: AI tools no longer includes Video stories / video generator.
- `morphai-chrome-simplify`: Morph AI header drops notes, knowledge, and JSON-export shortcuts; AI tools has no Open in tab.
- `morphai-skills-editor`: Skills upload supports markdown files, a wide split dark layout, and an AI improve pass before save.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines for these)

## Impact

- Root `README.md`, `start-all.sh` (drop missing Booki/Academi services), optional pointer-only nested `.env.example` files.
- Delete `platform-chat/`; update `formx`, `composerx`, `SharpReport`, `morph-engi` package.json and assistant drawers; local `aiProgress` copies.
- `bk/` frontend Header/App routes and backend `/video-stories*` plus `video_story_*` modules.
- Morph AI: `SkoolAiChat.js` header, `AiToolsWorkspaceDrawer.jsx`, `SkillsModal.js` / CSS, Morph `POST /api/skills` (md upload) and a skill-improve AI endpoint.
- Out of scope: MorphNotes Notes & TODOs page; Files workspace notes/knowledge tabs; light/dark theme switch; production `deploy/.env.production`.
