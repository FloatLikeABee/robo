## 1. Default agent skills

- [x] 1.1 Ensure builtin Research and Design skills by id (insert-if-missing). Always inject their full instruction bodies into the tool-loop context even when the picker is empty. Add the analysis contract (goal, materials, unknowns) to the agent instructions. No extra LLM round.
- [x] 1.2 Tests: missing builtins are inserted on existing stores; chat context without `skill_ids` still contains Research and Design bodies.

## 2. Session harness

- [x] 2.1 Add `agent_lesson` storage. After persist, distill at most one lesson when the session is significant (≥6 user turns, ≥2 tool rounds, or document context). Skip greetings. Do not duplicate per session.
- [x] 2.2 Inject recent lessons into later chat context. Tests: greeting skipped; significant session stores a lesson; later context includes the rule.

## 3. Verify

- [x] 3.1 `go test` for the new skill/lesson tests. Browser: Skills shows Research and Design; a chat with no picker still behaves as the agent (tools/docs), dark-only.
