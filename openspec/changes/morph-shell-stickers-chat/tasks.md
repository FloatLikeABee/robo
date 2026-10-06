## 1. Always-visible workspace

- [x] 1.1 Remove the Hide workspace / Show workspace control and stop applying the collapsed workspace layout on wide and phone screens
- [x] 1.2 Ignore a saved workspace-closed preference so the pane stays visible, including the phone lower band
- [x] 1.3 Put Context & Knowledge before Notes & TODOs, and select Context & Knowledge when the session has no stored tab
- [x] 1.4 Update workspace-open and phone-width checks so a closed workspace is no longer the expected phone state, and a stored Notes tab still opens Notes

## 2. Long assistant replies

- [x] 2.1 Add an enlarge control on an assistant reply whose rendered height is greater than ten line-heights, and leave shorter replies and user messages without it
- [x] 2.2 Open that reply in a dark near-full-viewport modal that scrolls, and dismiss it with Escape, backdrop click, or a close control without clipping the transcript
- [x] 2.3 Add a small check that the ten-line threshold uses rendered height, not newline count, and that a shorter reply does not qualify

## 3. AI tools loading

- [x] 3.1 Replace the white AI tools frame background with a dark surface and a loading indicator until the embedded app finishes loading
- [x] 3.2 Show that dark loading state again each time the modal opens

## 4. Project dark blue

- [x] 4.1 Retint Project page, header, cards, selected tab, primary buttons, and the create modal to the shared dark-blue surfaces and blue accents
- [x] 4.2 Move ordinary delete links off rose; keep rose only on the destructive confirm control

## 5. Stories and Stick notes

- [x] 5.1 Rename the Big notes navigation and page label to Stories without changing its address, and leave Timelines and `/stories` alone
- [x] 5.2 Rename the Tasks navigation, list heading, and create/detail headings to Stick notes without changing its address
- [x] 5.3 Render Stick notes as equal squares in a grid with an 8px gap, title plus a multi-line body, and a color from a twelve-color palette assigned by sorted id so the first twelve differ and search does not recolor
- [x] 5.4 Add search that filters title and body, shows no stickers when nothing matches, and restores the full grid when cleared
- [x] 5.5 Add a small check for sticker color assignment: stable per id order, unique for the first twelve, unchanged when the visible subset changes

## 6. Verification

- [x] 6.1 Run the Morph AI workspace, reply-threshold, and sticker-color checks
- [x] 6.2 Confirm the phone chat column still uses the full width with the workspace band visible
