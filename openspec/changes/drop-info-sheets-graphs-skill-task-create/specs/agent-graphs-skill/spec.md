## Purpose

Makes mermaid diagrams and charts the default visual when an in-stack agent reply can show structure or quantities, including DorphUtils and DorphNotes AI.

## ADDED Requirements

### Requirement: Graphs when a reply can show them

In-stack AI replies MUST include a mermaid diagram or chart when the answer is structure, process, comparison, or quantities. Short greeting or a single labeled record MUST NOT invent a chart.

#### Scenario: Structure question in a DorphUtils module

- **WHEN** the operator asks Event Logs, Content Maker, Data Access, or Project AI about a process, comparison, or quantities
- **THEN** the assistant reply includes a mermaid diagram or chart

#### Scenario: Dorph AI same contract

- **WHEN** the operator asks Dorph AI a structure or quantity question
- **THEN** the reply includes a mermaid diagram or chart

### Requirement: DorphNotes AI can create graphs

DorphNotes AI (notes/todos assist, task generate, research compose) MUST be allowed to emit mermaid in generated markdown when a graph helps.

#### Scenario: Task generate includes a diagram

- **WHEN** the operator generates a DorphNotes task from a prompt that describes a process or quantities
- **THEN** the markdown outcome MAY include a mermaid fence
- **AND** the HTML preview draws that diagram

#### Scenario: Notes assist may use mermaid

- **WHEN** DorphNotes AI assist generates or improves a note whose topic is structure or quantities
- **THEN** the body MAY include mermaid instead of being forced to plain text without fences

### Requirement: Shared skill, not four copies of policy

The graph/chart instruction MUST come from one shared platform contract (existing visual-first text plus a Dorph AI graphs skill), not a unique essay per product.

#### Scenario: Skill listed on Dorph AI

- **WHEN** the operator opens Dorph AI skills
- **THEN** a Graphs (or equivalent) builtin skill is present and enabled by default
