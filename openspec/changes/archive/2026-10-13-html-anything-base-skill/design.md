# Design

## Context

See proposal.md for why. html-anything (`nexu-io/html-anything`) is Apache-2.0. Vendoring is allowed if we keep the LICENSE and attribution. The repo has no root `SKILL.md`; the reusable skill is `SHARED_DESIGN_DIRECTIVES` in `next/src/lib/templates/shared.ts` plus 75 template skills. Dorph already has Ge's dark Arial multi-accent theme in `htmldoc` and form-HTML prompts (`openspec/specs/dorph-generated-document-theme`). Prompt assembly for replies lives in `pkg/dorphai`. HTML-producing callers are note text assist, `documentShapePrompt` (timelines / research / stories), `BuildFormHTMLPrompt`, and grill `WritePrompt`. Agents read `.cursor/skills/`, `docs/agents/16-dorph-for-agents.md`, and the embedded `dorph/mcp/agent_guide.md` (those two files must stay byte-identical).

## Goals / Non-Goals

**Goals:**

- One vendor tree with LICENSE, attribution, full Dorph-facing SKILL.md, and the upstream shared directives.
- One compact injection string from `pkg/dorphai` that every HTML-producing path prepends.
- Ge overlay wins on font and palette.
- 512-token budget test on the injected text.
- Agents can find the skill in the pack and chapter 16.

**Non-Goals:**

- Vendoring the Next.js app, the 75 templates, or screenshots.
- Injecting the full SKILL.md into ordinary Dorph chat.
- Changing `htmldoc` CSS (already dark sans).
- A new MCP tool or a `dorph_guide` topic (writing section is enough).
- Render, deploy, or Files workspace work.

## Self-grill (Superpowers)

Three alternatives were grilled before locking the approach. Recommendation is A3 + B2 + C1.

### A. What to put in the tree

1. **Vendor the whole html-anything repo (or all 75 skills).** Complete, but tens of thousands of lines and unused Next/e2e code. Local Ollama cannot load it. Rejected.
2. **Git submodule of `nexu-io/html-anything`.** Tracks upstream, but CI, shallow clones, and Windows checkouts get worse, and we still would not inject the submodule. Rejected.
3. **Vendor LICENSE + shared directives + a Dorph SKILL.md + a condensed file.** Smallest tree that keeps attribution and gives Dorph a base skill. **Chosen.**

### B. When to inject

1. **Always, on every chat turn.** One call site, but it burns Ollama context on "what model are you" and tool loops. Conflicts with `dorph-chat-reply-shape`. Rejected.
2. **Per HTML-producing path, via one helper.** Note AI, generated HTML/MD, form HTML, grill write, plus pack/docs for agents. **Chosen.** Matches the issue list and keeps chat short.
3. **On demand only (picker / MCP resource).** Smallest prompts, but a forgotten caller ships unstyled HTML. The issue requires every named path. Rejected as the only mechanism; the pack still exists for agents.

### C. How the compact text is stored

1. **Source file under `skills/vendor/html-anything/condensed.md`, copied/embedded in `pkg/dorphai`, test that they match.** Same pattern as `16` vs `agent_guide.md`. **Chosen.**
2. **Go string constant only.** Fewer files, but agents and LICENSE tree would drift from what the model sees. Rejected.
3. **Read the vendor file at runtime from disk.** Breaks when cwd is not the repo (Dorph binary, tests, Render). Rejected.

## Decisions

### 1. License check

Apache-2.0 §2 and §4 allow reproduction of source form with the license and attribution notices. There is no NOTICE file upstream. We copy `LICENSE` and add `NOTICE` naming the project and URL. If that reading were wrong we would stop and say so in the PR; it is not.

### 2. Condensed injection API

`pkg/dorphai` exports `HTMLSkillTokenBudget = 512`, `EstimateTokens` (runes/4 rounding up), `HTMLSkillInstructions()`, and `WithHTMLSkill(prompt)`. The injected text is condensed directives plus Ge overlay. It tells the model to keep the format the caller asked for (HTML document vs note/markdown body). Callers:

- `textAssistGenerateNotePrompt`
- `documentShapePrompt` (timelines, research, big notes)
- `BuildFormHTMLPrompt`
- `grill.WritePrompt`

Ordinary chat keeps `ReplyShapeInstructions` only.

### 3. Ge overlay

Overlay text (wins): `Arial, "Helvetica Neue", Helvetica, sans-serif`; background `#070b14`; text `#e8eef7`; accents `#38bdf8`, `#a78bfa`, `#34d399`; dark only; clear hierarchy; varied colours. This matches existing `htmldoc` tokens so prompts and rendered CSS agree.

### 4. Agent exposure

Add `.cursor/skills/html-anything/SKILL.md` (short, points at vendor path and chapter 16). Mention html-anything in the Dorph notes skill and in `docs/agents/16` section 5 (writing). Copy 16 onto `dorph/mcp/agent_guide.md`. Do not add a new initialize paragraph (2000-character cap).

## Risks / Trade-offs

- [Small models ignore craft rules] → Budget stays small; tests lock presence of Arial/dark/accents/html-anything, not model obedience.
- [Note/grill paths start emitting raw HTML] → Condensed text says keep the requested format; grill tests still require body-only rules.
- [Vendor condensed and embed drift] → Same-bytes test, like the agent guide.
- [Upstream later adds a root SKILL.md] → We can refresh the vendor tree; the helper API stays.

## Migration Plan

Code and tests only. Restart Dorph to pick up prompts. Rollback is revert. No data migration.

## Open Questions

None. License allows vendoring. The three grilled alternatives are resolved above.
