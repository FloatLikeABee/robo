## Purpose

Defines removal of Dorph AI task chains so the chat product no longer offers multi-step task chain management or scheduling.

## ADDED Requirements

### Requirement: Task chains feature removed
Dorph AI SHALL NOT expose Task chains entry points, modals, or background due-chain polling. Existing task-chain UI and client schedulers MUST be removed from the Dorph AI experience.

#### Scenario: No task chains control in Dorph AI
- **WHEN** a signed-in user opens Dorph AI chat
- **THEN** there is no control to open Task chains
- **AND** no task-chain modal or due-chain polling runs as part of the Dorph AI UI
