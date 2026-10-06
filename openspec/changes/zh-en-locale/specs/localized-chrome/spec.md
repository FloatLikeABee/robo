## Purpose

Shows Morph AI, MorphNotes, MorphTools, Event Logs, Content Maker, and Project chrome in the active language without rewriting the operator’s own content.

## ADDED Requirements

### Requirement: Chrome follows the active language
Navigation, buttons, empty states, placeholders, and errors shown by Morph AI, MorphNotes, MorphTools, Event Logs, Content Maker, and Project MUST appear in the active UI language. A missing translation MUST show the English string.

#### Scenario: Chinese empty state
- **WHEN** the UI language is Simplified Chinese and a screen has an empty state
- **THEN** that empty state is Simplified Chinese

#### Scenario: Missing translation
- **WHEN** a label has no Simplified Chinese string
- **THEN** the label is shown in English

### Requirement: Brand names stay in Latin script
The names Morph AI, MorphNotes, and MorphTools MUST stay in Latin script in both languages. Other chrome, including Event Logs, Content Maker, and Project, MUST be translated when the UI language is Chinese.

#### Scenario: MorphNotes title in Chinese
- **WHEN** the UI language is Simplified Chinese and MorphNotes is open
- **THEN** the product title is still MorphNotes

### Requirement: Stored content is not translated
Notes, chat messages, RAG passages, skill bodies, and other text the operator or an assistant already wrote MUST stay as stored. The language control MUST NOT rewrite that text.

#### Scenario: Existing note
- **WHEN** the operator switches the UI to Simplified Chinese and opens a note written in English
- **THEN** the note body is still English

### Requirement: Assistant replies follow the UI language
Morph AI MUST ask the model to reply in Simplified Chinese when the UI language is Chinese, and in English when the UI language is English, unless the operator’s message is in the other language. The product MUST treat any other locale value as absent and MUST NOT add a reply-language instruction for it.

#### Scenario: Chinese UI
- **WHEN** the UI language is Simplified Chinese and the operator sends a message in Chinese
- **THEN** the request asks the model to reply in Simplified Chinese

#### Scenario: User writes in the other language
- **WHEN** the UI language is English and the operator writes the message in Chinese
- **THEN** the request allows the reply to follow the message language

#### Scenario: Unknown locale value
- **WHEN** a chat request carries a locale other than `en` or `zh`
- **THEN** the request does not add a reply-language instruction
