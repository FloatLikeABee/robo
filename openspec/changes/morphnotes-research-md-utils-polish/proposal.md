## Why

MorphNotes and MorphUtils still show most document bodies as raw `<pre>` / textarea, and Timeline, Big notes, and Project open with an empty detail pane. Operators also need a first-class Research module (prompt + optional files → RAG → 20 verified online rounds → publishable conclusion) and a way to remove published Content Maker pages.

## What Changes

- MorphNotes markdown panels **display rendered Markdown by default** and switch to **raw Markdown for edit** (stored bodies save; derived Case/task Markdown stays view-only in Raw).
- **Timelines** and **Big notes** auto-select the first list item when the module opens and the list is not empty.
- New MorphNotes **Research** module: prompt, optional txt/pdf/csv/json uploads ingested into RAG, 20 online research rounds with subagent verification, stepwise conclusion pieces, a final refine, Markdown + HTML output. HTML is publishable; Markdown displays rendered (with Raw edit).
- Content Maker **Published contents** can **remove** published pages (not only Saved drafts).
- MorphUtils **Project**: auto-select the first project; Markdown displays rendered and switches to raw edit.

## Capabilities

### New Capabilities

- `notes-markdown-display-edit`: MorphNotes document panels show rendered Markdown with a Raw edit mode.
- `notes-list-default-first`: Timelines and Big notes select the first list item on enter.
- `morphnotes-research`: MorphNotes Research module — RAG uploads, 20 verified online rounds, stepwise MD/HTML conclusion, HTML publish.
- `content-maker-remove-published`: Published contents list can delete published pages so they leave the list and public URL.
- `project-markdown-and-default`: Project module selects the first project on enter and shows Markdown rendered with Raw edit.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- **MorphNotes** (`morph/`): `MarkdownEditor.jsx` reused across Timelines, Big notes, Case/task Markdown tab, Generic data content, Notes & TODOs body; `AppDrawer` + `appRouter` + `/api/tran/research*`; SQLite `research_*` tables; `pkg/webresearch` + existing knowledge chunking; Morph-native HTML publish like Big notes.
- **Content Maker** (`composerx/`): `DELETE /publishes/:id`, `ComposePublishRecordsPanel`, `publishedContents.js`.
- **Project** (`morph-engi/`): first-item select, Markdown/Raw UI, PATCH `markdown_content` (regenerate HTML); browser preview store too.
- **MorphUtils** shell: unchanged (iframe hosts). Dark-only. No new Research app embed.
- **Dependencies:** MorphNotes keeps `react-markdown`. Project may add `marked` (already used by Content Maker) for live Markdown preview. No MySQL/Neo4j required; GraphRAG remains optional.
