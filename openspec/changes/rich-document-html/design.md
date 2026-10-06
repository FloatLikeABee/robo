## Context

See proposal.md for why. `buildTimelineHTML`, `buildResearchHTML`, `buildBigNoteHTML`, and `buildCaseTaskHTML` all drop `markdownToHTMLFragment` into one `.prose` wrapper and style it with `htmldoc.DarkDocumentCSS`. The timeline prompt asks only for dated headings or bullets. The preview iframe shows that HTML as `srcDoc`. Chat already renders mermaid in React; this iframe does not.

## Goals / Non-Goals

**Goals:**

- One layout pass over the existing markdown fragment, shared by the four HTML builders.
- CSS for cards, a rail, callouts, stat chips, and an inline SVG bar chart.
- Prompt lines so generators emit the structure the layout understands.

**Non-Goals:**

- Raw HTML from the model, a chart library, or a mermaid script inside the iframe.
- Restyling Morph AI chat, stick-note faces, or adding a theme switch.
- Inventing statistics the source does not contain.

## Decisions

### 1. Markdown stays the source

The model keeps returning markdown JSON. A stat fence is the only new syntax:

```stat
Places | 5
Talks | 20
```

A blockquote is the callout. `##` headings are the cards. Bullets under a card are the rail.

**Why not mermaid in the iframe:** The preview is a `srcDoc` document with no app bundle. A CDN script can fail offline and can clash with the dark frame. An SVG built from the stat rows is the chart.

**Why not raw HTML from the model:** The current escape path assumes markdown. Model HTML breaks the dark frame and can inject markup.

### 2. Layout is a function of the fragment, not four templates

After `markdownToHTMLFragment`, one function wraps `h2` sections in cards, turns the following `ul` into a rail, turns `blockquote` into a callout, and replaces a stat fence with chips plus an SVG. Each builder keeps its own title and meta line. `DarkDocumentCSS` gains the classes.

### 3. Prompts ask for structure and forbid invented numbers

Timeline, story, and research prompts ask for a lede, `##` sections, a blockquote for a real constraint, and a stat fence only when the source states quantities. Case-task HTML uses the same renderer when its markdown already has those marks. No new case-task prompt.

## Risks / Trade-offs

- [Old documents stay plain until regenerated] → The renderer upgrades any markdown that already uses `##` and lists, including saved pages, the next time HTML is built. A stored page that is only a single bullet list gets the rail, not a chart.
- [A stat row is not a number] → Skip that row. If none remain, render no chart.
- [A long section looks like a card wall] → Cap visual treatment at heading, list, quote, and stat. Paragraphs inside a card stay prose.

## Migration Plan

No data move. Regenerating a document, or any path that rebuilds HTML from markdown, picks up the layout. Rollback is reverting the renderer and the prompt lines. Saved markdown remains valid.

## Open Questions

None.
