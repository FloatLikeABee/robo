## Purpose

Makes Morph AI default as an agent with always-on Research and Design skills that analyse each user prompt against available documents before using tools or answering.

## ADDED Requirements

### Requirement: Default Research and Design skills

Morph AI MUST ensure builtin Research and Design skills exist (insert-if-missing). Every management-tool chat MUST include those skills’ full instruction bodies without the operator selecting them in the picker.

#### Scenario: Fresh or existing skill store

- **WHEN** Morph AI starts and Research or Design builtins are missing
- **THEN** those skills are inserted enabled
- **AND** existing operator-created skills are left in place

#### Scenario: Chat gets full default bodies

- **WHEN** the operator sends a Morph AI chat message with no skill picker selection
- **THEN** the agent context still includes the full Research and Design instruction bodies

### Requirement: Analyse prompt and document materials

For each Morph AI chat request, the agent MUST analyse the user prompt against available session documents (pinned files, notes, knowledge) before calling tools or giving the final answer. The analysis MUST identify the goal, which materials apply, and unknowns.

#### Scenario: Document-backed question

- **WHEN** the operator asks about content that is in pinned files or knowledge
- **THEN** the agent uses those materials (search or excerpts) rather than answering from the model alone

#### Scenario: Ambiguous prompt

- **WHEN** the operator prompt is underspecified
- **THEN** the agent states the unknowns (briefly) and either asks one clarifying question or proceeds with an explicit assumption
