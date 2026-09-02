## Purpose

Lets Morph Data Assets fill Detail (JSON) from a plain-text description using AI extract, so users do not have to hand-write JSON.

## ADDED Requirements

### Requirement: Describe an asset and extract Detail JSON
On Assets create or edit, the user SHALL be able to enter plain-text description of the asset and run AI extract into the Detail (JSON) field. The extracted JSON MUST be shown for review; it MUST NOT overwrite existing detail until the user applies it.

#### Scenario: Extract JSON onto an empty Detail field
- **WHEN** a user is creating or editing an asset with empty Detail (JSON)
- **AND** they paste or type a description and run AI extract
- **THEN** Detail (JSON) is filled with a JSON object derived from that description
- **AND** the user can edit the JSON before saving the asset

#### Scenario: Extract does not auto-save the asset
- **WHEN** a user runs AI extract on Assets Detail (JSON)
- **THEN** the asset record is not saved until the user saves the asset
- **AND** cancelling the drawer/editor discards unsaved extracted JSON

### Requirement: Existing Detail JSON is preserved until apply
If Detail (JSON) already has content, AI extract MUST NOT silently replace it. The user MUST confirm applying the extracted JSON (replace or merge as presented in the UI) before the field changes.

#### Scenario: Confirm replace on existing detail
- **WHEN** an asset already has Detail (JSON)
- **AND** the user runs AI extract and confirms applying it
- **THEN** Detail (JSON) updates to the extracted result the user confirmed

#### Scenario: Cancel apply on existing detail
- **WHEN** an asset already has Detail (JSON)
- **AND** the user runs AI extract but cancels applying it
- **THEN** the previous Detail (JSON) remains unchanged
