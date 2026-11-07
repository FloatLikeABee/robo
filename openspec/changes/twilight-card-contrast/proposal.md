# Proposal

## Why

Twilight glass put cards, chat bubbles, and card titles on fills that sit next to the page color `#2c2640`. The shapes are still there, but they do not read as separate objects.

## What Changes

- Lift card surfaces so they are clearly lighter than the page, with a stronger edge, without leaving twilight purple.
- Apply that lift to Dorph AI chat bubbles and titled cards, DorphNotes papers, the Content shell panels, Event Logs cards, the Skills modal, and DorphTools papers.
- Card titles use brighter ink than the body text on the same card.
- Leave the page floor, user-bubble purple, and public Event Logs form background as they are.

## Capabilities

### New Capabilities

- `twilight-card-contrast`: Cards and their titles stay visually separate from the twilight page.

### Modified Capabilities

- None.

## Impact

- `dorph/frontend` chat CSS and the DorphNotes theme paper color.
- Content (`composerx`) elevated surfaces, Event Logs operator cards, Skills modal, and DorphTools (`bk`) paper.
- No new theme, no API, no change to generated HTML palettes.
