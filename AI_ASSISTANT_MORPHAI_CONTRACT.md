# MorphAI Cross-Project Assistant Contract (Phase 1)

This document defines a shared contract for AI assistants in:

- FormsX (`tranform`)
- TranMail (`tranmail`)
- SharpReport (`SharpReport`)
- Booki (`booki`)
- UsersPanel (`UsersPanel`)

## Goals

1. Keep assistant behavior consistent across projects.
2. Support conversational requirement gathering before writes.
3. Preserve existing APIs while adding assistant capabilities safely.

## Request Shape

```json
{
  "messages": [
    { "role": "user", "content": "create form name: Survey A" }
  ],
  "state": {
    "intent": "create_form",
    "fields": {
      "name": "Survey A"
    }
  }
}
```

## Response Shape

```json
{
  "assistant_message": "I can create the form. Please provide: slug",
  "intent": "create_form",
  "missing_fields": ["slug"],
  "state": {
    "intent": "create_form",
    "fields": {
      "name": "Survey A"
    }
  },
  "completed": false
}
```

## Behavioral Rules

1. If user intent is write/create/update, detect required fields.
2. If required fields are missing, ask follow-up questions.
3. Only execute write when required fields are complete.
4. Keep `state.intent` + `state.fields` to continue multi-turn conversation.
5. Support general Q&A outside platform-specific actions.

## Current Phase Status

- `tranform`: assistant endpoint runs safe create flows for forms/contacts and reads Events & Info (MongoDB); `GET /api/v1/ai/mcp-tools` documents HTTP tools for Morph-style clients.
- `tranmail`: assistant endpoint executes safe create flows for templates/contacts.
- `SharpReport`: assistant endpoint provides stateful clarification for report creation.
- `booki`: assistant endpoint provides stateful clarification for create intents.
- `UsersPanel`: assistant endpoint provides stateful clarification for users/roles.

## Shared infrastructure (Phase 0)

- **Model config:** `MORPH_AI_API_KEY`, `MORPH_AI_MODEL` (default `qwen3-max`) — see `DEVELOPER_BASELINE.md`.
- **Go client:** `pkg/morphai/`
- **Rust client:** `pkg/morphai-rs/`
- **Chat UI:** `platform-chat/` (`@robo/platform-chat`)
- **Sessions (spec):** [`AI_ASSISTANT_SESSIONS_API.md`](./AI_ASSISTANT_SESSIONS_API.md)

## Current Phase Status

- `formx`: assistant + **Survey Bot** (`survey bot …`, `/survey-bot` module, `ui_blocks` MCP-app widgets).
- `composerx`: assistant + reference RAG; templates enqueue GraphRAG outbox.
- `morph`: Knowledge Library + `/api/graph/search`; entity detail writes enqueue outbox; `morphgraph-worker` for Neo4j backfill/daily sync.
- `SharpReport` / `booki` / `UsersPanel`: assistant endpoints as before.

## Survey Bot + UI blocks (FormsX)

When intent is `survey_bot` (user message contains “survey bot”), FormsX may return:

```json
{
  "assistant_message": "Please select your gender.",
  "intent": "survey_bot",
  "state": { "intent": "survey_bot", "survey_bot": { "status": "awaiting_ui", "answers": {} } },
  "ui_blocks": [
    {
      "type": "mcp_app",
      "widget": "select",
      "id": "gender",
      "label": "Gender",
      "options": [{ "value": "female", "label": "Female" }],
      "submit_as": { "field": "gender" }
    }
  ]
}
```

Clients (`platform-chat`) render `ui_blocks` and submit `survey_bot_answer:field=value` as the next user message.

## GraphRAG / Morph Knowledge

Morph exposes `POST /api/graph/search` and Morph-only Knowledge Library uploads (`/api/knowledge/files`). Prefer graph search for uploaded knowledge questions. See `docs/MORPH_GRAPH_RAG_PLAN.md` and `docs/SURVEY_BOT_AND_GRAPHRAG_PLAN.md`.

## Next Phase

1. Neo4j entity sync worker + daily reconcile (Morph / FormsX / ComposerX).
2. Full CRUD tool execution for all major modules in each project.
3. Add confirmation and audit trails for destructive actions.
