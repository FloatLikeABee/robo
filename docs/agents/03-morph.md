# 03 — Morph (MorphData / Morph AI)

## Overview

**Morph** is the flagship app. It's a Go monolith (Gin) serving both REST API and React SPA. It has the most complex architecture: 4 databases, self-hosted auth, AI chat, Tran admin (school management), forms, knowledge/GraphRAG, and hybrid context RAG.

- **Module:** `idongivaflyinfa` (Go 1.24)
- **Backend:** `morph/` — Go + Gin
- **Frontend:** `morph/frontend/` — React 18 (CRA)
- **Ports:** API `9090`, UI `3031` (dev)
- **Auth:** Self-hosted JWT + bcrypt (NOT UsersPanel)

## Backend architecture

### Entry point: `morph/main.go`

Startup sequence:
1. `config.GetConfig()` — loads `.env` via godotenv
2. `db.New(cfg.DBPath)` — opens BadgerDB
3. `cache.New()` — in-memory go-cache (5-min TTL)
4. `ai.New(apiKey, model, cache)` — creates MorphAI client (DashScope qwen3-max)
5. Optional: MySQL, MongoDB, Redis connections (each logs warning if unavailable)
6. `handlers.New(...)` — creates the Handlers struct with all dependencies
7. Gin router: CORS, `AuthzMiddleware()`, Swagger, route registration, static files
8. Listen on `cfg.Port` (default `9090`)

### Handlers struct (`morph/handlers/handlers.go`)

```go
type Handlers struct {
    db                    *db.DB              // BadgerDB
    aiService             *ai.AIService       // MorphAI client
    externalAPIBase       string              // Image/PDF reader base URL
    jwtCfg                auth.TokenConfig    // JWT secret + expiry
    TranMySQL             *db.TranMySQL       // optional
    TranMongo             *db.TranMongo       // optional
    TranRedis             *db.TranRedis       // optional
    hybridStore           *hybridcontext.Store // per-session RAG context
    sharpReportBase       string
    tranFormBase          string
    tranMailBase          string
    bookiBase             string
    entityAttachmentMax   int
    entityAttachmentRootDir string
    importJobs            *importJobStore
    ginEngine             *gin.Engine
}
```

### Route registration (`morph/handlers/register_routes.go`)

All routes are registered in a single function `RegisterAPIRoutes`. Key route groups:

| Group | Prefix | Purpose |
|-------|--------|---------|
| Auth | `/api/auth` | Login, me, user, permissions (self-hosted) |
| Admin | `/api/admin` | User CRUD (admin-only) |
| Chat | `/api/chat` | AI chat, sessions, messages |
| Hybrid Context | `/api/chat/hybrid-context` | Temporary RAG corpora per session |
| Knowledge | `/api/knowledge` | Knowledge file upload, GraphRAG search |
| Tran Admin | `/api/tran/*` | Districts, facilities, members, employees, contacts, assets, activities, generic data, case tasks, story posts |
| Forms | `/api/forms` | Form templates, answers, assignments |
| Data Collector | `/api/data-collector` | CSV/XLSX import |
| Products | `/api/products` | Product files |
| Proxies | `/api/messages`, `/api/users-panel`, `/api/composerx` | Legacy proxies |

### Database layer

| DB | File | Purpose |
|----|------|---------|
| **BadgerDB** | `morph/db/db.go` | Primary embedded KV: chat sessions, messages, forms, voice profiles. Key prefixes: `chat_sess:`, `chat_msg:`, `form_template:`, `form_answer:`, `voice_profile:` |
| **MySQL** | `morph/db/tran_mysql.go` | Tran admin data: districts, facilities, members, employees, contacts, assets, activities, generic data, case tasks, story posts, comments, notes, big notes, grid configs, **plat_users** (auth) |
| **MongoDB** | `morph/db/tran_mongo.go` | Large JSON detail documents for Tran entities + form results. DB name from `TRAN_MONGO_DB` (default `athena`) |
| **Redis** | `morph/db/tran_redis.go` | Session caching for Tran subsystem |

### Auth (`morph/auth/`, `morph/handlers/auth_local.go`, `morph/handlers/authz_middleware.go`)

**Self-hosted.** See `01-auth-flow.md` for full details. Key points:
- JWT HS256, secret from `JWT_SECRET` env
- Users in MySQL `plat_users` table, passwords bcrypt-hashed
- Bootstrap admin from `ADMIN_EMAIL`/`ADMIN_PASSWORD` env vars
- `AuthzMiddleware()` applied to all routes; `/api/auth/login` is public

### AI layer (`morph/ai/ai.go`)

- Wraps `pkg/morphai` client
- Default model: `qwen3-max` (DashScope)
- Methods: `GenerateSQL`, `CorrectSpelling`, `ChatCompletion`, `ChatCompletionLong`
- Prompt helpers in `PromptHelper.go`

### Key packages

| Package | Path | Purpose |
|---------|------|---------|
| `config` | `morph/config/config.go` | Env loading, `Config` struct |
| `db` | `morph/db/` | BadgerDB, MySQL, Mongo, Redis, plat_users |
| `cache` | `morph/cache/cache.go` | In-memory TTL cache |
| `ai` | `morph/ai/` | AI client + prompts |
| `handlers` | `morph/handlers/` | All HTTP handlers + route registration |
| `auth` | `morph/auth/` | JWT encode/decode |
| `models` | `morph/models/` | Shared data structures |
| `hybridcontext` | `morph/hybridcontext/` | Per-session RAG corpora |
| `importcol` | `morph/importcol/` | CSV/XLSX import logic |
| `validation` | `morph/validation/` | Input validation |
| `cmd` | `morph/cmd/` | CLI tools (seed, backfill, prune) |
| `migrations` | `morph/migrations/` | SQL migration files |

## Frontend architecture

### Build & tooling
- **CRA** (Create React App) with `@craco/craco` for customization
- **Proxy:** `"proxy": "http://localhost:9090"` — dev server proxies `/api` to Go backend
- **Port:** 3031 (dev)

### Key dependencies
- React 18, MUI v5, axios, react-router-dom v7, react-markdown, react-quill, leaflet, dayjs, jspdf + html2canvas

### Source structure (`morph/frontend/src/`)

| File/Dir | Purpose |
|----------|---------|
| `App.js` | Root: renders `SkoolAiChat` with dark/light theme |
| `appRouter.js` | React Router v7: `/login`, protected routes |
| `SkoolAiChat.js` | Main AI chat (~1200 lines): messages, input, hybrid context, notes/todos, sessions |
| `apiBase.js` | `API_BASE_URL` from `REACT_APP_API_URL` env |
| `api/tranClient.js` | Axios instance with Bearer token, 401 redirect |
| `auth/morphSession.js` | Token management: cookie + localStorage |
| `auth/isMorphAdmin.js` | Admin role check |
| `HybridContextDrawer.js` | RAG source management UI |
| `PlatformUiContext.jsx` | Platform UI configuration |
| `components/` | `admin/` (People, Districts, Vehicles, Trips, etc.), `chat/`, `common/`, `Forms/`, `notesTodos/` |
| `pages/` | `LoginPage.js`, `admin/` (all Tran entity pages, BigNotes, UsersAdmin, DataImport) |

### Production serving
In production, Go serves the React build from `frontend/build/` as static files. `r.NoRoute` serves `index.html` for SPA fallback.

## How to add a new Tran admin entity

Follow the pattern in existing entities (e.g., `handlers/tran_assets.go`):

1. **MySQL schema:** Add table in `morph/db/tran_mysql.go` `ensureTranMySQLSchema()`
2. **MySQL CRUD:** Add methods in `morph/db/tran_mysql.go` (Create, Get, List, Update, Delete)
3. **MongoDB detail:** Add detail methods in `morph/db/tran_mongo.go` if entity has JSON detail
4. **Handlers:** Create handler file in `morph/handlers/` following the pattern:
   - `List<Entity>`, `Get<Entity>`, `Get<Entity>Full`, `Create<Entity>`, `Update<Entity>`, `Delete<Entity>`
5. **Routes:** Register in `morph/handlers/register_routes.go`:
   ```go
   tran := api.Group("/tran")
   tran.GET("/<entities>", h.List<Entity>)
   tran.GET("/<entities>/:id", h.Get<Entity>)
   tran.POST("/<entities>", h.Create<Entity>)
   tran.PUT("/<entities>/:id", h.Update<Entity>)
   tran.DELETE("/<entities>/:id", h.Delete<Entity>)
   ```
6. **Models:** Add structs in `morph/models/`
7. **Frontend:** Add page in `morph/frontend/src/pages/admin/`, add route in `appRouter.js`, add nav item in `SkoolAiChat.js` or `AppDrawer.js`

## macOS note

On macOS, `go run main.go` fails with BadgerDB `LC_UUID` error. Use `go build -o morph-server main.go && ./morph-server` instead. `start-all.sh` handles this automatically.

## Key env vars

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `9090` | Backend port |
| `MORPH_AI_API_KEY` | — | DashScope API key |
| `MORPH_AI_MODEL` | `qwen3-max` | AI model |
| `DB_PATH` | `./data/badger` | BadgerDB path |
| `JWT_SECRET` | `morph-dev-jwt-secret-change-me` | JWT signing secret |
| `ADMIN_EMAIL` | `admin@local.com` | Bootstrap admin email |
| `ADMIN_PASSWORD` | `admin` | Bootstrap admin password |
| `TRAN_MYSQL_DSN` | `root:Dafuq@911@tcp(127.0.0.1:3306)/tran?...` | MySQL DSN |
| `TRAN_MONGO_URI` | `mongodb://localhost:27017/` | MongoDB URI |
| `TRAN_MONGO_DB` | `athena` | MongoDB DB name |
| `TRAN_REDIS_ADDR` | `127.0.0.1:6379` | Redis address |