# Tasks

## 1. Content name and twilight shell

- [x] 1.1 Change the Content Maker chrome string, document title, and login subtitle to Content (Chinese catalog 内容) and verify the composerx message catalog still has matching keys
- [x] 1.2 Restyle the Content shell (page, panels, accent) to the twilight glass hexes from `dorph/frontend/src/App.css` and verify the primary accent is `#8b5cf6` rather than blue `#3b82f6`

## 2. HTML preview modal

- [x] 2.1 Add an enlarge control that opens the current preview `srcdoc` in a near-viewport modal, hidden while the starter marker is present, and verify a starter page has no enlarge control and a non-starter page opens and closes the modal

## 3. Generate from Data

- [x] 3.1 Rename Generate from sources to Generate from Data, list generic-data records from the parent origin with the existing Dorph bearer, and verify a failed list shows an error without disabling paste generation
- [x] 3.2 On a chosen record, build an article source for PDF or markdown and a capped table-plus-counts source for CSV, Excel, or JSON, then run the existing generate path; verify a unit check of that source builder (50-row cap, article vs table)

## 4. Twilight generated pages

- [x] 4.1 Shift `skills/vendor/html-anything/condensed.md` and the injected copy to `#2c2640`, `#f4f0fa`, `#8b5cf6`, `#a78bfa`, `#f97316`, drop `#38bdf8` `#34d399` `#070b14`, and verify the compact skill test stays at or under 512 tokens and no longer requires cyan
- [x] 4.2 Point `contentsDataPublish.js`, `darkPreviewSrcDoc.js`, and the Go/Rust HTML builders that stamp page colors at that same set for new output, and verify the publish-payload test expects the purple page and the three accents

## 5. Event Logs

- [x] 5.1 Restyle the Event Logs operator shell to twilight glass and keep `public-form-page` on a solid `#2c2640` background without a mesh, and verify the public-form rule has no backdrop animation

## 6. Skills and DorphTools

- [x] 6.1 Restyle the Skills modal with twilight chat tokens and verify `.skills-modal` no longer sets a slate `#0f172a` background
- [x] 6.2 Retheme `bk/frontend/src/theme.js` to the twilight hexes with solid purple paper and no theme switch, and verify primary main is `#8b5cf6`

## 7. Check

- [x] 7.1 Run `openspec validate twilight-content-and-tools --strict` and verify it reports the change valid
