# Proposal

## Why

`morph-mcp` can already create, list, and get the signed-in user's Notes & TODOs, but an agent that is only told "use Morph" still has no skill, slash command, or instruction that names those tools. Without that pack, the agent either waits for a human to click Notes & TODOs at `http://localhost:3031/` or invents an HTTP write that the MCP design already rejected.

## What Changes

- Add a Cursor skill that says when to post or fetch a Morph note and how to call the existing `morph-mcp` tools.
- Add slash commands for the same create, list, and get flows. Each command is a prompt that calls those tools. There is no second API.
- Add a self-use instruction another bot can follow without a human click-path through the Notes & TODOs UI.
- Add one check that the pack names the real tools and does not teach the rejected HTTP write.

## Capabilities

### New Capabilities

- `agent-morph-notes`: An agent creates, lists, and gets the session user's Morph notes by following repo skills, slash commands, and a self-use instruction that call `create_note`, `list_my_tasks`, and `get_task`.

### Modified Capabilities

- None. `morph-mcp-stdio` already defines the tools. This change does not alter server behavior.

## Impact

- New files under `.cursor/skills/`, `.cursor/commands/`, and `docs/agents/`.
- A short pointer from the existing MCP agent doc so the instruction is findable.
- One Go test in `morph/mcp` that reads the pack and fails if it drifts from the tool names in `server.go`.
- No change to Render, deploy config, the Notes UI, or the MCP tool implementations.
