## Context

Morph Data today ships **Stories** as `StoryBoard` + `/api/tran/story-posts` (title/content posts, attachments, prompt→AI image generate). That model does not match the Timelines job. **Big Notes** already demonstrates the target UX pattern: list → create with AI → persist `markdown_content` + `html_content` → detail preview → `POST …/publish` → public HTML at `/api/tran/public/…`. ComposerX also publishes HTML pages, but Big Notes deliberately publishes inside Morph without requiring ComposerX — Timelines should follow that Morph-hosted path for parity and lower coupling.

See `proposal.md` for motivation and scope.

## Goals / Non-Goals

**Goals:**
- Replace Stories UX/nav with Timelines.
- One create flow with three source channels (file / URL / paste), validating that at least one is present.
- Extract text from `.txt`/`.md`/`.pdf` (≤5 MB), optionally fetch URL text, send combined source to MorphAI, store timeline MD + HTML.
- List/detail/delete/publish like Big Notes.

**Non-Goals:**
- Migrating historical StoryPost bodies into timelines (optional best-effort later; not required for launch).
- Multi-file batch upload in v1 (single file per create is enough).
- Publishing exclusively through ComposerX (Morph public HTML is sufficient).
- Interactive timeline editing canvas / drag-reorder events in v1 (markdown edit may be added later; v1 can be regenerate-or-delete).
- Keeping Stories AI image generation on the Timelines page.

## Decisions

### 1. New Timelines API and table (not overload StoryPost)
- **Choice**: Add `timeline` table + `/api/tran/timelines` handlers (list/get/create/delete/publish/public), mirror Big Notes field shapes (`title`, source summary fields, `markdown_content`, `html_content`, `published_slug`, `published_path`, ownership).
- **Why**: StoryPost is social-post shaped (attachments/comments mindset). Timelines are document outcomes. Clean schema avoids awkward columns and dual UX.
- **Alternatives**: Rename StoryPost in place — rejected (wrong shape, attachment model unused). Proxy through ComposerX drafts — rejected for v1 (extra product boundary).

### 2. Source intake on create (multipart + JSON fields)
- **Choice**: `POST /api/tran/timelines` as multipart form:
  - `file` (optional): one `.txt`/`.pdf`/`.md`, max 5_242_880 bytes
  - `url` (optional string)
  - `paste` / `content` (optional string)
  - optional `title` (else AI-derived)
- Server validates ≥1 source, extracts/fetches text, calls MorphAI, builds HTML, inserts row, returns full record.
- **Why**: Matches create UX; one round-trip; same pattern as other Morph file+AI flows.
- **Alternatives**: Separate upload-then-generate endpoints — more states, worse for “at least one source” validation.

### 3. Extraction and URL fetch
- **Choice**: Reuse Morph’s existing PDF text extraction path (`pkg/morphgraph` / shared helper used by other Morph AI imports). `.txt`/`.md` read as UTF-8. URL fetch: server-side HTTP GET with timeouts, size cap on response body, HTML→text stripping (or plain text if Content-Type is text). Block private/link-local addresses if SSRF helpers already exist; otherwise apply basic host deny list.
- **Why**: Consistent with SurveyX/Projects import work; keeps secrets and fetch on server.
- **Alternatives**: Client-only paste of URL content — weaker and fails CORS.

### 4. AI contract
- **Choice**: MorphAI returns JSON with `title` + `markdown` (chronological timeline: headings/events with dates or ordered milestones). Server owns HTML via goldmark (same approach as Big Notes `buildBigNoteHTML`, timeline-specific chrome/CSS).
- **Why**: Keeps HTML deterministic and publish-safe; avoids model inventing broken full HTML documents.
- **Alternatives**: Model emits full HTML — harder to sanitize/publish consistently.

### 5. UI modeled on Big Notes
- **Choice**: New `Timelines.js` (or rewrite `StoryBoard.js`) with list + detail + create dialog (three source fields + client-side “need at least one” warning) + MD/HTML tabs + Publish. Nav label **Timelines**; routes `/timelines`; redirects from `/stories` and `/story-board`.
- **Why**: User asked for Big Notes–like listing; Big Notes already teaches publish/preview.
- **Alternatives**: Keep tile/card StoryBoard layout — poorer fit for long markdown documents.

### 6. Stories deprecation
- **Choice**: Stop linking Stories in nav; leave `story-posts` API in place temporarily (unused by UI) or gate behind no routes. Do not auto-migrate StoryPost rows in v1.
- **Why**: Fastest safe cutover; avoids bad automatic “timeline” conversions from short social posts.
- **Alternatives**: One-shot SQL rename — misleading data quality.

### 7. Publish
- **Choice**: `POST /api/tran/timelines/:id/publish` + `GET /api/tran/public/timelines/:slug` serving stored/rebuilt HTML, parallel to Big Notes.
- **Why**: User asked “published like composer”; Morph Big Notes is the in-product equivalent already used by Morph Data.

## Risks / Trade-offs

- **[Risk] PDF with only scanned images yields empty text** → Mitigation: fail create with clear “no extractable text” error; optional future vision pass is out of scope.
- **[Risk] URL fetch SSRF / large pages** → Mitigation: timeouts, max body bytes, reject non-http(s), block obvious private hosts.
- **[Risk] Combined sources exceed model context** → Mitigation: truncate with preserved notice in source summary; prefer chronological sections from remaining text.
- **[Risk] Users expect Stories posts to appear under Timelines** → Mitigation: empty new list + short empty-state copy that Timelines are document-based; optional later migration tool.
- **[Trade-off] Single file per create** → Simpler validation; multi-file can be a follow-up.

## Migration Plan

1. Ship migration creating `timeline` (+ indexes on owner/created).
2. Deploy API + public serve routes.
3. Deploy frontend Timelines page, nav rename, redirects.
4. Verify create (file/URL/paste), list, preview, publish, public GET.
5. Rollback: revert frontend routes/nav; timeline table can remain unused. StoryPost data untouched.

## Open Questions

- Whether to allow optional title override + regenerate endpoint in the first ship (lean: create-only regenerate via new create, or add `POST …/:id/regenerate` if Big Notes parity is desired in the same PR).
- Exact empty-state copy mentioning that prior Stories are not imported.
