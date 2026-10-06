## Context

See proposal.md for why. There is no locale layer today. Morph AI and MorphNotes are one CRA. MorphTools, Event Logs, Content Maker, and Project are separate frontends loaded in iframes. Cookies are not port-specific, but production hosts are separate, so a cookie cannot be the contract. Product UIs stay dark. Chat is Morph AI only.

## Goals / Non-Goals

**Goals:**

- One saved choice per browser, detected once, shared with embedded modules.
- Chrome translated in the six in-scope frontends, with English as the fallback string.
- A reply-language hint that only accepts `en` or `zh`.

**Non-Goals:**

- A translation library, a user-table column, or Traditional Chinese.
- Translating MorphUtils, Data Access, invite signup, or stored content.
- A light/dark control beside the language control.

## Decisions

### 1. A small helper in each frontend, not a shared i18n package

Each in-scope frontend gets the same rules: read a saved `en` | `zh`, otherwise detect, look up a string, fall back to English, set `document.documentElement.lang` to `en` or `zh-Hans`.

**Why not i18next or a new workspace package:** Two languages and no plurals do not need a dependency, and these apps do not already share a UI package.

**Alternative considered:** One cookie on `localhost`. It would sync local ports and still fail across production hosts. Query plus a message works in both places.

### 2. Morph AI is the source of truth when a module is embedded

Save the choice in `localStorage` under `morph-locale` on that origin. MorphNotes reads the same key because it is the same origin.

Cross-origin iframes get `lang=en` or `lang=zh` on the existing embed URL, next to the session token. When the operator switches, the parent posts `{ type: 'morph-locale', lang }` to the frame and sends it again when the frame loads. A frame that is opened on its own, with no `lang` and no saved choice, detects for itself. A `lang` query wins over that frame’s own saved choice so the parent stays in charge.

**Why not reload the iframe:** A reload drops the operator’s place inside the module. A message updates chrome in place.

**Phone:** At 768px and below, the control lives in the existing More menu so the header does not grow another chip. Wider screens show it in the header.

### 3. Catalogs are English keys with a Chinese map

Keep the current English string as the key’s fallback. Every Chinese string has the same key. A test in each frontend fails if a Chinese key is missing. An inline script in each `index.html` applies the saved or detected language before the app paints, so the first paint is not stuck on `lang="en"`.

### 4. Chat locale is a closed field

The chat request may include `locale`. The handler appends a reply-language line only for `en` or `zh`: reply in that language unless the operator’s message is in the other one. Any other value is ignored.

**Why not a client-side prefix in the visible message:** The operator would see the instruction, and a model-facing string does not belong in the transcript.

## Risks / Trade-offs

- [A screen is only partly translated] → Missing keys stay English, and each frontend has a key-parity test. Tasks still have to cover the visible chrome, not a sample.
- [zh-TW browsers see Simplified Chinese] → Accepted. A third language is out of scope.
- [An iframe loads before it listens] → `lang` is on the URL for first paint, and the parent posts again on load.
- [A client sends `locale` as free text] → The server allows only `en` and `zh`.

## Migration Plan

No stored data moves. The first visit detects; after that the saved choice wins. Rollback is reverting the UI: English strings remain the fallback, and old clients that omit `locale` get no reply-language line.

## Open Questions

None. Traditional Chinese, account-level locale, and the leftover shells would change this plan, and they are excluded.
