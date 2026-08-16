## Purpose

Defines asynchronous mirroring of durable uploads and content writes into Neo4j so Morph AI can retrieve platform knowledge without blocking product APIs on graph availability.

## ADDED Requirements

### Requirement: Async Neo4j ingest for durable writes
The platform SHALL enqueue an asynchronous Neo4j ingest job when durable content is created or updated (including file uploads, published pages, knowledge documents, survey/sheet assets where applicable, and AI skills), without requiring Neo4j success before acknowledging the primary write.

#### Scenario: Upload succeeds when Neo4j is briefly down
- **WHEN** a user uploads a durable file or document and Neo4j is temporarily unreachable
- **THEN** the upload API SHALL still persist the primary store write successfully and SHALL enqueue or retain an ingest job for later Neo4j sync

#### Scenario: Ingest eventually upserts graph content
- **WHEN** an ingest job runs successfully for an uploaded or updated durable item
- **THEN** Neo4j SHALL contain a corresponding node/relationship representation usable by Morph AI retrieval paths for that item type

### Requirement: Retry and observability for ingest
The platform SHALL retry failed Neo4j ingest jobs with bounded backoff and SHALL expose enough status for operators to detect stuck or failed syncs (logs and/or an admin-visible queue/status surface).

#### Scenario: Failed ingest is retried
- **WHEN** a Neo4j ingest attempt fails with a transient error
- **THEN** the system SHALL retry according to the configured policy until success or a terminal failure state is recorded

### Requirement: Scope of mirrored content
Anything the product treats as a durable user or operator upload into Morph AI / Morph Data / Utils / skills stores SHALL be included in the async Neo4j mirror path unless explicitly marked ephemeral.

#### Scenario: Skill upload is mirrored
- **WHEN** a user uploads a skill into the AI skills store
- **THEN** the system SHALL enqueue Neo4j ingest for that skill so assistants can discover it via graph/AI retrieval after sync
