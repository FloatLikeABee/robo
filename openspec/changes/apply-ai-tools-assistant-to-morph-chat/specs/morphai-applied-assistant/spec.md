## Purpose

Lets a user apply one AI Tools assistant to Morph chat from the Assistants list, see that it is in effect, and dismiss it so chat returns to unscoped behavior.

## ADDED Requirements

### Requirement: Apply replaces Run on AI Tools assistants

The AI Tools Assistants list SHALL present an Apply control on each assistant instead of Run. Apply MUST NOT open a one-shot query dialog. The Assistants page MUST NOT offer a Run dialog or a Run action that posts a standalone query to the assistant run endpoint.

#### Scenario: Apply control is shown

- **WHEN** a user views the AI Tools Assistants list
- **THEN** each assistant has an Apply control
- **AND** no Run control or Run dialog is offered on that page

#### Scenario: Apply does not run a one-shot query

- **WHEN** a user activates Apply on an assistant
- **THEN** the system does not open a query dialog for that assistant
- **AND** the system does not treat the click as a one-shot assistant run

### Requirement: Apply attaches the assistant to Morph chat

Activating Apply on an assistant SHALL attach that assistant to Morph chat. Subsequent Morph chat messages MUST be sent with that assistant selected so Morph uses its prompt, model, and RAG. At most one assistant SHALL be applied at a time. Applying a different assistant SHALL replace the previously applied one.

#### Scenario: Apply from AI Tools while Morph chat is open

- **WHEN** a user activates Apply on an assistant from AI Tools with Morph chat available
- **THEN** that assistant is the applied assistant for Morph chat
- **AND** later messages in Morph chat use that assistant

#### Scenario: Applying another assistant replaces the current one

- **WHEN** assistant A is applied and the user activates Apply on assistant B
- **THEN** assistant B is the applied assistant
- **AND** assistant A is no longer applied

#### Scenario: Apply when Morph chat is not available

- **WHEN** a user activates Apply and Morph chat cannot receive the assistant
- **THEN** the assistant is not silently treated as applied
- **AND** the user is told that Morph chat is required to apply the assistant

### Requirement: Dismiss removes the applied assistant

The system SHALL provide a Dismiss control wherever the applied assistant is shown (AI Tools Assistants list and Morph chat). Activating Dismiss SHALL clear the applied assistant. After dismiss, Morph chat messages MUST be sent with no assistant selected.

#### Scenario: Dismiss from AI Tools

- **WHEN** an assistant is applied and the user activates Dismiss on that assistant in the AI Tools Assistants list
- **THEN** no assistant is applied
- **AND** later Morph chat messages are not scoped to that assistant

#### Scenario: Dismiss from Morph chat

- **WHEN** an assistant is applied and the user activates Dismiss in Morph chat
- **THEN** no assistant is applied
- **AND** the AI Tools Assistants list no longer shows that assistant as applied

#### Scenario: Dismiss control only for the applied assistant

- **WHEN** assistant A is applied
- **THEN** assistant A shows Dismiss instead of Apply
- **AND** other assistants still show Apply

### Requirement: Morph chat shows the applied assistant

While an assistant is applied, Morph chat SHALL show the applied assistant's name and a Dismiss control. Chat requests MUST include the applied assistant id. After dismiss or when none is applied, Morph chat MUST NOT send an assistant id.

#### Scenario: Applied assistant visible in Morph chat

- **WHEN** an assistant is applied
- **THEN** Morph chat displays that assistant's name
- **AND** a Dismiss control is available in Morph chat without opening AI Tools

#### Scenario: Chat request carries the applied assistant

- **WHEN** the user sends a Morph chat message while an assistant is applied
- **THEN** the chat request includes that assistant's id

#### Scenario: Chat request after dismiss has no assistant

- **WHEN** the user sends a Morph chat message after dismissing the applied assistant
- **THEN** the chat request does not include an assistant id
