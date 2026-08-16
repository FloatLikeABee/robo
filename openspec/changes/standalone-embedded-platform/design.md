## Context

See proposal.md — Why. Morph already has SQLite + Badger + go-cache for Morph Data (`morph-embedded-dbs`), but ComposerX, FormsX/SheetX, Booki, and related AI tools still open MySQL/Mongo/Redis and fail when those servers are absent. Neo4j already exists for GraphRAG/AI; uploads are not consistently mirrored async. There is no first-class user-uploadable Morph AI skills catalog.

## Goals / Non-Goals

**Goals:**
- Stop and remove MySQL/Mongo/Redis as project runtime dependencies (first implementation step).
- One embedded stack across Morph AI / Morph Data / Utils / AI tools: SQLite (relational), Badger (documents), go-cache (ephemeral).
- Async Neo4j ingest for durable uploads and skills so AI stays current without blocking writes.
- Skills store with upload + list + assistant consumption.

**Non-Goals:**
- Replacing Neo4j.
- Distributed multi-writer SQLite/Badger clusters.
- Full redesign of every Utils frontend; keep HTTP shapes stable where possible.
- Perfect historical migration when operators already wiped MySQL/Mongo (fresh embeds OK).

## Decisions

### 1. Decommission networked DBs first
- **Choice:** Implementation starts by stopping/uninstalling Homebrew MySQL, MongoDB, and Redis used for this monorepo, clearing required DSNs from `.env` examples, and accepting that Utils will stay broken until embeds land—explicitly preferred over keeping temporary servers.
- **Why:** Matches the user’s primary intent and prevents accidental dependency on reinstalled servers.
- **Alternatives:** Keep servers until each app migrates (safer cutover, rejects “delete first” mandate).

### 2. Embedded technology choices (fixed)
- **Relational:** `modernc.org/sqlite` (CGO-free, already used by Morph).
- **Documents:** BadgerDB (already used by Morph entity details / graph workers patterns).
- **Cache:** `patrickmn/go-cache` (already used by Morph).
- **Alternatives:** Postgres embedded (heavier), LiteFS (ops complexity), Redis-compatible embed (unnecessary).

### 3. Per-app data directories, shared Morph auth
- **Choice:** Each service owns files under its `./data/` (or Morph-shared paths where tables are truly shared). Auth remains Morph JWT / shared cookie; Utils do not reintroduce login UIs.
- **Why:** Avoids one giant SQLite file locking across processes; Morph auth already centralized.
- **Alternatives:** Single shared SQLite for all apps (lock contention / coupling).

### 4. ComposerX / FormsX migration approach
- **Choice:** Replace `sql.Open("mysql")` / GORM MySQL with SQLite DSN; replace Mongo repositories with Badger-backed stores (JSON documents keyed like existing collections); make Redis optional then remove.
- **Why:** Fastest path to restore Utils without networked DBs while keeping handlers mostly intact.
- **Alternatives:** Proxy all Utils data through Morph `/api/tran` (large product coupling).

### 5. Neo4j async ingest via outbox
- **Choice:** Primary write commits to SQLite/Badger, then inserts an outbox row (or Badger queue record); a worker drains to Neo4j with retries. HTTP returns after primary write.
- **Why:** Non-blocking UX; works when Neo4j is down; fits “everything uploaded async to Neo4j.”
- **Alternatives:** Sync Neo4j in request path (fragile); best-effort fire-and-forget without durable outbox (lossy).

### 6. Skills store shape
- **Choice:** Skills live as Badger documents (definition/body) + SQLite index (id, name, description, enabled, owner, timestamps). Upload accepts structured skill packages (markdown/JSON/SKILL.md zip as decided in tasks). Morph AI lists enabled skills and loads by id; uploads enqueue Neo4j ingest.
- **Why:** Matches embedded stack; indexable list without scanning Badger only.
- **Alternatives:** Skills only on disk under `.cursor/skills` (not multi-user / not API-managed).

### 7. Morph leftover cleanup
- **Choice:** After Utils/tools migrate, strip Morph `go.mod` leftover mysql/mongo/redis deps and env defaults that still advertise networked DSNs.
- **Why:** Morph is mostly embedded but still carries legacy client deps/env noise.

## Risks / Trade-offs

- **[Risk] Utils downtime while embeds are built after DB deletion** → Mitigate with clear task order: decommission → Morph verify → ComposerX → FormsX → remaining tools → Neo4j outbox → skills; communicate empty fresh DBs if migration sources gone.
- **[Risk] SQLite locking under concurrent Utils + Morph** → Mitigate with per-app DB files and busy timeout; avoid cross-process sharing of one file.
- **[Risk] Badger key design mismatches Mongo query patterns** → Mitigate by implementing repository methods used by handlers first; add secondary indexes in SQLite where list/filter queries need them.
- **[Risk] Neo4j outbox backlog** → Mitigate with retries, metrics/logs, and admin visibility; do not block uploads.
- **[Risk] Skill package format ambiguity** → Mitigate by starting with a simple validated schema (name, description, instructions/body, optional tools hints) and documenting upgrade path for zip/SKILL.md later if needed.

## Migration Plan

1. **Stop/uninstall** local MySQL, MongoDB, Redis for this project; update env examples so they are not required.
2. **Verify Morph** already boots on SQLite + Badger + go-cache; finish leftover dependency cleanup.
3. **Migrate ComposerX** then **FormsX/SheetX** to SQLite + Badger (+ go-cache); smoke Morph Utils embeds.
4. **Migrate remaining** Booki/other AI tools still on networked DBs.
5. **Add Neo4j outbox worker** and wire upload/write paths.
6. **Ship skills store** APIs + Morph AI consumption + upload UX; mirror skills to Neo4j via outbox.
7. **Rollback:** Only possible by restoring previous binaries and reinstalling networked DBs from backup—assume operators accept embedded-only forward path once cut over.

## Open Questions

- Exact skill package format for v1 (JSON definition vs zip containing `SKILL.md`) — default to JSON/markdown definition unless apply-time product preference says otherwise.
- Whether DataX/SharpReport needs any MySQL/Mongo paths beyond its own store (confirm during apply inventory).
