## Context

See proposal.md for motivation.

Create case/task (`CaseTasks.js` drawer + `POST /api/tran/case-tasks`) is a blank form: title, description, start/end, map `location` JSON `{label, area:[[lat,lng],...]}`, and entity `detail` JSON. `POST /api/tran/extract-json` already turns pasted text into a JSON object for Generic Data / Assets, but it is text-only and does not map onto case/task columns or location. Timeline create already combines prompt + file + URL, extracts PDF/txt/md, and **inserts** a row — the wrong persistence model here. PDF text extraction (`pkg/docextract` / `morphgraph.ExtractPDFBytes`) and `importcol.ParseUpload` for CSV/xlsx already exist. Detail JSON is capped at 5 nesting levels in the UI validator.

## Goals / Non-Goals

**Goals:**

- One authenticated draft endpoint that accepts prompt, file, or both and returns a structured case/task draft.
- Apply that draft into the existing create/edit drawer fields, including JsonDetailEditor and map area.
- Reuse existing extractors; no new tables.

**Non-Goals:**

- Auto-inserting a CaseTask on generate (unlike Timeline create).
- OCR / vision for image-only PDFs or photos.
- Auto-assigning members, employees, or contacts.
- Calling a geocoder to turn a place name into a polygon.
- Changing save, email, attachments, or list layout.

## Decisions

### 1. Dedicated draft endpoint, not `extract-json` purpose
- **Choice**: `POST /api/tran/case-tasks/ai-draft` (auth same as other `/api/tran/case-tasks` routes). Accept `multipart/form-data` (`prompt`, `file`) and `application/json` (`{ "prompt": "..." }`). Response:

```
{
  "title": "string",
  "description": "string",
  "start_at": "ISO-8601 or empty",
  "end_at": "ISO-8601 or empty",
  "location": { "label": "string", "area": [[lat, lng], ...] } | null,
  "detail": { ...object... }
}
```

- **Rationale**: File upload needs multipart; the contract is a task draft, not a free-form JSON blob. Extending `extract-json` would overload purpose strings and still need a second file path.
- **Alternatives**: New `purpose: case_task` on `extract-json` plus a separate upload helper — two round trips and a weaker location/date contract. Client-side extract — still needs the model and file extraction belongs on the server.

### 2. Fill the form; persist only on existing Save
- **Choice**: Generate applies fields to `draft` in the drawer. Create/update still goes through `CreateCaseTask` / `UpdateCaseTask`. If title, description, or detail is already non-empty, confirm before replace (same idea as Assets extract).
- **Rationale**: Matches Generic Data / Assets review-before-save. Operators can fix bad dates or JSON. Closing the drawer without Save is a no-op.
- **Alternatives**: Auto-create then open edit — harder to undo and contradicts “most importantly the detail json” as something to inspect.

### 3. Combine prompt and file like Timeline sources
- **Choice**: Extract file text (PDF via `docextract.ExtractPDFBytes`; txt/md as plain text; csv/xlsx via `importcol.ParseUpload` serialized to text). Concatenate labeled sections: prompt, then file. Cap combined runes (same order as timeline / extract-json, ~16k–48k). Truncate with a clear error or silent trim — prefer **400 if empty after extract**, truncate with a note in the prompt if over cap.
- **Rationale**: User asked for prompt or file or both. Timeline already concatenates sources; reuse that shape, not its INSERT.
- **Alternatives**: Two sequential model calls — slower, no benefit. File as attachment-only after save — misses “create from material”.

### 4. File types: txt, md, pdf, csv, xlsx
- **Choice**: Allow `.txt`, `.md`, `.markdown`, `.pdf`, `.csv`, `.xlsx`. Reject `.xls` unless parse already works; reject images. Map `ErrNoText` to 400. Size cap aligned with timeline (~5 MiB).
- **Rationale**: Briefs are often PDF or a pasted prompt; spreadsheets show up as work lists. Vision/OCR is out of scope.
- **Alternatives**: PDF/txt only — too tight. Any file stored as an attachment first — does not fill the form.

### 5. Location: label always when named; `area` only from source coordinates
- **Choice**: Prompt the model: if a place name is present, set `location.label`; set `location.area` only when the source contains explicit lat/lng (or a coordinate list). Never invent a polygon. Frontend applies via existing `parseLocation` / map preview; user can still open **Select map area**.
- **Rationale**: Without a geocoder, guessed polygons are wrong. A label still shows on the chip; drawing remains manual.
- **Alternatives**: Nominatim geocode on the server — extra dependency, rate limits, not requested. Skip location entirely — weaker than “select map area if any”.

### 6. Detail JSON is the required structured payload
- **Choice**: Model MUST return a JSON **object** for `detail` (snake_case keys, nested objects/arrays allowed, stay within the existing 5-level UI cap). Prefer operational structure similar to seeded case tasks (`case_summary`, `workflow`, `tags`, plus fields implied by the source) but **do not** require the demo seed keys. Server validates `detail` is a parseable object; frontend runs `validateJsonDetailStructure` before apply. Title is required; description may be short if the JSON carries the body.
- **Rationale**: User called detail JSON the most important field. Seed shape is a hint, not a schema lock-in, so arbitrary briefs still extract.
- **Alternatives**: Stuff everything into description — loses the JSON editor. Strict JSON Schema — too rigid for mixed source docs.

### 7. UI placement
- **Choice**: At the top of the Create/Edit drawer, a compact “Generate from material” block: prompt textarea, file name + picker, Generate button, progress/error. Keep the rest of the form as it is so Generate is optional.
- **Rationale**: Same drawer, no extra wizard. Edit can re-fill an existing task from new material.
- **Alternatives**: Separate “AI create” dialog that then opens the drawer — extra click. Timeline-style full-page create — inconsistent with Tasks cards.

## Risks / Trade-offs

- [Model invents dates or coordinates] → Prompt: omit if not in source; server drops `area` points that are not finite numbers; user reviews before Save.
- [Detail JSON nesting > 5] → Frontend validation blocks apply; user can trim in the editor.
- [Large PDFs / spreadsheets] → Size + rune caps; 400 on empty extract.
- [AI key missing] → 503 with the same “AI service not configured” message used elsewhere.
- [Replacing a filled edit form] → Confirm before apply.

## Migration Plan

- Deploy Morph API + Morph Data frontend together (new route + UI). No schema migration.
- Rollback: hide Generate / 404 the new route; manual create unchanged.

## Open Questions

None that change specs or the task breakdown. Geocoding a label into a polygon can be a later change without altering generate-before-save.
