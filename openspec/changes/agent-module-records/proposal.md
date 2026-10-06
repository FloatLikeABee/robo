## Why

Morph AI can already call the admin APIs, but when someone asks an agent for a note it files a personal Notes & TODOs row. Stick notes, timelines, stories, research, and event logs are separate records, and timelines and research are not even named in the tool instructions.

## What Changes

- When the operator asks an agent to create, check, or get a stick note, timeline, story, research item, or event log, the agent uses that module’s existing list, get, and create routes.
- Check means list or get and say whether a matching record exists, with its main fields. It does not add a new status.
- A personal note or TODO still uses Notes & TODOs. These five modules are not saved there instead.
- Stories, timelines, and research keep their current create behavior, including generation. The agent does not add a second plain-text store.
- The agent does not delete, publish, or email these records unless the operator asks for that.

Out of scope: new MCP tools, Info Sheets, a language switch, and generic data.

## Capabilities

### New Capabilities

- `agent-module-records`: Morph AI agents create, check, and get stick notes, timelines, stories, research, and event logs through the existing module APIs.

### Modified Capabilities

## Impact

- Morph AI management tool instructions and catalog in `morph/handlers/management_chat.go`.
- Existing routes stay: `/api/tran/case-tasks`, `/api/tran/timelines`, `/api/tran/big-notes`, `/api/tran/research`, `/api/sheetx/events-info`.
- No new tables and no change to the morph-mcp Notes & TODOs tools.
