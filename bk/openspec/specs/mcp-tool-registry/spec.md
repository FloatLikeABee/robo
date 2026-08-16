# MCP Tool Registry

## Purpose

Define how tools are registered, discovered, and invoked by the MCP service and agent runtime in Ground Control. This spec establishes a unified registry abstraction so that built-in tools, custom tools, and external MCP hosts can be exposed consistently to agents and MCP clients.

## Requirements

### Requirement: 
The system MUST maintain a single tool registry that holds every available tool regardless of source (built-in, custom, or external MCP host).

#### Scenario: Registry initialization

- **GIVEN** the Ground Control backend starts
- **WHEN** the ToolManager is instantiated
- **THEN** all built-in tools are loaded into the registry
- **AND** each entry has a unique tool_id, a ToolConfig, and an invocable function

#### Scenario: External host registration

- **GIVEN** an MCPHostProfile is marked active
- **WHEN** the host is connected
- **THEN** its exposed tools are registered under the same registry with prefixed tool_ids

### Requirement: 
Tools MUST be discoverable by both the internal agent runtime and external MCP clients through a common list interface.

#### Scenario: Agent tool selection

- **GIVEN** an agent configuration references a list of tool_ids
- **WHEN** the agent is run
- **THEN** only tools that exist in the registry are bound to the agent
- **AND** missing tool_ids are logged as warnings

#### Scenario: MCP tools/list

- **GIVEN** an MCP client sends a tools/list message
- **WHEN** MCPService processes the message
- **THEN** it returns the merged list of built-in, custom, and external host tools

### Requirement: 
Tool invocation MUST accept a structured arguments dict and route to the correct underlying implementation.

#### Scenario: Built-in tool call

- **GIVEN** a tools/call message names a built-in tool_id
- **WHEN** MCPService handles the message
- **THEN** it invokes the registered function with the arguments dict
- **AND** returns a text content response

#### Scenario: External host tool call

- **GIVEN** a tools/call message names an external host tool
- **WHEN** MCPService handles the message
- **THEN** it forwards the call to the external host
- **AND** returns the host's response unchanged
