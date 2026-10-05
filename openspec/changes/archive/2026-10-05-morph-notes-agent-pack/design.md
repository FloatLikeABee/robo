# Design

## Context

See proposal.md for why. Code checked before this design:

- `morph/mcp/server.go` registers exactly four tools: `whoami`, `list_my_tasks`, `get_task`, and `create_note`. There is no `list_notes`, `get_note`, or `create_todo`.
- `create_note` input is `title` and `body` (json tags on `createNoteInput`). At least one is required after trim. Title max 200 runes. Body max 32000 runes, before the source line. The server stores `[morph-mcp]` plus the title, and `source: morph-mcp` plus the body (`agentTitle`, `agentBody` in `morph/mcp/tasks.go`). A second call with the same stored title and body returns the existing id.
- `list_my_tasks` input is `type` (`all`, `note`, `todo`), `status` (`all`, `open`, `done`), and `limit` (default 50, cap 100, negative errors). Result is `{tasks, limit}`. For `type=note`, rows sort `CreatedOn DESC`.
- `get_task` input is integer `id`. `id <= 0` is `invalid id`. A missing row or another user's row is `not found`.
- Each result task is `id`, `item_type`, `title`, `status`, `body`, and optional `deadline_at`, `created_on`, `last_updated`.
- `create_note` always inserts `item_type` `note`. It does not create a todo, a CaseTask, a big note, or a tool note.
- The HTTP handler `POST /api/tran/notes-todos` is the path Bridge rejected: `tranUserIDFromContext` can store the row as user id 1. Confirmed in `openspec/changes/archive/2026-10-05-mcp-create-morph-notes/design.md` and left unchanged here.
- Cursor already loads `.cursor/skills/<name>/SKILL.md` and `.cursor/commands/<name>.md`. There is no repo-root `AGENTS.md`. Agent handbook chapters live in `docs/agents/`.
- Local human review is `http://localhost:3031/`. `morph-mcp` is stdio. It is not an HTTP server on 3031 or 9090.

## Goals / Non-Goals

**Goals:**

- One skill an agent can load and then create, list, and get notes with the real tools.
- Three slash commands that map onto those three calls.
- One instruction doc a non-Cursor bot can follow with no human click-path.
- One test that fails if the pack drops a real tool name or teaches the HTTP write.

**Non-Goals:**

- New MCP tools, prompts, or resources. Edits to `create_note` behavior. MCP app UI (#130).
- A `/morph-whoami` command. `whoami` stays in the skill and the doc.
- Render, deploy config, the Notes SPA, and `pkg/morphai`.
- An always-on rule that injects this pack into unrelated chats.

## Decisions

### 1. One skill holds the contract; the doc repeats it; commands stay short

`.cursor/skills/morph-notes/SKILL.md` is the Cursor entry. Its description mentions Morph notes, Notes & TODOs, and the three tool names so a notes request can load it. The body states when to call each tool, the arguments, the stored marker, the list cap, session scope, and the two forbidden paths (clicking `:3031`, posting `/api/tran/notes-todos`).

`docs/agents/15-morph-notes-self-use.md` is the same contract in handbook form, including the "tools missing" branch and a pointer to the `mcp.json` example in `docs/agents/14-morph-mcp.md`. It does not embed a token or a secret. A one-line pointer in `14-morph-mcp.md` and a Next bullet in `00-architecture-overview.md` make the chapter findable. Those lines do not restate the tool schema.

Each command under `.cursor/commands/` names one tool, how `$ARGUMENTS` map onto that tool's fields, and the same two forbidden paths. It points at the skill for the rest. The list command defaults `type` to `note`. The get command refuses a missing or non-integer id. The create command refuses empty arguments.

Alternatives:

- Always-apply `.cursor/rules` text. Rejected. The pack is for a notes task, not every session, and a rule does not help a bot that never loads Cursor rules.
- Three skills. Rejected. Create without the list-cap and the forbidden HTTP path is how an agent posts twice or posts to the wrong API. One skill is the unit that "knows when and how".
- Doc only, skill is a link. Rejected. Loading the skill has to be enough. Agents skip linked files.
- Skill only, no handbook chapter. Rejected. The acceptance test is another bot told to use Morph, not another Cursor session that auto-loads skills.
- Generate the doc from the skill. Rejected. A second file plus one assertion is smaller than a generator.

Failure mode: the two long copies drift. The test in decision 4 fails when either copy drops a required phrase.

### 2. Commands are prompts that call MCP, not HTTP stubs and not click stubs

Cursor slash commands are markdown the agent receives. They cannot shell to `morph-mcp`. The mapping is: the command text tells that agent to call the tool. If the tool is not in the session, the command says to stop and report that `morph-mcp` is not connected, and to point at `docs/agents/14-morph-mcp.md` for the local `mcp.json` shape. It does not open a browser.

Alternatives:

- Documented stubs that say "ask a human to click Notes & TODOs". Rejected. That is the failure this story exists to remove. The issue allows a stub only where the command cannot map to a flow. Here the command can name the tool.
- A small CLI that wraps the tools. Rejected. The tools already exist. A second client would need the JWT in the shell and would duplicate `create_note`.
- Teach `POST /api/tran/notes-todos` with the session cookie. Rejected. Bridge's design showed that handler can write as Tran user id 1 when the email matches no active Tran user, or when two active users share the email and the helper returns 0. A skill that uses it would store a note the MCP list cannot see, on the wrong user. Re-checked against the archived design; this change does not "fix" that helper.

Failure mode: an agent treats the command as permission to improvise a client. The command's first lines name the only tool and the two forbidden paths. The test asserts those lines.

### 3. Argument mapping is literal

- `/morph-note`: `$ARGUMENTS` required. The first non-empty line is `title` (trim to 200 Unicode code points). The full argument text is `body`, so the note text is not title-only. A later human edit of the title keeps the body, which is the ceiling Bridge already documented. One line still sends both. Empty arguments, or a body over 32000 code points: do not call.
- `/morph-notes`: `$ARGUMENTS` optional. Default call is `list_my_tasks` with `type=note`. Tokens `type=`, `status=`, and `limit=` override that. Omit `limit` for the default 50. `limit` 0 is that default, not an empty page. A negative limit errors. Anything else is not a filter the tool accepts; ignore it rather than invent a query language.
- `/morph-note-get`: `$ARGUMENTS` is the integer id. Missing or not an integer: do not call.

The server still applies its own limits. The pack states them so the agent does not retry a rejected create through HTTP. Send a plain title. The server adds `[morph-mcp]` once; a title that already has that prefix is not prefixed twice.

`whoami` is optional. Call it when it is available and the acting user is unknown. Do not refuse a create solely because `whoami` was not called. It is not a fourth command. Do not teach the agent to run the login curl, read `.env`, or invent `MORPH_MCP_TOKEN`. A missing server is a report to the human, plus the existing `mcp.json` section in `docs/agents/14-morph-mcp.md`.

### 4. One Go test locks the pack to `server.go`

`morph/mcp/agentpack_test.go` finds the repo root by walking parents for `docs/agents/14-morph-mcp.md`, then reads the skill, the three commands, and the instruction. It extracts tool names from `server.go` (`readOnly(` first string, and `Name:` strings on the `create_note` tool). It asserts:

- the extracted set contains `create_note`, `list_my_tasks`, and `get_task`. A rename in `server.go` fails this test even if the pack is unchanged
- the skill and the instruction contain each of those three names, plus `title`, `body`, `type`, `status`, `limit`, `[morph-mcp]`, `source: morph-mcp`, `localhost:3031`, `Do not POST /api/tran/notes-todos`, and `not a notes user`
- `morph-note.md` contains `create_note`, `morph-notes.md` contains `list_my_tasks`, `morph-note-get.md` contains `get_task`
- the skill, the instruction, and all three commands contain the literal `Do not POST /api/tran/notes-todos`
- none of the pack files contain `eyJ`

The test does not parse the full prose and does not start `morph-mcp`.

Alternatives:

- A Python script under `scripts/`. Rejected. The tool names live in the Go server. A test in that package fails in the same `go test ./mcp` run as the tools.
- No test, review the markdown by hand. Rejected. Two copies will drift, and the ponytail rule wants one runnable check for this contract.

Failure mode: the walk-up misses the repo root in a nested module. The test fails closed if `14-morph-mcp.md` is not found. `go test` cwd is `morph/mcp`; two parents up is the repo root today, and the walk does not assume that depth.

## Review

Proposer: put the procedure only in the skill and tell other bots to open `.cursor/skills/morph-notes/SKILL.md`. Reviewer: the story asks for an instruction doc another bot can follow. A Cursor skill path is not the agent handbook. Keep `docs/agents/15-morph-notes-self-use.md` as a full procedure. The skill still contains the contract, because "loads the skill" has to be sufficient.

Proposer: slash commands should be stubs that say MCP-only, go click `:3031`. Reviewer: the issue allows a stub when the command cannot map. These commands map: they name `create_note`, `list_my_tasks`, and `get_task`. A click stub fails the acceptance line about no human operating every step. The missing-tool branch reports the gap. It does not click.

Proposer: one skill per tool. Reviewer: the dangerous cases are cross-tool. List is capped. Create is not a todo. HTTP create can write as user id 1. Those warnings belong in the one text the agent loads for any note task.

Proposer: skip the test and the pointers. Reviewer: without the test, nothing fails when a later edit renames a tool in prose only. Without the pointer in `14-morph-mcp.md`, a bot that starts from the MCP chapter never finds the instruction. Both stay. Neither changes tool behavior.

Proposer: require `whoami` before every create. Reviewer: a session that has `create_note` and not `whoami` would stall, and the tool already acts as the token subject. `whoami` stays optional.

Proposer: repeat the login curl so the agent can mint `MORPH_MCP_TOKEN`. Reviewer: that sends the agent after the human's password and `.env`. The note tools are usable only after the human has connected `morph-mcp`. The pack reports a missing server. It does not collect secrets. The test rejects an `eyJ` prefix in the pack.

Proposer: a one-line note can be title-only, since the list column is the title. Reviewer: Bridge already showed a title edit drops a title-only marker, and the same editor would drop title-only content. The body gets the full argument.

Proposer: assert the pack merely contains the substring `POST /api/tran/notes-todos`. Reviewer: a sentence that recommends that route would pass. The locked phrase is `Do not POST /api/tran/notes-todos`.

Three failure modes that stay explicit in the pack and the test: the HTTP notes route appears only in that forbid phrase; `:3031` is the human review origin; the tool names in the pack are the names extracted from `server.go`, including `list_my_tasks` and `get_task` rather than names invented from the issue's "list notes" / "get note" wording. A create error of `not a notes user` stops the agent. It does not fall through to user id 1.

## Risks / Trade-offs

- [Two copies of the contract] → The Go test checks the shared phrases. It will not catch a subtle contradiction that still contains those phrases.
- [Commands assume the morph MCP server is already configured] → The missing-tool branch reports that and points at `14-morph-mcp.md`. This change does not launch the server or mint a JWT.
- [Duplicate create returns the old id] → The pack says a repeat of the same stored title and body is the existing note. The agent must not "fix" that by posting HTTP.
- [List page is capped] → The pack says to `get_task` when the id is known, and not to treat a short page as the full set.
- [Marker is editable by the human] → Same ceiling as #129. The pack tells the agent the server writes the marker. It does not ask the agent to add a column.

## Migration Plan

Docs and a test only. Rollback is deleting the skill, the three commands, the new chapter, the two pointers, and `agentpack_test.go`. No data migration. No process restart. Local Morph stays `http://localhost:3031/`.

## Open Questions

None. The file split, the command mapping, and the forbidden HTTP path are decided above.
