## Context

See proposal.md for motivation. Today Morph already uses:

- **Badger** (`DB_PATH`, default `./data/badger`) for forms, chat sessions, complaints, voice profiles, etc. (`morph/db/db.go`).
- **go-cache** (`morph/cache`) for AI/response caching.
- **MySQL** (`TranMySQL`) for all Tran relational tables plus `plat_users` auth.
- **MongoDB** (`TranMongo` + `entity_details` collection) for large JSON detail payloads.
- **Redis** (`TranRedis`) wired in `main`/handlers but **not actively used** by handlers today.

Hundreds of handler SQL statements use MySQL dialect (backticks, `AUTO_INCREMENT`, `information_schema`, JSON column quirks). Seed/backfill/prune CLIs dual-write MySQL + Mongo. Cutover must preserve API JSON and migrate live data carefully.

## Goals / Non-Goals

**Goals:**

- Single-process Morph with data under `./data/` (SQLite file + Badger directories + uploads).
- Preserve `/api/tran/*` and auth behavior with minimal frontend changes.
- Phased implementation: storage adapters → dialect/SQL port → migrate tool → cutover → delete networked drivers.
- Verified migration from existing MySQL + Mongo before dropping them.

**Non-Goals:**

- Migrating FormX / Booki / ComposerX stores in this change (follow-on).
- HA multi-node Morph writing the same SQLite/Badger files.
- Redesigning entity schemas or Morph Data UI.
- Replacing optional Neo4j GraphRAG in this change (unless it hard-blocks boot; keep optional).

## Decisions

### D1 — SQLite via `database/sql` + pure Go driver
- **Choice:** Use SQLite file store with `modernc.org/sqlite` (CGO-free) behind the existing `*sql.DB` usage pattern (`TranMySQL`-like wrapper renamed to `TranSQL` / `EmbeddedSQL`).
- **Why:** Handlers already speak `database/sql`; pure Go eases macOS/Linux builds; matches “lightweight / no deploy deps.”
- **Alternatives:** `mattn/go-sqlite3` (CGO), keep MySQL optional forever (rejects goal), LiteFS/Turso (overkill).

### D2 — Keep `*sql.DB` call sites; fix dialect incrementally
- **Choice:** Introduce a thin store type that returns `*sql.DB` opened on SQLite; port schema DDL to SQLite; systematically replace MySQL-only SQL (backticks → quotes or unquoted, `AUTO_INCREMENT` → `INTEGER PRIMARY KEY AUTOINCREMENT`, drop `information_schema` probes in favor of SQLite `pragma` / `CREATE TABLE IF NOT EXISTS`).
- **Why:** Rewriting every handler to an ORM is higher risk than dialect adaptation.
- **Alternatives:** Full repository layer rewrite (cleaner long-term, larger blast radius now).

### D3 — Entity details in Badger with explicit key namespace
- **Choice:** Store documents under keys like `entity_detail:{entity}:{id}` (JSON body bytes), implementing the same methods currently on `TranMongo` (`GetEntityDetailJSON`, `SetEntityDetailJSON`, `DeleteEntityDetail`) on a Badger-backed type. Prefer a **dedicated Badger path** (e.g. `./data/entity_details`) or a clear key prefix inside existing Badger to avoid colliding with chat/form keys.
- **Why:** User requested Badger for document DB; Morph already depends on Badger; API can stay identical so handlers keep calling `h.TranMongo`-shaped interface (rename to `EntityDetails` / `DocStore`).
- **Alternatives:** Stuff JSON into SQLite TEXT/JSON columns (simpler one-store, but user asked for Badger for documents); keep Mongo optional.

### D4 — Redis → go-cache only
- **Choice:** Remove Redis client; if any session/cache keys appear later, use `morph/cache` (TTL). Today Redis is unused in handlers — deletion is mostly config/wiring cleanup.
- **Why:** Matches user request; zero ops surface.
- **Alternatives:** Persist cache to Badger with TTL (unnecessary for current usage).

### D5 — Schema strategy for SQLite
- **Choice:** Maintain a consolidated SQLite schema (generated or hand-ported from Morph migrations + `ensureTranMySQLSchema` side effects) applied with `CREATE TABLE IF NOT EXISTS` / idempotent alters on boot. Do not run MySQL `.sql` migrations as-is.
- **Why:** Morph migrations are MySQL-specific; SQLite needs its own DDL set.
- **Alternatives:** Runtime SQL rewriter (fragile).

### D6 — Migration tool as Morph `cmd/`
- **Choice:** Add `cmd/migrate_embedded` (name flexible) that: connects to source MySQL + Mongo using existing env; opens target SQLite + Badger; supports `--dry-run`, `--apply`, `--verify`; copies tables in FK-safe order; upserts entity_details into Badger; never drops sources.
- **Why:** Same module boundaries as `seed_tran` / backfills; operators already use Morph CLIs.
- **Alternatives:** External scripts only (harder to share types/key conventions).

### D7 — Cutover flags / dual-read period (short)
- **Choice:** During implementation, allow temporary env to choose backend (`STORAGE_BACKEND=embedded|legacy`) **or** prefer embedded when SQLite path set and legacy DSNs empty. After verify, remove legacy code paths in a final task — avoid long dual-write forever.
- **Why:** Careful cutover without big-bang untested switch; still end at embedded-only.
- **Alternatives:** Big-bang only (riskier).

### D8 — Scope boundary for “everything in morph”
- **Choice:** This design’s implementation scope is the **`morph/` Go service** (and its CLIs/docs). Satellite Morph apps (FormX, Booki, ComposerX) remain on their DBs until separate changes.
- **Why:** Morph is the hard dependency for Morph Data; satellites are separate processes; bundling all would explode risk.
- **Alternatives:** Platform-wide in one change (rejected for carefulness).

## Risks / Trade-offs

- **[Risk] MySQL-specific SQL breaks on SQLite** → Mitigation: inventory SQL via tests/smoke; port helpers for common patterns; run seed + Morph Data smoke checklist per entity.
- **[Risk] Incomplete migration (missing tables/collections)** → Mitigation: dry-run inventory from `information_schema` / Mongo list; verify counts; checklist of Morph-owned tables.
- **[Risk] SQLite write concurrency under load** → Mitigation: acceptable for single Morph process; set busy timeout; serialize heavy writers if needed. Document single-writer assumption.
- **[Risk] Badger vs Mongo type quirks (int/string record ids)** → Mitigation: reuse `uniqueRecordIDs` lookup ideas when reading migrated keys; normalize ids to string or int consistently in new key format.
- **[Risk] Operators lose data by wiping destination and re-migrate wrong** → Mitigation: migrate tool refuses apply if destination non-empty unless `--force`; document backups first.
- **[Trade-off] No networked DB** → Simpler local/dev; not suited for multi-host shared DB without external sync (accepted).

## Migration Plan

1. **Inventory** live MySQL tables and Mongo `entity_details` counts on the current Morph env.
2. **Implement** SQLite open + schema ensure; Badger doc store; go-cache-only cache; feature flag or path-based backend selection.
3. **Port** handlers/CLIs off MySQL dialect and Mongo client interface.
4. **Build** migrate CLI: dry-run → apply → verify.
5. **Backup** MySQL dump + Mongo dump (or filesystem snapshots).
6. **Apply** migration into fresh `./data/tran.sqlite` + Badger details dir; **verify** counts and spot-check entities in UI.
7. **Cut over** Morph config to embedded paths; unset legacy DSNs; smoke auth + Morph Data modules.
8. **Decommission** Morph’s MySQL/Mongo/Redis deps and Docker services for Morph; keep dumps until confidence window ends.
9. **Rollback** (if needed): restore env DSNs to legacy, point Morph at previous binary/config; embedded files remain as artifacts.

## Open Questions

- Exact Badger layout: separate directory vs prefixed keys in existing `DB_PATH` — prefer separate `./data/entity_details` for clearer backups; confirm at implement time.
- Whether any Morph MySQL tables are owned by other apps sharing the same `tran` schema — migrate only Morph-owned; leave others untouched in source.
