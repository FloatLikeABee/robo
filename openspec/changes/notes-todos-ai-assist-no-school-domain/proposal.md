## Why

Notes & TODOs **AI assist** still tells the model it is writing for a transportation / school operations staff member. A title like “make money on AI stock” therefore comes back as bus-route / student / semester copy from the old Skool product — the title is treated as decoration, not the topic.

## What Changes

- Rewrite the Notes & TODOs generate (and related text-assist) prompts so the **user title / seed is the only domain**. No school, bus, student, parent-portal, or transportation persona unless the user wrote that.
- Empty title still MUST NOT invent school-operations examples.
- Keep the same `/api/tran/text-assist` API and the same AI assist button. Output still fills the body only.
- Add tests so leftover school-ops prompt strings cannot return unnoticed.

## Capabilities

### New Capabilities

- `notes-todos-ai-assist`: Notes & TODOs AI assist drafts or polishes text from the user’s title (and body when improving), with no leftover school-transportation domain.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- Morph backend: `morph/handlers/tran_text_assist.go` (`generate_todo`, `generate_note`, `task_chain_step` prompts). Same POST `/api/tran/text-assist`.
- Morph frontend: `NotesTodosContent.jsx` only if the request omits title/seed (should already send `seed: title`).
- Tests: new Go tests on prompt construction (no live model required for the leftover-string check).
