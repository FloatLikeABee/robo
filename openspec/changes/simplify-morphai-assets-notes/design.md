## Context

See proposal.md for motivation. Current surfaces:

- Morph AI empty state in `SkoolAiChat.js` renders logo plus a tip list ("Here's what you can do:").
- `GET /api/ai-agents` merges Morph DB agents (`source: morph`) with BK assistants (`source: bk`); the sidebar lists both.
- Morph Data uses a combined Resources page (`Resources.js`) with People + Assets tabs; drawer label is Resources; routes redirect people/assets → resources.
- Big Notes create dialog uses a free-text Theme `TextField` (default `dark`, fallback `default` on submit).

## Goals / Non-Goals

**Goals:**

- Minimal empty Morph AI welcome (logo/brand only).
- Assistants picker = AI tools assistants only, with a coherent default when Morph agents disappear from the list.
- Assets-only Morph Data module and nav; People UI and Resources naming removed from primary UX.
- Big Notes create theme = select dark | light.

**Non-Goals:**

- Deleting People data from the database or migrating people records into assets automatically.
- Removing the AI tools product header link or BK assistant management UI.
- Changing Big Notes generation/backends beyond accepting dark/light theme values.
- Redesigning Morph AI chrome beyond the empty welcome and assistants list contents.

## Decisions

### 1. Assistants: filter at API (preferred) + UI safety net
- **Choice**: Change `ListMorphAIAgentsHandler` so the chat agents list returns BK assistants only (or add a query such as `source=bk` used by Morph AI). Frontend also filters `source === 'bk'` so a mixed payload cannot regress the UI.
- **Rationale**: Single contract for the Assistants picker; Morph built-ins stay out of chat selection without leaving dead Morph defaults in localStorage forever.
- **Alternatives**: Frontend-only filter (faster, but API still advertises Morph agents). Keep Morph `general` as fallback when BK is empty — rejected; proposal says assistants are AI tools ones; empty list is acceptable.

### 2. Default agent when Morph agents removed
- **Choice**: Prefer previously stored agent id if it is still in the BK list; otherwise first BK assistant; if none, leave selection empty / no Morph fallback.
- **Rationale**: Avoid silently reintroducing Morph specialists.

### 3. Assets route and nav rename
- **Choice**: Primary route becomes `/assets` (under admin base). Drawer label **Assets**. `/resources` and `/people` redirect to `/assets`. Replace `Resources.js` tabs with the existing Assets/Vehicles page (or rename Resources page to Assets without People tab).
- **Rationale**: Matches "menu should be just assets". Reuses existing assets UI.
- **Alternatives**: Keep `/resources` URL with Assets label only — rejected; user asked for Assets naming, not Resources.

### 4. People module removal scope
- **Choice**: Remove People from nav, Resources tabs, and Data Import "Resources — People" option. Leave backend people APIs untouched for now unless something breaks the Assets page.
- **Rationale**: UX removal first; data migration is out of scope.

### 5. Big Notes theme control
- **Choice**: Replace free-text field with a two-option control (ToggleButtonGroup, Select, or Radio) bound to `'dark' | 'light'`. Default `'dark'`. Stop sending `'default'` fallback.
- **Rationale**: Exactly two selections as requested; matches existing dark default.

### 6. Empty welcome markup
- **Choice**: Keep the neon logo + MORPHAI title block; delete the `<p>Here's what you can do:</p>` and the entire `<ul>` tip list.
- **Rationale**: "just a logo is fine" — brand title already sits with the logo and is harmless.

## Risks / Trade-offs

- [Users relied on Morph specialist agents in Assistants] → Mitigation: they still chat without a specialist; AI tools assistants cover custom setups. Document in release notes if needed.
- [People bookmarks / import flows break] → Mitigation: redirects from legacy paths; remove People import option explicitly so users are not sent into a dead UI.
- [Stored `agentId` points at Morph built-in] → Mitigation: re-resolve selection against BK-only list on load.
- [Backend still generates themes other than dark/light for old notes] → Mitigation: create path only constrains new notes; existing notes unchanged.

## Migration Plan

1. Ship frontend + agents list change together so Assistants never flash Morph agents then disappear.
2. Deploy Assets route/nav; keep redirects for Resources/People.
3. No DB migration required for this change.
4. Rollback: revert frontend/API list filter and restore Resources page/tabs if needed.

## Open Questions

- None that block implementation; People data retention without UI is intentional.
