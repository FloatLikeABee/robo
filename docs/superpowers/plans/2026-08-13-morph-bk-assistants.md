# Morph AI + AI tools Assistants Implementation Plan

> **For agentic workers:** Implement task-by-task. Steps use checkbox syntax.

**Goal:** Card-tile Assistants UI in BK; Morph AI picker includes BK assistants with prompt + RAG in Morph chat.

**Architecture:** Morph merges agent lists server-side; BK RAG queried via `EXTERNAL_API_BASE`. BK frontend switches table → card grid.

**Tech Stack:** React/MUI (BK), React/CSS (Morph), Go/Gin (Morph), FastAPI (BK).

## Global Constraints

- BK assistant ids in Morph: `bk:<id>` prefix.
- Do not break Morph `image-generator` special path.
- Fail soft when BK unreachable.

---

### Task 1: BK Assistants card grid

**Files:** `bk/frontend/src/pages/AssistantManager.js`

- [x] Replace `Table` with responsive card grid; keep dialogs/actions.

### Task 2: Morph BK client + list merge

**Files:** `morph/handlers/bk_assistants.go` (new), `morph/handlers/morph_ai_agents.go`, models if needed

- [x] HTTP helpers: list assistants, get assistant, query RAG.
- [x] Merge into `GET /api/ai-agents`.

### Task 3: Morph chat enrichment for `bk:` agents

**Files:** `morph/handlers/chat.go` (+ helper)

- [x] Resolve `bk:` → prompt + RAG context → Morph tool loop.

### Task 4: Morph sidebar UX polish

**Files:** `morph/frontend/src/SkoolAiChat.js`, `App.css` if needed

- [x] Show source badge for AI tools assistants in sidebar list.
