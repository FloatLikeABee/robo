## 1. Header overflow

- [x] 1.1 Add a failing test that the phone More menu lists Skills, AI tools, MorphNotes, Clear chat, and MorphUtils only when a link is provided, and that the backdrop closes it
- [x] 1.2 Implement the More menu and wire it into the Morph AI header with the existing chip handlers
- [x] 1.3 Add phone CSS so the inline chips hide, More shows, header actions do not clip or scroll, and header controls are at least 44×44 CSS pixels

## 2. Phone sheets

- [x] 2.1 Add a failing test for sheet width (`100%`, not `100vw` or a narrow rail), session focus cycling, and Escape dismiss
- [x] 2.2 Make the MorphNotes navigation drawer a full-width phone sheet with 44px rows and a 44px close control
- [x] 2.3 Make the Morph AI sessions sidebar a full-width phone dialog with an in-sheet close control, focus held inside, and no sideways background scroll
- [x] 2.4 Size case/task and story drawer papers with `100%` at the `xs` breakpoint, and give hybrid sheet close controls a 44×44 phone target

## 3. Verification

- [x] 3.1 Run Morph frontend unit tests and the production build
