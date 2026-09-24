# Tasks

## 1. Registry and resolution

- [x] 1.1 Add the provider registry (openai, anthropic, xai, gemini, openrouter, ollama, dashscope, mistral, groq, openai-compatible) and verify `ListProviders` / `LookupProvider` cover ids, key requirement, default base URLs, and suggested models.
- [x] 1.2 Resolve per-call, then client config, then that provider's env, then defaults, and verify OpenAI does not use `MORPH_AI_API_KEY`, DashScope does not use `GEMINI_API_KEY`, Ollama needs no key, and a missing key or compatible base URL returns `ErrProviderNotConfigured` with no HTTP call.
- [x] 1.3 Keep `LoadFromEnv` on the legacy path and verify existing config tests still pass (`go test ./pkg/morphai -count=1 -run 'TestLoadFromEnv|TestVision'`).

## 2. Client and adapters

- [x] 2.1 Keep `ChatCompletion`, `ChatCompletionLong`, and `ChatCompletionVision` signatures, and verify a legacy compatible-mode call still posts `/chat/completions` with Bearer auth and `enable_thinking: false`, and native vision still mentions `MORPH_AI_API_URL`.
- [x] 2.2 Route named providers through the OpenAI-compatible adapter or the Anthropic Messages adapter, and verify each preset's auth header, URL, streaming parse, tool round trip, and HTTP error mapping with httptest.
- [x] 2.3 Reject unsupported capabilities (Anthropic JSON mode, native DashScope stream/tools/vision/JSON) with `ErrCapabilityUnsupported` and verify no request is sent.

## 3. Docs

- [x] 3.1 Document providers, resolution order, the later stored-key plug-in (admin/workspace first, per-user later, no encryption here), and empty env vars in `pkg/morphai/README.md`, `docs/agents/02-ai-integration.md`, `docs/agents/10-shared-libraries.md`, and `.env.example`.

## 4. Integration

- [x] 4.1 Run `go test ./...` and `go vet ./...` in `pkg/morphai`, `morph`, `formx/backend`, and `composerx/backend`, and record the results.
