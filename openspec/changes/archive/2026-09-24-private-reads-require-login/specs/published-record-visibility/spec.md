## Purpose

One published-or-not rule that HTTP public pages and in-process callers, including the read-only Morph MCP server, share so an unauthenticated caller cannot see a private record.

## ADDED Requirements

### Requirement: Published visibility is an in-process check
A record MUST be treated as publicly visible only when its published slug is non-empty after trimming. That decision MUST be available to Go callers in the same process, including the read-only MCP server, and MUST NOT depend on HTTP middleware having already run. An unauthenticated caller MUST NOT receive a record that fails this check.

#### Scenario: Empty slug is private
- **WHEN** a caller with no verified user asks whether a record with an empty published slug is publicly visible
- **THEN** the answer is no

#### Scenario: Whitespace slug is private
- **WHEN** a caller with no verified user asks whether a record whose published slug is only whitespace is publicly visible
- **THEN** the answer is no

#### Scenario: Non-empty slug is published
- **WHEN** a caller with no verified user asks whether a record with a non-empty published slug is publicly visible
- **THEN** the answer is yes

#### Scenario: MCP does not expose a private record without a user
- **WHEN** the read-only MCP server is asked to expose a record and the caller has no verified user id
- **THEN** the record is exposed only when it is publicly visible
- **AND** a record with an empty published slug is not exposed
