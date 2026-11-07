# Design

## Context

See proposal.md. The publish workspace is `ComposePublishPanel.svelte`. The source column is `.publish-source-box`. Its textarea is locked at `5.6rem`, and `.publish-source-actions` is a horizontal row. `.btn-secondary` paints `--color-bg-elevated`, the same fill as the panel, and a border of `--color-border-subtle`, the same pale stroke as the panel edge. Global `button` styles in `app.css` stay as they are so the rest of Content Maker is untouched.

## Goals / Non-Goals

**Goals:**

- One source column that spends leftover height on the text area.
- One vertical action stack, with From data last and filled as the primary.
- The same filled-button treatment on the other publish actions.

**Non-Goals:**

- Restyling Content Maker outside the publish workspace and its records list.
- Changing generate, upload, or publish behavior.
- A new color outside the existing twilight tokens.

## Decisions

### 1. The text area takes the leftover height

**Choice:** `.publish-source-box` is a column that stretches to its grid cell (`height: 100%`, `min-height: 0`). The textarea is `flex: 1` with `min-height: 12rem` and no max height. The action stack comes after it, so the field grows and the buttons sit at the bottom of the column. Column gap is 16px.

Rejected: raising the fixed `5.6rem` to a larger fixed height. A fixed box still leaves a hole when the column is tall, and overflows when it is short.

### 2. Three stacked actions, data last

**Choice:** Move Upload out of the label row into the stack. Order is Upload, From text, From data. Each is full width. From data uses the existing `.primary-action` fill (`--color-primary`, `#8b5cf6`) and slightly taller padding. The other two use the secondary fill below. Idle labels are Upload, From text, and From data. Upload still shows Summarizing… while files process. The placeholder no longer says the upload control is above the field.

Rejected: leaving Upload in the header and stacking only the two generate buttons. Upload is the same outline problem, and the column reads as one set of actions.

### 3. Filled buttons, scoped to this module

**Choice:** In this panel and in `ComposePublishRecordsPanel.svelte`, `.btn-secondary` uses `--color-surface` (`#3d3148`), which is distinct from the panel `#322a4a`, with a transparent border and `#f4f0fa` text. Hover brightens the fill. It does not use `--color-border-subtle`. Header Publish already uses `.primary-action` and stays that way. Save becomes Save. Apply HTML to preview becomes Apply. Close, Clear, Clear processed files, and Enlarge stay `.btn-ghost` with no border. The round AI send control keeps its shape and switches from the leftover blue gradient to the theme accent fill.

Rejected: changing the global `button` rule. That would restyle every Content Maker screen. Rejected: a thicker accent outline. The stroke is what makes these look like labels.

### 4. Records list uses the same secondary fill

**Choice:** Refresh and any other `.btn-secondary` in the records panel get the same fill. Remove and Delete stay quiet text actions.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| A short window clips the stack | Textarea min-height is 12rem and the column already scrolls with the page grid; the stack stays after the field |
| Secondary `#3d3148` is close to the panel | It is the warm tile, not the panel fill, and it has no pale stroke; primary From data is the one that must pop |
| Shorter labels hide the summarize behavior | The placeholder still says files are summarized |

## Migration Plan

Style and label change only. Rollback is reverting the two Svelte files. No stored data changes.
