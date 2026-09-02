## Why

Morph AI Skills’ left upload pane uses a tall Instructions field, so the form overflows and a vertical scrollbar appears. Operators should see the whole left pane (name, description, instructions, actions) without scrolling the pane.

## What Changes

- Shorten the Instructions control so the left Skills pane fits in the dialog height.
- The left pane MUST NOT show a vertical scrollbar for the form. The right catalog may still scroll.
- Keep the split dark Skills dialog, markdown upload, Improve with AI, and Morph JWT. No under-title lede. Dark-only.

## Capabilities

### New Capabilities

- `morphai-skills-left-pane-fit`: Left Skills upload pane, including a compact Instructions field, fits without a pane scrollbar.

### Modified Capabilities

- (none — `openspec/specs/` has no archived `morphai-skills-editor` baseline)

## Impact

- Morph AI frontend only: `SkillsModal.js` and Skills styles in `App.css`.
- No skills API, Neo4j sync, chat picker, MorphUtils, or MorphNotes.
