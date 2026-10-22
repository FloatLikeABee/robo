## Context

Notes & TODOs AI assist is one `POST /api/tran/text-assist` call. The prompt includes `dorphai.ReplyShapeInstructions` ("Answer the question in plain language"), so a title that looks like a request is answered as chat. The notes MCP app is a closed HTML document at `ui://dorph/notes` that calls tools through the host bridge (`postMessage` `tools/call`). dorph-mcp is one verified session, does not open Badger, and does not write a note except through `create_note`. Private `/api/tran/` routes already require a Dorph JWT. See proposal.md for why the one-shot path is the bug.

## Goals / Non-Goals

**Goals:**

- One interview state machine and one HTML page, used by the SPA modal and by MCP hosts.
- Two model calls at most (plan, then write). The server owns the question list, the answers, and the stop rules.
- Owner checks match the notes tools: the verified user, no caller-supplied user id, not-found for someone else's grill.
- A draft that is still a greeting is replaced before the operator sees it.

**Non-Goals:**

- Big notes, timelines, research, stories, and task-chain steps. Those generators stay on their own routes.
- Persisting the interview. Restarting the API or dorph-mcp drops open grills.
- Auto-saving the SPA note on confirm. Confirm fills the editor. Create and Save stay where they are.
- Changing Render, images, or `text-assist` for task-chain steps.

## Decisions

### 1. Plan once, ask locally, write once

`QuestionCap` scores the title only. A long existing body is context for the prompts (truncated), and it does not raise the count. The cap is 1 for a short single intent, and at most 4 when the title is long, has a clause break, or walks back the idea ("actually"). The plan prompt asks for exactly that many questions, one per line, each ending in `?`, with no greeting.

The model is not asked again between questions. Skip, next, and finish are state transitions. Finish, or an answer that is only a stop phrase ("enough", "write it now"), moves to the draft. That keeps a small Ollama model on two short prompts instead of a chat transcript it will answer in character.

**Alternatives rejected:**

- One-shot generation with a stricter prompt. The screenshot is that path. The operator still cannot correct a misunderstood title.
- A model call after every answer. Small models drift, and each call is a long wait on local Ollama.
- A fixed questionnaire. It ignores a one-line title and a product-shaped title the same way.

### 2. The server rejects chat, the model does not get a vote

Parsed questions must be on the title: at least one significant token from the title, or the whole set is replaced by fallback questions that name the title. The write prompt asks for the body only. It does not include `ReplyShapeInstructions` and does not say "answer the question".

If the draft greets, introduces the assistant, or offers more help, and there is no substantial remainder, the shown draft is built from the title, the context, and the answers. Skipped questions are omitted. A TODO becomes bullets. The operator still confirms.

A stop phrase is the whole answer ("enough", "write it now"), not a sentence that merely contains those words. An empty answer does not advance. If the write call fails, the session stays on the interview and keeps the answers. The editor body does not change.

**Alternative rejected:** show the raw model text and hope the operator edits it. That is the current bug, moved into a modal.

### 3. One HTML document, two hosts

`dorph/grill` embeds `app.html`. dorph-mcp serves it at `ui://dorph/grill`. `GET /api/tran/grill/ui` returns the same bytes to the signed-in SPA. The modal fetches that HTML with the bearer client and sets the iframe `srcDoc`. A raw iframe navigation would not attach the Authorization header. The modal answers the app bridge:

- `ui/initialize` gets a host result.
- `tools/call` `note_grill` is `POST /api/tran/grill` for the SPA, and the real tool for an MCP host.
- After initialize, the SPA sends a host-context message with kind, title, and description so the title is not put in a URL.
- On review, the SPA shows the draft in the page and enables Use this on the dialog. That button inserts the draft from the last grill result, not a string the page invents. The page does not call `create_note` on the SPA surface. An MCP host shows Save note only after the draft is on screen, and that button calls `create_note`.

The page uses `textContent`, a closed content security policy, and the dark Arial palette from `dorph/internal/htmldoc` (background `#070b14`, accents cyan, violet, and mint).

**Alternative rejected:** a second React conversation. It would drift from the MCP page, which is the UI the issue asks to embed.

### 4. Sessions stay in memory and on the verified user

The API store keys sessions by `auth_user_id` from the JWT middleware. It does not use `tranUserIDFromContext`, which can fall through to user id 1. A missing user id is 401 and does not call the model. Another user's id is 404 with no draft.

dorph-mcp keeps its own store inside the process, keyed by the token subject. `note_grill` is not read-only. It does not open SQLite. `create_note` is unchanged, including the `[dorph-mcp]` marker.

A process holds at most 32 grills per user and drops the oldest. That ceiling is in-memory and per process. A shared store would be a table; this interview does not need to survive a restart.

`POST /api/tran/grill` takes `action` (`start`, `answer`, `skip`, `finish`), `id`, `kind` (`note` or `todo`), `title`, `description`, and `answer`. The response is `id`, `phase` (`asking` or `review`), `question`, `asked`, `planned`, and `draft` when reviewing.

### 5. Tests do not call Ollama

The completer is an argument. Tests pass a function. The default completer is the existing Dorph AI client. Prompt builders, the cap, the parser, the chatty-draft replacement, and the owner check are pure tests. The MCP test reads `ui://dorph/grill` and drives `note_grill` with the fake completer. The SPA test drives the host bridge and checks that confirm inserts and cancel does not.

## Risks / Trade-offs

- [In-memory grills vanish on restart] → The modal is short. The operator starts again. No partial note is saved.
- [API and dorph-mcp do not share grill ids] → Each client interviews on its own process. That matches how notes tools already split HTTP from stdio.
- [Fallback prose is plain] → It is body text about the title and the answers, which is the bug we are fixing. The operator can edit after confirm.
- [Spec used to say the only MCP resource is the notes app] → This change adds `ui://dorph/grill`. Module tools that already exist stay listed by inclusion, not by a new exclusive count.
- [iframe bridge is the SPA integration] → Cancel stays outside the iframe so a failed script can still dismiss the modal.

## Migration Plan

Local only. No data migration. Rollback is reverting the commit. Existing notes are untouched. `text-assist` remains for task-chain steps.

## Self-grill

Proposer and reviewer were the same pass, before implementation.

- Scoring the existing body would make every long edit ask four questions. The cap uses the title only. The body stays in the write prompt, truncated.
- An iframe `src` to the HTML route would drop the bearer token. The SPA fetches the document and uses `srcDoc`.
- The page must not be able to insert a greeting the server already replaced. Use this inserts the last review draft from the server.
- "enough detail about the API" is an answer. Only a whole-answer stop phrase finishes early.
- A failed write call must not drop the answers or touch the editor.

## Open Questions

None that change the spec. Big notes stay out until someone asks for that editor.
