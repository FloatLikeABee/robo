## Why

The product chrome is English only. People whose browser is set to Chinese get the same English labels, and there is no way to choose a language. The first visit should follow the browser, and an explicit choice should stick.

## What Changes

- Two UI languages: English (`en`) and Simplified Chinese (`zh`, shown as 中文). Traditional Chinese is not a third language.
- On first visit, with no saved choice, pick `zh` when any preferred browser language starts with `zh`, otherwise `en`. Do not use IP, account country, or a server profile.
- A visible English / 中文 control saves the choice on this browser and wins over detection after that. The login screen can switch before anyone is signed in.
- The same choice covers Morph AI, MorphNotes, MorphTools, Event Logs, Content Maker, and Project, including when the last three are inside the MorphNotes modal.
- Chrome is translated: navigation, buttons, empty states, placeholders, and errors the UI shows. User notes, chat history, RAG text, and skill bodies stay as written.
- When the locale is Chinese, Morph AI asks the model to reply in Simplified Chinese unless the user writes in another language. English does the reverse.
- Brand names stay as they are: Morph AI, MorphNotes, MorphTools.

Out of scope: MorphUtils, Data Access, invite signup, API log lines, and translating stored content.

## Capabilities

### New Capabilities

- `locale-choice`: Detect the browser language once, remember an explicit English or 中文 choice, and apply that choice to every in-scope screen, including embedded modules.
- `localized-chrome`: In-scope screens show their own chrome in the active language, and Morph AI replies follow that language unless the user writes in the other one.

### Modified Capabilities

## Impact

- Morph frontend (`morph/frontend`): header control, string maps, `document.documentElement.lang`, embed URLs, chat system hint.
- MorphTools (`bk/frontend`), Event Logs (`formx/frontend`), Content Maker (`composerx/frontend`), Project (`morph-engi/frontend`): the same two languages and the parent’s choice when embedded.
- No new translation dependency and no user-table locale column.
