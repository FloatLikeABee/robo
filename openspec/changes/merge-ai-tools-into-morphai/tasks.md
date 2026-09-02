> **Amendment (iframe pivot):** Per follow-up, the AI Tools drawer now embeds the full bk UI via iframe so ALL modules (Assistants, RAG, Video Stories, Documents, System) are available, and the header "AI tools" link opens the drawer instead of navigating to `localhost:3000`. Groups 1 and 3 (native proxy endpoints + native chat windows) were implemented then removed in favor of the iframe; they are marked superseded below.

## 1. Morph proxy endpoints for AI Tools — SUPERSEDED (removed in iframe pivot)

- [x] 1.1 ~~Add `GET /api/ai-tools/assistants`~~ (added, then removed — iframe embeds bk UI directly)
- [x] 1.2 ~~Add `POST /api/ai-tools/assistants/:id/chat`~~ (superseded)
- [x] 1.3 ~~Add `GET /api/ai-tools/rag/collections`~~ (superseded)
- [x] 1.4 ~~Add `POST /api/ai-tools/rag/collections/:name/chat`~~ (superseded)
- [x] 1.5 ~~Register routes~~ (removed from `register_routes.go`; `ai_tools_proxy.go` deleted)
- [x] 1.6 ~~Clear error status when bk unavailable~~ (superseded)
- [x] 1.7 ~~Verify endpoints against running bk~~ (superseded)

## 2. AI Tools workspace drawer (Morph frontend)

- [x] 2.1 Create an AI Tools workspace drawer component using the `.hybrid-drawer` overlay/panel pattern (reference `ChatNotesTodosDrawer.jsx`)
- [x] 2.2 Add an AI Tools toggle control in `SkoolAiChat.js` header with local `useState` open/close (mirror `notesDrawerOpen`)
- [x] 2.3 Embed the full bk UI via iframe so every module is available (Assistants, RAG, Video Stories, Documents, System) — carries Morph session token; "Open in tab" fallback
- [x] 2.4 Header "AI tools" entry opens the in-app drawer; external `localhost:3000` link removed (fixes navigating to a dead/404 bk-ui)
- [x] 2.5 Add error/failure state when the embed cannot load (message + open-in-tab link)
- [x] 2.6 Add drawer + iframe styles in `App.css` (wide panel, full-height frame)

## 3. Assistant & RAG chat windows — SUPERSEDED (removed in iframe pivot)

- [x] 3.1 ~~Reusable native chat-window component~~ (built, then removed — `AiToolsChatWindow.jsx` deleted; embedded bk UI provides Assistants/RAG)
- [x] 3.2 ~~Wire Assistant chat to proxy~~ (superseded)
- [x] 3.3 ~~Wire RAG chat to proxy~~ (superseded)
- [x] 3.4 ~~Follow-up turns / progress~~ (superseded)
- [x] 3.5 ~~Error/empty states~~ (superseded)
- [x] 3.6 ~~Chat windows as entry point~~ (superseded — NOTE: the embedded bk UI keeps bk's existing single-prompt Assistant/RAG dialogs)

## 4. Skills picker in MorphAI chat input

- [x] 4.1 Add a skills button beside the attachment button inside `.input-wrapper.chat-input-inner` in `SkoolAiChat.js`
- [x] 4.2 Build a compact popover above the input listing skills from `GET /api/skills` with multi-select and selected-state indication
- [x] 4.3 Persist selection while composing; close on outside click / re-activation without altering the draft
- [x] 4.4 Include selected `skill_ids` in the `/api/chat` request for both JSON and FormData send paths
- [x] 4.5 Add popover styles in `App.css` (position above input, mobile-safe width)
- [x] 4.6 Verify backend applies skills (selected skills influence the response via `management_chat.go`)

## 5. Remove Images module from bk (keep core generation)

- [x] 5.1 Grep bk for internal callers of `/images/*` vs `generate_image_and_save`; confirm Graphic Documents & Video Stories use the helper, not the HTTP route
- [x] 5.2 Delete `bk/frontend/src/pages/ImageGenerator.js` and remove its `/images` route + nav item (`App.js`, `components/Header.js`)
- [x] 5.3 Remove image-management helpers from `bk/frontend/src/services/api.js` (`generateImage`, `polishImagePrompt`, `getGeneratedImages`, delete image)
- [x] 5.4 Remove `/images/*` management routes from `bk/src/api.py`; keep `bk/src/image_generation.py`. NOTE: retained `GET /images/file/{filename}` (core file-serving) because Graphic Documents reference it for inline image previews
- [x] 5.5 Verify Graphic Documents and Video Stories still generate images end-to-end (static: `generate_image_and_save` import + `/images/file/` route preserved; bk frontend build passes)

## 6. Remove Readers module from bk

- [x] 6.1 Delete `bk/frontend/src/pages/ReadersHub.js`, `ImageReader.js`, `PDFReader.js`
- [x] 6.2 Remove `/readers` route and legacy `/image-reader`, `/pdf-reader` redirects and the nav item (`App.js`, `components/Header.js`)
- [x] 6.3 Remove reader helpers from `bk/frontend/src/services/api.js` (`readImage`, `readMultipleImages`, `readImageAndProcess`, `readImageAndProcessMultiple`, `readPDF`)
- [x] 6.4 Remove `/image-reader/*` and `/pdf-reader/*` routes from `bk/src/api.py` and delete `bk/src/image_reader.py`. DEVIATION: kept `bk/src/pdf_reader.py` — `PDFReader` is still used by ScholarForge project preparation (`stream_prepare_session`), so only the user-facing reader routes were removed
- [x] 6.5 Confirm no remaining imports/references to the removed reader modules

## 7. Validation & cleanup

- [x] 7.1 Build/run Morph frontend and backend; smoke-test the AI Tools drawer, both chat windows, and the skills picker (Morph `go build ./...` passes; frontend `npm run build` passes. Live end-to-end UI smoke test requires all three servers running with a configured bk — deferred to manual QA)
- [x] 7.2 Build/run bk; confirm Images/Readers routes return not found and no dead imports remain (bk app imports cleanly; route enumeration confirms `/images/generate`, `/images/polish-prompt`, `/images/{filename}`, `/image-reader/*`, `/pdf-reader/*` are absent while `/images/file/{filename}` and `/graphic-document/generate` remain)
- [x] 7.3 Run `openspec validate merge-ai-tools-into-morphai --strict` and fix any issues (valid)
- [x] 7.4 Update any docs/nav references that mentioned bk Images/Readers or standalone AI Tools chat (updated `docs/agents/03-morph.md` `externalAPIBase` comment; bk/README has no Images/Readers feature entries)
