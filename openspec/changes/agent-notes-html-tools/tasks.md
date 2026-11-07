# Tasks

## 1. Agent Notes

- [x] 1.1 Select the first note when the list load returns rows, and verify a test that an empty selection after load becomes the first id while a later back does not reselect

## 2. DorphNotes HTML colors

- [x] 2.1 Move `htmldoc`, Contents & Data publish, the preview override, case-task HTML, and the form prompt to page `#221c34`, card `#322a4a`, and accents `#8b5cf6` `#a78bfa` `#f97316`, and verify the theme and publish tests expect those hexes

## 3. DorphTools

- [x] 3.1 Remove the Graphic documents tab from Documents and verify Articles and ScholarForge remain and the generator is not mounted
- [x] 3.2 Set the DORPHTOOLS title to solid `#f4f0fa` without a clipped gradient, and verify the header style has no `WebkitTextFillColor: 'transparent'`
- [x] 3.3 Run `openspec validate agent-notes-html-tools --strict` and verify the change is valid
