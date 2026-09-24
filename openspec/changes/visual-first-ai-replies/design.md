## Context

See proposal.md. Specs: `visual-first-ai-replies`, `visual-first-documents`.

Today assistant markdown is GFM only (`ChatMarkdown` / `react-markdown`, `marked` in Content Maker / Project / Data Access). Mermaid fences stay as code. `pkg/morphai.ToolFollowUpPrompt` asks for a markdown summary. Morph AI already has an `image-generator` agent (`generateStoryImageBytes`) and chat `Images[]`. Data Access reports already use ECharts — that path is for table reports, not chat. Dark-only UIs; Morph AI is the system chat (no platform-chat drawers).

## Goals / Non-Goals

**Goals:**

- One shared visual-first instruction string in `pkg/morphai`, used by Morph AI, Event Logs, and Content Maker tool loops.
- Mermaid render in those chats and in markdown/HTML document previews (dark mermaid theme).
- Occasional pixel art via existing image agent + a compact ```pixel``` grid in ordinary chat (canvas/CSS, no per-turn image API).
- HTML iframe previews inject a mermaid runtime so stored documents match chat.

**Non-Goals:**

- ECharts (or other chart SDKs) inside chat — SharpReport keeps ECharts for Data Access report builder only.
- Calling the image API on every assistant reply.
- Mermaid inside email HTML (clients will not run JS).
- Reviving `@robo/platform-chat` drawers.
- Light/dark theme switch.
- Changing MorphNotes Research round counts or Content Maker publish semantics.

## Decisions

### 1. Shared prompt, per-frontend render

**Choice:** Add `VisualFirstInstructions` (and fold it into `ToolFollowUpPrompt`) in `pkg/morphai`. Each markdown UI renders mermaid itself (React `ChatMarkdown` vs `marked` + post-process). No new shared npm package.

**Why:** One Go string updates Morph AI, Event Logs, and Content Maker loops. Frontends do not share a bundler. A new `@robo/markdown` package is extra machinery.

**Alternative:** Prompt-only — rejected; mermaid would still show as a fence. **Alternative:** One shared frontend package — rejected (YAGNI).

### 2. Mermaid is the chart/diagram engine in chat and docs

**Choice:** Add `mermaid` only on frontends that render these surfaces. Programming-like content → flowchart / sequence / class. Numbers → `pie` / `xychart-beta`. Dark theme via mermaid `theme: dark` (or equivalent init). Invalid mermaid → show source.

**Why:** One language the model already knows; no second chart library in Morph AI CRA.

**Alternative:** ECharts in chat — rejected (already in Data Access reports; heavy for bubbles). **Alternative:** Server-side SVG render in Go — extra service, skip.

### 3. Pixel art: existing image agent + ```pixel``` grid

**Choice:** Image-generator agent keeps generating a PNG; prompt it toward pixel-art (limited palette, chunky pixels) when the user asked for art. Ordinary chats MAY emit a fenced pixel grid (rows of hex or `.`/`#` cells) drawn with canvas nearest-neighbor upscale. Do not call `generateStoryImageBytes` from the default tool loop.

**Why:** “Sometimes” must not mean an image round-trip on every “list districts”. A tiny grid is local and cheap.

**Alternative:** New image model / always-on picture — rejected (latency, cost, not asked). **Alternative:** Raw unsanitized SVG in markdown — XSS risk; skip.

Pixel fence (locked): language `pixel`, body is newline-separated rows of `#RRGGBB` tokens or `.` (empty) / `#` (accent). Ceiling: naive parse, max 64×64. Upgrade: palette DSL.

### 4. HTML previews: inject mermaid runtime, not email

**Choice:** Extend `withDarkPreviewSrcDoc` (and the morph-engi twin) to include mermaid init when the HTML contains mermaid. Markdown→HTML keeps `class="mermaid"` on fences. Content Maker / Project `marked` previews run mermaid after parse.

**Why:** Stored HTML is viewed in iframes that can run JS. Email cannot.

**Alternative:** Pre-render mermaid to SVG at save time — extra pipeline; do later if email needs diagrams.

### 5. Single-record exception stays in Morph AI field instructions

**Choice:** Keep “every meaningful field” for one Morph Data record. Visual-first applies to comparisons, lists-as-groups, processes, and code — not to one address.

**Why:** A pie chart of one employee is worse than **Label:** value.

## Risks / Trade-offs

- [Mermaid bundle size on CRA] → Dynamic import mermaid in `ChatMarkdown` so first chat without a fence does not pay. ponytail: if the import is painful, load on first fence only.
- [Model ignores the motto] → Put the contract in both system instructions and `ToolFollowUpPrompt`; Morph AI Research/Design skills can one-line “diagram first”.
- [XSS via mermaid/HTML] → Keep mermaid as text→svg through mermaid’s renderer; do not `dangerouslySetInnerHTML` arbitrary assistant HTML in the chat bubble.
- [xychart-beta flaky] → If a chart type fails, fallback to source; pie + flowchart are enough for v1.
- [Pixel grid ugly] → Cap size; image agent remains the high-quality path.

## Migration Plan

1. `pkg/morphai` prompt + Morph AI `ChatMarkdown` mermaid/pixel + management instructions.
2. Event Logs + Content Maker assistant markdown + the same follow-up string (already imported).
3. Document previews: darkPreviewSrcDoc, Research/Project/Content Maker/Data Access analysis.
4. Rollback: revert prompts; mermaid dep unused if renderers are gated.

## Open Questions

None. Chart types can be tuned without changing the specs.
