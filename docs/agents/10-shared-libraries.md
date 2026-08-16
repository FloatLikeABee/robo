# 10 — Shared Libraries (`pkg/`)

## Overview

The `pkg/` directory contains shared Go and Rust libraries used by multiple apps. They're linked via `replace` directives in `go.mod` (Go) or `path` dependencies in `Cargo.toml` (Rust).

```
pkg/
├── morphai/          ← Go AI client (DashScope)
├── morphai-rs/       ← Rust AI client (DashScope)
├── assistmd/         ← Go markdown formatting helpers
├── docextract/       ← Go PDF/text extraction for AI prompts
├── morphgraph/       ← Go GraphRAG types + Neo4j operations
└── webresearch/      ← Go web research (Wikipedia, DuckDuckGo, arXiv)
```

---

## pkg/morphai/ — Go DashScope AI Client

**Module:** `github.com/robo/morphai` (Go 1.21, zero external dependencies)

**Used by:** morph, formx, composerx, booki

### Key types

```go
type Config struct {
    APIKey       string
    Model        string
    APIURL       string  // Native DashScope endpoint
    BaseURL      string  // OpenAI-compatible endpoint
    UseNativeAPI bool
}

type Client struct { ... }

type Message struct {
    Role    string
    Content string
}

// Multimodal turn for vision requests; Message is unchanged.
type MultiMessage struct {
    Role    string
    Content []ContentPart
}

type ContentPart struct {
    Type     string     // "text" or "image_url"
    Text     string
    ImageURL *ImageURL
}
```

### Key functions

```go
// Load config from environment
cfg := morphai.LoadFromEnv()
// Reads: MORPH_AI_API_KEY → MORPH_AI_MODEL → MORPH_AI_API_URL → MORPH_AI_BASE_URL
// Fallbacks: GEMINI_API_KEY, TRAN_QWEN_API_KEY, TRAN_QWEN_MODEL, etc.
// Default model: "qwen3-max"

// Create client
client := morphai.NewClient(cfg)
// Or shortcut:
client := morphai.NewClientFromEnv()

// Check if configured (API key is set)
if client.Configured() { ... }

// Send chat completion (120s timeout, 3 retries, 200ms rate limit)
response, err := client.ChatCompletion(ctx, messages)

// Long generation (300s timeout)
response, err := client.ChatCompletionLong(ctx, messages)

// Vision: text plus images (300s timeout). Pass "" to use the configured vision model.
reply, err := client.ChatCompletionVision(ctx, []morphai.MultiMessage{
    morphai.UserMultiMessage(
        morphai.TextPart("Transcribe this form"),
        morphai.ImageDataPart("image/png", raw),
    ),
}, "")

// Can this client send images at all?
if client.VisionSupported() { ... }
```

### Vision requests

`ChatCompletionVision` POSTs an OpenAI-style `content` array to
`{BaseURL}/chat/completions`, reusing the same retry, rate-limit, and error
handling as text chat. Because only the compatible endpoint accepts a content
array, a client configured for the native DashScope text-generation endpoint
returns an error naming `MORPH_AI_API_URL` rather than sending an unusable payload.

The vision model comes from the optional `MORPH_AI_VISION_MODEL` (default
`qwen-vl-max`). The chat model is never used as a fallback — a text-only model
would reject or silently ignore image content. No new required environment
variable: an API key alone is enough.

### Utility functions (`context.go`)

```go
morphai.TruncateRunes(s string, max int) string
morphai.ExtractJSONObject(s string) (string, bool)
morphai.ToolFollowUpPrompt(toolResult string) string
morphai.ToolFollowUpPromptWithInstruction(toolResult, instruction string) string

// Constants
morphai.DefaultHistoryMaxMessages = 8
morphai.DefaultHistoryMaxRunes = 800
morphai.DefaultToolResultMaxRunes = 12_000
morphai.DefaultToolMaxRounds = 8
```

### Dual protocol

The client auto-detects which API protocol to use:

- **Native DashScope:** If `MORPH_AI_API_URL` contains `"text-generation/generation"` → POSTs to DashScope text-generation API
- **OpenAI-compatible:** Otherwise → POSTs to `{BaseURL}/chat/completions` with `enable_thinking: false`

Both paths share retry logic: 3 attempts, exponential backoff (2s, 4s, 8s), rate-limit handling on HTTP 429.

### How apps link it

```go
// In go.mod:
replace github.com/robo/morphai => ../../pkg/morphai

// In code:
import "github.com/robo/morphai"
```

---

## pkg/morphai-rs/ — Rust DashScope AI Client

**Crate:** `morphai` v0.1.0 (edition 2021)

**Used by:** UsersPanel, SharpReport, morph-engi

### Key types

```rust
pub struct Config {
    pub api_key: String,
    pub model: String,
    pub api_url: String,
    pub base_url: Option<String>,
}

pub struct Client { ... }

pub struct Message {
    pub role: String,
    pub content: String,
}
```

### Key functions

```rust
// Load config from environment
let cfg = Config::from_env();
// Or with MESSAGE_AI_* prefix fallback:
let cfg = Config::from_message_ai_env();

// Create client
let client = Client::new(cfg);
// Or shortcut:
let client = Client::from_env();

// Check if configured
if cfg.configured() { ... }

// Send chat completion (180s timeout, 3 retries)
let messages = vec![Message::user("Hello")];
let response = client.chat_completion(&messages).await?;
```

### Utility functions

```rust
morphai::truncate_chars(s: &str, max_chars: usize) -> String
morphai::extract_json_object(s: &str) -> Option<String>
morphai::tool_follow_up_prompt(tool_result: &str) -> String
morphai::tool_follow_up_prompt_with_instruction(tool_result: &str, instruction: &str) -> String

// Constants
morphai::DEFAULT_HISTORY_MAX_CHARS       // 500
morphai::DEFAULT_HISTORY_MAX_MESSAGES    // 4
morphai::DEFAULT_TOOL_MAX_ROUNDS         // 8
morphai::DEFAULT_TOOL_RESULT_MAX_CHARS   // 8_000
```

### How apps link it

```toml
# In Cargo.toml:
[dependencies]
morphai = { path = "../pkg/morphai-rs" }
```

---

## pkg/assistmd/ — Go Markdown Formatting

**Module:** `github.com/robo/assistmd` (Go 1.21, zero dependencies)

**Used by:** formx, composerx, booki

### All functions

```go
import "github.com/robo/assistmd"

assistmd.Title("Form Created")              // → **Form Created**
assistmd.Empty("No results")                // → _No results_
assistmd.Success("Form", "Survey A")        // → ✅ **Form** — Survey A
assistmd.BulletList("Fields:", items)       // → Fields:\n- item1\n- item2
assistmd.KVBlock("Details", pairs)          // → - **key:** value
assistmd.NamedSlug("My Form", "my-form")    // → **My Form** — `my-form`
assistmd.NamedID("User", 42)                // → **User** — #42
```

Pure formatting — no side effects, no network calls. Used to render consistent assistant reply formatting across all Go backends.

---

## pkg/docextract/ — Go Document Text Extraction

**Module:** `github.com/robo/docextract` (Go 1.21)

**Dependencies:** `ledongthuc/pdf` only

**Used by:** formx

Turns uploaded files into AI-readable text: classifies a file, extracts text from
PDFs and text files, and caps content so callers stay inside their prompt budget.
`pkg/morphgraph` also parses PDFs, but it carries the Neo4j driver, which is why
this lightweight module exists separately.

### All functions

```go
import "github.com/robo/docextract"

// Is this an extractable document, a vision-readable image, or neither?
// The declared MIME type wins when meaningful; otherwise the extension decides.
kind := docextract.Classify(name, mime) // KindDocument | KindImage | KindUnsupported
docextract.IsPDF(name, mime)

// Text extraction. ExtractPDF recovers from parser panics on malformed files and
// returns ErrNoText for a PDF with no text layer (typically a scan).
text, err := docextract.ExtractPDF(path)
text, err := docextract.ExtractPDFBytes(raw)
text := docextract.ExtractText(raw) // TXT/MD/CSV; replaces invalid UTF-8, keeps CSV rows

// Size control — the marker makes truncation visible to the model and the user.
capped, wasCut := docextract.Truncate(s, docextract.MaxPerFileChars)

// For error messages: "accepted file types: PDF, TXT, MD, CSV"
msg := docextract.AcceptedTypesMessage(docextract.KindDocument)
```

Accepted types: PDF, TXT, MD, CSV as documents; JPEG, PNG, GIF, WebP as images.
Caps: `MaxPerFileChars` (12000), `MaxRequestChars` (24000).

Morph Engi has an equivalent in Rust at `morph-engi/backend/src/services/doc_text.rs`,
built on the `pdf-extract` crate with `spawn_blocking` and `catch_unwind` so a
malformed PDF cannot block the async runtime or abort the request.

---

## pkg/morphgraph/ — Go GraphRAG Types + Neo4j

**Module:** `github.com/robo/morphgraph` (Go 1.24)

**Dependencies:** `neo4j/neo4j-go-driver/v5`, `ledongthuc/pdf`

**Used by:** morph (for knowledge graph features)

### Key types

```go
type Config struct {
    Enabled          bool
    Neo4jURI         string
    Neo4jUser        string
    Neo4jPassword    string
    Neo4jDatabase    string
    EmbeddingModel   string
    OpenAIAPIKey     string
    OpenAIBaseURL    string
}

type Store struct { ... }     // Neo4j-backed graph client
type Embedder struct { ... }  // OpenAI-compatible embedding API
type NodeProps struct {
    UID, Source, Type, SourceID, Title, Summary string
    Labels []string
    UpdatedAt time.Time
}
```

### Key functions

```go
// Config
cfg := morphgraph.LoadFromEnv()
// Reads: MORPH_GRAPH_ENABLED, NEO4J_URI, NEO4J_USER, NEO4J_PASSWORD, etc.

// Store (nil-safe when disabled)
store, err := morphgraph.OpenStore(cfg)
store.BootstrapSchema(ctx)                           // Create constraints + vector index
store.UpsertEntity(ctx, NodeProps{...})              // MERGE entity
store.UpsertChunk(ctx, chunkUID, entityUID, text, embedding)
store.VectorSearch(ctx, embedding, limit)            // Cosine similarity search
store.Close(ctx)

// Embedder
emb := morphgraph.NewEmbedder(cfg)
vectors, err := emb.Embed(ctx, []string{"text"})     // → [][]float32

// Chunking
chunks := morphgraph.ChunkText(text, 900, 120)       // 900-rune windows, 120 overlap

// Utilities
uid := morphgraph.UID("morph", "entity", "123")      // → "morph:entity:123"
```

---

## pkg/webresearch/ — Go Web Research

**Module:** `github.com/robo/webresearch` (Go 1.21, zero dependencies)

**Used by:** formx, composerx

### Key types

```go
type Source struct {
    Title string
    Type  string   // "wiki", "web", "paper"
    URL   string
}
```

### Key function

```go
notes, sources := webresearch.Gather("quantum computing")
// notes: combined text from Wikipedia + DuckDuckGo + arXiv
// sources: []Source with citations
```

Calls three free public APIs (no API keys required):
1. **Wikipedia** — search + REST summary (title, excerpt, URL)
2. **DuckDuckGo** — Instant Answer API (AbstractText)
3. **arXiv** — API query (top 2 results, title + ID URL)

User-Agent: `"RoboMorphAI/1.0 (web-grounded generation; local development)"`