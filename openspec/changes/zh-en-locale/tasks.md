## 1. Morph locale helper

- [x] 1.1 Add a locale helper in `morph/frontend` that saves `en` or `zh` under `morph-locale`, detects from the browser language list only when nothing is saved (`zh*` → `zh`, else `en`), lets a `lang` query win, ignores any other value, and sets the document language to `en` or `zh-Hans`
- [x] 1.2 Apply that helper from `index.html` before the app paints
- [x] 1.3 Add an English / 中文 control on the Morph AI header, inside the phone More menu below 768px, and on the login screen

## 2. Morph AI and MorphNotes chrome

- [x] 2.1 Translate Morph AI chat chrome (header, composer, workspace, drawers, empty states, placeholders, shown errors). Keep Morph AI, MorphNotes, and MorphTools in Latin script
- [x] 2.2 Translate MorphNotes navigation and screen chrome the same way, including Event Logs, Content Maker, and Project labels
- [x] 2.3 Add a test that every English catalog key has a Simplified Chinese string

## 3. Embedded modules follow the parent

- [x] 3.1 Add `lang=en` or `lang=zh` to MorphNotes embed URLs and the MorphTools frame URL
- [x] 3.2 When the operator switches language, post `{ type: 'morph-locale', lang }` to each open frame, and post again when a frame loads

## 4. Reply language

- [x] 4.1 Accept `locale` on the chat request and append a reply-language line only for `en` or `zh` (reply in that language unless the message is in the other one)
- [x] 4.2 Test that `zh` and `en` add the line and any other value adds nothing

## 5. MorphTools

- [x] 5.1 Use the same locale rules, early script, and parent message. A standalone visit detects for itself; a `lang` query wins over its own saved choice
- [x] 5.2 Translate MorphTools chrome and add the English/Chinese key-parity test

## 6. Event Logs, Content Maker, and Project

- [x] 6.1 Event Logs: same locale rules, parent message, translated chrome, and key-parity test
- [x] 6.2 Content Maker: same locale rules, parent message, translated chrome, and key-parity test
- [x] 6.3 Project: same locale rules, parent message, translated chrome, and key-parity test
