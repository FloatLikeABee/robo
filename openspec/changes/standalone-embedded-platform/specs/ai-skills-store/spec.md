## Purpose

Defines Morph AI’s skills library: many skills available to assistants, with user upload and management so the project can grow a reusable skills catalog.

## ADDED Requirements

### Requirement: Skills catalog
The platform SHALL provide a skills store that holds many skills (bundled defaults and user-uploaded skills) addressable by id, with metadata sufficient for listing and assistant selection (name, description, version or updated time, enabled flag).

#### Scenario: List skills
- **WHEN** an authenticated Morph AI client requests the skills list
- **THEN** the system SHALL return the available skills with id and display metadata

### Requirement: User can upload skills
Authenticated users SHALL be able to upload a skill package or skill definition into the skills store so it becomes available to Morph AI according to enablement rules.

#### Scenario: Successful upload
- **WHEN** a user uploads a valid skill through the skills upload API or UI
- **THEN** the skill SHALL be stored in the embedded platform stores and SHALL appear in subsequent skills list responses

#### Scenario: Invalid skill rejected
- **WHEN** a user uploads an invalid or unsupported skill payload
- **THEN** the system SHALL reject the upload with a clear error and SHALL NOT register the skill as available

### Requirement: Assistants can consume skills
Morph AI assistants SHALL be able to load or reference skills from the skills store during chat/tool loops according to product rules (for example enabled skills only).

#### Scenario: Assistant uses an enabled skill
- **WHEN** an assistant run is configured or prompted to use an enabled skill from the store
- **THEN** the runtime SHALL be able to retrieve that skill’s definition from the skills store without requiring MySQL, MongoDB, or Redis

### Requirement: Manage skills lifecycle
Users with appropriate access SHALL be able to disable or delete user-uploaded skills they manage, without breaking the store for remaining skills.

#### Scenario: Delete uploaded skill
- **WHEN** an authorized user deletes a user-uploaded skill
- **THEN** the skill SHALL no longer appear as available in the skills list and SHALL not be loadable by new assistant runs
