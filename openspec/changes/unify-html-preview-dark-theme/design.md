## Context

See proposal.md — Why. Today MorphNotes Go handlers embed per-feature CSS with `--accent:#38bdf8` and `#1e293b` gradient stops. Project's `build_project_html` uses `--accent:#2dd4bf` and `#134e4a` (green wash). `darkPreviewSrcDoc.js` only injects `color-scheme` and scrollbars—it does not retheme body colors. morph-engi frontend `.md-prose` links are teal; MorphUtils lists Project with `accent: '#2dd4bf'`.

## Goals / Non-Goals

**Goals:**

- One token set for generated HTML (Go + Rust builders and client export helpers).
- Preview normalization for legacy stored HTML in iframes.
- Project UI + MorphUtils accent aligned to Morph blue-grey.
- Rendered Markdown prose CSS updated in Project panel.

**Non-Goals:**

- Event Logs / Survey Bot emerald UI (formx).
- Content Maker theme editor or published-page themes beyond verifying dark starter is already `#0b1220`.
- Light mode.
- Bulk DB migration to rewrite all existing `html_content` rows (re-save or preview normalization covers legacy).

## Decisions

### 1. Canonical dark document tokens

**Choice:**

```text
--bg: #0b1220
--card: #111827
--line: #1e293b
--ink: #e8eef7
--muted: #94a3b8
--accent: #38bdf8   (links, blockquote border)
--accent-purple: #818cf8  (optional heading/meta emphasis)
gradient: radial-gradient(1200px 600px at 10% -10%, #1e1b4b 0%, var(--bg) 55%)
```

Purple stop `#1e1b4b` (indigo-950) tints the wash; blue stop `#1e293b` remains acceptable for MorphNotes parity. **Replace** Project's `#134e4a` entirely.

**Why:** Matches MorphNotes base with requested purple depth; avoids green.

**Alternative:** Copy Morph AI `--chat-accent` only — less purple; user asked for both blue and purple.

### 2. Single CSS fragment in Go, mirror in Rust

**Choice:** Add `morph/internal/htmldoc/theme.go` (or similar) exporting `DarkDocumentCSS()` prose + layout block. Refactor `buildTimelineHTML`, `buildBigNoteHTML` dark branch, `buildCaseTaskHTML`, and reuse from Research (already delegates to timeline builder). Update `morph-engi` `build_project_html` to the same string constants (duplicate minimal block in Rust—YAGNI on cross-lang package until a third consumer appears).

**Why:** One edit point per language; Project was the outlier.

### 3. Preview normalization for legacy HTML

**Choice:** Extend `withDarkPreviewSrcDoc` (and morph-engi's copy in `ProjectDocumentPanel.svelte`) to inject a trailing `<style id="morph-dark-theme-override">` that sets `body` background/color and overrides `.prose a`, `a`, and common `:root` accent vars when the srcDoc already has `<style>` but uses teal/green accents. Detect `#2dd4bf`, `#134e4a`, `#10b981`, `#059669` in inline styles and apply override anyway (simple substring check).

**Why:** Operators see correct preview without mass re-save.

**Alternative:** Force re-generate all HTML on deploy — heavy, needs migration job.

### 4. Project frontend shell

**Choice:** In `morph-engi/frontend/src/app.css`, shift `--color-bg` toward `#0c1220`, `--color-surface` `#141c2a`, retire `--color-teal` for links in favor of `--color-accent: #38bdf8` / `--color-accent-soft: #818cf8`. Replace `text-teal` / `text-rose-300` on non-destructive chrome with `text-muted` / accent blue; keep `text-rose-*` only on explicit delete/remove actions.

**Why:** User called out red/pink chrome; rose was used on delete and links—separate destructive from navigation.

### 5. MorphUtils Project accent

**Choice:** `config.ts` projects `accent: '#3b82f6'` (same family as `--mu-accent`).

**Why:** Matches shell; one-line fix.

## Risks / Trade-offs

- [Published external URLs still serve old HTML] → Preview normalized in-app; publish/save regenerates HTML. Operators can re-save to fix live slug.
- [Duplicated CSS in Go and Rust] → Keep fragments identical; comment points to twin file.
- [Override style fights author HTML] → Override only targets `body`, `a`, `.prose a`, and `:root` vars; narrow scope.

## Migration Plan

1. Land shared CSS + generators + preview override.
2. Update Project / MorphNotes frontend prose styles.
3. Smoke one record per module (Timeline, Big note, Research, Case/task, Project).
4. Rollback: revert CSS strings; stored HTML unchanged.

## Open Questions

None.
