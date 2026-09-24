## Context

See proposal.md — Why. Event Logs Layout still has an Info Sheets tab (`/survey-bot`). `pkg/morphai.VisualFirstInstructions` already defines mermaid-first replies and is used on Morph AI tool follow-ups and Content Maker follow-ups; Event Logs’ first-turn prompt, Data Access, Project, MorphNotes text-assist, case-task draft, and research compose do not. Case-task create still shows description, map, JSON detail, and attachments; AI draft returns location + detail JSON.

## Goals / Non-Goals

**Goals:**

- Remove Info Sheets from operator Event Logs UI and MorphUtils copy; redirect `/survey-bot` to `/events-info`.
- Reuse `VisualFirstInstructions` as the graphs contract; seed a Morph AI builtin Graphs skill; append the same string to MorphUtils and MorphNotes AI prompts.
- Create-task UI: prompt, upload, Generate, title, start/end; Markdown + HTML tabs. AI draft keys: title, start_at, end_at, markdown.

**Non-Goals:**

- Deleting survey-bot backend tables or public `/s/:slug` collect URLs in this change (operator UI gone; leftover APIs can rot).
- Forcing a chart on every greeting or a single Morph Data record.
- New chart libraries (mermaid already in Morph AI / Event Logs / Content Maker previews).
- Redesigning task **edit** of historical JSON/map data.

## Decisions

### 1. Redirect, don’t keep a hidden tab

**Choice:** Drop the Info Sheets `NavLink`. `App.tsx` sends `/survey-bot` (and old forms aliases) to `/events-info`. MorphUtils Event Logs description drops Info Sheets.

**Why:** User said the module is not needed. Redirect avoids dead bookmarks.

**Alternative:** Keep SurveyBot unlinked — still ships a product surface.

### 2. One graphs contract, many prompts

**Choice:** Keep `morphai.VisualFirstInstructions`. Mirror the same string as `VISUAL_FIRST_INSTRUCTIONS` in `pkg/morphai-rs` and put it into Rust `tool_follow_up_prompt` (today those follow-ups are prose-only). Add Morph AI builtin skill `builtin-graphs` (seed like Research/Design). Append the constant to first-turn prompts: Event Logs `formsXAssistantInstructions`, Content Maker `tranMailAssistantInstructions`, Data Access `DATA_AI_INSTRUCTIONS`, Project `ENGI_INSTRUCTIONS`, MorphNotes `textAssist*` (allow mermaid fences), `caseTaskAIDraftPrompt`, research compose/write prompts.

**Why:** User asked for a common skill anywhere. Go follow-ups already have VisualFirst; Rust Data Access / Project follow-ups do not. Duplicating essays per app drifts.

**Alternative:** Only the Morph AI skill — MorphUtils first-turn and Rust follow-ups would still skip graphs.

### 3. Task document is markdown in description

**Choice:** AI returns `markdown` (accept legacy `description`). Create form stores it in `description`. Markdown tab shows/edits that document. HTML tab uses `buildCaseTaskHTML` / `darkPreviewSrcDoc`; `markdownToHTMLFragment` MUST keep mermaid fence language (`code.language-mermaid`) so the preview draws diagrams. Do not add a SQLite column. Create POST sends title, times, description=markdown; empty location/detail.

**Why:** No schema migration; CaseTask already has description. Today the HTML fragment drops fence language, so mermaid would stay as a code block.

**Alternative:** New `markdown_content` column — extra migration for a field we already have.

### 4. Create vs edit

**Choice:** Hide extra widgets when `!editing`. Edit keeps map/JSON/attachments for old rows.

**Why:** Spec says create-only; historical tasks would otherwise lose data.

## Risks / Trade-offs

- [Public Info Sheet links still exist] → Operator cannot manage them in UI; acceptable until a later cleanup.
- [Data AI “analyze” template vs graphs] → Keep report headings; add mermaid when quantities exist.
- [Notes were “plain text, no fences”] → Allow mermaid; still no replacement title.

## Migration Plan

1. Ship UI redirects + prompt/skill + create-form together.
2. Rollback: restore Info Sheets tab; drop builtin-graphs seed (existing rows remain); restore create fields.

## Open Questions

None.
