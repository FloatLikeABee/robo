# Tasks

## 1. Vendor tree and compact helper (tests first)

- [x] 1.1 Add failing `pkg/morphai` tests that `HTMLSkillInstructions` stays at or under 512 tokens (runes/4 rounding up), names html-anything, Arial or sans, a dark theme, and at least two accent colours, and that `WithHTMLSkill` prefixes a prompt. Verify they fail before the helper exists.
- [x] 1.2 Add a failing test that `skills/vendor/html-anything/` has Apache-2.0 LICENSE, a NOTICE naming nexu-io/html-anything, SKILL.md, and that `condensed.md` matches the morphai embed. Verify it fails on the empty tree.
- [x] 1.3 Vendor LICENSE, NOTICE, upstream shared directives, Morph SKILL.md, condensed.md, and the pkg/morphai helper so those tests pass. Verify `go test ./...` in `pkg/morphai`.

## 2. Wire HTML-producing paths (tests first)

- [x] 2.1 Add failing tests that note generate, `documentShapePrompt`, `BuildFormHTMLPrompt`, and grill `WritePrompt` include the compact skill, and that grill still asks for body-only text. Verify they fail.
- [x] 2.2 Call `morphai.WithHTMLSkill` (or `HTMLSkillInstructions`) from those four paths so the new tests pass and existing prompt tests still pass. Verify `go test` in `pkg/morphai`, `morph`, and `morph/grill`.

## 3. Agent pack and chapter 16 (tests first)

- [x] 3.1 Extend `morph/mcp/agentpack_test.go` so the pack and chapter 16 name html-anything and the vendor path. Verify the test fails.
- [x] 3.2 Add `.cursor/skills/html-anything/SKILL.md`, mention the skill in the Morph notes skill and in `docs/agents/16` writing guidance, and copy 16 onto `morph/mcp/agent_guide.md`. Verify `go test ./mcp -count=1 -run 'TestAgentPack|TestGuide|TestMorphGuide'` passes.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `pkg/morphai` and `morph`, and any Morph frontend test file this change touched (none if frontend was not edited). Verify all pass.

## Workflow follow-up

- Archive the change after implementation and verification.
- Open a PR whose body says `Closes #149`. Do not merge.
