# Design

## Context

See proposal.md — Why. On current `main`, `morph/frontend/src/components/chat/AgentFilesTab.js` and `morph/frontend/src/lib/filesWorkspaceStore.js` are gone. `AgentWorkspace.js` tabs are Notes & TODOs and Context & Knowledge, and `readWorkspaceTab` maps a stored `files` value to `knowledge`. `SkoolAiChat.js` sends `include_files: false` and `pinned_files: []`.

`openspec/changes/archive/` already uses `2026-09-24-<change-name>/` (for example `2026-09-24-morph-mcp-stdio-skeleton`). `openspec archive` merges delta specs into `openspec/specs/` unless `--skip-specs` is set. Live specs today are only `morph-data-api-auth` and `morph-mcp-stdio`. There is no `openspec/README.md` or `openspec/project.md`. `openspec/config.yaml` `context` is what the CLI injects when creating artifacts.

`morphai-drop-files-workspace` (still active, all tasks checked) names the prior changes it supersedes for the agent shell.

## Goals / Non-Goals

**Goals:**

- Move the five Files-workspace changes into the dated archive, with a superseded marker, without deleting their files and without merging their deltas.
- Leave `morphai-drop-files-workspace` active and unsynced.
- Put the product-owner warning where OpenSpec agents read it (`openspec/project.md` and the CLI project context).
- Mark the two active changes that still say to build the Files tab, without archiving the behavior they still describe.

**Non-Goals:**

- Archiving the other complete changes.
- Editing `docs/agents/*` or `README.md`.
- Changing Morph AI code, lessons, or tests.
- Syncing `morphai-drop-files-workspace` into `openspec/specs/` in this change.

## Decisions

### 1. Archive these five, and only these five, as Files-workspace changes

| Change | Why it qualifies |
| --- | --- |
| `morphai-files-session-workspace` | Per-session IndexedDB folder, pins, and recents on the Files tab. Tasks 5/5. `AgentFilesTab` is gone. |
| `morphai-restore-folder-workspace-session` | Reload opens the folder-bound session on Files and retitles from the folder. Tasks 13/13. Drop-files removes `boundIds` and folder titles. Last-chat restore without a folder is restated by drop-files and still implemented by `resolveRestoredSessionId`. |
| `morphai-files-folder-survives-refresh` | Persist folder name and listing in IndexedDB. Tasks 9/9. The store module is deleted. |
| `morphai-remember-files-workspace-open` | Return visit reopens the Files tab and folder binding. Tasks 10/10. Drop-files calls out its Files-tab bits. Pane open/closed and Notes/Knowledge tab persistence remain in `AgentWorkspace.js`; a stored `files` tab becomes `knowledge`. |
| `morphai-pinned-files-readable-context` | Cache pinned local-folder text and send it. Tasks 12/12. The UI no longer sends folder pins. Its delta also modifies `morphai-files-folder-persist`, which must not become a live spec. |

Each proposal is checkbox-complete, specifies the local-folder / pins / IndexedDB Files tab, and contradicts current code and `morphai-drop-files-workspace`.

**Rejected:** Treat “about six” as a quota and archive a sixth folder to match the issue’s estimate. The reversing design names these five. A sixth archive would be a mixed change (below).

### 2. Do not archive `morphai-agent-workspace`

That change introduced the Files tab, and its spec still says the Files tab MUST open a local folder. It also specifies the collapsible session rail, the split shell, Notes & TODOs, Context & Knowledge, and sub-agent orchestration. Those are still in the product (`AgentWorkspace.js`, `inferSubAgents`). Archiving the folder with a blanket “superseded by drop-files” banner would tell the next agent the shell was removed.

**Choice:** Leave the folder active. Add a superseded banner on its proposal: the Files tab / local-folder / pin-from-folder instructions are superseded by `morphai-drop-files-workspace` and are not permission to restore the tab. Do not delete the old requirement text.

**Rejected:** Archive it anyway. The still-current shell would sit in `archive/` next to removed Files work and look equally obsolete.

### 3. Do not archive `morphai-restore-missing-webpack-modules`

The change exists so webpack can resolve `AgentWorkspace`, `agentContext`, `AiToolsWorkspaceDrawer`, `appliedAssistantChannel`, and `ExtractJsonFromTextDialog`. Those modules are still imported. One scenario still says the Files tab folder picker must load, and the proposal says to keep `AgentFilesTab`.

**Choice:** Leave the folder active. Banner the proposal: do not restore `AgentFilesTab` or `filesWorkspaceStore`; that part is superseded by `morphai-drop-files-workspace`. Keep the rest of the module list.

**Rejected:** Archive it as a Files change. The next agent could treat `AgentWorkspace.js` as part of the removed tab and delete it.

### 4. Do not edit `platform-trim-readme-chat-skills`

Its chrome spec used to say “Files-workspace tabs MUST still exist.” That sentence is reworded to Notes & TODOs and Context & Knowledge, and the proposal banner says the IndexedDB Files tab is superseded by `morphai-drop-files-workspace`. The rest of that change stays active.

**Rejected:** Patch the chrome requirement in place. It is a real stale MUST, but this change does not take ownership of that chrome change.

### 5. `morphai-drop-files-workspace` stays active and is not synced

**Choice:** Do not move it and do not merge `morphai-no-files-workspace` into `openspec/specs/`. The issue requires that change to remain the source of truth. It is checkbox-complete and matches the code, but it is the policy record agents should open by name. Parking it in `archive/` beside the five superseded folders makes them look like the same kind of history.

**Rejected:** Archive it and sync its delta, so `openspec/specs/morphai-no-files-workspace/spec.md` becomes the policy. That matches how `morph-mcp-stdio` was archived, and a later change may do it. Doing it here puts the policy change in the same archive directory as the changes it reversed, and creates a second copy of the policy in this PR. A later change can archive it by modifying the “policy is this active change” requirement and syncing only the drop-files delta.

**Rejected:** Archive it with `--skip-specs`. Then neither active changes nor live specs state the policy; only an archived folder does, next to the old Files recipes.

### 6. Archive with `openspec archive --skip-specs`

Default `openspec archive` merges every delta into `openspec/specs/`. For these five, that would add live requirements for the Files tab, session folder binding, pin text, and “restore Files on return.” The CLI flag `--skip-specs` skips that merge and still moves the folder to `openspec/changes/archive/YYYY-MM-DD-<name>/`, which is the same layout as the archives already on `main`.

Add the superseded banner to each proposal **before** the move, so the archived history carries it. Do not delete tasks or specs inside the folder.

Banner (all five):

> **Superseded** by `morphai-drop-files-workspace`. This change specified Morph AI’s IndexedDB Files workspace (local folder, pins, Files tab). That behavior was removed. Do not re-implement it, and do not sync these delta specs into `openspec/specs/`.

Extra sentence on `morphai-remember-files-workspace-open`: pane open/closed and Notes or Knowledge tab persistence remain current and are restated by `morphai-drop-files-workspace`. A stored `files` tab means Context & Knowledge. That is not a reason to restore the Files tab.

Extra sentence on `morphai-restore-folder-workspace-session`: last-chat restore without a folder binding remains current and is restated by `morphai-drop-files-workspace`. Folder-bound restore and folder-derived titles do not.

**Rejected:** Move the directories by hand with `git mv` only. The CLI is what enforces the date prefix and the “do not double-prefix” rule. Use the CLI, then confirm `openspec/specs/` gained nothing from these five.

**Rejected:** Sync the deltas, then archive. That publishes the removed tab as live behavior.

### 7. Warning lives in OpenSpec docs, in two places that say the same thing

`openspec/project.md` is the doc the issue asked for (there is no OpenSpec README today). Repeat one short paragraph in `openspec/config.yaml` `context`, and point it at `openspec/project.md`, because that context is what `openspec instructions` shows to agents. Do not add the note to `docs/agents/*` or `README.md` (docs catch-up is a separate change).

The warning says: do not restore the IndexedDB Files tab (local folder picker, recents, pins, `AgentFilesTab`, `filesWorkspaceStore`, database `morphai-files-workspace`) unless the product owner explicitly says yes. Current policy is the active change `morphai-drop-files-workspace`. The five archived changes are history. Active mentions in `morphai-agent-workspace`, `morphai-restore-missing-webpack-modules`, and `platform-trim-readme-chat-skills` are not permission to bring the tab back.

### 8. Prove the archive with a check that fails first

Add `openspec/check_files_workspace_archive.py`. It fails if any of the five is still active, if an archived proposal lacks the superseded marker, if `openspec/specs/` contains a Files-tab requirement or `morphai-no-files-workspace`, if `morphai-drop-files-workspace` is not active, if the two mixed changes lost their disclaimers or were archived, or if `openspec/project.md` lacks the product-owner warning. Run it before the moves (red), then after (green). No Go, frontend, or Rust code changes, so those suites are not the proof for this change.

## Risks / Trade-offs

- [Default archive merges the Files deltas] → `--skip-specs`, then diff `openspec/specs/` and confirm the five capability names are absent.
- [`morphai-remember-files-workspace-open` also specified pane persistence that is still live] → Extra banner sentence. Drop-files and `AgentWorkspace.js` remain the description of that kept behavior. The archived delta is not synced, so “reopen Files” does not become a live spec.
- [`morphai-agent-workspace` still contains a Files MUST after the banner] → The banner is the first thing in the proposal, and the project warning names the change. The requirement text stays so history is not deleted.
- [`platform-trim-readme-chat-skills` still says Files-workspace tabs must exist] → Named in the warning. Not edited here.
- [`openspec/project.md` is not auto-loaded] → Same warning in `openspec/config.yaml` context.
- [A later archive of drop-files would contradict “it stays active” once this spec is synced] → That later change must modify this requirement and sync only the drop-files delta. This change does not do that.
- [Date-prefix collision] → None of these names exist under `archive/` yet. Today’s date is 2026-09-24, matching the existing archive folders.

## Migration Plan

1. Add the check and watch it fail.
2. Banner the five proposals, then `openspec archive <name> --skip-specs --yes` for each.
3. Banner the two mixed proposals. Add `openspec/project.md` and the config context line.
4. Re-run the check. Confirm `openspec validate` and `openspec list`.
5. Rollback is reverting this change. No data migration. IndexedDB leftovers stay ignored, as drop-files already decided.

## Open Questions

None.

## Design review

Proposer: archive every checkbox-complete change that mentions Files, including the agent shell and the webpack restore, and archive drop-files with its spec synced so live specs carry the policy.

Reviewer: that fails against the code. `AgentWorkspace.js` still has Notes and Knowledge and remaps `files` → `knowledge`. `inferSubAgents` still runs. Webpack still needs `AgentWorkspace.js`. Drop-files is the only change that matches that code. Merging the older deltas would make `openspec/specs/` require a tab the code deleted. Merging drop-files and moving it into `archive/` would hide the policy in the same list as the recipes it cancelled.

Proposer, revised: archive the five the drop-files design names, skip their spec merge, leave drop-files active, banner the two mixed active changes, and put the product-owner warning in OpenSpec docs and CLI context.

Reviewer: `morphai-remember-files-workspace-open` is mixed. Archiving it whole can be read as “delete pane persistence.”

Proposer, revised: archive it, because its Files-tab restore is the hazard and drop-files already restates pane open/closed. The banner says the kept part out loud. Do not sync `morphai-workspace-open-persist`, because that delta also requires selecting the Files tab.

Reviewer: the issue said six changes.

Proposer: the sixth implementer is `morphai-agent-workspace`, and it is not archived for the reason in decision 2. It is marked superseded for the Files tab only. That is the “archived or clearly marked” acceptance path for a mixed folder. Webpack is the same kind of mark, not a sixth archive.

Reviewer: accepted. No open question changes the spec, the archive set, or the tasks.
