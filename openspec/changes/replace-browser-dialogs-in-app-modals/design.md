## Context

See proposal.md — Why. **MorphNotes** wraps the app in `ConfirmProvider` (`ConfirmDialog.jsx`, MUI `Dialog`) and most admin pages call `useConfirm()`. Gaps: `SkillsModal.js` still uses `window.confirm`; `useConfirm` falls back to native dialogs when context is missing. **Event Logs** has a Tailwind `ConfirmProvider` (confirm only). **Project**, **Content Maker**, and **Data Access** (Svelte) call global `confirm()`. **AI tools** (`bk`) uses `window.confirm` and `alert()`.

## Goals / Non-Goals

**Goals:**

- Zero native `alert`/`confirm`/`prompt` in supported product UIs after migration.
- Reuse existing patterns where they exist (Morph MUI dialog, formx overlay).
- Minimal new code: one small dialog helper per Svelte app; extend formx with `alert()` if needed.
- Dark-themed, accessible (`role="alertdialog"`, focus trap, Escape).

**Non-Goals:**

- Replacing MUI `Snackbar`, inline error text, or success toasts.
- Web Push / `Notification` API.
- `prompt()` — none found in stack; add only if a call site appears during migration.
- MorphUtils shell (no native dialogs today).

## Decisions

### 1. MorphNotes: tighten existing provider

**Choice:** Keep `ConfirmProvider` at `index.js`. Remove `window.confirm` / `window.alert` fallback in `useConfirm`; log `console.warn` and resolve `false` / no-op if provider missing (catches mis-mounts in dev). Migrate `SkillsModal.js` to `useConfirm({ danger: true })`.

**Why:** Provider is already global; fallback defeats the goal.

### 2. Svelte apps: lightweight `confirmDialog` store

**Choice:** Add `lib/confirmDialog.ts` (or `.js`) per Svelte app with `confirm(opts)` and `alert(opts)` returning Promises, plus one root `<ConfirmDialog />` in `App.svelte`. Style with existing app CSS tokens (Project blue-grey, ComposerX/SharpReport dark surfaces). Copy shape from formx `ConfirmContext` (title, message, labels, danger).

**Why:** YAGNI — no shared cross-repo package; three small copies beats a new dependency.

**Alternative:** Import formx ConfirmContext into embeds — wrong runtime (React vs Svelte).

### 3. AI tools (bk): React dialog component

**Choice:** Add `ConfirmProvider` mirroring Morph's MUI dialog (bk already uses React). Mount at app root. Replace three call sites.

**Why:** Same stack as MorphNotes admin patterns.

### 4. Event Logs: add `alert()` to ConfirmContext

**Choice:** Extend `ConfirmContext` with `alert({ title, message })` (single OK button), matching Morph's API. Use where error notices need a modal.

**Why:** Parity for future error paths; confirm-only provider is incomplete for alert migration.

### 5. Enforcement

**Choice:** Grep-based checklist in tasks; optional CI grep later (`window\.(alert|confirm|prompt)` in product `src/`). No ESLint rule in this change.

## Risks / Trade-offs

- [Duplicate Svelte dialog code ×3] → Keep each under ~120 lines; same API shape.
- [Modal under embedded iframe] → Each product root owns its overlay `z-index`; no cross-frame dialogs.
- [Missed call site] → Final task: repo-wide grep on supported `src/` trees.

## Migration Plan

1. MorphNotes + SkillsModal + remove fallback.
2. Svelte: Project → Content Maker → Data Access.
3. bk AI tools.
4. formx `alert()` if needed.
5. Grep verification.

## Open Questions

None.
