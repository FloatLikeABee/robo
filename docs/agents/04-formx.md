# 04 — FormsX (SheetX / SurveyX)

## Overview

FormsX is a forms platform with Survey Bot, Events & Info, public forms, and AI assistant. It uses MySQL (GORM) for form structure and MongoDB for responses/documents.

- **Module:** `github.com/formsx/backend` (Go 1.21)
- **Backend:** `formx/backend/` — Go + Gin + GORM
- **Frontend:** `formx/frontend/` — React 19 + Vite + TailwindCSS 4
- **Ports:** API `29909`, UI `19909`
- **Auth:** UsersPanel-dependent

## Backend architecture

### Entry point: `formx/backend/cmd/server/main.go`

1. `godotenv.Load(".env")` then `config.Load()`
2. Initialize MySQL (GORM), MongoDB (official driver)
3. Create all repository layers
4. Create `Handler` struct (includes MorphAI client)
5. Register all routes under `/api/v1`
6. Swagger UI at `/swagger/*any`
7. Static uploads at `/uploads`
8. Listen on `SERVER_PORT` (default `29909`)

### Package layout

```
formx/backend/
├── cmd/server/main.go          ← Entry point
├── internal/
│   ├── config/config.go        ← Env loading
│   ├── handler/                 ← All HTTP handlers
│   │   ├── handler.go           ← Handler struct + route registration
│   │   ├── assistant.go         ← Rule-based assistant
│   │   ├── assistant_llm.go     ← LLM tool-calling loop
│   │   ├── forms.go             ← Form CRUD
│   │   ├── questions.go         ← Question CRUD
│   │   ├── pages.go             ← Page CRUD
│   │   ├── rules.go             ← Visibility rules
│   │   ├── responses.go         ← Response list/export/delete
│   │   ├── public.go            ← Public form + submit
│   │   ├── events_info.go       ← Events & Info CRUD
│   │   ├── survey_bot_handlers.go ← Survey Bot CRUD + AI draft
│   │   ├── survey_bot_from_file.go ← Survey draft from an uploaded PDF or image
│   │   ├── survey_bot_from_file.go ← Survey draft from an uploaded PDF or image
│   │   ├── auth.go              ← Auth (proxy to UsersPanel)
│   │   ├── form_template_ai.go  ← AI form template generation
│   │   ├── mongodb_mcp.go       ← MCP tool gateway
│   │   └── app_abilities_mcp.go ← App abilities docs
│   ├── models/                  ← GORM + MongoDB models, DTOs
│   ├── mysql/                   ← GORM repos (Form, Page, Question, Rule)
│   ├── mongo/                   ← MongoDB repos (Response, EventInfo, AIDoc, SurveyBot)
│   ├── mail/smtp.go             ← Email broadcast
│   ├── surveybot/               ← Survey Bot template engine
│   └── pkg/validator/           ← Submission validation
├── go.mod
└── go.sum
```

### Handler struct

```go
type Handler struct {
    FormRepo              *mysql.FormRepo
    PageRepo              *mysql.PageRepo
    QuestionRepo          *mysql.QuestionRepo
    RuleRepo              *mysql.RuleRepo
    ResponseRepo          *mongo.ResponseRepo
    EventInfoRepo         *mongo.EventInfoRepo
    AIDocRepo             *mongo.AIDocumentRepo
    SurveyBotTemplateRepo *mongo.SurveyBotTemplateRepo
    SurveyBotResultRepo   *mongo.SurveyBotResultRepo
    AI                    *morphai.Client
}
```

### Route groups

| Group | Auth | Key endpoints |
|-------|------|---------------|
| `/api/v1/auth` | Public | `POST /login`, `GET /me` |
| `/api/v1/assistant` | Protected | `POST /chat` |
| `/api/v1/ai` | Protected | MCP tools, web search, form template AI, MongoDB MCP |
| `/api/v1/survey-bot` | Protected | Template CRUD, compile, publish, results, `POST /templates/from-file` |
| `/api/v1/forms` | Protected | Form/page/question/rule CRUD, responses |
| `/api/v1/events-info` | Protected | Events & Info CRUD, share, AI context |
| `/api/v1/public` | Public | `GET /forms/:slug`, `POST /forms/:slug/submit`, AI sheets |

### Database layer

| DB | ORM/Driver | Purpose |
|----|-----------|---------|
| MySQL | GORM | Forms, pages, questions, rules (structured) |
| MongoDB | Official driver | Form responses, Events & Info, AI documents, Survey Bot templates/results |

### AI assistant

Two modes:
1. **Rule-based** (`assistant.go`): Regex intent detection for create_form, create_event, list_forms, list_events. Extracts fields, asks follow-ups.
2. **LLM tool-calling** (`assistant_llm.go`): When MorphAI is configured, uses tool-calling loop (up to 8 rounds). Tools: list_forms, get_form, create_form, create_question, update_question, list_events, get_event, get_event_context, create_event, sync/list/search/get system documents, web_search, list/get survey results, list/get survey templates.

### Survey Bot

Markdown-based survey templates. Flow:
1. Create template (manually, AI draft via `/ai-draft`, or from an uploaded file via `/from-file`)
2. Compile markdown → UI blocks (`/compile`)
3. Publish (`/publish`)
4. Public AI sheet at `/s/:slug` — users interact with AI-powered survey
5. Results stored in MongoDB

#### Survey from a document or photo

`POST /api/v1/survey-bot/templates/from-file` takes a multipart upload with one
`file` part plus optional `title_hint` and `instructions` text fields, and returns
a validated Survey Bot markdown draft. It creates nothing: the caller saves through
the normal create route, so the user always reviews the draft first.

| Upload | How the content is read |
|--------|------------------------|
| PDF | Text layer extracted locally with `pkg/docextract` (no AI call) |
| JPEG, PNG, GIF, WebP | Transcribed by a vision model via `morphai.ChatCompletionVision` |
| MD, TXT | Rejected with guidance — the frontend loads these straight into the editor |

The generated markdown is validated with `surveybot.ParseMarkdown`. On failure the
parser's complaint is fed back to the model for one retry; a second failure returns
HTTP 422 with both the error and the rejected markdown so it can be repaired by hand.

Dependencies and failure modes:
- PDF text extraction needs no AI. A scan without a text layer returns a distinct
  "no text could be extracted" error rather than an empty draft.
- Image reading needs an OpenAI-compatible endpoint. With `MORPH_AI_API_URL` pointed
  at the native DashScope text-generation endpoint, image uploads return HTTP 503
  naming the setting to change. `MORPH_AI_VISION_MODEL` overrides the vision model
  (default `qwen-vl-max`).
- With no `MORPH_AI_API_KEY`, the endpoint returns HTTP 503 with an
  "AI is not configured" message.

## Frontend architecture

### Build & tooling
- Vite + React 19 + TypeScript 5.9 + TailwindCSS 4
- Port: 19909 (dev)
- Proxy: `/api` → `localhost:29909`, `/uploads` → `localhost:29909`

### Source structure (`formx/frontend/src/`)

| File/Dir | Purpose |
|----------|---------|
| `App.tsx` | Router: public routes + authenticated routes |
| `pages/` | `SurveyBot.tsx`, `EventsInfo.tsx`, `PublicForm.tsx`, `PublicAISheet.tsx`, `Login.tsx`, `EditForm.tsx`, `FormResults.tsx`, `MyForms.tsx`, `Settings.tsx` |
| `components/` | `Layout.tsx`, `FormTemplateAIPanel.tsx`, `EventInfoSubmitCard.tsx` |
| `lib/api.ts` | Centralized API client with typed methods, JWT management, 401 handling |

### API client pattern (`lib/api.ts`)

```typescript
const api = {
  auth: { login, me },
  forms: { list, get, create, update, delete },
  eventsInfo: { list, get, create, update, delete },
  surveyBot: { templates, compile, publish, results },
  publicAISheet: { get, chat },
  assistant: { chat },
  // ...
};
```

JWT stored in localStorage/sessionStorage/cookie. Axios/fetch interceptor adds Bearer token. On 401: clear tokens, dispatch `AUTH_EXPIRED_EVENT`, redirect to `/login`.

## How to add a new form feature

1. **Backend model:** Add GORM model in `internal/models/` or MongoDB model
2. **Repository:** Add repo in `internal/mysql/` or `internal/mongo/`
3. **Handler:** Add handler methods in `internal/handler/`
4. **Routes:** Register in `internal/handler/handler.go` `Register()` method
5. **Frontend API:** Add typed methods in `frontend/src/lib/api.ts`
6. **Frontend page:** Add page component in `frontend/src/pages/`
7. **Route:** Add route in `frontend/src/App.tsx`

## Key env vars

| Variable | Default | Purpose |
|----------|---------|---------|
| `SERVER_PORT` | `29909` | Backend port |
| `TRAN_MYSQL_DSN` | `root:Dafuq@911@tcp(127.0.0.1:3306)/tran?...` | MySQL DSN |
| `TRAN_MONGO_URI` | `mongodb://localhost:27017/` | MongoDB URI |
| `TRAN_MONGO_DB` | `athena` | MongoDB DB name |
| `UPLOAD_DIR` | `./uploads` | File upload directory |
| `USERS_PANEL_BASE_URL` | `http://127.0.0.1:9090` | UsersPanel URL |
| `MORPH_AI_API_KEY` | — | DashScope API key |
| `MORPH_AI_MODEL` | `qwen3-max` | AI model |