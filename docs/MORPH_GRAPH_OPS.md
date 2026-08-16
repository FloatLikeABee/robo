# GraphRAG ops (Neo4j + daily sync)

## Native Neo4j (no Docker)

```bash
brew install neo4j
neo4j start
# set password on first login; bolt://127.0.0.1:7687
```

Env (Morph + worker):

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

### Local Neo4j with no password

Neo4j enables auth by default. For local-only no-password:

1. Edit Neo4j config (Homebrew often: `$(brew --prefix)/etc/neo4j/neo4j.conf` or `~/.neo4j/...`):

```conf
dbms.security.auth_enabled=false
```

2. Restart Neo4j (`neo4j restart`).

3. Leave `NEO4J_PASSWORD` unset — the worker uses `NoAuth`.

## Worker

Apply Morph MySQL migration first (creates `graph_sync_outbox` + knowledge tables):

```bash
mysql -u root -p -D tran < morph/migrations/045_graph_knowledge.sql
```

```bash
cd morphgraph-worker
go build -o morphgraph-worker .
./morphgraph-worker bootstrap-schema
./morphgraph-worker backfill --all
./morphgraph-worker run          # outbox daemon
./morphgraph-worker sync --mode=daily
./morphgraph-worker status
```

## Daily cron / launchd (03:00)

```cron
0 3 * * * cd /opt/robo/morphgraph-worker && ./morphgraph-worker sync --mode=daily >>/var/log/robo/morphgraph-daily.log 2>&1
```

macOS launchd example: `deploy/alibaba/systemd` pattern — or a local LaunchAgent calling the same command.

## What syncs

| Source | Entities |
|--------|----------|
| Morph | districts, facilities, members, employees, contacts, assets, activities, case tasks + knowledge files |
| FormsX | forms (+ outbox on create/update/delete) |
| ComposerX | email_templates (+ outbox on create/update) |
| Knowledge | Morph Knowledge Library chunks → Neo4j `Chunk` nodes when enabled |

Knowledge Library uploads (Morph UI → Context drawer → **Knowledge Library**) always index into MySQL chunks; Neo4j receives them via outbox/worker when `MORPH_GRAPH_ENABLED=true`.
