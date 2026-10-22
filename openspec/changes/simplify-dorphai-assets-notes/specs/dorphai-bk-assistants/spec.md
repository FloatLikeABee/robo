## Purpose

Defines that the Dorph AI Assistants sidebar lists only assistants created in AI tools (BK), not Dorph built-in specialists.

## ADDED Requirements

### Requirement: Assistants are AI tools only
The Dorph AI Assistants picker SHALL list only assistants that originate from AI tools (BK). Dorph system-defined or Dorph-store specialist agents MUST NOT appear in that list.

#### Scenario: Assistants list populated from AI tools
- **WHEN** a user expands Assistants in the Dorph AI sidebar
- **THEN** every listed assistant is an AI tools assistant
- **AND** Dorph built-in agents such as general/specialist Dorph agents are not listed

#### Scenario: Default selection when AI tools assistants exist
- **WHEN** at least one AI tools assistant is available
- **THEN** the selected assistant is one of those AI tools assistants (or a previously stored AI tools assistant id that still exists)

#### Scenario: No AI tools assistants available
- **WHEN** AI tools returns no assistants
- **THEN** the Assistants list is empty or shows an empty state
- **AND** Dorph built-ins are still not offered as substitutes in that list
