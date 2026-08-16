## 1. Data model and routes scaffolding

- [x] 1.1 Add SQL migration for `timeline` table (owner/user keys, title, source summary fields, markdown_content, html_content, published_slug/path, timestamps)
- [x] 1.2 Register `/api/tran/timelines` CRUD + publish + public serve routes in `register_routes.go` (parallel to big-notes)
- [x] 1.3 Add `tranEndpoints` timeline helpers in `tranClient.js`

## 2. Source intake and generation backend

- [x] 2.1 Implement create handler: multipart parse of optional `file`, `url`, `paste`/`content`, optional `title`; reject when all sources empty
- [x] 2.2 Enforce file rules: only `.txt`/`.pdf`/`.md`, max 5 MB; clear unsupported-type and size errors
- [x] 2.3 Extract text from txt/md/pdf; fail when extraction yields empty usable text
- [x] 2.4 Fetch URL text with timeout, body size cap, and basic SSRF-safe host checks; fail clearly on unreachable/unreadable URLs
- [x] 2.5 Call MorphAI for `title` + timeline `markdown`; build HTML server-side (goldmark / Big Notes-style page wrapper)
- [x] 2.6 Persist timeline row and return full record; implement list/get/delete for owner

## 3. Publish

- [x] 3.1 Implement `POST /api/tran/timelines/:id/publish` with unique slug allocation
- [x] 3.2 Implement `GET /api/tran/public/timelines/:slug` serving HTML without auth
- [x] 3.3 Reject publish when HTML/markdown content is empty

## 4. Timelines UI (Big Notes–like)

- [x] 4.1 Add Timelines page: list saved timelines, open detail, delete with confirm, empty-state copy that prior Stories are not imported
- [x] 4.2 Create dialog with three sources (file picker, URL field, paste area) plus client-side warning when none provided
- [x] 4.3 Detail view with markdown + HTML preview tabs and Publish action wired to API
- [x] 4.4 Wire create to multipart API; surface validation/AI/extraction errors in the dialog

## 5. Nav rename and Stories cutover

- [x] 5.1 Rename AppDrawer item Stories → Timelines; route to `/timelines`
- [x] 5.2 Update `appRouter.js`: Timelines route; redirect `/stories` and `/story-board` → `/timelines`
- [x] 5.3 Remove or stop mounting StoryBoard as the primary module (keep redirects only)

## 6. Verification

- [x] 6.1 Manual/API check: paste-only, file-only, URL-only creates succeed; no-source returns warning/error
- [x] 6.2 Manual/API check: reject >5 MB and non txt/pdf/md files
- [x] 6.3 Manual check: list/detail/delete; publish then open public HTML URL unauthenticated
