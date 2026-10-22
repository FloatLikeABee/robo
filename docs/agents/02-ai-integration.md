# 02 — AI Integration

## Overview

**Dorph AI** (`dorph/frontend`, port 3031) is the system chat. DorphUtils modules (Event Logs, Content Maker, Data Access, Project) do **not** ship a floating `@robo/platform-chat` drawer. Compose/progress UIs may copy a local `aiProgress` helper for status text.

Backends still use the shared Go client for in-app LLM features (compose, tools, generate-document). The default env path is DashScope. A named provider is optional.

```
  Dorph AI UI (:3031)
        │
        ▼
  Dorph API  /api/chat  + Skills + optional AI tools proxy (bk)
        │
        ▼
  pkg/dorphai  (Go)     pkg/dorphai-rs  (Rust)
        │
        ▼
  Selected provider (default: DashScope / Qwen via DORPH_AI_*, model qwen3-max)
```

## Shared client: `pkg/dorphai` (Go)

**Module:** `github.com/robo/dorphai` (`replace` in each app `go.mod`).

```go
cfg := dorphai.LoadFromEnv()
client := dorphai.NewClientFromEnv()
if client.Configured() {
    response, err := client.ChatCompletion(ctx, messages)
}
```

Reads `DORPH_AI_API_KEY`, `DORPH_AI_MODEL`, `DORPH_AI_API_URL`, `DORPH_AI_BASE_URL`. Fallbacks: `GEMINI_API_KEY`, `TRAN_QWEN_*`. Default model: `qwen3-max`. ChatCompletion: 120s timeout, 3 retries. Long/vision helpers exist for longer jobs.

Named providers (`openai`, `anthropic`, `xai`, `gemini`, `openrouter`, `ollama`, `dashscope`, `mistral`, `groq`, `openai-compatible`) are selected on `Config.Provider` or per call. Each uses its own env key. A missing key returns `ErrProviderNotConfigured` and is not sent to another vendor. Gemini uses Google's OpenAI-compatible endpoint. Anthropic uses the Messages API. Mistral and Groq are named OpenAI-compatible presets. The library does not store keys; a later resolver can pass an admin/workspace key, then a per-user key, on `Config.APIKey`.

**Used by:** dorph, formx, composerx.

See [`pkg/dorphai/README.md`](../../pkg/dorphai/README.md).

## Shared client: `pkg/dorphai-rs` (Rust)

Path dependency from Data Access and Project. Same env keys. Used for Project document generation and Data Access table/report AI — **not** as a second chat product.

## Dorph AI product

- Chat sessions and messages in Dorph Badger.
- Header **Skills** (markdown upload + catalog). The catalog also lists that user's enabled session lessons.
- Session lessons (`agent_lesson`) are per user. Only enabled lessons for the bearer-authenticated user are injected into `/api/chat` prompts. A client `X-User-ID` does not select lessons. Operators list (including disabled), disable, and delete their own lessons at `/api/agent-lessons`; another user's id is 404. Legacy unowned rows are claimed at most once, and only when the first ownership decision sees a single account. A session the account already owns keeps that lesson and drops the conflicting legacy row. The management-chat exact cache follows the bearer user's enabled lessons (see [`03-dorph.md`](03-dorph.md)).
- **AI tools** workspace can open `bk` (Assistants, RAG, Documents, System).
- Notes and knowledge belong in DorphNotes and in the agent workspace Notes & TODOs and Context & Knowledge tabs, not a duplicate header shortcut.

## Leftover `/assistant/chat` APIs

Some module backends still expose `POST .../assistant/chat` for tool loops. Treat those as backend helpers. Do not reintroduce satellite assistant drawers in Event Logs, Content Maker, Data Access, or Project.

Typical request/response shape (messages + optional `state` / `ui_blocks`) is implemented per app. There is no separate `AI_ASSISTANT_DORPHAI_CONTRACT.md` in the repo.

## Content Maker embeddings

Optional `TRAN_OPENAI_API_KEY` for reference-library embeddings. Independent of the DashScope chat key.

## Optional GraphRAG

Neo4j + `dorphgraph-worker` are optional. See [`docs/DORPH_GRAPH_OPS.md`](../DORPH_GRAPH_OPS.md).
