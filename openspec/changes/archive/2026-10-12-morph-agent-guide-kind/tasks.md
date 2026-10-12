# Tasks

## 1. Kind classifier

- [x] 1.1 Add `SuggestNoteKind` in a new `morph/mcp/agent_runtime.go` with the ordered rules from design.md decision 2. Verify `go test ./mcp -count=1 -run TestSuggestNoteKind` from `morph/` fails because the table file is missing.
- [x] 1.2 Add `agent_runtime_test.go` with about 30 table rows covering note, todo, task, timeline, event_log, stick_note, research, big_note, grill-on, and grill-off. Verify every row matches expected `kind` and `use_grill`, empty intent is `invalid_input`, and the test file does not import an AI client.
- [x] 1.3 Export `AgentRuntimeToolCount = 2`. Verify the constant is 2 and `records_tools.go` still has `ToolCount = 20`.

## 2. MCP tools, resource, and initialize

- [x] 2.1 Register `suggest_note_kind` and `morph_guide` as read-only tools from `registerAgentRuntime`, and add `morph://guide`. Call that registrar from `NewServerWith`. Verify `go test ./mcp -count=1 -run 'TestHandshakeListAndWhoami|TestToolsAreWhoamiListAndGet|TestNotesAppSurface'` fails on the old tool/resource counts.
- [x] 2.2 Update handshake, notes-app, and stdio tests to expect `mcp.ToolCount + mcp.AgentRuntimeToolCount` tools, the two new read-only names, and `morph://guide` alongside the two UI resources. Verify those tests pass after the count change.
- [x] 2.3 Rewrite the `instructions` const to under 2000 characters, keep `read-only`, name `morph://guide` or `morph_guide`, and state never-share-without-consent. Verify `len(instructions) < 2000` in a test and `TestHandshakeListAndWhoami` still passes.

## 3. Playbook and morph_guide

- [x] 3.1 Write `docs/agents/16-morph-for-agents.md` with every section in the plan outline (setup/PAT, kinds, writing, grill, checking, sharing, pool, errors, per-client, privacy, tool reference). Verify it names all phase 1B pool tools, the five error codes, CC BY 4.0 alias attribution, and untrusted shared content.
- [x] 3.2 Copy that file to `morph/mcp/agent_guide.md` and embed it. Add a test that the two files are byte-identical and `resources/read` of `morph://guide` returns that markdown. Verify `go test ./mcp -count=1 -run 'TestGuide|TestMorphGuide'` passes.
- [x] 3.3 Implement `morph_guide` topic excerpts (`kinds`, `sharing`, `grill`, `limits`, `setup`, `checking`, `pool`, `errors`, `tools`, `writing`, `privacy`). Empty topic returns outline plus golden rules. Unknown topic is `invalid_input` with the valid list. Verify those cases in `TestMorphGuide`.
- [x] 3.4 Point `docs/agents/14-morph-mcp.md` and `docs/agents/00-architecture-overview.md` at chapter 16. Verify both files name `16-morph-for-agents.md`.

## 4. Skills and commands pack

- [x] 4.1 Update `.cursor/skills/morph-notes/SKILL.md` (still under 150 lines) to name `suggest_note_kind`, `morph_guide`, chapter 16, and never-share-without-consent, while keeping the #131 locked phrases. Verify `TestAgentPack` still passes the old assertions.
- [x] 4.2 Add `.cursor/commands/morph-todo.md`, `morph-log-session.md`, `morph-review.md`, `morph-share.md`, and `morph-pool-find.md`, each under 150 lines, each pointing at chapter 16. `/morph-share` MUST require an explicit human yes. Verify each file exists and `wc -l` is under 150.
- [x] 4.3 Add `agent-pack/morph/CLAUDE.md` and `agent-pack/morph/cursor-rule.mdc` as portable snippets that link to chapter 16 and the Cursor skill. Verify they contain `suggest_note_kind` and no `eyJ`.
- [x] 4.4 Extend `agentpack_test.go` to scan `agent_runtime.go` for the new tool names and to require the new commands plus the consent line. Keep the existing create/list/get locks. Verify `go test ./mcp -count=1 -run TestAgentPack` passes.
- [x] 4.5 Add a one-line pointer in `docs/agents/15-morph-notes-self-use.md` to chapter 16 without dropping any phrase `TestAgentPack` already locks. Verify that test still passes.

## 5. Integration

- [x] 5.1 From `morph/`, run `go vet ./mcp ./cmd/morph-mcp` and `go test ./mcp ./cmd/morph-mcp -count=1`. Verify both succeed.

## Workflow follow-up

- Archive the change after the Morph API tests are green.
- Open a PR to main with `Refs #144`. Do not merge.
