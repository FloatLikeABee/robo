## Why

Morph AI on this machine used to call SiliconFlow. The repo-root `.env` now sends the env-loaded client to DashScope (`qwen3.7-plus` on the DashScope compatible-mode URL), while the SiliconFlow key, base URL, and model sit on the OpenAI-compatible variables, which that client does not use. The DashScope profile and the local Ollama provider should stay available.

## What Changes

- Add `siliconflow` as a named MorphAI provider that talks to `https://api.siliconflow.cn/v1` with its own key, base URL, and model variables.
- Select that provider for the local env-loaded client, using the SiliconFlow model already stored locally (`Qwen/Qwen3.6-27B`), and send the same Qwen thinking-off flag the previous SiliconFlow setup used.
- Leave the current DashScope `MORPH_AI_API_KEY`, `MORPH_AI_MODEL`, and `MORPH_AI_BASE_URL` values in the local env file, and leave the existing `OPENAI_*` SiliconFlow lines in place. Ollama stays a named provider and is not selected.
- Hosted Morph stays on `MORPH_AI_PROVIDER=dashscope`. This does not change `render.yaml`.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `morphai-providers`: The provider list includes `siliconflow`. Selecting it uses SiliconFlow's own key, base URL, and model, including when `MORPH_AI_MODEL` still names the local DashScope model. SiliconFlow chat requests disable Qwen thinking. DashScope and Ollama resolution stay as they are when those providers are selected.

## Impact

- `pkg/morphai` provider catalog, env loading, and request flags. Tests in that package.
- `pkg/morphai/README.md` and `.env.example`.
- The gitignored repo-root `.env` (select SiliconFlow, keep the DashScope and OpenAI-compatible lines).
- Morph, Event Logs, Content Maker, and other callers of `LoadFromEnv` pick up the selected provider. No new HTTP API.
- `render.yaml` and the Morph container contract stay on DashScope.
