## 1. Go document builder (TDD)

- [x] 1.1 Add failing tests: Markdown for `{"priority":"high"}` includes a Detail section with Priority / high and MUST NOT include a fenced `json` dump of that object; nested `case_summary` + `tags` uses a heading and a list; empty `{}` omits `## Detail`; invalid JSON still includes title and the raw text
- [x] 1.2 Add failing tests: HTML for the same object includes a structured Detail UI (e.g. `detail-ui`) with Priority / high, stays dark (`#0b1220`), and MUST NOT use `<pre><code>` pretty-printed JSON as the primary detail
- [x] 1.3 Implement JSON→Markdown and JSON→HTML fragment walkers and wire them into `buildCaseTaskMarkdown` / `buildCaseTaskHTML` until those tests pass (GET full computed fields follow)

## 2. MorphNotes drawer helper

- [x] 2.1 Update `caseTaskViewDocs.js` to the same Detail rules (Title Case labels, nested headings, bullets; HTML injects the card fragment, not a JSON fence) so the live draft tabs match Go

## 3. Verify

- [x] 3.1 Run MorphNotes handler tests covering the new Detail display
- [x] 3.2 Browser: Tasks → Case/task details → Markdown and HTML show labeled detail UI (not a JSON blob); empty detail omits the section; edit detail on Details without save and confirm both view tabs update
