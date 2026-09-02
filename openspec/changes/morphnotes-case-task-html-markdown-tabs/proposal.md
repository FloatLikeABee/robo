## Why

MorphNotes Case/task details is an edit form only. Operators also need to **read** a case as Markdown and as HTML, the way Timelines and Big notes already do, without leaving the details drawer.

## What Changes

- Case/task details gains tabs: keep the current edit form, plus **Markdown** and **HTML** views of that case.
- Markdown shows a document built from the case (title, description, times, location, JSON detail).
- HTML shows a rendered preview of that document (dark page, same family as Timeline HTML).
- Create and edit both can switch tabs; Markdown/HTML follow the current draft, including unsaved field changes.
- No new CaseTask columns: views are derived, not a second stored body.

## Capabilities

### New Capabilities

- `case-task-html-markdown-views`: Case/task details can be viewed as Markdown and HTML via tabs alongside the existing form.

### Modified Capabilities

- (none — `openspec/specs/` has no archived case-task baseline)

## Impact

- **MorphNotes frontend**: `CaseTasks.js` details drawer (tabs + Markdown/HTML panes).
- **MorphNotes backend**: optional `markdown_content` / `html_content` on GET full, or a small builder reused from timeline/big-note HTML helpers — no SQLite migration.
- **Out of scope**: Publishing a public case URL; changing PDF/email; editing the case as a raw Markdown file instead of the form.
