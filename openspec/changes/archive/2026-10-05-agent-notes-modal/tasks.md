## 1. Failing tests

- [x] 1.1 Add a failing modal test: a signed-in list shows title, time, and status; choosing a row shows the body; an empty list explains the Morph note tool; a failed load shows the error and no rows.
- [x] 1.2 Add a failing dismiss test: Escape and the close control call close, return focus to the opener, and stop the Escape event so a bubble listener does not also run.
- [x] 1.3 Add a failing entry test: the header Agent notes control opens the dialog and does not assign `/agent-notes`; `/agent-notes` with no token goes to login and does not call the list; with a token it lands on the chat with the dialog open and the `agent-notes` query removed.
- [x] 1.4 Point the phone CSS assertions at the modal panel: safe areas, wrap, one column at 768px, and a 44px back control.

## 2. Modal over the chat

- [x] 2.1 Implement `AgentNotesModal` with the drawer chrome, dialog semantics, capture-phase Tab and Escape via `onSheetKeyDown`, scroll lock, and the existing list and detail reading. Load `GET /api/tran/agent-notes` only.
- [x] 2.2 Open it from the header and more-menu item. Remove `window.location.assign('/agent-notes')`. Remember the chip or the More apps button as the focus restore target.
- [x] 2.3 Replace the `/agent-notes` page with a replace redirect to `/?agent-notes=1`, and have the chat open the modal from that flag and then remove the flag without remounting. Delete the standalone page component.

## 3. Docs and check

- [x] 3.1 Update the agent-notes sentences in `docs/agents/14-morph-mcp.md` so the human list is the chat modal, and `/agent-notes` opens that same modal.
- [x] 3.2 Run the frontend tests and production build. Do not change Go handlers.
