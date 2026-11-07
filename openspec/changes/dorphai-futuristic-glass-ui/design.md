# Design

## Context

See `proposal.md` (Why). Today Dorph AI hard-codes `data-chat-theme='dark'` with near-black surfaces and blue accents in `App.css`. Admin MUI theme mirrors those blues in `theme.js`. OpenSpec project context previously said “dark-only”; this change **redefines** the single product preset as twilight glass (still no end-user theme toggle).

## Goals / Non-Goals

**Goals**

- One cohesive twilight glass look: lighter than current dark, darker than existing `light` tokens.
- Purple (primary) + orange (secondary) accents with frosted panels and soft mesh background.
- Token-driven CSS so embeds and drawers pick up changes automatically.
- Accessibility and phone chrome preserved.

**Non-Goals**

- Light/dark switch, per-user themes, or reskinning other products.
- WebGL, canvas shaders, or new CSS frameworks.
- Backend or API changes.

## Decisions

### 1. Token strategy: new preset value `twilight` (recommended)

**Choice:** Add `html[data-chat-theme='twilight']` token block in `App.css`; set SkoolAiChat default to `twilight` and persist under existing `THEME_KEY` (migrate stored `dark` → `twilight` once on read).

**Alternatives considered**

| Approach | Pros | Cons |
|----------|------|------|
| Retune `dark` tokens in place | Smallest diff | Breaks “dark” semantic; admin sync confusing |
| Reuse `light` tokens + overlay | Reuses block | Beige/warm wrong vibe; fails “not that light” |
| **New `twilight` preset** | Clear contract; embeds opt in via one attribute | One extra token block (~40 lines) |

### 2. Glass implementation: CSS only

**Choice:** `background: rgba(...)` + `backdrop-filter: blur(12–20px) saturate(1.2)` on `.skool-ai-chat` panels; `-webkit-backdrop-filter` duplicate; `@supports not (backdrop-filter: blur(1px))` raises opacity to ~0.92.

**Self-grill (rejected)**

- **Heavy blur everywhere:** kills scroll perf on older iPads → blur only on 4–5 shell layers, not every message bubble.
- **Glass on user bubbles:** hurts contrast → user bubble stays more opaque tinted purple; assistant bubble lighter frost.
- **npm glass UI kit:** YAGNI.

### 3. Dispersion background

**Choice:** Fixed pseudo-element on `.skool-ai-chat` root: multi-stop `radial-gradient` (purple top-left, orange bottom-right, deep slate base). Optional 60s `background-position` animation; disabled under `prefers-reduced-motion`.

**Alternatives:** SVG noise texture (extra asset); video loop (never).

### 4. Color anchors (starting point, tunable in implementation)

| Token role | Direction |
|------------|-----------|
| Page base | `#12101a` – `#1a1628` (twilight, not black) |
| Purple accent | `#8b5cf6` / `#a78bfa` |
| Orange highlight | `#f97316` / `#fb923c` |
| Text | `#e9e4f2` primary, `#a89bb8` muted |
| Glass panel | `rgba(255,255,255,0.06)` border `rgba(255,255,255,0.12)` |

Send button: purple→orange gradient (reuse pattern from current blue gradient).

### 5. Admin alignment

**Choice:** Update `getAdminTheme('dark')` primary/secondary to purple/orange hues and table header bg to twilight slate; keep `getStoredThemeMode()` returning `'dark'` for MUI `mode` (still dark palette mode, new chroma). Document in code comment that MUI dark mode tracks chat twilight tokens.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Contrast fails on glass | Pre-ship check with WebAIM on composer + assistant bubble; bump opacity if needed |
| `backdrop-filter` jank | Limit blurred nodes; test iOS Safari; opaque fallback media query |
| Embedded iframe double blur | Chat root owns background; panels blur once |
| Hard-coded blues in stray components | Grep `#3b82f6`, `#2563eb` in chat tree during apply; replace with `var(--chat-accent)` |
| Policy drift vs “dark-only” | Update `openspec/project.md` in same change to “single twilight preset” |

## Migration Plan

1. Ship CSS tokens + SkoolAiChat attribute default.
2. One-time localStorage migration `dark` → `twilight` (and write `twilight` on load).
3. Visual QA: standalone chat, DorphNotes embed, phone sessions sheet, reduced-motion/transparency.
4. Rollback: revert attribute to `dark` and old token block (no data migration needed).

## Open Questions

- None blocking: exact hex values and blur radii are implementation tuning within the spec constraints.
