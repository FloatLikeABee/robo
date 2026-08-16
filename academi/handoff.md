# Academi — Handoff for Cursor CLI

Last updated: 2026-06-06  
Purpose: Context for continuing work in **Cursor CLI** (or any agent) without relying on prior chat history.

---

## What this project is

**Academi** is an AI-powered study assistant: chat, docs, community, and guided learning. The product vision lives in [`design.md`](design.md).

There are **three surfaces**:

| Surface | Path | Status |
|---------|------|--------|
| **Web app (primary dev focus)** | `web/` | Fully wired to backend; ES modules; mobile-first UI |
| **Go API** | `backend/` | Auth, AI chat, docs, community, guides, chat sessions |
| **React Native app** | root (`App.js`, `screens/`, …) | Scaffold / partial; not the main loop for recent work |

**Day-to-day development** = `./start.sh` → frontend on **8765**, backend on **8978** (from `backend/.env`).

---

## Quick start

```bash
# From repo root
./start.sh
```

- Frontend: http://localhost:8765  
- Backend health: http://localhost:8978/health  
- API base (web): `http://<host>:8978/api/v1` (auto-detected; see `web/js/api.js`)

**Requirements:** Go 1.21+, Python 3 (static server for ES modules).

**First run:** `start.sh` copies `backend/.env.example` → `backend/.env` if missing. **Add your AI API key before chat works.**

```bash
# Stop both processes
Ctrl+C
```

---

## Environment (`backend/.env`)

Gitignored. Template: [`backend/.env.example`](backend/.env.example).

| Variable | Notes |
|----------|--------|
| `SERVER_PORT` | Default `8978`; must match `web/index.html` meta `academi-api-port` |
| `AI_DEFAULT_PROVIDER` | `siliconflow` recommended for current setup |
| `SILICONFLOW_API_KEY` | From [SiliconFlow](https://cloud.siliconflow.cn) |
| `SILICONFLOW_BASE_URL` | `https://api.siliconflow.cn/v1` |
| `SILICONFLOW_MODEL` | e.g. `deepseek-ai/DeepSeek-V3.2` — use ids from their model list |
| `OPENAI_*` | Fallback if `SILICONFLOW_*` unset (same OpenAI-compatible shape) |

**SiliconFlow compatibility:** OpenAI-style `POST /v1/chat/completions`, `Authorization: Bearer <key>`. Academi does **not** stream yet (`stream=true` unsupported).

**After changing `.env`:** restart `./start.sh` (config loads once at process start).

---

## Architecture

```
web/index.html
  └── js/main.js          # bootstrap, viewport
        └── js/app.js     # AcademiApp shell, navigation, theme
              ├── js/chat.js       # chat, sessions, AI calls, uploads
              ├── js/docs.js       # docs grid, modals, learn flow
              ├── js/community.js  # feed, posts, votes
              ├── js/guide.js      # subjects, guides, editor
              ├── js/api.js        # base URL, debounce, auth headers
              └── js/markdown.js   # escapeHtml, markdown, learn modal

backend/cmd/main.go       # Gin router, CORS, service registration
backend/internal/
  ai/          # chat, learn, prompts, message normalization
  auth/        # JWT, mock session for web demo
  docs/        # upload, extract, CRUD
  community/   # posts, comments, votes
  guide/       # subjects + guides
  chatsessions/# persisted threads (BadgerDB)
  research/    # Wikipedia/web/arXiv snippets when research enabled
  database/    # BadgerDB
```

**Web ↔ API:** Mock auth on load (`POST /auth/mock`). Chat uses `POST /ai/chat` with `messages`, `doc_ids`, `document_mode`, `disable_research`, `help_you_learn`.

**Data:** BadgerDB at `backend/data/badger/` (gitignored). Uploads at `backend/data/uploads/`.

---

## Recent work (this branch / uncommitted)

Much of the following is **local only** — see [Git state](#git-state).

### Web app
- Split monolithic `web/script.js` → `web/js/*` (ES modules).
- **Copy** on AI messages wired (`copy-ai-btn`).
- Incremental chat rendering, 30s screen cache, parallel uploads, debounced doc search.
- Chat errors surface upstream message text (not only generic sorry).

### Backend / AI
- **`siliconflow` provider** in `backend/internal/config/config.go` with `firstEnv` / `envDefault` (fixed bug where literal `"OPENAI_API_KEY"` was sent as the Bearer token).
- **`normalizeChatMessages`** (`backend/internal/ai/messages_normalize.go`): SiliconFlow allows **one** system message at index 0; doc + research prompts are merged before API call.
- Research runs when **doc attached** + Research checkbox, not only Document agent mode.
- Clearer AI upstream errors on non-2xx responses.

### Tooling
- [`start.sh`](start.sh): build backend binary, serve `web/`, clean shutdown on Ctrl+C.
- `.gitignore`: `backend/bin/`, badger data.

---

## Known quirks & fixes already applied

| Symptom | Cause | Fix location |
|---------|--------|--------------|
| `AI provider 'siliconflow' not found` | Provider not registered | `config.go` |
| `401 Invalid token` after renewing key | `envOr` returned string `"OPENAI_API_KEY"` instead of env value | `firstEnv()` in `config.go` |
| `400 System message must be at the beginning` | Multiple `system` roles (prompt + doc + research) | `messages_normalize.go` |
| Chat works on LAN phone but not desktop | `localhost` on phone = phone itself | Set `localStorage.academi_api_base` or open via machine IP |
| ES modules fail from `file://` | Browsers block module imports | Must use HTTP server (`start.sh`) |

---

## Not done / obvious next tasks

1. **Share button** on AI messages (web) — UI only, no handler (`web/js/chat.js`).
2. **Profile stats** hardcoded in `web/index.html` — not from API.
3. **React Native app** — screens exist but lag behind web; `services/academiApi.js` is the RN API helper.
4. **Streaming chat** — backend uses non-streaming completions only.
5. **Service worker / PWA** — stub commented in `web/js/main.js`.
6. **README** — still describes RN-first setup; web module layout partially documented.
7. **Commit & push** — large uncommitted diff; branch **behind `origin/main` by 1** (pull/rebase before push).

---

## Git state (snapshot)

```
Branch: main (behind origin/main)
Modified:  .gitignore, README.md, backend/.env.example, backend/internal/ai/*,
           backend/internal/config/config.go, web/index.html
Deleted:   web/script.js
Untracked: start.sh, web/js/, backend/internal/ai/messages_normalize*.go,
           backend/data/ (runtime — do not commit uploads/DB)
```

**Do not commit:** `backend/.env`, `backend/data/`, API keys.

Suggested commit split (optional):
1. `start.sh` + `.gitignore`
2. Web module split + web improvements
3. Backend AI/config fixes + message normalization tests

---

## Key files to read first (Cursor CLI)

```text
handoff.md              ← this file
start.sh                ← how to run
design.md               ← product spec
web/js/app.js           ← app shell + mixin wiring
web/js/chat.js          ← chat + AI request shape
backend/cmd/main.go     ← routes
backend/internal/ai/service.go      ← buildChatPayload, ChatHandler
backend/internal/ai/messages_normalize.go
backend/internal/config/config.go   ← providers
backend/.env.example
```

---

## Useful API endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/health` | No |
| POST | `/api/v1/auth/mock` | No — web demo session |
| POST | `/api/v1/ai/chat` | Optional (public in current setup) |
| POST | `/api/v1/ai/learn` | Optional |
| GET/POST | `/api/v1/docs` | Bearer |
| POST | `/api/v1/docs/upload` | Bearer |
| GET/POST | `/api/v1/community/posts` | Bearer |
| GET/POST/PATCH | `/api/v1/chat-sessions` | Bearer |
| GET/POST | `/api/v1/guides/*` | Bearer |

Full list: run backend and watch Gin debug routes, or see [`backend/README.md`](backend/README.md).

---

## Cursor CLI tips

1. **Working directory:** repo root `/Users/floatinbee/robo/academi` (or your clone path).

2. **Verify before/after changes:**
   ```bash
   ./start.sh
   curl -s http://localhost:8978/health
   # chat smoke test (needs valid .env key):
   TOKEN=$(curl -s -X POST http://localhost:8978/api/v1/auth/mock | jq -r .token)
   curl -s -X POST http://localhost:8978/api/v1/ai/chat \
     -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
     -d '{"messages":[{"role":"user","content":"hi"}],"context":{}}'
   ```

3. **Backend tests:**
   ```bash
   cd backend && go test ./internal/ai/ -count=1
   ```

4. **Scope agents to the active stack** unless task says otherwise:
   - UI/UX bugs → `web/`
   - API/AI bugs → `backend/internal/`
   - RN → `screens/`, `components/`, `services/`

5. **SiliconFlow / model errors:** check model id in `.env`; error body often appears in chat bubble after recent frontend change.

6. **Optional:** add `.cursor/rules` or `AGENTS.md` pointing to this handoff and “prefer minimal diffs; match `web/js` module pattern.”

---

## Contact / continuity

- Product intent: [`design.md`](design.md)
- Backend ops detail: [`backend/README.md`](backend/README.md)
- Web entry: [`web/index.html`](web/index.html) → [`web/js/main.js`](web/js/main.js)

When in doubt: run `./start.sh`, reproduce on http://localhost:8765, read Network tab for `/api/v1/ai/chat` request/response, and backend stdout for `AI Response Status` lines.
