# Spec Delta

## MODIFIED Requirements

### Requirement: Slash commands map to the same tools
The repository MUST provide a create command, a list command, and a get command under `.cursor/commands/`. The create command MUST tell the agent to call `create_note`. The list command MUST tell the agent to call `list_my_tasks`. The get command MUST tell the agent to call `get_task`. A command MUST NOT name a different note API as the way to finish that flow. The pack MUST also provide commands that map todo, session-log, review, share, and pool-find onto the Morph tools named in those command files.

#### Scenario: Create command
- **WHEN** the create-note slash command is read
- **THEN** it tells the agent to call `create_note`
- **AND** it does not tell the agent to POST `/api/tran/notes-todos`

#### Scenario: List command
- **WHEN** the list-notes slash command is read
- **THEN** it tells the agent to call `list_my_tasks`

#### Scenario: Get command
- **WHEN** the get-note slash command is read
- **THEN** it tells the agent to call `get_task` with the id argument

#### Scenario: Missing tool is not a click-path
- **WHEN** a command is read and the morph-mcp tools are not available
- **THEN** the command tells the agent to report that morph-mcp is not connected
- **AND** it does not tell the agent to finish by clicking the Notes & TODOs UI

#### Scenario: New flow commands exist
- **WHEN** the commands directory is read
- **THEN** it includes morph-todo, morph-log-session, morph-review, morph-share, and morph-pool-find

## ADDED Requirements

### Requirement: Skill names kind suggestion and the guide
The Morph notes skill MUST name `suggest_note_kind` and `morph_guide`. It MUST say to suggest a kind when the ask is not obviously one kind, to grill when `use_grill` is true, and to read `docs/agents/16-morph-for-agents.md` for the full contract.

#### Scenario: Agent loads the skill for an unclear ask
- **WHEN** an agent reads the Morph notes skill
- **THEN** the text names `suggest_note_kind` and `morph_guide`
- **AND** it points at `docs/agents/16-morph-for-agents.md`
