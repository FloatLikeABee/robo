## Purpose

Makes a skill checked in Morph AI chat govern the next assistant reply.

## ADDED Requirements

### Requirement: Selected skills are sent and followed
When the operator checks one or more skills in the chat picker and sends a message, the request MUST include those skill ids. The assistant system prompt for that turn MUST include each selected skill’s instruction body and MUST tell the model to follow those skills for the reply. A selected skill that exists in the picker MUST NOT be omitted because it is absent from a second list.

#### Scenario: A checked skill is in the prompt
- **WHEN** the operator checks a skill that has instructions and sends a message
- **THEN** the chat request includes that skill id
- **AND** the system prompt contains that skill’s instructions
- **AND** the prompt says to follow the selected skills for this reply

#### Scenario: Unchecked skills are not required
- **WHEN** the operator sends a message with no skills checked
- **THEN** the request does not include selected skill ids
- **AND** the prompt does not require a picker skill
