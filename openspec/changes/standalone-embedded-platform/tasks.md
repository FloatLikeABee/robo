## 1. Decommission MySQL, MongoDB, and Redis

- [x] 1.1 Stop and uninstall (or disable) local Homebrew MySQL, MongoDB, and Redis used for this monorepo; confirm ports 3306 / 27017 / 6379 are not required for Morph boot
- [x] 1.2 Remove required `TRAN_MYSQL_DSN` / `TRAN_MONGO_*` / `TRAN_REDIS_ADDR` from Morph, ComposerX, FormsX, and other covered `.env.example` files; document embedded paths instead
- [x] 1.3 Update `start-all.sh` / README notes so default startup does not start or depend on MySQL, MongoDB, or Redis

## 2. Morph embedded cleanup

- [x] 2.1 Verify Morph AI / Morph Data boot and smoke on SQLite + Badger + go-cache with networked DBs stopped
- [x] 2.2 Remove leftover Morph MySQL/Mongo/Redis client usage and `go.mod` deps once unused; drop dual legacy DSN defaults from Morph config docs

## 3. ComposerX (Content Maker) embedded storage

- [x] 3.1 Add ComposerX SQLite open/migrate (schema for templates, emails, publishes, drafts, contacts, users index) under ComposerX `./data/`
- [x] 3.2 Replace Mongo email/publish/reference document bodies with Badger-backed content store; keep repository APIs used by handlers
- [x] 3.3 Remove Redis hard-fail at boot (go-cache only if any cache keys remain); delete mysql/mongo/redis drivers when unused
- [x] 3.4 Smoke: ComposerX `/publishes/history`, `/emails`, reference docs with Morph AI JWT via Morph Utils embed

## 4. SheetX / FormsX embedded storage

- [x] 4.1 Switch FormsX GORM/MySQL relational tables to SQLite with AutoMigrate or explicit schema under FormsX `./data/`
- [x] 4.2 Replace Mongo collections (responses, events-info, AI docs, survey-bot templates/results) with Badger-backed repositories
- [x] 4.3 Remove MySQL/Mongo/Redis runtime requirements from FormsX config and `go.mod` when unused
- [x] 4.4 Smoke: SheetX survey-bot / forms list APIs with Morph AI session through Morph Utils

## 5. Remaining Utils / AI tools inventory and cutover

- [x] 5.1 Inventory Booki, DataX/SharpReport, Morph Engi, and other AI tools for MySQL/Mongo/Redis usage
- [x] 5.2 Migrate each remaining offender to SQLite + Badger + go-cache (or confirm already embedded / out of scope with rationale)
- [x] 5.3 End-to-end Morph Utils smoke: Survey Maker, Content Maker, Data Access, Project without MySQL/Mongo/Redis running

## 6. Neo4j async ingest

- [x] 6.1 Add durable outbox (SQLite table or Badger queue) for Neo4j ingest jobs after primary writes
- [x] 6.2 Implement worker with retry/backoff that upserts uploaded/durable content into Neo4j
- [x] 6.3 Wire upload and durable create/update paths (Morph knowledge uploads, ComposerX publishes, FormsX durable assets, skills) to enqueue outbox jobs without blocking HTTP on Neo4j
- [x] 6.4 Add operator-visible logging and/or status for failed/stuck ingest jobs

## 7. AI skills store

- [x] 7.1 Define v1 skill schema (id, name, description, body/instructions markdown or JSON, enabled, owner, timestamps) in SQLite index + Badger body
- [x] 7.2 Implement Morph skills APIs: list, get, upload, enable/disable, delete (auth via Morph session)
- [x] 7.3 Seed or ship a small set of built-in skills; ensure user uploads appear in the catalog
- [x] 7.4 Hook Morph AI assistants to load enabled skills from the store during runs
- [x] 7.5 Add Morph AI UI entry for browsing/uploading skills; enqueue Neo4j ingest on upload/update
- [x] 7.6 Reject invalid skill payloads with clear errors; verify delete removes skill from list and new runs

## 8. Docs and validation

- [x] 8.1 Document standalone boot (clone → build → run) with embedded paths and optional Neo4j only for AI graph
- [x] 8.2 Validate change specs/tasks (`openspec validate standalone-embedded-platform`) and record smoke checklist results
