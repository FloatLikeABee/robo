## Context

See proposal.md for why. Code checked on `main` at `5de4ca9`, which contains #137 (`9873cbc`). Ge's Mac may have uncommitted modal chrome for Morph Notes. That tree is not on this branch. The decisions below use only what is on `main`.

What `main` actually does:

- The header chip "MorphNotes" is an `href` to `/morphdata`. `HeaderMoreMenu` opens `href` items in a new tab. It is not an overlay.
- `ChatNotesTodosDrawer` is the notes-shaped overlay (`hybrid-drawer-overlay` plus a panel `min(94vw, 900px)`). `notesDrawerOpen` is never set to true, so it does not open.
- AI tools (`AiToolsWorkspaceDrawer`) is the live large overlay: same overlay class, panel `min(96vw, 1200px)`, close button, overlay `mousedown` dismisses. It has no Escape handler, no `role="dialog"`, and no focus move.
- Skills (`SkillsModal`) is the live dialog: centered panel, `role="dialog"`, Escape, `body` scroll lock, overlay click dismisses. It does not move focus in or restore it.
- The enlarge overlay saves `document.activeElement` and restores it. `onSheetKeyDown` (`phoneSheet.js`) traps Tab and handles Escape for the sessions sheet.
- Agent notes is `AgentNotesPage` at `/agent-notes` under `ProtectedLayout`. The menu item calls `window.location.assign('/agent-notes')`. The list request is `GET /api/tran/agent-notes`. The response rows already include `body`, so the page never calls get-by-id.
- At `max-width: 768px` the page is one column, with a back control of at least 44px, `overflow-wrap: anywhere`, and safe-area padding. `hybrid-drawer` at that breakpoint is full width and pads the safe areas.

## Goals / Non-Goals

**Goals:**

- One agent-notes surface, opened over the chat, with the same list, detail, empty, and error behavior.
- Escape, a close control, and a click on the overlay dismiss it. Focus moves in and returns to the opener.
- `/agent-notes` with no session still goes to login. With a session it opens this same surface on the chat.
- No change to the agent-notes handlers or to `tranUserIDFromContext`.

**Non-Goals:**

- Restyling MorphNotes, AI tools, or Skills.
- Enabling `ChatNotesTodosDrawer`.
- Editing notes, calling get-by-id, or changing ownership rules.
- MorphUtils, Render, or a shared overlay framework.

## Decisions

### 1. A chat modal that copies the drawer chrome, not a new shared component

`AgentNotesModal` is opened from `SkoolAiChat` and portaled to `document.body`. The agent shell sets `.chat-container` to `display: contents`, so an overlay left in that tree becomes a grid item of `.app` and is clipped by `overflow: hidden`. AI tools and Skills sit in that tree on `main`; copying that placement does not produce a viewport overlay. The portal is the smallest way to get a real one. It still paints above the chat. The shell uses `hybrid-drawer-overlay` and `hybrid-drawer` so the dark tokens, phone full-bleed, and safe-area padding stay the ones the shell already ships. The panel width matches the AI tools panel (`min(96vw, 1200px)`), in CSS rather than an inline style. The list and detail markup moves out of `AgentNotesPage` unchanged in behavior: title, time, Done/Open, body as text, empty copy, and `err.response.data.error` with no rows.

The modal adds the dialog contract those drawers do not have, using pieces that already exist:

- `role="dialog"`, `aria-modal="true"`, `aria-labelledby` the title.
- Close control labeled "Close agent notes". Overlay `mousedown` dismisses only when the target is the overlay.
- On open, focus the close control. On close, focus the opener if it is still connected. The desktop chip is that opener. A more-menu item is unmounted when the menu closes, so the opener saved for that path is the "More apps" button, which stays mounted. A deep link has no opener; focus the composer.
- Tab and Escape go through `onSheetKeyDown` on a window `keydown` listener in the capture phase. Escape calls `preventDefault` and `stopPropagation` before other listeners. The header more menu, the skills modal, and the sessions sheet listen on bubble, so one Escape closes agent notes only.
- `document.body.style.overflow` is `hidden` while the modal is mounted and restored on unmount. The parent passes a stable `onClose`.

`HeaderMoreMenu` already closes itself when an item is clicked, and it opens `href` in a new tab. The item stays an `onClick`. It sets `hasPopup: 'dialog'` and `expanded` the way Skills and AI tools do. It does not assign a URL.

Alternatives:

- Extract one `ChatOverlay` and migrate Skills, AI tools, and this modal. Rejected. The three shells differ (iframe drawer, form dialog, read-only list). Migrating the others is an unrelated redesign. Ceiling: the overlay chrome is copied. The upgrade is a shared shell if a fourth overlay needs the same focus contract.
- Reuse `SkillsModal`. Rejected. It is the skills editor, not a panel slot.
- Open `ChatNotesTodosDrawer` and filter agent rows in the browser. Rejected. That drawer is unwired, and its list is `GET /api/tran/notes-todos`, which still honors `user_id` and can fall through to user id 1. A client filter cannot undo that.
- Native `<dialog>` with `showModal()` and `closedby="any"`. The current web guidance prefers this for focus trap, Escape, and light dismiss. Rejected here. The chat overlays on `main` are divs; a top-layer dialog would not share their backdrop or z-index. `closedby` is unsupported in Safari, so light dismiss still needs a click handler. CRA's jsdom (react-scripts 5) does not implement `showModal`, so the dismiss and focus contract could not be tested the way this frontend tests. The div overlay plus `onSheetKeyDown` is the contract we can match and test.
- Keep `AgentNotesPage` and iframe it. Rejected. That is a second document and a second chrome.

### 2. Deep link redirects into chat and is consumed once

`/agent-notes` stays under `ProtectedLayout`. Its element is a redirect to `/?agent-notes=1` (replace). `SkoolAiChat` reads that flag in the initial state, opens the modal, then `setSearchParams` removes the flag with replace. The route element does not change, so the chat does not remount and the modal stays open. Refresh after that does not reopen it. A bookmark of `/agent-notes` still works the next time.

Logged out, `ProtectedLayout` sends the human to `/login?returnTo=` with the agent-notes path before the redirect runs. The view does not call the API. After login, the existing return path hits `/agent-notes` and the redirect opens the modal.

The menu path does not write the query. Opening from the header leaves the address on the chat.

Alternatives:

- Delete the route. Rejected. `/agent-notes` would fall through to the chat catch-all with the modal closed. The bookmark would not open the list.
- Leave the flag in the query while the modal is open. Rejected. Refresh reopens it, and the menu would have to write the URL to share one mechanism. That is a navigation the menu must not do.
- Render a second `App` on `/agent-notes` and navigate to `/` on close. Rejected. Those are different route elements, so closing remounts the chat and drops the in-memory thread.
- Keep the full page for the deep link. Rejected. Two chromes.

### 3. Same list API, no get-by-id, no handler edits

The modal calls `tranEndpoints.agentNotes` only. Rows include `body`. Get-by-id stays on the server for other callers and is not used by this view. No Go change. A 401 still follows the existing `tranApi` login redirect. A 409 or other error shows the server `error` string and no rows.

### 4. Phone layout stays the #132 column

Below 768px the existing rule remains: one column, list until a row is chosen, detail with a back control of at least 44px, long tokens wrap. The panel itself is the full-bleed `hybrid-drawer` (safe areas, no horizontal shell scroll). There is no text field, so the phone keyboard does not open; focusing the close control blurs the composer. The back control is "back to the list", not "close the modal".

## Review

Proposer: treat Morph Notes on `main` as the modal to clone. Reviewer: the header MorphNotes control is a new-tab link, and `notesDrawerOpen` is never set. Checked in `SkoolAiChat.js` and `HeaderMoreMenu.js`. Clone the AI tools / notes-drawer chrome that is actually rendered, and add the dialog behavior from Skills, the enlarge overlay, and `onSheetKeyDown`. Say in the PR that the Mac WIP was not available.

Proposer: use `<dialog showModal()>` because the web guidance requires it. Reviewer: it would not match the div overlays, Safari still needs a light-dismiss fallback, and jsdom cannot open it. Keep the div. Record the rejection.

Proposer: one shared overlay component now, so Escape is consistent. Reviewer: Skills and AI tools are out of scope. A capture listener that stops Escape is enough so this modal does not also dismiss them. Do not add Escape to AI tools in this change.

Proposer: leave `/agent-notes` as the page so the tests keep passing. Reviewer: that is the second UI the story forbids. Move the tests onto the modal and the redirect.

Proposer: clear the deep-link query by navigating to `/`. Reviewer: a different route element remounts `SkoolAiChat`. Remove the query on the same `*` route. Test that the dialog is still open after the query is gone.

Proposer: call get-by-id when a row is selected so the detail is fresh. Reviewer: the list payload already has `body`, and a second request is a new failure mode. Select from the list.

Failure modes that stay tested: the menu opens the dialog and does not assign `/agent-notes`; Escape and the close control dismiss it and return focus, and a bubble Escape listener does not also run; the deep link opens the dialog and then drops the query without closing it; a logged-out `/agent-notes` visit goes to login and does not call the list; empty and error states still render; phone CSS keeps one column, a 44px back control, and safe-area padding.

## Risks / Trade-offs

- [Overlay chrome is copied] → Named in the modal file. A shared shell is the upgrade if another overlay needs this focus contract.
- [Escape does not close AI tools] → Unchanged. Agent notes consumes Escape only while it is open.
- [More-menu item unmounts before focus returns] → Restore target is the More apps button, saved during the click.
- [Deep link has no opener] → Close focuses the composer.
- [`text-wrap: pretty` is not in Firefox] → Same as #132. `overflow-wrap: anywhere` still wraps.
- [Query clear depends on the chat not remounting] → Same route element. Covered by the deep-link test.

## Migration Plan

No schema change and no API change. Rollback is restoring the page route and the `location.assign` menu item. Existing notes stay.

## Open Questions

None. The chrome, the deep link, and the API are decided above.
