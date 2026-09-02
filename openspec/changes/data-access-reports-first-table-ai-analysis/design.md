## Context

See proposal.md — Why. Data Access header nav is `Sidebar.svelte`: Data tables (`/data-tables`), Data reports (`/reports/builder`), Help. Home `+page.ts` redirects to `/data-tables`. Table detail is `data-tables/[id]/+page.svelte`. Existing `POST /api/v1/data-tables/analyze` converts uploaded files into table JSON — not a narrative analysis. Morph AI (`morphai` crate) already runs in this API. Publish history uses a fixed overlay modal (`DataPagePublishPanel.svelte`). Stay dark-only. User-facing names: Data Access, Data tables, Data reports, MorphUtils.

## Goals / Non-Goals

**Goals:**

- Swap header item order only (reports then tables).
- New table-scoped AI analysis: modal, rendered markdown, `.md` download.
- Authenticated API; reuse Morph AI config.

**Non-Goals:**

- Changing `/` redirect to Data reports.
- Replacing Data reports builder or the file-import analyze endpoint.
- Persisting analysis on the table record (download is the keep path).
- MorphNotes Generic data, Event Logs, Content Maker.
- A theme switch or under-title ledes.

## Decisions

### 1. Nav order in `Sidebar.svelte` only

- **Choice:** Put `{ href: '/reports/builder', label: 'Data reports' }` before Data tables. Help last.
- **Why:** That array is the header. Title mapping in `+layout.svelte` stays as-is.
- **Alternative:** Also send `/` to `/reports/builder`. Rejected — user asked order, not a new home.

### 2. New `POST /api/v1/data-tables/:id/ai-analysis`

- **Choice:** Distinct from file `analyze`. Load table schema + a capped row sample (and aggregates if already available), prompt Morph AI for markdown (Overview / Key findings / Gaps / Actions, grounded in the sample). Return `{ markdown }`.
- **Why:** File-analyze returns columns/rows JSON. Report assistant is a multi-turn tool loop; a one-shot table id is enough for this modal.
- **Alternative:** Drive the existing Data AI chat drawer. Rejected — user asked a modal on Data tables.

### 3. Shared modal on list + detail

- **Choice:** One Svelte modal component. Detail page: **AI analysis** next to existing actions. List: an Analyze control per card that does not navigate away (stop card click).
- **Why:** Spec requires both surfaces; one modal avoids two implementations.
- **Alternative:** Detail only. Weaker vs “on data tables.”

### 4. Render markdown in the modal; download raw `.md`

- **Choice:** Sanitize + render markdown in the dialog (small renderer, e.g. `marked` + sanitizer, or an existing in-repo helper). Download via Blob `text/markdown` named from the table (e.g. `{slug}-analysis.md`).
- **Why:** Display is markdown *style*; download is the source they can keep.
- **Alternative:** Download HTML only. Rejected — user said markdown download.

### 5. Dark modal chrome

- **Choice:** Same overlay pattern as saved builds: `bg-bg-elevated`, `border-border`, no light theme branch.
- **Why:** Data Access is dark-only.

## Risks / Trade-offs

- [Large tables] → Cap sample rows/chars in the prompt; say in the markdown if truncated.
- [No Morph AI key] → 503/error in modal; no fake success download.
- [XSS from markdown] → Sanitize HTML before `innerHTML` / equivalent.

## Migration Plan

- Deploy API + UI together. Rollback: revert Sidebar order and the new route/modal.

## Open Questions

- None.
