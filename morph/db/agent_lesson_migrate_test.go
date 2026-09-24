package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openMemorySQLite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func insertPlatUser(t *testing.T, db *sql.DB, id, createdAt string) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO plat_users (
			id, email, username, password_hash, is_verified, roles, permissions,
			default_channel_id, created_at, updated_at
		) VALUES (?, ?, ?, '', 1, '["Admin"]', '[]', '', ?, ?)`,
		id, id+"@example.com", id, createdAt, createdAt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAgentLessonMigrationDefaultsEnabledAndClaimsSingleOwner(t *testing.T) {
	sqlDB := openMemorySQLite(t)
	_, err := sqlDB.Exec(`
		CREATE TABLE agent_lesson (
			id TEXT NOT NULL PRIMARY KEY,
			trigger TEXT NOT NULL,
			rule TEXT NOT NULL,
			source_session_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`CREATE UNIQUE INDEX idx_agent_lesson_session ON agent_lesson(source_session_id)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at)
		VALUES ('legacy', 'when old', 'keep the rule', 'sess-old', '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}

	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner := lessonEnabledOwner(t, sqlDB, "legacy")
	if enabled != 1 || owner != "" {
		t.Fatalf("legacy row after migrate: enabled=%d owner=%q", enabled, owner)
	}

	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner = lessonEnabledOwner(t, sqlDB, "legacy")
	if enabled != 1 || owner != "" {
		t.Fatalf("second migrate changed legacy row: enabled=%d owner=%q", enabled, owner)
	}

	insertPlatUser(t, sqlDB, "only-user", "2024-06-01T00:00:00Z")
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner = lessonEnabledOwner(t, sqlDB, "legacy")
	if enabled != 1 || owner != "only-user" {
		t.Fatalf("single-user claim: enabled=%d owner=%q", enabled, owner)
	}
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner = lessonEnabledOwner(t, sqlDB, "legacy")
	if enabled != 1 || owner != "only-user" {
		t.Fatalf("claim not idempotent: enabled=%d owner=%q", enabled, owner)
	}

	insertPlatUser(t, sqlDB, "second-user", "2025-01-01T00:00:00Z")
	_, err = sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at, enabled, owner_user_id)
		VALUES ('orphan', 'when multi', 'do not leak', 'sess-new', '2025-02-01T00:00:00Z', 1, '')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner = lessonEnabledOwner(t, sqlDB, "orphan")
	if enabled != 1 || owner != "" {
		t.Fatalf("multi-user legacy row must stay unowned: enabled=%d owner=%q", enabled, owner)
	}
	_, owner = lessonEnabledOwner(t, sqlDB, "legacy")
	if owner != "only-user" {
		t.Fatalf("already claimed lesson changed owner to %q", owner)
	}

	var oldIndex string
	err = sqlDB.QueryRow(`SELECT name FROM sqlite_master WHERE type='index' AND name='idx_agent_lesson_session'`).Scan(&oldIndex)
	if err != sql.ErrNoRows {
		t.Fatalf("old session-only unique index still present: %v %q", err, oldIndex)
	}
	_, err = sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at, enabled, owner_user_id)
		VALUES ('shared-a', 'when', 'a', 'default', '2025-03-01T00:00:00Z', 1, 'only-user')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at, enabled, owner_user_id)
		VALUES ('shared-b', 'when', 'b', 'default', '2025-03-01T00:00:01Z', 1, 'second-user')`)
	if err != nil {
		t.Fatalf("two owners must be able to share a session id: %v", err)
	}
}

func TestAgentLessonFreshSchemaDefaultsEnabled(t *testing.T) {
	sqlDB := openMemorySQLite(t)
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	_, err := sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at)
		VALUES ('fresh', 'when new', 'rule', 'sess', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	var platUsers string
	if err := sqlDB.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='plat_users'`).Scan(&platUsers); err != nil || platUsers != "plat_users" {
		t.Fatalf("brand-new database must include plat_users: %v %q", err, platUsers)
	}
	enabled, owner := lessonEnabledOwner(t, sqlDB, "fresh")
	if enabled != 1 || owner != "" {
		t.Fatalf("column default: enabled=%d owner=%q", enabled, owner)
	}
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner = lessonEnabledOwner(t, sqlDB, "fresh")
	if enabled != 1 || owner != "" {
		t.Fatalf("rerun changed fresh row: enabled=%d owner=%q", enabled, owner)
	}
}

func TestMigrateAgentLessonColumnsSkipsBackfillWhenPlatUsersMissing(t *testing.T) {
	sqlDB := openMemorySQLite(t)
	_, err := sqlDB.Exec(`
		CREATE TABLE agent_lesson (
			id TEXT NOT NULL PRIMARY KEY,
			trigger TEXT NOT NULL,
			rule TEXT NOT NULL,
			source_session_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at)
		VALUES ('partial', 'when partial', 'keep', 'sess', '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrateAgentLessonColumns(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner := lessonEnabledOwner(t, sqlDB, "partial")
	if enabled != 1 || owner != "" {
		t.Fatalf("partial db: enabled=%d owner=%q", enabled, owner)
	}
	if err := migrateAgentLessonColumns(sqlDB); err != nil {
		t.Fatal(err)
	}
	enabled, owner = lessonEnabledOwner(t, sqlDB, "partial")
	if enabled != 1 || owner != "" {
		t.Fatalf("second partial migrate: enabled=%d owner=%q", enabled, owner)
	}
}

func lessonEnabledOwner(t *testing.T, db *sql.DB, id string) (int, string) {
	t.Helper()
	var enabled int
	var owner string
	if err := db.QueryRow(`SELECT enabled, owner_user_id FROM agent_lesson WHERE id=?`, id).Scan(&enabled, &owner); err != nil {
		t.Fatal(err)
	}
	return enabled, owner
}
