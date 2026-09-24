# Spec Delta

## REMOVED Requirements

### Requirement: Single tool registry
The system MUST maintain a single tool registry that holds every available tool regardless of source (built-in, custom, or external MCP host).

**Reason**: AI tools never had a real unified tool registry wired to a compliant MCP server. The TCP listener and host-tool placeholders are being deleted, not finished. The live spec heading for this block is blank (`### Requirement:`); match this removal by the sentence above.
**Migration**: Do not call BK for tools/list or tools/call. Protocol MCP is `morph/cmd/morph-mcp`.

#### Scenario: Registry is not a BK capability
- **WHEN** an operator looks for a BK tool registry that merges built-in, custom, and external MCP host tools
- **THEN** that capability is not part of AI tools

### Requirement: Common tool list interface
Tools MUST be discoverable by both the internal agent runtime and external MCP clients through a common list interface.

**Reason**: There is no BK MCP client surface. Agent runtime tool binding through this registry was never the live assistant path. The live spec heading for this block is blank; match this removal by the sentence above.
**Migration**: Assistants run without this registry. MCP clients connect to `morph/cmd/morph-mcp`.

#### Scenario: No BK tools/list
- **WHEN** a client sends a tools/list message to AI tools
- **THEN** AI tools does not answer it on a TCP MCP server

### Requirement: Structured tool invocation
Tool invocation MUST accept a structured arguments dict and route to the correct underlying implementation.

**Reason**: Host tool calls were placeholders (metadata or echo) and the host manager was not passed into the TCP server. The live spec heading for this block is blank; match this removal by the sentence above.
**Migration**: Do not store MCP host profiles in AI tools or call `/mcp/hosts`.

#### Scenario: No BK tools/call
- **WHEN** a client sends a tools/call message to AI tools
- **THEN** AI tools does not forward it to an external host

## ADDED Requirements

### Requirement: No BK TCP MCP
AI tools MUST NOT start a length-prefixed TCP server, MUST NOT expose `/mcp/start` or `/mcp/hosts`, and MUST NOT advertise that surface as the Model Context Protocol. The platform MCP server is the Morph Go program `morph/cmd/morph-mcp`.

#### Scenario: API startup
- **WHEN** the AI tools API process starts
- **THEN** it does not listen on port 8196 for MCP
- **AND** its HTTP API does not include `/mcp/start` or `/mcp/hosts`

#### Scenario: Operator docs
- **WHEN** an operator reads the AI tools docs that used to describe a BK MCP server
- **THEN** those docs do not claim that server works
- **AND** they name `morph/cmd/morph-mcp` as the protocol server
