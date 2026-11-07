# Design

## Context

See proposal.md. Today Dorph AI uses one `dorphai.Client` from `LoadFromEnv()` inside `chatWithManagementTools` and `distillSessionLesson`. Skills come from `ai_skills` + picker `skill_ids`; durable memory from `agent_lesson` (harvest after significant sessions). `RouteSubAgents` in `agent_context.go` is keyword-based, not model-routed. SiliconFlow support exists in Go `pkg/dorphai`; NVIDIA integrate is not registered yet. Rust `pkg/dorphai-rs` remains single-profile (Project API)—not the first implementation target.

## Goals / Non-Goals

**Goals:**

- Primary model: NVIDIA `integrate.api.nvidia.com/v1`, model `nvidia/nemotron-3-ultra-550b-a55b`, with `enable_thinking` via request extra body and streaming support for reasoning vs answer text.
- Coordinator model: SiliconFlow (e.g. `Qwen/Qwen3-8B`) for cheap structured side work: turn plan, working-memory patch, sub-agent task list, lesson distill.
- Keep the existing JSON tool loop and `/api/tran/*` catalog; coordinator does not execute tools directly.
- Skills system unchanged in storage; injection order documented and tested.
- Memory: durable `agent_lesson` + new session **working memory** blob (coordinator-maintained summary of facts/decisions for the session).

**Non-Goals:**

- Replacing Dorph AI workspace UI (`dorphai-agent-workspace`) or restoring the Files tab.
- New OS processes or Cursor-style coding agents.
- Storing API keys in repo, Render yaml, or OpenSpec artifacts.
- Forcing Project (`dorph-engi`) onto dual-model in v1 (optional follow-up in tasks).

## Decisions

### 1. Two resolved configs, one agent loop (recommended over a new runtime)

**Choice:** Add `dorphai.Config` resolution helpers `LoadOrchestratorFromEnv()` and `LoadCoordinatorFromEnv()` (names TBD) mapping to:

| Role | Env surface | Default when unset |
|------|-----------|-------------------|
| Orchestrator | `DORPH_AI_ORCHESTRATOR_PROVIDER=nvidia` + `NVIDIA_API_KEY`, optional `NVIDIA_BASE_URL`, `NVIDIA_MODEL` | Falls back to legacy single client |
| Coordinator | `DORPH_AI_COORDINATOR_PROVIDER=siliconflow` + existing `SILICONFLOW_*` | Falls back to orchestrator client |

**Why:** Reuses `chatWithManagementTools`; avoids a second agent framework. Matches user ask: Nemotron “base”, SiliconFlow “sub agent” for coordination/memory.

**Alternatives considered:**

- **Single model with role prompts only** — cheaper but fails the explicit dual-model requirement and overloads one context window.
- **Separate microservice for coordinator** — ops weight; YAGNI for local `start-all.sh`.
- **SiliconFlow primary, NVIDIA sub** — inverted vs user intent.

### 2. Turn pipeline (ECC-style: explicit phases, one owner for tools)

```
User message
  → Coordinator (optional): emit { plan, memory_patch, sub_agents[] } as JSON
  → Apply memory_patch to session working memory (SQLite)
  → Merge: skills + lessons + working memory + hybrid context into orchestrator system prompt
  → Orchestrator loop: tool JSON rounds until prose reply
  → Coordinator (optional): post-turn memory_patch if significant
  → maybeHarvestSession (distill) uses coordinator client when configured
```

**Why:** Coordinator never competes with orchestrator for final prose; tool execution stays on orchestrator only (one place to audit).

**Alternative:** Coordinator runs tools — rejected (two actors calling `/api/tran` complicates auth and logging).

### 3. NVIDIA provider in `pkg/dorphai`

**Choice:** Register `ProviderNVIDIA` with default base `https://integrate.api.nvidia.com/v1`, key env `NVIDIA_API_KEY`, model env `NVIDIA_MODEL` defaulting to `nvidia/nemotron-3-ultra-550b-a55b`. OpenAI-compatible chat completions with extra body:

```json
{ "chat_template_kwargs": { "enable_thinking": true } }
```

**Streaming:** Extend client with optional `ChatCompletionStream` returning `(reasoningChunk, contentChunk, err)`; Dorph AI frontend can show thinking in a collapsible block in a later task.

**Why:** Matches the user’s Python sample; keeps one HTTP path in Go.

### 4. Memory model

**Choice:**

- **Durable:** keep `agent_lesson` (trigger + rule); distill via coordinator when `DORPH_AI_COORDINATOR_*` configured, else orchestrator/legacy.
- **Working:** new table `agent_working_memory (session_id, owner_user_id, summary, updated_at)` updated only from coordinator JSON `memory_patch` (max length cap, e.g. 4k chars).
- **Injection:** orchestrator system prompt: enabled skills (existing) → last N lessons → working memory summary → coordinator plan (one paragraph).

**Why:** User asked for memory “alongside skills”; lessons already exist; working memory fills session continuity without distilling every turn.

**Alternative:** Vector store — deferred; GraphRAG stays for knowledge library, not chat scratchpad.

### 5. Sub-agents

**Choice:** When coordinator returns `sub_agents`, map to existing worker ids (`files`, `knowledge`, `notes`, `dorphdata`) and run `RouteSubAgents` logic **or** coordinator list, whichever is narrower (intersection). Cap still `maxSubAgents = 3`. Workers prefetch context blobs; orchestrator sees `TOOL_RESULT`-style summaries in the user message prefix.

**Why:** Reuses `dorphai-agent-workspace` direction without requiring the workspace UI to ship first.

### 6. Security

- Keys only in `.env` / host env; document `NVIDIA_API_KEY` in `.env.example` empty.
- Log model id and role, never keys.
- User pasted a live NVIDIA key in chat — **rotate it** in the NVIDIA console; do not write it to disk in this change.

## Risks / Trade-offs

- [Latency] Two model calls per turn → run coordinator with tight max tokens; skip on `isSimpleChat`.
- [Cost] Nemotron ultra + SF → gate coordinator on non-trivial turns; document env toggles.
- [Rust apps out of sync] Project still single-client until `dorphai-rs` dual profile → document in tasks.
- [Streaming UX incomplete] v1 may buffer stream server-side → acceptable if API returns full text first.
- [Coordinator JSON parse failures] → log and continue with keyword `RouteSubAgents` only.

## Migration Plan

1. Ship `pkg/dorphai` NVIDIA provider + role loaders behind env flags.
2. Wire Dorph handlers; default off → no behavior change.
3. Enable orchestrator + coordinator in local `.env`; restart `dorph-api`.
4. Add tests with httptest transports for both roles.
5. Rollback: unset `DORPH_AI_ORCHESTRATOR_PROVIDER`; single client path remains.

## Open Questions

- Whether Dorph AI streaming UI ships in the same change or immediately after (tasks split).
- Exact coordinator JSON schema versioning (`coord_v1`) — fixed in spec scenarios.
