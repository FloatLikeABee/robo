## Context

See proposal.md — Why. Related shipped work (`morphai-files-folder-survives-refresh`) persists folder **listing** in IndexedDB via `toFileSnapshot()` but strips handles. `readPinnedFileBodies` (`agentContext.js`) only reads `handle.getFile()` or `file.text()`. When `loadWorkspaceForSession` returns snapshot rows (permission not granted, webkit-only open, or post-refresh), pins appear in UI but send yields `content: ""`. Backend `AssemblePinnedBlob` skips empty bodies; the model correctly reports missing files.

**Files (N)** chip only toggles `includeFiles` — it does not validate readability.

## Goals / Non-Goals

**Goals:**

- Pinned files always send text when Files include is on and content is available.
- Cache pin text at pin-time; persist small bodies in session IDB.
- Block send + show reconnect/warning when pins are on but unreadable.
- Minimal diff: extend existing `filesWorkspaceStore` binding, no Go API change.

**Non-Goals:**

- Server-side file fetch from user disk.
- Pinning files larger than existing 512 KiB cap.
- Changing sub-agent routing or context-cache fingerprint algorithm (only ensure bodies are non-empty when pins claim readability).

## Decisions

### 1. Pin-time eager read + session cache

**Choice:** On `handleTogglePin` (add pin), call shared `readFileEntryText(f)`; store result in `setSessionBinding` as `pinnedContents: { [path]: { hash, content } }` (only when non-empty and under cap). `readPinnedFileBodies` checks cache first, then live read.

**Why:** Matches operator mental model (“pin = include this file”); survives refresh without handles.

**Alternative:** Force reconnect before every send — worse UX; user already sees listing.

### 2. Send guard

**Choice:** Before `tranApi.post('/api/chat')`, if `includeFiles && pinnedPaths.size > 0 && pinnedBodies.every(empty)`, set error state and return early with message linking to Reconnect.

**Why:** Fail loud per ponytail; avoids wasting tokens on hollow context.

### 3. UI warnings

**Choice:** `AgentFilesTab` — pinned row with no cache and no live handle shows “needs reconnect” badge. Files chip gets `is-warning` class when any pin unreadable.

**Why:** Explains why chat failed before send.

### 4. Cap and eviction

**Choice:** Reuse `MAX_FILE_BYTES` (512 KiB); max ~10 pinned texts stored (same order as pin set); drop oldest on overflow. Unpin removes entry.

**Why:** IDB size bound; aligns with skipped-file rules.

### 5. Reconnect auto-refresh cache

**Choice:** After successful `reconnectRecentFolder`, re-read all pinned paths into `pinnedContents`.

**Why:** Restores live reads without re-pinning.

## Risks / Trade-offs

- [Stale pin text if file edited on disk] → Accept for v1; hash mismatch could invalidate cache later (optional follow-up).
- [IDB size] → Cap pin count + file size; only store text for pinned paths.
- [Privacy] → Text stays local in IDB like listing snapshot already does.

## Migration Plan

1. Extend session binding schema (additive `pinnedContents` field).
2. Update pin toggle + `readPinnedFileBodies`.
3. Send guard + UI warnings.
4. Unit tests in `agentContext.test.js` + `filesWorkspaceStore.test.js`.

No server deploy dependency; frontend-only.

## Open Questions

None for v1.
