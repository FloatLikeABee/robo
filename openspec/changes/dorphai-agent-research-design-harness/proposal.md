## Why

Dorph AI already has a tool loop, but it still behaves like a model caller: builtin skills are a thin catalog (concise / DorphData / knowledge), research and design are not default, documents and prompts are only keyword-routed, and finished sessions are discarded. Operators need an agent that analyses each request against available materials and keeps lessons from significant sessions.

## What Changes

- Treat the existing management tool loop as the default Dorph AI **agent** (not a new runtime). Always load full instruction bodies for default **Research** and **Design** skills (plus a short request-analysis rule).
- Seed those skills even when the skills table is already populated (insert-if-missing by id).
- Every chat request gets a short analysis contract: goal, which documents/workspace/knowledge apply, unknowns — then tools or the answer. No extra model family.
- After **significant** sessions only (enough turns, tools, or document context — not greetings), distill one durable lesson and inject matching lessons on later chats.
- Dark-only. Skills picker still adds extra skills; Research/Design are on by default. No theme switch.

## Capabilities

### New Capabilities

- `dorphai-default-agent-skills`: Always-on Research and Design skill bodies, plus request analysis of prompts and document materials.
- `dorphai-session-harness`: Harvest significant sessions into durable agent lessons used on later chats.

### Modified Capabilities

- None. `openspec/specs/` has no archived Dorph AI agent-skill baseline.

## Impact

- `dorph/handlers/skills.go` (ensure builtins, always inject default bodies)
- `dorph/handlers/management_chat.go` (analysis contract, lesson context)
- `dorph/handlers/chat.go` / persist path (significance + background distill)
- SQLite `agent_lesson` (or equivalent) via `dorph/db`
- Tests for seed, default injection, significance heuristic
- Optional: Skills UI lists learned lessons (read-only). No other products.
