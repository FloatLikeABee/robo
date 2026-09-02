## Context

See proposal.md for motivation. Specs: `morphai-agent-shell`, `morphai-agent-workspace`, `morphai-agent-orchestration`.

Today Morph AI (`SkoolAiChat.js`) is a full-width chat: session sidebar (220px, overlay on small screens), message list, single-line composer (`rows={1}`, `min-height: 24px`, `max-height: 120px`). Notes & TODOs and HybridContext/Knowledge are **overlay drawers**. Chat already sends `session_id`, `skill_ids`, `agent_id`, attachments, and HybridContext attach/detach. Backend `POST /api/chat` plus `/api/chat/hybrid-context*` and `/api/knowledge/*` are the durable stores. Embedded Morph Data uses `singleSession` and MUST stay compact.

## Goals / Non-Goals

**Goals:**

- In-flow agent layout: collapsible sessions, ~50/50 chat | workspace, composer only in the chat column at ~3× height.
- Reuse existing notes and HybridContext/knowledge implementations inside tabs instead of duplicating data models.
- Session-scoped pin set + server cache keyed by user + session so repeat turns skip full re-upload.
- Task fan-out using existing Morph internal tools / HybridContext / knowledge search, with visible sub-agent status.

**Non-Goals:**

- Writing or deleting files in the user’s local folder.
- Replacing Morph Utils navigation or other apps’ assistants.
- A general coding agent (apply patches, run terminals) in this change.
- New streaming protocol unless the existing progress helper is insufficient; prefer extending `aiProgress` / current chat response.

## Decisions

### 1. Layout: CSS grid, sessions collapse in document flow
- **Choice:** Main Morph AI uses a three-region grid: `sessions | chat | workspace`. Collapsed sessions width is a slim rail (~48px, toggle + new-chat) or 0px with the header ☰ remaining. Expanded width stays ~220px. Persist collapse in `sessionStorage`. Below a breakpoint (~900px), workspace stacks under chat; sessions may keep today’s overlay pattern.
- **Rationale:** Overlay sidebars hide the chat; in-flow collapse is what “more space” means. Spec allows stack on narrow viewports.
- **Alternatives:** Keep overlay-only sessions — fails the collapse-for-width requirement. Three independently resizable panes — nicer later, more CSS/js for v1.

### 2. Composer stays in the left column; height ×3
- **Choice:** Chat column is a flex column: header (shared or chat-only), messages (`flex: 1`), composer. Workspace has its own header (tabs). Composer `min-height` ≈ 72px (3× 24px) and `max-height` ≈ 360px; start at 3 visible rows.
- **Rationale:** Spec requires input left of workspace and taller paste area.
- **Alternatives:** Composer full-width under both panes — rejected by spec.

### 3. Files tab: browser directory picker, read-only
- **Choice:** Prefer File System Access API (`showDirectoryPicker`) with IndexedDB-stored handles per `session_id`. Fallback: `<input webkitdirectory>`. List a tree; skip obvious junk (`node_modules`, `.git`) and files over a size cap; pin/unpin in session state. Read bytes in the browser and send to Morph for cache + prompt; never `createWritable` on the user’s folder.
- **Rationale:** Users asked to “open a folder”; browsers cannot silently scan the disk. Chrome/Edge get a real folder; Safari gets the fallback.
- **Alternatives:** Only Morph Knowledge uploads — weaker “folder as workspace”. A desktop sidecar — out of scope.

### 4. Tabs wrap existing drawers; header icons retarget tabs
- **Choice:** Extract the bodies of `ChatNotesTodosDrawer` (`NotesTodosContent`) and `HybridContextDrawer` into the Notes and Context & Knowledge tabs. Remove overlay as the primary path on main Morph AI. Header note/knowledge buttons set `workspaceTab` and ensure the workspace is visible.
- **Rationale:** Same data, no second notes stack. Spec: shortcuts open the tab, not an overlay.
- **Alternatives:** Keep overlays plus tabs — two UIs for one store.

### 5. Context controls as an include bar on the chat column
- **Choice:** Compact chips above the composer: Files (pin count), Notes, Knowledge/Hybrid. Each chip toggles include for the next `POST /api/chat`. Pin list lives in the Files tab; chips only toggle “send this category.”
- **Rationale:** Controls sit with the send action so the user sees what the agent will use.
- **Alternatives:** Controls only inside the workspace — easy to miss when sending.

### 6. Context cache: hash of include set, server-side per user/session
- **Choice:** Client computes hashes of pinned file contents. Request sends `context_cache_key` + hashes; if the server has a matching blob for that user/session, skip bodies. Store in SQLite (e.g. `morph_agent_context_cache`) with include-set fingerprint. Invalidate when pins, notes snapshot, or knowledge/hybrid sources change. Do not put cache in a shared global table without `user_id`.
- **Rationale:** Avoids re-uploading large folders every turn; matches the cache requirement without depending on a provider-specific prompt-cache API (optional extra later).
- **Alternatives:** Provider prompt caching only — not portable across DashScope/SiliconFlow. Client-only cache — lost on refresh and never seen by the API.

### 7. Sub-agents: orchestrate existing tools, not new processes
- **Choice:** On each send, a cheap router (rules + one model call) decides `{files, morphdata, knowledge, notes, web}` workers. Run allowed workers (existing internal API / graph search / HybridContext extract) concurrently with a small cap (e.g. 3). Surface labels through the existing AI progress ticker. Parent model writes one markdown reply from `TOOL_RESULT`s. No extra OS processes.
- **Rationale:** Morph already has tool-shaped internal APIs; fan-out is composition, not a new runtime. Simple questions skip the router’s extra workers.
- **Alternatives:** Always one model — fails multi-sub-agent. True multi-process agents — operationally heavy for v1.

## Risks / Trade-offs

- [Directory picker unsupported] → webkitdirectory fallback + empty state; chat still works.
- [Huge folders] → skip lists, size cap, pin-only inclusion (do not dump the whole tree into the prompt).
- [Cache serving stale files] → fingerprint includes content hashes; unpin/re-pin or file change busts cache.
- [Sub-agent cost/latency] → cap concurrency; skip fan-out for short/no-tool prompts.
- [Embedded chat regressions] → gate layout on `!singleSession && !embedded` (or equivalent); leave Morph Data embed as today.

## Migration Plan

1. Ship shell layout + taller composer behind the main Morph AI route only.
2. Move drawer content into tabs; point header buttons at tabs; delete overlay as default.
3. Add pin/include chips and cache API; then router + progress for sub-agents.
4. Rollback: revert frontend layout and ignore new request fields; leftover cache table is unused.

## Open Questions

None that change specs or the task breakdown. Provider-side prompt cache can be added later without changing user-visible include/exclude behavior.
