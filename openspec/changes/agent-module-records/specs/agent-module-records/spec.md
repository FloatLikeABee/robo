## Purpose

Lets a Morph AI agent create, check, and get stick notes, timelines, stories, research, and event logs in the modules the operator already uses.

## ADDED Requirements

### Requirement: Five modules stay on their own records
When the operator asks the agent to create, check, or get a stick note, timeline, story, research item, or event log, the agent MUST use that module’s existing API. It MUST NOT create a Notes & TODOs row for that request. A request for a personal note or TODO MUST still use Notes & TODOs.

#### Scenario: Stick note is not a personal note
- **WHEN** the operator asks the agent to create a stick note
- **THEN** the agent calls `POST /api/tran/case-tasks` and does not call `POST /api/tran/notes-todos`

#### Scenario: Personal todo stays a todo
- **WHEN** the operator asks the agent to add a personal TODO
- **THEN** the agent uses Notes & TODOs

### Requirement: Each module can be listed, read, and created
The agent instructions MUST name these routes:

| Module | List | Get | Create |
| --- | --- | --- | --- |
| Stick notes | `GET /api/tran/case-tasks` | `GET /api/tran/case-tasks/:id/full` | `POST /api/tran/case-tasks` |
| Timelines | `GET /api/tran/timelines` | `GET /api/tran/timelines/:id` | `POST /api/tran/timelines` |
| Stories | `GET /api/tran/big-notes` | `GET /api/tran/big-notes/:id` | `POST /api/tran/big-notes` |
| Research | `GET /api/tran/research` | `GET /api/tran/research/:id` | `POST /api/tran/research` |
| Event logs | `GET /api/sheetx/events-info` | `GET /api/sheetx/events-info/:id` | `POST /api/sheetx/events-info` |

Create bodies MUST follow the existing APIs: a stick note needs `title`, `start_at`, and `end_at`; a timeline needs `title` plus `paste`, `content`, or `url`; a story needs `idea`; research needs `prompt`; an event log needs `title` and `time`.

#### Scenario: Research is in the tool instructions
- **WHEN** the agent instructions are loaded
- **THEN** they include `GET /api/tran/research` and `POST /api/tran/research`

#### Scenario: Timeline is in the tool instructions
- **WHEN** the agent instructions are loaded
- **THEN** they include `GET /api/tran/timelines` and `POST /api/tran/timelines`

### Requirement: Check reports the record
Check MUST list or get the module and tell the operator whether a matching record exists, including its main fields. Check MUST NOT add a new status field.

#### Scenario: Missing stick note
- **WHEN** the operator asks the agent to check a stick note that is not in the list
- **THEN** the agent says it was not found and does not create a Notes & TODOs row

### Requirement: Failed creates are reported
If a create call fails, the agent MUST report that error. It MUST NOT store the same request as a Notes & TODOs row.

#### Scenario: Event log create fails
- **WHEN** `POST /api/sheetx/events-info` returns an error
- **THEN** the agent tells the operator the error and does not call `POST /api/tran/notes-todos`

### Requirement: Extra actions stay opt-in
The agent MUST NOT delete, publish, or email a stick note, timeline, story, research item, or event log unless the operator asks for that action. Story, timeline, and research creates MUST keep the generator those endpoints already run.

#### Scenario: Create a story
- **WHEN** the operator asks the agent to create a story from an idea
- **THEN** the agent calls `POST /api/tran/big-notes` with that idea and does not publish it

#### Scenario: Create does not email
- **WHEN** the operator asks the agent to create a stick note
- **THEN** the agent does not call the case-task send-email route
