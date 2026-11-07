# Design

## Context

See `proposal.md`. First twilight pass (`App.css` `data-chat-theme='twilight'`) uses page `#14101c`, `--chat-glass-bg: rgba(255,255,255,0.06)`, blur 16px. Mesh blooms sit under a 85% opaque dark veil. That is why it still feels like a dark theme with a tint, not glass.

## Goals / Non-Goals

**Goals**

- Same `twilight` attribute; retune tokens and glass CSS only.
- Dusk floor ~2–3 stops lighter; frost that actually transmits the mesh.
- Contrast and a11y fallbacks still pass.

**Non-Goals**

- New theme id, light mode, extra layers/libraries.

## Decisions

### 1. Lift in place vs new `mist` theme vs true light glass

| Approach | Pros | Cons |
|----------|------|------|
| **Retune `twilight` (recommended)** | One preset; localStorage already `twilight`; smallest diff | Must keep AA on lighter frost |
| New `mist` attribute | Could keep old twilight as fallback | Second preset; user did not ask for a toggle |
| Switch to light tokens + dark text | Maximum “lighter” | Breaks “not that light”; beige/daytime from prior reject |

**Choice:** retune `twilight`.

### 2. How to make it actually glass (self-grill)

Rejected:

- Only raising `--chat-page-bg` and leaving 6% frost → still a dark sheet on a slightly lighter room.
- Blurring every bubble at 40px → scroll cost; assistant already frosted.
- White wash overlay on `.app` → kills purple/orange.

**Choice:** three knobs together:

1. Page/mesh: base ~`#2c2640`–`#352e4a`; blooms brighter (purple ~0.5, orange ~0.4); inner veil lighter (`rgba(48,40,72,0.45)` not `0.85`).
2. Glass fill: `--chat-glass-bg` ~`rgba(255,255,255,0.16)`–`0.20`; borders `rgba(255,255,255,0.28)` plus `inset 0 1px 0 rgba(255,255,255,0.35)`.
3. Blur: chrome 16px → ~28px; assistant ~16px. Keep blur on the same 4–5 shell nodes.

### 3. Text contrast

Lighter frost can wash `#e9e4f2`. Keep text at `#f4f0fa` (or similar) on glass; user bubble stays opaque-ish purple (`>= 0.88` alpha). If AA fails on assistant bubble, raise bubble fill slightly—not the whole page.

### 4. Admin MUI

`background.default/paper` follow the new dusk hexes (`#2c2640` / `#352e4a` range). Accents stay purple/orange from the first change.

### 5. Fallbacks

Opaque fallbacks use lifted `--chat-surface` / `--chat-surface2` (~0.92 dusk), never `#14101c`.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Text contrast on 20% white frost | Check composer + assistant bubble; bump fill or text |
| iOS blur jank at 28px | Same node cap; opaque media-query fallback |
| “Still too dark” vs “now too pale” | Stay in dusk band; no beige |

## Migration Plan

CSS + `theme.js` only. No localStorage migration. Rollback = revert those files.

## Open Questions

None. Exact hex/blur are apply-time tuning inside dusk + stronger-frost.
