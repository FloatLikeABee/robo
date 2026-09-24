## 1. Regression check first

- [x] 1.1 Add a Morph frontend unit test that fails on the current shell: viewport meta keeps `width=device-width` and `viewport-fit=cover`; `index.css` does not yet clip horizontal overflow on `html`/`body`/`#root`; the phone `.header-actions` rule still uses `overflow-x: auto`; MorphNotes still gates the home-indicator spacer on `useMediaQuery`; `/skills` padding does not reference `safe-area-inset`.

## 2. Fit the shell

- [x] 2.1 In `index.css`, constrain `html`, `body`, and `#root` to the viewport (`max-width: 100%`, `overflow-x: hidden` then `overflow-x: clip`) and wrap alerts (`.MuiAlert-root`, `.alert`, `.skills-panel-alert`) with `max-width: 100%` and `overflow-wrap: anywhere`.
- [x] 2.2 In `App.css`, give the chat shell column `min-width: 0` and `max-width: 100%`, stop the phone header actions from scrolling horizontally (wrap, shrink, no `overflow-x: auto`), ellipsize the title, and use `overflow-wrap: break-word` on message bubbles and `anywhere` on `.error-bubble`.
- [x] 2.3 Keep chat header and composer safe-area insets, including left/right, without padding `#root`. Add safe-area padding on `.skills-panel-page`. On the login error line, set `max-width: 100%` and `overflow-wrap: anywhere`.
- [x] 2.4 In `AdminLayout.js`, show the home-indicator spacer with the `xs`/`sm` display breakpoint instead of `useMediaQuery`, and set the phone top bar min height to `calc(56px + env(safe-area-inset-top))`.
- [x] 2.5 At the `xs` breakpoint, change case/task and timeline drawer paper from `100vw` to `100%` with `maxWidth: '100%'`. At `max-width: 768px`, keep `.app.app--agent .chat-sidebar` `position: fixed` with `min-width: 0` so the 80px rail cannot re-enter the grid.

## 3. Verify

- [x] 3.1 Re-run the unit test and confirm it passes. Run the Morph frontend production build (`CI=true npm run build` in `morph/frontend`). Measure a phone-width layout (about 390px) so document `scrollWidth` does not exceed `clientWidth` on the sign-in shell and the chat shell markup.
