# 00 — Architecture Overview

## What is robo?

**robo** is a monorepo of Morph platform apps that share:

- Central authentication hosted by **Morph** (JWT + bcrypt in SQLite `plat_users`)
- Shared AI clients (`pkg/morphai` for Go, `pkg/morphai-rs` for Rust)
- **Morph AI** as the system chat (no shared `platform-chat` drawer)
- **MorphUtils** (`morph-utils/`) — iframe shell for Event Logs, Content Maker, Data Access, and Project
- One repo-root `.env`

Booki, Academi, and a standalone UsersPanel app are **not** in the supported stack (`start-all.sh` does not launch them).

## App map

```
                         Morph AI  :3031 / API :9090
                         login, chat, MorphNotes
                         Tasks, Timelines, Big notes,
                         Research, Generic data
                                   │
                    JWT cookie  userspanel_session_token
                                   │
        ┌──────────────┬───────────┼──────────────┬──────────────┐
        ▼              ▼           ▼              ▼              ▼
   MorphUtils      Invite       AI tools       optional
   :3040           Signup       bk             Neo4j :7687
   iframe shell    :3051        :3000/:8000
        │           UI only
        │           /api → :9090
        ├── Event Logs     formx      UI :19909  API :29909
        ├── Content Maker  composerx  UI :8044   API :8043
        ├── Data Access    SharpReport UI :5178  API SHARPREPORT_PORT
        └── Project        morph-engi UI :5179  API :9096
```

## Tech stack summary

| Product (folder) | Backend | Frontend | Local data | Auth |
|------------------|---------|----------|------------|------|
| Morph AI / MorphNotes (`morph/`) | Go (Gin) | React (CRA) | SQLite + Badger | Morph JWT (auth hub) |
| Invite Signup (`invite-signup/`) | — (Morph API) | React (Vite) | — | Admin uses Morph JWT; redeem is open |
| Event Logs (`formx/`) | Go (Gin) | React (Vite) | SQLite + Badger | Morph SSO |
| Content Maker (`composerx/`) | Go (Gin) | Svelte (Vite) | SQLite + Badger | Morph SSO |
| Data Access (`SharpReport/`) | Rust (Axum) | SvelteKit | SQLite | Morph SSO |
| Project (`morph-engi/`) | Rust (Axum) | Svelte (Vite) | SQLite | Morph SSO |
| MorphUtils (`morph-utils/`) | — | React (Vite) | — | Morph cookie |
| AI tools (`bk/`) | Python (FastAPI) | React (CRA) | Chroma (local) | Morph token via Morph AI |

**Leftover package names:** Event Logs and Content Maker Go packages may still be named `mysql` / `mongo` while `main` opens **SQLite** and **Badger**. Do not install MySQL, MongoDB, or Redis for the default stack.

## Port map

| Service | API | UI |
|---------|-----|-----|
| Morph | 9090 | 3031 |
| MorphUtils | — | 3040 |
| Invite Signup | — | 3051 |
| Event Logs | 29909 | 19909 |
| Content Maker | 8043 | 8044 |
| Project | 9096 | 5179 |
| Data Access | `SHARPREPORT_PORT` (often 3050) | 5178 |
| AI tools | 8000 | 3000 |

## User-facing names vs ids

| Say this | Folder / URL id (do not “fix”) |
|----------|--------------------------------|
| Morph AI, MorphNotes (Tasks, Timelines, Big notes, Research, Generic data) | `morph/`, UI `/morphdata` |
| Invite Signup | `invite-signup/`, launcher `invite-signup-ui`, UI :3051 |
| MorphUtils | `morph-utils/` |
| Event Logs | `formx/`, embed id `sheetx`, default embed `/events-info` |
| Content Maker | `composerx/` |
| Data Access | `SharpReport/`, iframe `/datax` |
| Project | `morph-engi/` |
| AI tools | `bk/` |

Auth env is still `USERS_PANEL_BASE_URL` → Morph `:9090`. Cookie is still `userspanel_session_token`.

## Shared libraries

`pkg/morphai`, `pkg/morphai-rs`, `pkg/repoenv`, `pkg/assistmd`, `pkg/docextract`, `pkg/morphgraph`, `pkg/webresearch`. See `10-shared-libraries.md`.

## Optional GraphRAG

Neo4j and `morphgraph-worker` are optional. The rest of the stack starts without them. See [`docs/MORPH_GRAPH_OPS.md`](../MORPH_GRAPH_OPS.md).

## Next

- Auth: `01-auth-flow.md`
- AI: `02-ai-integration.md`
- Conventions: `13-conventions.md`
- Run: `12-build-deploy.md`
- Local MCP (stdio): `14-morph-mcp.md`
