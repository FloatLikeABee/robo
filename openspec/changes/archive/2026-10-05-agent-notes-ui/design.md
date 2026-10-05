## Context

See proposal.md for why. Code checked before this design:

- `create_note` (PR #133) inserts `user_note_todo` with `ItemType = 'note'` for the Tran user matched by `ownerClause`: `plat_users.id` equals the JWT subject, joined to one active `User` on email (`Deactivated = 0`, `LIMIT 1`). Stored title starts with `[morph-mcp]`. Stored body starts with `source: morph-mcp`. A human can edit either string in Notes & TODOs.
- Notes & TODOs in the SPA (`NotesTodosContent`) loads `GET /api/tran/notes-todos`. `ListUserNotesTodos` and `GetUserNoteTodo` call `tranUserIDFromContext` (`morph/handlers/tran_tools.go`). Verified on `4592fed`:
  - `?user_id=` is returned as-is when it parses as a positive integer. The session is not checked.
  - No active Tran user falls through to numeric `X-User-ID`, then **user id 1**.
  - Two or more active email matches return **0** and do not fall through (the comment and the code agree). PR #133's design text said that case fell through to user id 1; the code does not. A new read must not reintroduce that fallthrough.
- `AuthzMiddleware` requires a session for `/api/tran/` and strips client identity headers before setting `X-User-ID` to `plat_users.id` (a string, not the Tran integer). `ProtectedLayout` sends a missing token to `/login`.
- The Morph AI shell (`App.css`, issues #95–#99 and #126) is the phone shell: dark theme, `100dvh`, safe-area insets, header more-menu at `max-width: 768px`, no horizontal shell scroll. The workspace pane on a phone is a short strip under the chat (`38vh` from the 900px rule). Skills page (`skills-panel-page`) is the full-viewport pattern that already clears safe areas.
- Open PRs #82, #76, and #74 do not touch notes handlers, `appRouter.js`, `SkoolAiChat.js`, or `App.css`. No open Bridge or Echo PR was on the branch list.

## Goals / Non-Goals

**Goals:**

- One page a signed-in human can open to audit agent notes that belong to their session.
- List and detail fail closed: no `user_id`, no user id 1, no note text when logged out or when the session is not exactly one active Tran user.
- Readable at about 390px and at desktop width, reusing the existing shell chrome.

**Non-Goals:**

- Changing `tranUserIDFromContext`, Notes & TODOs writes, or `create_note`.
- A schema column. MorphUtils, Render, and new agent write tools.
- Editing agent notes from this page. The existing editor remains the place a human can change title and body.

## Decisions

### 1. New read routes, not the Notes list

`GET /api/tran/agent-notes` and `GET /api/tran/agent-notes/:id` resolve the owner from `auth_user_id` (the middleware's JWT subject). They count active Tran `User` rows for that `plat_users.id` and email. The count must be exactly one. The statements ignore `user_id` and `X-User-ID`.

Zero rows or two or more rows: HTTP 409, body `{"error":"session does not map to one notes user"}`, no titles. A missing or foreign id: HTTP 404 `{"error":"not found"}` with no title or body. No session: the existing middleware returns 401 before the handler. The handler also returns 401 when `auth_user_id` is empty, so a route registered without the middleware still fails closed.

Alternatives:

- Reuse `GET /api/tran/notes-todos?type=note` and filter in the browser. Rejected. Verified: `?user_id=` returns that user's rows, and a session with no Tran user becomes user id 1. A client filter cannot undo either leak.
- Add `source=morph-mcp` on the existing list and keep `tranUserIDFromContext`. Rejected. The same owner bugs remain, and every Notes caller would share the new parameter.
- Fix `tranUserIDFromContext` for every Tran route. Rejected for this story. The user-id-1 miss is the demo-install behavior of comments, timelines, big notes, and tool notes. OpenSpec `tran-user-admin-guard` left that miss in place on purpose. Changing it here is a different bug and a wide diff.

Failure modes that stay tested: logged-out list and get are 401 and omit the seeded title; `X-User-ID` alone is 401; user B's list and get omit A's title and body; `?user_id=` of A's Tran id still returns only B's notes; a note owned by user id 1 is absent when the session has no Tran user and when two active users share the email.

### 2. Agent rows are a marker match, not a new column

A row is included when `ItemType = 'note'` and either `Title` starts with `[morph-mcp]` or `Body` starts with `source: morph-mcp` (`instr` = 1, case-sensitive). Either marker is enough, because the Notes editor can save a new title and keep the body, or the reverse. Todos are excluded even if a title happens to carry the prefix. Order is `CreatedOn DESC`. The JSON shape is the existing note object (`id`, `title`, `body`, `completed`, `created_on`, `item_type`). The UI maps `completed` to Done or Open.

No schema change. `create_note` already writes both markers, and this story does not change the writer. A column would be empty for notes already stored, and the editor would not maintain it.

Alternatives:

- `source` column. Rejected. It does not identify rows #129 already wrote, and filling it means changing the writer, which is out of scope.
- Title prefix only. Rejected. Saving a new title in Notes & TODOs drops the only marker. PR #133 kept the body line for that reason.
- Body line only. Rejected. A human who clears the body and keeps the title would lose the note from this audit view.
- Trust `LIMIT 1` the way MCP does when two users share an email. Rejected. That picks one of the two and can show the other person's notes. This read returns none.

Ceiling: a human can delete both strings. The note then disappears from this view and remains in Notes & TODOs. A `source` column the editor does not rewrite is the upgrade if attribution must survive an edit.

### 3. Full-page `/agent-notes` inside the existing shell

The route sits under `ProtectedLayout`, next to `/skills`. No token navigates to `/login?returnTo=` and the page does not call the API. The page uses the skills-page shell: dark tokens, `100dvh`, safe-area padding, `overflow-x: hidden`. Desktop is a two-column list and detail. At `max-width: 768px` the list is the screen; choosing a note shows the detail and a back control of at least 44px. Titles use `text-wrap: balance` only on the short page heading. Body copy uses `text-wrap: pretty` with `overflow-wrap: anywhere`, so a long token wraps inside the column. Unsupported browsers keep normal wrap.

The header more menu (the phone `⋯` control, and the desktop chip bar) gains an "Agent notes" item that assigns `/agent-notes` in the same tab. `HeaderMoreMenu` opens `href` in a new tab, so this item is an `onClick`, not an `href`.

Empty copy: agents post notes for this sign-in with the Morph note tool; nothing from an agent is here yet. A 409 says the sign-in does not map to one notes user, and the list stays empty.

Alternatives:

- A filter chip on Notes & TODOs. Rejected. That list is the unsafe endpoint, and the column uses `noWrap` in a narrow pane, which hides the marker and the body.
- A third Agent workspace tab. Rejected as the primary view. On a phone the workspace is a `38vh` strip under the chat, so list and detail are not readable. The tab would also fight the saved workspace tab.
- A new drawer beside `ChatNotesTodosDrawer`. Rejected. That drawer is never opened (`notesDrawerOpen` stays false). A drawer on a phone covers the shell and does not add a login gate of its own.

### 4. Detail is read-only text

The body is shown as text (`white-space: pre-wrap`), not the Notes markdown editor. The audit view must not invite a save that strips the marker, and it must show the `source: morph-mcp` line as stored. The human edits in Notes & TODOs.

## Review

Proposer: filter the existing notes list in the client. Reviewer: `tranUserIDFromContext` honors `?user_id=` and, with no Tran user, returns user id 1. Checked in `tran_tools.go` on this tip. A browser filter still downloads the wrong rows. New routes only.

Proposer: ambiguous email can use MCP's `LIMIT 1`, since that is who `create_note` writes. Reviewer: two active users with one email are not a safe owner. `LIMIT 1` can attach the note to the other person. Return 409 and no text. MCP's write behavior stays unchanged.

Proposer: put the list in the workspace tab so the phone shell is reused. Reviewer: the phone workspace height is the 900px rule's `38vh`. A full page that reuses the skills shell (safe area, dark, `100dvh`) and the header more menu (44px, phone overflow) is the readable reuse. Do not restyle the chat grid.

Proposer: add a `source` column so an edit cannot hide a note. Reviewer: existing rows have no column, and the writer is out of scope. Match either stored marker. Document the ceiling.

Three failure modes that stay tested: logged-out request is 401 and contains no seeded title; user B cannot see user A's agent note; no Tran user and two Tran users do not return user id 1's note.

## Risks / Trade-offs

- [Marker can be edited away] → Either prefix still matches. Both removed means the note leaves this view. Documented ceiling, no column.
- [Email `LIMIT 1` on the writer] → This read refuses a session that matches two active users, so it will not display a note MCP attached to one of them. That is the closed failure.
- [Existing Notes list stays leaky] → Called out so this page does not call it. Not fixed here.
- [`text-wrap: pretty` is not in Firefox] → Ignored there; `overflow-wrap: anywhere` still wraps long tokens.
- [Header entry touches `SkoolAiChat.js`] → Open PRs #82, #76, and #74 do not edit that file. The change is one menu item.

## Migration Plan

No schema change. Rollback is removing the routes, the page, and the header item. Existing notes stay. The API process keeps the old Notes endpoints.

## Open Questions

None. The read path, the marker rule, and the page are decided above.
