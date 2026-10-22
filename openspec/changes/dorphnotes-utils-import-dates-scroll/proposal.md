## Why

DorphNotes still exposes Settings File import that operators no longer use, Task start/end dates are optional so records can be saved without a window, and HTML/Markdown previews in DorphNotes and DorphUtils Project show nested or unthemed scrollbars. DorphUtils also still shows Sign out / Out even though session is Dorph JWT and operators do not need a second logout.

## What Changes

- Remove DorphNotes **Settings → File import** entirely: nav item, route, page, and the dedicated import API used only by that page. Generic data CRUD and its extract-for-review flow stay.
- **BREAKING:** `POST /api/tran/generic-data/import` is removed. Settings `/admin/configuration/file-import` (and `/configuration/data-import` alias) go away.
- DorphNotes **Tasks** create/edit: **Start** and **End** dates are required. New drafts default both to the current local calendar day.
- DorphNotes **Timelines** HTML view: one vertical scrollbar only (the HTML content). Outer panel does not scroll. Inner scrollbar is dark to match the theme.
- DorphNotes **Big Notes** preview (and HTML tab if it has the same double-scroll): same rule — outer overflow gone, inner content scroll, dark scrollbar.
- DorphUtils shell and every module shown inside it (Event Logs, Content Maker, Data Access, Project): remove Logout, Sign out, and Out controls. Dorph AI Sign out on the Dorph home chat is unchanged.
- DorphUtils **Project** Markdown and HTML preview scrollbars use the dark theme instead of default light browser chrome.

## Capabilities

### New Capabilities

- `remove-dorphnotes-file-import`: Drop Settings File import UI and its dedicated backend import endpoint.
- `case-task-required-dates`: Task start and end dates required; new forms default to today.
- `dorphnotes-preview-scroll`: Timelines HTML and Big Notes preview use a single inner, dark-themed vertical scrollbar.
- `dorphutils-no-logout`: DorphUtils and its modules have no logout/sign-out/out actions.
- `dorphutils-project-themed-scroll`: Project Markdown/HTML preview scrollbars match the dark theme.

### Modified Capabilities

- (none)

## Impact

- DorphNotes: `AppDrawer.js`, `appRouter.js`, `DataImport.js`, `tranClient.js`, generic-data import handler/route, `CaseTasks.js` + create/update validation in `tran_case_tasks.go`, `Timelines.js`, `BigNotes.js` (and related CSS).
- DorphUtils: `dorph-utils/frontend/src/App.tsx` Sign out; Event Logs (`formx` Layout logout), Content Maker (`composerx` Out), Data Access (`SharpReport` UserMenu Sign out), Project (`dorph-engi` AppLayout Sign out).
- Project preview: `ProjectDocumentPanel.svelte` overflow/scrollbar styles.
- DorphUtils left-nav ids (`sheetx`, `/survey-bot`, `composerx`, `datax`, `projects`) stay. Product UIs stay dark-only. Session still Dorph JWT; operators leave by closing the tab or using Dorph AI Sign out.
