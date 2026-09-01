## Purpose

Makes MorphUtils Project Markdown and HTML preview vertical scrollbars match the dark product theme instead of default light browser chrome.

## ADDED Requirements

### Requirement: Project markdown and HTML scrollbars are themed
In MorphUtils Project document preview, both the Markdown tab and the HTML tab MUST use dark-themed vertical scrollbars that fit the Project dark UI. Default light OS/browser scrollbar chrome MUST NOT be the visible style on those preview surfaces.

#### Scenario: Long markdown in Project
- **WHEN** an operator opens a Project document Markdown preview that overflows vertically
- **THEN** the vertical scrollbar is dark-themed to match the Project UI

#### Scenario: Long HTML in Project
- **WHEN** an operator opens a Project document HTML preview that overflows vertically
- **THEN** the vertical scrollbar is dark-themed to match the Project UI
- **AND** if the HTML document itself scrolls, that inner scrollbar is also dark-themed
