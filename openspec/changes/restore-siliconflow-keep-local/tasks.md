## 1. SiliconFlow provider

- [x] 1.1 Add provider id `siliconflow` to the catalog: OpenAI chat-completions, key required, default base `https://api.siliconflow.cn/v1`, key env `SILICONFLOW_API_KEY`, base env `SILICONFLOW_BASE_URL`, first suggested model `Qwen/Qwen3.6-27B`. Update the provider-list test so the list includes it and still marks Ollama as not requiring a key.
- [x] 1.2 When the loaded provider is `siliconflow` and `SILICONFLOW_MODEL` is set, use that model instead of `MORPH_AI_MODEL`. A model set on a hand-built config is still sent. Other providers keep the current `MORPH_AI_MODEL` rule.
- [x] 1.3 Set `enable_thinking` to false on SiliconFlow text chat, and leave it off for OpenAI, xAI, and Ollama.

## 2. Tests for both profiles

- [x] 2.1 httptest: `MORPH_AI_PROVIDER=siliconflow`, `SILICONFLOW_MODEL=Qwen/Qwen3.6-27B`, `MORPH_AI_MODEL=qwen3.7-plus`, and `MORPH_AI_BASE_URL` on DashScope. The chat goes only to `https://api.siliconflow.cn/v1/chat/completions` with the SiliconFlow key, model `Qwen/Qwen3.6-27B`, and `enable_thinking` false.
- [x] 2.2 httptest: siliconflow selected, `SILICONFLOW_API_KEY` unset, `OPENAI_API_KEY` and `OPENAI_BASE_URL` set. The call is not-configured and sends no request.
- [x] 2.3 httptest: `MORPH_AI_PROVIDER=dashscope` with the parked `MORPH_AI_*` DashScope values and a SiliconFlow key also set. The chat uses the DashScope URL, the Morph key, and `qwen3.7-plus`.
- [x] 2.4 httptest: `MORPH_AI_PROVIDER=ollama` with a SiliconFlow key set. The chat goes to the Ollama base and omits Authorization.

## 3. Local env and docs

- [x] 3.1 In the gitignored repo-root `.env`, set `MORPH_AI_PROVIDER=siliconflow` and `SILICONFLOW_MODEL=Qwen/Qwen3.6-27B`. Copy `OPENAI_API_KEY` into `SILICONFLOW_API_KEY` only if `OPENAI_BASE_URL` is `https://api.siliconflow.cn/v1`. Do not print the key. Leave `MORPH_AI_API_KEY`, `MORPH_AI_MODEL`, `MORPH_AI_BASE_URL`, and `OPENAI_*` unchanged.
- [x] 3.2 Document commented `SILICONFLOW_API_KEY`, `SILICONFLOW_BASE_URL`, and `SILICONFLOW_MODEL` in `.env.example` and the provider row in `pkg/morphai/README.md`. Do not set `MORPH_AI_PROVIDER` in the example file. Do not change `render.yaml`.

## 4. Check

- [x] 4.1 Run `go test ./pkg/morphai -count=1`. Confirm `render.yaml` still sets `MORPH_AI_PROVIDER` to `dashscope`.
