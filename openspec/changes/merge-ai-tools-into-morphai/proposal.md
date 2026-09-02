## Why

The AI Tools app (`bk/`) is a separate React/FastAPI product whose most valuable features — Assistants and RAG — are hidden behind single-prompt modal dialogs (one query in, one result out) with no conversation. Meanwhile MorphAI already offers a polished multi-turn chat with sessions, attachments, right-drawer panels, and a skills backend that the UI never uses. Consolidating AI Tools into MorphAI gives users one coherent chat experience, removes dead weight (Images, Readers), and finally exposes the existing skills capability in the chat.

## What Changes

- **Remove the Images module from `bk`** — delete the user-facing Images page/route (`/images`) and its `/images/*` endpoints. **Keep** the core image-generation helper that Graphic Documents and Video Stories depend on. **BREAKING**: `/images` route and `/images/*` HTTP endpoints are removed.
- **Remove the Readers module from `bk`** — delete Image Reader and PDF Reader pages, the `/readers` route (plus legacy `/image-reader`, `/pdf-reader` redirects), and their `/image-reader/*` and `/pdf-reader/*` endpoints and backend readers. **BREAKING**: those routes/endpoints are removed.
- **Merge AI Tools into MorphAI** — add a large right-drawer panel in MorphAI (same `.hybrid-drawer` pattern as Notes & TODOs) that hosts the AI Tools workspace: Assistants and RAG only. Rebuilt natively in Morph's React frontend, talking to the `bk` backend through Morph proxy endpoints (extending the existing `bk_assistants.go` / `/api/ai-agents` pattern).
- **Turn Assistant chat and RAG chat into real chat windows** — replace the single-prompt `Dialog`s with multi-turn chat windows presented as right-drawer modals, styled like MorphAI chat (message history, streaming/progress status, input bar).
- **Add a skills picker to MorphAI chat** — a small button beside the attachment button opens a compact popover above the input to select one or more skills; selected `skill_ids` are sent with the chat request (backend already supports this).

## Capabilities

### New Capabilities
- `morphai-ai-tools-workspace`: AI Tools (Assistants + RAG) surfaced inside MorphAI as a large right-drawer, with multi-turn chat windows (right-drawer modals) for Assistant chat and RAG chat, backed by Morph proxy endpoints to the `bk` backend.
- `morphai-chat-skill-selection`: A skills picker in the MorphAI chat input that lets users choose skills from a compact popover and include them in the chat request.
- `ai-tools-module-cleanup`: Removal of the Images and Readers modules (routes, pages, and endpoints) from the AI Tools (`bk`) app, while preserving core image generation used by other modules.

### Modified Capabilities
<!-- No existing specs in openspec/specs/; all behavior here is introduced as new capabilities. -->

## Impact

- **`bk` frontend** (`bk/frontend/src`): delete `pages/ImageGenerator.js`, `pages/ReadersHub.js`, `pages/ImageReader.js`, `pages/PDFReader.js`; remove routes/redirects and nav items in `App.js` and `components/Header.js`; prune image/reader helpers from `services/api.js`.
- **`bk` backend** (`bk/src`): remove `image_reader.py`, `pdf_reader.py`, and the `/images/*`, `/image-reader/*`, `/pdf-reader/*` routes in `api.py`. Keep `image_generation.py` (used by `graphic_document_service.py` and video story pipeline) but drop only the user-facing image endpoints.
- **Morph frontend** (`morph/frontend/src`): new AI Tools right-drawer + Assistants/RAG chat window components (reusing `SkoolAiChat`/chat styling and the `.hybrid-drawer` pattern); add a skills button in `SkoolAiChat.js` input area with a popover; wire `skill_ids` into the `/api/chat` request.
- **Morph backend** (`morph/handlers`): extend `bk_assistants.go` / `morph_ai_agents.go` with proxy endpoints for multi-turn assistant chat and RAG query used by the new drawer.
- **Dependencies/systems**: existing `bk` assistant/RAG endpoints consumed via Morph proxy; existing Morph skills backend (`handlers/skills.go`, `ChatRequest.SkillIDs`) now driven from the UI.
