## Context

See proposal.md. Specs: `morphai-default-agent-skills`, `morphai-session-harness`.

Today `chatWithManagementTools` is the Morph AI agent loop. `SeedBuiltinSkills` only runs when `ai_skills` is empty and seeds concise / MorphData / knowledge-first — catalog names always, **full bodies only if** `skill_ids` are selected. `inferSubAgents` is keyword heuristics. Sessions persist messages; nothing is distilled into durable agent memory. Dark-only Morph AI; reuse the existing loop.

## Goals / Non-Goals

**Goals:**

- Always-on Research + Design instruction bodies.
- Request analysis of prompt + documents in the same loop (no second model family).
- Harvest **significant** sessions into SQLite lessons; inject recent lessons on later chats.

**Non-Goals:**

- Replacing the management tool loop or adding OS subagents / Cursor-style coding.
- Auto-harvesting every session.
- Per-user lesson UI beyond a read-only list if cheap; no theme switch.
- Changing MorphNotes Research (20→5) or Content Maker.

## Decisions

### 1. Stay on the existing tool loop

**Choice:** Keep `chatWithManagementTools`. Strengthen system prompt + always-injected skill bodies. Do not add a parallel agent runtime.

**Why:** Morph AI already issues JSON tool calls. A second loop is the model-caller problem again.

**Alternative:** New planner/executor — rejected (YAGNI, two agents to debug).

### 2. Ensure builtins by id, always inject Research + Design bodies

**Choice:** Replace empty-table-only seed with `ensureBuiltinSkill(id)` insert-if-missing. New ids: `builtin-research`, `builtin-design`. Always append their instruction bodies in `buildEnabledSkillsContext` (in addition to picker `skill_ids`). Research: graph search first for docs, then allowed web/tool catalog; structure by argument; don’t invent citations. Design: constraints, 2–3 options, recommend, don’t jump to implementation. Analysis: fold into those bodies + a short block in `managementToolInstructions`: goal, materials, unknowns — then JSON tool or answer. **No extra LLM round.**

**Why:** User asked for default perfect research/design and better analysis. Extra round doubles latency.

**Alternative:** Force picker selection — fails “default”. **Alternative:** Rewrite `inferSubAgents` as the analysis — too shallow; keep it as a hint only.

### 3. Significant-session harness

**Choice:** Table `agent_lesson` (`id`, `trigger`, `rule`, `source_session_id`, `created_at`). After persist, if **any** of: ≥6 user messages in the session, tool log ≥2, or files/knowledge were in the applied context — enqueue one background distill (same LLM as ImproveSkill-style JSON: trigger + rule). Skip if a lesson already exists for that `source_session_id`. Inject last ~8 lessons into the agent prompt (newest first). Ceiling: naive distill, one lesson per session. Upgrade: embedding match.

**Why:** User said not every session — only important / high-context ones.

**Alternative:** Operator clicks “Remember” — extra UI, not asked. **Alternative:** Dump full transcripts into the prompt — context blowup.

### 4. Observability

**Choice:** `GET /api/skills` (or a small `GET /api/agent-lessons`) can list learned lessons; Skills page shows a compact “Learned from sessions” list if the endpoint is already in hand. Do not add a second chrome product.

## Risks / Trade-offs

- [Prompt grows with skills + lessons] → Cap lessons at 8; keep skill bodies tight.
- [Distill invents a bad rule] → Lessons are advisory text; operator can delete via Skills later if we expose delete (optional; skip if it expands scope).
- [Existing DBs never got new builtins] → ensure-by-id on startup.
- [Analysis in the same turn is ignored by the model] → Put it in both system instructions and Research/Design bodies.

## Migration Plan

1. Schema + ensure builtins + always inject + analysis sentence.
2. Lesson table + significance + distill + inject.
3. Rollback: stop injecting defaults; table can remain unused.

## Open Questions

None. Significance thresholds (6 turns / 2 tools) can be tuned without changing the spec.
