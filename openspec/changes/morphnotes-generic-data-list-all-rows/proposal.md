## Why

After saving a MorphNotes Generic data import, the grid often still shows an older row instead of the new one, and the table looks incomplete. The list API skips every other database row, so roughly half of the records never appear.

## What Changes

- `GET /api/tran/generic-data` returns every stored Generic data row (same order as today: newest `last_updated` first), not every other row.
- After Save on import, the grid includes the saved record (typically at the top). The detail drawer and the grid row match.
- Tests lock in an even-count insert (e.g. 4 rows → 4 in the list, including the newest id).

## Capabilities

### New Capabilities

- `generic-data-list-complete`: Generic data list/grid shows all saved records; a newly saved import appears in the table.

### Modified Capabilities

- (none — `openspec/specs/` has no archived `generic-data` baseline)

## Impact

- **MorphNotes backend**: `ListGenericData` row scan (`querySingleRowMapFromRows` currently calls `Next` while the list loop already did).
- **MorphNotes frontend**: Generic data DataGrid (reload after Save already exists; it will show the new row once the list is complete).
- **Out of scope**: Raising the existing 500-row list cap; paste-vs-file extract UI.
