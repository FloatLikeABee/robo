# 02 — AI Integration

## Overview

Every app in the platform has an AI assistant. They all follow the same contract and use the same shared libraries.

```
┌─────────────────────────────────────────────────────────────────┐
│                      AI ARCHITECTURE                             │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Frontend: @robo/platform-chat                            │   │
│  │  ┌─────────────────┐  ┌─────────────────┐                 │   │
│  │  │ React wrapper   │  │ Svelte wrapper  │                 │   │
│  │  └────────┬────────┘  └────────┬────────┘                 │   │
│  │           └──────────┬─────────┘                           │   │
│  │                      ▼                                     │   │
│  │           usePlatformChat.ts (framework-agnostic)          │   │
│  │           • Message list state                             │   │
│  │           • send() → POST /assistant/chat                  │   │
│  │           • UI blocks rendering                            │   │
│  │           • Markdown rendering (marked)                    │   │
│  │           • Progress ticker (aiProgress.ts)                │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Backend: /assistant/chat endpoint (per app)              │   │
│  │  ┌────────────────────────────────────────────────────┐   │   │
│  │  │  Request:  { messages, state: { intent, fields } } │   │   │
│  │  │  Response: { assistant_message, state, ui_blocks } │   │   │
│  │  └────────────────────────────────────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  AI Client: pkg/morphai (Go) / pkg/morphai-rs (Rust)     │   │
│  │  ┌────────────────────────────────────────────────────┐   │   │
│  │  │  DashScope (Alibaba Cloud) — qwen3-max default     │   │   │
│  │  │  Dual protocol:                                    │   │   │
│  │  │  • Native DashScope API (text-generation)          │   │   │
│  │  │  • OpenAI-compatible (/chat/completions)           │   │   │
│  │  │  Rate limiting, retry (3x), 120s/300s timeouts     │   │   │
│  │  └────────────────────────────────────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## The MorphAI contract

Defined in `AI_ASSISTANT_MORPHAI_CONTRACT.md`. All `/assistant/chat` endpoints follow this:

### Request shape

```json
{
  "messages": [
    { "role": "user", "content": "create form name: Survey A" }
  ],
  "state": {
    "intent": "create_form",
    "fields": { "name": "Survey A" }
  }
}
```

### Response shape

```json
{
  "assistant_message": "I can create the form. Please provide: slug",
  "intent": "create_form",
  "missing_fields": ["slug"],
  "state": {
    "intent": "create_form",
    "fields": { "name": "Survey A" }
  },
  "completed": false,
  "ui_blocks": []
}
```

### Behavioral rules

1. If user intent is write/create/update, detect required fields.
2. If required fields are missing, ask follow-up questions.
3. Only execute write when required fields are complete.
4. Keep `state.intent` + `state.fields` for multi-turn conversation.
5. Support general Q&A outside platform-specific actions.

## Shared AI client: pkg/morphai (Go)

**Module:** `github.com/robo/morphai` (linked via `replace` directive in go.mod)

### Key API

```go
// Load config from environment
cfg := morphai.LoadFromEnv()
// Reads: MORPH_AI_API_KEY, MORPH_AI_MODEL, MORPH_AI_API_URL, MORPH_AI_BASE_URL
// Fallbacks: GEMINI_API_KEY, TRAN_QWEN_API_KEY, etc.
// Default model: qwen3-max

// Create client
client := morphai.NewClient(cfg)
// Or shortcut:
client := morphai.NewClientFromEnv()

// Check if configured
if client.Configured() { ... }

// Send chat completion (120s timeout, 3 retries)
response, err := client.ChatCompletion(ctx, messages)

// Long generation (300s timeout)
response, err := client.ChatCompletionLong(ctx, messages)
```

### Utility functions (pkg/morphai/context.go)

```go
morphai.TruncateRunes(s, maxRunes)            // Truncate to N runes
morphai.ExtractJSONObject(s)                  // Extract { ... } from model output
morphai.ToolFollowUpPrompt(toolResult)        // Build tool-loop message
morphai.DefaultHistoryMaxMessages = 8         // History truncation constants
morphai.DefaultToolMaxRounds = 8
```

### Dual protocol

The client auto-detects which protocol to use:
- If `MORPH_AI_API_URL` contains `text-generation/generation` → **Native DashScope** API
- Otherwise → **OpenAI-compatible** `/chat/completions` (with `enable_thinking: false` to suppress Qwen3.x chain-of-thought)

## Shared AI client: pkg/morphai-rs (Rust)

**Crate:** `morphai` (linked via `path = "../pkg/morphai-rs"` in Cargo.toml)

### Key API

```rust
use morphai::{Client, Message, Config};

// Load config from environment
let cfg = Config::from_env();
// Or with MESSAGE_AI_* prefix fallback:
let cfg = Config::from_message_ai_env();

// Create client
let client = Client::new(cfg);
// Or shortcut:
let client = Client::from_env();

// Send chat completion
let messages = vec![Message::user("Hello")];
let response = client.chat_completion(&messages).await?;
```

### Utility functions

```rust
morphai::truncate_chars(s, max_chars)
morphai::extract_json_object(s)
morphai::tool_follow_up_prompt(tool_result)
```

## Markdown formatting: pkg/assistmd (Go)

Used by Go backends to format assistant responses consistently.

```go
import "github.com/robo/assistmd"

assistmd.Title("Form Created")           // → **Form Created**
assistmd.Success("Form", "Survey A")     // → ✅ **Form** — Survey A
assistmd.BulletList("Fields:", items)    // → Fields:\n- item1\n- item2
assistmd.KVBlock("Details", pairs)       // → - **key:** value
assistmd.NamedSlug("My Form", "my-form") // → **My Form** — `my-form`
```

## Per-app assistant implementation

| App | Endpoint | Key files | Pattern |
|-----|----------|-----------|---------|
| morph | `POST /api/chat` | `handlers/chat.go` | Direct AI chat + hybrid context RAG |
| formx | `POST /api/v1/assistant/chat` | `internal/handler/assistant.go`, `assistant_llm.go` | Rule-based + LLM tool-calling loop (8 rounds) |
| composerx | `POST /ai/assistant/chat` | `platform_assistant.go`, `platform_assistant_llm.go` | Regex intent detection + LLM tool loop |
| booki | `POST /api/v1/assistant/chat` | `internal/handlers/assistant.go` | Intent-based: create_customer, create_booking, etc. |
| UsersPanel | `POST /api/assistant/chat` | `handlers/assistant.rs` | Stateful clarification for users/roles |
| SharpReport | `POST /api/v1/assistant/chat` | `api/report_assistant.rs` | Stateful clarification for report creation |
| morph-engi | `POST /api/v1/assistant/chat` | `api/assistant.rs` | Full CRUD tool suite |

## Adding AI to a new feature

### Go pattern (formx, composerx, booki)

```go
// 1. Create morphai client (usually in handler struct)
aiClient := morphai.NewClient(morphai.LoadFromEnv())

// 2. Build messages
messages := []morphai.Message{
    {Role: "system", Content: "You are a helpful assistant..."},
    {Role: "user", Content: userMessage},
}

// 3. Call AI
response, err := aiClient.ChatCompletion(ctx, messages)

// 4. Format response with assistmd
formatted := assistmd.Success("Created", detail)
```

### Rust pattern (UsersPanel, SharpReport, morph-engi)

```rust
// 1. Create client
let cfg = Config::from_env();
let client = Client::new(cfg);

// 2. Build messages
let messages = vec![
    Message::system("You are a helpful assistant..."),
    Message::user(&user_message),
];

// 3. Call AI
let response = client.chat_completion(&messages).await?;
```

## Web research: pkg/webresearch (Go)

Used by formx and composerx to ground AI responses with real-world context.

```go
import "github.com/robo/webresearch"

notes, sources := webresearch.Gather("quantum computing")
// notes: combined text from Wikipedia, DuckDuckGo, arXiv
// sources: []Source with Title, Type ("wiki"/"web"/"paper"), URL
```

No API keys required — uses free public APIs (Wikipedia, DuckDuckGo Instant Answer, arXiv).

## GraphRAG: pkg/morphgraph (Go)

Used by morph for Neo4j-backed knowledge graph search. See `docs/MORPH_GRAPH_RAG_PLAN.md` and `docs/MORPH_GRAPH_OPS.md`.

```go
import "github.com/robo/morphgraph"

cfg := morphgraph.LoadFromEnv()
store, _ := morphgraph.OpenStore(cfg)
store.VectorSearch(ctx, embedding, 10)
```

## Platform chat drawer: @robo/platform-chat

See `11-platform-chat.md` for full integration guide. Quick summary:

```tsx
// React
import PlatformChatDrawer from '@robo/platform-chat/react';
import '@robo/platform-chat/chat-drawer.css';

<PlatformChatDrawer
  open={open}
  onClose={() => setOpen(false)}
  chatEndpoint="/api/v1/assistant/chat"
  getHeaders={() => ({ Authorization: `Bearer ${token}` })}
  title="AI Assistant"
/>
```

```svelte
<!-- Svelte -->
import PlatformChatDrawer from '@robo/platform-chat/svelte';
import '@robo/platform-chat/chat-drawer.css';

<PlatformChatDrawer
  bind:open
  {chatEndpoint}
  {getHeaders}
  title="AI Assistant"
/>
```