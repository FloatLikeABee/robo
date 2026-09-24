## Why

Destructive actions and error notices still call `window.alert`, `window.confirm`, or bare `confirm()` in several products. Those browser chrome dialogs break the dark in-app experience, block the wrong layer, and cannot be styled. MorphNotes and Event Logs already have in-app confirm modals; the rest of the stack should match.

## What Changes

- Replace every **browser-native** `alert`, `confirm`, and `prompt` in supported product UIs with **in-app modals** (confirm/cancel for destructive actions; OK-only for notices/errors).
- **MorphNotes / Morph AI** (`morph/frontend`): remove remaining `window.confirm` call sites; stop `useConfirm` from falling back to `window.alert` / `window.confirm` when the provider is mounted (keep a dev-only console warning if miswired).
- **Project** (`morph-engi`), **Content Maker** (`composerx`), **Data Access** (`SharpReport`), **AI tools** (`bk`): add a small shared confirm/alert dialog pattern and migrate delete/error flows off native dialogs.
- **Event Logs** (`formx`): already uses `ConfirmProvider`; add `alert()` for error notices if any native alerts remain; no new UX pattern.
- Modals MUST use each product's dark theme, trap focus, support Escape to dismiss (cancel), and label destructive confirms clearly.
- Out of scope: OS push notifications (`Notification` API), toast/snackbar success messages, and backend error responses.

## Capabilities

### New Capabilities

- `in-app-dialogs`: User-facing confirms and alerts render as in-app modals across supported Morph stack products, not browser-native dialogs.

### Modified Capabilities

- (none)

## Impact

- `morph/frontend/src/components/ConfirmDialog.jsx`, `SkillsModal.js`
- `morph-engi/frontend` (new dialog helper + `App.svelte`, `ProjectDocumentPanel.svelte`)
- `composerx/frontend` (`ComposePublishRecordsPanel.svelte` + shared dialog)
- `SharpReport/frontend` (data-tables, docs pages)
- `bk/frontend` (`AssistantManager.js`, `AgentManager.js`, `GraphicalFlowEditor.js`)
- `formx/frontend/src/context/ConfirmContext.tsx` (optional `alert` parity)
- No API or database changes.
