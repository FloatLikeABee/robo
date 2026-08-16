## Purpose

Defines the project-wide rule that Morph AI, Morph Data, Morph Utils, and AI tool backends persist only with embedded relational SQL, Badger document storage, and in-process cache—never requiring MySQL, MongoDB, or Redis at runtime.

## ADDED Requirements

### Requirement: No networked MySQL Mongo Redis at boot
Morph AI, Morph Data, Morph Utils backends, and AI tool services covered by this platform SHALL start and serve their APIs without requiring a running MySQL, MongoDB, or Redis server.

#### Scenario: Fresh machine without those servers
- **WHEN** an operator starts Morph AI, Morph Data APIs, ComposerX, SheetX/FormsX, and other covered Utils/AI backends with embedded data paths configured and MySQL/Mongo/Redis stopped or absent
- **THEN** each service SHALL open or create its embedded stores and SHALL NOT fail solely because MySQL, MongoDB, or Redis are unreachable

#### Scenario: Decommission is the default
- **WHEN** project documentation and default env examples are followed for local development
- **THEN** they SHALL NOT instruct operators to install or start MySQL, MongoDB, or Redis as required dependencies for Morph AI, Morph Data, or Morph Utils

### Requirement: Relational data in embedded SQL
Each covered service that previously used MySQL for relational tables SHALL persist those tables in an embedded SQL database file (Go-native embedded relational store) owned by that service or a clearly shared Morph data directory.

#### Scenario: ComposerX publish history without MySQL
- **WHEN** a client lists published pages through ComposerX after the storage swap
- **THEN** the API SHALL return results from the embedded SQL store and SHALL NOT call MySQL

#### Scenario: SheetX form metadata without MySQL
- **WHEN** a client lists forms through SheetX/FormsX after the storage swap
- **THEN** the API SHALL return results from the embedded SQL store and SHALL NOT call MySQL

### Requirement: Document payloads in Badger
Each covered service that previously stored large documents or collections in MongoDB SHALL persist equivalent document payloads in BadgerDB on disk with stable keys and get/upsert/delete semantics for those product APIs.

#### Scenario: Email or publish body round-trip via Badger
- **WHEN** a client saves a ComposerX saved-email or published-page body and later loads it
- **THEN** the service SHALL return the stored body from Badger without requiring MongoDB

#### Scenario: Survey or event documents without Mongo
- **WHEN** SheetX/FormsX stores survey templates, results, or event-info documents after the storage swap
- **THEN** those documents SHALL be readable through existing product APIs backed by Badger (or an equivalent Badger-backed repository)

### Requirement: Cache without Redis
Covered services SHALL use in-process cache with TTL semantics for any behavior formerly backed by Redis and SHALL NOT require Redis to boot or serve core APIs.

#### Scenario: Service runs with Redis stopped
- **WHEN** Redis is stopped or `TRAN_REDIS_ADDR` is unset
- **THEN** covered services SHALL continue serving APIs that do not depend on Redis, using in-process cache where caching is needed

### Requirement: Stable product HTTP contracts
Covered services SHALL keep existing HTTP request and response shapes for their primary CRUD and list endpoints unless a change is explicitly documented as breaking.

#### Scenario: Morph Utils embeds keep working
- **WHEN** Morph Utils embeds ComposerX or SheetX after the storage swap with a valid Morph AI session
- **THEN** successful list/create/update responses SHALL remain JSON-compatible with the pre-migration clients without a mandatory frontend rewrite
