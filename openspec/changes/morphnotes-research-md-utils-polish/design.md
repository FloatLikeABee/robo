## Context

See proposal.md for motivation. Specs: `notes-markdown-display-edit`, `notes-list-default-first`, `morphnotes-research`, `content-maker-remove-published`, `project-markdown-and-default`.

Today MorphNotes Timelines / Big notes / Case/task Markdown tabs dump source in `<pre>`. Generic data already uses `ReactMarkdown`. Unused `MarkdownEditor.jsx` is Write-default Preview. Timelines and Big notes start with `selectedId = null`. Knowledge Library upload is gated on `TranMySQL` even though SQLite already has `morph_knowledge_*` tables. ComposerX Published contents sets `canDelete: false` for published rows; only `DELETE /publish-drafts/:id` exists. Project (`morph-engi`) Markdown is a raw `<pre>`; `PATCH /projects/:id` does not write `markdown_content`. `pkg/webresearch.Gather` is keyless (Wikipedia / DDG / arXiv). Morph `Gather` to AI tools is optional and already caps iterations at 20.

Dark-only. MorphUtils is an iframe shell — Research is not a fifth embed.

## Goals / Non-Goals

**Goals:**

- One MorphNotes Markdown/Raw widget reused on stored bodies; Case/task Markdown is rendered + read-only Raw.
- First-item select on Timelines, Big notes, Research, and Project lists.
- Async Research job: file RAG on SQLite, 20 checkpointed online rounds, verifier call per round, concatenate, refine, goldmark HTML, Morph-native publish.
- Content Maker can delete published pages (list + public URL).
- Project Markdown rendered + Raw save that regenerates HTML.

**Non-Goals:**

- Research as a MorphUtils module or ComposerX-only publish.
- Requiring MySQL, Mongo, Redis, or Neo4j.
- Real OS-level subagent processes or a new queue service.
- Saving Case/task by editing derived Markdown.
- Theme switch, chat-bubble changes, HTML-source tabs as Markdown.
- Making Knowledge Library’s MySQL gate the Research ingest path.

## Decisions

### 1. Reuse `MarkdownEditor` as Markdown | Raw, default Markdown

**Choice:** Change the unused MorphNotes `MarkdownEditor` default from Write to **Markdown** (rendered `ReactMarkdown` + `remark-gfm`) with **Raw** (textarea). Wire it into Timelines Markdown tab, Big notes Markdown + analysis Markdown, Generic data content, Notes & TODOs body, Research conclusion. Case/task Markdown tab uses the same chrome with `onChange` omitted / disabled Raw (copyable source only). Keep existing HTML iframe tabs where they already exist.

**Why:** Widget already exists; Generic data and chat already depend on `react-markdown`. Default preview matches the spec. Case/task stays derived (existing `case-task-html-markdown-views` decision).

**Alternative:** New package or MDX editor — rejected (new dep, unused widget already here). **Alternative:** HTML iframe as the “markdown display” — rejected (that is the HTML tab).

### 2. Default-select first list item after load (and after delete)

**Choice:** After the list fetch, if nothing is selected and `items[0]` exists, set that id. After deleting the selected row, select the new `items[0]` or clear. Same for Research and Project. Do not create rows. Do not change sort order.

**Why:** Notes & TODOs already auto-selects `visibleItems[0]`. Copy that, don’t invent a URL scheme.

**Alternative:** Remember last id in `localStorage` — rejected (user asked for first in the menu).

### 3. Research lives in MorphNotes, cloned from Big notes

**Choice:** Nav + route `/morphdata/research`, handlers `/api/tran/research*`, SQLite tables next to `big_note` / `timeline`. List/detail UI cloned from `BigNotes.js`. Publish like `PublishBigNote` (public slug, no ComposerX).

**Why:** Same product shape (AI generate → MD/HTML → publish). MorphUtils would be a new iframe app for no gain.

**Alternative:** Wrap AI tools Gathering UI — rejected (wrong product, no MorphNotes list/publish, MySQL/bk coupling).

### 4. Job-scoped RAG on SQLite, not Knowledge Library HTTP

**Choice:** `research` row + `research_chunk` (file_id, chunk_index, text, optional embedding_json). On upload: `morphgraph` / `docextract` for pdf/txt/csv/json, then existing chunk+optional embed helpers. Retrieval per round: same token-overlap / cosine ranking as `SearchKnowledgeChunks`, filtered by `research_id`.

**Why:** Knowledge upload currently 503s without `TranMySQL`. Spec forbids Neo4j. Job-scoped chunks avoid polluting the global library and keep RAG working on `./data` SQLite.

**Alternative:** Call `POST /api/knowledge/files` — rejected (MySQL gate). **Alternative:** Full scan of file text into the LLM every round — rejected for large PDFs; chunk retrieve is the existing pattern.

Ceiling: O(n) chunk scan per round (same as today’s knowledge search). Upgrade: ANN/Neo4j when GraphRAG is on.

### 5. Twenty rounds are an async in-process worker, not one HTTP request

**Choice:** `POST` creates the job, ingests files, returns the id immediately, starts **one in-process goroutine** ( Morph already runs an ingest worker). Each round: (1) LLM proposes the next query from prompt + top RAG hits + prior piece titles, (2) `webresearch.Gather`, (3) writer LLM → piece Markdown, (4) verifier LLM (second call, “subagent”) appends a Verification subsection. Persist piece `round_index` 1–20 after each round. Then concatenate in order, refine LLM, goldmark → `html_content`, status `complete`. GET polls. Cancel sets a flag the loop checks. Failed round writes an error piece and continues. Resume: worker restarts from `max(round_index)+1` if the process died mid-job (on Morph boot, pick up `running` rows).

**Why:** 20 × (web ~14s + two LLM calls) will exceed any browser/Gin timeout. Pieces must survive refresh (spec). `webresearch` needs no API key. AI tools `Gather` is optional extra context if that process is up — do not block the job on it.

**Alternative:** One `Gather(maxIterations=20)` then split — rejected (no per-round pieces or verifier). **Alternative:** 20 OS processes / Morph `RouteSubAgents` fan-out — rejected (`RouteSubAgents` is prompt-level, max 3, and “never disk-write workers”). Verifier = second LLM call with sources in context.

Ceiling: one research job running at a time (mutex). Upgrade: a real job table worker pool if concurrent runs matter.

### 6. Content Maker `DELETE /publishes/:id`

**Choice:** Repository delete by id: drop SQLite `published_pages` row and the HTML body store. `canDelete: true` for published rows. Merged draft+published Remove deletes **both** so the row leaves the list. Confirm dialog. Public `GET /public/p/:slug` 404s after.

**Why:** Spec is remove from list + stop serving. Draft-only delete already exists.

**Alternative:** Unpublish flag keeping the row — rejected (user asked to remove). **Alternative:** Delete published but keep draft for merged rows — would leave a Saved row; spec says the name leaves the list.

### 7. Project Markdown/Raw + PATCH markdown

**Choice:** Default-select first project like Timelines. Markdown tab: rendered preview + Raw textarea. Add `marked` to morph-engi frontend (Content Maker already uses it; morph-engi has no markdown lib). Save calls `PATCH /api/v1/projects/:id` with `markdown_content`; backend regenerates `html_content` with the existing `build_project_html`. Mirror PATCH in `browserStore.ts`. HTML tab and publish unchanged.

**Why:** `update_project` already exists but omits markdown columns — smallest API change is extending that PATCH. Zero-dep preview via stale `html_content` would show unsaved Raw incorrectly.

**Alternative:** Iframe of `html_content` as the Markdown tab — rejected (that is the HTML tab). **Alternative:** New save endpoint — rejected (PATCH already there).

## Risks / Trade-offs

- [20-round runtime / cost] → Async job, persist every round, show N/20, mutex one job. Operator can cancel.
- [webresearch is shallow vs a crawler] → Enough for keyless local stack; optional AI tools Gather if up. Spec is 20 rounds of online gather, not a new search engine.
- [Verifier is not an independent agent] → Second constrained LLM call with sources; honest vs fake multi-agent infra.
- [Deleting published HTML is irreversible] → Confirm dialog; no soft-delete table in v1.
- [marked in morph-engi] → Same library as Content Maker; MorphNotes stays on `react-markdown`.

## Migration Plan

1. SQLite `CREATE TABLE IF NOT EXISTS` for `research`, `research_piece`, `research_chunk` (and file metadata). No MySQL migration.
2. ComposerX delete route is additive; old clients without Remove still list pages.
3. Project PATCH ignores missing markdown fields so old clients keep working.
4. Rollback: hide Research nav; unused tables; revert `canDelete`; MarkdownEditor default can revert without data loss.

## Open Questions

None — round count (20), MorphNotes home, continue-on-round-failure, and Morph-native HTML publish are fixed in the specs.
