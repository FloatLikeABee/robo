## Context

See proposal.md for motivation. Case/task details is a right-hand drawer in `morph/frontend/src/pages/admin/CaseTasks.js`: one scrollable form (title, description, times, map, JSON `detail`, attachments). There are no tabs. PDF (`buildCaseTaskPDF`) and email HTML already flatten the same fields; they are not a reading surface in the drawer.

Timelines and Big notes already store `markdown_content` / `html_content` and use Markdown + HTML tabs (iframe `srcDoc`, dark `#0b1220`). CaseTask has no those columns; description is TEXT and entity `detail` is JSON. `markdownToHTMLFragment` / `buildTimelineHTML` in MorphNotes handlers are the HTML wrapping pattern to reuse.

Product: MorphNotes (not Morph Data). Dark UI only. Do not auto-save. No CaseTask schema migration.

## Goals / Non-Goals

**Goals:**

- Tabs in the details drawer: Details | Markdown | HTML.
- Markdown and HTML are derived from the **current draft** (create and edit).
- HTML preview matches Timeline-style dark document; Markdown is copyable source.
- Backend helper + tests pin the document shape; GET full may expose computed `markdown_content` / `html_content` without persisting them.

**Non-Goals:**

- New SQLite columns or a second editable Markdown body.
- Public publish URL for a case (unlike Timelines).
- Changing PDF download or email send.
- Light-theme HTML.

## Decisions

### 1. Derived views, not stored columns

**Choice:** Build Markdown (and HTML from it) from title, description, assignees, start/end, location, and JSON detail. Do not migrate CaseTask.

**Why:** User asked to *view* the task in those formats. Storing a second copy would desync from the form on save.

**Alternative:** Add `markdown_content`/`html_content` like Big notes — rejected (duplicate source of truth, migration).

### 2. Live draft in the frontend; Go builder for tests and GET full

**Choice:** Drawer Markdown/HTML tabs always render from React draft state so unsaved edits appear without Save. A Go `buildCaseTaskMarkdown` + HTML wrap (reuse `markdownToHTMLFragment` and a dark shell like `buildTimelineHTML`, labeled as Case/task) is unit-tested and optionally returned on `GET /api/tran/case-tasks/:id/full` as computed fields.

**Why:** Spec requires tabs to follow unsaved fields. GET fields help API clients and keep one documented document shape.

**Alternative:** Preview-only POST endpoint — extra round-trip on every tab switch; skip unless JS/Go drift becomes a problem.

**Frontend Markdown tab:** `<pre>` of the generated source (same as Timelines Markdown tab). **HTML tab:** iframe `srcDoc` of the wrapped HTML (sandbox like Timelines). Optional: also render Markdown with ReactMarkdown — not required; source + iframe is enough.

### 3. Document shape

Markdown outline:

```markdown
# {title}

**Assignees:** …
**Start:** …  **End:** …
**Location:** …

{description}

## Detail

```json
{pretty json}
```
```

Omit empty optional lines. Empty title → `# Case/task` or similar fallback so HTML still has a heading. JSON pretty-print like `parseJSONPretty` used by PDF.

HTML: wrap fragment in dark CSS (reuse Timeline `--bg:#0b1220` / `.prose` family). Description is passed through the existing Markdown-to-HTML fragment helper so `**bold**` in description renders.

### 4. Tab chrome

MUI `Tabs`/`Tab` inside the details `Paper`, above the form/preview body. Default tab **Details**. `textTransform: 'none'`. Footer Save/Cancel stays visible for Details; Markdown/HTML are view-only (no second save).

## Risks / Trade-offs

- [JS vs Go document drift] → Keep the outline in this design; tests lock Go output; frontend helper should follow the same headings (`#`, `## Detail`, fenced json).
- [Untrusted description Markdown in iframe] → Use the same sanitizing/fragment path as timelines; iframe sandbox without scripts.
- [Large JSON detail] → Same as PDF; show full pretty JSON in the fence (operators already edit it).

## Migration Plan

- No DB migration.
- Deploy API + MorphNotes frontend together if GET full grows new computed keys (additive JSON).
- Rollback: revert UI tabs; extra GET keys are ignored by old UI.

## Open Questions

None. Preview POST vs client helper is an implementation detail; start with client draft + Go tests.
