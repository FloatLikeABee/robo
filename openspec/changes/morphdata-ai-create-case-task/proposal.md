## Why

Creating a Morph Data case/task is still a blank form: title, description, start/end, map area, and especially the detail JSON must be typed by hand even when the work already exists as a briefing note, PDF, or pasted prompt. Operators need to turn that material into a filled task without re-keying structured fields.

## What Changes

- On **Create case/task** (and the same drawer when editing), the user can supply a **prompt**, an **uploaded file**, or **both**.
- AI reads that material and fills the draft: **title**, **description**, **start/end** when the source implies dates, **map area / location** when the source implies a place, and **detail JSON** as the primary structured payload.
- The user **reviews and edits** the filled fields, then saves through the existing create/update path. AI does **not** auto-save a task.
- Manual create (empty form, type everything) stays available.

## Capabilities

### New Capabilities

- `case-task-ai-draft`: From a prompt and/or uploaded file, Morph Data AI proposes a case/task draft (title, description, optional dates, optional map location, required detail JSON) for review before save.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- **Morph Data backend**: New authenticated draft endpoint (multipart + JSON) on `/api/tran/case-tasks`; PDF/text extraction reuse; Morph AI prompt that returns a structured task object. No new database tables.
- **Morph Data frontend**: `CaseTasks.js` create/edit drawer gains prompt + file + “Generate” that applies the draft into existing fields (including JsonDetailEditor and map area).
- **API catalog**: `management_chat.go` / `tranClient.js` document the new route.
- **Config**: Existing Morph AI key; clear error when AI is not configured. No new env vars.
- **Out of scope**: Auto-creating records; OCR for scanned image-only PDFs; auto-assigning members/employees; geocoding a named place into a polygon when the source has no coordinates.
