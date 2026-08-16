# Users Panel

A user authentication and authorization admin stack: a **Rust (Axum)** API with **MySQL**, and a **Svelte** admin UI (Vite). Roles, JWT sessions, optional Google OAuth, and an admin screen to manage user roles.

---

## Repository layout

| Path | Purpose |
|------|---------|
| `backend/` | HTTP API (`users-panel-api`), SQL migrations (MySQL) |
| `admin/` | Admin SPA (Vite + Svelte 5) |
| `docs/microservices-and-api.md` | OpenAPI / Swagger, microservice patterns, users & roles & permissions flows |
| `design.md` | Original system design notes |

---

## Prerequisites

- **Rust** (latest stable) with `cargo`
- **MySQL** 8+ (or compatible), database created (e.g. `tran`) and a user that can run migrations
- **Node.js** 20+ and **npm** (for the admin UI)

---

## Local development

### 1. API server

From the repo root:

```bash
cd backend
```

Optional: create `backend/.env` (or export variables in your shell). Important variables are listed in [Environment variables](#environment-variables) below.

Run:

```bash
cargo run
```

By default the API listens on **`http://127.0.0.1:5001`**. On first start it applies SQL migrations to the MySQL database from **`DATABASE_URL`**.

**OpenAPI (Swagger):** with the server running, open **`http://127.0.0.1:5001/swagger-ui`** and use **`http://127.0.0.1:5001/openapi.json`** to generate clients. In local dev, Vite also proxies these paths, e.g. **`http://localhost:5173/swagger-ui`**. See **`docs/microservices-and-api.md`** for microservice-style integration and for adding users, roles, and permissions.

**Connection string (sqlx / this project):** use a URL, not the Go DSN style.

| Go-style DSN | Equivalent `DATABASE_URL` for Rust (sqlx) |
|--------------|---------------------------------------------|
| `root:PASSWORD@tcp(127.0.0.1:3306)/tran?parseTime=true&charset=utf8mb4` | `mysql://root:PASSWORD@127.0.0.1:3306/tran?charset=utf8mb4` |

If the password contains `@`, URL-encode it as **`%40`** (e.g. `secret@911` → `secret%40911`).

Example:

```bash
export DATABASE_URL='mysql://root:Dafuq%40911@127.0.0.1:3306/tran?charset=utf8mb4'
```

The default in code (if `DATABASE_URL` is unset) is a **dev-only** URL: user `root`, password `Dafuq@911`, host `127.0.0.1:3306`, database `tran` (password encoded as `Dafuq%40911` in the URL). Set **`DATABASE_URL`** in production and do not commit real credentials there if the repo is shared.

**Optional first admin user:** if `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` are **not** set, the server uses a **development default** (unless `USERS_PANEL_NO_DEFAULT_ADMIN=1`):

- Email: **`admin@example.com`**
- Password: **`AdminExample2026!`**

If no user with that email exists, the server creates a verified **Admin** user once. TranMail, TranForm, TranDemo (Morph login), and the Users Panel admin UI all use this same account when they talk to this API.

Set `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` together to override, or set `USERS_PANEL_NO_DEFAULT_ADMIN=1` to skip auto-bootstrap entirely.

### 2. Admin UI

In a second terminal:

```bash
cd admin
npm install
npm run dev
```

Vite usually serves the app at **`http://localhost:5173`**. It proxies **`/api`**, **`/swagger-ui`**, and **`/openapi.json`** to **`http://127.0.0.1:5001`**, so the browser can call the API and load Swagger on the Vite port without CORS issues during development.

Open the app, sign in (or register); you land on **Users**. Admins can open **Permissions** (permission slugs + per-role toggles) and **Roles** from the nav. **Admin** role is required for Users, Permissions, and Roles.

Migrations create **`plat_users`**, **`plat_roles`** (canonical role names), **`plat_permissions`**, **`plat_role_permissions`** (with FK to `plat_roles` and `plat_permissions`), and SQLx’s **`_sqlx_migrations`** table.

**Admin APIs (JWT + Admin role):** user list; **PATCH/DELETE** `/api/admin/users/:id` (edit email/username; delete — not yourself); **PATCH** `/api/admin/users/:id/roles`; **CRUD** `/api/admin/roles` and `/api/admin/roles/:id`; permissions CRUD as before.

---

## Production build

### API (release binary)

```bash
cd backend
cargo build --release
```

The binary is `backend/target/release/users-panel-api`. Run it on your server with the same environment variables as in development. Point **`DATABASE_URL`** at your MySQL instance and ensure the database exists before the first run (migrations create tables).

### Admin (static files)

```bash
cd admin
npm ci
npm run build
```

Static output is in **`admin/dist/`**. Serve that folder with any static host (nginx, Caddy, S3 + CloudFront, etc.).

---

## Deployment patterns

### A. Single host (reverse proxy)

Typical setup:

1. Run the API process (e.g. systemd) listening on `127.0.0.1:5001` or a Unix socket behind the proxy.
2. Serve `admin/dist/` at your public site (e.g. `https://panel.example.com`).
3. Configure the proxy so `/api` is forwarded to the Rust service (same origin avoids CORS).

Set **`FRONTEND_ORIGIN`** to your public admin URL (e.g. `https://panel.example.com`) and **`API_PUBLIC_URL`** to the public API base if it differs from the default `http://HOST:PORT` (used in verification/reset links in logs).

### B. Separate API and admin origins

If the admin is on another origin than the API:

- Set **`FRONTEND_ORIGIN`** to the admin’s origin (CORS).
- Build the admin with **`VITE_API_ORIGIN`** pointing at the public API origin, e.g. `https://api.example.com`, so browser requests go to the correct host:

```bash
cd admin
VITE_API_ORIGIN=https://api.example.com npm run build
```

Or put `VITE_API_ORIGIN=...` in `admin/.env.production` before `npm run build`.

---

## Environment variables

### API (`backend`)

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *(dev default below)* | MySQL URL; override in production |
| `JWT_SECRET` | *(dev default)* | **Required in production** — signing key for JWTs |
| `JWT_EXPIRY_HOURS` | `1` | Access token lifetime |
| `HOST` | `127.0.0.1` | Bind address |
| `PORT` | `5001` | Listen port |
| `FRONTEND_ORIGIN` | `http://localhost:5173` | Allowed CORS origin (admin URL) |
| `API_PUBLIC_URL` | `http://{HOST}:{PORT}` | Base URL for verify/reset links (logged for dev) |
| `GOOGLE_CLIENT_ID` | — | Enable Google OAuth if set with secret |
| `GOOGLE_CLIENT_SECRET` | — | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | `http://{HOST}:{PORT}/api/auth/google/callback` | Must match Google Cloud console |
| `BOOTSTRAP_ADMIN_EMAIL` | *(see below)* | Overrides dev default `admin@example.com` when set **with** password |
| `BOOTSTRAP_ADMIN_PASSWORD` | *(see below)* | Overrides dev default when set **with** email |
| `USERS_PANEL_NO_DEFAULT_ADMIN` | — | Set `1` to disable automatic dev bootstrap admin |
| `MORPH_AI_API_KEY` | — | DashScope API key for LLM assistant + Data Collector AI mapping |
| `MORPH_AI_MODEL` | `qwen3-max` | Chat / mapping model |
| `MORPH_AI_API_URL` | DashScope text-generation URL | Optional native endpoint override |

When neither `BOOTSTRAP_*` is set, the API uses **`admin@example.com` / `AdminExample2026!`** once (until you set `USERS_PANEL_NO_DEFAULT_ADMIN=1`).

Email sending is not wired to SMTP yet; verification and password-reset URLs are printed to the server logs for development.

**MorphAI legacy fallbacks:** `GEMINI_API_KEY`, `GEMINI_MODEL`. See [AI provider configuration](#ai-provider-configuration) below.

### Admin (`admin`)

| Variable | When | Description |
|----------|------|-------------|
| `VITE_API_ORIGIN` | Build / dev | Public API origin (empty in dev uses Vite proxy for `/api`) |

---

## AI provider configuration

UsersPanel uses **MorphAI** ([`pkg/morphai-rs`](../../pkg/morphai-rs)) for:

- The **admin AI assistant** drawer (`POST /api/assistant/chat`)
- **Data Collector** column mapping when uploaded CSV/JSON headers do not match the strict template ([`docs/data-collector.md`](docs/data-collector.md))

Without `MORPH_AI_API_KEY`, the assistant returns **rule-based** help and Data Collector requires **exact** template headers.

### 1. Get a DashScope API key

Create a key at [Alibaba Cloud DashScope](https://dashscope.aliyun.com/) (Qwen text generation).

### 2. Configure the API

Create or edit **`backend/.env`**:

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
# optional:
# MORPH_AI_API_URL=https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
```

Do not commit real keys. Restart after changes:

```bash
cd backend && cargo run
```

Or: `./start-all.sh restart userspanel-api`

### 3. Verify

**Assistant:** sign in to the admin UI (`http://localhost:5173`), open the AI drawer, ask e.g. `list roles`.

**Data Collector:** Admin → Data Collector → upload a file with non-standard headers; import should call AI to map columns (see server logs if mapping fails).

See also: [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](../../AI_ASSISTANT_MORPHAI_CONTRACT.md).

---

## Google OAuth

1. Create OAuth credentials in Google Cloud Console (Web application).
2. Add authorized redirect URI: **`{API_PUBLIC_URL or your public API}/api/auth/google/callback`** (must match `GOOGLE_REDIRECT_URL` if you override it).
3. Set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` for the API process.

---

## Health check

The API does not expose a dedicated `/health` route in this repo; you can use **`GET /api/auth/roles`** with a valid `Authorization` header, or add a small health route later. For a bare TCP check, ensure the configured **PORT** accepts connections.

---

## Troubleshooting

### `Error: Dirty(20260412000002)` (or another version)

A migration **failed partway through**. SQLx records that in **`_sqlx_migrations`** with **`success = false`** for that version and refuses to run further migrations until you fix it.

1. Use the **current** migration files from this repo (the duplicate `idx_prp_role` issue is fixed in `20260412000002`).
2. In MySQL, **remove only the failed migration row** (dev / after you are sure the schema matches what you expect):

```sql
DELETE FROM _sqlx_migrations WHERE version = 20260412000002 AND success = 0;
```

If nothing was applied for that version, you can use `DELETE FROM _sqlx_migrations WHERE version = 20260412000002;` instead.

3. From `backend/`, run:

```bash
sqlx migrate run
```

(or start the API if your app runs migrations on boot).

If tables from that migration already exist, the SQL uses `CREATE TABLE IF NOT EXISTS` and `INSERT IGNORE`, so re-running is usually safe. If you had a half-applied state you do not trust, fix or drop the affected tables first, then delete the migration row and run again.

### `Error: VersionMismatch(…)` from sqlx migrations

This happens when a migration file was **edited after it already ran** once. SQLx stores a checksum of the file; if the file changes, startup fails.

**Option A — repair checksums (recommended)**  
Install [sqlx-cli](https://crates.io/crates/sqlx-cli) matching your sqlx version (e.g. `0.8.x`), set `DATABASE_URL`, then from `backend/`:

```bash
sqlx migrate repair
```

**Option B — re-apply the migration record (dev only)**  
In MySQL, remove the row for that version so the migration runs again on next start:

```sql
DELETE FROM _sqlx_migrations WHERE version = 20260412000001;
```

Then run `cargo run` again. This is safe if the migration uses `CREATE TABLE IF NOT EXISTS` and you are okay re-running that SQL. If you still have an old `users` table from an earlier schema, rename or drop it after moving data.

### `JWT_SECRET not set; using default`

Set a secret for anything beyond quick local tests, e.g. in `backend/.env`:

```bash
JWT_SECRET=your-long-random-string
```

---

## Security notes for production

- Use a long random **`JWT_SECRET`**.
- Protect **`DATABASE_URL`** (credentials) and back up MySQL on your usual schedule.
- Prefer HTTPS everywhere and lock down **`FRONTEND_ORIGIN`** to your real admin URL.
