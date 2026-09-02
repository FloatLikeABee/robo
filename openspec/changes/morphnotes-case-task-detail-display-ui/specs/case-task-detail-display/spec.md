## Purpose

Lets MorphNotes operators read Case/task JSON detail as a structured document in Markdown and HTML instead of a raw JSON dump.

## ADDED Requirements

### Requirement: Markdown view shows structured detail, not a JSON fence
When Case/task Markdown is derived from a draft whose `detail` is a JSON object with at least one key, the Markdown document MUST include a Detail section that presents those fields as readable labels and values. Keys MUST be shown as human-readable labels (snake_case and camelCase converted to Title Case). Nested objects MUST appear under nested headings. Arrays MUST appear as bullet lists (or numbered items). The Detail section MUST NOT wrap the payload in a fenced `json` code block solely to dump the object.

#### Scenario: Object fields appear as labeled lines
- **WHEN** the draft detail is `{"priority":"high"}` and the operator opens the Markdown tab
- **THEN** the Markdown includes a Detail section with a readable Priority label and the value high
- **AND** the Markdown does not contain a ` ```json ` fence whose body is that object

#### Scenario: Nested object and list are readable
- **WHEN** the draft detail is `{"case_summary":{"priority":"high"},"tags":["fleet"]}`
- **THEN** the Markdown includes a nested heading for Case summary with a Priority label, and a Tags list that includes fleet
- **AND** it does not present that payload as a single fenced JSON block

### Requirement: HTML view shows a document UI for detail, not a JSON dump
When Case/task HTML is derived from the same draft, the HTML preview MUST render `detail` as a visual document section (labeled fields, nested groups, lists or cards) on the existing dark Case/task page. It MUST NOT show the parsed object as a monospace JSON dump (`<pre>` / `<code>` of pretty-printed JSON) as the primary presentation. The preview MUST stay dark (not a light-theme document).

#### Scenario: HTML shows labeled detail, not pretty JSON
- **WHEN** the draft detail is `{"priority":"high"}` and the operator opens the HTML tab
- **THEN** they see a rendered Detail section that includes the Priority label and high as visible text
- **AND** they do not see a code block whose text is the pretty-printed JSON object

#### Scenario: HTML stays dark
- **WHEN** the operator opens the HTML tab for a case with non-empty detail
- **THEN** the preview page remains a dark document suitable for MorphNotes

### Requirement: Empty and invalid detail do not break the views
Empty detail (`{}`, empty string, or missing) MUST omit the Detail section rather than showing an empty JSON object or a `{}` code fence. If `detail` is not valid JSON, Markdown and HTML MUST still render the rest of the case document and MUST show the raw detail text so the operator can see what failed to parse.

#### Scenario: Empty object omits Detail
- **WHEN** the draft detail is `{}` or blank
- **THEN** Markdown and HTML do not include a Detail heading whose body is `{}` or a fenced empty object

#### Scenario: Invalid JSON still shows the raw text
- **WHEN** the draft detail is not valid JSON
- **THEN** Markdown and HTML still show title and description
- **AND** the views include the raw detail text (not a crash or blank preview)

### Requirement: Views stay derived from the current draft
Markdown and HTML MUST continue to rebuild from the current Case/task draft (including unsaved Detail-tab edits). The system MUST NOT add persisted markdown/html columns on CaseTask for this change. The Details tab remains the editor for JSON `detail`.

#### Scenario: Unsaved detail edit appears in views
- **WHEN** the operator changes a detail field on Details without saving, then opens Markdown and HTML
- **THEN** both views reflect the updated detail as structured display, not a stale JSON dump
