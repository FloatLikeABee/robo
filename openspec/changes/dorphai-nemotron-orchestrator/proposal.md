# Proposal

## Why

Dorph AI already runs an agent loop (`chatWithManagementTools`) with skills, session history, and harvested `agent_lesson` rows, but every turn still uses a single shared `DORPH_AI_*` client. We need a stronger primary reasoning model (NVIDIA Nemotron on the integrate API, with optional thinking streams) while keeping a fast, local-friendly coordinator for routing, memory maintenance, and sub-agent fan-out—without replacing the existing tool catalog or skills store.

## What Changes

- Add a **primary orchestrator** profile (NVIDIA integrate OpenAI-compatible API) used for the main Dorph AI agent loop: tool JSON, user-facing replies, and optional streamed `reasoning_content` + answer content.
- Add a **coordinator** profile (existing SiliconFlow provider) used for lightweight sub-calls: pre-turn plan, memory distill/update, sub-agent worker briefs, and lesson harvest—never for direct user-facing final answers when the primary is configured.
- Extend **`pkg/dorphai`** (Go) with an NVIDIA provider id, `enable_thinking` / `chat_template_kwargs` on compatible chat requests, and optional streaming deltas for reasoning vs content.
- Wire **Dorph API** (`dorph/handlers`) so `POST /api/chat` resolves two clients from env, runs coordinator → primary → tools in a defined order, and keeps **skills** + **agent_lesson** injection in the primary system prompt.
- Add **working memory** snippets (coordinator-written, session-scoped) alongside durable lessons; cap injected tokens.
- Document repo-root env keys (`DORPH_AI_ORCHESTRATOR_*`, `DORPH_AI_COORDINATOR_*` or mapped provider ids); **never commit API keys**.
- **Non-breaking default:** if orchestrator env is unset, behavior stays on today’s single `LoadFromEnv()` client.

## Capabilities

### New Capabilities

- `dorphai-nemotron-orchestrator`: Primary NVIDIA Nemotron model configuration, chat completion options (thinking, streaming), and ownership of the management-tool agent loop.
- `dorphai-siliconflow-coordinator`: SiliconFlow-backed coordinator role for planning, sub-agent coordination, and memory-oriented JSON/text tasks without replacing the primary reply path.
- `dorphai-agent-memory-skills`: How session working memory, harvested lessons, and selected/catalog skills compose into each agent turn.

### Modified Capabilities

- (none — no existing `openspec/specs/` baseline defines dual-model agent roles)

## Impact

- **`pkg/dorphai`**: provider registry, resolve, client streaming, env loading for role-specific configs.
- **`dorph`**: `management_chat.go`, `chat.go`, `agent_lesson.go`, `agent_context.go`, `skills.go`; possible SQLite table for `agent_working_memory`.
- **`.env.example`**: documented orchestrator/coordinator variables (placeholders only).
- **`pkg/dorphai-rs`**: follow-up task if Project/dorph-engi should share the same split (out of scope for first slice unless noted in tasks).
- **Frontend**: optional SSE/streaming UI for thinking blocks on Dorph AI (phased in tasks).
- **Related in-flight changes:** complements `dorphai-agent-workspace` orchestration UI and `dorphai-agent-research-design-harness` lesson harvest—this change defines **model roles**, not the workspace shell.
