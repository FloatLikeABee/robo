## Context

See proposal.md for why. Today dorph-mcp verifies one session JWT at startup and on each call, opens `TRAN_SQLITE_PATH` read-only, and writes through a second `mode=rw` connection. It does not run migrations. Notes & TODOs, timelines, stories, and research are already owner-scoped. Stick notes write `CaseTask` with no owner. That table is also the DorphNotes board: one table, two kinds (`stick`, `task`). #134: JWT_SECRET in client configs, no revoke list, deleted users keep working until the process exits.

## Goals / Non-Goals

**Goals:**

- PATs hashed in SQLite; dorph-mcp verifies them without a client JWT secret.
- One kind-dispatch for create/get/list/update/append/archive plus search, with old names as aliases.
- Owner on `CaseTask`, backfill, MCP scoped, admin HTTP list unchanged.
- Grill start accepts every kind.

**Non-Goals:**

- Pool, share, OAuth, Streamable HTTP, Settings UI, `dorph_guide`, `suggest_note_kind`.
- Changing `tranUserIDFromContext` or the SPA Notes & TODOs editor.
- Echo's `16-dorph-for-agents.md` and the skills pack.
- Argon2, cursor pagination, a separate `set_todo_status` tool.

## Decisions

### 1. SHA-256 of the full token, constant-time compare

The secret is 32 random bytes. A fast hash is enough. `golang.org/x/crypto` is already in the module for bcrypt; we still do not add argon2. Store `hex(sha256(token))`. Never log the token. Format `dorph_pat_<id>_<secret>` where `id` is 16 hex chars.

**Rejected:** argon2id (extra API, no gain on high-entropy secrets). Hashing only the secret suffix (a leaked id would still need the secret; hashing the whole string is one compare).

### 2. PAT path skips JWT_SECRET; JWT path unchanged

`CheckStartup` / `IdentityFromEnv` / `RecheckFromEnv` branch on the `dorph_pat_` prefix. PAT verify opens the same SQLite file and reads `agent_token`. JWT still refuses an empty or default secret. A session JWT is treated as `notes:read` and `notes:write`. Tests that build `Identity` with empty scopes keep that implicit notes access so existing `NewServer` tests do not all grow scope lists.

**Rejected:** dropping JWT support this phase. The plan keeps it behind a flag; keeping it working is the flag.

### 3. Recheck hits plat_users and the token row every call

`RecheckFromEnv` for a PAT re-reads the row (revoked, expired) and `ConfirmPlatUser`. For a JWT it still decodes, then `ConfirmPlatUser` on the already-open read connection. That closes the #134 deleted-user gap. `last_used_at` is a best-effort write on the rw path; a read-only recheck that cannot write is still a valid deny/allow.

### 4. One CaseTask table, `board_kind`, `owner_user_id`

`owner_user_id` is `plat_users.id` (TEXT). `board_kind` is `stick` or `task`, default `stick`. Existing rows backfill owner to the oldest Admin plat user, else the oldest plat user, recorded once like the agent-lesson claim. MCP filters by owner and kind. `GET /api/tran/case-tasks` keeps its column list and has no owner filter, so the admin board still shows every sticker. HTTP create sets `owner_user_id` from the session when present.

**Rejected:** two tables (they are already one). Filtering the admin list by owner (would hide backfilled-to-admin rows from a non-admin viewer and change the UI).

### 5. Unified record, integer `id` on `create_note` when kind is note

New tools return `kind`, `record_id` (`kind:rowid`), `title`, `updated_at`, `visibility` (`private`). `create_note` with omitted kind or `note` still returns the existing integer `id` plus those fields so `get_task` and the notes app keep working. Aliases keep their current input shapes. `update_note` / `append_to_note` / `archive_note` take `kind:rowid`.

Idempotency lives in `note_idempotency(owner_user_id, idempotency_key, kind, source_id)`. Version and `archived_at` are columns on the backing tables; archived rows are omitted from get/list/search.

### 6. Tags and FTS5 are write-through

`note_tag(owner_user_id, source_kind, source_id, tag)` with a lowercase slug, cap 10, length 32. `my_notes_fts` is a standalone FTS5 table with `owner_user_id` UNINDEXED. Writes upsert; archive deletes the FTS row. `search_my_notes` keeps only `[A-Za-z0-9_]+` tokens, joins them with `AND`, and binds `owner_user_id`. No tokens → empty page.

### 7. Grill kinds are labels and write-prompt variants

`grill.ValidKind` accepts the seven kinds. Plan/fallback questions stay title-based. The write prompt asks for timeline dated lines, research findings, or a short stick/task body. Still no SQLite in the grill.

## Risks / Trade-offs

- [SHA-256 if a low-entropy name-as-secret is ever stored] → API generates the secret; callers cannot set it.
- [Admin UI shows other users' new stickies] → Same as today. MCP is the isolation boundary this story owns.
- [FTS can lag if a write path forgets the upsert] → One helper used by every kind write.
- [Integer `id` on create_note plus `kind:id` elsewhere] → Documented; aliases exist for one phase.
- [last_used_at may skip on a RO recheck] → List still works; the security checks do not depend on it.

## Migration Plan

API `ensureTranSQLiteSchema` adds tables and columns. Existing CaseTask rows get an owner once. Rollback is drop-the-binary; leftover columns are unused. dorph-mcp never migrates. Deploy the API first, or run it once locally, then launch dorph-mcp against that file.
