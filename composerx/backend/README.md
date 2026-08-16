# MergeEmailX Backend (Golang + Gin)

Backend service that powers MergeEmailX: templates, merge data, attachments, email jobs, and report orchestration.

## Prerequisites

- Go 1.21+ (Go modules enabled)
- MySQL 8+
- MongoDB 6+
- Redis 6+

## Environment configuration

The service reads the following environment variables, with sensible defaults taken from `design.md`:

```text
TRAN_MYSQL_DSN="root:Dafuq@911@tcp(127.0.0.1:3306)/tran?parseTime=true&charset=utf8mb4"
TRAN_MONGO_URI="mongodb://localhost:27017/"
TRAN_MONGO_DB="alterathena"
TRAN_REDIS_ADDR="127.0.0.1:6379"
TRAN_FILE_STORAGE_PATH="./storage"
PORT="8043"
```

You can override any of these in your shell or a process manager.

Copy the template and add AI settings:

```bash
cp .env.example .env
```

## AI provider configuration

ComposerX (TranMail) uses **MorphAI** for the platform assistant and email composer chat. It can also load optional JSON config and uses a **separate** OpenAI-compatible key for reference-library **embeddings** (RAG).

### Chat / assistant (Qwen via DashScope)

**Option A — environment (recommended)**

Add to `backend/.env`:

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
MORPH_AI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `MORPH_AI_API_KEY` | _(empty)_ | **Required** for LLM assistant + composer chat |
| `MORPH_AI_MODEL` | `qwen3-max` | Chat model |
| `MORPH_AI_BASE_URL` | DashScope compatible-mode `/v1` | OpenAI-compatible base URL for composer |

**Legacy fallbacks:** `TRAN_QWEN_API_KEY`, `TRAN_QWEN_MODEL`, `TRAN_QWEN_BASE_URL`.

**Option B — JSON file (optional)**

Copy `ai.config.example.json` → `ai.config.json` next to the binary (or set `TRAN_AI_CONFIG_PATH`):

```json
{
  "qwen_api_key": "sk-your-dashscope-key",
  "qwen_base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
  "qwen_default_model": "qwen3-max"
}
```

**Precedence:** `MORPH_AI_*` env vars override JSON; JSON fills in when env is unset.

The platform assistant also uses `github.com/robo/morphai` (`LoadFromEnv()`) for `POST /ai/assistant/chat`.

### Reference library RAG (optional, separate key)

Uploading PDFs/images for **AI reference docs** (`/ai/reference-docs`, `/ai/composer-chat`) uses **embedding** APIs when configured:

```bash
TRAN_OPENAI_API_KEY=sk-...   # OpenAI-compatible embedder for chunk vectors
```

If unset, reference upload/list may work but semantic RAG search in composer chat will report that AI is not configured.

### Restart and verify

```bash
go run .
```

Or from monorepo root: `./start-all.sh restart composerx-api`

1. Open `http://localhost:8044` (frontend).
2. Use the **AI assistant** drawer or **composer chat** on a new email.
3. Without keys you see guided rule-based help; with `MORPH_AI_API_KEY` you get full LLM replies.

**Endpoints:** `POST /ai/assistant/chat`, `POST /ai/composer-chat`, `POST /ai/reference-docs/upload`.

See also: [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](../../AI_ASSISTANT_MORPHAI_CONTRACT.md).

## 1. Prepare MySQL schema

From the `backend` directory:

```bash
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS tran CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -uroot -p tran < schema.sql
```

If you already applied an older `schema.sql`, run incremental migrations as needed, for example:

```bash
mysql -uroot -p tran < schema_migration_email_templates_meta.sql
```

That adds `tag`, `description`, and `builtin_key` on `email_templates` (used for built-in starter templates and AI-friendly metadata). Fresh installs from the current `schema.sql` include these columns already.

For more details see `README-db.md`.

## 2. Install Go dependencies

From the `backend` directory:

```bash
go mod tidy
```

This will download Gin, MySQL driver, Mongo driver, Redis client, and related packages.

## 3. Run the backend

From the `backend` directory:

```bash
go run ./...
```

The API will listen on:

```text
http://localhost:8043
```

You can verify it is healthy:

```bash
curl http://localhost:8043/health
```

## Saved emails (MySQL + MongoDB)

- **MySQL** (`saved_emails`, `saved_email_recipients`): name, recipient count/preview, `content_mongo_id`, timestamps.
- **MongoDB** collection `email_bodies`: document `_id` (ObjectId) + `html` (large body).

New installs: `schema.sql` includes these tables and nullable `email_jobs.template_id` / `email_jobs.saved_email_id`.

Existing DBs: run **`schema_migration_saved_emails.sql` once** (skip any `ALTER` that errors if already applied).

For Compose & Publish pages, run **`schema_migration_published_pages.sql` once** on existing environments.
For saved publish HTML drafts, run **`schema_migration_publish_drafts.sql` once** on existing environments.

## 4. Core endpoints

Once running, the main JSON APIs are:

- `GET /emails` — list index rows (no HTML)
- `POST /emails` — create (`name`, `html_content`, `recipients[]`, `created_by`)
- `GET /emails/{id}` — full detail including HTML from Mongo
- `PUT /emails/{id}` — update body + recipients
- `DELETE /emails/{id}` — delete SQL row + Mongo document
- `GET /contacts` — list address book (paginated)
- `POST /contacts` — create (`name`, `email`, `phone`, `note` ≤ 1000 chars)
- `PUT /contacts/{id}` — update
- `DELETE /contacts/{id}`
- `POST /contacts-batch-delete` or `POST /contacts/batch-delete` — JSON `{"ids":[1,2,3]}` (deduped, max 500 per request); response `{"deleted":n}` (same handler; use the flat URL if a proxy strips nested paths)
- `POST /contacts/import` — multipart `file` (CSV; headers: `email` required, `name`, `phone`/`phone_number`, `note`)
- `POST /contacts/sync-api` — JSON `{"url":"https://...","headers":{"Authorization":"Bearer …"}}` — server `GET`s JSON array or `{contacts|items|data|results:[]}` and upserts by email
- `GET /templates`
- `POST /templates`
- `PUT /templates/{id}`
- `DELETE /templates/{id}`
- `GET /merge-data`
- `POST /merge-data/upload` (multipart)
- `DELETE /merge-data/{id}`
- `GET /attachments`
- `POST /attachments/upload` (multipart)
- `DELETE /attachments/{id}`
- `POST /email/send` — body must include **exactly one** of `template_id` or `saved_email_id`, plus `recipients[]`
- `GET /email/job/{guid}`
- `GET /publishes/resolve-path?name=...` — resolve public slug (auth)
- `GET /publishes/history` — list published pages (auth)
- `POST /publishes` — publish HTML page and reserve public path (auth)
- `GET /publish-drafts` — list saved HTML drafts (auth)
- `POST /publish-drafts` — save HTML draft (auth)
- `GET /publish-drafts/{id}` — fetch one HTML draft (auth)
- `DELETE /publish-drafts/{id}` — delete one HTML draft (auth)
- `GET /public/published/{slug}` — fetch published page JSON (public)
- `GET /public/p/{slug}` — public HTML page endpoint (no login)
- `GET /reports/available`
- `POST /reports/order`
- `GET /reports/status/{guid}`

**Fixtures:** sample contacts live in `fixtures/` (`contacts_mock.csv`, `.sql`, `.json` — see `fixtures/README.md`).

These are consumed directly by the Svelte frontend.

