## Context

See proposal.md — Why. `.hybrid-drawer` in `App.css` sets `background: var(--chat-bg-elevated, #1e1f24)`. `--chat-bg-elevated` is not defined on `html[data-chat-theme='dark']`; the fallback is neutral grey. Context & Knowledge (`.hybrid-drawer--panel`) and the Morph AI Notes & TODOs drawer share that class. MorphNotes Notes & TODOs uses MUI `background.paper` (`#060a12`) plus a header gradient with leftover purple (`rgba(156,39,176,…)`). Chat tokens already exist: `--chat-page-bg` `#020408`, `--chat-panel` `#060a12`, `--chat-surface` `#0c1220`, `--chat-border` `rgba(37, 99, 235, 0.22)`. Stay dark-only.

## Goals / Non-Goals

**Goals:**

- Point hybrid-drawer (and related fallbacks) at existing `--chat-*` tokens.
- Align MorphNotes Notes & TODOs paper/header with that blue surface.
- Keep Files / chat column already on `--chat-surface` as the visual reference.

**Non-Goals:**

- New palette or a light/dark switch.
- Rewriting Context & Knowledge copy or Notes & TODOs layout.
- Restyling MorphUtils, Data Access, Event Logs, or Content Maker.

## Decisions

### 1. Reuse `--chat-surface` / `--chat-panel`, do not invent `--chat-bg-elevated`

- **Choice:** `.hybrid-drawer` background = `var(--chat-surface)` (or `--chat-panel`); color = `var(--chat-text)`; borders = `var(--chat-border)`. Replace `#1e1f24` / `#3a3c48` / `--chat-text-soft` fallbacks with `--chat-text-muted` where that is the muted text token.
- **Why:** Undefined `--chat-bg-elevated` is why Knowledge looks grey next to Files (`agent-workspace` already uses `--chat-surface`).
- **Alternative:** Define `--chat-bg-elevated: #1e1f24`. Rejected — that keeps the grey.

### 2. MorphNotes Notes panel: same blues, drop purple header wash

- **Choice:** `adminRightPanelPaperSx` `bgcolor` = `#0c1220` (surface) or `#060a12` (panel) to match chat; header background = blue-only gradient (`rgba(37, 99, 235, …)`), no magenta.
- **Why:** Header still uses Skool-era purple; paper is close but not the same as `--chat-surface`.
- **Alternative:** Leave MUI paper and only fix CSS drawers. Rejected — user named MorphNotes Notes & TODOs.

### 3. Shared Notes content stays on theme paper

- **Choice:** Only add an explicit dark-blue `bgcolor` on the NotesTodosContent root if MUI fields still punch grey holes after the panel fill is fixed. Prefer theme tokens already in `getAdminTheme('dark')`.
- **Why:** Avoid restyling every TextField unless the panel still looks grey.
- **Alternative:** Custom CSS for every MUI input. Overkill unless verification shows grey holes.

## Risks / Trade-offs

- [Light `data-chat-theme` leftovers] → Product is dark-only; do not add a switch. Hybrid-drawer light overrides can stay unused or be retargeted later.
- [Contrast on `#0c1220`] → Existing chat text tokens already pair with that surface.

## Migration Plan

- Frontend CSS + MorphNotes panel sx only. Rollback: revert those files.

## Open Questions

- None.
