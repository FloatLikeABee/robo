---
name: dorphai-fast-ops
description: Improves DorphAI, DorphData, SheetX (FormsX), ComposerX, and DataX work efficiency by using MCP-style APIs, live product endpoints, and direct read-only database/repository lookups first. Use when working on DorphAI assistants, tool loops, MCP APIs, database search, data operations, SheetX, FormsX, ComposerX, DorphData, or DataX.
---

# DorphAI Fast Ops

## Goal

React fast by grounding answers and operations in the smallest live data source available: MCP-style tool catalogs, product APIs, repositories, or read-only database queries.

## Fast Workflow

1. Identify the product context from the request: DorphData, SheetX (FormsX), ComposerX, DataX, or shared DorphAI.
2. Prefer a tool catalog or schema endpoint before broad code search.
3. Run the smallest live lookup that can answer the request.
4. If more data is needed, list first, then fetch detail by id.
5. Summarize compactly in user language. Do not paste raw JSON unless requested.

## Product Map

- DorphData: `/api/chat`, `/api/tran/*`, `/api/forms/*`, `/api/sheetx/*` (alias `/api/formsx/*`), `/api/composerx/*`, `/api/knowledge/*`, `/api/graph/search`; internal execution uses `dorph/handlers/internal_api.go`.
- SheetX (repo `formx/`): `/api/v1/ai/mongodb-mcp`, `/api/v1/ai/app-abilities`, `/api/v1/events-info`, `/api/v1/survey-bot/*` (AI Sheets), and assistant tools in `formx/backend/internal/handler/assistant_llm.go`. Say **survey bot** or use **AI Sheets** to enter guided Survey Bot mode.
- ComposerX: `/ai/mcp-tools`, `/ai/assistant/chat`, templates, saved emails, reference docs, and UsersPanel inbox tools.
- DataX: `/api/v1/assistant/chat`, `/api/v1/data-ai/chat`, `/api/v1/databases`, `/api/v1/databases/:id/schema`, `/api/v1/data-tables`, `/api/v1/data-tables/:id/query`.
- Shared DorphAI: Go helpers in `pkg/dorphai`; Rust helpers in `pkg/dorphai-rs`; GraphRAG helpers in `pkg/dorphgraph`.

## Tool Use Rules

- Use read-only SQL unless the user explicitly asks to create, update, or delete.
- For destructive operations, fetch or list the target first and confirm the exact id/path unless the user already gave a precise target.
- Preserve auth forwarding when proxying between products.
- Prefer direct repository calls inside a product service; use HTTP when crossing product boundaries.
- For DorphData single records, prefer `/full` routes when available so extended `detail` fields are included.
- For Dorph knowledge / uploaded docs, prefer `POST /api/graph/search` before broad listing.
- For SheetX system knowledge, use Mongo-backed system documents before public web search.
- For SheetX AI Sheets (Survey Bot) results/templates, use `/api/v1/survey-bot/*` or assistant tools `list_survey_results` / `list_survey_templates`.
- For ComposerX reference work, list/search reference docs before using generic generation.
- For DataX data questions, use schema/table/query tools and keep answers within DataX data scope.

## Response Pattern

When an assistant tool loop is needed, the model should produce only one JSON object for the tool call. After `TOOL_RESULT`, either issue another JSON tool call or answer in markdown.

Use concise result summaries:

```text
TOOL_RESULT
[compact JSON or text result]

Summarize for the user in markdown. If another tool is needed, reply with only one JSON object.
```

## Safety

- Do not invent ids, table names, columns, or endpoint capabilities.
- Do not bypass allowlists for internal DorphData tool calls.
- Do not expose secrets, connection strings, or bearer tokens in user-facing answers.
- Keep large tool results truncated and fetch detail only for selected records.
