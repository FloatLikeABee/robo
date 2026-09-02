## Purpose

Let MorphAI chat users choose one or more skills from a compact picker next to the attachment button so the selected skills are included in the chat request and influence the assistant's behavior.

## ADDED Requirements

### Requirement: Skills picker button beside the attachment button

The MorphAI chat input SHALL include a skills button positioned next to the attachment button. Activating it SHALL open a compact picker positioned above the input area.

#### Scenario: Opening the skills picker

- **WHEN** a user clicks the skills button in the chat input
- **THEN** a compact picker opens above the input showing the available skills

#### Scenario: Dismissing the skills picker

- **WHEN** the skills picker is open and the user clicks outside it or reactivates the skills button
- **THEN** the picker closes without altering the current message draft

### Requirement: Selecting skills for a chat request

The skills picker SHALL allow selecting and deselecting one or more skills. The current selection SHALL be visually indicated, and SHALL persist while composing the message.

#### Scenario: Selecting multiple skills

- **WHEN** a user selects two skills in the picker
- **THEN** both skills are shown as selected and remain selected when the picker is reopened

#### Scenario: Deselecting a skill

- **WHEN** a user deselects a previously selected skill
- **THEN** that skill is no longer marked selected and will not be included in the next request

### Requirement: Selected skills are sent with the chat request

When a user sends a message with one or more skills selected, MorphAI SHALL include the selected skill identifiers in the chat request so the backend applies them.

#### Scenario: Sending a message with skills selected

- **WHEN** a user has one or more skills selected and sends a message
- **THEN** the chat request includes the selected skill identifiers

#### Scenario: Sending a message with no skills selected

- **WHEN** a user sends a message with no skills selected
- **THEN** the chat request is sent without any skill identifiers and behaves as before
