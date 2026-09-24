## 1. MorphNotes Markdown | Raw widget

- [x] 1.1 Change `MarkdownEditor.jsx` default tab to Markdown (rendered), rename Write → Raw, keep `react-markdown` + remark-gfm, dark surface, no theme switch. Disabled/read-only Raw when `onChange` is omitted.
- [x] 1.2 Replace Timelines Markdown `<pre>` with `MarkdownEditor`. Add PATCH (or smallest existing update) that saves `markdown_content` and regenerates `html_content`. Test: save round-trip.
- [x] 1.3 Replace Big notes Markdown and analysis Markdown `<pre>` with `MarkdownEditor`; persist body via existing Big notes update; analysis stays saved with the response record.
- [x] 1.4 Case/task Markdown tab: default rendered Markdown, Raw is derived source and does not persist a second body. Details save still writes form fields only.
- [x] 1.5 Generic data content: keep rendered Markdown, add Raw edit that saves `content_markdown`. Notes & TODOs body: same Markdown | Raw using the existing note save path.

## 2. Default-select first list item

- [x] 2.1 Timelines: after list load, if nothing selected and the list is non-empty, select `items[0]`. After delete of the selected row, select the new first item or empty state.
- [x] 2.2 Big notes: same first-item and post-delete behavior as Timelines.

## 3. Research schema, ingest, and worker

- [x] 3.1 SQLite `research`, `research_piece`, `research_chunk` (and file name/kind on chunk or a tiny files table) in `sqlite_schema.go`. No MySQL/Neo4j required.
- [x] 3.2 `POST /api/tran/research`: require non-empty prompt; accept optional txt/pdf/csv/json; reject other types with 400 and no row. Extract + chunk into `research_chunk` before rounds. Test: prompt-only create; bad file type rejected.
- [x] 3.3 Async in-process worker (mutex: one running job): 20 rounds — query LLM, `webresearch.Gather`, RAG retrieve by `research_id`, writer LLM, verifier LLM, persist piece each round. Failed round stores an error piece and continues. After 20, concatenate in order, refine, goldmark HTML, status complete. Resume unfinished `running` jobs on Morph boot. Cancel stops further rounds.
- [x] 3.4 `GET` list/detail (pieces + status + markdown/html), cancel, PATCH markdown (regenerate HTML), publish + public HTML like Big notes. Tests: 20 pieces in order after a stubbed run; public GET after publish; empty prompt rejected.

## 4. Research MorphNotes UI

- [x] 4.1 Nav + route `/morphdata/research` (label Research). List/detail cloned from Big notes. Create dialog: prompt + optional files. Poll while running; show round N/20 and completed pieces.
- [x] 4.2 Default-select first research job (and after delete). Conclusion uses `MarkdownEditor`. Publish + open public URL. HTML preview tab like Big notes.

## 5. Content Maker remove published

- [x] 5.1 `DELETE /publishes/:id`: delete published row and stored HTML so `GET /public/p/:slug` no longer serves it. Test: delete then public 404.
- [x] 5.2 `canDelete: true` for published rows. Published contents Remove confirms; cancel leaves the row. Merged draft+published Remove deletes both so the name leaves the list. Update `publishedContents` tests.

## 6. Project first item and Markdown | Raw

- [x] 6.1 Auto-select first project on load; after delete select the new first or empty state.
- [x] 6.2 Markdown tab: rendered default + Raw edit. Add `marked` (same as Content Maker). PATCH `/api/v1/projects/:id` writes `markdown_content` and regenerates `html_content`. Mirror in `browserStore.ts`. HTML tab and publish unchanged. Test: PATCH markdown updates HTML.

## 7. Verify

- [x] 7.1 Static: MorphNotes MarkdownEditor tests/compile; research handler tests; ComposerX publishedContents + delete tests; Project PATCH test. No light/dark toggle added.
- [x] 7.2 Browser MorphNotes: Timelines and Big notes open on first item with rendered Markdown; Raw edit saves (Timeline/Big notes); Case/task Markdown rendered + Raw does not replace Details; Research create (prompt, optional file), progress, rendered conclusion, publish HTML.
- [x] 7.3 Browser MorphUtils: Content Maker Published contents can remove a published page (list gone, public URL dead); Project opens on first item, Markdown rendered, Raw save updates HTML preview.
