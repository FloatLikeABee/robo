# Tasks

## 1. Pack contract test

- [x] 1.1 Add `morph/mcp/agentpack_test.go` as specified in design.md decision 4. Run `go test ./mcp -count=1 -run TestAgentPack` from `morph/` and confirm it fails because the skill, the three commands, or the instruction doc is missing.

## 2. Skill, commands, and self-use instruction

- [x] 2.1 Add `.cursor/skills/morph-notes/SKILL.md` with the when/how contract from design.md decisions 1 and 3. Verify the file names `create_note`, `list_my_tasks`, and `get_task` and includes `Do not POST /api/tran/notes-todos`.
- [x] 2.2 Add `.cursor/commands/morph-note.md`, `morph-notes.md`, and `morph-note-get.md` that map `$ARGUMENTS` onto those tools. Verify each file names its tool, contains `Do not POST /api/tran/notes-todos`, and tells a missing-tool session to report that morph-mcp is not connected.
- [x] 2.3 Add `docs/agents/15-morph-notes-self-use.md` as a standalone procedure, plus a one-line pointer in `docs/agents/14-morph-mcp.md` and a Next bullet in `docs/agents/00-architecture-overview.md`. Verify the new chapter names the three tools and does not contain a login curl or a token.
- [x] 2.4 Re-run `go test ./mcp -count=1 -run TestAgentPack` from `morph/` and confirm it passes. Then run `go test ./mcp -count=1` from `morph/` and confirm the package passes.
