## Why

Outside agents reach Dorph through dorph-mcp. That bridge can create, list, and get Notes & TODOs only. Stick notes, timelines, stories, research, and event logs are still unreachable unless the agent clicks the UI or posts HTTP on its own.

## What Changes

- dorph-mcp gains create, list, and get tools for stick notes, timelines, stories, research, and event logs, acting as the verified session user with no user-id argument.
- Stick notes, timelines, stories, and research are stored in the same SQLite file as notes. The text is saved as written. Those creates do not start the in-app generators.
- Event logs stay in Event Logs. The bridge calls the existing Dorph events API with the session token and does not open Badger.
- A skill tells an outside agent which tool to call. Notes & TODOs stay on `create_note`, `list_my_tasks`, and `get_task`.
- Check is list or get. The tools do not delete, publish, or email.

## Capabilities

### New Capabilities

- `mcp-module-bridge`: Outside agents create, list, and get stick notes, timelines, stories, research, and event logs through dorph-mcp.

### Modified Capabilities

## Impact

- `dorph/mcp` tool list and SQLite writes. Event log tools call `POST/GET /api/sheetx/events-info` on the Dorph API.
- New skill `.cursor/skills/dorph-module-records/SKILL.md`.
- Existing notes tools and the dorph-notes skill stay as they are.
