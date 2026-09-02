## Why

Content Maker **Published contents** shows “failed to list published pages” instead of the pages operators already published. **Compose content** also throws away the in-progress HTML/name/AI chat when they click Published contents (or another inner tab), so returning to Compose starts from the starter page.

## What Changes

- `GET /publishes/history` MUST list published pages (or an empty list) instead of HTTP 500 with `failed to list published pages`.
- The same SQLite timestamp scan issue MUST not break `GET /publish-drafts` if that list is used on the same screen.
- Compose content MUST keep the current draft (name, HTML, AI thread) when the operator switches to Published contents and back. Explicit Publish / Save draft still write to the server.
- No change to MorphUtils outer nav, route ids (`composerx` / `/composerx`), or dark-only chrome.

## Capabilities

### New Capabilities

- `published-pages-list`: Authenticated list of published pages (and drafts on that screen) succeeds against SQLite TEXT timestamps.
- `compose-content-session`: Compose content in-progress work survives switching away to Published contents and back.

### Modified Capabilities

- (none — `openspec/specs/` has no archived Content Maker baseline)

## Impact

- **Content Maker backend**: `published_pages` / `publish_drafts` list Scan of `created_at` / `updated_at`.
- **Content Maker frontend**: `App.svelte` mounts Compose vs Published as a destroy/create `{#if}`; `ComposePublishPanel.svelte` holds all compose `$state`.
- **Out of scope**: Changing public `/public/p/:slug` URLs; MorphUtils iframe keep-alive (already keeps the Content Maker iframe mounted).
