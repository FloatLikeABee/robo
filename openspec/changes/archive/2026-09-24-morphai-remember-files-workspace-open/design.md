## Context

See proposal.md — Why. Today `AgentWorkspace` writes the active tab to `sessionStorage` via `workspaceTabStorageKey(sessionId)`. `SkoolAiChat.js` reads it on session change. Folder bindings and handles already live in IndexedDB (`filesWorkspaceStore.js`). The right workspace column is always rendered on the agent shell grid; there is no show/hide control yet.

## Goals / Non-Goals

**Goals:**

- `localStorage` persistence for workspace pane open/closed (global per origin).
- `localStorage` persistence for active workspace tab (per `sessionId`).
- Restore open state + tab on mount; when a session has a folder binding, land on Files with `loadWorkspaceForSession` output visible.
- Minimal toggle control (header icon or workspace edge chevron) gated to `isAgentShell`.

**Non-Goals:**

- Changing IndexedDB schema or folder/recents logic (reuse `filesWorkspaceStore`).
- Persisting include-bar chips (Files / Notes / Knowledge toggles above composer).
- Embedded / `singleSession` workspace.
- Server-side preferences or multi-device sync.

## Decisions

### 1. `localStorage` for cross-visit UI prefs

**Choice:** Keys `morphai-workspace-open` (`'1'` / `'0'`) and reuse `morphai-workspace-tab:{sessionId}` but read/write via `localStorage` instead of `sessionStorage`. On write, also mirror to `sessionStorage` once for same-tab backward compatibility during rollout (optional one-release shim; can skip if YAGNI).

**Why:** User asked for "next time using our product" — that means survive browser close. Tab key already exists; only the storage API changes.

**Alternative:** Keep `sessionStorage` — fails the cross-visit requirement.

### 2. Default open when unset

**Choice:** If `morphai-workspace-open` is missing, default **open** so existing users keep today's always-visible workspace. First explicit collapse sets `'0'`.

**Why:** Non-breaking; only users who hide the pane get a collapsed default afterward.

### 3. Collapsed layout via grid class

**Choice:** Add `app--workspace-collapsed` on the agent shell when closed. CSS sets workspace column to `0` / `display: none` and lets chat span both chat+workspace grid areas on desktop; stacked mobile layout hides the bottom workspace row the same way.

**Why:** Matches session-rail collapse pattern; no portal/overlay.

**Alternative:** Keep DOM mounted off-screen — wastes layout on narrow screens.

### 4. Toggle placement

**Choice:** Header icon in `chat-header` (agent shell only), `aria-expanded` tied to open state, label "Workspace" / "Files workspace". Click toggles open and writes `localStorage`.

**Why:** Discoverable, same band as other chrome; avoids fighting tab strip.

**Alternative:** Chevrons on pane edge — fine but more CSS; header is fewer lines.

### 5. Files restore on load

**Choice:** After `loadWorkspaceForSession` resolves with a `folderId`, if saved tab is missing or workspace was never used, set tab to `files`. Do not override a valid saved tab of `notes` or `knowledge` when there is no folder binding.

**Why:** Folder + Files tab go together; other tabs stay user-chosen.

### 6. Tests

**Choice:** One small unit test (or assert-based helper test) for storage read/write helpers if extracted; manual smoke checklist in tasks. No new test framework.

## Risks / Trade-offs

- [Stale closed state on first visit after deploy] → Default open when key absent.
- [localStorage blocked / private mode] → try/catch; fall back to in-memory default (open).
- [Tab restored before session list loads] → Restore tab in the same `sessionId` effect that already loads messages/workspace.

## Migration Plan

1. Ship storage + restore + toggle; CSS collapsed class.
2. Rollback: remove toggle and class; revert to always-open grid and `sessionStorage` tab only.

## Open Questions

None.
