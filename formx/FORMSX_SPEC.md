# SheetX (SheetX) — Microservice Specification & Build Prompt

## Copy-paste prompt (condensed)

Build a microservice **SheetX** in **Golang** with **Swagger**. Users create forms (name, description, published URL slug, “single response only” option) and add questions with types: **text**, **select**, **multiselect**, **boolean**, **image upload**, **document/file upload**. Store **forms and questions in MySQL**; store **form results (user answers) in MongoDB**. Use env: `TRAN_MYSQL_DSN`, `TRAN_MONGO_URI`, `TRAN_MONGO_DB`, `TRAN_REDIS_ADDR` with defaults: `root:Dafuq@911@tcp(127.0.0.1:3306)/tran?parseTime=true&charset=utf8mb4`, `mongodb://localhost:27017/`, `athena`, `127.0.0.1:6379`. Enforce single response per respondent when “single response only” is set. Expose REST APIs for form/question CRUD, public form by slug, submit response, and list/export responses. Implement full spec below.

---

SheetX is a form/sheet-builder microservice. Users create forms with configurable questions (multiple data types), publish them via URL, and collect responses. Form metadata lives in MySQL; response payloads live in MongoDB.

---

## 1. Overview

- **Name:** SheetX (legacy code name FormsX)  
- **Role:** Form creation, publishing, and response collection  
- **Backend:** Go (Golang), with **Swagger** (OpenAPI) for all HTTP APIs  
- **Databases:**  
  - **MySQL:** Forms and questions (schema, config).  
  - **MongoDB:** Form results (user-filled answers).  
  - **Redis:** Optional (e.g. rate limiting, single-response checks, caching).

---

## 2. Environment & DB Configuration

Use these env vars (with fallbacks as below):

| Variable           | Default | Purpose |
|--------------------|---------|--------|
| `TRAN_MYSQL_DSN`   | `root:Dafuq@911@tcp(127.0.0.1:3306)/tran?parseTime=true&charset=utf8mb4` | MySQL connection |
| `TRAN_MONGO_URI`   | `mongodb://localhost:27017/` | MongoDB connection URI |
| `TRAN_MONGO_DB`    | `athena` | MongoDB database name |
| `TRAN_REDIS_ADDR`  | `127.0.0.1:6379` | Redis address (optional) |

Helper: `getEnv(key, defaultVal string) string` — return env value or default.

---

## 3. Data Models

### 3.1 MySQL — Forms & Questions

**Form (table: `forms`)**

| Field              | Type         | Notes |
|--------------------|-------------|-------|
| id                 | BIGINT PK   | Auto-increment |
| name               | VARCHAR(255)| Form display name |
| description        | TEXT        | Optional description |
| slug               | VARCHAR(255)| Unique URL slug (e.g. `my-survey-2024`) — used in published URL |
| single_response_only | TINYINT(1) | 1 = one response per respondent (e.g. by identifier), 0 = multiple allowed |
| created_at         | DATETIME    | |
| updated_at         | DATETIME    | |

- **Unique constraint:** `slug`.
- Published form URL format: `{base_url}/forms/{slug}` or `{base_url}/f/{slug}`.

**Question (table: `questions`)**

| Field        | Type         | Notes |
|-------------|-------------|-------|
| id          | BIGINT PK   | Auto-increment |
| form_id     | BIGINT FK   | References `forms.id` |
| title       | VARCHAR(512)| Question text (e.g. "What's your name?") |
| type        | VARCHAR(32) | Question/answer data type (see 3.3) |
| required    | TINYINT(1)  | 1 = required, 0 = optional |
| sort_order  | INT         | Display order (ascending) |
| config      | JSON        | Type-specific config (options for select, file constraints, etc.) |
| created_at  | DATETIME    | |
| updated_at  | DATETIME    | |

- **Index:** `form_id`, `sort_order`.

### 3.2 Question / Answer Data Types

Support at least these **types** (stored in `questions.type` and used when validating/storing answers):

| Type        | Description | Example / Config |
|-------------|-------------|-------------------|
| `text`      | Short or long text | Single-line or paragraph; config: `multiline` bool |
| `integer`   | Integer (e.g. selector value) | Male=1, Female=2; config: options `[{value, label}]` or min/max |
| `select`    | Single choice from options | Same as integer with labels; config: `options` |
| `multiselect` | Multiple choices | config: `options`, max_selections |
| `float`     | Decimal number | Optional min/max in config |
| `boolean`   | Yes/No | — |
| `date`      | Date only | — |
| `datetime`  | Date and time | — |
| `image`     | Image upload | config: max_size, allowed_mime (e.g. image/*) |
| `document`  | Document upload | config: max_size, allowed_extensions or mime |
| `qrcode`    | QR code input (user scans or enters code) | Store as string; config: optional validation pattern |

- **File types (`image`, `document`):** Store in object storage or filesystem; in MongoDB store reference (path/URL) and metadata (filename, size, mime).

### 3.3 MongoDB — Form Results (Responses)

**Collection:** e.g. `form_responses` (or `responses`).

**Document shape (flexible; adapt to your conventions):**

```json
{
  "_id": "ObjectId",
  "form_id": 123,
  "slug": "my-survey-2024",
  "respondent_id": "optional-identifier-for-single-response",
  "submitted_at": "ISODate",
  "answers": [
    {
      "question_id": 1,
      "type": "text",
      "value": "John Doe"
    },
    {
      "question_id": 2,
      "type": "integer",
      "value": 1
    },
    {
      "question_id": 3,
      "type": "image",
      "value": "https://storage/path/to/image.jpg",
      "filename": "photo.jpg",
      "size": 102400
    }
  ]
}
```

- **Indexes:** `form_id`, `slug`, `respondent_id` (for single-response check), `submitted_at`.
- For **single response only:** before inserting, check (e.g. by `form_id` + `respondent_id` or IP/token) that no document exists; use Redis or MongoDB query.

---

## 4. API Design (REST + Swagger)

Implement in Go and expose **Swagger UI** (e.g. `/swagger/`) and **OpenAPI JSON/YAML** (e.g. `/swagger/doc.json`).

### 4.1 Form Management (MySQL)

- **POST   /api/v1/forms**  
  Create form: body `{ "name", "description", "slug", "single_response_only" }`.  
  Return created form (id, slug, etc.).

- **GET    /api/v1/forms**  
  List forms (paginated): query `page`, `limit`, optional `search`.

- **GET    /api/v1/forms/:id**  
  Get one form by ID.

- **GET    /api/v1/forms/by-slug/:slug**  
  Get form by slug (for public form view).

- **PUT    /api/v1/forms/:id**  
  Update form (name, description, single_response_only; slug update optional and must stay unique).

- **DELETE /api/v1/forms/:id**  
  Soft-delete or hard-delete form (and optionally cascade questions).

### 4.2 Question Management (MySQL)

- **POST   /api/v1/forms/:formId/questions**  
  Add question: body `{ "title", "type", "required", "sort_order", "config" }`.

- **GET    /api/v1/forms/:formId/questions**  
  List questions for form (ordered by `sort_order`).

- **PUT    /api/v1/forms/:formId/questions/:id**  
  Update question.

- **DELETE /api/v1/forms/:formId/questions/:id**  
  Delete question.

### 4.3 Public Form & Submissions (MongoDB)

- **GET    /api/v1/public/forms/:slug**  
  Get form + questions by slug (for rendering the published form). No auth required (or optional).

- **POST   /api/v1/public/forms/:slug/submit**  
  Submit response: body `{ "respondent_id": "optional", "answers": [ { "question_id", "value" } ] }`.  
  If form is `single_response_only`, check `respondent_id` (or fallback) and reject duplicate.  
  Validate types and required fields; store in MongoDB. For file types, accept uploads (multipart) or pre-uploaded URLs.

- **POST   /api/v1/public/forms/:slug/upload** (optional)  
  Upload file for an `image` or `document` question; return reference to use in `answers[].value`.

### 4.4 Response Retrieval (MongoDB)

- **GET    /api/v1/forms/:formId/responses**  
  List responses (paginated): query `page`, `limit`, optional `since`, `until`.  
  Return documents from MongoDB.

- **GET    /api/v1/forms/:formId/responses/export**  
  Export as CSV or JSON (optional).

---

## 5. Backend Implementation (Golang)

- **Structure:**  
  - `cmd/server` — main, config loading (env), server start.  
  - `internal/config` — DSN, Mongo URI/DB, Redis.  
  - `internal/mysql` — form and question repo (CRUD).  
  - `internal/mongo` — response repo (insert, list, single-response check).  
  - `internal/redis` — optional: rate limit, single-response lock.  
  - `internal/handler` — HTTP handlers for forms, questions, public submit, responses.  
  - `internal/models` — form, question, request/response DTOs.  
  - `pkg/validator` — validate question types and required fields before saving to MongoDB.

- **Swagger:**  
  - Use **swaggo/swag** (go-swagger) or **swaggo/gin-swagger** if using Gin.  
  - Annotate handlers with comments; generate `docs/` and serve `/swagger/` and `/swagger/doc.json`.  
  - All endpoints must be documented in OpenAPI.

- **DB:**  
  - MySQL: use `database/sql` with a driver (e.g. `go-sql-driver/mysql`) or GORM.  
  - MongoDB: official `go.mongodb.org/mongo-driver`.  
  - Redis: optional `go-redis/redis/v8`.

- **Routing:** Use **Gin** or **Chi**; middleware for CORS, logging, recovery.

---

## 6. Single-Response-Only Behavior

- When `single_response_only` is true:  
  - Require a stable `respondent_id` (e.g. email, user id, or anonymous token from cookie).  
  - Before inserting into MongoDB, check if a response already exists for `form_id` + `respondent_id`.  
  - If exists, return HTTP 409 or 400 with message "Already responded."  
  - Optionally use Redis key `form:{form_id}:respondent:{respondent_id}` for fast check and lock.

---

## 7. File Uploads (Image / Document)

- Accept multipart form or separate upload endpoint; validate size and type from question `config`.  
- Store files in local disk or S3-compatible storage; in MongoDB store only reference (path/URL) and metadata in the corresponding answer `value` (or nested object).

---

## 8. Summary Checklist

- [ ] Go backend with clear layout (cmd, internal, pkg).  
- [ ] MySQL: tables `forms`, `questions` with migrations or auto-migrate.  
- [ ] MongoDB: collection for responses; indexes on form_id, respondent_id, submitted_at.  
- [ ] All question types: text, integer, select, multiselect, float, boolean, date, datetime, image, document, qrcode.  
- [ ] Form fields: name, description, URL slug, single_response_only.  
- [ ] REST APIs for form/question CRUD, public form by slug, submit response, list/export responses.  
- [ ] Swagger generated and served; every endpoint documented.  
- [ ] Env: TRAN_MYSQL_DSN, TRAN_MONGO_URI, TRAN_MONGO_DB, TRAN_REDIS_ADDR with given defaults.  
- [ ] Single-response enforcement when enabled.  
- [ ] File upload handling for image and document types.

Use this spec as the single source of truth to implement FormsX.
