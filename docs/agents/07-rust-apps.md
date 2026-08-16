# 07 — Rust Apps (UsersPanel, SharpReport, morph-engi)

## Overview

Three Rust backends share common patterns: Axum framework, SQLx for database, `pkg/morphai-rs` for AI, and UsersPanel SSO for auth (except UsersPanel itself which is the auth hub).

```
┌─────────────────────────────────────────────────────────────────┐
│                     RUST APP PATTERNS                            │
│                                                                 │
│  All three share:                                               │
│  • Axum 0.7 + Tokio                                             │
│  • SQLx for database                                            │
│  • pkg/morphai-rs for AI                                        │
│  • JWT auth (self-issued or UsersPanel SSO)                     │
│  • Config from env vars                                         │
│                                                                 │
│  Differences:                                                   │
│  • UsersPanel: MySQL, central auth, messaging                   │
│  • SharpReport: SQLite, embedded Metabase JAR, analytics        │
│  • morph-engi: SQLite/MySQL, civil engineering domain           │
└─────────────────────────────────────────────────────────────────┘
```

---

## UsersPanel — Central Auth & User Admin

- **Crate:** `users-panel-api` (Rust edition 2021)
- **Port:** API `5001`, Admin UI `5173`
- **Database:** MySQL (SQLx 0.8)
- **Auth:** Self-hosted JWT + bcrypt + Google OAuth2

### Backend structure (`UsersPanel/backend/src/`)

| File | Purpose |
|------|---------|
| `main.rs` | Entry point: config, DB pool, router, listen |
| `config.rs` | `DATABASE_URL`, `JWT_SECRET`, Google OAuth, bootstrap admin |
| `handlers/auth.rs` | Register, login, verify-email, Google OAuth, forgot/reset password |
| `handlers/admin.rs` | User CRUD, morph-users sync |
| `handlers/assistant.rs` | MorphAI contract assistant |
| `handlers/messages.rs` | Inbox, threads, messaging |
| `handlers/data_collector.rs` | CSV/Excel import |
| `handlers/integration.rs` | Main panel, forms, email, reports |
| `permissions.rs` | Permission definitions |
| `roles_db.rs` | Role management |
| `public_channel.rs` | Public channel logic |
| `message_ai.rs` | AI-powered message features |

### Key endpoints (all under `/api`)

| Group | Key endpoints |
|-------|---------------|
| Auth | `POST /auth/register`, `POST /auth/login`, `GET /auth/google`, `GET /auth/google/callback`, `POST /auth/forgot-password`, `POST /auth/reset-password`, `GET /auth/roles`, `GET /auth/permissions`, `GET /auth/user` |
| Assistant | `POST /assistant/chat` |
| Admin | `GET/POST /admin/users`, `PATCH/DELETE /admin/users/:user_id`, `GET /admin/morph-users` |
| Messages | `GET /messages/inbox`, `GET /messages/users`, `POST /messages/threads`, `GET/POST /messages/threads/:thread_id/messages` |
| Data Collector | `GET /data-collector/entities`, `GET /data-collector/templates/:entity`, `POST /data-collector/validate`, `POST /data-collector/jobs` |
| Integration | `GET /main-panel`, `POST /forms`, `POST /email/compose`, `GET /reports` |

### Admin UI (`UsersPanel/admin/`)
- Svelte 5 + Vite 8 + TypeScript 6.0
- Port: 5173
- Routes: Login, Register, Users (admin), DataCollector (admin), AuthCallback, ResetPassword
- Uses `@robo/platform-chat` for AI assistant drawer

### ⚠️ Critical role

UsersPanel is the **central auth service** for the entire platform (except morph). All other apps validate credentials through it. Changes to auth endpoints or JWT format affect every other app.

---

## SharpReport — DataPulse Analytics

- **Crate:** `datapulse` (Rust edition 2024)
- **Port:** API `3050`, UI `5178`
- **Database:** SQLite (primary, via SQLx 0.7), also supports PostgreSQL/MySQL
- **Auth:** JWT (self-issued) + UsersPanel SSO
- **Special:** Embedded Metabase JAR (auto-download, auto-start, health monitoring, API proxy)

### Backend structure (`SharpReport/backend/src/`)

| File/Dir | Purpose |
|----------|---------|
| `main.rs` | Entry point: config, DB, Metabase orchestrator, router |
| `config.rs` | Layered config (default.toml + env-specific + env vars) |
| `api/` | 19 API modules: auth, databases, data_tables, dashboards, queries, embed, metabase proxy, report builder, report assistant, data page publishing, setup wizard, MCP tools, AI status |

### Key endpoints (all under `/api/v1`)

| Group | Key endpoints |
|-------|---------------|
| Health/AI | `GET /health`, `GET /ai/status`, `GET /ai/mcp-tools` |
| Auth | `POST /auth/login`, `POST /auth/logout`, `GET /auth/me` |
| Setup | `GET /setup/status`, `POST /setup/initialize`, `POST /setup/admin`, `POST /setup/database` |
| Databases | CRUD at `/databases`, test, schema |
| Data Tables | CRUD, analyze, import, rows, query, page-build, page-chat |
| Publishing | `GET /publishes/resolve-path`, `GET/POST /publishes`, `GET /public/p/:slug` (public, no auth) |
| Reports | `POST /reports/builder/execute`, `POST /reports/assistant/chat`, `POST /assistant/chat`, `POST /data-ai/chat` |
| Dashboards | CRUD at `/dashboards`, embed |
| Queries | CRUD at `/queries`, execute |
| Metabase | `ANY /metabase/*path` — transparent proxy |

### Metabase integration

SharpReport embeds a Metabase JAR as a managed subprocess:
- Auto-downloads Metabase JAR on first startup
- Auto-starts Metabase as a child process
- Monitors Metabase health
- Proxies all `/metabase/*` requests to the Metabase instance
- Generates signed embedding tokens for dashboard/card embedding

### Frontend (`SharpReport/frontend/`)
- SvelteKit 2 + Svelte 5 + Vite 5 + TypeScript 5 + TailwindCSS 4
- ECharts (via echarts-for-svelte), Bits UI, Lucide Svelte
- Port: 5178
- Uses `@robo/platform-chat`

---

## morph-engi — Civil Engineering Platform

- **Crate:** `morph-engi` (Rust edition 2021)
- **Port:** API `9096`, UI `5179`
- **Database:** SQLite (default, `morph_engi.db`) or MySQL (SQLx 0.7)
- **Auth:** JWT (self-issued) + UsersPanel SSO

### Backend structure (`morph-engi/backend/src/`)

| File | Purpose |
|------|---------|
| `main.rs` | Entry point |
| `config.rs` | `DATABASE_URL`, `JWT_SECRET`, `APP_PORT`, `USERS_PANEL_BASE_URL` |
| `api/auth.rs` | Login, platform-session, dev-login, me |
| `api/assistant.rs` | MorphAI contract assistant with full CRUD tool suite |
| `api/modules.rs` | Projects, tasks, site-logs, materials, resource-files, finance, budget-lines, resources, contractors, contracts, relations, communications |
| `api/verification.rs` | Verification sessions |
| `api/extract.rs` | Data extraction |

### Key endpoints (all under `/api/v1`)

| Group | Key endpoints |
|-------|---------------|
| Auth | `POST /auth/login`, `POST /auth/platform-session`, `POST /auth/dev-login`, `GET /auth/me` |
| Organization | `GET /organization`, `PATCH /organization` |
| Assistant | `POST /assistant/chat` |
| Projects | `GET/POST /projects`, `PATCH/DELETE /projects/:id` |
| Tasks | `GET/POST /tasks` |
| Site Logs | `GET/POST /site-logs` |
| Materials | `GET/POST /materials`, `PATCH /materials/:id`, `GET/POST /material-usages` |
| Files | `GET/POST /resource-files`, `PATCH /resource-files/:id`, `POST /resource-files/upload` |
| Finance | `GET/POST /finance`, `GET/POST /budget-lines`, `PATCH /budget-lines/:id` |
| Resources | `GET/POST /resources`, `GET/POST /resource-allocations` |
| Contractors | `GET/POST /contractors`, `GET/POST /contracts` |
| Relations | `GET/POST /relations`, `GET/POST /communications` |
| Verification | `GET /verification/sessions`, `GET /verification/sessions/:id`, `POST /verification/run` |

### Frontend (`morph-engi/frontend/`)
- Svelte 5 + Vite 7 + TypeScript 5.9 + TailwindCSS 4
- Port: 5179
- Pages: Projects, Files, People, Settings
- Uses `@robo/platform-chat`

---

## Common Rust patterns

### Config loading

```rust
// All Rust apps follow this pattern:
pub struct Config {
    pub database_url: String,
    pub jwt_secret: String,
    pub port: u16,
    pub users_panel_base_url: String,
}

impl Config {
    pub fn from_env() -> Self {
        Self {
            database_url: std::env::var("DATABASE_URL").unwrap_or_default(),
            jwt_secret: std::env::var("JWT_SECRET").unwrap_or_default(),
            port: std::env::var("APP_PORT").or(std::env::var("PORT")).unwrap_or("5001").parse().unwrap(),
            users_panel_base_url: std::env::var("USERS_PANEL_BASE_URL").unwrap_or_default(),
        }
    }
}
```

### AI integration

```rust
use morphai::{Client, Config, Message};

let cfg = Config::from_env();
let client = Client::new(cfg);
let messages = vec![Message::user("Hello")];
let response = client.chat_completion(&messages).await?;
```

### Database (SQLx)

```rust
// MySQL (UsersPanel)
let pool = MySqlPool::connect(&config.database_url).await?;

// SQLite (SharpReport, morph-engi)
let pool = SqlitePool::connect(&config.database_url).await?;

// Query
let rows = sqlx::query_as::<_, MyStruct>("SELECT * FROM table")
    .fetch_all(&pool)
    .await?;
```

### Adding a new endpoint (Axum)

```rust
// 1. Define handler
async fn my_handler(
    State(state): State<AppState>,
    Json(body): Json<MyRequest>,
) -> Result<Json<MyResponse>, AppError> {
    // ...
}

// 2. Register route in main.rs or router module
let app = Router::new()
    .route("/api/v1/my-endpoint", post(my_handler))
    .with_state(state);
```