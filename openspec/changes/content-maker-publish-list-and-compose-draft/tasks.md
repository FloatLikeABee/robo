## 1. Published and drafts list (TDD)

- [x] 1.1 Add a failing test: insert a `published_pages` row with TEXT `created_at`/`updated_at`, `List` (or `GET /publishes/history` with test auth) returns that row — not a Scan error
- [x] 1.2 Add a failing test: empty `published_pages` list returns 0 items without error
- [x] 1.3 Add a failing test: `publish_drafts` list with TEXT timestamps returns the row
- [x] 1.4 Implement flexible timestamp Scan for published list and drafts list/get until those tests pass

## 2. Compose session keep-alive

- [x] 2.1 Keep `ComposePublishPanel` mounted (hidden) when the operator is on Published contents so name, HTML, and AI chat are not destroyed
- [x] 2.2 Confirm remount `$effect` does not overwrite non-starter HTML when returning to Compose content

## 3. Verify

- [x] 3.1 Run Content Maker backend tests covering list Scan
- [x] 3.2 Browser: Published contents lists pages (no “failed to list published pages”); Compose content, edit HTML, open Published contents, return — draft still there; tab switch does not create a new published page
