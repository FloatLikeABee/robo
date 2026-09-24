# morphai-providers Specification

## Purpose

Lets Morph backends call more than one model vendor through the shared Go client, while today's DashScope env setup keeps working and a missing key is reported instead of being routed elsewhere.

## Requirements

### Requirement: Named providers
The shared Go client SHALL support these provider ids: `openai`, `anthropic`, `xai`, `gemini`, `openrouter`, `ollama`, `dashscope`, `mistral`, `groq`, and `openai-compatible`. Each SHALL expose metadata a settings screen can list: id, display name, whether a key is required, default base URL (empty when the caller must supply one), and suggested models. Ollama SHALL NOT require an API key. OpenRouter SHALL remain its own provider. Mistral and Groq SHALL be named presets that speak the OpenAI chat-completions protocol.

#### Scenario: List includes every named provider
- **WHEN** a caller asks for the provider list
- **THEN** the list contains openai, anthropic, xai, gemini, openrouter, ollama, dashscope, mistral, groq, and openai-compatible
- **AND** ollama is marked as not requiring a key
- **AND** openai-compatible has no default base URL

#### Scenario: Unknown provider
- **WHEN** a caller selects a provider id that is not in the registry
- **THEN** the call fails with an unknown-provider error
- **AND** no HTTP request is sent

### Requirement: Stable client for existing callers
Callers that construct the client from `MORPH_AI_*` (and the legacy `GEMINI_API_KEY`, `GEMINI_MODEL`, and `TRAN_QWEN_*` fallbacks) without selecting a provider SHALL keep today's behavior: default model `qwen3-max`, DashScope compatible-mode chat completions unless `MORPH_AI_API_URL` points at the native text-generation endpoint, the same missing-key error text containing `MORPH_AI_API_KEY`, and the same vision rejection text when the native endpoint is selected. Existing chat, long-timeout, and vision methods SHALL keep their signatures.

#### Scenario: Env-only client still calls DashScope compatible-mode
- **WHEN** a client is loaded from the environment with `MORPH_AI_API_KEY` set and no provider id
- **THEN** chat completions POST to the configured compatible-mode base URL plus `/chat/completions`
- **AND** the Authorization header is `Bearer` plus that key
- **AND** Qwen thinking is disabled on that text chat request

#### Scenario: Native DashScope vision still explains the endpoint
- **WHEN** a client is configured for the native DashScope text-generation endpoint and a vision call is made
- **THEN** the error text contains `MORPH_AI_API_URL`
- **AND** no image payload is sent

### Requirement: Config resolution without cross-provider fallback
A call SHALL resolve settings in this order: per-call provider, model, API key, and base URL when set; otherwise the client config; otherwise that provider's own environment variables; otherwise the provider default model and base URL. A key or base URL that belongs to a different provider SHALL NOT be used. When the selected provider still has no key (or, for Ollama, no base URL), the call SHALL fail with an error that unwraps to the not-configured error, and SHALL NOT contact another provider. Ollama SHALL be configured when a base URL resolves, including its localhost default, with no key.

#### Scenario: Explicit OpenAI key beats the OpenAI env var
- **WHEN** the caller passes an API key and `OPENAI_API_KEY` is also set
- **THEN** the request Authorization header uses the passed key

#### Scenario: OpenAI does not borrow the Morph key
- **WHEN** the caller selects openai, passes no API key, `OPENAI_API_KEY` is unset, and `MORPH_AI_API_KEY` is set
- **THEN** the call fails with the not-configured error
- **AND** no HTTP request is sent

#### Scenario: Ollama needs no key
- **WHEN** the caller selects ollama and passes no API key
- **THEN** the client is configured
- **AND** the request omits the Authorization header
- **AND** the URL is the Ollama base plus `/chat/completions`

#### Scenario: Generic compatible endpoint requires a base URL
- **WHEN** the caller selects openai-compatible and supplies a key but no base URL and no `OPENAI_COMPATIBLE_BASE_URL`
- **THEN** the call fails with the not-configured error

### Requirement: Chat, stream, tools, vision, and JSON mode
Chat completions, streaming, tool calls, vision, and JSON mode SHALL work through the selected provider when that provider's protocol supports them. Anthropic SHALL use its Messages API, including its tool-use content blocks. Providers that share the OpenAI chat-completions protocol SHALL use that protocol. A request for a capability the selected provider lacks SHALL fail with an error that unwraps to the capability-unsupported error and SHALL NOT be sent. Provider HTTP errors SHALL surface status, code, and message.

#### Scenario: OpenAI-compatible tool round trip
- **WHEN** a caller sends tools and the provider returns a tool call, then the caller sends the tool result
- **THEN** the first request includes the tool definition
- **AND** the second request includes the tool result in that provider's wire format
- **AND** the final assistant text is returned

#### Scenario: Streaming text is reassembled
- **WHEN** a provider streams text as server-sent events
- **THEN** the caller can reassemble the full text from the deltas

#### Scenario: Anthropic rejects JSON mode
- **WHEN** a caller asks Anthropic for JSON mode
- **THEN** the call fails with the capability-unsupported error
- **AND** no HTTP request is sent

#### Scenario: HTTP error mapping
- **WHEN** a provider responds with a non-success status and an error message
- **THEN** the error includes that status and message

### Requirement: Later key sources plug in without storage here
This library SHALL NOT store, encrypt, or read keys from a database. A later story SHALL be able to supply an already-resolved key on the client config or on the call. The intended precedence for that later resolver is: per-call key, then a stored per-user key, then a stored admin or workspace key, then the env fallback this story implements. Admin and workspace keys are the first stored source to be added; per-user keys come after. Encryption belongs to a separate secrets package.

#### Scenario: A pre-filled key is used as-is
- **WHEN** a caller sets the API key on the client config for a named provider
- **THEN** that key is sent
- **AND** the library does not write it to disk
