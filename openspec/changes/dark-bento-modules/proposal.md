# Proposal

## Why

The sample set in Documents is a light-mode card system: bento grids, an 8px spacing scale, icons with labels, skeletons, and short text with Read more. Our modules are one flat twilight fill, so those rules have nowhere to land, and copying the samples’ cream page would wash the product out.

## What Changes

- Give every operator module a dark bento: a deeper page, mid-dark tiles of a few sizes, 16px gaps, 16px corners. Not a white or cream page, and not neon pink tiles.
- Follow the sample rules that fit an app: 8px spacing, icon plus label on navigation, at most two nav levels, skeleton placeholders instead of a lone spinner, long card text clamped with Read more, and a side-by-side column when a preview sits next to text.
- Apply it to Dorph AI, DorphNotes, Content, Event Logs (operator), and DorphTools.
- Leave public Event Logs forms and already published HTML pages alone.

## Capabilities

### New Capabilities

- `dark-bento-layout`: Shared dark bento structure and the sample layout rules across operator modules.

### Modified Capabilities

- None.

## Impact

- Shell and primary screens in `dorph/frontend`, `composerx/frontend`, `formx/frontend` (operator only), and `bk/frontend`.
- No new dependency, no API, no change to generated document colors.
