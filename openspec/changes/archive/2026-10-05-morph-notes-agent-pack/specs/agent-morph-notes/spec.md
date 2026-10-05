# Spec Delta

## Purpose

Let an AI agent create, list, and get the signed-in user's Morph notes by following repo skills, slash commands, and a self-use instruction that call the existing morph-mcp tools.

## ADDED Requirements

### Requirement: Skill names the real note tools
The Morph notes skill MUST name `create_note`, `list_my_tasks`, and `get_task`. It MUST say when to create a note, when to list, and when to get by id. It MUST document `title` and `body` for create, `type`, `status`, and `limit` for list, and `id` for get.

#### Scenario: Agent loads the skill
- **WHEN** an agent reads the Morph notes skill
- **THEN** the text names `create_note`, `list_my_tasks`, and `get_task`
- **AND** it says title or body is required to create
- **AND** it says list accepts type, status, and limit
- **AND** it says get accepts an integer id

#### Scenario: List cap is stated
- **WHEN** an agent reads how to list notes
- **THEN** the skill says the default limit is 50 and the cap is 100
- **AND** it says a note missing from that page is not proof the id does not exist

### Requirement: Skill forbids click-paths and the HTTP write
The skill MUST identify `http://localhost:3031/` Notes & TODOs as the human review surface. It MUST tell the agent to create, list, and get by calling the morph-mcp tools. It MUST tell the agent not to POST `/api/tran/notes-todos` and not to click that UI to finish the note.

#### Scenario: Create does not use the browser
- **WHEN** an agent reads the skill in order to leave a note
- **THEN** the write instruction is to call `create_note`
- **AND** the skill says not to POST `/api/tran/notes-todos`
- **AND** the skill says not to click Notes & TODOs to create the note

### Requirement: Slash commands map to the same tools
The repository MUST provide a create command, a list command, and a get command under `.cursor/commands/`. The create command MUST tell the agent to call `create_note`. The list command MUST tell the agent to call `list_my_tasks`. The get command MUST tell the agent to call `get_task`. A command MUST NOT name a different note API as the way to finish that flow.

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

### Requirement: Self-use instruction stands alone
The self-use instruction MUST be a document another bot can follow to create, list, and get a note without a human click-path. It MUST name `create_note`, `list_my_tasks`, and `get_task` and the same arguments as the skill. If those tools are absent, it MUST tell the bot to report that morph-mcp is not connected.

#### Scenario: Another bot leaves a note
- **WHEN** a bot is given only the self-use instruction and asked to leave a note
- **THEN** the steps call `create_note` with title, body, or both
- **AND** the steps do not require a human to click Notes & TODOs

#### Scenario: Tools are missing
- **WHEN** a bot following the instruction cannot call the morph-mcp tools
- **THEN** the instruction says to report that morph-mcp is not connected
- **AND** it does not substitute a click-path or `POST /api/tran/notes-todos`

### Requirement: Pack states attribution and session scope
The skill and the self-use instruction MUST say the stored title starts with `[morph-mcp]` and the stored body starts with `source: morph-mcp`. They MUST say the tools act as the verified session user, that the calls take no user-id argument, and that another user's id is not found. They MUST say a create error that the session is not a notes user stops the agent, with no HTTP retry.

#### Scenario: Agent expects the marker
- **WHEN** an agent reads the create steps in the skill or the instruction
- **THEN** it is told the stored title starts with `[morph-mcp]`
- **AND** it is told the stored body starts with `source: morph-mcp`

#### Scenario: No cross-user read
- **WHEN** an agent reads how get behaves
- **THEN** the text says an id that is missing or belongs to someone else is not found

#### Scenario: No Tran user is not user id 1
- **WHEN** an agent reads a create error that the session is not a notes user
- **THEN** the skill and the instruction say to stop
- **AND** they say not to retry through `POST /api/tran/notes-todos`

### Requirement: Pack does not collect secrets
The skill, the three commands, and the self-use instruction MUST NOT contain a JWT or a JWT secret. They MUST tell the agent not to put `MORPH_MCP_TOKEN` or `JWT_SECRET` into a note, a commit, or the chat.

#### Scenario: Secrets stay out of the note
- **WHEN** an agent reads how to create a note
- **THEN** the skill and the instruction say not to write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the note
- **AND** the pack files do not contain a bearer token

### Requirement: Pack stays on local Morph
The skill and the self-use instruction MUST name `http://localhost:3031/` as the local human review origin. They MUST NOT instruct the agent to change Render or a deploy configuration.

#### Scenario: Local review origin
- **WHEN** an agent reads where a human reviews the note
- **THEN** the text names `http://localhost:3031/`
- **AND** it does not tell the agent to edit Render or deploy config
