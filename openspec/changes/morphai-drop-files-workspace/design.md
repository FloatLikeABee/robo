## Context

See proposal.md — Why. The Morph AI agent shell (`SkoolAiChat` + `AgentWorkspace`) currently has three tabs: Files (`AgentFilesTab` + IndexedDB `morphai-files-workspace`), Notes & TODOs, Context & Knowledge (`HybridContextDrawer` panel). Composer chips: Files (local pins), Notes, Knowledge. Last-session restore also falls back to sessions with a folder binding. Embedded `singleSession` never had this Files tab.

Prior unarchived changes (`morphai-files-session-workspace`, `morphai-restore-folder-workspace-session`, `morphai-files-folder-survives-refresh`, `morphai-remember-files-workspace-open` Files-tab bits, `morphai-pinned-files-readable-context`) specified keeping Files. This change supersedes them for the agent shell.

## Goals / Non-Goals

**Goals:**

- Delete the Files tab and its store, picker, pins, and Files include chip.
- Leave Notes & TODOs and Context & Knowledge as the workspace tabs; default leftover `files` tab keys to Context & Knowledge.
- Keep last-chat restore and workspace pane open/closed; drop folder-binding restore and folder-derived session titles.

**Non-Goals:**

- Removing HybridContext, Knowledge Library, or `include_knowledge` / `include_notes`.
- Dropping chat API fields `include_files` / `pinned_files` (frontend stops sending folder pins; empty payloads are fine).
- Wiping operators’ leftover IndexedDB `morphai-files-workspace` (stop reading it).
- Changing MorphNotes, MorphUtils, Event Logs, Content Maker, Data Access, or Project.

## Decisions

### 1. Delete Files, don’t hide it

**Choice:** Remove `AgentFilesTab`, `filesWorkspaceStore` (and its test), folder/pin state and handlers in `SkoolAiChat`, the Files include chip, and Files-only CSS. `AgentWorkspace` tabs: Notes & TODOs, Context & Knowledge.

**Why:** The user asked it gone; Context & Knowledge already uploads and indexes files. Hiding a tab leaves dead IndexedDB and pin send paths.

**Alternative:** Keep the store “just in case” — contradicts YAGNI and would still ship a second file surface.

### 2. Last session stays; folder binding does not

**Choice:** Keep `morphai-last-session` restore when that `sessionId` still exists, then default. Drop `boundIds` / folder-session fallback and folder-name retitle (`isGenericSessionTitle` only if still used). Do not auto-open the pane on Files.

**Why:** Restoring the last *chat* is useful without a folder workspace. Folder fallback exists only for Files.

**Alternative:** Also drop last-session restore — extra product change the user did not ask for.

### 3. Stale `files` tab key → knowledge

**Choice:** If `morphai-workspace-tab:{sessionId}` (or equivalent) is `files`, treat it as `knowledge`. Default tab when unset: Context & Knowledge.

**Why:** One-line migration; no localStorage wipe.

**Alternative:** Wipe all tab keys — noisy and unnecessary.

### 4. Leave backend pin fields

**Choice:** Frontend sends `include_files: false` (or omits pins). Do not change `AssemblePinnedBlob` / `agent_context.go` unless a compile or test breaks from unused UI-only helpers.

**Why:** Shortest working diff. Empty pinned files already mean “no folder context.”

**Alternative:** Remove `pinned_files` from the API — extra backend churn for no operator-visible gain.

### 5. Lessons

**Choice:** On apply, replace the “Files folder must survive refresh” lesson with: Morph AI has no Files workspace; use Context & Knowledge.

**Why:** That lesson would make the next agent restore a feature the product deleted.

## Risks / Trade-offs

- [Operators who pinned a local folder lose that include path] → Use Context & Knowledge upload / Knowledge Library. No data migration from IndexedDB listings.
- [Unarchived Files OpenSpec changes still describe Files] → This change is the new contract; do not re-apply those Files restore tasks.
- [Backend still accepts `pinned_files`] → Harmless if the UI never fills them.

## Migration Plan

1. Ship frontend deletion together (tab, store, chips, tests).
2. Ignore leftover IndexedDB; optional later `indexedDB.deleteDatabase` is not required.
3. Rollback: restore Files tab from git; IndexedDB recents may already be empty after new sessions.

## Open Questions

None.
