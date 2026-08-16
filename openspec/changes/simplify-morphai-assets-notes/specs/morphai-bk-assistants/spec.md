## Purpose

Defines that the Morph AI Assistants sidebar lists only assistants created in AI tools (BK), not Morph built-in specialists.

## ADDED Requirements

### Requirement: Assistants are AI tools only
The Morph AI Assistants picker SHALL list only assistants that originate from AI tools (BK). Morph system-defined or Morph-store specialist agents MUST NOT appear in that list.

#### Scenario: Assistants list populated from AI tools
- **WHEN** a user expands Assistants in the Morph AI sidebar
- **THEN** every listed assistant is an AI tools assistant
- **AND** Morph built-in agents such as general/specialist Morph agents are not listed

#### Scenario: Default selection when AI tools assistants exist
- **WHEN** at least one AI tools assistant is available
- **THEN** the selected assistant is one of those AI tools assistants (or a previously stored AI tools assistant id that still exists)

#### Scenario: No AI tools assistants available
- **WHEN** AI tools returns no assistants
- **THEN** the Assistants list is empty or shows an empty state
- **AND** Morph built-ins are still not offered as substitutes in that list
