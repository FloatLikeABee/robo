# Assistant Module

## Purpose

Define the unified Assistant module that replaces both Adviser and Customization, and define how the system operates after the legacy tool system is removed.

## Requirements

### Requirement: Assistant unifies adviser and customization
The Assistant module MUST provide a single profile type that captures both adviser-style file-backed agents and customization-style instruction overrides.

#### Scenario: Create assistant

- **GIVEN** a user provides name, system prompt, and optional files or RAG collections
- **WHEN** the assistant is created
- **THEN** a single AssistantProfile is persisted
- **AND** an underlying Agent is created or linked

#### Scenario: Migrate legacy adviser

- **GIVEN** existing adviser profiles exist
- **WHEN** migration runs
- **THEN** each adviser becomes an AssistantProfile with source='adviser'

#### Scenario: Migrate legacy customization

- **GIVEN** existing customization profiles exist
- **WHEN** migration runs
- **THEN** each customization becomes an AssistantProfile with source='customization'

### Requirement: Tool system removal
ToolManager, ToolRegistry, src/tools.py, and built-in tool wrappers MUST be removed. Agents and services MUST operate without tool binding.

#### Scenario: Agent without tools

- **GIVEN** an agent configuration has no tools
- **WHEN** the agent runs
- **THEN** it uses direct LLM invocation
- **AND** no ToolManager lookup occurs

#### Scenario: MCP tools/list after removal

- **GIVEN** the tool system has been removed
- **WHEN** MCPService receives tools/list
- **THEN** it returns an empty list or host-only tools

### Requirement: API and storage consolidation
API endpoints and TinyDB storage for advisers and customizations MUST be replaced by a single /assistants resource and assistants.json storage.

#### Scenario: List assistants

- **GIVEN** multiple assistants exist
- **WHEN** GET /assistants is called
- **THEN** all profiles are returned in one list

#### Scenario: Delete legacy endpoints

- **GIVEN** the refactor is complete
- **WHEN** checking API routes
- **THEN** /advisers and /customizations endpoints are removed
- **AND** frontend routes point to /assistants
