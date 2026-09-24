## 1. Shared HTML document CSS

- [x] 1.1 Add `morph/internal/htmldoc/theme.go` with dark document + prose CSS constants (blue/purple palette)
- [x] 1.2 Refactor `buildTimelineHTML`, Big notes dark HTML, Case/task HTML, and Research (via timeline) to use the shared fragment
- [x] 1.3 Update `morph-engi/backend/src/api/project_docs.rs` `build_project_html` to the same palette (remove teal/green gradient)
- [x] 1.4 Align `caseTaskViewDocs.js` client export CSS with the shared tokens

## 2. Preview normalization

- [x] 2.1 Extend `morph/frontend/src/lib/darkPreviewSrcDoc.js` to inject dark-theme override for legacy green/teal stored HTML
- [x] 2.2 Mirror the same override helper in `morph-engi/frontend` (`ProjectDocumentPanel.svelte` / shared `lib` if extracted)

## 3. In-app rendered Markdown prose

- [x] 3.1 Update `ProjectDocumentPanel.svelte` `.md-prose` link/accent colors to blue/indigo
- [x] 3.2 Spot-check MorphNotes `MarkdownEditor` / `themed-preview-scroll` surfaces for green MUI primary links; tune if they read teal on dark panels

## 4. Project chrome + MorphUtils accent

- [x] 4.1 Retheme `morph-engi/frontend/src/app.css` tokens to Morph blue-grey; remove teal from default chrome
- [x] 4.2 Replace non-destructive `text-teal` / rose chrome in `App.svelte` and `ProjectDocumentPanel.svelte` with muted blue-grey accents
- [x] 4.3 Set MorphUtils Project module `accent` in `morph-utils/frontend/src/config.ts` to blue (`#3b82f6` or `--mu-accent`)

## 5. Verification

- [x] 5.1 Visual smoke: Project HTML tab + new save — dark blue/purple, no green wash
- [x] 5.2 Visual smoke: MorphNotes Timeline / Big note / Research / Case/task HTML tabs match palette
- [x] 5.3 Open legacy Project record with old teal HTML — iframe preview still reads on-brand via override
