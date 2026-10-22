# dorph-generated-document-theme Specification

## Purpose

Gives AI-generated Dorph documents one dark, readable, multi-accent look instead of a serif page or a single hue.

## Requirements

### Requirement: Generated documents use a dark sans theme
HTML documents Dorph generates for stories, research, timelines, case tasks, and project documents MUST use an Arial or equivalent sans-serif font stack, a dark background, and light body text with enough contrast to read. Headings MUST be distinct from body text. The document MUST NOT use a serif body face or a light page background, including when a stored theme name asks for a light theme.

#### Scenario: Story or research HTML
- **WHEN** Dorph renders a story, research note, timeline, or case task as HTML
- **THEN** the document font stack includes Arial
- **AND** the page background is dark and the body text is light

#### Scenario: Light theme name
- **WHEN** a generated note is stored with a light theme name
- **THEN** the rendered HTML is still the dark theme

### Requirement: Accents vary
A generated HTML document MUST use at least three distinct accent colors for links, headings, and callouts or charts. It MUST NOT style the whole document with only one accent hue.

#### Scenario: Accent variety
- **WHEN** Dorph renders a generated HTML document
- **THEN** the stylesheet contains at least three different accent colors

### Requirement: Generation prompts ask for that theme
When Dorph asks a model to write an HTML document, the prompt MUST request the dark sans multi-accent theme and MUST NOT request a light page or a single-accent palette.

#### Scenario: HTML generation prompt
- **WHEN** a prompt asks the model to produce an HTML page
- **THEN** the prompt names a dark theme, a sans-serif font, and more than one accent color

### Requirement: Chat markdown stays dark and sans
Assistant markdown rendered in Dorph AI chat MUST use a sans-serif font stack and the existing dark chat colors.

#### Scenario: Assistant markdown
- **WHEN** Dorph AI renders an assistant markdown reply
- **THEN** the markdown block uses a sans-serif font stack
- **AND** it does not switch the chat to a light background

### Requirement: Generation prompts load the html-anything skill
When Dorph asks a model to write HTML or markdown that becomes a generated document, the prompt MUST include the compact html-anything skill and Ge's Arial dark multi-accent overlay. It MUST NOT request a light page or a single-accent palette.

#### Scenario: Form HTML prompt
- **WHEN** Dorph builds the form HTML generation prompt
- **THEN** the prompt includes the compact html-anything skill
- **AND** it names a dark theme, a sans-serif font, and more than one accent color

#### Scenario: Generated markdown prompt
- **WHEN** Dorph builds a timeline, research, or story generation prompt
- **THEN** the prompt includes the compact html-anything skill
- **AND** it still asks for markdown that the server can render, not raw HTML unless the path is form HTML

### Requirement: Generated header chrome uses Arial
The header meta line and other generated document chrome on stories, research, timelines, big notes, and case-task HTML MUST use the Arial sans stack. Those rules MUST NOT use `system-ui`.

#### Scenario: Big-note header is Arial
- **WHEN** Dorph renders a big note as HTML
- **THEN** the header meta font stack includes Arial
- **AND** the stylesheet does not set the header font to `system-ui`

### Requirement: Big-note generate stays on the idea
When Dorph asks a model to write a big note from an idea, the prompt MUST tell the model to use only facts named in the idea and MUST forbid inventing features, products, agent types, or capabilities the idea does not name.

#### Scenario: Idea-grounded prompt
- **WHEN** Dorph builds the big-note generate prompt
- **THEN** the prompt includes the idea text
- **AND** it forbids inventing features not in the idea

### Requirement: Big notes have a minimum depth
A generated big note MUST have at least three `##` sections with body text under them. The generate prompt MUST ask for that depth. A one-paragraph reply MUST be expanded from the idea (retry once, then a local expand) so the stored markdown is not a single paragraph.

#### Scenario: Prompt asks for sections
- **WHEN** Dorph builds the big-note generate prompt
- **THEN** the prompt requires at least three `##` sections

#### Scenario: One paragraph is expanded
- **WHEN** the model returns a single paragraph about the idea
- **THEN** the stored markdown has at least three `##` headings
- **AND** the headings or body still use words from the idea

### Requirement: Heading count ignores ###
A big note is thin unless it has at least three line-start `## ` headings. `###` headings MUST NOT count. If the note is still thin after one regenerate, Dorph MUST append deterministic `##` sections built from the idea.

#### Scenario: Two ## and many ### is still thin
- **WHEN** the markdown has `## Introduction`, `## Key Milestones`, and several `###` headings
- **THEN** the stored markdown has at least three `## ` headings

### Requirement: Big-note timeout is retryable
Big-note generate MUST wait longer than 90 seconds for Ollama. When the wait expires, the API MUST return a clean retryable error, not `context deadline exceeded`.

#### Scenario: Deadline becomes a retry error
- **WHEN** the model call hits the generate deadline
- **THEN** the error text tells the caller to retry
- **AND** it does not contain `context deadline exceeded`
