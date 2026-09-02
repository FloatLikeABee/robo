## Context

See `proposal.md` (Why) for motivation. Relevant current state:

- **AI Tools (`bk/`)** is a React CRA + MUI app on a FastAPI monolith (`bk/src/api.py`). Assistants run via `POST /assistants/{id}/run` (single query → single result); RAG search via `POST /rag/collections/{name}/query`. There is no server-side multi-turn assistant session; the run endpoint is stateless.
- **MorphAI** (`morph/frontend`) is React CRA. Its chat is `SkoolAiChat.js` with a custom-CSS right drawer pattern (`.hybrid-drawer` in `App.css`) already used by `ChatNotesTodosDrawer.jsx` and `HybridContextDrawer.js`. The attachment button lives inside `.input-wrapper.chat-input-inner`.
- **Morph backend** already proxies to `bk`: `handlers/bk_assistants.go` exposes `bkAPIBase()`, `bkHTTPGetJSON/PostJSON`, `listBKAssistants`, `getBKAssistant`, `queryBKRAGCollection`; `handlers/morph_ai_agents.go` exposes `GET /api/ai-agents`. Skills are fully wired server-side: `ChatRequest.SkillIDs` → `management_chat.go` → `buildEnabledSkillsContext`; frontend never sends them. `GET /api/skills` exists.

Constraints: reuse existing Morph auth (`morphSession.js` bearer), the existing drawer CSS pattern, and the existing `bk` proxy helpers. No browser-to-`bk` calls (browser talks only to Morph, which proxies to `bk`).

## Goals / Non-Goals

**Goals:**
- Add one MorphAI right-drawer that hosts an AI Tools workspace limited to Assistants and RAG.
- Provide real multi-turn chat windows (right-drawer modals) for Assistant chat and RAG chat, reusing MorphAI chat styling/UX.
- Add a skills picker beside the attachment button that sends `skill_ids` with `/api/chat`.
- Remove Images and Readers from `bk` (routes, pages, endpoints, readers backend) without breaking Graphic Documents / Video Stories image generation.

**Non-Goals:**
- No changes to Video Stories, Documents, or System modules beyond preserving image generation.
- No server-side persistence of AI Tools assistant/RAG conversations (multi-turn is maintained client-side and replayed to the stateless `bk` endpoints).
- No redesign of the existing MorphAI `/api/chat` session model.
- No removal of `bk` backend managers for assistants/RAG (still used via proxy).

## Decisions

### D1: Native rebuild in Morph frontend, not iframe
Rebuild the Assistants/RAG UI as native Morph React components consuming Morph proxy endpoints. Rationale: consistent styling with `.hybrid-drawer` and MorphAI chat, shared auth, and the ability to turn single-prompt flows into multi-turn windows. Alternative (iframe embed of `bk` UI) rejected: keeps MUI styling and separate auth, and cannot deliver the "chat window like MorphAI" requirement cleanly.

### D2: Reuse the `.hybrid-drawer` pattern; nest chat windows as right-drawer modals
The AI Tools workspace opens as a `.hybrid-drawer` (like Notes & TODOs). Selecting an assistant or collection opens a chat window, also a right-drawer modal. Rationale: matches the requested UX ("very big modal from right drawer" + "chat window should be right drawer modal") and reuses existing CSS/animation. Alternative (MUI `Drawer`) rejected for the chat surface to stay visually consistent with MorphAI.

### D3: Client-maintained multi-turn over stateless `bk` endpoints
`bk`'s assistant run and RAG query are stateless. The chat window keeps message history in component state and, on each turn, sends the new user message (optionally with prior-turn context as needed) to the proxy. Rationale: avoids new `bk` session storage; keeps scope contained. Trade-off: history is not persisted across reloads (acceptable; see Risks). Alternative (add `bk` conversation storage) rejected as out of scope.

### D4: New Morph proxy endpoints built on existing helpers
Add Morph endpoints so the browser never calls `bk` directly:
- `GET /api/ai-tools/assistants` — list (wraps `listBKAssistants`).
- `POST /api/ai-tools/assistants/{id}/chat` — run a turn (wraps the `bk` assistant run; body `{ message }`, returns `{ response }`).
- `GET /api/ai-tools/rag/collections` — list collections.
- `POST /api/ai-tools/rag/collections/{name}/chat` — run a RAG query turn (wraps `queryBKRAGCollection`; body `{ query }`, returns results).
Rationale: reuse `bkHTTPGetJSON/PostJSON`, keep auth/session on the Morph side, and isolate `bk` URL/config. Registered in `handlers/register_routes.go`. Alternative (reuse `/api/ai-agents` + `/api/chat` with `agent_id=bk:*`) rejected because that routes through Morph's management-tool chat loop; a thin dedicated proxy is simpler and keeps AI Tools chat independent from Morph session chat.

### D5: Skills picker sends `skill_ids` on the existing `/api/chat`
Add a skills button inside `.input-wrapper.chat-input-inner` (next to attach) opening a compact popover above the input, populated from `GET /api/skills`. Selected ids are added to the `/api/chat` request body (`skill_ids`) for both JSON and FormData paths. Rationale: backend already consumes `ChatRequest.SkillIDs`; this is the missing UI wiring. Alternative (reuse centered `SkillsModal`) rejected — requirement is a small popover above the input, not a fullscreen modal (the existing header Skills modal can remain for catalog management).

### D6: Images removal keeps core generation
Delete only the user-facing Images surface: `bk/frontend` page `ImageGenerator.js`, its route/nav, and `services/api.js` image helpers; and the `/images/*` routes in `bk/src/api.py`. Keep `bk/src/image_generation.py` because `graphic_document_service.py` and the video story pipeline import `generate_image_and_save`. Rationale: satisfies "remove images module" without breaking dependents. Alternative (full removal) explicitly rejected by scoping decision.

### D7: Readers removal is complete
Delete `bk/frontend` `ReadersHub.js`, `ImageReader.js`, `PDFReader.js`; the `/readers` route and legacy redirects (`/image-reader`, `/pdf-reader`); reader helpers in `services/api.js`; and backend `image_reader.py`, `pdf_reader.py` plus the `/image-reader/*` and `/pdf-reader/*` routes in `api.py`. Rationale: no other module depends on the readers.

## Risks / Trade-offs

- **Client-only history loses conversations on reload** → Acceptable for v1; document it. Future work can add persistence if needed.
- **Removing `/images/*` may break external callers or the graphic-document/video paths if they call the HTTP route instead of the helper** → Before deleting routes, grep for internal usages; confirm dependents call `generate_image_and_save` directly, not `/images/*`.
- **`bk` backend availability** → Proxy endpoints must surface a clear error state; the workspace shows an error rather than hanging (reuse the timeout pattern in `bk_assistants.go`).
- **Skills popover overlap with input on small screens** → Position above input with mobile-safe width, mirroring existing popover/drawer responsive rules in `App.css`.
- **Stateless assistant run ignores prior turns** → If the `bk` run endpoint cannot accept history, follow-ups may lack context; mitigate by sending a compact transcript in the message payload where the proxy/`bk` supports it, otherwise treat each turn independently and note the limitation.

## Migration Plan

1. Land Morph proxy endpoints (D4) behind existing auth; verify against a running `bk`.
2. Add the AI Tools drawer + Assistants/RAG chat windows (D1–D3) in Morph frontend.
3. Add the skills picker (D5) and wire `skill_ids`.
4. Remove Images user surface (D6) and Readers module (D7) from `bk`; verify Graphic Documents and Video Stories still generate images.
5. Rollback: revert frontend drawer/skills commits and the `bk` deletions independently; proxy endpoints are additive and safe to leave.

## Open Questions

- Does the `bk` assistant run endpoint accept any conversation/history parameter, or must multi-turn context be embedded in the message text? (Deferrable: affects only how D3/D5 payloads are shaped, not the specs or task breakdown.)
