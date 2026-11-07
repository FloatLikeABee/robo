# Spec Delta

## Purpose

Defines the primary NVIDIA Nemotron orchestrator used for Dorph AI’s tool-based agent loop and user-facing replies.

## ADDED Requirements

### Requirement: Orchestrator resolves from dedicated env

When `DORPH_AI_ORCHESTRATOR_PROVIDER` is set to a known provider id (including `nvidia`), the Dorph AI agent loop SHALL resolve API key, base URL, and model from that provider’s env vars only. It SHALL NOT reuse another provider’s key. When unset, the orchestrator SHALL use the legacy single `LoadFromEnv()` configuration.

#### Scenario: NVIDIA orchestrator configured

- **WHEN** `DORPH_AI_ORCHESTRATOR_PROVIDER=nvidia`, `NVIDIA_API_KEY` is set, and `NVIDIA_MODEL` is unset
- **THEN** chat completions use base URL `https://integrate.api.nvidia.com/v1` and model `nvidia/nemotron-3-ultra-550b-a55b`

#### Scenario: Legacy fallback

- **WHEN** `DORPH_AI_ORCHESTRATOR_PROVIDER` is empty and `DORPH_AI_API_KEY` is set
- **THEN** the agent loop behaves as today with one client from `LoadFromEnv()`

### Requirement: Orchestrator owns management tools

The orchestrator client SHALL be the only model that emits JSON tool calls against the Dorph management tool catalog and SHALL produce the final user-visible prose reply for `POST /api/chat` when configured.

#### Scenario: Tool round uses orchestrator

- **WHEN** the user asks for a Tran list operation and the orchestrator is configured
- **THEN** JSON tool requests are sent to the orchestrator model endpoint, not the coordinator endpoint

### Requirement: Thinking mode on compatible chat

When the orchestrator uses OpenAI-compatible chat completions, the system SHALL send `chat_template_kwargs.enable_thinking=true` in the request body for NVIDIA integrate unless an env flag explicitly disables it.

#### Scenario: Nemotron thinking enabled

- **WHEN** the orchestrator provider is `nvidia` and a non-streaming completion is requested
- **THEN** the outbound JSON body includes `chat_template_kwargs` with `enable_thinking` true

### Requirement: Optional streamed reasoning

When the client requests streaming completions for the orchestrator, the system SHALL surface separate reasoning and answer text deltas when the provider returns them.

#### Scenario: Stream carries reasoning delta

- **WHEN** streaming is enabled and the provider returns `reasoning_content` on a chunk
- **THEN** the stream consumer can read reasoning separately from `content`
