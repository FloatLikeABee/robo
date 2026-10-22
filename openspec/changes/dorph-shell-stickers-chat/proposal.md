## Why

Dorph AI still lets the right workspace collapse behind a Hide workspace control, opens on Notes & TODOs ahead of Context & Knowledge, and shows a white frame while AI tools loads. Long assistant replies stay trapped in a narrow bubble. DorphNotes Tasks are uneven cards with one-line text, and the Project module still uses violet and gray chrome instead of the shared dark blue.

## What Changes

- Remove the Dorph AI Hide workspace / Show workspace control. The right workspace stays visible. On a phone the same pane stays in the existing lower band; it is no longer closed on first load, and a saved "closed" choice is ignored.
- Put **Context & Knowledge** before **Notes & TODOs**, and select Context & Knowledge when that session has no stored tab.
- Assistant chat bubbles taller than about ten visual lines gain an enlarge control that opens the reply in a dark, near-full-viewport modal. Existing diagram and image enlarge stays as it is.
- AI tools, on first open and on every reopen, shows a dark loading surface with an indicator until the embedded frame finishes loading. The frame wrapper is no longer white.
- DorphUtils **Project** chrome (page, header, cards, selected tabs, primary buttons, create modal) uses the shared dark-blue palette. Deep violet and neutral-gray panel fills go away. A destructive confirm may stay a distinct danger color.
- DorphNotes visible label **Big notes** becomes **Stories**. **Tasks** becomes **Stick notes**. Stick notes list as equal square stickers in a tight grid, each a stable different color from a fixed palette, with more of the note body visible, plus search. Routes and APIs stay `/big-notes` and `/case-tasks`. `/stories` stays Timelines.

## Capabilities

### New Capabilities

- `dorphai-workspace-chrome`: Workspace pane cannot be hidden; Context & Knowledge is the first tab and the default when nothing is stored.
- `chat-reply-enlarge`: Long assistant replies can open in a dark read modal.
- `ai-tools-loading-shell`: AI tools modal uses a dark loading state instead of a white frame.
- `project-dark-blue`: Project module chrome is dark blue.
- `dorphnotes-stick-notes`: Stories and Stick notes labels; Stick notes is a searchable equal-size sticker grid.

### Modified Capabilities

- `dorph-phone-chat`: Phone workspace is no longer closed until the operator opens it, and there is no open/close control. The chat column still uses the full width.

## Impact

- Dorph AI: `SkoolAiChat.js`, `AgentWorkspace.js`, `App.css`, `AiToolsWorkspaceDrawer.js`, assistant message render, workspace-open tests, phone layout tests.
- DorphNotes: `AppDrawer.js`, `CaseTasks.js`, `BigNotes.js`, `platformUiDefaults.js` labels only. No API or route rename.
- Project (`dorph-engi` frontend): `app.css`, `AppLayout.svelte`, `App.svelte`, `ProjectDocumentPanel.svelte`, module icon. No API change.
- `openspec/specs/dorph-phone-chat`: replace the closed-until-asked requirement.
- Do not restore a Dorph AI Files workspace.
