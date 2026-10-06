## 1. Selected skills

- [x] 1.1 Load instruction bodies for every skill id the chat picker sends, from the same source as the picker
- [x] 1.2 Put those bodies in the system prompt as skills to follow for that reply, and do not drop a selected id that the picker listed
- [x] 1.3 Add a check that a selected id is included even when it is absent from the enabled-catalog list

## 2. Assistant picker

- [x] 2.1 Add a composer assistant picker, same pattern as skills, listing `GET /api/ai-agents`, single choice plus clear
- [x] 2.2 Send the chosen assistant id on the next message and keep using that assistant’s system prompt
- [x] 2.3 Remove the Apply button from the AI tools assistant panel

## 3. RAG picker

- [x] 3.1 List RAG collection names from the AI tools API through a Morph route
- [x] 3.2 Add a composer RAG picker, same pattern as skills, with multiple checks and clear
- [x] 3.3 On send, retrieve the checked collections and include the snippets, including when no assistant is selected
- [x] 3.4 When an assistant with its own collections is also selected, retrieve both sets and keep the existing snippet caps

## 4. Verification

- [x] 4.1 Run the selected-skill check and a check that a chat-selected RAG collection is queried without an assistant id
