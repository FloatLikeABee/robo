# DorphTools (`bk/`)

Python FastAPI + React workbench for **Assistants**, **RAG**, **Documents**, and **System**. Dorph AI can open this module. Folder stays `bk/`.

Agent notes: [`docs/agents/06-ai-tools.md`](../docs/agents/06-ai-tools.md).

## Ports

| | |
|--|--|
| API | http://localhost:8000/docs |
| UI | http://localhost:3000 |

## Run

From the repo root:

```bash
cp .env.example .env
./start-all.sh start bk
```

Manual:

```bash
# API (from bk/)
python -m uvicorn src.api:app --reload --host 0.0.0.0 --port 8000

# UI
cd frontend && npm start
```

Python **3.11+**, Node 18+. Nested `bk/.env` is ignored; use the **repo-root** `.env`.

Typical local keys (see `.env.example`): `API_PORT=8000`, Chroma persist directory, optional SMTP. Ollama is optional if you use local models; Dorph AI chat itself uses DashScope via Dorph.

## Nav

`/assistants`, `/rag`, `/documents`, `/status` (System).

Platform chat stays in Dorph AI (`:3031`). This app is the RAG/assistants workbench, not a second DorphUtils module.

AI tools does not speak the Model Context Protocol and does not listen on port 8196. The platform MCP server is [`dorph/cmd/dorph-mcp`](../docs/agents/14-dorph-mcp.md).
