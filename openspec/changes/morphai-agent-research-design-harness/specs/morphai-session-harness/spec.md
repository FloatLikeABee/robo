## Purpose

Turns significant Morph AI chat sessions into durable lessons so later chats reuse what worked, without harvesting every short exchange.

## ADDED Requirements

### Requirement: Harvest only significant sessions

After a Morph AI session becomes significant, the system MUST distill at most one new lesson from it. Short greetings and low-context chats MUST NOT produce a lesson.

#### Scenario: Significant session is harvested

- **WHEN** a session has substantial user turns or used tools or document context
- **THEN** a durable lesson is stored (trigger + rule)
- **AND** the lesson is attributed to that session

#### Scenario: Greeting is skipped

- **WHEN** the operator only sent a short greeting with no tools and no documents
- **THEN** no lesson is stored

### Requirement: Later chats use stored lessons

Later Morph AI chats MUST include recent matching lessons in the agent context so the agent can follow them.

#### Scenario: Lesson applies on a later chat

- **WHEN** a lesson exists from a prior significant session
- **AND** the operator starts a new chat that matches the lesson trigger
- **THEN** that lesson’s rule is present in the agent context
