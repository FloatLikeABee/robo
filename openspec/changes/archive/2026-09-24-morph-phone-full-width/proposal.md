# Proposal

## Why

On a phone, Morph AI chat is either crushed into an ~80px left strip or the Notes/Knowledge panel stays open and eats the conversation. Ge's iPhone screenshots (issue #126) show both, and the chat is not usable until both are fixed.

## What Changes

- At the existing phone shell breakpoint (`max-width: 768px`), a collapsed agent workspace uses one flexible column so the header, transcript, and composer fill the viewport width. Words wrap as normal prose.
- On a phone, the Notes & TODOs / Context & Knowledge panel starts closed. The header workspace control still opens it. An explicit saved open or closed choice is kept. Desktop still starts with the side panel open.
- On a phone, the composer Notes / Knowledge include chips are not shown. The panel tabs are the one labeled path to those surfaces. Include-on-send stays at its current default (both on).

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `morph-phone-chat`: The phone chat column must fill the layout viewport when the workspace is collapsed; the workspace must not occupy a bottom band by default; composer Notes/Knowledge chips must not duplicate the panel tabs.

## Impact

- `morph/frontend/src/App.css` phone agent-grid rules and the include bar.
- `morph/frontend/src/components/chat/AgentWorkspace.js` initial open state, plus `SkoolAiChat.js` call site.
- Frontend unit tests only. No API, MorphUtils, or desktop layout change. Keyboard-inset and safe-area rules from the phone composer work stay.
