# Design

## Context

See proposal.md. The twilight page is `#2c2640`. Chat panels use `rgba(53, 46, 74, 0.42)`, assistant bubbles use `rgba(255, 255, 255, 0.16)`, and DorphNotes paper is `#352e4a`. Those sit on the same hue, so cards disappear into the floor.

## Goals / Non-Goals

**Goals:**

- One lighter card fill and one stronger edge, copied into each twilight surface.
- Titles brighter than secondary text on that card.

**Non-Goals:**

- A new theme or a lighter page.
- Restyling public Event Logs forms.
- Changing generated HTML document colors.

## Decisions

### 1. Lift the card, keep the floor

| Approach | Pros | Cons |
|----------|------|------|
| **Lighter solid card on the same page (chosen)** | The object reads immediately | Slightly less “frosted” |
| Darken the page | Also separates | Throws away the current floor |
| Border only | Small diff | The fill still matches, so titles still vanish |

**Choice:** card fill `#5c5182`, edge `rgba(232, 222, 255, 0.55)`, top inset highlight. Page stays `#2c2640`. Comment points at `App.css` twilight tokens. User bubbles stay `rgba(109, 40, 217, 0.92)`.

### 2. Titles, not every line

**Choice:** card titles and chat headings use `#f4f0fa` at 700, or `#a78bfa` when the title is a label. Body and hints stay `#c4b8d4`. Do not recolor every paragraph.

### 3. Same lift in each app

**Choice:** set the card fill in Dorph AI bubble and panel rules, DorphNotes `theme.js` paper, Content `--color-bg-elevated` / `--color-surface`, Event Logs card classes, the Skills modal surface, and DorphTools paper. No shared package.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Glass blur hides the new fill | Use a solid fill; keep blur only as a shadow, not the color |
| User and assistant bubbles converge | Leave the user gradient alone |
| MUI paper lightens every dialog | That is the point for cards; buttons keep the purple accent |

## Migration Plan

CSS and theme tokens only. Rollback is restoring the previous fills.

## Open Questions

None.
