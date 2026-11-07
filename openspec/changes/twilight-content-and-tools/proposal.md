# Proposal

## Why

Contents & Data’s Publish pane still introduces itself as Content Maker, in blue chrome, while Dorph AI is twilight glass. Generated pages, Event Logs, the Skills modal, and DorphTools use other palettes (cyan/mint, slate, sci-fi blue), so the same product looks like several apps.

## What Changes

- Rename the Publish module’s visible name from **Content Maker** to **Content**. The drawer stays **Contents & Data**. The `composerx` folder, embed id, and routes stay.
- Restyle that Content shell to the same purple glass as Dorph AI (translucent panels, blur, light ink, purple accent, orange only as the existing highlight).
- When the HTML preview has real page content, the operator can open it in a large modal. The empty starter preview stays inline.
- Rename **Generate from sources** to **Generate from Data**. It opens a picker of existing generic-data records. PDF (and other prose files) become an article page. CSV, Excel, and JSON become a table plus a short dashboard. Paste-and-generate stays.
- Keep the vendored html-anything skill (structure, no invented facts, self-contained HTML, contrast). Shift its page palette from cyan/mint/`#070b14` to the twilight set, still three hex accents, still under 512 tokens. Grill write stays plain markdown and does not receive the skill.
- Hard-coded HTML builders (generic-data publish, research, timelines, big notes, case tasks, project docs, preview wrappers) use that same twilight set for new output. Stored pages are not migrated.
- Event Logs operator chrome uses twilight glass. Public respondent forms stay solid and readable, with the same ink and accent, without a frosted mesh.
- Skills modal and the DorphTools app (`bk`) use twilight glass. No theme picker.

## Capabilities

### New Capabilities

- `content-publish-surface`: Content naming, twilight glass shell, enlargeable HTML preview, and generate-from-data-record.
- `event-logs-operator-theme`: Event Logs operator chrome in twilight glass; public forms stay readable.
- `dorph-tools-theme`: Skills modal and DorphTools (`bk`) chrome in twilight glass.

### Modified Capabilities

- `html-anything-base-skill`: Compact overlay hex accents become the twilight set (purple, violet, orange). License, budget, grill exclusion, and hex-only rule stay.
- `dorph-generated-document-theme`: Generated documents and their prompts use that twilight set. Dark sans, three accents, no light page, and big-note rules stay.

## Impact

- `composerx/frontend` shell, messages, and Compose publish panel.
- `dorph/frontend` Skills modal CSS, generic-data publish HTML, preview wrapper.
- `skills/vendor/html-anything/condensed.md` and the copies Dorph injects; tests that lock the old cyan hex.
- Event Logs UI (`formx/frontend`), DorphTools UI (`bk/frontend/src/theme.js`).
- Existing HTML builders in `dorph/handlers` and `dorph-engi` that stamp document colors.
- No new publish API, no new database, no rename of `composerx`.
