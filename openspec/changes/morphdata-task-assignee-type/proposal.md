## Why

Saving a new Morph Data **Tasks** row fails with `NOT NULL constraint failed: CaseTask.assignee_type (1299)` when the form has no assignee. Assignees are already soft-deprecated on write; operators must still be able to create a title-only (or otherwise unassigned) task.

## What Changes

- Creating a case/task **without** assignees succeeds on embedded SQLite (including databases whose `CaseTask` table was created with `assignee_type` NOT NULL and no usable DEFAULT).
- The create path always writes a safe legacy `assignee_type` / `assignee_id` so SQLite does not insert NULL.
- The UI still omits assignees on save; no new assignee picker is required.
- Failures surface a clear error only for real validation problems (empty title, bad JSON), not a raw SQLite constraint code.

## Capabilities

### New Capabilities

- `case-task-unassigned-create`: Morph Data Tasks can save a new case/task that has no assignees without hitting a NOT NULL constraint on `CaseTask.assignee_type`.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- **Morph Data backend**: `CreateCaseTask` insert in `morph/handlers/tran_case_tasks.go`; possibly SQLite schema ensure in `morph/db/sqlite_schema.go` / `tran_sql.go` for older `CaseTask` files.
- **Morph Data frontend**: `CaseTasks.js` save payload stays assignee-free; no product change unless the error toast needs a friendlier message (backend should stop returning the constraint error).
- **Tests**: Handler or SQLite insert test covering create with title only against NOT NULL `assignee_type`.
- **Out of scope**: Restoring assignee editing; MySQL-only deployments; changing the Tasks form fields besides making save work.
