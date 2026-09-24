## 1. MorphNotes / Morph AI

- [x] 1.1 Remove `window.alert` / `window.confirm` fallback from `ConfirmDialog.jsx` `useConfirm` (dev warn only)
- [x] 1.2 Migrate `SkillsModal.js` delete flow to `useConfirm({ danger: true })`

## 2. Project (morph-engi)

- [x] 2.1 Add `lib/confirmDialog.ts` + root `ConfirmDialog.svelte` (confirm + alert)
- [x] 2.2 Replace `confirm()` in `App.svelte` and `ProjectDocumentPanel.svelte`

## 3. Content Maker (composerx)

- [x] 3.1 Add shared confirm/alert dialog helper and mount in `App.svelte`
- [x] 3.2 Migrate `ComposePublishRecordsPanel.svelte` delete flows off native `confirm()`

## 4. Data Access (SharpReport)

- [x] 4.1 Add shared confirm dialog helper and mount in app shell
- [x] 4.2 Migrate `confirm()` in data-tables and docs pages

## 5. AI tools (bk)

- [x] 5.1 Add `ConfirmProvider` (or equivalent) at bk app root
- [x] 5.2 Replace `window.confirm` / `alert` in `AssistantManager.js`, `AgentManager.js`, `GraphicalFlowEditor.js`

## 6. Event Logs (formx)

- [x] 6.1 Add `alert()` to `ConfirmContext.tsx` (OK-only modal) for error notices
- [x] 6.2 Grep formx `src/` for any remaining native dialogs and migrate if found

## 7. Verification

- [x] 7.1 Grep supported product `src/` trees: no `window.alert`, `window.confirm`, `window.prompt`, or bare `confirm(` / `alert(` except in dialog helper implementations
- [x] 7.2 Smoke: delete confirm in Project, Content Maker draft delete, MorphNotes Skills delete, AI tools assistant delete — all in-app modals
