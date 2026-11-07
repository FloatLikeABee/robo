# Proposal

## Why

Chat still takes its model from server environment variables. An operator cannot point Dorph AI at their own API or a local model, and the session bar has no place to do that.

## What Changes

- Add a settings control at the bottom of the chat sessions bar. It opens a modal.
- The modal lists the existing providers, including Ollama as the local model. It asks for an API URL and an API key. Ollama does not require a key.
- Scan asks that server for its models and lets the operator pick one. Chat uses the saved choice. With nothing saved, chat keeps the current environment provider.
- Store the key on the server for that user. The browser does not keep the raw key, and a later read shows only a mask.

## Capabilities

### New Capabilities

- `chat-provider-settings`: Session-bar settings for one saved AI provider, including a model scan and a local model.

### Modified Capabilities

- None. The existing provider registry stays the list. This change adds the settings screen and the per-user save.

## Impact

- Dorph AI session sidebar and a small settings modal.
- Authenticated chat routes to list providers, save the choice, and scan models.
- One row per user in the Dorph SQLite store.
- No new provider brand and no change to other modules' environment keys.
