## Why

Morph Data’s Stories board is a free-form post + image tool that does not help users turn real source material (documents, pages, or pasted text) into a structured chronological narrative they can review and publish. Renaming and reshaping it into **Timelines** gives a clearer product job: ingest content from file, URL, or paste, have AI produce timeline markdown (and publishable HTML), then keep every outcome in a Big Notes–style list.

## What Changes

- **BREAKING (UX / naming)**: Replace Morph Data **Stories** (nav, routes, page copy) with **Timelines**. Legacy `/stories` and `/story-board` paths redirect to `/timelines`.
- **Create Timeline** accepts **three optional source channels** in one create flow:
  1. **File upload** — `.txt`, `.pdf`, or `.md` only, max **~5 MB** per file
  2. **URL** — fetch and read page/document text
  3. **Paste** — free-text / markdown content
- At least **one** of the three sources MUST be provided; otherwise the UI/API returns a clear **warning / validation error** (no silent create).
- AI turns the combined source text into a **timeline markdown** document; the server also produces **HTML** suitable for display and for **publish** (same pattern as Big Notes / Composer-style public HTML pages).
- Persist each generated timeline (title, source metadata, markdown, HTML, publish fields) and **list** them like Big Notes so users can open, preview (MD + HTML), and publish again.
- Retire Stories-specific create UX (prompt-only AI post + multi-image generate) from the Timelines surface; existing StoryPost data may be migrated, archived, or left inaccessible behind redirects (decide in design — prefer a clean Timelines table/API).

## Capabilities

### New Capabilities
- `morphdata-timelines`: Morph Data Timelines module — create from file/URL/paste (with required-source validation and size/type limits), AI timeline markdown + HTML generation, persisted outcomes listed and viewable like Big Notes, and public publish of HTML.

### Modified Capabilities
- (none — `openspec/specs/` has no published Stories baseline; Stories lived only in product code)

## Impact

- **Frontend (Morph Data)**: `AppDrawer.js` label/route; `appRouter.js` (`/timelines` + redirects); replace or heavily rewrite `StoryBoard.js` into a Timelines page modeled on `BigNotes.js` (list + detail + create dialog + MD/HTML preview + publish).
- **API client**: `tranClient.js` — new timeline endpoints; deprecate Stories endpoints from UI.
- **Backend (Morph)**: New (or evolved) handlers under `/api/tran/timelines` (CRUD, generate, publish, public serve); source intake (multipart file, URL fetch, paste); PDF/txt/md extraction; MorphAI prompt for timeline markdown; HTML build + slug publish similar to `tran_big_notes.go`.
- **DB**: New `timeline` (or equivalent) table for outcomes + optional source blob refs; migration plan for `StoryPost` / story attachments.
- **Dependencies**: Reuse existing PDF extract helpers (`pkg/morphgraph` or patterns from file-import / Big Notes); MorphAI for generation; goldmark (or shared helper) for MD→HTML.
- **ComposerX**: Not required for Morph-hosted publish (Big Notes already publishes without ComposerX); optional future handoff to Composer publish is out of scope unless design chooses proxy.
