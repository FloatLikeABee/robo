## Why

MorphNotes Case/task **Markdown** and **HTML** tabs currently dump entity `detail` as a fenced JSON blob (` ```json ` → `<pre><code>`). Operators already have a card-style display on the Details tab; the reading views should show the same fields as a document UI (labels, nested sections, lists), not source JSON.

## What Changes

- Markdown view renders `detail` as readable fields: Title Case labels, nested headings, bullets for lists — not a fenced JSON block.
- HTML view renders `detail` as a dark document UI (cards / definition lists / nested sections), matching the MorphNotes dark Case/task preview — not a monospace JSON dump.
- Empty `{}` omits the Detail section. Invalid JSON still shows the raw text so operators can see what failed to parse.
- Details tab JSON editor (Card view / tree / raw) stays the edit surface. Views stay derived from the current draft. No CaseTask schema change.

## Capabilities

### New Capabilities

- `case-task-detail-display`: Case/task Markdown and HTML views present JSON `detail` as structured, human-readable UI instead of raw JSON text.

### Modified Capabilities

- (none — `openspec/specs/` has no archived case-task baseline; this follows `case-task-html-markdown-views` without replacing the tabs themselves)

## Impact

- MorphNotes Go builder: `buildCaseTaskMarkdown` / `buildCaseTaskHTML` in `morph/handlers/tran_case_task_views.go` (and tests).
- MorphNotes frontend helper: `morph/frontend/src/pages/admin/caseTaskViewDocs.js` (keep JS/Go document shape aligned).
- GET case-task full computed `markdown_content` / `html_content` follow the new shape (still not persisted).
- Out of scope: PDF/email flattening, Details-tab editor, public publish URL, light theme, MorphUtils.
