# DorphAI Knowledge Graph + GraphRAG Milestone Plan

**Status:** Ready for implementation  
**Decisions locked:** Neo4j (native, no Docker) · DorphData + FormsX + ComposerX only  
**Core requirement:** MySQL/Mongo remain source of truth; data syncs into Neo4j so assistants retrieve faster via GraphRAG.

---

## 1. Decisions

| Decision | Choice |
|----------|--------|
| Graph DB | **Neo4j 5.x** with native vector indexes |
| Deploy | **Native install only** (Homebrew / Neo4j Desktop / tarball) — **no Docker** |
| V1 products | **DorphData**, **FormsX**, **ComposerX** |
| Source of truth | Existing **MySQL + MongoDB**; Neo4j is a derived AI index |
| Sync model | **Write-path hooks + outbox + bulk backfill** (no Kafka/CDC in repo) |
| Embeddings | Centralize ComposerX’s OpenAI-compatible pattern into shared Go package |
| Assistant integration | Shared GraphRAG tools; HybridContext stays session-only; ComposerX reference RAG migrates to graph |

**Out of scope:** Booki, DataX/SharpReport, UsersPanel, dorph-engi, Academi, Docker Neo4j, replacing MySQL/Mongo, Debezium CDC, multi-tenant partitioning.

---

## 2. Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│ Source of truth                                              │
│  MySQL `tran`  ·  Mongo `athena`  ·  Mongo `alterathena`     │
└───────────────┬─────────────────────────────┬───────────────┘
                │ writes                      │ backfill read
                ▼                             ▼
┌──────────────────────────┐     ┌────────────────────────────┐
│ Product handlers         │     │ dorphgraph-worker backfill │
│ Dorph / FormsX / CompX   │     └─────────────┬──────────────┘
│ → enqueue outbox         │                   │
└────────────┬─────────────┘                   │
             ▼                                 │
┌──────────────────────────┐                   │
│ MySQL graph_sync_outbox  │◄──────────────────┘
└────────────┬─────────────┘
             ▼
┌──────────────────────────┐
│ dorphgraph sync worker   │
│ upsert nodes/edges/chunks│
│ + embeddings             │
└────────────┬─────────────┘
             ▼
┌──────────────────────────┐
│ Neo4j                    │
│ Entity nodes · Rels      │
│ Chunk nodes + vector idx │
└────────────┬─────────────┘
             ▼
┌──────────────────────────┐
│ DorphAI assistants       │
│ graph_search / expand /  │
│ get  → grounded answers  │
└──────────────────────────┘
```

**Flow summary**

1. Apps keep writing MySQL/Mongo exactly as today.
2. After a successful write, enqueue a row in `graph_sync_outbox` (never fail the user request if Neo4j is down).
3. Worker claims outbox events, MERGEs graph state, chunks text, embeds, updates vector index.
4. Assistants call `graph_search` first for platform-data questions, then fall back to existing MCP/REST tools.

---

## 3. Ontology (V1)

Stable node key: `source:type:id`  
Examples: `dorph:member:42`, `formsx:form:7`, `composerx:template:3`.

Shared properties on all nodes: `uid`, `source`, `type`, `source_id`, `title`, `summary`, `updated_at`, `tenant` (default `"default"`).

### 3.1 DorphData (`dorph` / MySQL `tran` + Mongo `athena.entity_details`)

| Node label | Source | Key edges |
|------------|--------|-----------|
| `District` | MySQL `District` | — |
| `Facility` | MySQL `facility` | `IN_DISTRICT` → District |
| `Member` | MySQL `member` | `AT_FACILITY` → Facility |
| `Employee` | MySQL `employee` | `WORKS_AT` → Facility |
| `Contact` | MySQL `Contact` | — |
| `Asset` | MySQL `Asset` | — |
| `Activity` | MySQL `Activity` | `HAS_PARTICIPANT` → Member, `HAS_STAFF` → Employee, `USES_ASSET` → Asset |
| `CaseTask` | MySQL `CaseTask` | `ASSIGNED_TO` → Member/Employee |
| `StoryPost` | MySQL `StoryPost` | `ABOUT` → entity |
| `GenericData` | MySQL `generic_data` | — |
| `EntityDetail` | Mongo `entity_details` | `DETAIL_OF` → parent entity |
| `Chunk` | derived text | `DESCRIBES` → any entity |

`RecordContact` → `HAS_CONTACT` / `LINKED_TO` with `relationship` property.

Hook points: after CRUD + `attachEntityDetail` in [`dorph/handlers/`](../dorph/handlers/) (facilities, members, employees, activities, contacts, assets, case tasks, generic data, districts).

### 3.2 FormsX

| Node label | Source | Key edges |
|------------|--------|-----------|
| `Form` | MySQL `forms` | — |
| `FormPage` | MySQL `form_pages` | `PAGE_OF` → Form |
| `Question` | MySQL `questions` | `QUESTION_OF` → FormPage |
| `QuestionRule` | MySQL `question_rules` | `RULE_OF` → Question |
| `FormResponse` | Mongo `form_responses` | `RESPONSE_TO` → Form |
| `WorkspaceEvent` | Mongo `workspace_events` | `RELATED_TO` → Form (when known) |
| `AISystemDoc` | Mongo `ai_system_documents` | `DOC_FOR` → Form |
| `Chunk` | titles/summaries/bodies | `DESCRIBES` → above |

Hook points: form/page/question/rule repos; response create; event create; `syncSystemDocuments` also enqueues graph upserts.  
See [`formx/backend/internal/handler/assistant_llm.go`](../formx/backend/internal/handler/assistant_llm.go), [`mongodb_mcp.go`](../formx/backend/internal/handler/mongodb_mcp.go).

### 3.3 ComposerX

| Node label | Source | Key edges |
|------------|--------|-----------|
| `EmailTemplate` | MySQL `email_templates` | — |
| `SavedEmail` | MySQL `saved_emails` + Mongo `email_bodies` | body chunks via `DESCRIBES` |
| `PublishedPage` | MySQL `published_pages` | — |
| `MergeDataSource` | MySQL `merge_data_sources` | — |
| `ReferenceDocument` | Mongo `reference_documents` | migrate existing chunk embeddings into Neo4j `Chunk` |
| `Chunk` | HTML/markdown/reference text | `DESCRIBES` → docs |

Existing vector RAG lives in [`composerx/backend/reference_ai.go`](../composerx/backend/reference_ai.go) (`text-embedding-3-small`, cosine over Mongo chunks). Phase 3 lifts that into Neo4j.

---

## 4. New shared packages

### 4.1 `pkg/dorphgraph/` (Go) — primary deliverable

| Module | Responsibility |
|--------|----------------|
| `config.go` | `NEO4J_*`, embedding env, `DORPH_GRAPH_ENABLED` |
| `client.go` | Neo4j Go driver; health; constraint/index bootstrap |
| `schema.go` | Cypher constraints + vector index on `Chunk.embedding` |
| `upsert.go` | Idempotent MERGE for nodes/edges by `uid` |
| `chunker.go` | Text chunking (~900 runes / 120 overlap, match ComposerX) |
| `embed.go` | Shared embeddings client (lift from `reference_ai.go`) |
| `retrieve.go` | GraphRAG: vector → expand hops → rerank → context pack |
| `tools.go` | Canonical tool catalog JSON for assistants |
| `outbox.go` | Enqueue + claim/process helpers |
| `sync/*.go` | Per-product mappers: Dorph / FormsX / ComposerX |

### 4.2 Sync worker

- Path: `dorphgraph-worker/` (or `pkg/dorphgraph/cmd/worker`)
- Polls `graph_sync_outbox`, applies upserts/deletes, marks done/failed
- Subcommands: `run` (daemon), `backfill`, `bootstrap-schema`, `status`

### 4.3 DorphData HTTP gateway

Allowlisted under Dorph (same pattern as FormsX/ComposerX proxy in [`dorph/handlers/internal_api.go`](../dorph/handlers/internal_api.go)):

| Route | Purpose |
|-------|---------|
| `GET /api/graph/health` | Neo4j + outbox depth |
| `POST /api/graph/search` | Vector + structural GraphRAG |
| `POST /api/graph/expand` | Hop expansion from `uid`s |
| `GET /api/graph/entity/:uid` | Exact entity + key edges |
| `POST /api/graph/sync/backfill` | Admin trigger |
| `GET /api/graph/sync/status` | Lag / failures |

FormsX/ComposerX call `pkg/dorphgraph` **in-process** when possible; Dorph management chat can use HTTP tools like other APIs.

---

## 5. Sync design (MySQL/Mongo → Neo4j)

### 5.1 Outbox table (MySQL `tran`)

```sql
CREATE TABLE graph_sync_outbox (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source VARCHAR(32) NOT NULL,      -- dorph | formsx | composerx
  entity_type VARCHAR(64) NOT NULL,
  entity_id VARCHAR(128) NOT NULL,
  op ENUM('upsert','delete') NOT NULL,
  payload_json JSON NULL,
  created_at DATETIME(3) NOT NULL,
  available_at DATETIME(3) NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  locked_by VARCHAR(64) NULL,
  locked_at DATETIME(3) NULL,
  processed_at DATETIME(3) NULL,
  last_error TEXT NULL,
  KEY idx_claim (processed_at, available_at, id)
);
```

Migration: `dorph/migrations/NNN_graph_sync_outbox.sql`.

### 5.2 Write-path pattern

```go
// After successful MySQL/Mongo write — never fail the user request:
_ = dorphgraph.Enqueue(ctx, db, dorphgraph.Event{
    Source: "dorph", EntityType: "member", EntityID: strconv.Itoa(id), Op: "upsert",
})
```

### 5.3 Backfill (day-0 + rebuild)

```bash
dorphgraph-worker bootstrap-schema
dorphgraph-worker backfill --source=dorph
dorphgraph-worker backfill --source=formsx
dorphgraph-worker backfill --source=composerx
dorphgraph-worker backfill --all
```

**Order:** districts → facilities → people/assets → activities/links → FormsX forms tree → responses/events → ComposerX templates/emails/refs → chunk+embed pass.

### 5.4 Consistency rules

1. MySQL/Mongo always win; Neo4j may lag seconds.
2. Upserts are idempotent `MERGE` on `uid`.
3. Deletes remove node + connected `Chunk`s.
4. Failed outbox rows retry with backoff; dead-letter after N attempts.
5. Upsert entity first, then chunk/embed (same worker phase or follow-up op).
6. `DORPH_GRAPH_ENABLED=false` → enqueue no-ops / assistants skip graph tools.

---

## 6. GraphRAG retrieval

Implemented in `pkg/dorphgraph/retrieve.go`:

1. **Embed** user query (same model as chunks).
2. **Vector search** top-K `Chunk` nodes; optional `source` / `type` filters.
3. **Collect seed entity uids** from `DESCRIBES` edges.
4. **Expand** 1–2 hops for structural context.
5. **Rerank** by vector score + edge relevance + recency.
6. **Pack** compact context under ~8–12k runes (align with [`pkg/dorphai/context.go`](../pkg/dorphai/context.go)).

### Assistant tools (all three products)

| Tool | Args | Purpose |
|------|------|---------|
| `graph_search` | `query`, `sources?`, `types?`, `limit?` | Primary GraphRAG entry |
| `graph_expand` | `uids`, `hops?` | Relationship neighborhood |
| `graph_get` | `uid` | Exact entity + summary + key edges |

**Prompt rule:** prefer `graph_search` before broad list/search tools when the question is about existing platform data. Update `FastToolFirstInstructions` / product system prompts accordingly.

No arbitrary Cypher from end-user chat in V1.

---

## 7. Neo4j native setup (no Docker)

macOS developer path:

1. `brew install neo4j` (or Neo4j Desktop).
2. Set initial password; start: `neo4j start` → bolt `neo4j://127.0.0.1:7687`.
3. Env (Dorph, FormsX, ComposerX, worker):

```bash
NEO4J_URI=neo4j://127.0.0.1:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=...
NEO4J_DATABASE=neo4j
DORPH_GRAPH_ENABLED=true
DORPH_GRAPH_EMBEDDING_MODEL=text-embedding-3-small
# Reuse TRAN_OPENAI_API_KEY / OpenAI-compatible base URL for embeddings
```

4. Worker bootstrap creates uniqueness on `uid` and:

```cypher
CREATE VECTOR INDEX chunk_embedding IF NOT EXISTS
FOR (c:Chunk) ON (c.embedding)
OPTIONS {indexConfig: {
  `vector.dimensions`: 1536,
  `vector.similarity_function`: 'cosine'
}};
```

---

## 8. Implementation phases

### Phase 0 — Foundation

1. Create `pkg/dorphgraph` module + Neo4j Go driver dependency.
2. Config, client, health, schema bootstrap.
3. Outbox migration SQL under `dorph/migrations/`.
4. Worker stub: claim outbox, log, mark processed (no-op mapper).
5. Env examples in Dorph / FormsX / ComposerX `.env.example`.
6. Short ops section in this doc (install + env).

**Exit:** `dorphgraph-worker bootstrap-schema` + health check against local Neo4j.

### Phase 1 — DorphData sync + GraphRAG API

1. Mappers: District, Facility, Member, Employee, Contact, Asset, Activity (+ joins), CaseTask, EntityDetail.
2. Hooks on Dorph write handlers → outbox.
3. Backfill CLI for Dorph.
4. Chunk+embed for entity summaries (name, description, key fields + Mongo detail excerpt).
5. `retrieve.go` + Dorph `/api/graph/search|expand|entity|health`.
6. Wire Dorph management chat ([`management_chat.go`](../dorph/handlers/management_chat.go)) to prefer graph tools; allowlist in `internal_api.go`.
7. Keep HybridContext ([`dorph/hybridcontext/`](../dorph/hybridcontext/)) for **session uploads only**.

**Exit:** “Which members are at facility X?” answered via graph without scanning all REST list tools.

### Phase 2 — FormsX sync + assistant tools

1. Mappers: Form tree, response summaries, events, AISystemDoc.
2. Write hooks + backfill.
3. Tools in `assistant_llm.go` + MCP catalog (`mongodb_mcp.go` / mcp-tools).
4. `search_system_documents` remains; `graph_search` preferred for relational questions.
5. `sync_system_documents` also upserts Form/AISystemDoc nodes.

**Exit:** “Questions on form Z / events related to form Y” via graph.

### Phase 3 — ComposerX sync + unify reference RAG

1. Mappers: templates, saved emails (body chunks), published pages, reference docs.
2. Backfill embeddings into Neo4j from Mongo `reference_documents`.
3. Dual-read during transition; then cut over composer-chat retrieval to `dorphgraph.Retrieve`.
4. Assistant tools in platform assistant LLM.

**Exit:** Template/email and reference-doc questions both hit Neo4j chunks.

### Phase 4 — Hardening + milestone close

1. Metrics: outbox depth, oldest unprocessed, embed failures.
2. Admin: sync status, requeue failed, trigger backfill.
3. Auth: graph routes require UsersPanel JWT like other APIs.
4. Golden assistant tests (graph on/off).
5. Update [`AI_ASSISTANT_DORPHAI_CONTRACT.md`](../AI_ASSISTANT_DORPHAI_CONTRACT.md), [`DEVELOPER_BASELINE.md`](../DEVELOPER_BASELINE.md), [`.cursor/skills/dorphai-fast-ops/SKILL.md`](../.cursor/skills/dorphai-fast-ops/SKILL.md).
6. Confirm `DORPH_GRAPH_ENABLED=false` leaves apps fully functional.

**Exit:** Milestone acceptance checklist below all green.

---

## 9. Security and data hygiene

- Neo4j is an **AI index**, not a second CRUD store for UI.
- Chunk text: titles, summaries, non-sensitive fields; truncate Mongo detail bodies; light redaction for emails/phones in summaries where easy.
- No end-user Cypher in V1.
- Graph disabled cleanly when `DORPH_GRAPH_ENABLED=false`.

---

## 10. Testing

| Layer | What |
|-------|------|
| Unit | chunker, uid scheme, mapper fixtures, retrieve packing |
| Integration | Local Neo4j; outbox → upsert → search |
| Product | Dorph / FormsX / ComposerX golden questions with graph on/off |
| Regression | Existing tool loops still work when Neo4j is down |

---

## 11. Milestone acceptance criteria

1. Native Neo4j running; schema + vector index applied by worker bootstrap.
2. Full backfill of Dorph + FormsX + ComposerX completes successfully.
3. Live writes appear in Neo4j within outbox SLA (target &lt; 5s local).
4. All three assistants expose `graph_search` / `graph_expand` / `graph_get` and prefer them in prompts.
5. Dorph `/api/graph/health` and search return grounded entity context.
6. ComposerX reference questions work via Neo4j (or dual-read with documented cutover).
7. Docs + env examples + DorphAI skill updated; graph can be disabled without breaking apps.

---

## 12. Key existing code to extend

| Area | Path |
|------|------|
| Shared DorphAI Go | [`pkg/dorphai/`](../pkg/dorphai/) |
| Dorph chat / tools | [`dorph/handlers/management_chat.go`](../dorph/handlers/management_chat.go), [`internal_api.go`](../dorph/handlers/internal_api.go) |
| Dorph entity detail | [`dorph/handlers/entity_detail.go`](../dorph/handlers/entity_detail.go) |
| Dorph models | [`dorph/models/tran.go`](../dorph/models/tran.go) |
| HybridContext (keep session-only) | [`dorph/hybridcontext/store.go`](../dorph/hybridcontext/store.go) |
| FormsX assistant | [`formx/backend/internal/handler/assistant_llm.go`](../formx/backend/internal/handler/assistant_llm.go) |
| FormsX AI docs | [`formx/backend/internal/mongo/ai_document_repo.go`](../formx/backend/internal/mongo/ai_document_repo.go) |
| ComposerX RAG | [`composerx/backend/reference_ai.go`](../composerx/backend/reference_ai.go) |
| ComposerX assistant | [`composerx/backend/platform_assistant_llm.go`](../composerx/backend/platform_assistant_llm.go) |
| Contract | [`AI_ASSISTANT_DORPHAI_CONTRACT.md`](../AI_ASSISTANT_DORPHAI_CONTRACT.md) |

---

## 13. Suggested implementation order (checklist)

- [ ] Phase 0: `pkg/dorphgraph` + outbox + worker stub + Neo4j env
- [ ] Phase 1: Dorph sync + `/api/graph/*` + management chat tools
- [ ] Phase 2: FormsX sync + assistant/MCP tools
- [ ] Phase 3: ComposerX sync + reference RAG cutover
- [ ] Phase 4: metrics, admin, docs, golden tests, disable-flag verification

---

*Approve this plan to begin Phase 0 implementation.*
