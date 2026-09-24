## Why

Operators pin workspace files and turn on **Files (N)** expecting Morph AI to read them. Today the UI can show a pin and the include chip while the chat receives **no file text**: `readPinnedFileBodies` only reads live `FileSystemHandle` / `File` objects, but after refresh (or when folder permission is not granted) `workspaceFiles` are IDB snapshots (`path`, `size`, `skipped` only). Pinning still works; send silently ships empty `content` → backend `AssemblePinnedBlob` skips it → the model honestly says it cannot see the file. This breaks the core Files→chat workflow.

## What Changes

- **Eager pin cache:** When a file is pinned, read its text immediately (if readable) and store on the session binding (`pinnedContents` map or array keyed by path, capped like file snapshot).
- **Send path:** `readPinnedFileBodies` uses cached pin text first, then live handle/`File` as fallback.
- **Guardrails:** If pins are on but zero pinned bodies are readable at send time, block the send with a clear inline error (“Reconnect folder to read pinned files”) instead of silently omitting context.
- **UI feedback:** Pinned rows that lack readable content show a warning; **Files (N)** chip shows a warning dot when any pin is unreadable; reconnect CTA when handle permission is needed.
- **Tests:** Unit tests for snapshot-only workspace + cached pin content; handler test unchanged (already skips empty content).

## Capabilities

### New Capabilities

- `morphai-pinned-files-context`: Pinned workspace files MUST contribute readable text to `POST /api/chat` when Files include is on; unreadable pins MUST surface explicit operator feedback instead of silent empty context.

### Modified Capabilities

- `morphai-files-folder-persist`: Extend session snapshot to optionally store pinned file text (small files only) so refresh + pin survives for chat context, not just listing display.

## Impact

- `morph/frontend/src/lib/agentContext.js` — `readPinnedFileBodies` + pin-read helpers
- `morph/frontend/src/lib/filesWorkspaceStore.js` — session binding shape, pin cache on toggle
- `morph/frontend/src/SkoolAiChat.js` — send guard, error state, reconnect prompt
- `morph/frontend/src/components/chat/AgentFilesTab.js` — unreadable pin affordance
- `openspec/changes/morphai-files-folder-survives-refresh/specs/...` delta via modified capability above
- No Go API contract change (still `pinned_files[]` with `path`, `hash`, `content`)
