# Proposal

## Why

External agents connecting to Morph over MCP still get operator docs and a three-tool self-use pack. They have no rules-first way to pick a note kind, no playbook for PAT setup, grill, checking notes, or the shared pool, and `initialize` does not point them at one. Epic #144 (Echo) needs that surface now so Bridge's unified notes and pool tools have a documented client contract.

## What Changes

- Add a read-only `suggest_note_kind` MCP tool: rules-first and deterministic (no LLM). Input is the user's ask (`intent`) plus optional `content_hint`. Output is `kind` (`note`, `todo`, `big_note`, `timeline`, `research`, `stick_note`, `task`, and `event_log` when that is the match), `reason`, `use_grill`, and `confidence`.
- Add `docs/agents/16-morph-for-agents.md` as the canonical playbook for any external agent. Serve it as `morph://guide` and through a read-only `morph_guide` tool. Keep `initialize` instructions under 2k characters and point at that guide.
- Document the target tool names from the plan (unified notes, grill, and pool). Mark pool tools as coming in phase 1B when they are not implemented yet. Do not implement Bridge's unified notes, PAT, or pool tools in this change.
- Update the #131 skills/commands pack for kind selection, grill, review, share (consent first), and pool-find. Add the commands listed in the plan. Each pack file stays under 150 lines and links to the canonical doc.
- Register the two new tools in a new `morph/mcp` file so Bridge's module-tool file stays untouched.

## Capabilities

### New Capabilities

- `suggest-note-kind`: Deterministic rules that map an ask to a Morph note kind, a reason, a grill flag, and a confidence score.
- `morph-for-agents`: The external-agent playbook (kinds, every tool with examples and error codes, grill, checking, pool consent and attribution, setup, etiquette) served as `morph://guide` and `morph_guide`, plus the short initialize pointer and the skills/commands pack for those flows.

### Modified Capabilities

- `morph-mcp-stdio`: `tools/list` also advertises `suggest_note_kind` and `morph_guide` (both read-only). `resources/list` also includes `morph://guide`. Initialize instructions stay under 2k characters and point at the guide.
- `agent-morph-notes`: The pack still drives create/list/get through the existing tools, and now also names kind suggestion, the guide, and the new slash commands. It still forbids the HTTP notes write and click-path.

## Impact

- New file under `morph/mcp/` for the two tools and the guide resource; a one-call hook and a shorter `instructions` string in `server.go`.
- Handshake and pack tests that count tools or resources.
- `docs/agents/16-morph-for-agents.md`, an embedded copy used by the server, pointers in `14-morph-mcp.md` and `00-architecture-overview.md`.
- `.cursor/skills/morph-notes`, `.cursor/commands/` (existing plus `/morph-todo`, `/morph-log-session`, `/morph-review`, `/morph-share`, `/morph-pool-find`), and `agent-pack/morph/` (CLAUDE snippet and Cursor rule).
- No change to `records_tools.go`, grill, identity, PAT, unified notes, pool schema, frontend, Render, or deploy config.
