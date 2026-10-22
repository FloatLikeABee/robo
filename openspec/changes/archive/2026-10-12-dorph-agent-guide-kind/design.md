# Design

## Context

See proposal.md for why. Code checked before this design:

- `dorph/mcp/server.go` registers whoami, list, get, create_note; `registerModuleTools` (20 total via `ToolCount` in `records_tools.go`); `registerGrillTools`. `instructions` is one long paragraph. Handshake tests require that string to contain `read-only` and `len(tools) == mcp.ToolCount`. Resource list tests require exactly `ui://dorph/notes` and `ui://dorph/grill`.
- `agentpack_test.go` extracts tool names from `server.go` only (`readOnly("` and `Name:`). The pack lives at `.cursor/skills/dorph-notes/SKILL.md`, three commands, and `docs/agents/15-dorph-notes-self-use.md`.
- `go:embed` cannot leave the package directory. The playbook canonical path is `docs/agents/16-dorph-for-agents.md`.
- Bridge owns PATs, unified `kind` notes, `search_my_notes`, owners, and pool tools on a parallel PR. This change must not edit `records_tools.go`, `records.go`, `tasks.go`, grill, or identity.
- Story #144 Echo: rules-first, no LLM, ~30-example table. Plan E2.4 mentioned an Ollama tie-break and 40 examples; the story wins.

## Goals / Non-Goals

**Goals:**

- Two read-only tools and `dorph://guide` registered from a new file.
- A deterministic classifier that the 30-row table pins.
- One thorough playbook, short initialize text, pack updates.

**Non-Goals:**

- Implementing pool, PAT, unified notes, or grill-for-all-kinds.
- An LLM fallback.
- The deferred 25-ask live-client eval (E3.3).
- Render, frontend, or deploy config.

## Decisions

### 1. New file `dorph/mcp/agent_runtime.go`; leave `ToolCount` at 20

Register `suggest_note_kind`, `dorph_guide`, and `dorph://guide` from `registerAgentRuntime(server, recheck)`. `server.go` gains that one call and a rewritten `instructions` const. `const AgentRuntimeToolCount = 2` lives next to the new tools. Handshake and stdio tests compare `len(tools)` to `mcp.ToolCount + mcp.AgentRuntimeToolCount`.

`ToolCount` stays in `records_tools.go`. Bumping it there is a merge collision with Bridge adding unified tools. Adding our increment beside our tools keeps both diffs additive.

Alternatives:

- Bump `ToolCount` to 22 in `records_tools.go`. Rejected. That file is Bridge's module registration.
- Put the two `AddTool` calls inline in `server.go`. Rejected. The story asked for a separate file.

Failure mode: a test still uses bare `mcp.ToolCount`. The apply tasks list every current comparison.

### 2. Rules are an ordered match, not an LLM and not a learned ranker

`SuggestNoteKind(intent, contentHint)` lowercases both strings and applies the plan's decision procedure as first-match rules:

1. Timed reminder / calendar / stick → `stick_note` (before generic todo, so "reminder Friday 9am" is not a todo).
2. Board / assignee / column / kanban / stages → `task`.
3. Dated steps / itinerary / chronology → `timeline`.
4. Incident / decision log / "log the … at <time>" → `event_log`.
5. Single action verbs (buy, call, send, submit, fix, pick up, "todo:") → `todo`.
6. Research / cite sources / investigate + findings → `research`.
7. Chapter / long-form / content_hint over ~2000 words → `big_note`.
8. Else `note`.

`use_grill` is true when intent word count is under 15 and the text has no specifics (no URL, ISO date, weekday+clock, or a content_hint over ~40 words), or the ask contains help-me-write / help-me-draft, or it talks about sharing to the pool. A long `content_hint` forces false.

`confidence` is 0.9 on a strong keyword hit, 0.75 on default `note` with specifics, 0.55 on a short vague ask. Empty intent is `invalid_input`. Kind values are the plan names (`stick_note`, not `stick`).

The ~30-row table is the contract. No Ollama client is imported.

Alternatives:

- LLM tie-break as in plan E2.4. Rejected. The story says no LLM; CI unsets provider keys.
- Score every kind and pick the max. Rejected. Ordered rules are smaller and the table can pin them.

### 3. Canonical guide on disk; embed a byte-identical copy

Write `docs/agents/16-dorph-for-agents.md`. Copy it to `dorph/mcp/agent_guide.md` and `//go:embed` that file. A test fails if the two bytes differ. `resources/read` of `dorph://guide` returns the embed (`text/markdown`). `dorph_guide` slices by heading: empty topic returns the outline plus golden rules; a known topic (`kinds`, `sharing`, `grill`, `limits`, `setup`, `checking`, `pool`, `errors`, `tools`, `writing`, `privacy`) returns that section; anything else is `invalid_input` listing those topics.

Initialize instructions stay under 2k, keep the substring `read-only` (existing handshake test), name `dorph://guide` / `dorph_guide`, state private-by-default and never-share-without-consent, and mention pool tools as phase 1B.

Alternatives:

- Read the docs path at runtime. Rejected. The binary's cwd is not the repo.
- Guide only in `dorph/mcp`. Rejected. The plan's canonical path is `docs/agents/16-…`.

### 4. Pack stays where #131 put it; `agent-pack/dorph/` is the portable snippet

Keep updating `.cursor/skills/dorph-notes/SKILL.md` and `.cursor/commands/`. Add the five new commands. Each file stays under 150 lines and points at chapter 16. `docs/agents/15-dorph-notes-self-use.md` keeps the #131 phrases the pack test locks and adds a pointer to 16. `agent-pack/dorph/CLAUDE.md` plus `agent-pack/dorph/cursor-rule.mdc` are the portable copies the plan asked for; they link at the Cursor paths rather than duplicating the skill.

`/dorph-share` and `/dorph-pool-find` name the phase 1B tools and stop if those tools are missing. They do not invent HTTP.

`agentpack_test.go` also scans `agent_runtime.go` for tool names, requires the new commands, and requires the skill to name `suggest_note_kind`, `dorph_guide`, and the consent line.

Alternatives:

- Move the pack to `agent-pack/dorph/` only. Rejected. Cursor loads `.cursor/`.
- Always-apply Cursor rule. Rejected in #131; still rejected. Notes are not every chat.

### 5. Document Bridge's target names; do not implement them

The playbook's tool reference uses the plan names: `create_note` / `get_note` / `list_notes` / `update_note` / `append_to_note` / `archive_note` / `search_my_notes` with `kind`, plus today's aliases (`list_my_tasks`, `get_task`, module create/list/get). Pool tools: `preview_share`, `share_note`, `unshare_note`, `list_my_shares`, `pool_topics`, `pool_search`, `pool_get`, `pool_save_copy`, `pool_report`. Each unimplemented pool tool is labeled coming in phase 1B. PAT setup documents `DORPH_MCP_TOKEN=dorph_pat_…` and the legacy JWT for one phase.

No new handler, table, or token verifier.

## Review

Proposer: return kind `stick` as the story lists it. Reviewer: `create_note` `kind` in the plan is `stick_note`. Returning `stick` teaches the wrong argument. Return `stick_note`. Mention `stick` only as prose.

Proposer: include an Ollama tie-break so E2.4's 85% eval is reachable. Reviewer: the story forbids an LLM call. CI unsets keys. The 30-row table is the acceptance test. Skip Ollama.

Proposer: implement `preview_share` stubs that error. Reviewer: stubs become a second API. Document the names and stop. Phase 1B is Bridge.

Proposer: put both tools in `server.go` next to whoami. Reviewer: Bridge will touch that file for unified notes. One `registerAgentRuntime` line is the whole conflict.

Proposer: bump `ToolCount` in `records_tools.go`. Reviewer: that constant is Bridge's module count. Add `AgentRuntimeToolCount` beside our tools.

Proposer: `dorph_guide` always returns the full markdown. Reviewer: the plan says excerpt by topic. Empty topic returns outline plus golden rules; `dorph://guide` is the full doc.

Proposer: a live 25-ask eval. Reviewer: the epic deferred E3.3. The table test is the check.

Proposer: always-on Cursor rule injecting the playbook. Reviewer: #131 already rejected that. Skill + commands + initialize pointer are enough.

## Risks / Trade-offs

- [Rules misfire on a mixed ask] → Ordered match plus the 30-row table. Ambiguous leftover asks become `note` with lower confidence; the playbook says ask the human if confidence is below 0.6.
- [Guide embed drifts from docs/agents] → Byte-equality test.
- [Handshake tests still expect 20 tools / 2 resources] → Update those assertions in this change. grill_app_test already checks URI presence, not exact count.
- [Bridge lands `get_note` / `list_notes` while the pack still names `get_task` / `list_my_tasks`] → Keep the old names (they stay as aliases). The playbook lists both.
- [Agents share from the new `/dorph-share` command] → The command requires an explicit yes in this conversation and names `consent_required` if the human did not confirm.

## Migration Plan

Additive MCP tools and docs. Rollback is delete the new file, revert the `server.go` hook and instructions, revert tests, and delete chapter 16 plus the new commands. No schema, no process flag.

## Open Questions

None. Kind names, no-LLM, file split, and phase 1B documentation are decided above.
