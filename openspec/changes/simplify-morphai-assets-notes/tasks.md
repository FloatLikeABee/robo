## 1. Morph AI empty welcome

- [x] 1.1 In `SkoolAiChat.js`, remove the "Here's what you can do:" paragraph and the entire tip `<ul>` from the empty-chat welcome panel
- [x] 1.2 Keep the logo / MORPHAI brand block only; verify empty session shows no quick links

## 2. Assistants = AI tools only

- [x] 2.1 Update `ListMorphAIAgentsHandler` (or equivalent) so Morph AI chat agents listing returns BK/AI tools assistants only
- [x] 2.2 In `SkoolAiChat.js`, filter assistants to `source === 'bk'` and re-resolve default/`agentId` against that list (no Morph built-in fallback)
- [x] 2.3 Remove or hide "AI tools" source badge if the list is BK-only; confirm empty BK list does not resurrect Morph specialists

## 3. Morph Data Assets-only nav

- [x] 3.1 Change `AppDrawer` nav label from Resources to Assets and point selection/routing at the assets path
- [x] 3.2 Update `appRouter.js` so `/assets` is the primary module route; redirect `/resources` and `/people` to `/assets`
- [x] 3.3 Replace Resources People+Assets tabs with Assets-only page (reuse Vehicles/assets UI; stop mounting People)
- [x] 3.4 Update Data Import entity options: remove "Resources — People"; relabel assets import to Assets-oriented naming
- [x] 3.5 Sweep remaining Resources/People primary-nav copy in Morph Data admin UI related to this module

## 4. Big Notes dark/light theme

- [x] 4.1 Replace Big Notes create dialog free-text Theme field with a two-option control (dark | light), default dark
- [x] 4.2 Ensure create payload sends only `dark` or `light` (no `default` fallback)

## 5. Verify

- [x] 5.1 Smoke-check Morph AI empty welcome, Assistants list, Assets nav/route redirects, and Big Notes create theme selection
