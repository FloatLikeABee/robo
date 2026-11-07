# Design

## Context

See proposal.md. Samples live in `/Users/floatinbee/Documents/samples web ui`. They are light-mode posts: bento compartments, skeleton instead of a spinner, icons with labels, an 8px spacing scale (4, 8, 16, 32, 40), columns instead of wrapped text, nav no deeper than two levels, and Read more on long copy. The attached bento reference uses a cream page around a dark tile grid. The product is already twilight and dark-only. The last card pass used one fill, `#5c5182`, which is even and a bit light.

## Goals / Non-Goals

**Goals:**

- One dark tile set and an 8px rhythm, copied into each operator shell.
- Bento structure on each module’s primary screen, not a rewrite of every dialog.

**Non-Goals:**

- A light or cream theme.
- Public Event Logs forms.
- Generated HTML page colors.
- A third navigation level or a new component library.

## Decisions

### 1. Dark tiles, not the sample’s light page

| Approach | Pros | Cons |
|----------|------|------|
| **Deeper floor, mid-dark tiles (chosen)** | Matches the bento photo’s inner grid and stays on-brand | Less “frosted glass” |
| Copy the cream page and white cards | Matches the posts exactly | User said it must not be too light |
| Keep `#5c5182` everywhere | Already shipped | One slab, not a bento |

**Choice:** page `#221c34`. Tiles `#322a4a` and `#3d3148`. One accent tile `#5a3d62` for the feature cell. Text `#f4f0fa`, muted `#c4b8d4`. Gap 16, radius 16, padding 16 or 24. Replace `#5c5182` card fills with these tiles.

### 2. Where the bento goes

**Choice:** each module’s main view becomes a grid. Dorph AI: transcript in a wide tile, workspace in a side tile. DorphNotes: the open module’s header is a wide tile and the list is smaller tiles or one tile that holds the existing grid. Content: preview and source are the two columns already required. Event Logs and DorphTools: their cards use the tile colors and the same gap and radius. Do not turn a data grid into one card per row.

### 3. Sample rules, scoped

**Choice:** navigation items that are already icon-plus-label stay that way. Do not add labels to every toolbar icon that already has a tooltip. Skeletons replace the first-load spinner of a list, not the spinner on a button. Read more clamps card descriptions and analysis excerpts, not the chat transcript the operator is reading in full. Wide layouts use two columns when a preview and text are both on screen.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| “All modules” becomes a rewrite | Primary screen per module only |
| Skeletons on every request flicker | First load only, when the list is empty |
| Darker tiles fail contrast | Text stays `#f4f0fa` on `#322a4a` |

## Migration Plan

CSS and layout on the primary screens. Rollback restores the previous fills. No data migration.

## Open Questions

None.
