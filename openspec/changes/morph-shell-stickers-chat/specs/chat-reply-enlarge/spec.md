## Purpose

Lets operators read a long Morph AI assistant reply in a large dark modal without leaving the conversation.

## ADDED Requirements

### Requirement: Long assistant replies can enlarge
An assistant message whose rendered body is taller than ten times its line height MUST offer a control that opens that reply in a modal. The control MUST NOT appear on a shorter assistant message. User messages MUST NOT gain this control. The thread MUST keep the full reply visible; enlarging MUST NOT replace the bubble with a ten-line clamp.

#### Scenario: Tall reply offers enlarge
- **WHEN** an assistant reply renders taller than ten line-heights
- **THEN** the message offers an enlarge control

#### Scenario: Short reply has no enlarge control
- **WHEN** an assistant reply renders at ten line-heights or fewer
- **THEN** the message has no enlarge control

#### Scenario: Enlarge does not clip the thread
- **WHEN** an operator has not opened the modal
- **THEN** the full assistant reply remains in the transcript

### Requirement: Enlarge modal is a dark reading stage
The modal MUST use a dark background and cover nearly the full viewport. The reply text MUST be readable and scrollable inside it. Escape, a backdrop click, and a close control MUST dismiss it. Diagram and image enlarge already in the product MUST keep working on the same reply.

#### Scenario: Open and dismiss
- **WHEN** the operator activates enlarge on a long assistant reply
- **THEN** a dark near-full-viewport modal shows that reply
- **AND** Escape dismisses the modal and returns focus to the conversation
