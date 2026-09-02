## Context

See proposal.md — Why. After rebasing `main` onto origin, untracked Morph frontend modules were not on disk while `SkoolAiChat.js` and `AdminDataGrid.js` still imported them. CRA reports `Can't resolve './…'` with no extension. The sources currently exist as untracked `.jsx` files; importers are `.js`. ChatNotesTodosDrawer already uses `.jsx`, so CRA *can* resolve `.jsx`, but the overlay the user hit was a missing file (stash/reset gap). Prefer `.js` so these modules match `SkoolAiChat.js` / `AdminDataGrid.js` and survive the same way other tracked chat files do.

No conflict markers remain in those two importers.

## Goals / Non-Goals

**Goals:**
- Every extensionless import from `SkoolAiChat.js` and `AdminDataGrid.js` maps to a file CRA 5 resolves.
- `AgentWorkspace` keeps `AgentFilesTab` in the same folder.
- `craco build` (or `craco start` overlay) no longer lists those six modules as missing.

**Non-Goals:**
- Re-implementing agent workspace, files IndexedDB, or apply-assistant behavior (already specified in other changes).
- Committing unrelated dirty-tree work (UsersPanel deletions, README trim).
- Changing MorphUtils iframe wiring.

## Decisions

1. **Rename the untracked `.jsx` modules to `.js`**  
   Same basename, same folder as the import path. Alternative: keep `.jsx` — CRA should resolve it, but the reported errors match missing files, and `.js` matches the importers. Alternative: add explicit `.jsx` in imports — more churn in `SkoolAiChat.js`.

2. **Keep module bodies; do not rewrite APIs**  
   Named exports (`workspaceTabStorageKey`, `contextFingerprint`, `APPLIED_ASSISTANT_MSG`, …) stay as they are so `SkoolAiChat.js` does not need import edits.

3. **Leave git conflict markers as a verify step, not a merge**  
   A repo-wide search found none. If webpack still fails after rename, re-check the running CRA process is using this tree (port 3031).

## Risks / Trade-offs

- [Stale webpack overlay] → Restart Morph AI UI after the files land; a process started during the stash gap can keep showing old `Module not found` lines.
- [Only `.jsx` present while webpack looks for `.js`] → Rename removes that ambiguity.
- [Untracked again after another stash] → Implementation should leave the files in `src/` with the rest of Morph frontend so a later commit can include them; this change does not require committing the whole dirty tree.

## Migration Plan

1. Rename the six frontend modules (plus `AgentFilesTab`) to `.js`.
2. Restart or rebuild Morph AI frontend and confirm the overlay is gone.
3. Rollback: restore the `.jsx` names; importers stay extensionless either way.
