# Tasks

## 1. Regression check (red first)

- [x] 1.1 Add `openspec/check_files_workspace_archive.py` covering the spec: the five Files-workspace changes are under `openspec/changes/archive/2026-09-24-*` with a superseded-by-`morphai-drop-files-workspace` proposal and their design, tasks, and specs still present; those names are not active; `openspec/specs/` has none of their capabilities and no `morphai-no-files-workspace`; `morphai-drop-files-workspace` is still active; `morphai-agent-workspace` and `morphai-restore-missing-webpack-modules` are still active and their proposals disclaim the Files tab / `AgentFilesTab`; `openspec/project.md` warns not to restore the IndexedDB Files tab without an explicit product-owner yes. Verify the script exits non-zero before any archive move, and the failure names the still-active changes.

## 2. Archive the five Files-workspace changes

- [x] 2.1 Prepend the superseded banner from design.md (including the extra sentences for `morphai-remember-files-workspace-open` and `morphai-restore-folder-workspace-session`) to each of the five proposals. Verify each proposal still has its `## Why` section and now names `morphai-drop-files-workspace`.
- [x] 2.2 Run `openspec archive <name> --skip-specs --yes` for each of the five. Verify the CLI prints that spec updates were skipped, each folder landed at `openspec/changes/archive/2026-09-24-<name>/`, and `openspec/specs/` did not gain `morphai-files-session-workspace`, `morphai-folder-workspace-session`, `morphai-files-folder-persist`, `morphai-pinned-files-context`, or `morphai-workspace-open-persist`.

## 3. Disclaimers and OpenSpec warning

- [x] 3.1 Add the Files-tab superseded banner to the active proposals for `morphai-agent-workspace` and `morphai-restore-missing-webpack-modules` without moving those folders or deleting their other requirements. Verify both proposals name `morphai-drop-files-workspace` and both directories are still under `openspec/changes/`.
- [x] 3.2 Add `openspec/project.md` with the product-owner warning, and add the same warning to `openspec/config.yaml` context pointing at `openspec/project.md`. Verify neither `docs/agents/` nor `README.md` changed, and `openspec/project.md` tells agents not to restore the IndexedDB Files tab without an explicit product-owner yes.

## 4. Integration check

- [x] 4.1 Re-run `openspec/check_files_workspace_archive.py` and verify it exits 0. Run `openspec validate --all` and `openspec list` and verify the five names are absent from the active list, `morphai-drop-files-workspace` is still listed, and no live spec requires a Morph AI Files tab.
