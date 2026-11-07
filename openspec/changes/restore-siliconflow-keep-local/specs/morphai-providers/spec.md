## MODIFIED Requirements

### Requirement: Named providers
The shared Go client SHALL support these provider ids: `openai`, `anthropic`, `xai`, `gemini`, `openrouter`, `ollama`, `dashscope`, `mistral`, `groq`, `openai-compatible`, and `siliconflow`. Each SHALL expose metadata a settings screen can list: id, display name, whether a key is required, default base URL (empty when the caller must supply one), and suggested models. Ollama SHALL NOT require an API key. OpenRouter SHALL remain its own provider. Mistral and Groq SHALL be named presets that speak the OpenAI chat-completions protocol. SiliconFlow SHALL be a named preset that speaks the OpenAI chat-completions protocol and SHALL require an API key. Its default base URL SHALL be `https://api.siliconflow.cn/v1`. Its suggested models SHALL include `Qwen/Qwen3.6-27B`.

#### Scenario: List includes every named provider
- **WHEN** a caller asks for the provider list
- **THEN** the list contains openai, anthropic, xai, gemini, openrouter, ollama, dashscope, mistral, groq, openai-compatible, and siliconflow
- **AND** ollama is marked as not requiring a key
- **AND** openai-compatible has no default base URL
- **AND** siliconflow requires a key
- **AND** siliconflow's default base URL is `https://api.siliconflow.cn/v1`

#### Scenario: Unknown provider
- **WHEN** a caller selects a provider id that is not in the registry
- **THEN** the call fails with an unknown-provider error
- **AND** no HTTP request is sent

### Requirement: Config resolution without cross-provider fallback
A call SHALL resolve settings in this order: per-call provider, model, API key, and base URL when set; otherwise the client config; otherwise that provider's own environment variables; otherwise the provider default model and base URL. A key or base URL that belongs to a different provider SHALL NOT be used. A key copied by the legacy env loader (`MORPH_AI_API_KEY`, `GEMINI_API_KEY`, or `TRAN_QWEN_API_KEY`) SHALL be valid only while no provider is selected. Selecting a named provider SHALL drop that copied key and SHALL use a caller-set key or that provider's own env var. The loader SHALL remember the key it copied. A later change to those env vars SHALL NOT make the copied key look caller-set. A base URL filled by the same loader (`MORPH_AI_BASE_URL`, `TRAN_QWEN_BASE_URL`, a compatible-mode `MORPH_AI_API_URL`, or the DashScope default used when those are unset) SHALL be valid only while no provider is selected. Selecting a named provider SHALL drop that loaded base URL and SHALL use a base URL the caller set on the config or on the call, or that provider's own env var, or the provider default. If the provider still has no key, the call SHALL fail with the not-configured error. When the selected provider still has no key (or, for Ollama, no base URL), the call SHALL fail with an error that unwraps to the not-configured error, and SHALL NOT contact another provider. Ollama SHALL be configured when a base URL resolves, including its localhost default, with no key. `MORPH_AI_PROVIDER`, when set to a known provider id, SHALL select that provider for a client loaded from the environment. The call SHALL use that provider's own key and base URL environment variables, and SHALL NOT send the legacy key or base URL snapshot. An unknown id SHALL fail with the unknown-provider error and SHALL NOT fall back to DashScope. Whitespace and letter case SHALL NOT change the id. When `MORPH_AI_PROVIDER` is unset or blank, the client SHALL stay on the legacy path. When it is set, `MORPH_AI_MODEL` SHALL be the chat model unless that value is empty or is the DashScope default `qwen3-max` on a non-DashScope provider that has its own default model, in which case that provider's default model SHALL be used, or the selected provider is `siliconflow` and `SILICONFLOW_MODEL` is set, in which case `SILICONFLOW_MODEL` SHALL be the chat model. A model written on a hand-built config SHALL still be sent. `OPENAI_API_KEY` and `OPENAI_BASE_URL` SHALL NOT authorize or target `siliconflow`.

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

#### Scenario: A key loaded from the legacy env is not sent to OpenAI
- **WHEN** a client is loaded from `MORPH_AI_API_KEY` and the provider is then set to openai, and `OPENAI_API_KEY` is unset
- **THEN** the call fails with the not-configured error
- **AND** no HTTP request is sent

#### Scenario: A key loaded from the legacy env is not attached to Ollama
- **WHEN** a client is loaded from `MORPH_AI_API_KEY` and the provider is then set to ollama
- **THEN** the request omits the Authorization header

#### Scenario: The Gemini fallback does not authorize explicit DashScope
- **WHEN** the only key in the environment is `GEMINI_API_KEY`, the client is loaded from the environment, and the provider is then set to dashscope
- **THEN** the call fails with the not-configured error
- **AND** no HTTP request is sent

#### Scenario: A caller-set key still wins
- **WHEN** a client is loaded from the legacy env and the caller then sets a different API key for the selected provider
- **THEN** the request uses that caller key

#### Scenario: A custom legacy base URL is not used for OpenAI
- **WHEN** `MORPH_AI_BASE_URL` points at one server, the client is loaded from the environment, the provider is then set to openai, and `OPENAI_BASE_URL` points at a different server
- **THEN** the request is sent only to the OpenAI server
- **AND** the OpenAI key is not sent to the legacy server

#### Scenario: A custom legacy base URL is not used for Ollama
- **WHEN** `MORPH_AI_BASE_URL` points at one server, the client is loaded from the environment, the provider is then set to ollama, and Ollama's base URL points at a different server
- **THEN** the request is sent only to the Ollama server
- **AND** the request omits the Authorization header

#### Scenario: A compatible-mode legacy API URL is not used for OpenAI
- **WHEN** `MORPH_AI_API_URL` is an OpenAI-compatible URL for one server, the client is loaded from the environment, the provider is then set to openai, and `OPENAI_BASE_URL` points at a different server
- **THEN** the request is sent only to the OpenAI server

#### Scenario: An explicit base URL still wins
- **WHEN** a client is loaded from a legacy base URL and the caller then sets a different base URL for the selected provider
- **THEN** the request is sent to that caller base URL

#### Scenario: Rotating the legacy env key after load does not authorize another provider
- **WHEN** a client is loaded from `MORPH_AI_API_KEY` and that env var is then changed, and the provider is set to openai with no OpenAI key
- **THEN** the call fails with the not-configured error
- **AND** no HTTP request is sent

#### Scenario: MORPH_AI_PROVIDER selects that provider's endpoint
- **WHEN** `MORPH_AI_PROVIDER` is `xai`, `XAI_API_KEY` is set, `XAI_BASE_URL` points at one server, and `MORPH_AI_BASE_URL` points at another
- **THEN** the request is sent only to the xAI server
- **AND** the Authorization header uses the xAI key

#### Scenario: MORPH_AI_PROVIDER does not borrow the legacy key
- **WHEN** `MORPH_AI_PROVIDER` is `xai` and the only key set is `MORPH_AI_API_KEY`
- **THEN** the call fails with the not-configured error
- **AND** no HTTP request is sent

#### Scenario: An unknown MORPH_AI_PROVIDER is an error
- **WHEN** `MORPH_AI_PROVIDER` is not a known provider id
- **THEN** the call fails with the unknown-provider error
- **AND** no HTTP request is sent
- **AND** the client does not call DashScope

#### Scenario: The DashScope default model is not sent to another provider
- **WHEN** `MORPH_AI_PROVIDER` is `xai` and `MORPH_AI_MODEL` is unset or `qwen3-max`
- **THEN** the request uses xAI's default model

#### Scenario: An explicit MORPH_AI_MODEL is sent
- **WHEN** `MORPH_AI_PROVIDER` is `xai` and `MORPH_AI_MODEL` is a model other than `qwen3-max`
- **THEN** the request uses that model

#### Scenario: SiliconFlow ignores the parked DashScope profile
- **WHEN** `MORPH_AI_PROVIDER` is `siliconflow`, `SILICONFLOW_API_KEY` is set, `SILICONFLOW_MODEL` is `Qwen/Qwen3.6-27B`, `MORPH_AI_API_KEY` is set, `MORPH_AI_MODEL` is `qwen3.7-plus`, and `MORPH_AI_BASE_URL` is the DashScope compatible-mode URL
- **THEN** the chat request is sent only to `https://api.siliconflow.cn/v1/chat/completions`
- **AND** the Authorization header uses the SiliconFlow key
- **AND** the request model is `Qwen/Qwen3.6-27B`

#### Scenario: SiliconFlow does not borrow the OpenAI slot
- **WHEN** `MORPH_AI_PROVIDER` is `siliconflow`, `SILICONFLOW_API_KEY` is unset, and `OPENAI_API_KEY` plus `OPENAI_BASE_URL` point at SiliconFlow
- **THEN** the call fails with the not-configured error
- **AND** no HTTP request is sent

#### Scenario: A hand-built SiliconFlow model still wins
- **WHEN** a caller builds a siliconflow config with a model other than `Qwen/Qwen3.6-27B` and `SILICONFLOW_MODEL` is also set
- **THEN** the request uses the caller model

#### Scenario: DashScope still uses the parked local profile
- **WHEN** `MORPH_AI_PROVIDER` is `dashscope`, `MORPH_AI_API_KEY` is set, `MORPH_AI_MODEL` is `qwen3.7-plus`, `MORPH_AI_BASE_URL` is the DashScope compatible-mode URL, and `SILICONFLOW_API_KEY` is set
- **THEN** the chat request is sent only to that DashScope URL plus `/chat/completions`
- **AND** the Authorization header uses the Morph key
- **AND** the request model is `qwen3.7-plus`

#### Scenario: Ollama stays available beside SiliconFlow
- **WHEN** `MORPH_AI_PROVIDER` is `ollama`, `SILICONFLOW_API_KEY` is set, and Ollama's base URL is its localhost default
- **THEN** the chat request is sent only to the Ollama base plus `/chat/completions`
- **AND** the request omits the Authorization header

## ADDED Requirements

### Requirement: SiliconFlow disables Qwen thinking
A SiliconFlow text chat request SHALL disable Qwen thinking, the same way a DashScope compatible-mode text chat request does.

#### Scenario: SiliconFlow text chat turns thinking off
- **WHEN** a client loaded with `MORPH_AI_PROVIDER=siliconflow` and a SiliconFlow key sends a text chat
- **THEN** Qwen thinking is disabled on that request
