# 02 — AI Integration

## Overview

**Morph AI** (`morph/frontend`, port 3031) is the system chat. MorphUtils modules (Event Logs, Content Maker, Data Access, Project) do **not** ship a floating `@robo/platform-chat` drawer. Compose/progress UIs may copy a local `aiProgress` helper for status text.

Backends still use the shared Go client for in-app LLM features (compose, tools, generate-document). The default env path is DashScope. A named provider is optional.

```
  Morph AI UI (:3031)
        │
        ▼
  Morph API  /api/chat  + Skills + optional AI tools proxy (bk)
        │
        ▼
  pkg/morphai  (Go)     pkg/morphai-rs  (Rust)
        │
        ▼
  Selected provider (default: DashScope / Qwen via MORPH_AI_*, model qwen3-max)
```

## Shared client: `pkg/morphai` (Go)

**Module:** `github.com/robo/morphai` (`replace` in each app `go.mod`).

```go
cfg := morphai.LoadFromEnv()
client := morphai.NewClientFromEnv()
if client.Configured() {
    response, err := client.ChatCompletion(ctx, messages)
}
```

Reads `MORPH_AI_API_KEY`, `MORPH_AI_MODEL`, `MORPH_AI_API_URL`, `MORPH_AI_BASE_URL`. Fallbacks: `GEMINI_API_KEY`, `TRAN_QWEN_*`. Default model: `qwen3-max`. ChatCompletion: 120s timeout, 3 retries. Long/vision helpers exist for longer jobs.

Named providers (`openai`, `anthropic`, `xai`, `gemini`, `openrouter`, `ollama`, `dashscope`, `mistral`, `groq`, `openai-compatible`) are selected on `Config.Provider` or per call. Each uses its own env key. A missing key returns `ErrProviderNotConfigured` and is not sent to another vendor. Gemini uses Google's OpenAI-compatible endpoint. Anthropic uses the Messages API. Mistral and Groq are named OpenAI-compatible presets. The library does not store keys; a later resolver can pass an admin/workspace key, then a per-user key, on `Config.APIKey`.

**Used by:** morph, formx, composerx.

See [`pkg/morphai/README.md`](../../pkg/morphai/README.md).

## Shared client: `pkg/morphai-rs` (Rust)

Path dependency from Data Access and Project. Same env keys. Used for Project document generation and Data Access table/report AI — **not** as a second chat product.

## Morph AI product

- Chat sessions and messages in Morph Badger.
- Header **Skills** (markdown upload + catalog).
- **AI tools** workspace can open `bk` (Assistants, RAG, Documents, System).
- Notes and knowledge belong in MorphNotes and in the agent workspace Notes & TODOs and Context & Knowledge tabs, not a duplicate header shortcut.

## Leftover `/assistant/chat` APIs

Some module backends still expose `POST .../assistant/chat` for tool loops. Treat those as backend helpers. Do not reintroduce satellite assistant drawers in Event Logs, Content Maker, Data Access, or Project.

Typical request/response shape (messages + optional `state` / `ui_blocks`) is implemented per app. There is no separate `AI_ASSISTANT_MORPHAI_CONTRACT.md` in the repo.

## Content Maker embeddings

Optional `TRAN_OPENAI_API_KEY` for reference-library embeddings. Independent of the DashScope chat key.

## Optional GraphRAG

Neo4j + `morphgraph-worker` are optional. See [`docs/MORPH_GRAPH_OPS.md`](../MORPH_GRAPH_OPS.md).
