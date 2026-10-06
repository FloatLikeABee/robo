## Purpose

Lets an agent outside Morph create, list, and get stick notes, timelines, stories, research, and event logs through morph-mcp, the same way it already handles Notes & TODOs.

## ADDED Requirements

### Requirement: Bridge tools are named and session-scoped
morph-mcp MUST advertise create, list, and get tools for stick notes, timelines, stories, research, and event logs. The tools MUST act as the verified session user and MUST NOT take a user-id argument. Stored titles for the SQLite modules MUST start with `[morph-mcp]`.

#### Scenario: Server lists the stick-note tools
- **WHEN** a client lists morph-mcp tools
- **THEN** the list includes `create_stick_note`, `list_stick_notes`, and `get_stick_note`

#### Scenario: Server lists the other module tools
- **WHEN** a client lists morph-mcp tools
- **THEN** the list includes `create_timeline`, `list_timelines`, `get_timeline`, `create_story`, `list_stories`, `get_story`, `create_research`, `list_research`, `get_research`, `create_event_log`, `list_event_logs`, and `get_event_log`

### Requirement: SQLite modules store the agent's text
`create_stick_note` MUST insert a stick note. `create_timeline` MUST insert a timeline for the session user. `create_story` MUST insert a story for the session user. `create_research` MUST insert a research row for the session user. These creates MUST NOT start the in-app generators. List and get MUST return only that user's timelines, stories, and research. Stick notes are the shared board, so list and get are not limited to one owner. An id that is missing, or that is another user's timeline, story, or research, MUST be not found.

#### Scenario: Create a stick note
- **WHEN** an agent calls `create_stick_note` with a title
- **THEN** a stick note is stored whose title starts with `[morph-mcp]`
- **AND** no Notes & TODOs row is created for that call

#### Scenario: Another user's story is not found
- **WHEN** an agent gets a story id that belongs to someone else
- **THEN** the tool result is not found

### Requirement: Event logs use the events API
`create_event_log`, `list_event_logs`, and `get_event_log` MUST use the Morph events API. They MUST NOT open Badger and MUST NOT write a Notes & TODOs row. A missing token or a failed API call MUST be reported as an error.

#### Scenario: Create an event log
- **WHEN** an agent calls `create_event_log` with a title and a time and the events API accepts it
- **THEN** the tool returns the API result
- **AND** no Notes & TODOs row is created

### Requirement: Skill points at the tools
A skill MUST name these tools and say to call them instead of posting the module HTTP APIs or clicking MorphNotes. It MUST say Notes & TODOs still use `create_note`, `list_my_tasks`, and `get_task`. If the tools are missing, it MUST say morph-mcp is not connected.

#### Scenario: Agent reads the skill
- **WHEN** an agent reads the module-records skill
- **THEN** it names `create_stick_note`, `create_timeline`, `create_story`, `create_research`, and `create_event_log`
- **AND** it says not to click MorphNotes to create the record
