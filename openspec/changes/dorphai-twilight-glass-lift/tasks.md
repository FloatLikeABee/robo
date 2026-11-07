# Tasks

## 1. Token lift

- [x] 1.1 Raise twilight page/panel/sidebar/surface tokens into dusk (lighter than `#14101c`, not near-white)
- [x] 1.2 Raise `--chat-glass-bg` / `--chat-glass-border` and lighten mesh veil so blooms read through chrome
- [x] 1.3 Point reduced-transparency and no-blur fallbacks at the lifted opaque dusk surfaces

## 2. Stronger glass CSS

- [x] 2.1 Increase backdrop-filter on header, sidebar, composer, workspace (target ~28px); assistant bubble ~16px
- [x] 2.2 Add specular inset highlight on those glass regions
- [x] 2.3 Keep user bubble more opaque; confirm assistant/composer text contrast on the new frost

## 3. Admin alignment

- [x] 3.1 Update `theme.js` dark `background.default` / `paper` / table header to the dusk floor

## 4. Verification

- [x] 4.1 Desktop + 375px: mesh visible through header/composer; not a black slab
- [x] 4.2 Embedded DorphNotes chat matches standalone dusk glass
- [x] 4.3 `npm test --watchAll=false` in `dorph/frontend`
- [x] 4.4 `openspec validate dorphai-twilight-glass-lift --strict`
