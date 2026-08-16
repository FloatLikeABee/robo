# DataPulse - Metabase-Powered Analytics Platform

## Overview

DataPulse is a modern analytics platform that embeds Metabase as its reporting engine, providing a sleek, intuitive interface for database connectivity, data manipulation, and visualization. Built with Rust backend and Svelte frontend, it delivers enterprise-grade performance with a futuristic user experience.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Svelte)                        │
│  ┌─────────────┐ ┌──────────────┐ ┌────────────┐ ┌───────────┐ │
│  │ Dashboard   │ │ Query Builder│ │ Visualizer │ │ Settings  │ │
│  │ Component   │ │ Component    │ │ Component  │ │ Component │ │
│  └──────┬──────┘ └──────┬───────┘ └─────┬──────┘ └─────┬─────┘ │
│         └────────────────┴───────────────┴──────────────┘       │
│                            │                                    │
│                     REST/WebSocket API                          │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                     Backend (Rust/Axum)                         │
│  ┌─────────────┐ ┌────────▼──────┐ ┌───────────┐ ┌──────────┐  │
│  │ Auth &      │ │  Metabase     │ │ Database  │ │ Embed    │  │
│  │ Session Mgr │ │  Orchestrator │ │ Connector │ │ API      │  │
│  └──────┬──────┘ └──────┬────────┘ └─────┬─────┘ └────┬─────┘  │
│         └────────────────┴────────────────┴────────────┘        │
│                            │                                    │
│  ┌─────────────────────────▼──────────────────────────────┐    │
│  │              Metabase Process Manager                   │    │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────────┐  │    │
│  │  │ First-run   │ │ Health Check │ │ API Proxy/Embed │  │    │
│  │  │ Setup       │ │ & Restart    │ │ Token Generator │  │    │
│  │  └─────────────┘ └──────────────┘ └─────────────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
        ┌─────▼─────┐ ┌─────▼─────┐ ┌─────▼─────┐
        │ Metabase  │ │  User     │ │  App      │
        │ H2/Postgres│ │ Databases │ │ Database  │
        │ (internal)│ │ (MySQL,   │ │ (SQLite/  │
        │           │ │  PG, etc) │ │  Postgres)│
        └───────────┘ └───────────┘ └───────────┘
```

---

## Tech Stack

### Backend


| Component          | Technology                    | Purpose                          |
| ------------------ | ----------------------------- | -------------------------------- |
| Runtime            | Rust 2024 Edition             | Core language                    |
| Web Framework      | Axum                          | HTTP server, routing             |
| Async Runtime      | Tokio                         | Async execution                  |
| Serialization      | Serde + serde_json            | Data serialization               |
| Database ORM       | SQLx                          | App database (PostgreSQL/SQLite) |
| Process Management | tokio::process                | Metabase JVM management          |
| HTTP Client        | reqwest                       | Metabase API communication       |
| Auth               | jsonwebtoken + argon2         | JWT + password hashing           |
| Config             | config + dotenv               | Configuration management         |
| Logging            | tracing + tracing-subscriber  | Structured logging               |
| Validation         | validator                     | Input validation                 |
| Embedding          | metabase-embed crate (custom) | Signed embedding tokens          |


### Frontend


| Component         | Technology                       | Purpose             |
| ----------------- | -------------------------------- | ------------------- |
| Framework         | Svelte 5 (Runes)                 | Reactive UI         |
| Build Tool        | Vite                             | Fast dev/build      |
| Styling           | TailwindCSS v4                   | Utility-first CSS   |
| Component Library | Bits UI + CVA                    | Headless components |
| Icons             | Lucide Svelte                    | Icon system         |
| State Management  | Svelte Stores                    | Client state        |
| HTTP Client       | Fetch API + custom hooks         | API communication   |
| Charts            | ECharts (via echarts-for-svelte) | Data visualization  |
| Theme             | CSS variables + Tailwind         | Dark/light themes   |
| Animations        | Svelte transitions + CSS         | Smooth interactions |
| Routing           | SvelteKit                        | Client-side routing |


### Infrastructure


| Component        | Technology                        | Purpose            |
| ---------------- | --------------------------------- | ------------------ |
| Metabase         | metabase.jar (latest)             | Reporting engine   |
| Containerization | Docker + docker-compose           | Deployment         |
| Reverse Proxy    | Nginx/Caddy                       | Request routing    |
| Process Manager  | systemd (Linux) / launchd (macOS) | Service management |


---

## Core Features

### 1. First-Run Setup Wizard

- Automatic detection of Metabase installation
- Download metabase.jar if not present
- JVM detection and validation
- Interactive setup flow:
  1. Welcome screen
  2. JVM check/installation guide
  3. Metabase initialization
  4. Admin account creation
  5. First database connection
  6. Completion dashboard

### 2. Database Connectivity

- Support for all Metabase-compatible databases:
  - PostgreSQL, MySQL, SQL Server
  - MongoDB, BigQuery, Snowflake
  - SQLite, MariaDB, Oracle
  - ClickHouse, Druid, Presto
- Connection testing and validation
- SSL/TLS configuration UI
- Connection pooling management
- Schema browser with table/column preview

### 3. Metabase Integration

- **Embedding Modes:**
  - Signed embedding with JWT tokens
  - Static embedding for public dashboards
  - Full-app embedding for admin views
- **API Proxy:**
  - Transparent Metabase API proxy
  - Request/response transformation
  - Caching layer for performance
- **Lifecycle Management:**
  - Automatic start/stop
  - Health monitoring
  - Graceful shutdown
  - Auto-restart on failure

### 4. Dashboard & Analytics

- Pre-built dashboard templates
- Drag-and-drop widget placement
- Real-time data refresh
- Export options (PDF, CSV, PNG)
- Scheduled report delivery
- Embeddable widgets for external sites

### 5. Query Builder

- Visual query builder (no-code)
- SQL editor with syntax highlighting
- Query history and favorites
- Query parameters and variables
- Result pagination and sorting

### 6. User Management

- Role-based access control (RBAC):
  - Admin, Editor, Viewer, API User
- SSO integration (OAuth2, SAML)
- API key management
- Audit logging
- Team/workspace organization

### 7. Settings & Configuration

- Platform settings UI
- Metabase configuration panel
- Email/SMTP setup
- Webhook integrations
- Backup and restore
- System health monitoring

---

## Project Structure

```
datapulse/
├── backend/
│   ├── Cargo.toml
│   ├── Cargo.lock
│   ├── src/
│   │   ├── main.rs                 # Entry point
│   │   ├── config.rs               # Configuration management
│   │   ├── db/                     # Application database
│   │   │   ├── mod.rs
│   │   │   ├── models.rs
│   │   │   ├── migrations/
│   │   │   └── repositories/
│   │   ├── metabase/               # Metabase integration
│   │   │   ├── mod.rs
│   │   │   ├── orchestrator.rs     # Process management
│   │   │   ├── api_client.rs       # Metabase API wrapper
│   │   │   ├── embedding.rs        # Token generation
│   │   │   └── health.rs           # Health monitoring
│   │   ├── api/                    # HTTP routes
│   │   │   ├── mod.rs
│   │   │   ├── auth.rs
│   │   │   ├── databases.rs
│   │   │   ├── dashboards.rs
│   │   │   ├── queries.rs
│   │   │   ├── users.rs
│   │   │   ├── settings.rs
│   │   │   └── embed.rs
│   │   ├── services/               # Business logic
│   │   │   ├── mod.rs
│   │   │   ├── auth_service.rs
│   │   │   ├── database_service.rs
│   │   │   └── report_service.rs
│   │   ├── middleware/             # HTTP middleware
│   │   │   ├── mod.rs
│   │   │   ├── auth.rs
│   │   │   ├── logging.rs
│   │   │   └── cors.rs
│   │   └── utils/                  # Utilities
│   │       ├── mod.rs
│   │       └── crypto.rs
│   ├── migrations/                 # SQLx migrations
│   └── config/
│       ├── default.toml
│       ├── production.toml
│       └── development.toml
│
├── frontend/
│   ├── package.json
│   ├── svelte.config.js
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── src/
│   │   ├── app.html
│   │   ├── app.css
│   │   ├── routes/
│   │   │   ├── +layout.svelte
│   │   │   ├── +page.svelte
│   │   │   ├── setup/
│   │   │   │   └── +page.svelte
│   │   │   ├── dashboard/
│   │   │   │   ├── +page.svelte
│   │   │   │   └── [id]/
│   │   │   │       └── +page.svelte
│   │   │   ├── databases/
│   │   │   │   ├── +page.svelte
│   │   │   │   └── [id]/
│   │   │   │       └── +page.svelte
│   │   │   ├── queries/
│   │   │   │   ├── +page.svelte
│   │   │   │   └── new/
│   │   │   │       └── +page.svelte
│   │   │   ├── reports/
│   │   │   │   └── +page.svelte
│   │   │   ├── settings/
│   │   │   │   ├── +page.svelte
│   │   │   │   ├── database/
│   │   │   │   ├── metabase/
│   │   │   │   ├── users/
│   │   │   │   └── system/
│   │   │   └── embed/
│   │   │       └── [token]/
│   │   │           └── +page.svelte
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── ui/             # Reusable UI components
│   │   │   │   │   ├── button/
│   │   │   │   │   ├── card/
│   │   │   │   │   ├── input/
│   │   │   │   │   ├── modal/
│   │   │   │   │   ├── table/
│   │   │   │   │   ├── sidebar/
│   │   │   │   │   └── theme-toggle/
│   │   │   │   ├── dashboard/
│   │   │   │   ├── database/
│   │   │   │   ├── query/
│   │   │   │   └── embed/
│   │   │   ├── stores/
│   │   │   │   ├── auth.svelte.ts
│   │   │   │   ├── theme.svelte.ts
│   │   │   │   └── app.svelte.ts
│   │   │   ├── api/
│   │   │   │   ├── client.ts
│   │   │   │   ├── auth.ts
│   │   │   │   ├── databases.ts
│   │   │   │   └── dashboards.ts
│   │   │   ├── utils/
│   │   │   │   ├── formatters.ts
│   │   │   │   └── validators.ts
│   │   │   └── types/
│   │   │       └── index.ts
│   │   └── hooks/
│   │       └── auth.ts
│   └── static/
│       └── favicon.png
│
├── deploy/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   └── nginx.conf
│
├── metabase/
│   └── .gitkeep                     # metabase.jar downloaded here
│
├── docs/
│   ├── API.md
│   ├── DEPLOYMENT.md
│   └── DEVELOPMENT.md
│
├── .env.example
├── .gitignore
└── README.md
```

---

## Database Schema (Application DB)

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    avatar_url TEXT,
    metabase_user_id INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Database Connections
CREATE TABLE database_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    engine VARCHAR(50) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_encrypted TEXT NOT NULL,
    ssl_enabled BOOLEAN DEFAULT false,
    ssl_config JSONB,
    additional_config JSONB,
    metabase_database_id INTEGER,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Dashboards (local metadata)
CREATE TABLE dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metabase_dashboard_id INTEGER,
    database_id UUID REFERENCES database_connections(id),
    layout_config JSONB,
    is_public BOOLEAN DEFAULT false,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Saved Queries
CREATE TABLE saved_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query_text TEXT NOT NULL,
    database_id UUID REFERENCES database_connections(id),
    metabase_card_id INTEGER,
    created_by UUID REFERENCES users(id),
    is_favorite BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id),
    permissions JSONB,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit Log
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- System Settings
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Setup State
CREATE TABLE setup_state (
    id INTEGER PRIMARY KEY DEFAULT 1,
    is_completed BOOLEAN DEFAULT false,
    metabase_initialized BOOLEAN DEFAULT false,
    metabase_port INTEGER DEFAULT 3001,
    metabase_pid INTEGER,
    admin_user_created BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## API Design

### Authentication

```
POST   /api/v1/auth/login          # Login
POST   /api/v1/auth/logout         # Logout
POST   /api/v1/auth/refresh        # Refresh token
GET    /api/v1/auth/me             # Get current user
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
```

### Setup

```
GET    /api/v1/setup/status        # Check setup status
POST   /api/v1/setup/initialize    # Initialize Metabase
POST   /api/v1/setup/admin         # Create admin user
POST   /api/v1/setup/database      # Add first database
```

### Databases

```
GET    /api/v1/databases           # List connections
POST   /api/v1/databases           # Create connection
GET    /api/v1/databases/:id       # Get connection
PUT    /api/v1/databases/:id       # Update connection
DELETE /api/v1/databases/:id       # Delete connection
POST   /api/v1/databases/:id/test  # Test connection
GET    /api/v1/databases/:id/schema # Browse schema
GET    /api/v1/databases/:id/schema/:table # Table details
```

### Dashboards

```
GET    /api/v1/dashboards          # List dashboards
POST   /api/v1/dashboards          # Create dashboard
GET    /api/v1/dashboards/:id      # Get dashboard
PUT    /api/v1/dashboards/:id      # Update dashboard
DELETE /api/v1/dashboards/:id      # Delete dashboard
GET    /api/v1/dashboards/:id/embed # Get embed URL
```

### Queries

```
GET    /api/v1/queries             # List saved queries
POST   /api/v1/queries             # Create query
GET    /api/v1/queries/:id         # Get query
PUT    /api/v1/queries/:id         # Update query
DELETE /api/v1/queries/:id         # Delete query
POST   /api/v1/queries/execute     # Execute ad-hoc query
POST   /api/v1/queries/:id/run     # Run saved query
```

### Embedding

```
POST   /api/v1/embed/dashboard/:id # Generate embed token
POST   /api/v1/embed/card/:id      # Generate embed token
GET    /embed/:token               # Public embed endpoint
```

### Settings

```
GET    /api/v1/settings            # Get all settings
PUT    /api/v1/settings            # Update settings
GET    /api/v1/settings/metabase   # Metabase config
PUT    /api/v1/settings/metabase   # Update Metabase config
GET    /api/v1/settings/system     # System health
POST   /api/v1/settings/metabase/restart # Restart Metabase
```

### Users (Admin)

```
GET    /api/v1/users               # List users
POST   /api/v1/users               # Create user
GET    /api/v1/users/:id           # Get user
PUT    /api/v1/users/:id           # Update user
DELETE /api/v1/users/:id           # Delete user
GET    /api/v1/users/:id/api-keys  # List API keys
POST   /api/v1/users/:id/api-keys  # Create API key
DELETE /api/v1/api-keys/:id        # Revoke API key
```

---

## Metabase Process Management

### Lifecycle Flow

```
Application Start
       │
       ▼
┌──────────────────┐
│ Check setup_state│
└────────┬─────────┘
         │
    ┌────▼────┐
    │Completed?│
    └────┬────┘
         │
    ┌────▼────────────────────┐
    │ No                      │ Yes
    ▼                         ▼
┌─────────────┐        ┌──────────────────┐
│Setup Wizard │        │Check Metabase PID│
│Flow         │        └────────┬─────────┘
└──────┬──────┘                 │
       │                  ┌─────▼─────┐
       │                  │Running?   │
       │                  └─────┬─────┘
       │                  ┌─────▼────────────────────┐
       │                  │ No          │ Yes        │
       ▼                  ▼             ▼
┌─────────────┐   ┌──────────────┐ ┌──────────────┐
│1. Check JVM │   │Start Metabase│ │Health Check  │
│2. Download  │   │Process       │ │Ping API      │
│  metabase.jar│  │              │ │              │
│3. Initialize│   └──────────────┘ └──────────────┘
│  Metabase   │
│4. Create    │
│  Admin      │
│5. Mark Done │
└─────────────┘
```

### Process Configuration

```toml
# config/default.toml
[metabase]
jar_path = "./metabase/metabase.jar"
download_url = "https://downloads.metabase.com/latest/metabase.jar"
jvm_opts = "-Xmx2g -Xms512m"
port = 3001
host = "127.0.0.1"
db_type = "h2"  # or "postgres" for production
health_check_interval = 30  # seconds
restart_on_failure = true
max_restarts = 5
```

### Health Monitoring

```rust
// Pseudocode for health checker
async fn health_check_loop() {
    let mut interval = interval(Duration::from_secs(30));
    loop {
        interval.tick().await;
        match check_metabase_health().await {
            Ok(_) => update_status(Healthy).await,
            Err(_) => {
                increment_failure_count().await;
                if failure_count() >= MAX_RESTARTS {
                    alert_admin("Metabase repeatedly failing").await;
                } else {
                    restart_metabase().await;
                }
            }
        }
    }
}
```

---

## Embedding Strategy

### Signed Embedding

```rust
// Generate signed JWT for Metabase embedding
fn generate_embed_token(
    resource: EmbedResource,
    params: EmbedParams,
    secret: &[u8],
) -> Result<String> {
    let claims = EmbedClaims {
        resource: match resource {
            EmbedResource::Dashboard(id) => Resource::Dashboard(id),
            EmbedResource::Card(id) => Resource::Card(id),
        },
        params: params.into_hashmap(),
        exp: (Utc::now() + Duration::minutes(10)).timestamp(),
    };
    
    encode(&Header::default(), &claims, secret)
}
```

### Frontend Embed Component

```svelte
<!-- Embed.svelte -->
<script lang="ts">
    import { onMount } from 'svelte';
    
    export let embedUrl: string;
    export let height: string = '800px';
    export let loading = true;
    
    let iframe: HTMLIFrameElement;
    
    onMount(() => {
        iframe.addEventListener('load', () => {
            loading = false;
        });
    });
</script>

<div class="embed-container" style="height: {height}">
    {#if loading}
        <div class="loading-skeleton">
            <div class="skeleton-bar"></div>
            <div class="skeleton-chart"></div>
        </div>
    {/if}
    
    <iframe
        bind:this={iframe}
        src={embedUrl}
        class="embed-iframe"
        class:hidden={loading}
        frameborder="0"
        allowfullscreen
    />
</div>
```

---

## UI/UX Design

### Design System

#### Color Palette

```css
/* Light Theme */
:root {
    --bg-primary: #ffffff;
    --bg-secondary: #f8fafc;
    --bg-tertiary: #f1f5f9;
    --bg-elevated: #ffffff;
    
    --text-primary: #0f172a;
    --text-secondary: #475569;
    --text-tertiary: #94a3b8;
    
    --accent-primary: #6366f1;
    --accent-secondary: #8b5cf6;
    --accent-gradient: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #a855f7 100%);
    
    --success: #10b981;
    --warning: #f59e0b;
    --error: #ef4444;
    --info: #3b82f6;
    
    --border: #e2e8f0;
    --border-hover: #cbd5e1;
    
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
    --shadow-glow: 0 0 20px rgb(99 102 241 / 0.3);
}

/* Dark Theme */
[data-theme="dark"] {
    --bg-primary: #0f172a;
    --bg-secondary: #1e293b;
    --bg-tertiary: #334155;
    --bg-elevated: #1e293b;
    
    --text-primary: #f8fafc;
    --text-secondary: #cbd5e1;
    --text-tertiary: #64748b;
    
    --accent-primary: #818cf8;
    --accent-secondary: #a78bfa;
    --accent-gradient: linear-gradient(135deg, #818cf8 0%, #a78bfa 50%, #c084fc 100%);
    
    --border: #334155;
    --border-hover: #475569;
    
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.3);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.4);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.4);
    --shadow-glow: 0 0 20px rgb(129 140 248 / 0.4);
}
```

#### Typography

```css
:root {
    --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
    --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
    
    --text-xs: 0.75rem;
    --text-sm: 0.875rem;
    --text-base: 1rem;
    --text-lg: 1.125rem;
    --text-xl: 1.25rem;
    --text-2xl: 1.5rem;
    --text-3xl: 1.875rem;
    --text-4xl: 2.25rem;
}
```

### Key UI Components

#### Sidebar Navigation

```
┌─────────────────────────────────────────────────────┐
│ ◉ DataPulse                              [🌙] [👤] │
├──────────────┬──────────────────────────────────────┤
│              │                                      │
│ 📊 Overview  │   Main Content Area                  │
│              │                                      │
│ 🗄️ Databases │   ┌────────────────────────────┐    │
│              │   │  Quick Stats                 │    │
│ 📈 Dashboards│   │  ┌────────┐ ┌────────┐     │    │
│              │   │  │  12    │ │  156   │     │    │
│ 🔍 Queries   │   │  │Databases│ │Reports │     │    │
│              │   │  └────────┘ └────────┘     │    │
│ 📋 Reports   │   │                              │    │
│              │   │   ┌─────────────────────┐   │    │
│ ⚙️ Settings  │   │   │   Recent Activity   │   │    │
│              │   │   └─────────────────────┘   │    │
│              │   │                              │    │
│              │   │   ┌─────────────────────┐   │    │
│              │   │   │   Popular Reports   │   │    │
│              │   │   └─────────────────────┘   │    │
│              │   │                              │    │
├──────────────┤   │                              │    │
│ [?] Help     │   └──────────────────────────────┘    │
└──────────────┴──────────────────────────────────────┘
```

#### Database Connection Form

```
┌─────────────────────────────────────────────────────┐
│  Add Database Connection                    [✕]     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Database Name                                      │
│  ┌─────────────────────────────────────────────┐   │
│  │ Production Database                         │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  Database Type                                      │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │
│  │  🐘     │ │  🐬     │ │  📊     │ │  🍃     │  │
│  │PostgreSQL│ │ MySQL   │ │Snowflake│ │ MongoDB │  │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘  │
│                                                     │
│  Host                    Port                       │
│  ┌──────────────────┐   ┌──────────┐              │
│  │ db.example.com   │   │ 5432     │              │
│  └──────────────────┘   └──────────┘              │
│                                                     │
│  Database Name                                      │
│  ┌─────────────────────────────────────────────┐   │
│  │ myapp_production                            │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  Username                  Password                 │
│  ┌──────────────────┐   ┌──────────────────┐      │
│  │ admin            │   │ ••••••••••       │      │
│  └──────────────────┘   └──────────────────┘      │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ 🔒 Use SSL/TLS                              │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │  Test Connection                            │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────┐  ┌─────────────────────┐  │
│  │      Cancel         │  │   Add Database →    │  │
│  └─────────────────────┘  └─────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## Deployment

### Docker Compose (Development)

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: deploy/Dockerfile
    ports:
      - "3000:3000"
    volumes:
      - ./metabase:/app/metabase
      - app-data:/app/data
    environment:
      - RUST_LOG=info
      - DATABASE_URL=sqlite:///app/data/datapulse.db
      - METABASE_JAR_PATH=/app/metabase/metabase.jar
      - METABASE_PORT=3001
      - JWT_SECRET=change-me-in-production
    depends_on:
      - metabase-db
    restart: unless-stopped

  metabase-db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=metabase
      - POSTGRES_USER=metabase
      - POSTGRES_PASSWORD=metabase-secret
    volumes:
      - metabase-data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  app-data:
  metabase-data:
```

### Production Dockerfile

```dockerfile
# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend
FROM rust:1.84-slim AS backend-builder
WORKDIR /app/backend
COPY backend/Cargo.* ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release
COPY backend/ ./
COPY --from=frontend-builder /app/frontend/build ./static
RUN cargo build --release

# Stage 3: Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    openjdk-17-jre-headless \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy backend
COPY --from=backend-builder /app/backend/target/release/datapulse /app/datapulse

# Create metabase directory
RUN mkdir -p /app/metabase /app/data

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:3000/api/v1/health || exit 1

EXPOSE 3000

CMD ["/app/datapulse"]
```

---

## Configuration

### Environment Variables

```bash
# Server
PORT=3000
HOST=0.0.0.0
RUST_LOG=info

# Application Database
DATABASE_URL=postgresql://user:pass@localhost:5432/datapulse
# Or for SQLite
# DATABASE_URL=sqlite:///var/lib/datapulse/datapulse.db

# JWT
JWT_SECRET=your-super-secret-key-change-me
JWT_EXPIRY=24h

# Metabase
METABASE_JAR_PATH=./metabase/metabase.jar
METABASE_DOWNLOAD_URL=https://downloads.metabase.com/latest/metabase.jar
METABASE_PORT=3001
METABASE_HOST=127.0.0.1
METABASE_DB_TYPE=postgres
METABASE_DB_URL=postgresql://metabase:metabase-secret@localhost:5432/metabase
METABASE_JVM_OPTS=-Xmx2g -Xms512m

# Embedding
METABASE_SITE_URL=http://localhost:3001
METABASE_SECRET_KEY=your-metabase-secret-key

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# SMTP (Optional)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=password
```

---

## Security Considerations

1. **Password Encryption**: All database passwords encrypted using AES-256-GCM
2. **JWT Tokens**: Short-lived access tokens (15min) + refresh tokens (24h)
3. **Metabase Secret**: Stored encrypted, never exposed to frontend
4. **CORS**: Strict origin validation
5. **Rate Limiting**: API rate limiting per user/IP
6. **Input Validation**: All inputs validated server-side
7. **SQL Injection**: Parameterized queries via SQLx
8. **XSS Prevention**: Content Security Policy headers
9. **HTTPS**: Enforced in production
10. **Audit Logging**: All admin actions logged

---

## Performance Optimizations

1. **Response Caching**: Metabase API responses cached (configurable TTL)
2. **Connection Pooling**: SQLx connection pool for app database
3. **Static Assets**: Frontend assets served with gzip/brotli
4. **Lazy Loading**: Dashboard widgets loaded on demand
5. **WebSocket**: Real-time updates for query execution
6. **CDN Ready**: Static embed endpoints CDN-cacheable

---

## Monitoring & Observability

1. **Health Endpoints**:
  - `GET /api/v1/health` - Basic health
  - `GET /api/v1/health/detailed` - Full system status
2. **Metrics** (Future):
  - Prometheus metrics endpoint
  - Request latency, error rates
  - Metabase process stats
3. **Logging**:
  - Structured JSON logs
  - Request/response logging
  - Metabase process output capture

---

## Development Workflow

### Prerequisites

- Rust 1.84+
- Node.js 22+
- Java 17+ (for Metabase)
- PostgreSQL (optional, SQLite works for dev)

### Backend Development

```bash
cd backend
cargo run              # Start with hot reload
cargo test             # Run tests
cargo clippy           # Lint
cargo fmt              # Format
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev            # Start dev server (port 5173)
npm run build          # Production build
npm run preview        # Preview production build
```

### Full Stack

```bash
# Root directory
npm run dev            # Starts both backend and frontend
npm run build          # Build everything
npm run start          # Start production server
```

---

## Roadmap

### Phase 1: MVP (Weeks 1-4)

- Rust backend with Axum
- Metabase process management
- First-run setup wizard
- Basic database connectivity
- Svelte frontend with auth
- Dashboard embedding
- Docker deployment

### Phase 2: Core Features (Weeks 5-8)

- Full database management UI
- Query builder integration
- User management & RBAC
- API key management
- Dark/light theme
- Responsive design
- Audit logging

### Phase 3: Advanced Features (Weeks 9-12)

- Scheduled reports
- Email notifications
- Webhook integrations
- Export functionality
- Advanced embedding options
- Performance dashboard
- Backup/restore

### Phase 4: Polish & Scale (Weeks 13-16)

- SSO integration
- Multi-tenancy support
- Prometheus metrics
- Kubernetes deployment
- API documentation
- Load testing
- Production hardening

---

## License

Proprietary - All rights reserved

---

*Design Document v1.0 - Created April 2026*