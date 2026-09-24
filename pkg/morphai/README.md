# morphai

Shared Go client for Morph AI backends (`github.com/robo/morphai`).

An empty provider keeps today's DashScope client: `MORPH_AI_*`, with `GEMINI_API_KEY` / `GEMINI_MODEL` and `TRAN_QWEN_*` as fallbacks. Set `Config.Provider` (or a per-call provider) to talk to a named vendor. There is no silent switch to another provider when a key is missing.

## Providers

| ID | Protocol | Key env | Default base URL |
| --- | --- | --- | --- |
| `openai` | OpenAI chat completions | `OPENAI_API_KEY` | `https://api.openai.com/v1` |
| `anthropic` | Anthropic Messages | `ANTHROPIC_API_KEY` | `https://api.anthropic.com` |
| `xai` | OpenAI chat completions | `XAI_API_KEY` | `https://api.x.ai/v1` |
| `gemini` | OpenAI chat completions | `GEMINI_API_KEY` | `https://generativelanguage.googleapis.com/v1beta/openai` |
| `openrouter` | OpenAI chat completions | `OPENROUTER_API_KEY` | `https://openrouter.ai/api/v1` |
| `ollama` | OpenAI chat completions | none | `http://127.0.0.1:11434/v1` |
| `dashscope` | OpenAI chat completions, or native text-generation when `MORPH_AI_API_URL` is the generation endpoint | `DASHSCOPE_API_KEY`, then `MORPH_AI_API_KEY`, then `TRAN_QWEN_API_KEY` | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| `mistral` | OpenAI chat completions | `MISTRAL_API_KEY` | `https://api.mistral.ai/v1` |
| `groq` | OpenAI chat completions (no vision) | `GROQ_API_KEY` | `https://api.groq.com/openai/v1` |
| `openai-compatible` | OpenAI chat completions | `OPENAI_COMPATIBLE_API_KEY` | caller must set a base URL |

Each provider also reads `<KEY>_BASE_URL` when the config base URL is empty (`OPENAI_BASE_URL`, `ANTHROPIC_BASE_URL`, and so on). `ListProviders` returns ids, display names, whether a key is required, the default base URL, suggested models, and capabilities.

Mistral and Groq are named presets on the OpenAI-compatible adapter. OpenRouter stays its own provider. Gemini uses Google's OpenAI-compatible endpoint, not the native `generateContent` API. Anthropic uses the Messages API.

## Resolution

1. Per-call fields on `CompletionRequest` (provider, model, API key, base URL) when set.
2. Fields on the client `Config`.
3. That provider's own environment variables.
4. The provider's default model and base URL.

A key from another provider is not reused. Selecting `openai` on a client that only has `MORPH_AI_API_KEY` returns `ErrProviderNotConfigured` and does not call DashScope. Ollama is configured with no key when a base URL resolves. `openai-compatible` needs both a key and a base URL.

`LoadFromEnv` remembers the key and base URL it copied. That snapshot is used only while `Provider` is empty. A custom `MORPH_AI_BASE_URL`, a compatible-mode `MORPH_AI_API_URL`, and the DashScope default filled in when neither is set are all part of that snapshot. Setting a named provider drops them. The call then uses a base URL the caller set on `Config` or on that call, or that provider's own env var, or the provider default. The same rule applies to the key copied from `MORPH_AI_API_KEY`, `GEMINI_API_KEY`, or `TRAN_QWEN_API_KEY`: a later change to those env vars does not turn the loaded copy into a caller key. `GEMINI_API_KEY` does not authorize explicit DashScope. `TRAN_QWEN_API_KEY` and `MORPH_AI_API_KEY` still do, because those are DashScope env vars, not the copied field.

A per-call provider switch also drops the client's model and base URL, so a Qwen model name does not follow the call to another vendor. Passing `dashscope` on a legacy client keeps the client's key and URLs.

The unscoped client (empty provider) is unchanged: missing key text is `MORPH_AI_API_KEY is not configured`, which also unwraps to `ErrProviderNotConfigured`. Native DashScope vision still returns the existing `MORPH_AI_API_URL` explanation.

## Calling

```go
client := morphai.NewClientFromEnv()
text, err := client.ChatCompletion(ctx, []morphai.Message{
    {Role: "user", Content: "Hello"},
})

openai := morphai.NewClient(morphai.Config{
    Provider: morphai.ProviderOpenAI,
    Model:    "gpt-4.1",
})
resp, err := openai.Complete(ctx, morphai.CompletionRequest{
    Messages: []morphai.Message{{Role: "user", Content: "Hello"}},
})

ch, err := openai.Stream(ctx, morphai.CompletionRequest{
    Messages: []morphai.Message{{Role: "user", Content: "Hello"}},
})
```

`ChatCompletion`, `ChatCompletionLong`, and `ChatCompletionVision` keep their signatures. `Complete` / `CompleteLong` add tools and JSON mode. `Stream` emits `StreamEvent` values; `CollectStream` folds them into one `CompletionResponse`. Cancelling the stream context returns the reader goroutine and closes the response body even if the caller has stopped reading. A stream that closes before `[DONE]` or `message_stop` is an error, not a finished reply. Empty streamed tool-call arguments are `{}`. A capability the provider does not implement returns `ErrCapabilityUnsupported` (Anthropic has no JSON mode; native DashScope has chat only; Groq has no vision).

HTTP failures are `*APIError` with status, code, and message. The error string stays `API error (status N): code - message`.

## Later stored keys

This package does not store or encrypt keys. A later settings resolver fills `Config.APIKey` or `CompletionRequest.APIKey` before the call. Intended precedence once both stored sources exist: per-call key, then per-user key, then admin/workspace key, then the env fallback above. Admin/workspace keys are the first stored source to add; per-user keys come after. Encryption belongs to a separate secrets package.

## Defaults

| Setting | Value |
| --- | --- |
| Legacy model | `qwen3-max` |
| Legacy vision model | `qwen-vl-max` |
| Chat timeout | 120s |
| Long / vision timeout | 300s |
| Retries | 3, on 429 and transport errors |
| Minimum spacing | 200ms between requests on one client |
