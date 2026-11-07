## Context

See proposal.md for why. `pkg/morphai` already has named providers. `LoadFromEnv` with an empty `MORPH_AI_PROVIDER` sends `MORPH_AI_API_KEY`, `MORPH_AI_MODEL`, and `MORPH_AI_BASE_URL`. On this machine those three are the DashScope profile (`qwen3.7-plus`, DashScope compatible-mode). SiliconFlow is already stored beside them as `OPENAI_API_KEY`, `OPENAI_BASE_URL=https://api.siliconflow.cn/v1`, and `OPENAI_MODEL=Qwen/Qwen3.6-27B`. The legacy client does not read that OpenAI slot. Ollama is a named provider with a localhost default and is not selected. `render.yaml` pins hosted Morph to `dashscope`.

Compatible-mode DashScope, and the legacy path even when its base URL is SiliconFlow, set `enable_thinking` to false. A named non-DashScope provider omits that field today.

## Goals / Non-Goals

**Goals:**

- Env-loaded Morph AI on this machine calls SiliconFlow with `Qwen/Qwen3.6-27B` and thinking off.
- The DashScope `MORPH_AI_*` values stay in the local env and still win when the provider is `dashscope`.
- Ollama remains selectable and is not probed or preferred.

**Non-Goals:**

- Changing `render.yaml` or the Morph container contract.
- Making Ollama primary, or falling back from SiliconFlow to Ollama.
- Reading `OPENAI_*` as the SiliconFlow credentials inside the client.
- A settings UI.

## Decisions

### 1. Named `siliconflow` provider, not a rewrite of `MORPH_AI_*`

Add `siliconflow` next to the other OpenAI-compatible presets. Key env `SILICONFLOW_API_KEY`, base env `SILICONFLOW_BASE_URL`, default base `https://api.siliconflow.cn/v1`, first suggested model `Qwen/Qwen3.6-27B`. `LoadFromEnv` already drops the legacy key and base snapshot for a named provider, so the DashScope lines can stay in `.env`.

**Why not point `MORPH_AI_BASE_URL` at SiliconFlow:** that slot is the local DashScope profile. Overwriting it, or parking the model only in a comment, drops a working config. One `MORPH_AI_MODEL` cannot be `qwen3.7-plus` and `Qwen/Qwen3.6-27B` at the same time.

**Why not `MORPH_AI_PROVIDER=openai`:** the OpenAI slot is already the SiliconFlow host on this machine, but the id would mean OpenAI, and `MORPH_AI_MODEL=qwen3.7-plus` would still be sent. Leaving `OPENAI_*` untouched keeps that slot as it is.

**Why not invert the unmerged Ollama-primary probe:** that probed localhost on every chat and preferred Ollama. This request puts SiliconFlow back in front and keeps the local profiles, without a new probe.

### 2. `SILICONFLOW_MODEL` wins over the parked `MORPH_AI_MODEL`

When the loaded provider is `siliconflow` and `SILICONFLOW_MODEL` is set, that value is the chat model. `qwen3.7-plus` stays in `MORPH_AI_MODEL` for DashScope. A model set on a hand-built config is still sent. Other providers keep the current rule, including "an explicit `MORPH_AI_MODEL` other than `qwen3-max` is sent".

### 3. SiliconFlow text chat sets `enable_thinking` to false

Same flag the previous SiliconFlow setup got on the legacy path, and the same flag DashScope compatible-mode sends. Do this only for `siliconflow` and the existing DashScope cases. Do not start sending it for OpenAI, xAI, or Ollama.

### 4. Local env selects it; the example file only documents it

In the gitignored repo-root `.env`, set `MORPH_AI_PROVIDER=siliconflow`, copy the current `OPENAI_API_KEY` into `SILICONFLOW_API_KEY` only because `OPENAI_BASE_URL` is `https://api.siliconflow.cn/v1`, and set `SILICONFLOW_MODEL=Qwen/Qwen3.6-27B`. Leave `MORPH_AI_API_KEY`, `MORPH_AI_MODEL`, `MORPH_AI_BASE_URL`, and `OPENAI_*` unchanged. `.env.example` gains commented `SILICONFLOW_*` lines and does not select the provider. Hosted Morph stays `dashscope`.

## Risks / Trade-offs

- [A SiliconFlow model rejects `enable_thinking`] → The previous working SiliconFlow path sent it. If one model rejects the field, stop sending it for that model in a follow-up; do not drop it for DashScope.
- [The OpenAI key is not a SiliconFlow key] → Copy it only when `OPENAI_BASE_URL` is `https://api.siliconflow.cn/v1`. The client never reads `OPENAI_*` for this provider.
- [`qwen3.7-plus` is sent to SiliconFlow] → The loader must apply `SILICONFLOW_MODEL` after the `MORPH_AI_MODEL` copy. A test covers that pair of values.
- [Restart required] → Env is read at process start. `start-all.sh` must be restarted after the `.env` edit or the old DashScope client stays in memory.

## Migration Plan

1. Ship the provider and tests.
2. Edit the local `.env` as in decision 4. Do not commit it.
3. Restart Morph (and any other process that loaded MorphAI from the env).
4. Rollback: remove `MORPH_AI_PROVIDER` (or set it to `dashscope`). The DashScope lines are still there. `SILICONFLOW_*` can stay.

## Open Questions

None.
