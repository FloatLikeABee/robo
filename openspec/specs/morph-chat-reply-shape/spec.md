# morph-chat-reply-shape Specification

## Purpose

Keeps Morph assistant replies matched to the question, concise enough for a small local model, and honest about which model is configured.

## Requirements

### Requirement: Plain questions get plain answers
Morph AI chat and note or TODO AI assist MUST answer the operator's question in concise plain language. A diagram, chart, table, or code block MUST be included only when the question is about structure, quantities, or code and that form makes the answer clearer. An ordinary question MUST NOT be answered as a caption plus a diagram, a pixel grid, or a code or architecture write-up.

#### Scenario: Identity question
- **WHEN** the operator asks which model is in use
- **THEN** the reply is short prose
- **AND** it does not include a diagram, a caption label, or a JSON metadata block

#### Scenario: Structure question
- **WHEN** the operator asks how a process or architecture is structured
- **THEN** the reply MAY include one diagram or code block that explains that structure

### Requirement: Model identity is not invented
When the operator asks Morph AI chat which model or version is in use, the reply MUST name the configured provider and model. If the provider or model is not configured, the reply MUST say that part is unknown. The reply MUST NOT invent a product version, model name, or update date. Note and TODO AI assist does not answer that question; it only follows the plain-answer rule.

#### Scenario: Configured provider and model
- **WHEN** Morph is configured with a provider and a model and the operator asks which model is in use
- **THEN** the reply names that provider and that model
- **AND** it does not add a version or date that was not configured

#### Scenario: Unknown configuration
- **WHEN** no provider and no model can be resolved and the operator asks which model is in use
- **THEN** the reply says the model is unknown

### Requirement: Diagram skill is opt-in
The default chat prompt MUST NOT require a diagram on every reply. A graphs skill that asks for a diagram MUST apply only when the operator selects that skill.

#### Scenario: No graphs skill selected
- **WHEN** the operator sends a chat message without selecting the graphs skill
- **THEN** the prompt does not instruct the model to lead with a diagram
