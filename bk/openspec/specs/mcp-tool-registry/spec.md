# MCP Tool Registry

## Purpose

This capability is withdrawn. AI tools does not host a tool registry or a TCP MCP server. The platform Model Context Protocol server is the Morph Go program `morph/cmd/morph-mcp`.

## Requirements

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
