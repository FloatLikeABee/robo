## Why

MorphNotes **Notes & TODOs** and Morph AI **Context & Knowledge** sit on a neutral dark grey (`#1e1f24` and similar) while the rest of the chrome is dark blue (`--chat-surface` / `--chat-panel`). They look like a leftover palette, not the current theme.

## What Changes

- Paint those panels with the same dark-blue tokens as Morph AI / MorphNotes chrome (page, sidebar, chat). Stay dark-only; no theme switch.
- Context & Knowledge (workspace tab and drawer) MUST use `--chat-surface` / `--chat-panel` / `--chat-border`, not undefined `--chat-bg-elevated` falling back to `#1e1f24`.
- MorphNotes Notes & TODOs (header panel) MUST match that blue surface, not a flatter grey paper or leftover purple header wash.
- Morph AI Notes & TODOs drawer (same hybrid-drawer shell) MUST match, so notes opened from chat are not grey next to a blue workspace.

## Capabilities

### New Capabilities

- `notes-knowledge-dark-blue`: Notes & TODOs and Context & Knowledge panels use the product dark-blue theme.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- Morph frontend CSS: `App.css` `.hybrid-drawer` and related fallbacks (`--chat-bg-elevated`, `#1e1f24`, `#3a3c48`).
- MorphNotes: `adminRightPanelStyle.js` Notes & TODOs paper/header.
- Shared Notes & TODOs MUI surfaces only if they still read as grey on the blue panel.
- No API or copy changes. Do not add a light/dark switch.
