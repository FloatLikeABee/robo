## 1. Inventory and scaffolding

- [x] 1.1 Inventory Morph-owned MySQL tables (from migrations + runtime `ensure*` DDL) and Mongo `entity_details` usage; list FK-safe copy order
- [x] 1.2 Add SQLite dependency (`modernc.org/sqlite`) and config keys for `TRAN_SQLITE_PATH` (default `./data/tran.sqlite`) and entity-details Badger path (default `./data/entity_details`)
- [x] 1.3 Introduce `TranSQL` (or rename) open/ping/close on SQLite with busy timeout; keep a short dual-backend switch (`embedded` vs legacy MySQL) for safe cutover
- [x] 1.4 Define `EntityDetailStore` interface matching current Mongo get/set/delete; implement Badger-backed store with `entity_detail:{entity}:{id}` keys in the dedicated Badger path
- [x] 1.5 Wire `main.go` to prefer embedded stores when SQLite path is set / legacy DSNs empty; keep go-cache; stop requiring Redis

## 2. SQLite schema and SQL dialect

- [x] 2.1 Port Morph relational schema to SQLite `CREATE TABLE IF NOT EXISTS` (including `plat_users` and tables created in `ensureTranMySQLSchema`)
- [x] 2.2 Replace MySQL-only boot probes (`information_schema`, backtick DDL) with SQLite-safe ensure/migrate helpers
- [x] 2.3 Audit and fix handler SQL that breaks on SQLite (AUTO_INCREMENT, backticks, LIMIT quirks, JSON helpers) starting with auth + highest-traffic Tran entities
- [x] 2.4 Update `plat_users` / bootstrap admin to run on SQLite

## 3. Handler and CLI cutover

- [x] 3.1 Point handlers at `EntityDetailStore` instead of `TranMongo` concrete type; preserve staff/employee mirror delete/upsert behavior
- [x] 3.2 Remove Redis from `Handlers` / `New` once unused; route any future cache needs through `morph/cache`
- [x] 3.3 Update `cmd/seed_tran`, backfill_*, and `prune_empty_detail` to use SQLite + Badger detail store
- [x] 3.4 Update Morph Data-facing error paths so Tran APIs no longer 503 solely for “MySQL/Mongo not configured” when embedded stores are open

## 4. Migration tooling

- [x] 4.1 Add `cmd/migrate_embedded` with `--dry-run`, `--apply`, `--verify`, and refuse non-empty destination unless `--force`
- [x] 4.2 Implement MySQL → SQLite table copy in FK-safe order with per-table counts and non-zero exit on fatal errors
- [x] 4.3 Implement Mongo `entity_details` → Badger copy preserving JSON bodies and empty-detail semantics
- [x] 4.4 Implement verify: compare relational row counts and sample/detail key presence; never drop/truncate sources

## 5. Config, docs, and dependency cleanup

- [x] 5.1 Update `.env.example`, README, and any Morph start docs for embedded defaults and cutover checklist (backup → migrate → verify → switch → smoke → decommission)
- [x] 5.2 After embedded path is default and verified, remove Morph MySQL/Mongo/Redis client code and `go.mod` deps (`go-sql-driver/mysql`, `mongo-driver`, `go-redis`)
- [x] 5.3 Remove dual-backend switch once cutover is complete so Morph is embedded-only

## 6. Validation

- [x] 6.1 Dry-run + apply + verify migration against a real (or restored) MySQL/Mongo snapshot; spot-check entity detail JSON in API/UI
- [x] 6.2 Smoke Morph boot with no MySQL/Mongo/Redis: login, list/create/update/delete key Tran entities, detail panel save/load, seed path
- [x] 6.3 Confirm FormX/Booki/ComposerX are unchanged and documented as follow-on (out of scope)
