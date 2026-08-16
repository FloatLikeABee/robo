# 05 — ComposerX (TranMail)

## Overview

ComposerX is an email composition and publishing platform with AI-powered drafting, RAG reference documents, web search, and public page publishing.

- **Module:** `mergeemailx-backend` (Go 1.25)
- **Backend:** `composerx/backend/` — Go + Gin
- **Frontend:** `composerx/frontend/` — Svelte 5 + Vite
- **Ports:** API `8043`, UI `8044`
- **Auth:** UsersPanel-dependent

## Backend architecture

### Entry point: `composerx/backend/main.go`

Single-file backend (~5000+ lines). The `main()` function:
1. Loads config from env vars + optional `ai.config.json`
2. Opens MySQL, MongoDB, Redis connections
3. Creates Gin router with CORS, logging, recovery, auth middleware
4. Creates `App` struct with all repos + AI client
5. Registers all routes via `app.registerRoutes()`
6. Listens on `PORT` (default `8043`)

### App struct

```go
type App struct {
    Templates   *TemplateRepository      // MySQL
    MergeData   *MergeDataRepository     // MySQL + local FS
    SavedEmails *SavedEmailRepository    // MySQL index + MongoDB body
    Published   *PublishedPageRepository // MySQL + MongoDB HTML
    Drafts      *PublishDraftRepository  // MySQL
    EmailStore  *EmailContentStore       // MongoDB email_bodies
    AI          *morphai.Client          // MorphAI
    Router      *gin.Engine
}
```

### Key files

| File | Purpose |
|------|---------|
| `main.go` | Entry point, App struct, all handlers, route registration, auth middleware |
| `models.go` | Data structs: User, EmailTemplate, MergeDataSource, ReportConfig, SavedEmailDetail |
| `repository.go` | TemplateRepository, MergeDataRepository (MySQL CRUD) |
| `saved_emails_repository.go` | SavedEmailRepository (MySQL index + MongoDB body) |
| `published_pages_repository.go` | PublishedPageRepository (MySQL + MongoDB HTML, slug generation) |
| `publish_drafts_repository.go` | PublishDraftRepository (MySQL-only drafts) |
| `email_content_mongo.go` | EmailContentStore (MongoDB `email_bodies` collection) |
| `ai_config.go` | AI config loading, `mergeAIConfigFromFile`, `mergeMorphAIEnv`, `chatCompletionClient` |
| `composer_draft.go` | HTML-to-markdown conversion for AI composer |
| `reference_ai.go` | Reference doc RAG: upload, chunk, embed (OpenAI), vector search |
| `web_research.go` | Web search via `pkg/webresearch`, MCP abilities |
| `platform_assistant.go` | Regex-based intent detection + field extraction |
| `platform_assistant_llm.go` | LLM tool-calling loop (list/get templates, saved emails, reference docs) |
| `publish_handlers.go` | Publish page CRUD, AI chat for HTML generation, public serving |
| `publish_source.go` | File processing: PDF text extraction, image description via vision model |

### Route groups

| Group | Key endpoints |
|-------|---------------|
| Health | `GET /health` |
| Auth | `POST /auth/login`, `GET /auth/me` (proxy to UsersPanel) |
| Templates | CRUD at `/templates` |
| Merge Data | List, upload, download at `/merge-data` |
| Saved Emails | CRUD at `/emails` (MySQL index + MongoDB markdown body) |
| Published Pages | CRUD at `/publishes`, public at `/public/p/:slug` (no auth) |
| Publish Drafts | CRUD at `/publish-drafts` |
| Reports | Stub endpoints at `/reports` |
| AI Composer | `POST /ai/composer-chat` (markdown generation with RAG + web search) |
| AI Publish | `POST /ai/publish-chat` (HTML page generation) |
| AI Reference | Upload/chunk/embed/search at `/ai/reference-docs` |
| AI Assistant | `POST /ai/assistant/chat` (intent + LLM tool loop) |
| AI Web Search | `POST /ai/web-search` |
| AI Tools | `GET /ai/app-abilities`, `GET /ai/mcp-tools` |
| AI Sources | `POST /ai/publish-sources/process` (PDF/text/image → structured materials) |

### Database layer

| DB | Purpose |
|----|---------|
| MySQL | Templates, merge data metadata, saved email index, published pages, publish drafts |
| MongoDB | Email bodies (`email_bodies` collection), reference documents (`reference_documents` collection) |
| Redis | Configured but minimal usage in current code |
| Local FS | File uploads: merge data files, reference docs, publish sources (path from `TRAN_FILE_STORAGE_PATH`) |

### AI features

1. **Composer chat** (`POST /ai/composer-chat`): Generates markdown email drafts. Integrates RAG (vector search over reference doc chunks) and web research. Returns `assistant_message` + `proposed_markdown`.

2. **Publish chat** (`POST /ai/publish-chat`): Generates HTML pages. Returns `proposed_page_html`.

3. **Reference RAG** (`reference_ai.go`): Upload documents (text/PDF/image) → chunk → embed (OpenAI `text-embedding-3-small`) → store in MongoDB → vector search at query time.

4. **Platform assistant** (`POST /ai/assistant/chat`): Two modes:
   - Regex-based for simple flows (create template)
   - LLM tool-calling loop for general queries (up to 8 rounds)

### Config loading priority

1. Environment variables (`TRAN_*`, `MORPH_AI_*`)
2. JSON config file (`ai.config.json`, path from `TRAN_AI_CONFIG_PATH`)
3. Hardcoded defaults (Qwen base URL, `qwen3-max` model)

## Frontend architecture

### Build & tooling
- Svelte 5 + Vite 7
- Port: 8044 (dev)
- Proxy: all API paths (`/auth`, `/health`, `/emails`, `/templates`, `/merge-data`, `/publishes`, `/publish-drafts`, `/public`, `/ai`, `/reports`) → `http://127.0.0.1:8043`
- `/ai` proxy has 600s timeout for long AI generation

### Key dependencies
- `svelte ^5.45.2`, `marked ^15.0.7`, `@robo/platform-chat`

### Source structure (`composerx/frontend/src/`)

| File/Dir | Purpose |
|----------|---------|
| `main.js` | Mounts App component |
| `App.svelte` | Main app (~3659 lines): 5 pages, auth, AI composer chat, autosave, theme, navigation |
| `components/` | `ComposePublishPanel.svelte`, `ComposePublishRecordsPanel.svelte`, `ContentMarkdownPanel.svelte`, `PlatformAssistantDrawer.svelte` |
| `lib/` | `composerDraft.js` (HTML→markdown), `contentMarkdown.js` (rendering), `TableFooterBar.svelte`, `ButtonLeadingIcon.svelte` |

### App.svelte pages
1. **Compose & Publish** — HTML editor + AI chat + file source processing
2. **Publish Records** — Published pages history
3. **Compose** — Markdown editor + right-rail AI chat + web search + reference docs
4. **Contents** — Saved emails list
5. **Reference Docs** — RAG document management

### Auth
JWT in localStorage/sessionStorage/cookie. Login form in App.svelte. Token validated against UsersPanel.

## How to add a new AI feature

1. **Backend handler:** Add handler function in `main.go` (or new file in `composerx/backend/`)
2. **Route:** Register in `registerRoutes()` method
3. **AI call:** Use `a.ai.ChatCompletion()` (MorphAI) or `a.chatCompletionClient()` (OpenAI-compatible for JSON/vision)
4. **Frontend:** Add UI in `App.svelte` or new component
5. **API client:** Add fetch calls in the Svelte component

## Key env vars

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8043` | Backend port |
| `TRAN_MYSQL_DSN` | — | MySQL DSN |
| `TRAN_MONGO_URI` | — | MongoDB URI |
| `TRAN_MONGO_DB` | `alterathena` | MongoDB DB name |
| `TRAN_REDIS_ADDR` | — | Redis address |
| `TRAN_FILE_STORAGE_PATH` | `./storage` | Local file storage |
| `MORPH_AI_API_KEY` | — | DashScope API key |
| `MORPH_AI_MODEL` | `qwen3-max` | AI model |
| `MORPH_AI_BASE_URL` | DashScope compatible-mode | OpenAI-compatible endpoint |
| `TRAN_OPENAI_API_KEY` | — | OpenAI key (for embeddings) |
| `USERS_PANEL_BASE_URL` | — | UsersPanel URL |