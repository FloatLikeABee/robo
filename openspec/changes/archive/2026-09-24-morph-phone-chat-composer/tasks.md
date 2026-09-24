## 1. Tests first

- [x] 1.1 Add a failing unit test for keyboard overlap: positive when the visual viewport is shorter than the layout viewport and offset is 0, zero when the layout viewport already matches, zero when offset accounts for the difference.
- [x] 1.2 Add a failing source test that the last phone CSS block clips transcript inline overflow, pads the composer with left and right safe areas, sizes composer controls to at least 44px, and sizes the shell with `--keyboard-inset`.

## 2. Keyboard inset

- [x] 2.1 Implement the overlap helper and bind it from the Morph AI chat shell, including cleanup and `data-keyboard-open` only while the overlap is positive.
- [x] 2.2 Add `interactive-widget=resizes-content` to the existing viewport meta.

## 3. Phone chat CSS

- [x] 3.1 In the last `max-width: 768px` block, clip the transcript inline axis, constrain prose, welcome, errors, code, and diagrams, and keep an in-block scroll for wide code and diagrams.
- [x] 3.2 In that same block, restore composer side safe areas, wrap composer chips, and raise composer controls and the enlarge close control to at least 44px.
- [x] 3.3 Size the phone shell with `calc(100dvh - var(--keyboard-inset, 0px))`, and hide the agent workspace row while the keyboard is open.

## 4. Verification

- [x] 4.1 Run the Morph frontend unit tests and the production build.
- [x] 4.2 Check the chat at about 390px: wrapped prose, vertical-only transcript scroll, empty and error fit, and the send control stays in view when a keyboard inset is applied.
