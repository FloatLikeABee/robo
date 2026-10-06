## Why

Generated timelines, stories, research, and case-task pages render as one dark column of headings and bullets. A phase list with real counts still looks like a plain note, so the document does not show the shape of the work.

## What Changes

- Generated HTML keeps markdown as the source and lays it out as a document: a lede, phase cards, a vertical rail for milestone lists, callouts, and a chart when the source has quantities.
- The model does not emit raw HTML. It may add a stat block and a short callout. The renderer turns those, plus `##` sections and lists, into the visual layout.
- A chart uses only numbers present in the source. A document with no quantities still gets phase cards and a rail, and it does not invent a chart.
- Timelines, stories, research, and case-task HTML previews share this layout. Chat replies, stick-note faces, and the language setting stay as they are.

## Capabilities

### New Capabilities

- `document-visual-html`: Generated MorphNotes documents use expressive blocks and charts instead of a single plain prose column.

### Modified Capabilities

## Impact

- Shared document CSS and the markdown-to-HTML fragment used by timeline, story, research, and case-task HTML builders.
- The timeline, story, and research generation prompts, so they ask for sections, an optional stat block, and callouts.
- Preview iframes keep the existing dark document frame.
