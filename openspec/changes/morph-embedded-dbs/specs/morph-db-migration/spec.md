## Purpose

Defines a safe, verifiable one-time (and re-runnable) migration of Morph’s existing MySQL relational data and Mongo entity-detail documents into Morph’s embedded SQL and Badger stores before networked databases are decommissioned.

## ADDED Requirements

### Requirement: Relational MySQL to embedded SQL migration
Morph SHALL provide a migration command or documented procedure that copies Morph-owned tables from the existing Tran MySQL database into the embedded SQL file store.

#### Scenario: Dry-run reports counts without writing
- **WHEN** an operator runs migration in dry-run mode against reachable MySQL
- **THEN** the tool SHALL report per-table source row counts and planned destination actions without mutating the embedded SQL store

#### Scenario: Apply copies rows
- **WHEN** an operator runs migration apply with valid MySQL credentials and an empty or target embedded SQL path
- **THEN** destination tables SHALL contain the same row counts as source for migrated Morph-owned tables (within documented exclusions), and the tool SHALL exit non-zero on fatal copy errors

### Requirement: Mongo entity_details to Badger migration
Morph SHALL migrate Mongo `entity_details` documents into Badger keys that Morph’s document API uses for `(entity, record_id)` lookups, preserving JSON body content.

#### Scenario: Detail documents readable after migration
- **WHEN** migration completes successfully for Mongo entity details
- **THEN** loading an entity that had a non-empty Mongo body SHALL return that body via Morph’s detail get path backed by Badger

#### Scenario: Empty and missing details handled
- **WHEN** a source Mongo document is missing or has an empty body (including `{}` / whitespace-only as Morph already treats empty)
- **THEN** migration SHALL either skip writing or write an equivalent empty representation such that Morph’s empty-detail checks behave as they did pre-migration

### Requirement: Verification and safety
Migration SHALL support verification after apply and SHALL leave source MySQL/Mongo unmodified (read-only from the migrator’s perspective) so operators can roll back by pointing Morph at the old stores or restoring from backup until cutover is accepted.

#### Scenario: Verify compares source and destination
- **WHEN** an operator runs migration verify after apply
- **THEN** the tool SHALL compare row/document counts (and sample key presence for details) and SHALL report mismatches with a non-zero exit code when verification fails

#### Scenario: Sources remain intact
- **WHEN** migration apply finishes
- **THEN** MySQL tables and Mongo collections used as sources SHALL still be readable with their pre-migration data (migrator SHALL NOT drop or truncate sources)

### Requirement: Idempotent or restartable apply
Migration apply SHALL be safe to re-run after partial failure by either upserting deterministic keys/primary keys or documenting a clean wipe of the destination embedded stores before re-apply.

#### Scenario: Re-run after partial failure
- **WHEN** an operator re-runs apply after a failed mid-run migration using the documented recovery mode
- **THEN** the destination SHALL end in a consistent complete state without duplicate primary keys for relational tables

### Requirement: Cutover documentation
Morph documentation SHALL describe the ordered cutover: backup → migrate → verify → switch Morph config to embedded paths → smoke-test APIs → decommission MySQL/Mongo/Redis for Morph.

#### Scenario: Operator follows cutover checklist
- **WHEN** an operator completes the documented cutover steps
- **THEN** Morph SHALL run using only embedded paths and SHALL no longer require Morph’s MySQL/Mongo/Redis connection settings
