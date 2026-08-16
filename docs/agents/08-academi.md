# 08 — Academi (Study Assistant)

## Overview

Academi is an AI-powered academic study platform with a Go backend, React Native mobile app, and static web frontend. It's unique in the platform for its **multi-provider AI** support (not just DashScope).

- **Module:** `github.com/academi/backend` (Go 1.24)
- **Backend:** `academi/backend/` — Go + Gin
- **Frontend:** React Native 0.72 (+ static web at `academi/web/`)
- **Ports:** API `8978`, Web UI `8765`
- **Auth:** UsersPanel SSO + local BadgerDB fallback

## Backend architecture

### Entry point: `academi/backend/cmd/main.go`

1. Load config from env
2. Open BadgerDB
3. Create Gin router with CORS
4. Register all route groups
5. Listen on `SERVER_PORT` (default `8978`)

### Package layout

```
academi/backend/
├── cmd/main.go                  ← Entry point
├── internal/
│   ├── config/config.go         ← Multi-provider AI config, BadgerDB path, JWT, CORS
│   ├── auth/                    ← JWT + UsersPanel bridge
│   │   ├── jwt.go               ← Local JWT
│   │   └── userspanel.go        ← UsersPanel SSO
│   ├── ai/                      ← Multi-provider AI
│   │   ├── chat.go              ← Chat completion
│   │   ├── learn.go             ← "Help You Learn" mode
│   │   ├── docparse.go          ← Document parsing
│   │   ├── prompts.go           ← Prompt templates
│   │   └── learning_guide.go    ← Learning guide generation
│   ├── community/               ← Community posts, voting, AI moderation
│   ├── docs/                    ← Document upload, AI indexing, semantic search
│   ├── chatsessions/            ← Chat session persistence
│   ├── notifications/           ← Push + in-app notifications
│   └── research/                ← Web research / RAG
├── go.mod
└── go.sum
```

### Route groups (all under `/api/v1`)

| Group | Key endpoints |
|-------|---------------|
| Auth | `POST /auth/register`, `POST /auth/login`, `POST /auth/dev-login`, `POST /auth/mock` |
| AI | `GET /ai/providers`, `POST /ai/chat`, `POST /ai/learn`, `POST /ai/summarize`, `POST /ai/generate-guide`, `POST /ai/moderate` |
| Community | `GET/POST /community/*` (JWT-protected) — posts, upvote/downvote, tags, AI moderation |
| Docs | `GET/POST /docs/*` — upload PDFs/images, AI indexing, semantic search, auto-summarization |
| Chat Sessions | `GET/POST /chatsessions/*` — session persistence |
| Notifications | `GET/POST /notifications/*` (JWT-protected) — push + in-app |

### Multi-provider AI

Academi supports **multiple AI providers** simultaneously, unlike other apps which only use DashScope:

| Provider | Config key |
|----------|-----------|
| OpenAI | `OPENAI_API_KEY` |
| Anthropic | `ANTHROPIC_API_KEY` |
| Azure | `AZURE_API_KEY`, `AZURE_ENDPOINT` |
| Ollama (local) | `OLLAMA_BASE_URL` |
| Qwen (DashScope) | `QWEN_API_KEY` |
| SiliconFlow | `SILICONFLOW_API_KEY` |

The `GET /ai/providers` endpoint returns which providers are configured. Users can select their preferred provider.

### Auth flow

1. **Primary:** UsersPanel SSO — same credentials as Morph/Booki
2. **Fallback:** Local BadgerDB auth if UsersPanel is not configured
3. Dev mode: `POST /auth/dev-login` and `POST /auth/mock` for testing

### Database

- **BadgerDB** (embedded): All data — users, chat sessions, community posts, documents, notifications

## Frontend (React Native)

### Build & tooling
- React Native 0.72 + React 18 + React Navigation 6
- State: Zustand v4
- Entry: `academi/App.js`

### Bottom tabs
1. **Chat** (Home) — AI chat with multi-provider support
2. **Community** — Posts, upvote/downvote, AI moderation
3. **Docs** — Upload PDFs/images, AI indexing, search
4. **Guide** — "Help You Learn" structured learning guides
5. **Profile** — User settings

### Web frontend
- Static files served from `academi/web/` on port 8765
- `python3 -m http.server 8765` in dev

## How to add a new AI provider

1. Add config fields in `academi/backend/internal/config/config.go`
2. Add provider client in `academi/backend/internal/ai/`
3. Register in the provider registry (returned by `GET /ai/providers`)
4. Update frontend provider selector

## Key env vars

| Variable | Default | Purpose |
|----------|---------|---------|
| `SERVER_PORT` | `8978` | Backend port |
| `DB_PATH` | `./data/badger` | BadgerDB path |
| `JWT_SECRET` | — | Local JWT secret |
| `USERS_PANEL_BASE_URL` | — | UsersPanel URL |
| `OPENAI_API_KEY` | — | OpenAI key |
| `ANTHROPIC_API_KEY` | — | Anthropic key |
| `QWEN_API_KEY` | — | DashScope key |
| `OLLAMA_BASE_URL` | — | Local Ollama URL |

## ⚠️ Note

Academi is **excluded** from `start-all.sh` default stack. Run it separately per `academi/README.md`.