## Purpose

Notes & TODOs and Context & Knowledge use the same dark-blue chrome as MorphNotes and Morph AI, not a leftover dark-grey panel fill.

## ADDED Requirements

### Requirement: Context & Knowledge uses dark-blue chrome

The Context & Knowledge surface (Morph AI workspace tab and the HybridContext drawer) MUST use the product dark-blue panel colors (the same family as `--chat-surface` / `--chat-panel` / `--chat-border`). It MUST NOT fill with a neutral dark grey such as `#1e1f24`. The UI stays dark; there is no light-mode switch.

#### Scenario: Workspace Knowledge tab

- **WHEN** the user opens Context & Knowledge in the Morph AI agent workspace
- **THEN** the panel background is dark blue like the chat chrome
- **AND** it is not a flat dark grey distinct from Files / the message column

#### Scenario: HybridContext drawer

- **WHEN** Context & Knowledge opens as a right drawer
- **THEN** that drawer uses the same dark-blue fill and blue-tinted border family as the workspace panel

### Requirement: MorphNotes Notes & TODOs uses dark-blue chrome

The MorphNotes Notes & TODOs header panel MUST use the same dark-blue surface family as MorphNotes chrome. It MUST NOT read as a separate dark-grey card. Header wash MAY be blue-tinted; it MUST NOT be leftover purple-grey.

#### Scenario: Open Notes & TODOs on MorphNotes

- **WHEN** the user opens Notes & TODOs from the MorphNotes header
- **THEN** the panel background matches MorphNotes dark-blue chrome
- **AND** it does not look like a grey overlay on a blue app

### Requirement: Chat Notes & TODOs drawer matches

When Notes & TODOs opens in the Morph AI (or MorphNotes AI) hybrid-drawer shell, that drawer MUST use the same dark-blue tokens as Context & Knowledge, not the grey `#1e1f24` fallback.

#### Scenario: Notes drawer from chat

- **WHEN** the user opens Notes & TODOs from Morph AI chat
- **THEN** the drawer fill is dark blue like Context & Knowledge after this change
