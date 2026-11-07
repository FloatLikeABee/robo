# Design

## Context

See proposal.md. Dorph AI twilight tokens live on `html[data-chat-theme='twilight']` in `dorph/frontend/src/App.css` (`#2c2640` page, `#f4f0fa` text, `#8b5cf6` / `#a78bfa` / `#7c3aed`, highlight `#f97316`, glass fill and blur). Content (`composerx`) is still a blue shell whose header string is Content Maker. Publish HTML is generated in `ComposePublishPanel.svelte` (“Generate from sources”) and by hard-coded builders (`contentsDataPublish.js` is still a light page). The compact html-anything injection still requires `#38bdf8` and mint. Event Logs (`formx`) and DorphTools (`bk/frontend/src/theme.js`) are separate blue/slate themes. Skills modal CSS is slate `#0f172a`.

## Goals / Non-Goals

**Goals:**

- One copied token set, applied in each app. No shared package.
- Content naming, glass shell, preview modal, and record-based generation without a new publisher.
- html-anything structure stays; its hex accents move to twilight.
- Operator chrome for Event Logs, Skills, and DorphTools matches. Public forms stay solid.

**Non-Goals:**

- Renaming `composerx`, the embed id, or the Contents & Data drawer item.
- A theme picker, a chart library, or rewriting stored HTML.
- Putting colours into grill markdown.
- Vendoring the rest of the html-anything template catalog.

## Decisions

### 1. Copy tokens, do not add a package

| Approach | Pros | Cons |
|----------|------|------|
| **Duplicate the twilight hexes in each app (chosen)** | No new dependency; each UI already owns its CSS | Drift if someone edits one copy |
| Shared npm/CSS package | One file | New package across CRA, Vite, and Svelte for a palette |
| iframe CSS injection from Dorph | Parent controls children | Cross-origin; cannot style composerx, formx, or bk |

**Choice:** each surface sets the same hexes. A comment names `App.css` as the source. Glass means translucent panel, 1px light border, backdrop blur, light ink. Orange `#f97316` stays the third accent so existing “at least three accents” still holds. Do not drop to a single purple.

### 2. Content is a label, not a rename of the app

**Choice:** `composerx/frontend` `contentMaker` string, document title, and login subtitle become Content. Routes, folder, and embed id stay. Chinese catalog key stays in parity as 内容. The product UI stays English.

Rejected: renaming the folder or the drawer. That fights the last Contents & Data change and every deploy doc.

### 3. Enlarge is the existing preview, not a second renderer

**Choice:** a button beside the HTML PREVIEW label opens a near-viewport modal whose iframe uses the same `srcdoc` as the inline frame. Hide it while `pageHtml` still contains `data-composerx-starter`. Escape and backdrop close it.

Rejected: a new preview component, or always-on enlarge for the empty starter.

### 4. Generate from Data reads Dorph records through the parent origin

**Choice:** the button in `ComposePublishPanel.svelte` opens a modal. The list is `GET {parent}/api/tran/generic-data` and the body is `GET {parent}/api/tran/generic-data/:id/full`, with the Dorph bearer Content already uses. Parent origin is the embed referrer (DorphNotes on `:3031` proxies `/api`; production is the same host). PDF and markdown become an article from title, analysis, and markdown. CSV, Excel, and JSON become a table from `detail.columns` / `detail.rows`, capped at 50 rows, plus counts. That text is the existing generate-from-sources input. No new publish route.

Rejected: a picker in the Dorph shell that postMessages into the iframe (the control the operator sees is inside Content). Rejected: a composerx proxy route (CORS on Dorph is already open, and the parent origin already proxies `/api`). Rejected: a chart library. Summary figures are counts computed from the rows.

If the list fails, show the error and leave paste generation working.

### 5. html-anything keeps its craft and changes its hexes

**Choice:** edit `skills/vendor/html-anything/condensed.md` (and the injected copy) so the overlay background is `#2c2640`, text `#f4f0fa`, accents `#8b5cf6` `#a78bfa` `#f97316`. Drop `#38bdf8`, `#34d399`, and `#070b14`. Stay hex-only, Arial, dark, no light page, at or under 512 tokens. Update the tests that lock the old cyan. Grill write still does not receive the skill.

Rejected: loading the full upstream html-anything repo. Rejected: leaving Ge cyan/mint in the skill and wrapping pages in a purple frame (the page itself would still be the other theme).

Hard-coded builders (`contentsDataPublish.js`, `darkPreviewSrcDoc.js`, Go HTML in research/timelines/big notes/case tasks, `dorph-engi` project HTML) stamp the same three accents and the purple page on **new** output. Preview injection may restyle a legacy green document while it is on screen; it does not write that HTML back.

### 6. Event Logs glass stops at the operator shell

**Choice:** `formx/frontend` shell tokens and page background become twilight glass. `html.public-form-page` uses solid `#2c2640`, the same ink and purple accent, and no mesh animation.

Rejected: frosting the public form. Rejected: restyling every Tailwind utility by hand when the shell tokens cascade.

### 7. Skills modal and DorphTools

**Choice:** Skills modal rules use `var(--chat-surface)`, `var(--chat-text)`, `var(--chat-border)`, and the purple accent. `bk/frontend/src/theme.js` replaces the sci-fi blue palette with the twilight hexes. MUI paper stays a solid purple (`#352e4a`) because translucent paper breaks contrast; the app bar can blur. No theme switch.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Token copies drift | Comment points at `App.css`; tests lock the three accent hexes in the compact skill and in `contentsDataPublish` |
| Compact skill exceeds 512 tokens | Swap hexes in place; do not add sentences |
| Generic-data detail is huge | Cap table rows at 50, same ceiling as publish-from-analysis |
| Parent origin missing when Content is opened alone | Show the record-list error; paste generate still works |
| Old tests expect `#38bdf8` | Update those assertions in the same change |
| Public form contrast | Solid background, light text, no blur |

## Migration Plan

Ship UI and prompt/template changes. No database migration. Stored pages change on the next generate or publish. Rollback is reverting the CSS, `condensed.md`, and the builders.

## Open Questions

None.
