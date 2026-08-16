# Academi

AI-powered academic ecosystem for students and knowledge seekers. Blends AI-driven assistance, community collaboration, structured + unstructured learning resources, and task-oriented guidance.

Product vision: [`design.md`](design.md). Operational handoff: [`handoff.md`](handoff.md).

## Surfaces

| Surface | Path | Stack | Status |
|---------|------|-------|--------|
| **Web app** (primary dev focus) | `web/` | Vanilla JS ES modules, HTML5, CSS | Fully wired to backend |
| **Go API** | `backend/` | Go 1.24, Gin, BadgerDB | Auth, AI chat, docs, community, guides, sessions |
| **React Native app** | root (`App.js`, `screens/`, `components/`, …) | RN 0.72, Zustand, React Navigation | Scaffold / partial; lags behind web |

## Tech Stack

**Backend:**
- Go 1.24 with Gin web framework
- BadgerDB (embedded KV store at `backend/data/badger/`)
- JWT auth via `golang-jwt/jwt/v5`
- UUID generation, godotenv for config
- PDF text extraction via `ledongthuc/pdf`
- AI providers: SiliconFlow (primary, OpenAI-compatible), OpenAI, Anthropic, Azure, Ollama, Qwen

**Web frontend:**
- Vanilla JavaScript ES modules (`web/js/`)
- No build step; served via static HTTP server (Python `http.server`)
- Mobile-first responsive design, dark/light theme toggle
- Modules: `app.js` (shell), `chat.js`, `docs.js`, `community.js`, `api.js`, `markdown.js`

**React Native:**
- React 18.2, React Native 0.72
- Zustand for state management
- React Navigation (bottom tabs + native stack)
- SVG, markdown display, safe area context

## Building and Running

### Quick Start (recommended)

```bash
./start.sh
```

This builds the Go backend binary and serves the web frontend. Both shut down on Ctrl+C.

- Frontend: http://localhost:8765
- Backend health: http://localhost:8978/health
- API base: `http://<host>:8978/api/v1`

**Requirements:** Go 1.21+, Python 3 (for static file serving with ES module support).

**First run:** `start.sh` copies `backend/.env.example` → `backend/.env` if missing. Add your AI API key before chat works.

### Environment (`backend/.env`)

Gitignored. Template at `backend/.env.example`.

| Variable | Purpose |
|----------|---------|
| `SERVER_PORT` | Backend port (default `8978`) |
| `SERVER_MODE` | Gin mode (`debug` / `release`) |
| `DB_PATH` | BadgerDB path (default `./data/badger`) |
| `JWT_SECRET` | JWT signing secret |
| `AI_DEFAULT_PROVIDER` | `siliconflow` (recommended), `openai`, `anthropic`, `azure`, `ollama`, `qwen` |
| `SILICONFLOW_API_KEY` | API key from [SiliconFlow](https://cloud.siliconflow.cn) |
| `SILICONFLOW_BASE_URL` | `https://api.siliconflow.cn/v1` |
| `SILICONFLOW_MODEL` | e.g. `deepseek-ai/DeepSeek-V3.2` |
| `OPENAI_*` | Fallback if SiliconFlow vars unset |
| `USERS_PANEL_BASE_URL` | UsersPanel/Morph auth endpoint |

**SiliconFlow** uses OpenAI-compatible `POST /v1/chat/completions` with `Authorization: Bearer <key>`. Streaming is not supported.

### Running Individual Surfaces

```bash
# Backend only
cd backend && go build -o bin/academi-backend ./cmd/main.go && ./bin/academi-backend

# Web only (needs backend running)
cd web && python3 -m http.server 8765 --bind 0.0.0.0

# React Native
npm install
npm run android   # or npm run ios
```

### Testing

```bash
# Backend unit tests
cd backend && go test ./internal/ai/ -count=1

# All backend tests
cd backend && go test ./... -count=1

# React Native tests
npm test
```

### Smoke Test (curl)

```bash
curl -s http://localhost:8978/health

# Auth + chat
TOKEN=$(curl -s -X POST http://localhost:8978/api/v1/auth/mock | jq -r .token)
curl -s -X POST http://localhost:8978/api/v1/ai/chat \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"hi"}],"context":{}}'
```

## Architecture

```
web/index.html
  └── js/main.js          # bootstrap, viewport
        └── js/app.js     # AcademiApp shell, navigation, theme
              ├── js/chat.js       # chat, sessions, AI calls, uploads
              ├── js/docs.js       # docs grid, modals, learn flow
              ├── js/community.js  # feed, posts, votes
              ├── js/api.js        # base URL resolution, debounce, auth headers
              └── js/markdown.js   # escapeHtml, markdownToHtml

backend/cmd/main.go       # Gin router, CORS, service registration
backend/internal/
  ai/          # chat, learn, prompts, message normalization
  auth/        # JWT, UsersPanel SSO, mock session for web demo
  docs/        # upload, extract, CRUD
  community/   # posts, comments, votes
  guide/       # subjects + guides
  chatsessions/# persisted chat threads (BadgerDB)
  research/    # Wikipedia/web/arXiv snippets
  notifications/# notification service
  database/    # BadgerDB wrapper
```

**Web ↔ API:** Mock auth on load (`POST /auth/mock`). Chat uses `POST /ai/chat` with `messages`, `doc_ids`, `document_mode`, `disable_research`, `help_you_learn`.

**Auth:** UsersPanel/Morph SSO via `?userspanel_token=…` query param → exchanged via `POST /api/v1/auth/dev-login`. Falls back to mock session for local dev.

**Data:** BadgerDB at `backend/data/badger/` (gitignored). Uploads at `backend/data/uploads/`.

## API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/health` | No | Health check |
| POST | `/api/v1/auth/mock` | No | Demo session token |
| POST | `/api/v1/auth/dev-login` | No | UsersPanel SSO exchange |
| GET | `/api/v1/ai/providers` | No | List configured AI providers |
| POST | `/api/v1/ai/chat` | Optional | AI chat completion |
| POST | `/api/v1/ai/learn` | Optional | Learning guide generation |
| POST | `/api/v1/ai/summarize` | Optional | Content summarization |
| POST | `/api/v1/ai/generate-guide` | Optional | Step-by-step guide generation |
| POST | `/api/v1/ai/moderate` | Optional | Content moderation |
| GET/POST | `/api/v1/docs` | Bearer | Document list/create |
| POST | `/api/v1/docs/upload` | Bearer | File upload |
| GET/POST | `/api/v1/community/posts` | Bearer | Community feed |
| GET/POST/PATCH/DELETE | `/api/v1/chat-sessions` | Bearer | Chat session CRUD |
| GET/POST | `/api/v1/guides/*` | Bearer | Learning guides |
| GET/POST | `/api/v1/notifications` | Bearer | Notifications |

## Design System

- **Dark mode default:** `#0A0F1C` background
- **Gradients:** `#5B8CFF` → `#9B6DFF` → `#00D4FF` (neon blue/purple/cyan)
- **Style:** Glassmorphism, blurred surfaces, rounded corners (2xl+), micro-interactions
- **Typography:** Inter / SF Pro; Display 32px bold, H1 24px semibold, Body 16px regular
- **Spacing:** 4pt base system; tokens at 4/8/12/16/24/32px
- **Mobile width:** 390px base (iPhone); margin 16px
- **Light mode:** `#F7F9FC` background, reduced glow, more clarity

## Key Conventions

- **Web JS:** ES modules (no bundler); `web/js/app.js` is the shell, feature modules mixed onto `AcademiApp.prototype` via `Object.assign`.
- **Backend Go:** Service-per-domain pattern; each `internal/<domain>/` has `NewService()` and `RegisterRoutes()`.
- **Config:** Singleton loaded once via `sync.Once` from env vars + `.env` file. Use `firstEnv()` / `envDefault()` helpers for multi-key fallback.
- **Message normalization:** SiliconFlow requires exactly one system message at index 0; `normalizeChatMessages()` in `backend/internal/ai/messages_normalize.go` merges doc + research prompts before the API call.
- **CORS:** Permissive (any origin) for local dev; `AllowCredentials: true` with origin echo.
- **Git:** Do not commit `backend/.env`, `backend/data/`, or API keys.

## Known Quirks

| Symptom | Cause | Fix |
|---------|-------|-----|
| `AI provider 'siliconflow' not found` | Provider not registered in config | Check `backend/internal/config/config.go` |
| `401 Invalid token` after renewing key | `envOr` returned literal string instead of env value | `firstEnv()` in `config.go` |
| `400 System message must be at the beginning` | Multiple system roles | `messages_normalize.go` merges them |
| Chat works on LAN phone but not desktop | `localhost` on phone = phone itself | Set `localStorage.academi_api_base` or use machine IP |
| ES modules fail from `file://` | Browsers block module imports over file protocol | Must use HTTP server (`./start.sh`) |

## Project Structure

```
academi/
├── start.sh              # Dev launcher (backend + web)
├── App.js                # React Native entry
├── index.js              # RN bootstrap
├── package.json          # RN dependencies
├── design.md             # Product vision + UI spec
├── handoff.md            # Operational context for agents
├── backend/
│   ├── cmd/main.go       # Go entry point
│   ├── internal/         # Domain services
│   ├── go.mod / go.sum   # Go modules
│   ├── .env.example      # Environment template
│   └── data/             # Runtime data (gitignored)
├── web/
│   ├── index.html        # Web app entry
│   ├── styles.css        # Global styles
│   └── js/               # ES module source
├── screens/              # RN screen components
├── components/           # RN reusable UI components
├── services/             # RN API service
├── store/                # RN Zustand store
└── themes/               # RN theme config
```

## Next Steps (from handoff.md)

1. Share button handler on AI messages (web)
2. Profile stats from API (currently hardcoded)
3. React Native app parity with web
4. Streaming chat support
5. Service worker / PWA
6. README update for web module layout
