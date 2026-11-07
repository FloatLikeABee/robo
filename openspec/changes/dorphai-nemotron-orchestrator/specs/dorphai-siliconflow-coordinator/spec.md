# Spec Delta

## Purpose

Defines the SiliconFlow coordinator sub-model that plans turns, maintains session working memory, and supports sub-agent coordination without replacing the primary reply path.

## ADDED Requirements

### Requirement: Coordinator resolves from SiliconFlow env

When `DORPH_AI_COORDINATOR_PROVIDER` is `siliconflow`, the coordinator SHALL use `SILICONFLOW_API_KEY`, `SILICONFLOW_BASE_URL`, and `SILICONFLOW_MODEL` (when set). It SHALL NOT authorize SiliconFlow from `OPENAI_API_KEY` or legacy `DORPH_AI_API_KEY` alone.

#### Scenario: Coordinator model from SILICONFLOW_MODEL

- **WHEN** `DORPH_AI_COORDINATOR_PROVIDER=siliconflow`, `SILICONFLOW_API_KEY` is set, and `SILICONFLOW_MODEL=Qwen/Qwen3-8B`
- **THEN** coordinator completions use model `Qwen/Qwen3-8B` at the SiliconFlow base URL

#### Scenario: Coordinator disabled

- **WHEN** `DORPH_AI_COORDINATOR_PROVIDER` is empty
- **THEN** pre-turn planning and coordinator memory patches are skipped without failing the chat request

### Requirement: Coordinator output is structured side work

Before the orchestrator tool loop, the coordinator MAY return a JSON object with optional fields `plan`, `memory_patch`, and `sub_agents`. The coordinator SHALL NOT execute management tools or return the final user-facing answer when an orchestrator is configured.

#### Scenario: Plan only

- **WHEN** coordinator returns `{"plan":"Check generic-data then summarize"}`
- **THEN** that plan text is appended to the orchestrator system context and no tool HTTP calls were made by the coordinator

#### Scenario: Sub-agent list bounded

- **WHEN** coordinator returns `sub_agents` with more than three ids
- **THEN** the server truncates to three recognized worker ids before prefetch

### Requirement: Simple chat skips coordinator

When the user message qualifies as simple chat under existing server rules, the system SHALL skip the coordinator call.

#### Scenario: Greeting skips coordinator

- **WHEN** the user sends `hi` and simple-chat detection applies
- **THEN** no coordinator HTTP request is made

### Requirement: Lesson distill prefers coordinator

When a significant session is harvested into `agent_lesson`, the distill call SHALL use the coordinator client if configured; otherwise it SHALL use the orchestrator or legacy client.

#### Scenario: Harvest uses coordinator

- **WHEN** session significance thresholds are met and the coordinator is configured
- **THEN** the distill prompt is sent to the SiliconFlow coordinator endpoint
