## Why

Event Logs still carries an Info Sheets builder nobody needs. DorphUtils assistants (and DorphNotes AI) still answer in prose even when a chart would do. Creating a DorphNotes task dumps extra fields; operators only want generate-from-prompt or upload, a title, start/end, and a markdown/HTML document.

## What Changes

- **BREAKING (Event Logs):** Remove the Info Sheets operator module (header tab, Survey Bot page, DorphUtils copy). Events & Info stays. `/survey-bot` redirects to `/events-info`.
- Add a shared **graphs** skill for in-stack agents: when a reply can show structure or quantities, include a mermaid diagram/chart. Apply it in Dorph AI and all four DorphUtils modules (Event Logs, Content Maker, Data Access, Project), and in DorphNotes AI (notes/todos assist, task draft, research).
- **DorphNotes Tasks create:** Keep only generate-from-prompt, upload-generate, title, and start/end. Drop description, map, JSON detail, and attachments from the create form. Generated outcome is markdown and HTML only (mermaid allowed). Existing task edit may keep fields already stored.

## Capabilities

### New Capabilities

- `event-logs-without-info-sheets`: Event Logs is Events & Info only; Info Sheets is gone from operator UI.
- `agent-graphs-skill`: Shared graph/chart reply skill for Dorph AI, DorphUtils module assistants, and DorphNotes AI.
- `case-task-create-md-html`: Creating a DorphNotes task is prompt/upload generate plus title and times; the document is markdown and HTML.

### Modified Capabilities

- (none — no archived baselines under `openspec/specs/`)

## Impact

- Event Logs: `formx/frontend` Layout, App routes, SurveyBot page unused; DorphUtils `config.ts` copy
- Graphs: `pkg/dorphai` VisualFirst (reuse), Dorph AI builtin skill, Event Logs / Content Maker / Data Access / Project assistant prompts, DorphNotes text-assist + case-task AI draft + research compose
- Tasks: `dorph/frontend/src/pages/admin/CaseTasks.js`, `tran_case_tasks_ai.go` draft shape (markdown/html instead of location/detail JSON)
- Dark-only. No new product names.
