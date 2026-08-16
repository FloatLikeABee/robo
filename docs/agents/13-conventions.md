# 13 — Conventions & Cross-Cutting Patterns

## General rules

1. **Minimal diffs:** Change only files required for the task. Match existing naming, formatting, and patterns in that specific app.
2. **Scope to one app:** Do not refactor across apps unless explicitly asked. Each app is an independent project.
3. **Verify per-app:** Run the app-specific build/test command scoped to edited packages.
4. **Read app READMEs:** Always read each project's own `README.md` (or `design.md`) before non-trivial changes.
5. **Environment:** Each app uses `.env` or shell exports differently. Grep for `godotenv`, `dotenv`, or `DATABASE_URL` in that repo.
6. **Auth:** Know which auth pattern the app uses before touching login/session code. See `01-auth-flow.md`.
7. **Cross-project assistant behavior:** Align with `AI_ASSISTANT_MORPHAI_CONTRACT.md` when touching assistant endpoints.

---

## Go backend conventions

### Project structure

Go backends follow one of two patterns:

**Pattern A: Single-file (`main.go`)** — morph, composerx
```
app/
├── main.go            ← Entry point, App struct, all handlers, route registration
├── models.go          ← Data structures
├── repository.go      ← DB CRUD
├── go.mod
└── go.sum
```

**Pattern B: `cmd/internal` layout** — formx, booki, academi
```
app/
├── cmd/server/main.go ← Entry point
├── internal/
│   ├── config/        ← Env loading
│   ├── handler/       ← HTTP handlers + route registration
│   ├── models/        ← Data structures
│   ├── mysql/         ← DB repos (or mongo/, etc.)
│   └── middleware/     ← Auth middleware
├── go.mod
└── go.sum
```

### Adding a new endpoint

1. **Handler:** Add handler function in the handlers file/package
2. **Route:** Register in the route registration function
3. **Model:** Add request/response structs if needed
4. **DB:** Add repository methods if needed

### Gin patterns

```go
// Handler signature
func (h *Handler) MyEndpoint(c *gin.Context) {
    // Parse request
    var req MyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Get user from context (set by auth middleware)
    userID := c.GetString("auth_user_id")

    // Do work...

    // Return response
    c.JSON(200, myResponse)
}

// Route registration
func (h *Handler) Register(r *gin.RouterGroup) {
    r.GET("/my-endpoint", h.MyEndpoint)
    r.POST("/my-endpoint", h.MyCreateEndpoint)
}
```

### Config loading

```go
// Most apps use this pattern:
func Load() Config {
    return Config{
        Port:       getEnv("SERVER_PORT", "9090"),
        MySQLDSN:   getEnv("TRAN_MYSQL_DSN", "..."),
        MongoURI:   getEnv("TRAN_MONGO_URI", "..."),
        // ...
    }
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
```

### AI integration

```go
// Create client (once at startup)
aiClient := morphai.NewClient(morphai.LoadFromEnv())

// Use in handlers
messages := []morphai.Message{
    {Role: "system", Content: systemPrompt},
    {Role: "user", Content: userMessage},
}
response, err := aiClient.ChatCompletion(ctx, messages)
```

---

## Rust backend conventions

### Project structure

```
app/
├── src/
│   ├── main.rs         ← Entry point
│   ├── config.rs       ← Config struct + from_env()
│   ├── api/            ← Handler modules
│   │   ├── mod.rs
│   │   ├── auth.rs
│   │   └── ...
│   └── models.rs       ← Data structures
├── Cargo.toml
└── Cargo.lock
```

### Adding a new endpoint

```rust
// 1. Handler in api/my_module.rs
pub async fn my_handler(
    State(state): State<AppState>,
    Json(body): Json<MyRequest>,
) -> Result<Json<MyResponse>, AppError> {
    // ...
}

// 2. Route in main.rs or api/mod.rs
let app = Router::new()
    .route("/api/v1/my-endpoint", post(my_handler))
    .with_state(state);
```

### Config pattern

```rust
pub struct Config {
    pub database_url: String,
    pub jwt_secret: String,
    pub port: u16,
}

impl Config {
    pub fn from_env() -> Self {
        Self {
            database_url: std::env::var("DATABASE_URL").unwrap_or_default(),
            jwt_secret: std::env::var("JWT_SECRET").unwrap_or_default(),
            port: std::env::var("PORT").unwrap_or("5001".into()).parse().unwrap(),
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

---

## Frontend conventions

### React (Vite) — formx, booki, morph-utils

```
src/
├── App.tsx             ← Router + layout
├── pages/              ← Page components
├── components/         ← Reusable components
├── lib/api.ts          ← Centralized API client
└── main.tsx            ← Entry point
```

**API client pattern:**
```typescript
// lib/api.ts
const api = {
  forms: {
    list: () => fetch('/api/v1/forms', { headers: authHeader() }),
    create: (data) => fetch('/api/v1/forms', { method: 'POST', body: JSON.stringify(data), headers: authHeader() }),
  },
};
```

**Auth:** JWT in localStorage + cookie. Axios/fetch interceptor adds Bearer token. On 401: clear tokens, redirect to `/login`.

### React (CRA) — morph

```
src/
├── App.js              ← Root component
├── appRouter.js        ← React Router
├── api/tranClient.js   ← Axios instance with auth
├── auth/morphSession.js ← Token management
├── pages/              ← Page components
└── components/         ← Reusable components
```

### Svelte (Vite) — composerx, UsersPanel, morph-engi

```
src/
├── App.svelte          ← Main app (auth, routing, state)
├── components/         ← Reusable components
├── lib/                ← Utilities
└── main.js             ← Entry point
```

**Auth:** JWT in localStorage. Fetch wrapper adds Bearer token. On 401: redirect to login.

### SvelteKit — SharpReport

```
src/
├── routes/             ← File-based routing
│   ├── +page.svelte
│   └── +layout.svelte
└── lib/                ← Utilities
```

---

## Database conventions

| DB | Typical use | Driver/ORM |
|----|------------|------------|
| MySQL | Structured data (users, forms, templates, bookings) | Go: `database/sql` or GORM. Rust: SQLx |
| MongoDB | Documents, responses, large JSON | Go: `mongo-driver`. Rust: N/A |
| BadgerDB | Embedded KV (chat sessions, forms, voice) | Go: `badger/v4` |
| SQLite | Local dev / single-file DB | Rust: SQLx |
| Redis | Sessions, caching | Go: `go-redis`. Rust: N/A |

### Schema management

- **Go/MySQL:** Schema ensured idempotently on connect (e.g., `ensureTranMySQLSchema()`)
- **Go/MongoDB:** Collections created on first write
- **Rust/SQLx:** Migrations via `sqlx::migrate!()` or manual schema files
- **BadgerDB:** No schema — key-value store

---

## Naming conventions

- **Go packages:** lowercase, single word (e.g., `handlers`, `models`, `config`)
- **Go modules:** domain-style (e.g., `github.com/formsx/backend`, `mergeemailx-backend`)
- **Rust crates:** kebab-case (e.g., `users-panel-api`, `datapulse`)
- **Env vars:** `UPPER_SNAKE_CASE` with app prefix (`TRAN_*`, `MORPH_AI_*`)
- **API routes:** `/api/v1/<resource>` (RESTful)
- **React components:** PascalCase (e.g., `PlatformChatDrawer.tsx`)
- **Svelte components:** PascalCase (e.g., `PlatformChatDrawer.svelte`)

---

## Error handling

### Go

```go
// Return JSON errors
c.JSON(400, gin.H{"error": "Invalid request"})
c.JSON(500, gin.H{"error": "Internal server error"})

// Log warnings for non-critical failures
log.Printf("Warning: Failed to connect to MySQL: %v", err)
```

### Rust

```rust
// Use a unified error type
enum AppError {
    BadRequest(String),
    Internal(String),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        // Map to status code + JSON body
    }
}
```

---

## ⚠️ Critical: morph auth is self-hosted

Despite what `DEVELOPER_BASELINE.md` says, **morph no longer uses UsersPanel for auth**. It has its own JWT + bcrypt system in MySQL `plat_users` table. Do not add UsersPanel dependencies to morph auth code. See `01-auth-flow.md` for details.