## Why

On a phone, Morph AI chat is the product. The shell already fits the viewport, but the transcript can still scroll sideways, composer controls are under 44px, and the soft keyboard covers the input because the chat column is locked to `100dvh` with overflow hidden. Issue #98 is the conversation itself.

## What Changes

- Keep ordinary message prose, the empty welcome, and send/network errors inside the chat column at about 390px, with the transcript scrolling vertically only.
- Keep the composer input and send control in the visible viewport while the soft keyboard is open, including side safe-area insets on the composer.
- Size composer actions (send, attach, skills, cancel, clear) to at least 44×44 CSS pixels on a phone.
- Keep code, Mermaid, and pixel-art blocks inside the column, with the existing enlarge overlay as the path for a closer look.

## Capabilities

### New Capabilities

- `morph-phone-chat`: Phone layout of the Morph AI transcript, empty and error states, composer, keyboard, and in-column visual blocks.

### Modified Capabilities

- None. Shell viewport and header/drawer requirements stay as archived. This change does not alter those requirements.

## Impact

- Morph AI SPA only: `morph/frontend/public/index.html` viewport token, `morph/frontend/src/App.css`, `morph/frontend/src/SkoolAiChat.js`, and a small keyboard-inset helper under `morph/frontend/src/lib/`.
- No chat protocol, Go API, login rewrite, header/drawer redesign, or satellite apps.
