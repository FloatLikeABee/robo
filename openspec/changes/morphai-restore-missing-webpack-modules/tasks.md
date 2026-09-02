## 1. Restore resolvable modules

- [x] 1.1 Rename chat workspace `.jsx` files to `.js` (`AiToolsWorkspaceDrawer`, `AgentWorkspace`, `AgentFilesTab`) without changing exports
- [x] 1.2 Rename `ExtractJsonFromTextDialog.jsx` to `.js` next to `AdminDataGrid.js`
- [x] 1.3 Confirm `agentContext.js`, `filesWorkspaceStore.js`, and `appliedAssistantChannel.js` sit at the import paths `SkoolAiChat.js` uses

## 2. Conflicts and compile

- [x] 2.1 Confirm Morph frontend sources webpack compiles have no leftover conflict markers
- [x] 2.2 Run Morph frontend production compile and confirm the six `Can't resolve` overlay errors are gone
