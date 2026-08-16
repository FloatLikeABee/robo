## Purpose

Ensure Neo4j is available for Morph GraphRAG during local full-stack development without requiring a separate manual step.

## ADDED Requirements

### Requirement: start-all checks Neo4j before starting apps

When starting the dev stack (`start-all`, `start all`, or `restart all`), the launcher SHALL verify Neo4j bolt port **7687** is listening on the local machine before starting application services.

#### Scenario: Neo4j already running

- **WHEN** port 7687 is already listening
- **THEN** the launcher proceeds without attempting to start Neo4j again

### Requirement: start-all starts Neo4j when CLI is available

If port 7687 is not listening and the `neo4j` command exists, the launcher SHALL run `neo4j start` and wait briefly for the port to become available.

#### Scenario: Successful neo4j start

- **WHEN** port 7687 is closed and `neo4j` is on PATH
- **THEN** the launcher runs `neo4j start` and continues starting app services after Neo4j is up or a short timeout elapses

### Requirement: Missing Neo4j is non-fatal

If Neo4j cannot be started (CLI missing or start fails), the launcher SHALL print a warning and SHALL still start other application services.

#### Scenario: Neo4j not installed

- **WHEN** port 7687 is closed and `neo4j` is not on PATH
- **THEN** the launcher prints a warning that Morph graph features may be unavailable and continues starting apps
