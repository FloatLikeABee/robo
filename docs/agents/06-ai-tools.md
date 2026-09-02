# 06 — AI tools (`bk/`)

## Overview

User-facing product: **AI tools**. Folder stays `bk/`. Morph AI can open this workspace (Assistants, RAG, Documents, System). It is not a separate “Ground Control” product.

| | |
|--|--|
| Backend | `bk/` Python FastAPI |
| Frontend | `bk/frontend/` React (CRA) + MUI |
| Ports | API `8000`, UI `3000` |
| Auth | Morph session (opened from Morph AI) |
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

Morph AI is the system chat; AI tools is the RAG/assistants workbench linked from Morph AI.

See [`bk/README.md`](../../bk/README.md).
