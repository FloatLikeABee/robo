## Why

MorphNotes still exposes Settings File import that operators no longer use, Task start/end dates are optional so records can be saved without a window, and HTML/Markdown previews in MorphNotes and MorphUtils Project show nested or unthemed scrollbars. MorphUtils also still shows Sign out / Out even though session is Morph JWT and operators do not need a second logout.

## What Changes

- Remove MorphNotes **Settings → File import** entirely: nav item, route, page, and the dedicated import API used only by that page. Generic data CRUD and its extract-for-review flow stay.
- **BREAKING:** `POST /api/tran/generic-data/import` is removed. Settings `/admin/configuration/file-import` (and `/configuration/data-import` alias) go away.
- MorphNotes **Tasks** create/edit: **Start** and **End** dates are required. New drafts default both to the current local calendar day.
- MorphNotes **Timelines** HTML view: one vertical scrollbar only (the HTML content). Outer panel does not scroll. Inner scrollbar is dark to match the theme.
- MorphNotes **Big Notes** preview (and HTML tab if it has the same double-scroll): same rule — outer overflow gone, inner content scroll, dark scrollbar.
- MorphUtils shell and every module shown inside it (Event Logs, Content Maker, Data Access, Project): remove Logout, Sign out, and Out controls. Morph AI Sign out on the Morph home chat is unchanged.
- MorphUtils **Project** Markdown and HTML preview scrollbars use the dark theme instead of default light browser chrome.

## Capabilities

### New Capabilities

- `remove-morphnotes-file-import`: Drop Settings File import UI and its dedicated backend import endpoint.
- `case-task-required-dates`: Task start and end dates required; new forms default to today.
- `morphnotes-preview-scroll`: Timelines HTML and Big Notes preview use a single inner, dark-themed vertical scrollbar.
- `morphutils-no-logout`: MorphUtils and its modules have no logout/sign-out/out actions.
- `morphutils-project-themed-scroll`: Project Markdown/HTML preview scrollbars match the dark theme.

### Modified Capabilities

- (none)

## Impact

- MorphNotes: `AppDrawer.js`, `appRouter.js`, `DataImport.js`, `tranClient.js`, generic-data import handler/route, `CaseTasks.js` + create/update validation in `tran_case_tasks.go`, `Timelines.js`, `BigNotes.js` (and related CSS).
- MorphUtils: `morph-utils/frontend/src/App.tsx` Sign out; Event Logs (`formx` Layout logout), Content Maker (`composerx` Out), Data Access (`SharpReport` UserMenu Sign out), Project (`morph-engi` AppLayout Sign out).
- Project preview: `ProjectDocumentPanel.svelte` overflow/scrollbar styles.
- MorphUtils left-nav ids (`sheetx`, `/survey-bot`, `composerx`, `datax`, `projects`) stay. Product UIs stay dark-only. Session still Morph JWT; operators leave by closing the tab or using Morph AI Sign out.
