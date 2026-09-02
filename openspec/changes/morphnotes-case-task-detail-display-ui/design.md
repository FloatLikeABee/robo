## Context

See proposal.md for motivation. Case/task Markdown/HTML are derived in `buildCaseTaskMarkdown` / `buildCaseTaskHTML` (Go) and `caseTaskViewDocs.js` (drawer). Both currently append `## Detail` plus a fenced `json` pretty-print; HTML runs that Markdown through `markdownToHTMLFragment`, so JSON becomes `<pre><code>`. The Details tab already has a card display (`JsonDetailValueTable` in `jsonDetailViews.jsx`: Title Case labels, nested stacks, Yes/No). Specs: `specs/case-task-detail-display/spec.md`. Product: MorphNotes. Dark only. No CaseTask migration.

## Goals / Non-Goals

**Goals:**

- Replace the fenced JSON Detail section with a structured walker shared in spirit by Go and JS.
- HTML injects a dedicated Detail UI fragment (cards / labeled blocks) into the existing dark Case/task shell — not “render the JSON fence as HTML.”
- Markdown uses **Label:** value, nested `###` headings, and bullets.
- Empty `{}` omits Detail; invalid JSON shows escaped raw text.

**Non-Goals:**

- Changing the Details-tab JSON editor (Card / tree / raw).
- PDF or email body (still may pretty-print JSON).
- Persisted `markdown_content` / `html_content` columns.
- Light-theme HTML.

## Decisions

### 1. Walk JSON for Markdown; inject HTML fragment (do not round-trip HTML through the fence)

**Choice:** `detailJSONToMarkdown(raw)` builds the Detail markdown. `detailJSONToHTMLFragment(raw)` builds escaped HTML cards. `buildCaseTaskHTML` keeps title + description + meta via the existing markdown→HTML path, then appends the Detail fragment inside `.wrap` (after `.prose` or as a sibling `<section class="detail-ui">`). Do not put fenced JSON into Markdown and expect the fragment converter to look nice.

**Why:** `markdownToHTMLFragment` only handles headings, paragraphs, and fenced code. A real UI needs cards, definition-like rows, and nested groups with CSS already in the dark shell.

**Alternative:** Only change Markdown to **Label:** lines and reuse the fragment converter — rejected; that is still a wall of paragraphs, not a display UI.

### 2. Display rules (match Details card view)

**Choice:** Recurse parsed JSON:

- Object → one card (HTML) or `### Label` / `**Label:**` (Markdown) per key; keys Title Case (`priority` → Priority, `case_summary` → Case Summary, `Id` → ID). Nested objects nest; arrays become lists (`Item n` in HTML, `-` in Markdown).
- Boolean → Yes / No. Null / missing → omit or em dash. Empty object/array at the root → omit the whole Detail section. Nested empty → short “Empty” / “Empty list.”
- Invalid JSON → Detail heading plus escaped raw text (blockquote or `.detail-fallback`), never a crash.
- Cap nesting (reuse the existing JSON detail nesting limit) so a huge blob cannot explode headings.

**Why:** Operators already know this layout from Details → Card view.

**Alternative:** Markdown tables for every object — rejected for irregular nested AI detail.

### 3. Keep JS and Go aligned; TDD on Go

**Choice:** Same outline in `caseTaskViewDocs.js` so the drawer (live draft) matches GET full computed fields. Extend `tran_case_task_views_test.go`: `{priority:high}` MUST contain Priority and MUST NOT contain ` ```json `; nested `case_summary` + `tags`; empty `{}` omits `## Detail`; HTML contains `detail-ui` (or equivalent) and not `<pre><code>` of `{"priority"`.

**Why:** Previous tabs change already drifted if only one side updated.

**Alternative:** Frontend-only HTML — rejected; GET full would still dump JSON.

## Risks / Trade-offs

- [JS vs Go label drift] → Pin examples in Go tests; JS helper copies the same rules (Title Case, Yes/No, omit empty root).
- [Huge / cyclic-looking detail] → Nesting cap; escape all text in HTML.
- [Invalid JSON in the editor] → Fallback raw text; rest of the document still renders.

## Migration Plan

- No DB migration. Deploy MorphNotes API + frontend together so GET full matches the drawer.
- Rollback: revert builders; old fenced JSON is additive-document-shape only.

## Open Questions

None.
