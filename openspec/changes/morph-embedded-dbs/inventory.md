# Morph MySQL / Mongo inventory

Source: `morph/handlers/*.go`, `morph/db/*.go`, `morph/cmd/*`, `morph/importcol/insert.go`, `morph/models/tran.go`, migrations `032`–`047` (plus schema context from earlier migrations still reflected in live column names).

Scope: Morph Go service Tran MySQL tables **currently used** (handlers / `db/` / `cmd/`). Column lists below are **handler/CLI SELECT·INSERT names**, not ancient migration aliases (`Student`/`School`/`Trip`/`Vehicle`/`Note`/etc.).

---

## 1. FK-safe migrate order (parents → children)

Copy in this order (or equivalent topological sort). Soft logical refs (no MySQL FK) are still ordered parents-first.

1. `District`
2. `User`
3. `plat_users`
4. `contact`
5. `facility` *(soft `district_id` → District.ID)*
6. `member`
7. `employee` *(FK `facility_id` → facility.id)*
8. `Asset`
9. `Activity`
10. `ActivityEmployee` *(FK → Activity, employee)*
11. `ActivityParticipant` *(FK → Activity, member)*
12. `ActivityAsset` *(FK → Activity, Asset)*
13. `CaseTask`
14. `CaseTaskAssignee` *(FK → CaseTask)*
15. `StoryPost` *(logical `author_user_id` → User)*
16. `comment` *(self `ParentID`; logical entity/record)*
17. `generic_data`
18. `EntityAttachment`
19. `AdminGridSavedFilter`
20. `AdminGridColorConfig`
21. `PlatformUiConfig`
22. `user_note_todo` *(logical UserID → User)*
23. `record_contact` *(logical ContactID → contact)*
24. `tool_note` *(logical AuthorUserID / TargetUserID → User)*
25. `tool_broadcast` *(schema present; AttachmentsJSON via 043 — no active handler SQL found)*
26. `big_note` *(logical user_id → User)*
27. `big_note_response` *(→ big_note)*
28. `morph_knowledge_files`
29. `morph_knowledge_chunks` *(FK file_id → morph_knowledge_files)*
30. `graph_sync_outbox`
31. `StaffType` *(seed/CLI only)*
32. `StaffStaffType` *(seed/CLI; StaffID → employee.id)*
33. `email_agent` *(migration only; no Go handler SQL)*
34. `email_agent_log`
35. `user_message_auto_reply_log` *(migration only; User also has MessageAiAutoReply* cols unused by handlers)*

**Dropped / do not migrate:** `CustomAttribute*`, `CaseTaskAttachment` (data folded into `EntityAttachment` in 039), legacy tables from 030 (`StudentSchedule`, `SchoolGrade`, `Grade`, merge/form/code tables, mailing tables).

**Mongo (after or parallel to relational parents):** collection `entity_details` → Badger keys (see §3).

---

## 2. Core table columns (handler current usage)

### `member` (handlers: `tran_students.go`; import/seed/cmd)

| Role | Columns |
|------|---------|
| SELECT | `id`, `last_name`, `first_name`, `middle_name`, `dob`, `entry_date`, `facility`, `gender`, `email`, `participant_type`, `description` |
| INSERT (typical) | same write set; minimal create uses `last_name` only |

### `employee` (handlers: `tran_staff.go`)

| Role | Columns |
|------|---------|
| SELECT (list join) | `e.id`, `e.last_name`, `e.first_name`, `e.middle_name`, `e.email`, `e.phone_number`, `e.active_flag`, `e.employ_type`, `e.description`, `e.facility_id` (+ facility display) |
| WRITE allowed | `last_name`, `first_name`, `middle_name`, `staff_guid`, `active_flag`, `inactive_date`, `contractor_id`, `email`, `phone_number`, `date_of_birth`, `gender`, `user_id`, `employ_type`, `description`, `facility_id` |

### `facility` (handlers: `tran_schools.go`)

| Role | Columns |
|------|---------|
| SELECT | `id`, `facility_code`, `name`, `district_id`, `facility_type`, `description`, `location` |
| INSERT/UPDATE | `facility_code`, `name`, `district_id`, `facility_type`, `description`, `location` |

*Note: seed still writes `capacity`; not in handler SELECT.*

### `Asset` (handlers: `tran_vehicles.go`)

| Role | Columns |
|------|---------|
| SELECT list | `ID`, `asset_tag`, `description`, `AssetID`, `AssetType` |
| SELECT full | `ID`, `ContractorID`, `asset_tag`, `description`, `AssetID`, `AssetType` |
| WRITE | `asset_tag`, `description`, `AssetID`, `AssetType`, `ContractorID` |

### `Activity` (handlers: `tran_trips.go`)

| Role | Columns |
|------|---------|
| SELECT list | `ID`, `Name`, `start_date`, `end_date`, `location`, `ActivityType`, `description` |
| SELECT full | + `GUID` |
| WRITE | `Name`, `ActivityType`, `start_date`, `end_date`, `location`, `GUID`, `description` |

### `ActivityEmployee` / `ActivityParticipant` / `ActivityAsset`

Used by `cmd/seed_tran` (and prune/demo_limits deletes). Handlers do **not** query them today.

| Table | Columns (schema / seed) |
|-------|-------------------------|
| `ActivityEmployee` | `id`, `activity_id`, `employee_id`, `created_on` |
| `ActivityParticipant` | `id`, `activity_id`, `member_id`, `created_on` |
| `ActivityAsset` | `id`, `activity_id`, `asset_id`, `created_on` |

### `CaseTask` (handlers: `tran_case_tasks.go`)

| Role | Columns |
|------|---------|
| SELECT / list | `ID`, `title`, `description`, `start_at`, `end_at`, `location`, `assignee_type`, `assignee_id`, `created_on`, `last_updated` |
| INSERT | dynamic from those write fields (title required) |

### `CaseTaskAssignee`

| Role | Columns |
|------|---------|
| SELECT | `case_task_id`, `assignee_kind`, `assignee_id` (+ schema `id`, `created_on`) |
| INSERT | `case_task_id`, `assignee_kind`, `assignee_id` |

### `contact` (handlers: `tran_contacts.go`)

| Role | Columns |
|------|---------|
| SELECT | `ID`, `LastName`, `FirstName`, `Email`, `Phone`, `Mobile`, `description` |
| INSERT | same |

### `User` (handlers: `tran_users.go`, story author lookup)

| Role | Columns |
|------|---------|
| SELECT | `UserID`, `LoginID`, `FirstName`, `LastName`, `Email`, `Phone`, `Title`, `Administrator`, `Deactivated` |
| INSERT | `Administrator`, `LoginID`, `FirstName`, `LastName`, `Email`, [`Phone`, `Title`], `Deactivated` |
| UPDATE extras | `DeactivatedDate` on soft-offboard |

*Schema also has many unused legacy cols + `MessageAiAutoReplyEnabled` / `MessageAiAutoReplyPrompt` (038) — not in handler SELECT.*

### `plat_users` (`db/plat_users.go`)

| Role | Columns |
|------|---------|
| DDL / INSERT | `id`, `email`, `username`, `password_hash`, `google_id`, `is_verified`, `roles`, `permissions`, `default_channel_id`, `verification_token`, `verification_expires_at`, `reset_token`, `reset_expires_at`, `created_at`, `updated_at` |
| SELECT (runtime) | `id`, `email`, `username`, `password_hash`, `is_verified`, `roles`, `permissions`, `default_channel_id`, `created_at`, `updated_at` |

### `StoryPost` (handlers: `tran_story_posts.go`)

| Role | Columns |
|------|---------|
| SELECT | `ID`, `title`, `content`, `author_user_id`, `created_on`, `last_updated` |
| INSERT | `title`, `content`, `author_user_id` |

### `comment` (handlers: `tran_comments.go`)

| Role | Columns |
|------|---------|
| SELECT | `ID`, `EntityType`, `RecordID`, `ParentID`, `AuthorUserID`, `Body`, `CreatedOn`, `LastUpdated` |
| INSERT | `EntityType`, `RecordID`, `ParentID`, `AuthorUserID`, `Body` |

### `generic_data` (handlers: `tran_generic_data.go`)

| Role | Columns |
|------|---------|
| SELECT | `id`, `title`, `source_type`, `source_filename`, `record_count`, `description`, `ai_analysis`, `created_on`, `last_updated` |
| INSERT | `title`, `source_type`, `source_filename`, `record_count`, `description` (+ later `ai_analysis` UPDATE) |

### `EntityAttachment` (handlers: `tran_entity_attachments.go`)

| Role | Columns |
|------|---------|
| SELECT list | `id`, `entity_type`, `record_id`, `original_name`, `mime_type`, `size_bytes`, `created_on` |
| SELECT download | + `file_path`, `stored_name` |
| INSERT | `entity_type`, `record_id`, `original_name`, `stored_name`, `file_path`, `mime_type`, `size_bytes` |

### `AdminGridSavedFilter`

SELECT: `ID`, `GridKey`, `Name`, `FilterJSON`, `CreatedOn` · INSERT: `GridKey`, `Name`, `FilterJSON`

### `AdminGridColorConfig`

SELECT: `ConfigJSON` · INSERT/UPSERT: `GridKey`, `ConfigJSON` *(schema also `UpdatedOn`)*

### `PlatformUiConfig`

SELECT/UPSERT: `ID`, `ConfigJSON` *(schema also `UpdatedOn`)*

### `user_note_todo` (handlers: `tran_notes_todos.go`)

| Role | Columns |
|------|---------|
| SELECT | `ID`, `UserID`, `ItemType`, `Title`, `Body`, `Completed`, `DeadlineAt`, `CreatedOn`, `LastUpdated` |
| INSERT | `UserID`, `ItemType`, `Title`, `Body`, `Completed`, `DeadlineAt` |

### `record_contact` (handlers: `tran_record_contact.go`)

| Role | Columns |
|------|---------|
| SELECT | `ID`, `EntityType`, `RecordID`, `ContactID`, `Relationship`, `IsPrimary` |
| INSERT | `DBID`, `EntityType`, `RecordID`, `ContactID`, `Relationship`, `IsPrimary` |

*(schema also `CreatedOn`)*

### `District` (handlers: `tran_districts.go`)

| Role | Columns |
|------|---------|
| SELECT | `ID`, `DistrictID`, `District`, `Name`, `description` |
| INSERT | `DBID`, `DistrictID`, `District`, `Name`, `description` |

### `big_note` (handlers: `tran_big_notes.go`)

| Role | Columns |
|------|---------|
| SELECT | `id`, `user_id`, `owner_key`, `title`, `idea`, `note_kind`, `markdown_content`, `html_content`, `questions_json`, `theme`, `published_slug`, `published_path`, `created_on`, `last_updated` |
| INSERT | `user_id`, `owner_key`, `title`, `idea`, `note_kind`, `markdown_content`, `html_content`, `questions_json`, `theme` |

### `big_note_response`

| Role | Columns |
|------|---------|
| SELECT | `id`, `big_note_id`, `answers_json`, `analysis_markdown`, `created_on`, `last_updated` |
| INSERT | `big_note_id`, `answers_json` |

### `graph_sync_outbox` (`db/knowledge.go`)

INSERT: `source`, `entity_type`, `entity_id`, `op`, `payload_json`  
Full schema: + `id`, `created_at`, `available_at`, `attempts`, `locked_by`, `locked_at`, `processed_at`, `last_error`

### `morph_knowledge_files`

SELECT/INSERT: `id`, `title`, `filename`, `content_type`, `kind`, `storage_path`, `byte_size`, `text_excerpt`, `created_by`, `created_at`, `updated_at`

### `morph_knowledge_chunks`

SELECT/INSERT: `id`, `file_id`, `chunk_index`, `text_content`, `embedding_json`, `created_at`

### Other heavily used / related

| Table | Handler columns |
|-------|-----------------|
| `tool_note` | SELECT/INSERT: `ID`, `AuthorUserID`, `TargetType`, `TargetUserID`, `IsPrivate`, `Title`, `Body`, `ReadAt`, `CreatedOn`, `LastUpdated` |
| `StaffType` / `StaffStaffType` | seed only (`StaffTypeID`, `StaffTypeName`, … / `StaffID`, `StaffTypeID`, `PrimaryFlag`) |

---

## 3. Mongo collection: `entity_details`

**Canonical write fields** (`db/entity_details_mongo.go` `SetEntityDetailJSON`):

| Field | Role |
|-------|------|
| `entity` | string key |
| `record_id` | int (MySQL PK) |
| `body` | JSON text (or BSON object tolerated on read) |
| `updated_at` | set on upsert |

**Read aliases tolerated:** `Body`/`detail`/`Detail`/`payload`/`Payload`; `recordId`; staff↔employee entity variants.

**Entity keys used by Morph** (`handlers/entity_detail.go`):

`student`, `staff` (+ mirror `employee`), `school`, `vehicle`, `trip`, `contact`, `district`, `case_task`, `generic_data`

---

## 4. Usage map (quick)

| Area | Tables |
|------|--------|
| Handlers (hot path) | member, employee, facility, Asset, Activity, CaseTask, CaseTaskAssignee, contact, User, StoryPost, comment, generic_data, EntityAttachment, AdminGrid*, PlatformUiConfig, user_note_todo, record_contact, District, big_note*, plat_users (via auth/db), tool_note, knowledge/outbox |
| `db/` ensure DDL | StoryPost, generic_data, big_note*, facility.location, graph/knowledge, plat_users |
| `cmd/` seed/backfill/prune | same core entities + Activity* junction + StaffType/StaffStaffType; Mongo entity_details |
| Migrations 032–047 only / unused in handlers | email_agent*, user_message_auto_reply_log, tool_broadcast.AttachmentsJSON, User.MessageAiAutoReply* |

---

## 5. Notes for migrate tooling

- Prefer **handler column sets** when projecting into SQLite; leftover legacy columns on live MySQL can be ignored or copied as-is if present (`capacity`, unused `User` cols, etc.).
- Enforce FK order for: `employee.facility_id`, Activity* junctions, `CaseTaskAssignee`, `morph_knowledge_chunks`, `big_note_response`.
- Soft-delete legacy empty `StudentSchedule` checks in prune are defensive only — table dropped in 030.
- Dual-write pattern: MySQL row + Mongo `entity_details` for Morph Data entities listed in §3.
