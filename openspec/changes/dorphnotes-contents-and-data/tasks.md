# Tasks

## 1. Header title contrast

- [x] 1.1 Restyle `.app-title` to `color: var(--chat-text)` (no clipped accent fill); keep the accent bar

## 2. Nav and routes

- [x] 2.1 Drawer: one Contents & Data item; remove Generic data row; highlight for both old paths
- [x] 2.2 Add combined page (Data + Publish tabs) using existing GenericData and Content Maker embed
- [x] 2.3 Redirect `/dorphdata/generic-data` and `/dorphdata/content-maker` into the combined page; default index follows it
- [x] 2.4 Update `messages.js` (en/zh), `dorphNotesEmbeds.js` label, `EmbeddedModule` copy, and `dorphNotesEmbeds.test.js`

## 3. Publish from analysis

- [x] 3.1 Add Publish as HTML on a Generic data record (title + analysis + capped table excerpt)
- [x] 3.2 POST through `/api/composerx/publishes`; on success open the Publish tab or show the public path
- [x] 3.3 When Content Maker is down/unconfigured, show an error and leave import/analyze working

## 4. Docs and checks

- [x] 4.1 Update `docs/agents/03-dorph.md` (and any drawer-list docs) for Contents & Data
- [x] 4.2 Frontend tests for labels, redirects, and publish payload shape
- [x] 4.3 `openspec validate dorphnotes-contents-and-data --strict`
