# Proposal

## Why

Dorph already asks models to write notes, grill drafts, and generated HTML/MD, but it has no shared HTML craft skill. html-anything (Apache-2.0) is the quality bar for single-file HTML. Local Ollama cannot swallow the upstream repo, so Dorph needs a vendored, attributed, compact form plus Ge's dark sans overlay on every HTML-producing path.

## What Changes

- Vendor html-anything's skill text and needed assets under `skills/vendor/html-anything/`, keeping the Apache-2.0 LICENSE and attribution. Do not vendor the Next.js app or the 75 template skills.
- Add a Dorph-facing `SKILL.md` plus a condensed injection form. Layer Ge's defaults on top: Arial (or similar sans), futuristic dark mode, clear content, varied colours.
- Expose one helper in `pkg/dorphai` that every HTML-producing path loads: note AI output, generated HTML/MD, grill final content, and agent-facing MCP/skills.
- Expose the skill in the dorph-mcp skills pack and `docs/agents/16`.
- State a token budget for the injected text and enforce it with a test so local Ollama stays usable.

## Capabilities

### New Capabilities

- `html-anything-base-skill`: Vendored html-anything skill, Ge overlay, compact `pkg/dorphai` injection, and a prompt-size budget for HTML-producing AI paths.

### Modified Capabilities

- `dorph-generated-document-theme`: HTML/MD generation prompts MUST load the html-anything skill and the Ge overlay, not only name a dark sans theme.
- `dorph-for-agents`: The playbook and skills pack MUST name the html-anything base skill and where agents load it.
- `note-ai-grill`: The grill write prompt MUST load the same compact skill so the final draft follows it.

## Impact

- `skills/vendor/html-anything/` (LICENSE, attribution, SKILL.md, condensed assets).
- `pkg/dorphai` prompt/context helpers and tests (budget + content).
- Dorph HTML-producing prompts: note text assist, document shape / generated HTML/MD, form HTML, grill write.
- Agent pack: `.cursor/skills/`, `docs/agents/16-dorph-for-agents.md` (and the embedded `dorph/mcp/agent_guide.md`).
- No Render, deploy, or Files-workspace changes. No merge of this PR by the agent.
