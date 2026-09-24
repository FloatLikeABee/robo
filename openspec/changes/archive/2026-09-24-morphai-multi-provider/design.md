# Design

## Context

`pkg/morphai.Client` is constructed by morph, formx, and composerx as `NewClient(LoadFromEnv())` or `NewClient(Config{APIKey, Model, BaseURL, APIURL, UseNativeAPI})`. Chat is `ChatCompletion` / `ChatCompletionLong`. Vision is `ChatCompletionVision`. Tool loops in those apps parse JSON out of the reply; the library itself has no tool or stream API yet. `LoadFromEnv` fills `BaseURL` with the DashScope compatible-mode URL even when the env var is unset, and sets `UseNativeAPI` only when `MORPH_AI_API_URL` contains `text-generation/generation`. See proposal.md for why this story exists.

This is an architectural change to a shared client. The design below was grilled before the client was wired up: the same notes play proposer and reviewer.

## Goals / Non-Goals

**Goals:**

- One registry and one `Client`. Callers keep the methods they already call.
- Named providers, including Mistral and Groq as presets on the OpenAI chat-completions adapter. OpenRouter stays its own id.
- Resolution: per-call fields, then client `Config`, then that provider's env vars, then defaults. No cross-provider key.
- A place for a later story to drop in an already-resolved admin/workspace key, and later a per-user key, without this package storing anything.

**Non-Goals:**

- Settings UI, database rows, encryption, or a secrets package.
- Rust client (`pkg/morphai-rs`), Morph frontend, CI workflows.
- Inferring a provider from whichever env var happens to be set.

## Decisions

### 1. One client, a registry, two protocol adapters

Keep `*morphai.Client`. `Config.Provider` empty means the legacy path. A non-empty id selects a `providerSpec` (metadata, env var names, default URL, suggested models, which adapter).

Adapters:

| Adapter | Providers |
| --- | --- |
| OpenAI chat completions (`/chat/completions`, Bearer, SSE `data:` chunks, `tools` / `tool_calls`, `response_format.json_object`, `image_url`) | openai, xai, gemini, openrouter, ollama, dashscope (compatible-mode), mistral, groq, openai-compatible |
| Anthropic Messages (`/v1/messages`, `x-api-key`, `anthropic-version`, system hoisted, `tool_use` / `tool_result`) | anthropic |
| DashScope native text-generation | dashscope only when `UseNativeAPI` is set (today's `MORPH_AI_API_URL` path) |

Gemini uses Google's OpenAI-compatible base `https://generativelanguage.googleapis.com/v1beta/openai`, not the native `generateContent` API.

**Alternatives rejected:**

- A separate client type per vendor. Reviewer point: morph, formx, and composerx already store `*morphai.Client`. New types would churn every caller for no protocol benefit.
- One OpenAI adapter for Anthropic too (some proxies speak both). Rejected: the product calls Anthropic's own API. Auth, system, and tool blocks differ; a proxy would hide missing-key errors behind another host.
- A native adapter for every vendor (Gemini `generateContent`, Mistral, Groq, xAI). Rejected: those four plus OpenAI, OpenRouter, Ollama, and DashScope compatible-mode already accept chat completions, streaming, tools, and images. A native Gemini adapter is a second JSON dialect (contents/parts, `systemInstruction`, `functionDeclarations`) to test and keep in sync. If a Gemini capability is missing from the compat endpoint later, the registry can swap the adapter behind id `gemini` without a caller change.
- Treating Mistral and Groq as "paste a base URL" on `openai-compatible` only. Rejected by product decision: the settings list needs named presets. They still share the OpenAI adapter. OpenRouter is not folded into that preset because its model ids are `vendor/model` and it is already a named provider.

**Failure modes challenged:**

- `enable_thinking: false` is sent on every compatible-mode call today, including a custom base URL. Strict providers can reject unknown fields. Legacy calls keep sending it so today's Qwen and SiliconFlow behavior does not change. Explicit non-DashScope providers omit it. Explicit DashScope, and any base URL on `dashscope.aliyuncs.com`, still send it.
- Suggested OpenAI models stay on chat-completions models (`gpt-4.1`, `gpt-4o`, `gpt-4o-mini`). A newer flagship that only tool-calls through the Responses API is not the default.
- Ollama's OpenAI surface is under `/v1`. A base URL without that suffix gets `/v1` appended. A URL that already ends in `/v1` does not.
- Anthropic requires `max_tokens`. The adapter sends 4096, or 8192 for the long client, unless the call sets `MaxTokens`. Empty assistant text plus a tool call is sent as tool blocks only, because Anthropic rejects an empty content array.

### 2. Existing methods stay; new methods are additive

Unchanged signatures: `NewClient`, `NewClientFromEnv`, `LoadFromEnv`, `ChatCompletion`, `ChatCompletionLong`, `ChatCompletionVision`, `Configured`, `VisionSupported`, `VisionModel`, `Message.Content` as a string, `MultiMessage`.

Added: `Complete` / `CompleteLong` (tools and JSON mode), `Stream`, `ListProviders`, `LookupProvider`, `ResolveConfig`. `Message` gains `ToolCalls`, `ToolCallID`, and `Name` with `omitempty`, so a normal chat message still marshals as `role` and `content`.

`ChatCompletion` returns reply text only. Tool calls are visible on `Complete`. That keeps the JSON-in-the-prompt loops in morph, formx, and composerx on the same path.

### 3. Resolution order and the not-configured error

For a named provider:

1. Per-call `CompletionRequest` provider, model, API key, and base URL when non-empty.
2. Client `Config` fields. This is where a later story places a stored key.
3. That provider's env vars only (`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `XAI_API_KEY`, `GEMINI_API_KEY`, `OPENROUTER_API_KEY`, `OLLAMA_BASE_URL`, `DASHSCOPE_API_KEY`, `MISTRAL_API_KEY`, `GROQ_API_KEY`, `OPENAI_COMPATIBLE_API_KEY`, plus each `*_BASE_URL`).
4. Provider default model (first suggested model) and default base URL.

DashScope's extra env fallbacks are `MORPH_AI_API_KEY` and `TRAN_QWEN_API_KEY` (and the matching base URL vars), because those are the same vendor. `GEMINI_API_KEY` is not one of them.

If the provider changes between the client and the call, the call does not inherit the previous key or base URL. Exception: a legacy client (empty provider) selecting `dashscope` keeps its key and URLs, because that client is already DashScope.

`LoadFromEnv` always stores a base URL: `MORPH_AI_BASE_URL`, else `TRAN_QWEN_BASE_URL`, else a compatible-mode `MORPH_AI_API_URL`, else the DashScope compatible-mode default. That value is snapshotted with the loaded key. Selecting a named provider drops the snapshot and uses a caller-set base URL, that provider's own env base, or the provider default. A base URL written on `Config` or on the call is kept even when it equals the DashScope default.

Missing key (or missing base URL for `openai-compatible`) returns `*ProviderConfigError`, which unwraps to `ErrProviderNotConfigured`. The legacy empty-provider path still returns the exact string `MORPH_AI_API_KEY is not configured`, and that error also unwraps to `ErrProviderNotConfigured`. Nothing in this package then retries a different provider.

Ollama is configured with no key when a base URL resolves. The default is `http://127.0.0.1:11434/v1`. Reachability is the chat call's transport error, not a startup probe.

**Legacy env path (provider empty):** unchanged. Key order `MORPH_AI_API_KEY`, `GEMINI_API_KEY`, `TRAN_QWEN_API_KEY`. Model order `MORPH_AI_MODEL`, `GEMINI_MODEL`, `TRAN_QWEN_MODEL`, then `qwen3-max`. `OPENAI_API_KEY` does not configure this path.

**GEMINI_API_KEY challenge:** the same variable is the legacy single-key fallback and the Gemini provider key. That split is intentional. `LoadFromEnv` still copies it into `Config.APIKey` for the unscoped client, and keeps a private snapshot of that copy. A named provider ignores a key that still equals the snapshot and reads its own env var. It does not re-read the legacy env vars to decide whether the field is caller-set, so rotating those vars after load does not turn the snapshot into an explicit key. An explicit `gemini` provider therefore uses `GEMINI_API_KEY`. An explicit `dashscope` provider does not, so a Gemini key copied by `LoadFromEnv` is not sent to DashScope. `MORPH_AI_API_KEY` and `TRAN_QWEN_API_KEY` still authorize explicit DashScope through that provider's env list, not through the copied field.

### 4. How stored keys plug in later

This package does not open a database and does not encrypt. A later resolver, living next to the secrets package, fills `Config` or `CompletionRequest` before `NewClient` / `Complete`.

Precedence that resolver should apply once both stored sources exist:

1. Per-call key already on `CompletionRequest`.
2. Stored per-user key.
3. Stored admin/workspace key.
4. Leave `APIKey` empty and let this package use the provider env var.

Delivery order is the opposite of that precedence: the first stored source to build is admin/workspace; per-user keys are a later story. Until the per-user story exists, the resolver sets `Config.APIKey` from the admin/workspace secret and this package treats it as an explicit key, which already beats env. Encryption stays in the secrets package; this story never sees ciphertext.

### 5. Capability errors

`ErrCapabilityUnsupported` (via `*CapabilityError`) covers streaming, tools, vision, and JSON mode.

- Anthropic: chat, stream, tools, vision. No JSON mode (`response_format` does not exist on Messages).
- DashScope native: chat only. Vision keeps today's `MORPH_AI_API_URL` sentence so formx's image test still matches.
- Every OpenAI-compatible preset except Groq, including Mistral, Ollama, and generic: chat, stream, tools, vision, JSON mode at the protocol layer. Groq is chat, stream, tools, and JSON mode only. The public catalog has no general vision model (`llama-3.2-11b-vision-preview` and `gemma2-9b-it` are retired). Suggested Groq models are `openai/gpt-oss-120b` and `openai/gpt-oss-20b`. A model that rejects a capability returns the provider's HTTP error, mapped to `*APIError` (`API error (status N): code - message`), which matches the string shape callers already display.

## Risks / Trade-offs

- [Gemini compat lags a native-only feature] → id stays `gemini`; a later change can replace the adapter. This story does not half-implement `generateContent`.
- [Qwen thinking flag rejected by a strict compatible endpoint on the legacy path] → legacy behavior is preserved on purpose. New named providers do not send the flag unless the URL is DashScope.
- [Empty `Config.Model` on a hand-built legacy client used to send `"model":""`] → resolution now uses `qwen3-max`, matching `LoadFromEnv`. Callers that set a model are unchanged.
- [Tests that hit HTTP would sleep on the 200ms limiter and the 2s retry] → `NewClient` still uses those production values. Tests set the interval, retry count, and retry base to zero.
- [A future resolver forgets to clear `APIKey` when the user switches provider] → switching provider inside this package already drops the previous key. The resolver should still set the key for the provider it selected.

## Migration Plan

No caller edits. Deploy is a library update. Rollback is reverting the module. Env files gain empty `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `XAI_API_KEY`, `GEMINI_API_KEY`, `OPENROUTER_API_KEY`, `OLLAMA_BASE_URL`, `DASHSCOPE_API_KEY`, `MISTRAL_API_KEY`, `GROQ_API_KEY`, and `OPENAI_COMPATIBLE_API_KEY` (plus optional `*_BASE_URL`s). Unset, they do not change the legacy client.

## Open Questions

None that change the spec or the task breakdown. Model ids in the suggested lists are labels for a settings UI, not a pinned contract.

## Review revisions

Lens review on the first implementation found two holes. Both are fixed in the library, and the spec scenarios for key provenance and stream cancellation match this section.

**Legacy key provenance.** `LoadFromEnv` copies `MORPH_AI_API_KEY`, then `GEMINI_API_KEY`, then `TRAN_QWEN_API_KEY` into `Config.APIKey`. Using that field as an explicit key meant `cfg := LoadFromEnv(); cfg.Provider = openai` sent the DashScope key to OpenAI, and the same copy became an Ollama Bearer token or a DashScope call authorized by a Gemini key. The loader keeps a private snapshot of that key. `resolveLegacy` is the only path that sends it. A named provider drops a key that still equals the snapshot and falls through to that provider's own env var. A different string written to `Config.APIKey` after `LoadFromEnv`, or a per-call `APIKey`, is an explicit key and is sent. Changing the env var after load does not count as that write. The empty-provider path still uses the copied key.

**Legacy base provenance.** Comparing `BaseURL` to the DashScope default was not enough. A custom `MORPH_AI_BASE_URL`, or a compatible-mode `MORPH_AI_API_URL` stored as the base, survived `Provider = openai` and received `OPENAI_API_KEY`. Ollama targeted that URL too. The loader now snapshots the base the same way it snapshots the key. A named provider drops a base that still equals the snapshot and uses that provider's env base or default. A base URL set on `Config` or on the call is kept. DashScope still reads `MORPH_AI_BASE_URL` and `TRAN_QWEN_BASE_URL` from its own env list after the snapshot is dropped.

**Stream cancellation.** The OpenAI and Anthropic readers sent the terminal event with `context.Background()`. If the 16-slot channel was full and the caller had stopped reading, that send blocked, the goroutine never returned, and `resp.Body.Close` never ran. The terminal send now waits on the call context and, once that context is cancelled, tries the channel once without blocking. `defer resp.Body.Close()` therefore runs. A stream that reaches EOF before `[DONE]` or `message_stop` returns an error instead of a clean finish. Empty streamed tool arguments are normalized to `{}`. `rateLimit` records the wait under the mutex, releases it, and sleeps with the call context so cancel returns during the gap.

**Left for a later change.** Validating Anthropic tool JSON on `content_block_stop` would reject argument strings the caller can already see. A separate stream HTTP client (`ResponseHeaderTimeout` only) changes when a long stream is cut off and should not ride along with the leak fix. Bounding successful response bodies with `LimitReader` can truncate a long completion. Redacting keys inside provider error text is worthwhile and was not required to stop the cross-vendor send.
