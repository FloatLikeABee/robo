# Tasks

## 1. Shared dark tiles

- [x] 1.1 Replace the flat `#5c5182` card fill with page `#221c34` and tiles `#322a4a`, `#3d3148`, and accent `#5a3d62`, using 16px gap and radius, and verify the page hex is darker than the tile hexes

## 2. Module primary screens

- [x] 2.1 Lay Dorph AI’s transcript and workspace out as bento tiles and verify a wide viewport shows them side by side
- [x] 2.2 Apply the same tiles, gap, and radius to DorphNotes, Content, Event Logs operator cards, and DorphTools, and verify Content’s preview and source sit in two columns on a wide screen
- [x] 2.3 Show a skeleton shaped like the cards on first list load in those modules, and verify the empty loading state is not only a spinner

## 3. Sample rules

- [x] 3.1 Keep navigation items as icon plus label with no third level, and clamp long card body text behind Read more, and verify a long card shows that control
- [x] 3.2 Run `openspec validate dark-bento-modules --strict` and verify the change is valid
