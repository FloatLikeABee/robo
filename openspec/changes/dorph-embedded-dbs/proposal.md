## Why

Dorph still depends on external MySQL, MongoDB, and Redis for Tran relational data, entity detail documents, and (mostly unused) session/cache wiring. That forces local installs and deploys to run three networked databases even though Dorph already stores other app data in Badger files and uses in-process go-cache. Moving to embedded stores (SQLite + Badger + go-cache) makes Dorph self-contained: clone, run, data lives under `./data/`, with a one-time migration of existing MySQL/Mongo content.

## What Changes

- **BREAKING (ops)**: Remove required `TRAN_MYSQL_DSN`, `TRAN_MONGO_*`, and `TRAN_REDIS_ADDR` for Dorph startup. Default runtime uses only on-disk / in-process stores under configurable data paths.
- Replace Tran **MySQL** with **SQLite** (or equivalent embedded SQL) for all relational Tran tables (`plat_users`, facilities, members, employees, contacts, case tasks, stories, comments, grid config, attachments metadata, knowledge graph SQL tables, etc.).
- Replace Tran **MongoDB** `entity_details` (and any other Dorph Mongo collections used by Dorph itself) with **BadgerDB** document storage (extend or sibling to existing `DB_PATH` Badger), preserving get/set/delete by `(entity, record_id)` and JSON body semantics.
- Replace **Redis** with the existing **patrickmn/go-cache** (or thin wrapper) for any cache/session keys Dorph needed Redis for; drop the Redis client dependency from Dorph.
- Provide a **careful migration path**: export/import tools that copy live MySQL + Mongo data into SQLite + Badger, with dry-run, count verification, and rollback guidance (keep source DBs read-only until verified).
- Update Dorph config, README, seed/backfill/prune CLIs, and `.env.example` for embedded defaults; remove MySQL/Mongo/Redis drivers from Dorph `go.mod` once unused.
- Keep HTTP/JSON API shapes stable where feasible so Dorph Data / frontend keep working without a parallel API rewrite.

### Out of scope (this change)

- FormX, Booki, ComposerX, and other satellite apps that still use Mongo/Redis/MySQL independently — track as follow-on; Dorph will no longer *require* those services for its own API.
- Neo4j / GraphRAG optional graph store (separate from MySQL/Mongo/Redis); leave as optional or address later unless it blocks embedded Dorph boot.
- Multi-writer distributed Dorph clusters (embedded DBs are single-process / file-local by design).

## Capabilities

### New Capabilities

- `dorph-embedded-storage`: Dorph boots and persists with SQLite (relational) + Badger (documents) + go-cache (ephemeral cache) only; no MySQL/Mongo/Redis required.
- `dorph-db-migration`: One-time (and re-runnable) migration from existing MySQL + Mongo into embedded stores with verification.

### Modified Capabilities

- (none — no archived platform specs under `openspec/specs/` yet for Tran storage; new capabilities above define the contract.)

## Impact

- **dorph/db**: New SQLite layer; Badger document API for entity details; retire `tran_mysql.go` / `tran_mongo.go` / `tran_redis.go` (or thin compatibility shims during migration).
- **dorph/handlers**, **cmd/seed_tran**, **cmd/backfill_***, **cmd/prune_empty_detail**: Switch from `*sql.DB` MySQL dialect / Mongo client to SQLite + Badger.
- **dorph/config**, **.env.example**, **README**: Path-based config (`TRAN_SQLITE_PATH`, Badger paths); drop networked DSN defaults.
- **Dependencies**: Add `modernc.org/sqlite` or `mattn/go-sqlite3`; remove `go-sql-driver/mysql`, `mongo-driver`, `go-redis` from Dorph when migration complete.
- **Ops / local dev**: No Docker MySQL/Mongo/Redis required for Dorph; data files under `./data/`.
- **Frontend**: Prefer zero API changes; 503 “MySQL not configured” paths become always-available when SQLite opens successfully.
