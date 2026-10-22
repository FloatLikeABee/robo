## 1. Shared apply-assistant channel

- [x] 1.1 Add a small helper in Dorph frontend for apply/dismiss/state `postMessage` + `BroadcastChannel('dorph-ai-applied-assistant')`, origin checks, and `sessionStorage` key `dorph-ai:applied-assistant`
- [x] 1.2 Add a matching helper in AI Tools frontend to post apply/dismiss/request-state to `window.parent` (when framed) and the same BroadcastChannel

## 2. AI Tools Assistants list

- [x] 2.1 Replace the Run icon, Run dialog, and `runAssistant` call on `AssistantManager.js` with Apply (when not applied) and Dismiss (when this card is applied)
- [x] 2.2 On Assistants mount, request applied-assistant state; reflect Dorph’s reply so only the applied card shows Dismiss
- [x] 2.3 If Apply is used outside Dorph (not framed and Dorph does not receive it), show a toast that Dorph chat is required; do not mark the card applied

## 3. Dorph chat apply / dismiss

- [x] 3.1 In `SkoolAiChat.js`, restore applied assistant from `sessionStorage` and listen for apply/dismiss (postMessage from the AI Tools iframe origin + BroadcastChannel)
- [x] 3.2 Reply to `request-applied-assistant` and broadcast state after apply/dismiss so the iframe stays in sync
- [x] 3.3 Show an applied-assistant chip (name + Dismiss) in Dorph chat; Dismiss clears state and storage
- [x] 3.4 Include `agent_id` as `bk:<id>` on JSON and multipart `/api/chat` when an assistant is applied; omit it after dismiss
- [x] 3.5 If chat returns unknown/unavailable AI tools assistant, dismiss the applied assistant and surface the error once

## 4. Drawer wiring

- [x] 4.1 Ensure `AiToolsWorkspaceDrawer` iframe can receive Dorph state replies (`allow` is not required; confirm origin of `contentWindow.postMessage` uses the bk origin from `REACT_APP_BK_URL`)

## 5. Verify

- [x] 5.1 From Dorph AI → AI tools → Assistants: Apply an assistant, confirm the chip appears, send a message and confirm `agent_id` is `bk:<id>`
- [x] 5.2 Apply a second assistant: first is replaced; Dismiss from the card and from the Dorph chip both clear `agent_id`
- [x] 5.3 Confirm the Assistants page has no Run control or Run dialog
