## Purpose

Gives AI-generated Morph documents one dark, readable, multi-accent look instead of a serif page or a single hue.

## ADDED Requirements

### Requirement: Generated documents use a dark sans theme
HTML documents Morph generates for stories, research, timelines, case tasks, and project documents MUST use an Arial or equivalent sans-serif font stack, a dark background, and light body text with enough contrast to read. Headings MUST be distinct from body text. The document MUST NOT use a serif body face or a light page background, including when a stored theme name asks for a light theme.

#### Scenario: Story or research HTML
- **WHEN** Morph renders a story, research note, timeline, or case task as HTML
- **THEN** the document font stack includes Arial
- **AND** the page background is dark and the body text is light

#### Scenario: Light theme name
- **WHEN** a generated note is stored with a light theme name
- **THEN** the rendered HTML is still the dark theme

### Requirement: Accents vary
A generated HTML document MUST use at least three distinct accent colors for links, headings, and callouts or charts. It MUST NOT style the whole document with only one accent hue.

#### Scenario: Accent variety
- **WHEN** Morph renders a generated HTML document
- **THEN** the stylesheet contains at least three different accent colors

### Requirement: Generation prompts ask for that theme
When Morph asks a model to write an HTML document, the prompt MUST request the dark sans multi-accent theme and MUST NOT request a light page or a single-accent palette.

#### Scenario: HTML generation prompt
- **WHEN** a prompt asks the model to produce an HTML page
- **THEN** the prompt names a dark theme, a sans-serif font, and more than one accent color

### Requirement: Chat markdown stays dark and sans
Assistant markdown rendered in Morph AI chat MUST use a sans-serif font stack and the existing dark chat colors.

#### Scenario: Assistant markdown
- **WHEN** Morph AI renders an assistant markdown reply
- **THEN** the markdown block uses a sans-serif font stack
- **AND** it does not switch the chat to a light background
