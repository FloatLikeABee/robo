# Dorph AI + AI tools Assistants (card UI + shared picker)

**Date:** 2026-08-13  
**Status:** Approved

## Goals

1. AI tools (BK) Assistants page: replace table with good-looking card tiles.
2. Dorph AI sidebar assistant picker: list Dorph built-ins **and** AI tools assistants; selecting a BK assistant uses Dorph chat tools **plus** that assistant’s system prompt and RAG (option C).

## Design

### AI tools cards

- Responsive CSS grid of cards (name, description, provider/model chips, RAG chips, Run/Edit/Delete).
- Preserve create/edit/run dialogs; dark sci-fi theme aligned with BK.

### Dorph picker + chat

- `GET /api/ai-agents` returns Dorph Badger agents + BK `GET /assistants` (via `EXTERNAL_API_BASE`).
- BK entries use ids `bk:<assistant_id>`, `source: "bk"`.
- On chat with `bk:…`: fetch profile, query each RAG collection (`POST /rag/collections/{name}/query`), inject system prompt + truncated RAG into Dorph tool-loop instructions.
- Dorph-native agents (e.g. `image-generator`) unchanged.
- If BK is down, Dorph built-ins still list/work; BK section omitted.

## Out of scope

- Dorph UI for creating/editing BK assistants.
- Proxying full BK `/assistants/{id}/run` (would drop Dorph tools).
