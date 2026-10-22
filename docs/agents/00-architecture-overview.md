# 00 — Architecture Overview

## What is robo?

**robo** is a monorepo of Dorph platform apps that share:

- Central authentication hosted by **Dorph** (JWT + bcrypt in SQLite `plat_users`)
- Shared AI clients (`pkg/dorphai` for Go, `pkg/dorphai-rs` for Rust)
- **Dorph AI** as the system chat (no shared `platform-chat` drawer)
- **DorphUtils** (`dorph-utils/`) — iframe shell for Event Logs, Content Maker, Data Access, and Project
- One repo-root `.env`

Booki, Academi, and a standalone UsersPanel app are **not** in the supported stack (`start-all.sh` does not launch them).

## App map

```
                         Dorph AI  :3031 / API :9090
                         login, chat, DorphNotes
                         Tasks, Timelines, Big notes,
                         Research, Generic data
                                   │
                    JWT cookie  userspanel_session_token
                                   │
        ┌──────────────┬───────────┼──────────────┬──────────────┐
        ▼              ▼           ▼              ▼              ▼
   DorphUtils      Invite       AI tools       optional
   :3040           Signup       bk             Neo4j :7687
   iframe shell    :3051        :3000/:8000
        │           UI only
        │           /api → :9090
        ├── Event Logs     formx      UI :19909  API :29909
        ├── Content Maker  composerx  UI :8044   API :8043
        ├── Data Access    SharpReport UI :5178  API SHARPREPORT_PORT
        └── Project        dorph-engi UI :5179  API :9096
```

## Tech stack summary

| Product (folder) | Backend | Frontend | Local data | Auth |
|------------------|---------|----------|------------|------|
| Dorph AI / DorphNotes (`dorph/`) | Go (Gin) | React (CRA) | SQLite + Badger | Dorph JWT (auth hub) |
| Invite Signup (`invite-signup/`) | — (Dorph API) | React (Vite) | — | Admin uses Dorph JWT; redeem is open |
| Event Logs (`formx/`) | Go (Gin) | React (Vite) | SQLite + Badger | Dorph SSO |
| Content Maker (`composerx/`) | Go (Gin) | Svelte (Vite) | SQLite + Badger | Dorph SSO |
| Data Access (`SharpReport/`) | Rust (Axum) | SvelteKit | SQLite | Dorph SSO |
| Project (`dorph-engi/`) | Rust (Axum) | Svelte (Vite) | SQLite | Dorph SSO |
| DorphUtils (`dorph-utils/`) | — | React (Vite) | — | Dorph cookie |
| AI tools (`bk/`) | Python (FastAPI) | React (CRA) | Chroma (local) | Dorph token via Dorph AI |

**Leftover package names:** Event Logs and Content Maker Go packages may still be named `mysql` / `mongo` while `main` opens **SQLite** and **Badger**. Do not install MySQL, MongoDB, or Redis for the default stack.

## Port map

| Service | API | UI |
|---------|-----|-----|
| Dorph | 9090 | 3031 |
| DorphUtils | — | 3040 |
| Invite Signup | — | 3051 |
| Event Logs | 29909 | 19909 |
| Content Maker | 8043 | 8044 |
| Project | 9096 | 5179 |
| Data Access | `SHARPREPORT_PORT` (often 3050) | 5178 |
| AI tools | 8000 | 3000 |

## User-facing names vs ids

| Say this | Folder / URL id (do not “fix”) |
|----------|--------------------------------|
| Dorph AI, DorphNotes (Tasks, Timelines, Big notes, Research, Generic data) | `dorph/`, UI `/dorphdata` |
| Invite Signup | `invite-signup/`, launcher `invite-signup-ui`, UI :3051 |
| DorphUtils | `dorph-utils/` |
| Event Logs | `formx/`, embed id `sheetx`, default embed `/events-info` |
| Content Maker | `composerx/` |
| Data Access | `SharpReport/`, iframe `/datax` |
| Project | `dorph-engi/` |
| AI tools | `bk/` |

Auth env is still `USERS_PANEL_BASE_URL` → Dorph `:9090`. Cookie is still `userspanel_session_token`.

## Shared libraries

`pkg/dorphai`, `pkg/dorphai-rs`, `pkg/repoenv`, `pkg/assistmd`, `pkg/docextract`, `pkg/dorphgraph`, `pkg/webresearch`. See `10-shared-libraries.md`.

## Optional GraphRAG

Neo4j and `dorphgraph-worker` are optional. The rest of the stack starts without them. See [`docs/DORPH_GRAPH_OPS.md`](../DORPH_GRAPH_OPS.md).

## Next

- Auth: `01-auth-flow.md`
- AI: `02-ai-integration.md`
- Conventions: `13-conventions.md`
- Run: `12-build-deploy.md`
- Local MCP (stdio): `14-dorph-mcp.md`
- Agent notes (self-use): `15-dorph-notes-self-use.md`
- Dorph for agents (external playbook): `16-dorph-for-agents.md`
