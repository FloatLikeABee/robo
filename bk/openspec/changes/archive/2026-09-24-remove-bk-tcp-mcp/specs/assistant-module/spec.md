# Spec Delta

## MODIFIED Requirements

### Requirement: Tool system removal
ToolManager, ToolRegistry, src/tools.py, and built-in tool wrappers MUST be removed. Agents and services MUST operate without tool binding. AI tools MUST NOT include an MCP service that answers tools/list.

#### Scenario: Agent without tools

- **GIVEN** an agent configuration has no tools
- **WHEN** the agent runs
- **THEN** it uses direct LLM invocation
- **AND** no ToolManager lookup occurs

#### Scenario: MCP tools/list after removal

- **GIVEN** the BK TCP MCP server has been removed
- **WHEN** the AI tools API is running
- **THEN** it does not accept a tools/list message on a BK MCP server
