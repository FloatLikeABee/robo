# MorphAI Assistant Sessions API (Phase 0 spec)

This document defines **server-side chat session persistence** for satellite platform assistants. Implementation is planned per app in Phase 1+; this spec is the shared contract.

**Related:** [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](./AI_ASSISTANT_MORPHAI_CONTRACT.md) (message/state shape for `/assistant/chat`).

---

## Goals

1. Persist assistant conversations per authenticated user across page reloads.
2. Align with Morph Data session UX (`SkoolAiChat` embedded mode uses session id `default` when `singleSession=true`).
3. Keep chat messages separate from **staff messaging** (`UsersPanel /api/messages/*`).

---

## Authentication

All endpoints require the same JWT/session auth as the host app. Sessions are scoped to `user_id` (or org + user for multi-tenant apps like Booki).

---

## Endpoints

Base path pattern (adjust prefix per app):

```
GET    /api/v1/assistant/sessions
POST   /api/v1/assistant/sessions
GET    /api/v1/assistant/sessions/:id
PUT    /api/v1/assistant/sessions/:id
DELETE /api/v1/assistant/sessions/:id
GET    /api/v1/assistant/sessions/:id/messages
POST   /api/v1/assistant/sessions/:id/messages   # optional; chat may append via POST /assistant/chat
```

TranMail may use `/ai/assistant/sessions`; UsersPanel `/api/assistant/sessions`.

---

## Session object

```json
{
  "id": "default",
  "title": "Create registration form",
  "created_at": "2026-05-19T12:00:00Z",
  "updated_at": "2026-05-19T12:05:00Z"
}
```

- **`default`**: Reserved id for embedded single-session drawers (same as Morph Data).
- **`title`**: Auto-generated from first user message (optional LLM title) or user-editable.

---

## Message object

```json
{
  "id": "msg_01",
  "role": "user",
  "content": "Create a contact form with email and phone",
  "created_at": "2026-05-19T12:00:01Z"
}
```

Roles: `user`, `assistant` (same as MorphAI contract).

Optional metadata on assistant messages:

```json
{
  "intent": "create_form",
  "completed": false,
  "record": {}
}
```

---

## Chat flow with sessions

### Option A — Chat endpoint owns persistence (recommended)

`POST /assistant/chat` accepts optional `session_id`:

```json
{
  "session_id": "default",
  "messages": [{ "role": "user", "content": "..." }],
  "state": { "intent": "", "fields": {} }
}
```

Server:

1. Loads prior messages for `session_id` if client sends only the latest turn (future optimization).
2. Runs assistant pipeline (rules + LLM).
3. Appends user + assistant messages to session store.
4. Returns standard MorphAI response + `session_id`.

### Option B — Explicit message CRUD

Client loads `GET .../sessions/:id/messages` on drawer open; POST chat does not persist (client saves via separate API). Prefer **Option A** for consistency with Morph Data.

---

## Storage by app

| App | Suggested store | Notes |
|-----|-----------------|-------|
| TranDemo | BadgerDB (existing) | Reference implementation |
| TranMail | MongoDB | Same cluster as templates |
| FormsX | MongoDB or MySQL JSON | Match Events & Info store |
| Booki | App MySQL/SQLite | Org-scoped rows |
| SharpReport | App SQLite/Postgres | Report user id from JWT |
| UsersPanel | MySQL | Admin-only assistant |

---

## Retention

- Default: keep last **50 sessions** per user or **90 days**, whichever is stricter (app-configurable).
- `DELETE /sessions/:id` is hard delete for GDPR/user request support.

---

## Frontend integration (`@robo/platform-chat`)

Phase 1 will add props:

```ts
sessionEndpoint?: string;  // base URL for /assistant/sessions
singleSession?: boolean;   // default true → session id "default"
```

Until implemented, clients keep in-memory history via `createPlatformChatController`.

---

## Migration from client-only state

1. Ship session API behind feature flag `ASSISTANT_SESSIONS=1`.
2. On drawer open, `GET /sessions/default/messages` hydrates UI.
3. Remove local-only message arrays from app layouts.

---

_Last updated: Phase 0 foundation._
