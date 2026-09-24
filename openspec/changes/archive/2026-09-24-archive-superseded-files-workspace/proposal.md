# Proposal

## Why

Five checkbox-complete OpenSpec changes still tell agents to build Morph AI’s IndexedDB Files workspace (local folder, pins, Files tab). `morphai-drop-files-workspace` removed that surface, and `AgentWorkspace.js` now remaps a stored `files` tab to `knowledge`. Left active, those changes are a recipe to put the tab back.

## What Changes

- Archive the five Files-workspace changes under `openspec/changes/archive/` with a superseded marker pointing at `morphai-drop-files-workspace`. Keep their files. Do not merge their delta specs into `openspec/specs/`.
- Leave `morphai-drop-files-workspace` active and unsynced. It stays the source of truth for current Files policy.
- Add a short warning in OpenSpec project docs: do not restore the IndexedDB Files tab without an explicit product-owner yes.
- Do not archive the other complete changes, and do not edit `docs/agents/*` or `README.md`.

## Capabilities

### New Capabilities

- `openspec-files-workspace-archive`: Superseded Morph AI Files-workspace changes stay archived history, live specs do not require a Files tab, and OpenSpec docs warn against restoring that tab.

### Modified Capabilities

- (none)

## Impact

- `openspec/changes/` and `openspec/changes/archive/` only, plus `openspec/project.md` and the OpenSpec project context in `openspec/config.yaml`
- No Morph AI, MorphNotes, MorphUtils, or API code
- `openspec archive` defaults to merging delta specs; these five archives must use `--skip-specs`
