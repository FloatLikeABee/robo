## Purpose

Defines Dorph’s default persistence model: relational data in an embedded SQL file store, document payloads in Badger on disk, and ephemeral cache in process memory—so Dorph runs without MySQL, MongoDB, or Redis.

## ADDED Requirements

### Requirement: Embedded-only Dorph boot
Dorph SHALL start and serve Tran/Dorph Data APIs using only embedded stores (file-backed SQL for relational data, Badger for document payloads, in-process cache for ephemeral keys) without requiring a running MySQL, MongoDB, or Redis server.

#### Scenario: Fresh install without networked databases
- **WHEN** an operator starts Dorph with default embedded data paths and no `TRAN_MYSQL_DSN` / `TRAN_MONGO_URI` / `TRAN_REDIS_ADDR`
- **THEN** Dorph SHALL open or create the embedded stores under the configured data directory and SHALL NOT fail solely because MySQL, MongoDB, or Redis are unreachable

#### Scenario: Tran endpoints available when SQL store opens
- **WHEN** the embedded relational store opens successfully
- **THEN** Tran list/create/update/delete endpoints that previously returned 503 for “MySQL not configured” SHALL operate against the embedded SQL store instead

### Requirement: Relational data in embedded SQL
Dorph SHALL persist Tran relational entities (including auth `plat_users`, facilities, members, employees, contacts, case tasks, stories, comments, grid filters/colors, attachment metadata, and other MySQL-backed Tran tables Dorph owns) in an embedded SQL database file.

#### Scenario: User create and list round-trip
- **WHEN** a client creates a Tran entity via the existing HTTP API and then lists that entity type
- **THEN** the created row SHALL appear in the list with the same fields Dorph’s API already exposes

#### Scenario: Auth bootstrap without MySQL
- **WHEN** Dorph starts with embedded SQL and configured bootstrap admin credentials
- **THEN** Dorph SHALL ensure the admin account exists in the embedded SQL `plat_users` table (or equivalent) using the same login/JWT behavior as today

### Requirement: Document details in Badger
Dorph SHALL store entity detail JSON (today’s Mongo `entity_details` documents keyed by entity type and record id) in Badger such that get, upsert, and delete preserve empty/non-empty semantics used by Dorph Data detail panels and seed/prune tools.

#### Scenario: Detail JSON upsert and read
- **WHEN** a client saves detail JSON for an entity record and later loads that record
- **THEN** Dorph SHALL return the same JSON body Dorph would have returned from Mongo for a successful get

#### Scenario: Delete removes detail document
- **WHEN** a Tran record is deleted through Dorph’s API
- **THEN** Dorph SHALL remove the corresponding Badger detail document for that entity key and record id (including staff/employee mirror keys where Dorph already dual-writes)

### Requirement: Cache without Redis
Dorph SHALL provide any Dorph-owned cache/session behavior formerly intended for Redis via in-process cache with TTL semantics, and SHALL NOT require Redis for Dorph to run.

#### Scenario: Dorph runs with Redis unset
- **WHEN** `TRAN_REDIS_ADDR` is unset or Redis is stopped
- **THEN** Dorph SHALL continue serving APIs that do not depend on Redis, and any Dorph cache reads/writes SHALL use the in-process cache

### Requirement: Stable HTTP contracts
Dorph SHALL keep existing Dorph/Tran HTTP request and response shapes for entity CRUD and detail JSON unless a change is explicitly documented as breaking in the migration notes.

#### Scenario: Existing Dorph Data client still works
- **WHEN** Dorph Data frontend calls existing `/api/tran/...` routes after the storage swap
- **THEN** successful responses SHALL remain JSON-compatible with the pre-migration client without requiring a mandatory frontend rewrite

### Requirement: Configurable data locations
Dorph SHALL allow operators to configure filesystem paths for the embedded SQL database file and Badger document store, defaulting under Dorph’s local `./data/` tree.

#### Scenario: Custom data paths
- **WHEN** an operator sets custom path env vars for SQL and Badger document storage
- **THEN** Dorph SHALL read and write those paths and SHALL create parent directories as needed when permitted by the OS
