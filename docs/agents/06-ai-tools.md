# 06 — DorphTools (`bk/`)

## Overview

User-facing product: **DorphTools**. Folder stays `bk/`. Dorph AI can open this workspace (Assistants, RAG, Documents, System). It is not a separate “Ground Control” product. It does not run a Model Context Protocol server. The platform MCP program is [`dorph/cmd/dorph-mcp`](14-dorph-mcp.md).

| | |
|--|--|
| Backend | `bk/` Python FastAPI |
| Frontend | `bk/frontend/` React (CRA) + MUI |
| Ports | API `8000`, UI `3000` |
| Auth | Dorph session (opened from Dorph AI) |
| Data | Local Chroma (and related files under the app cwd) |

## Nav (`bk/frontend/src/components/Header.js`)

| Path | Label |
|------|--------|
| `/assistants` | Assistants |
| `/rag` | RAG |
| `/documents` | Documents |
| `/status` | System |

## Run

Prefer the launcher:

```bash
./start-all.sh start bk
```

API docs: http://localhost:8000/docs  
UI: http://localhost:3000

Python 3.11+ (see root README). Config is the **repo-root** `.env` (nested `bk/.env` is ignored).

Dorph AI is the system chat; DorphTools is the RAG/assistants workbench linked from Dorph AI.

See [`bk/README.md`](../../bk/README.md).
