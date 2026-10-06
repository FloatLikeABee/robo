## Purpose

Lets the operator choose an AI tools assistant from the Morph AI composer.

## ADDED Requirements

### Requirement: The composer selects one assistant
The chat composer MUST offer an assistant picker in the same control family as the skills picker. The picker MUST list AI tools assistants. The operator MUST be able to select one assistant or clear the selection. The next send MUST use that assistant’s system prompt. The AI tools assistant panel MUST NOT use an Apply button to attach an assistant to chat.

#### Scenario: Select an assistant and send
- **WHEN** the operator selects an assistant in the chat picker and sends a message
- **THEN** the request carries that assistant id
- **AND** the reply is generated with that assistant’s system prompt

#### Scenario: Clear the assistant
- **WHEN** the operator clears the selected assistant and sends a message
- **THEN** the request does not carry an assistant id
- **AND** the reply is not generated as that assistant

#### Scenario: No Apply button
- **WHEN** the operator opens the AI tools assistant panel
- **THEN** there is no Apply control that attaches an assistant to Morph chat
