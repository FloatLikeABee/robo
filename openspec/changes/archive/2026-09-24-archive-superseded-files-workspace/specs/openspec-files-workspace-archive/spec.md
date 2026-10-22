# Spec Delta

## Purpose

Keeps the removed Dorph AI IndexedDB Files workspace out of live OpenSpec requirements, and points agents at the change that deleted it.

## ADDED Requirements

### Requirement: Files-workspace history is archived and marked superseded

These checkbox-complete changes specified the Dorph AI local-folder Files workspace and MUST be stored as archive history, not as active changes:

- `dorphai-files-session-workspace`
- `dorphai-restore-folder-workspace-session`
- `dorphai-files-folder-survives-refresh`
- `dorphai-remember-files-workspace-open`
- `dorphai-pinned-files-readable-context`

Each MUST live under `openspec/changes/archive/` with a `2026-09-24-` name prefix. The proposal in each archived folder MUST say it is superseded by `dorphai-drop-files-workspace`. The original change files MUST remain in that archive folder.

#### Scenario: Active changes no longer include the Files workspace recipes

- **WHEN** an agent lists active OpenSpec changes
- **THEN** none of the five change names above is an active change
- **AND** each exists as an archived folder whose proposal names `dorphai-drop-files-workspace` as the change that superseded it

#### Scenario: Archive keeps the old files

- **WHEN** an agent opens one of those archived folders
- **THEN** its proposal, design, tasks, and delta specs are still present

### Requirement: Removed Files behavior is not a live spec

`openspec/specs/` MUST NOT require the Dorph AI agent shell to provide a Files tab, Open folder, recent folders, pin-from-folder, `filesWorkspaceStore`, or an IndexedDB `dorphai-files-workspace` database. Delta specs that belong only to the five archived changes MUST NOT be merged into `openspec/specs/`. The synced drop-policy spec `dorphai-no-files-workspace` MAY name those surfaces only to forbid them.

#### Scenario: Main specs do not describe a Files tab

- **WHEN** an agent reads requirements under `openspec/specs/`
- **THEN** no requirement tells them to build a Dorph AI Files tab or IndexedDB Files workspace

### Requirement: Drop policy is the Files source of truth

Current Dorph AI Files policy MUST come from `dorphai-drop-files-workspace`: that change while it is active, or, after it is archived, the synced live spec `openspec/specs/dorphai-no-files-workspace/spec.md`. Agents MUST NOT take Files policy from the five archived Files-workspace changes.

#### Scenario: Active drop change is the policy

- **WHEN** `dorphai-drop-files-workspace` is an active change
- **THEN** that change is the Files policy

#### Scenario: Archived drop change stays the policy through its synced spec

- **WHEN** `dorphai-drop-files-workspace` is archived and its delta has been synced
- **THEN** `openspec/specs/dorphai-no-files-workspace/spec.md` is the Files policy
- **AND** the five archived Files-workspace changes are not the policy

### Requirement: Mixed changes stay active and disclaim the Files tab

`dorphai-agent-workspace`, `dorphai-restore-missing-webpack-modules`, and `platform-trim-readme-chat-skills` MUST remain active changes. Each proposal MUST state that any instruction to build or keep the Dorph AI Files tab or IndexedDB Files workspace is superseded by `dorphai-drop-files-workspace` and is not permission to restore that tab. Requirements in those changes for the agent shell, Notes & TODOs, Context & Knowledge, orchestration, header chrome, and the other webpack modules MUST remain in the active folders.

#### Scenario: Agent workspace change disclaims the Files tab

- **WHEN** an agent opens the active `dorphai-agent-workspace` proposal
- **THEN** it states that the Files tab is superseded by `dorphai-drop-files-workspace`
- **AND** that change folder is not under `openspec/changes/archive/`

#### Scenario: Webpack change does not authorize AgentFilesTab

- **WHEN** an agent opens the active `dorphai-restore-missing-webpack-modules` proposal
- **THEN** it states that `AgentFilesTab` must not be restored
- **AND** the change still records the other chat modules the bundle has to resolve

#### Scenario: Chrome change disclaims the Files tab

- **WHEN** an agent opens the active `platform-trim-readme-chat-skills` proposal
- **THEN** it states that the IndexedDB Files tab is superseded by `dorphai-drop-files-workspace`
- **AND** that change folder is not under `openspec/changes/archive/`

### Requirement: OpenSpec docs forbid restoring the Files tab

OpenSpec project docs MUST warn that the IndexedDB Files tab (local folder picker, recents, pins, `AgentFilesTab`) MUST NOT be restored unless the product owner explicitly says yes, and MUST point at `dorphai-drop-files-workspace` for current policy. The warning MUST live under `openspec/` and MUST NOT be added to `docs/agents/` or `README.md`.

#### Scenario: Agent reads the OpenSpec warning

- **WHEN** an agent reads the OpenSpec project docs before changing Dorph AI
- **THEN** they see the warning not to restore the IndexedDB Files tab without an explicit product-owner yes
- **AND** the warning names `dorphai-drop-files-workspace` as the current Files policy
