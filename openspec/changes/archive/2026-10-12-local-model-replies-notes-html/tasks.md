## 1. Reply shape

- [x] 1.1 Replace the visual-first sentence with `ReplyShapeInstructions` and keep `VisualFirstInstructions` as that same text. Add `ModelIdentityInstructions` and a client helper that reports the resolved provider and model even when the key is missing.
- [x] 1.2 Prefix Morph chat and `GenerateChatResponseWithAgent` with the reply shape and the identity line. Keep the tool catalog. State that label fields and JSON tool calls are only for a real data call.
- [x] 1.3 Drop Graphs from always-on skills. Use code defaults for Research and Design, and replace a stored always-on body that still says diagram-first, visual-first, or short caption. Give Graphs its own opt-in sentence.
- [x] 1.4 Point note and TODO text assist, case-task drafts, and research compose at the reply shape, and remove lines that invite a mermaid diagram on every body.
- [x] 1.5 Update prompt tests so they require the reply-shape rule and the identity line, and reject a diagram-first or invented-version instruction.

## 2. Workspace pane

- [x] 2.1 Restore `initialWorkspaceOpen` so a saved choice wins and an unset phone starts closed. Restore the header show/hide control, the collapsed class, and the collapsed grid rules, including the phone band.
- [x] 2.2 Put the pane tabs back to English Notes & TODOs, then Context & Knowledge. Leave the Agent notes modal and the Files tab alone.
- [x] 2.3 Restore the workspace-open unit tests for phone, desktop, and a saved choice.

## 3. Document theme

- [x] 3.1 Switch `htmldoc` to the Arial stack, dark-only colors, and at least three accents. Use those accents on links, headings, and callouts or charts.
- [x] 3.2 Make big-note HTML always dark. Update the case-task preview and Project HTML to the same font and accents. Update HTML generation prompts that still ask for one accent or a light page.
- [x] 3.3 Set the chat markdown font to the same sans stack without changing the dark chat surface.
- [x] 3.4 Add or update tests that the document CSS includes Arial, stays dark, and contains at least three accent colors.

## 4. Checks

- [x] 4.1 Run the Go tests for `pkg/morphai` and the touched Morph handlers, and the Morph frontend unit tests that cover the pane.
