## Purpose

Defines Academi primary navigation labels and removal of Profile inside Morph Utils.

## ADDED Requirements

### Requirement: Chat renamed Academi AI
Academi SHALL label the former Chat surface as **Academi AI** in primary navigation.

#### Scenario: Academi AI label
- **WHEN** a signed-in user opens Academi
- **THEN** the chat/study assistant entry is labeled Academi AI

### Requirement: Community renamed Board
Academi SHALL label the former Community surface as **Board**.

#### Scenario: Board label
- **WHEN** a user views Academi primary navigation
- **THEN** the community entry is labeled Board

### Requirement: Profile removed
Academi SHALL NOT expose a Profile item in primary navigation or as a module-level profile settings entry.

#### Scenario: No Profile nav
- **WHEN** a signed-in user opens Academi
- **THEN** Profile is not present in primary navigation
