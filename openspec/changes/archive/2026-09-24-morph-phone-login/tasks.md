## 1. Regression checks first

- [x] 1.1 Add `authKeyboard` unit tests that fail before the helper exists: overlap is 0 when unfocused, when `visualViewport` is missing, or when the overlap is under 120px; a portrait keyboard overlap is returned in CSS pixels; `controlHidden` is true when a control sits below the visible bottom.
- [x] 1.2 Add a login source test that fails on the current page: Sign in is not yet `min-height: 44px` in a login stylesheet, the phone rule does not use `align-content: start`, the error is not `role="alert"`, and `public/index.html` viewport meta is unchanged (`viewport-fit=cover`, no `interactive-widget`).

## 2. Sign-in phone layout

- [x] 2.1 Add `authKeyboard.js` with `authKeyboardOverlap` and `controlHidden` so the unit tests pass.
- [x] 2.2 Add `LoginPage.css` and point `LoginPage.js` at it: phone `align-content: start`, `min-height: 100dvh` shell, 16px fields, 44px Sign in and fields, wrapping `role="alert"` error, safe-area padding, stable username/password names, and a keyboard spacer driven by the helper on focus. Clear the spacer on blur and unmount. Do not edit `App.css`, `index.css`, or the viewport meta.

## 3. Verify

- [x] 3.1 Re-run the Morph frontend unit tests and the production build. Measure `/login` at about 390px: no document horizontal scroll, Sign in at least 44×44, card top-aligned, a long error wraps, and a simulated keyboard overlap makes Sign in reachable. Open the default post-login landing at about 390px and confirm it does not scroll horizontally or leave a modal backdrop.
