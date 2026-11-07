# Design

## Context

See proposal.md. The session rail is `chat-sidebar` in `SkoolAiChat.js`. Provider metadata already exists as `pkg/dorphai.ListProviders()`, including Ollama with no key required. Chat today builds the client with `dorphai.LoadFromEnv()` in `dorph/ai/ai.go`. There is no settings screen and no per-user provider row.

## Goals / Non-Goals

**Goals:**

- One settings control, one modal, one saved provider per user.
- Scan through the Dorph API, not from the browser.
- Fall back to the environment provider when the user has not saved one.

**Non-Goals:**

- Several saved profiles.
- Changing Event Logs, Content, or DorphTools environment keys.
- A new provider id.

## Decisions

### 1. One row, server-side

**Choice:** SQLite table `user_ai_provider` keyed by the signed-in user id: provider id, base URL, API key, model. `GET` returns provider, URL, model, and `key_set: true`, never the key. `PUT` replaces the row. An empty key on update keeps the stored key.

Rejected: `localStorage`. The key would sit in the browser and other modules could not use the same session safely. Rejected: a list of profiles. The request is one setup.

### 2. Scan is a server call

**Choice:** `POST /api/chat/provider/scan` with provider, URL, and optional key. The handler calls `GET {base}/models` for OpenAI-compatible providers and `GET {base}/api/tags` for Ollama, using the key as a bearer when present. It returns `[{id}]`. The browser never calls the model host.

Rejected: scanning from the page. Local and cloud hosts block browser CORS, and the key would be sent from the page to an arbitrary origin.

### 3. Chat uses the row only when it is complete

**Choice:** if the user has a saved provider and model, that chat request builds the client from the row. Otherwise `LoadFromEnv()` stays. A scan error does not delete the row.

### 4. Where the control sits

**Choice:** a gear button after the session list, pinned to the bottom of the rail. The modal is the existing dialog pattern (portal, Escape closes). Provider select, URL, key, Scan, model select, Save.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Scan is used to probe internal hosts | Require a signed-in user; timeout the outbound call; do not return response bodies, only model ids |
| Key logged | Do not log the key or the scan URL with the key |
| Ollama model list shape differs | Map `/api/tags` names into the same `{id}` list |

## Migration Plan

New table, created on startup like the other SQLite tables. No backfill. Rollback drops the UI; env chat still works.

## Open Questions

None.
