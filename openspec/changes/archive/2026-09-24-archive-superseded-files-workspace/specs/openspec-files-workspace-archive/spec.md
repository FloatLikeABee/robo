# Spec Delta

## Purpose

Keeps the removed Morph AI IndexedDB Files workspace out of live OpenSpec requirements, and points agents at the change that deleted it.

## ADDED Requirements

### Requirement: Files-workspace history is archived and marked superseded

These checkbox-complete changes specified the Morph AI local-folder Files workspace and MUST be stored as archive history, not as active changes:

- `morphai-files-session-workspace`
- `morphai-restore-folder-workspace-session`
- `morphai-files-folder-survives-refresh`
- `morphai-remember-files-workspace-open`
- `morphai-pinned-files-readable-context`

Each MUST live under `openspec/changes/archive/` with a `2026-09-24-` name prefix. The proposal in each archived folder MUST say it is superseded by `morphai-drop-files-workspace`. The original change files MUST remain in that archive folder.

#### Scenario: Active changes no longer include the Files workspace recipes

- **WHEN** an agent lists active OpenSpec changes
- **THEN** none of the five change names above is an active change
- **AND** each exists as an archived folder whose proposal names `morphai-drop-files-workspace` as the change that superseded it

#### Scenario: Archive keeps the old files

- **WHEN** an agent opens one of those archived folders
- **THEN** its proposal, design, tasks, and delta specs are still present

### Requirement: Removed Files behavior is not a live spec

`openspec/specs/` MUST NOT require the Morph AI agent shell to provide a Files tab, Open folder, recent folders, pin-from-folder, or an IndexedDB `morphai-files-workspace` database. Delta specs that belong only to the five archived changes MUST NOT be merged into `openspec/specs/`.

#### Scenario: Main specs do not describe a Files tab

- **WHEN** an agent reads requirements under `openspec/specs/`
- **THEN** no requirement tells them to build a Morph AI Files tab or IndexedDB Files workspace

### Requirement: Drop-files stays the Files policy

`morphai-drop-files-workspace` MUST remain the source of truth for current Morph AI Files policy. This work MUST leave that change active and MUST NOT merge its delta into `openspec/specs/`.

#### Scenario: Policy change is still active

- **WHEN** this archive work is complete
- **THEN** `openspec/changes/morphai-drop-files-workspace/` is still an active change
- **AND** `openspec/specs/` has no `morphai-no-files-workspace` spec produced from it

### Requirement: Mixed changes stay active and disclaim the Files tab

`morphai-agent-workspace` and `morphai-restore-missing-webpack-modules` MUST remain active changes. Each proposal MUST state that any instruction to build or keep the Morph AI Files tab or `AgentFilesTab` is superseded by `morphai-drop-files-workspace` and is not permission to restore that tab. Requirements in those changes for the agent shell, Notes & TODOs, Context & Knowledge, orchestration, and the other webpack modules MUST remain in the active folders.

#### Scenario: Agent workspace change disclaims the Files tab

- **WHEN** an agent opens the active `morphai-agent-workspace` proposal
- **THEN** it states that the Files tab is superseded by `morphai-drop-files-workspace`
- **AND** that change folder is not under `openspec/changes/archive/`

#### Scenario: Webpack change does not authorize AgentFilesTab

- **WHEN** an agent opens the active `morphai-restore-missing-webpack-modules` proposal
- **THEN** it states that `AgentFilesTab` must not be restored
- **AND** the change still records the other chat modules the bundle has to resolve

### Requirement: OpenSpec docs forbid restoring the Files tab

OpenSpec project docs MUST warn that the IndexedDB Files tab (local folder picker, recents, pins, `AgentFilesTab`) MUST NOT be restored unless the product owner explicitly says yes, and MUST point at `morphai-drop-files-workspace` for current policy. The warning MUST live under `openspec/` and MUST NOT be added to `docs/agents/` or `README.md`.

#### Scenario: Agent reads the OpenSpec warning

- **WHEN** an agent reads the OpenSpec project docs before changing Morph AI
- **THEN** they see the warning not to restore the IndexedDB Files tab without an explicit product-owner yes
- **AND** the warning names `morphai-drop-files-workspace` as the current Files policy
