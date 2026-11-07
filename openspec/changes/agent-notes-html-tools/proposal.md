# Proposal

## Why

Agent Notes opens on an empty detail until the operator picks a row. DorphNotes HTML still uses an older purple page than the bento tiles. DorphTools Documents still offers Graphic documents, and the DORPHTOOLS title is a clipped gradient that disappears into the bar.

## What Changes

- When Agent Notes loads a non-empty list, select the first note. Going back to the list does not immediately reselect it.
- New HTML from DorphNotes (research, timelines, big notes, case tasks, and Contents & Data publish) uses the current bento colors: page `#221c34`, tile `#322a4a`, accents `#8b5cf6`, `#a78bfa`, `#f97316`. Stored pages stay until the next generate.
- Remove the Graphic documents tab from DorphTools Documents. Articles and ScholarForge stay. The graphic-document API stays.
- Paint the DORPHTOOLS title in solid light ink `#f4f0fa`.

## Capabilities

### New Capabilities

- `agent-notes-first-selection`: Agent Notes opens the first note when the list loads.
- `dorphtools-documents-trim`: Documents has no Graphic documents module, and the title is readable.

### Modified Capabilities

- `dorph-generated-document-theme`: New DorphNotes HTML uses the current bento page and tile colors.

## Impact

- `dorph/frontend` Agent Notes modal and publish HTML.
- `dorph/internal/htmldoc` and the prompts and tests that lock the old page hex.
- `bk/frontend` Documents tabs and header title.
- No new publisher and no theme picker.
