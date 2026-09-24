# Proposal

## Why

`pkg/morphai` can talk to one endpoint: DashScope/Qwen via `MORPH_AI_*` (with legacy `GEMINI_API_KEY` / `TRAN_QWEN_*` fallbacks). A later Settings UI needs named providers so an operator can pick a vendor and model and supply a key. This story is the Go library layer only (issue #21, epic #5).

## What Changes

- Add a provider registry and one client interface covering OpenAI, Anthropic, xAI, Google Gemini, OpenRouter, Ollama, DashScope/Qwen, Mistral, Groq, and a generic OpenAI-compatible endpoint.
- Let callers pass provider id, model, API key, and optional base URL on the client or on a single call. Empty provider keeps today's env-only behavior with no caller edits.
- Return `ErrProviderNotConfigured` when the selected provider has no key (Ollama needs a base URL, not a key). Never silently send the call to a different provider.
- Keep chat, vision, and the existing retry/error text. Add streaming, tool calls, and JSON mode where the provider's protocol supports them, and a typed capability error where it does not.
- Publish provider metadata (id, display name, key required, default base URL, suggested models) for a future settings screen.
- Document env vars. Do not store or encrypt keys.

## Capabilities

### New Capabilities

- `morphai-providers`: Named provider selection, config resolution, chat/stream/tools/vision/JSON behavior, and the not-configured error for the shared Go client.

### Modified Capabilities

- (none — `openspec/specs/` has no baseline for this client)

## Impact

- `pkg/morphai` (library, tests, README)
- Root `.env.example` and the agent docs that describe the Go client (`docs/agents/02-ai-integration.md`, `docs/agents/10-shared-libraries.md`)
- Callers in `morph/`, `formx/`, and `composerx/` keep compiling against the existing `Client` methods
- Out of scope: Settings UI, key storage, encryption, `pkg/morphai-rs`, `morph/frontend/`, `.github/`
