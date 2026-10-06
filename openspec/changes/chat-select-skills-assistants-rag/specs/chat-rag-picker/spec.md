## Purpose

Lets the operator choose RAG collections from the Morph AI composer and use them on the next send.

## ADDED Requirements

### Requirement: The composer selects RAG collections
The chat composer MUST offer a RAG picker in the same control family as the skills picker. The picker MUST list AI tools RAG collections. The operator MUST be able to check more than one collection and clear them. The next send MUST retrieve the checked collections for the message and include the snippets in the prompt. Retrieval MUST run even when no assistant is selected.

#### Scenario: Selected collections are retrieved
- **WHEN** the operator checks a RAG collection and sends a message
- **THEN** the request includes that collection
- **AND** the prompt includes snippets retrieved from that collection for the message

#### Scenario: No collection means no extra retrieval
- **WHEN** the operator sends a message with no RAG collection checked
- **THEN** the request does not include selected RAG collections
- **AND** the prompt does not add snippets from a chat-selected collection

#### Scenario: Assistant RAG still applies
- **WHEN** an assistant that has its own RAG collections is selected and the operator also checks a collection
- **THEN** the send uses the assistant’s collections and the checked collection
