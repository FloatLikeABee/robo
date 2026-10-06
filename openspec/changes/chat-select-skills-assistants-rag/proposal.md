## Why

Choosing skills in Morph AI chat does not reliably steer the reply. Assistants and RAG collections can only be applied from the AI tools panels, so the operator leaves the composer to attach the thing they want the next message to use.

## What Changes

- A skill checked in the chat picker MUST be followed on the next send. Its instruction body goes into the system prompt for that turn. A selected skill MUST NOT be dropped because it was missing from a second catalog.
- AI tools assistants are chosen from a chat picker with the same pattern as skills. One assistant at a time. The Apply button in the assistant panel is removed. The chosen assistant’s system prompt and its own RAG collections still apply to the send.
- RAG collections are chosen from a third chat picker, same pattern as skills, and can be more than one. The next send retrieves those collections and includes the snippets. This is independent of which assistant is selected.
- Creating and editing assistants and RAG collections stays in the AI tools modal. The chat pickers only select.

## Capabilities

### New Capabilities

- `chat-selected-skills`: Checked skills are injected into the chat turn and followed.
- `chat-assistant-picker`: The composer selects one AI tools assistant; the panel Apply button is gone.
- `chat-rag-picker`: The composer selects RAG collections and the send retrieves them.

### Modified Capabilities

- (none)

## Impact

- Morph AI composer: `SkoolAiChat.js` and the skills picker UI.
- Chat handler: `handlers/chat.go`, `handlers/skills.go`, `handlers/bk_assistants.go`. New list route for RAG collections if `/api/ai-agents` is the assistant list and RAG has no Morph route yet.
- AI tools assistant panel: remove Apply in `bk/frontend` `AssistantManager.js`.
- No new database. No Files workspace. Data Access stays out.
