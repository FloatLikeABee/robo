# 00 — Architecture Overview

## What is robo?

**robo** is a monorepo containing 8+ independent platform applications that share:
- A central auth service (UsersPanel) — except **morph**, which is self-hosted
- A shared AI client library (`pkg/morphai` for Go, `pkg/morphai-rs` for Rust)
- A shared chat drawer UI component (`platform-chat/`, published as `@robo/platform-chat`)
- A unified shell (`morph-utils/`) that embeds all apps in iframes
- A common AI assistant contract (`AI_ASSISTANT_MORPHAI_CONTRACT.md`)

## App map

```
┌──────────────────────────────────────────────────────────────────┐
│                        robo MONOREPO                             │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │  morph   │  │  formx   │  │composerx │  │  booki   │        │
│  │ Go+React │  │ Go+React │  │Go+Svelte │  │ Go+React │        │
│  │ :9090    │  │ :29909   │  │ :8043    │  │ :9095    │        │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘        │
│       │              │              │              │              │
│  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐        │
│  │morph-engi│  │UsersPanel│  │SharpRpt  │  │ academi  │        │
│  │Rust+Svel │  │Rust+Svel │  │Rust+SvK  │  │ Go+RN    │        │
│  │ :9096    │  │ :5001    │  │ :3050    │  │ :8978    │        │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘        │
│       │              │              │              │              │
│       └──────────────┼──────────────┘              │              │
│                      │                             │              │
│              ┌───────┴───────┐                     │              │
│              │   SHARED      │                     │              │
│              │  pkg/morphai  │ (Go AI client)      │              │
│              │ pkg/morphai-rs│ (Rust AI client)    │              │
│              │ platform-chat │ (TS chat drawer)    │              │
│              │ pkg/assistmd  │ (MD formatting)     │              │
│              │ pkg/morphgraph│ (GraphRAG types)    │              │
│              │ pkg/webresearch│ (web search)       │              │
│              └───────────────┘                     │              │
│                                                    │              │
│              ┌────────────────────────────────────┴───┐          │
│              │           INFRASTRUCTURE               │          │
│              │  MySQL · MongoDB · Redis · BadgerDB    │          │
│              │  SQLite · Neo4j (optional GraphRAG)    │          │
│              │  DashScope (Qwen) AI provider          │          │
│              └────────────────────────────────────────┘          │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  morph-utils/  —  Unified shell (React, :3040)           │   │
│  │  Embeds all apps in iframes with shared JWT cookie        │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  bk/  —  Ground Control RAG workspace (Python, :8000)     │   │
│  │  Linked from Morph AI header; UI on :3000                 │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

## Tech stack summary

| App | Backend | Frontend | Primary DB | Auth |
|-----|---------|----------|------------|------|
| **morph** | Go (Gin) | React (CRA) | BadgerDB + MySQL + Mongo + Redis | **Self-hosted** JWT+bcrypt |
| **formx** | Go (Gin, GORM) | React (Vite) | MySQL + MongoDB | UsersPanel |
| **composerx** | Go (Gin) | Svelte (Vite) | MySQL + MongoDB + Redis | UsersPanel |
| **booki** | Go (Gin) | React (Vite) | MySQL + Redis | UsersPanel SSO |
| **UsersPanel** | Rust (Axum) | Svelte 5 (Vite) | MySQL | Self-hosted (auth hub) |
| **SharpReport** | Rust (Axum) | SvelteKit | SQLite | UsersPanel SSO |
| **morph-engi** | Rust (Axum) | Svelte 5 (Vite) | SQLite | UsersPanel SSO |
| **academi** | Go (Gin) | React Native | BadgerDB | UsersPanel SSO + local fallback |
| **bk** | Python (FastAPI) | React (CRA) | ChromaDB | UsersPanel token (via Morph AI link) |

## Port map

| Service | API Port | UI Port |
|---------|----------|---------|
| UsersPanel | 5001 | 5173 |
| Morph | 9090 | 3031 |
| Morph Utils | — | 3040 |
| FormsX | 29909 | 19909 |
| ComposerX | 8043 | 8044 |
| Booki | 9095 | 5174 |
| Morph Engi | 9096 | 5179 |
| Academi | 8978 | 8765 |
| SharpReport | 3050 | 5178 |
| BK | 8000 | 3000 |

## Dependency matrix

| Capability | Apps that use it |
|------------|------------------|
| MySQL | morph, formx, composerx, booki, UsersPanel |
| MongoDB | morph, formx, composerx |
| Redis | morph, composerx, booki |
| BadgerDB (embedded) | morph, academi |
| SQLite | SharpReport, morph-engi |
| Neo4j (GraphRAG) | morph (optional, via morphgraph-worker) |
| ChromaDB | bk |

## Key documents

| File | Purpose |
|------|---------|
| `README.md` | Quick start, ports, service names |
| `DEVELOPER_BASELINE.md` | Architecture map, conventions (**note: morph auth section is outdated**) |
| `AI_ASSISTANT_MORPHAI_CONTRACT.md` | Cross-app assistant API contract |
| `AI_ASSISTANT_SESSIONS_API.md` | Session persistence spec |
| `DEPLOY-README.md` | Production deployment (Render, Alibaba) |

## How to use these agent docs

- **00-architecture-overview.md** (this file) — start here for orientation
- **01-auth-flow.md** — understand auth before touching any login/session code
- **02-ai-integration.md** — understand AI before touching any assistant/chat code
- **03–09** — per-app deep dives: read the one for the app you're modifying
- **10-shared-libraries.md** — reference for `pkg/*` usage
- **11-platform-chat.md** — reference for chat drawer integration
- **12-build-deploy.md** — reference for build, run, and deploy commands
- **13-conventions.md** — cross-cutting patterns to follow