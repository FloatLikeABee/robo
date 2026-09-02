# GraphRAG ops (optional)

Neo4j and `morphgraph-worker` are **not** required to run Morph AI, MorphNotes, MorphUtils, or the module apps. The default stack is SQLite + Badger. Skip this file unless you are enabling the graph.

`start-all.sh` may try `neo4j start` on full-stack boot when the CLI exists; other services still start if Neo4j is missing.

## Native Neo4j (no Docker)

```bash
brew install neo4j
neo4j start
# set password on first login; bolt://127.0.0.1:7687
```

Env (Morph + worker) in the **repo-root** `.env`:

```bash
MORPH_GRAPH_ENABLED=true
NEO4J_URI=neo4j://127.0.0.1:7687
NEO4J_USER=neo4j
# Optional locally: omit NEO4J_PASSWORD and disable Neo4j auth (see below).
# NEO4J_PASSWORD=...
NEO4J_DATABASE=neo4j
TRAN_MYSQL_DSN=...
TRAN_OPENAI_API_KEY=...   # embeddings (OpenAI-compatible)
MORPH_KNOWLEDGE_DIR=./data/knowledge
```

`TRAN_MYSQL_DSN` is a **worker** requirement (the Go worker still uses the MySQL driver). It is **not** required for Morph, Event Logs, or Content Maker in default local mode.

### Local Neo4j with no password

Neo4j enables auth by default. For local-only no-password:

1. Edit Neo4j config (Homebrew often: `$(brew --prefix)/etc/neo4j/neo4j.conf` or `~/.neo4j/...`):

```conf
dbms.security.auth_enabled=false
```

2. Restart Neo4j (`neo4j restart`).

3. Leave `NEO4J_PASSWORD` unset — the worker uses `NoAuth`.

## Worker

The worker still expects a SQL DSN (`TRAN_MYSQL_DSN` or `DATABASE_URL`). Historical Morph SQL migrations (for example `morph/migrations/045_graph_knowledge.sql`) targeted MySQL; do not treat that as the default Morph store (MorphNotes uses SQLite).

```bash
cd morphgraph-worker
go build -o morphgraph-worker .
./morphgraph-worker bootstrap-schema
./morphgraph-worker backfill --all
./morphgraph-worker run          # outbox daemon
./morphgraph-worker sync --mode=daily
./morphgraph-worker status
```

## Daily cron (03:00)

```cron
0 3 * * * cd /opt/robo/morphgraph-worker && ./morphgraph-worker sync --mode=daily >>/var/log/robo/morphgraph-daily.log 2>&1
```

## What syncs (when enabled)

| Source | Entities |
|--------|----------|
| Morph | MorphNotes entities + knowledge files |
| Event Logs (`formx`) | forms (+ outbox on create/update/delete) |
| Content Maker | published/template outbox when present |
| Knowledge | Morph knowledge chunks → Neo4j `Chunk` nodes when `MORPH_GRAPH_ENABLED=true` |

Dated design docs live under [`docs/archive/`](./archive/README.md), not on the operator path.
