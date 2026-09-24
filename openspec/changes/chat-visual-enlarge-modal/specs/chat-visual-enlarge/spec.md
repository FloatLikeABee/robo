## Purpose

Lets people enlarge diagrams, charts, pixel art, and images that appear in assistant chat so they can read them without leaving the conversation.

## ADDED Requirements

### Requirement: Click chat visual opens enlarge modal

When a mermaid diagram, chart, pixel-art grid, generated picture, or markdown image is shown in Morph AI, Event Logs, or Content Maker chat, the user MUST be able to activate it (click or keyboard) to open a dark in-app modal that shows an enlarged copy of that visual. The conversation MUST remain in place behind the modal. Invalid mermaid fallback text MUST NOT open the modal.

#### Scenario: Enlarge mermaid in Morph AI

- **WHEN** an assistant message in Morph AI contains a rendered mermaid diagram and the user clicks it
- **THEN** a dark modal opens with a larger copy of that diagram

#### Scenario: Enlarge pixel grid and generated image

- **WHEN** a Morph AI message shows a rendered pixel grid or a generated image and the user clicks it
- **THEN** a dark modal opens with a larger copy of that visual

#### Scenario: Enlarge mermaid in other in-stack chats

- **WHEN** Event Logs or Content Maker assistant markdown shows a rendered mermaid diagram and the user clicks it
- **THEN** a dark modal opens with a larger copy of that diagram

#### Scenario: Fallback source is not enlargable

- **WHEN** mermaid failed to render and the chat shows the source text instead
- **THEN** clicking that source text does not open the enlarge modal

### Requirement: Dismiss enlarge modal

The enlarge modal MUST be dismissible with Escape, a close control, and a click on the dimmed backdrop. While open, focus MUST stay in the modal. After dismiss, focus MUST return to the chat. The modal MUST use the product dark theme (no light/dark switch, no browser-native dialog).

#### Scenario: Escape and backdrop close

- **WHEN** the enlarge modal is open and the user presses Escape or clicks the backdrop
- **THEN** the modal closes and the chat is usable again

#### Scenario: Close control

- **WHEN** the enlarge modal is open and the user activates the close control
- **THEN** the modal closes
