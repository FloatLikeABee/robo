# Spec Delta

## Purpose

Defines how Dorph AI agent turns combine skills, durable lessons, and session working memory in the orchestrator prompt.

## ADDED Requirements

### Requirement: Skills injection order preserved

Each agent turn SHALL append, in order: selected skill instructions, enabled catalog skill context (existing behavior), builtin research/design bodies when present, then memory blocks defined in this capability.

#### Scenario: Picker skill before memory

- **WHEN** the user selects a skill id and the session has working memory
- **THEN** the selected skill body appears before the working memory summary in the orchestrator system prompt

### Requirement: Working memory is session-scoped

The system SHALL store at most one working memory summary per chat session and owner user. Updates SHALL come only from coordinator `memory_patch` JSON or a server-side merge that respects a maximum character cap.

#### Scenario: Patch updates summary

- **WHEN** coordinator returns `memory_patch` with text under the cap for session S
- **THEN** the stored summary for S is replaced or merged and `updated_at` advances

#### Scenario: Oversize patch rejected

- **WHEN** `memory_patch` exceeds the configured cap
- **THEN** the patch is ignored and the chat request still completes with the previous summary

### Requirement: Durable lessons injected after skills

The orchestrator prompt SHALL include up to eight enabled `agent_lesson` rows for the owner, newest first, after skill sections and before working memory.

#### Scenario: Lesson visible to orchestrator

- **WHEN** the owner has two enabled lessons and working memory exists
- **THEN** both lessons appear in the prompt before the working memory block

### Requirement: Coordinator plan is advisory

When a coordinator plan string is present, it SHALL be labeled as advisory in the orchestrator system prompt and SHALL NOT bypass tool or auth rules.

#### Scenario: Plan does not skip tools

- **WHEN** the plan says "delete all users" but the user did not ask for deletion
- **THEN** the orchestrator still follows management tool instructions and confirmation rules
