## Context

See proposal.md for why. The composer already posts `skill_ids` and `agent_id`. `buildEnabledSkillsContext` only copies a selected skill’s instructions when that id is also in `ListAISkills`, and the catalog line says “use when relevant.” Assistants are attached by an Apply button in `AssistantManager.js`, which posts into the chat. `GET /api/ai-agents` already lists AI tools assistants. `buildBKAssistantInstructions` already loads an assistant prompt and queries that assistant’s RAG collections. RAG collections themselves are listed by the AI tools API at `GET /rag/collections` and are not selectable from the composer.

## Goals / Non-Goals

**Goals:**

- Selected skills bind the reply.
- Assistant and RAG selection happen in the composer, in the skills picker pattern.
- A checked RAG collection is retrieved on send.

**Non-Goals:**

- A new skill or RAG editor in the chat. Editing stays in AI tools.
- Multi-assistant turns. One assistant id per send.
- Data Access, or a Files workspace.

## Decisions

### 1. Selected skills are required instructions, not a hint

Load each selected id’s instruction body from the same source the picker uses. Put those bodies under a “follow these selected skills” block. Keep the always-on Research and Design bodies. Stop treating the full enabled catalog as a substitute for a checked skill.

**Why not only “use when relevant”:** That line lets the model ignore a skill the operator just checked. The check is the instruction.

**Why not a second catalog filter:** The picker and the prompt must agree. If the id is in the picker, its body is loaded for the turn even when `ListAISkills` would have skipped it.

### 2. One picker pattern, two new menus

Reuse the skills popover: a composer button, a small menu, a badge for the count. Skills and RAG are checkboxes. Assistant is a single choice plus clear, because the chat API takes one `agent_id`. Do not add a second Apply path.

**Why remove Apply:** Two ways to attach an assistant will drift. The panel keeps create and edit. The composer is the only attach control.

### 3. Chat-selected RAG is its own list

Add a Morph route that lists RAG collection names from the AI tools API. The send includes those names. The handler queries them with the existing RAG query helper and appends snippets, including when no assistant is selected. If an assistant is also selected, query the assistant’s collections and the checked ones, and keep the existing snippet cap.

**Why not only the assistant’s RAG:** The operator asked to pick RAG in the chat the way they pick skills. An assistant’s collections stay attached to that assistant. The picker is the extra, explicit set.

## Risks / Trade-offs

- [AI tools is down, so the assistant and RAG menus are empty] → Show the empty state. The send still works without them.
- [Many collections make a long prompt] → Keep the current per-snippet and total caps. Do not raise them in this change.
- [A selected skill has an empty instruction body] → Include its name and description and do not pretend the body was applied.

## Migration Plan

- No data migration. Existing applied-assistant session storage can stay; the picker reads and writes that same selection.
- Rollback is restoring the Apply button and the old skill prompt wording.

## Open Questions

None. One assistant. Many skills. Many RAG collections. Snippet caps stay as they are.
