## 1. Go document builder (TDD)

- [x] 1.1 Add failing tests for `buildCaseTaskMarkdown`: title heading, description body, optional assignees/times/location lines omitted when empty, JSON detail in a fenced `json` block
- [x] 1.2 Implement `buildCaseTaskMarkdown` (+ HTML wrap using existing Markdown-to-HTML fragment + dark Timeline-style shell labeled Case/task) until those tests pass
- [x] 1.3 Optionally include computed `markdown_content` and `html_content` on GET case-task full (not persisted); add a test that GET full for a saved case returns both

## 2. Details drawer tabs

- [x] 2.1 Add Details / Markdown / HTML tabs in MorphNotes `CaseTasks.js` details drawer (create and edit); default Details; switching tabs must keep draft state
- [x] 2.2 Markdown tab: `<pre>` of a client helper matching the Go document shape, rebuilt from the current draft
- [x] 2.3 HTML tab: iframe `srcDoc` of the wrapped HTML (dark `#0b1220`, sandbox like Timelines); rebuilt from the same draft

## 3. Verify

- [x] 3.1 Run MorphNotes handler tests covering the new builder (and GET full if added)
- [x] 3.2 Browser: open Tasks → Case/task details → switch Details / Markdown / HTML; edit description without save and confirm both view tabs update; save still persists form fields only
