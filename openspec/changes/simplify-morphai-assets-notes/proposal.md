## Why

Morph AI and Morph Data still surface leftover onboarding and navigation from earlier modules: a verbose empty-chat guide, mixed Morph + AI-tools assistants, a combined People/Assets Resources module, and a free-text Big Notes theme field. Users want a quieter Morph AI home, assistants that only come from AI tools, Assets as the sole resource surface, and a simple dark/light theme choice for new Big Notes.

## What Changes

- **Morph AI empty chat**: Remove the "Here's what you can do:" copy and all quick-link / tip list items. Keep the logo (and brand title) only on the initial empty state.
- **Assistants list**: The sidebar Assistants section SHALL list only assistants created in AI tools (BK), not Morph built-in / specialist agents.
- **Resources → Assets**: Remove the People module and People tab. Keep Assets only. Rename nav and routes so the menu label is **Assets** (not Resources). Redirect or drop legacy People/Resources paths as needed.
- **Big Notes create theme**: Replace free-text theme input with a two-option selector: **dark** and **light** only.

## Capabilities

### New Capabilities
- `morphai-empty-welcome`: Empty Morph AI chat state shows logo/brand only; no tip list or quick links.
- `morphai-bk-assistants`: Assistants picker exposes only AI tools (BK) assistants.
- `morph-data-assets-nav`: Morph Data navigation and module surface are Assets-only (People/Resources combined module removed).
- `big-notes-theme-select`: Big Notes create flow offers only dark and light theme options.

### Modified Capabilities
- (none — main `openspec/specs/` has no archived baselines; prior Resources behavior lived only in change deltas)

## Impact

- **Frontend (Morph AI)**: `morph/frontend/src/SkoolAiChat.js` empty-state markup; assistants list filtering / default selection.
- **Backend (agents API)**: `morph/handlers/morph_ai_agents.go` (and related) may stop returning Morph built-ins for the chat picker, or frontend filters to `source === 'bk'` — prefer API/list contract that matches UI.
- **Frontend (Morph Data)**: `AppDrawer.js`, `appRouter.js`, `Resources.js`, `People.js`, `DataImport.js`, Vehicles/Assets pages — nav label, routes, tabs, import entity labels.
- **Frontend (Big Notes)**: `morph/frontend/src/pages/admin/BigNotes.js` create dialog theme control; ensure create payload sends `dark` or `light`.
- **BREAKING (UX)**: People UI and Resources combined module go away; users manage people-as-assets if needed via Assets. Morph specialist agents no longer appear in Assistants.
