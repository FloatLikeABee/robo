# Spec Delta

## MODIFIED Requirements

### Requirement: Tools and the notes app resource
The server MUST advertise the tools capability and the resources capability. `tools/list` MUST include `whoami`, `list_my_tasks`, `get_task`, `create_note`, `note_grill`, `suggest_note_kind`, and `dorph_guide`. `whoami`, `list_my_tasks`, `get_task`, `suggest_note_kind`, and `dorph_guide` MUST be marked read-only. `create_note` and `note_grill` MUST NOT be marked read-only. `resources/list` MUST succeed and MUST include the Dorph notes app at `ui://dorph/notes` and the Dorph grill app at `ui://dorph/grill`, each with MIME type `text/html;profile=mcp-app`, and the agent playbook at `dorph://guide`. The resources capability MUST NOT advertise `listChanged`. The server MUST NOT expose prompts, Engi project tools, or BK TCP tools.

#### Scenario: Three read-only tools
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the list includes `whoami`, `list_my_tasks`, `get_task`, `create_note`, `note_grill`, `suggest_note_kind`, and `dorph_guide`
- **AND** `whoami`, `list_my_tasks`, `get_task`, `suggest_note_kind`, and `dorph_guide` are marked read-only
- **AND** `create_note` and `note_grill` are not marked read-only

#### Scenario: Resources list is the notes app
- **WHEN** a client calls `resources/list` after initialize
- **THEN** the call succeeds
- **AND** the list includes `ui://dorph/notes`, `ui://dorph/grill`, and `dorph://guide`
- **AND** each of the UI resources has MIME type `text/html;profile=mcp-app`
- **AND** initialize capabilities do not set resources `listChanged` to true
