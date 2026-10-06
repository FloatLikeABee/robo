## Purpose

Lets a person use English or Simplified Chinese, starting from the browser language, and keeps that choice across the product screens they open.

## ADDED Requirements

### Requirement: Two UI languages
The product MUST offer exactly two UI languages: English (`en`) and Simplified Chinese (`zh`). The Chinese control MUST be labeled 中文. The product MUST NOT offer Traditional Chinese as a separate language.

#### Scenario: Language control lists both choices
- **WHEN** the operator opens the language control
- **THEN** the choices are English and 中文

### Requirement: First visit follows the browser
When this browser has no saved language choice, the product MUST select `zh` if any language in the browser’s preferred list starts with `zh`, including `zh-TW`, `zh-HK`, and `zh-Hant`. Otherwise it MUST select `en`. The product MUST NOT use IP address, account country, or a stored user profile to choose the language.

#### Scenario: Simplified Chinese browser
- **WHEN** the operator opens the product for the first time and the browser’s preferred languages start with `zh-CN`
- **THEN** the UI language is Simplified Chinese

#### Scenario: Traditional Chinese browser
- **WHEN** the operator opens the product for the first time and the browser’s preferred languages start with `zh-TW`
- **THEN** the UI language is Simplified Chinese

#### Scenario: Other browser language
- **WHEN** the operator opens the product for the first time and no preferred language starts with `zh`
- **THEN** the UI language is English

### Requirement: Explicit choice wins
Choosing English or 中文 MUST save that choice on this browser and MUST use it on the next visit instead of detecting again. The control MUST work on the login screen before anyone is signed in.

#### Scenario: Saved English on a Chinese browser
- **WHEN** the operator chooses English and later opens the product again in a browser whose preferred language starts with `zh`
- **THEN** the UI language is English

#### Scenario: Switch before sign-in
- **WHEN** the operator is on the login screen and chooses 中文
- **THEN** the login screen is shown in Simplified Chinese

### Requirement: One choice covers embedded modules
The language chosen in Morph AI MUST be the language used by MorphNotes, MorphTools, Event Logs, Content Maker, and Project, including when those modules are open inside Morph AI. Changing the language MUST update an already open module without signing in again.

#### Scenario: MorphNotes modal follows the chat
- **WHEN** the operator chooses 中文 in Morph AI and opens MorphNotes
- **THEN** MorphNotes chrome is Simplified Chinese

#### Scenario: Open module updates
- **WHEN** Event Logs, Content Maker, Project, or MorphTools is already open and the operator switches language in Morph AI
- **THEN** that module’s chrome switches to the new language

### Requirement: Document language matches the choice
The page MUST set its document language to `zh-Hans` when the UI language is Chinese and to `en` when the UI language is English.

#### Scenario: Chinese document language
- **WHEN** the UI language is Simplified Chinese
- **THEN** the document language is `zh-Hans`
