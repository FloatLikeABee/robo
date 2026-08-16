## Purpose

Defines the Morph AI empty-chat welcome state as a minimal brand moment with no instructional tip list or quick links.

## ADDED Requirements

### Requirement: Empty chat shows logo only
When Morph AI has no messages in the current session, the welcome panel SHALL show the Morph AI logo (and optional brand title) and MUST NOT show the "Here's what you can do:" heading, tip bullets, or other quick-link style instructional list items.

#### Scenario: Fresh empty session
- **WHEN** a user opens Morph AI with an empty message list
- **THEN** they see the logo/brand welcome
- **AND** they do not see "Here's what you can do:" or any feature tip list

#### Scenario: After messages exist
- **WHEN** the session already has one or more messages
- **THEN** the empty welcome panel is not shown
