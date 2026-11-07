# Tasks

## 1. Tokens and preset

- [x] 1.1 Add `html[data-chat-theme='twilight']` token block in `App.css` (page, surfaces, purple/orange accents, glass borders)
- [x] 1.2 Add dispersion background on `.skool-ai-chat` (static gradients + optional slow animation; respect `prefers-reduced-motion`)
- [x] 1.3 Add `@media (prefers-reduced-transparency: reduce)` and `@supports` fallbacks for opaque surfaces

## 2. SkoolAiChat wiring

- [x] 2.1 Default theme to `twilight`; migrate stored `dark` / `light` keys to `twilight` once
- [x] 2.2 Apply glass classes to header, sidebar, main column, composer (backdrop-filter + token backgrounds)
- [x] 2.3 Tune user vs assistant bubble styles for contrast on glass column

## 3. Shared chat UI cleanup

- [x] 3.1 Replace hard-coded blues in chat drawers/modals/hybrid context with chat CSS variables
- [x] 3.2 Style `chat-reasoning-block` / thinking `<details>` to match frosted secondary surface
- [x] 3.3 Verify embedded `.skool-ai-chat--embedded` still fills host without double margins

## 4. Admin token alignment

- [x] 4.1 Update `theme.js` primary/secondary and table header colors to purple/orange twilight pair
- [x] 4.2 Comment that MUI `mode: 'dark'` tracks chat twilight (not legacy blue)

## 5. Docs and policy

- [x] 5.1 Update `openspec/project.md` chat UI line: single twilight glass preset (supersedes flat dark-only wording)
- [x] 5.2 Optional: one screenshot note in `deploy/README.md` or agent doc if visual QA checklist exists

## 6. Verification

- [ ] 6.1 Manual: standalone Dorph AI at desktop + 375px width (sessions sheet, send, reasoning block)
- [ ] 6.2 Manual: DorphNotes embedded chat inherits twilight
- [x] 6.3 Run existing frontend tests (`npm test` in `dorph/frontend`) — fix snapshots only if intentional
- [x] 6.4 `openspec validate dorphai-futuristic-glass-ui --strict`
