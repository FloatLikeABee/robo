## Purpose

Stops MorphNotes Timelines HTML and Big Notes preview from stacking two vertical scrollbars, keeping one inner dark-themed scrollbar for the page content.

## ADDED Requirements

### Requirement: Timelines HTML has one vertical scrollbar
On MorphNotes Timelines HTML view, the outer preview panel MUST NOT show a vertical scrollbar. The HTML page content MUST be the only vertical-scrolling surface. That inner scrollbar MUST use a dark color that matches the MorphNotes dark theme (not the default light browser chrome).

#### Scenario: Long timeline HTML
- **WHEN** an operator opens a Timeline whose HTML is taller than the preview pane
- **THEN** they can scroll the HTML content inside the preview
- **AND** the outer Timelines panel around that HTML does not show a second vertical scrollbar
- **AND** the visible vertical scrollbar is dark-themed

#### Scenario: Short timeline HTML
- **WHEN** an operator opens a Timeline whose HTML fits in the pane
- **THEN** no extra outer vertical scrollbar appears for the HTML view

### Requirement: Big Notes preview has one vertical scrollbar
On MorphNotes Big Notes, the Preview tab (and the HTML tab if it embeds the same page) MUST follow the same rule: outer pane does not vertically scroll; only the inner HTML content scrolls; that scrollbar is dark-themed.

#### Scenario: Long big note preview
- **WHEN** an operator opens a Big Note whose preview HTML is taller than the pane
- **THEN** they scroll inside the preview content
- **AND** the outer Big Notes detail panel does not show a second vertical scrollbar
- **AND** the visible vertical scrollbar is dark-themed

#### Scenario: Big Notes markdown tab
- **WHEN** an operator is on the Big Notes Markdown tab
- **THEN** this requirement does not force removing markdown overflow
- **AND** the Preview/HTML dual-scrollbar rule still applies when those tabs are selected
