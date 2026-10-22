# Dorph AI + AI tools Assistants Implementation Plan

> **For agentic workers:** Implement task-by-task. Steps use checkbox syntax.

**Goal:** Card-tile Assistants UI in BK; Dorph AI picker includes BK assistants with prompt + RAG in Dorph chat.

**Architecture:** Dorph merges agent lists server-side; BK RAG queried via `EXTERNAL_API_BASE`. BK frontend switches table → card grid.

**Tech Stack:** React/MUI (BK), React/CSS (Dorph), Go/Gin (Dorph), FastAPI (BK).

## Global Constraints

- BK assistant ids in Dorph: `bk:<id>` prefix.
- Do not break Dorph `image-generator` special path.
- Fail soft when BK unreachable.

---

### Task 1: BK Assistants card grid

**Files:** `bk/frontend/src/pages/AssistantManager.js`

- [x] Replace `Table` with responsive card grid; keep dialogs/actions.

### Task 2: Dorph BK client + list merge

**Files:** `dorph/handlers/bk_assistants.go` (new), `dorph/handlers/dorph_ai_agents.go`, models if needed

- [x] HTTP helpers: list assistants, get assistant, query RAG.
- [x] Merge into `GET /api/ai-agents`.

### Task 3: Dorph chat enrichment for `bk:` agents

**Files:** `dorph/handlers/chat.go` (+ helper)

- [x] Resolve `bk:` → prompt + RAG context → Dorph tool loop.

### Task 4: Dorph sidebar UX polish

**Files:** `dorph/frontend/src/SkoolAiChat.js`, `App.css` if needed

- [x] Show source badge for AI tools assistants in sidebar list.
