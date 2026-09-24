package db

import (
	"database/sql"
	"log"
)

// ensureTranSQLiteSchema creates Morph-owned relational tables for embedded SQLite.
// Column sets follow openspec/changes/morph-embedded-dbs/inventory.md (handler current usage).
func ensureTranSQLiteSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS District (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			DistrictID INTEGER NOT NULL DEFAULT 0,
			District TEXT NOT NULL DEFAULT '',
			Name TEXT NULL,
			Description TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS plat_users (
			id TEXT NOT NULL PRIMARY KEY,
			email TEXT NOT NULL,
			username TEXT NOT NULL,
			password_hash TEXT NULL,
			google_id TEXT NULL,
			is_verified INTEGER NOT NULL DEFAULT 1,
			roles TEXT NOT NULL,
			permissions TEXT NOT NULL DEFAULT '[]',
			default_channel_id TEXT NOT NULL,
			verification_token TEXT NULL,
			verification_expires_at TEXT NULL,
			reset_token TEXT NULL,
			reset_expires_at TEXT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_plat_users_email ON plat_users(email)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_plat_users_username ON plat_users(username)`,
		`CREATE TABLE IF NOT EXISTS plat_invite_codes (
			id TEXT NOT NULL PRIMARY KEY,
			code TEXT NOT NULL,
			created_by TEXT NOT NULL,
			created_at TEXT NOT NULL,
			redeemed_at TEXT NULL,
			redeemed_by_user_id TEXT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_plat_invite_codes_code ON plat_invite_codes(code)`,
		`CREATE TABLE IF NOT EXISTS "User" (
			UserID INTEGER PRIMARY KEY AUTOINCREMENT,
			LoginID TEXT NULL,
			FirstName TEXT NULL,
			LastName TEXT NOT NULL DEFAULT '',
			Email TEXT NULL,
			Phone TEXT NULL,
			Title TEXT NULL,
			Administrator INTEGER NOT NULL DEFAULT 0,
			Deactivated INTEGER NOT NULL DEFAULT 0,
			DeactivatedDate TEXT NULL,
			MessageAiAutoReplyEnabled INTEGER NOT NULL DEFAULT 0,
			MessageAiAutoReplyPrompt TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS contact (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			LastName TEXT NOT NULL DEFAULT '',
			FirstName TEXT NULL,
			Email TEXT NULL,
			Phone TEXT NULL,
			Mobile TEXT NULL,
			description TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS facility (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			facility_code TEXT NOT NULL,
			name TEXT NULL,
			district_id INTEGER NULL,
			facility_type TEXT NULL,
			description TEXT NULL,
			location TEXT NULL,
			capacity INTEGER NULL DEFAULT 0,
			x_coord REAL NULL,
			y_coord REAL NULL,
			guid TEXT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS UQ_facility_facility_code ON facility(facility_code)`,
		`CREATE TABLE IF NOT EXISTS "member" (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			last_name TEXT NULL,
			first_name TEXT NULL,
			middle_name TEXT NULL,
			dob TEXT NULL,
			entry_date TEXT NULL,
			facility TEXT NULL,
			gender INTEGER NULL,
			email TEXT NULL,
			participant_type TEXT NULL,
			description TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS employee (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			last_name TEXT NOT NULL,
			first_name TEXT NULL,
			middle_name TEXT NULL,
			staff_guid TEXT NULL,
			active_flag INTEGER NOT NULL DEFAULT 1,
			inactive_date TEXT NULL,
			contractor_id INTEGER NULL DEFAULT 0,
			email TEXT NULL,
			phone_number TEXT NULL,
			date_of_birth TEXT NULL,
			gender INTEGER NULL,
			user_id INTEGER NULL,
			employ_type TEXT NULL,
			description TEXT NULL,
			facility_id INTEGER NULL
		)`,
		`CREATE TABLE IF NOT EXISTS Asset (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			ContractorID INTEGER NULL,
			asset_tag TEXT NULL,
			description TEXT NULL,
			AssetID TEXT NULL,
			AssetType TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS Activity (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL DEFAULT '',
			ActivityType TEXT NULL,
			start_date TEXT NULL,
			end_date TEXT NULL,
			location TEXT NULL,
			GUID TEXT NULL,
			description TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ActivityEmployee (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			employee_id INTEGER NOT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(activity_id, employee_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ActivityParticipant (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(activity_id, member_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ActivityAsset (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			asset_id INTEGER NOT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(activity_id, asset_id)
		)`,
		`CREATE TABLE IF NOT EXISTS CaseTask (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NULL,
			start_at TEXT NULL,
			end_at TEXT NULL,
			location TEXT NULL,
			assignee_type TEXT NOT NULL DEFAULT '',
			assignee_id INTEGER NOT NULL DEFAULT 0,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS CaseTaskAssignee (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			case_task_id INTEGER NOT NULL,
			assignee_kind TEXT NOT NULL,
			assignee_id INTEGER NOT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS StoryPost (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			author_user_id INTEGER NOT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS comment (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			EntityType TEXT NOT NULL,
			RecordID INTEGER NOT NULL,
			ParentID INTEGER NULL,
			AuthorUserID INTEGER NULL,
			Body TEXT NOT NULL,
			CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP,
			LastUpdated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS generic_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			source_type TEXT NOT NULL,
			source_filename TEXT NULL,
			record_count INTEGER NOT NULL DEFAULT 0,
			description TEXT NULL,
			ai_analysis TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS EntityAttachment (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_type TEXT NOT NULL,
			record_id INTEGER NOT NULL,
			original_name TEXT NOT NULL,
			stored_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			mime_type TEXT NULL,
			size_bytes INTEGER NOT NULL DEFAULT 0,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS AdminGridSavedFilter (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			GridKey TEXT NOT NULL,
			Name TEXT NOT NULL,
			FilterJSON TEXT NOT NULL,
			CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS AdminGridColorConfig (
			GridKey TEXT NOT NULL PRIMARY KEY,
			ConfigJSON TEXT NOT NULL,
			UpdatedOn TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS PlatformUiConfig (
			ID INTEGER NOT NULL PRIMARY KEY,
			ConfigJSON TEXT NOT NULL,
			UpdatedOn TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_note_todo (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			UserID INTEGER NOT NULL,
			ItemType TEXT NOT NULL,
			Title TEXT NULL,
			Body TEXT NULL,
			Completed INTEGER NOT NULL DEFAULT 0,
			DeadlineAt TEXT NULL,
			CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP,
			LastUpdated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS record_contact (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			DBID INTEGER NOT NULL DEFAULT 0,
			EntityType TEXT NOT NULL,
			RecordID INTEGER NOT NULL,
			ContactID INTEGER NOT NULL,
			Relationship TEXT NULL,
			IsPrimary INTEGER NOT NULL DEFAULT 0,
			CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tool_note (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			AuthorUserID INTEGER NOT NULL,
			TargetType TEXT NOT NULL,
			TargetUserID INTEGER NULL,
			IsPrivate INTEGER NOT NULL DEFAULT 0,
			Title TEXT NULL,
			Body TEXT NULL,
			ReadAt TEXT NULL,
			CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP,
			LastUpdated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tool_broadcast (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			AuthorUserID INTEGER NOT NULL,
			Title TEXT NULL,
			Body TEXT NULL,
			AttachmentsJSON TEXT NULL,
			CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP,
			LastUpdated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS big_note (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			owner_key TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			idea TEXT NOT NULL DEFAULT '',
			note_kind TEXT NOT NULL DEFAULT 'note',
			markdown_content TEXT NOT NULL DEFAULT '',
			html_content TEXT NOT NULL DEFAULT '',
			questions_json TEXT NULL,
			theme TEXT NOT NULL DEFAULT 'default',
			published_slug TEXT NULL,
			published_path TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS big_note_response (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			big_note_id INTEGER NOT NULL,
			answers_json TEXT NOT NULL,
			analysis_markdown TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS timeline (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			owner_key TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			source_summary TEXT NOT NULL DEFAULT '',
			source_file_name TEXT NULL,
			source_url TEXT NULL,
			has_paste INTEGER NOT NULL DEFAULT 0,
			markdown_content TEXT NOT NULL DEFAULT '',
			html_content TEXT NOT NULL DEFAULT '',
			published_slug TEXT NULL,
			published_path TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS research (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			owner_key TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			prompt TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'ingesting',
			current_round INTEGER NOT NULL DEFAULT 0,
			round_target INTEGER NOT NULL DEFAULT 5,
			markdown_content TEXT NOT NULL DEFAULT '',
			html_content TEXT NOT NULL DEFAULT '',
			error_text TEXT NULL,
			published_slug TEXT NULL,
			published_path TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS research_file (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			research_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			kind TEXT NOT NULL,
			text_excerpt TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS research_chunk (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			research_id INTEGER NOT NULL,
			file_id INTEGER NOT NULL DEFAULT 0,
			chunk_index INTEGER NOT NULL,
			text_content TEXT NOT NULL,
			embedding_json TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS research_piece (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			research_id INTEGER NOT NULL,
			round_index INTEGER NOT NULL,
			markdown TEXT NOT NULL DEFAULT '',
			verification TEXT NOT NULL DEFAULT '',
			sources_json TEXT NULL,
			status TEXT NOT NULL DEFAULT 'ok',
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(research_id, round_index)
		)`,
		`CREATE TABLE IF NOT EXISTS morph_knowledge_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			filename TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT '',
			kind TEXT NOT NULL,
			storage_path TEXT NOT NULL,
			byte_size INTEGER NOT NULL DEFAULT 0,
			text_excerpt TEXT NULL,
			created_by TEXT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS morph_knowledge_chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			chunk_index INTEGER NOT NULL,
			text_content TEXT NOT NULL,
			embedding_json TEXT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(file_id, chunk_index)
		)`,
		`CREATE TABLE IF NOT EXISTS graph_sync_outbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			op TEXT NOT NULL,
			payload_json TEXT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			available_at TEXT DEFAULT CURRENT_TIMESTAMP,
			attempts INTEGER NOT NULL DEFAULT 0,
			locked_by TEXT NULL,
			locked_at TEXT NULL,
			processed_at TEXT NULL,
			last_error TEXT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS neo4j_ingest_outbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kind TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			payload_json TEXT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			last_error TEXT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_neo4j_ingest_outbox_status ON neo4j_ingest_outbox(status, id)`,
		`CREATE TABLE IF NOT EXISTS ai_skills (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			owner_user_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_skills_enabled ON ai_skills(enabled)`,
		`CREATE TABLE IF NOT EXISTS agent_lesson (
			id TEXT NOT NULL PRIMARY KEY,
			trigger TEXT NOT NULL,
			rule TEXT NOT NULL,
			source_session_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			owner_user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_lesson_created ON agent_lesson(created_at)`,
		`CREATE TABLE IF NOT EXISTS morph_agent_context_cache (
			user_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			blob TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (user_id, session_id)
		)`,
		`CREATE TABLE IF NOT EXISTS StaffType (
			StaffTypeID INTEGER PRIMARY KEY AUTOINCREMENT,
			StaffTypeName TEXT NOT NULL,
			StaffTypeDescription TEXT NULL,
			IsSystemDefined INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS StaffStaffType (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			StaffID INTEGER NOT NULL,
			StaffTypeID INTEGER NOT NULL,
			PrimaryFlag INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS email_agent (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			config_json TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS email_agent_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_agent_id INTEGER NOT NULL,
			message TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_message_auto_reply_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			message TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	// Idempotent column adds for older SQLite files.
	_ = sqliteAddColumnIfMissing(db, "facility", "location", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "facility", "description", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "member", "description", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "employee", "description", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "employee", "facility_id", "INTEGER NULL")
	_ = sqliteAddColumnIfMissing(db, "big_note", "owner_key", "TEXT NOT NULL DEFAULT ''")
	_ = sqliteAddColumnIfMissing(db, "big_note", "note_kind", "TEXT NOT NULL DEFAULT 'note'")
	_ = sqliteAddColumnIfMissing(db, "big_note", "questions_json", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "user_note_todo", "DeadlineAt", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "User", "DeactivatedDate", "TEXT NULL")
	_ = sqliteAddColumnIfMissing(db, "research", "round_target", "INTEGER NOT NULL DEFAULT 5")
	_, _ = db.Exec(`UPDATE research SET round_target = 20 WHERE current_round > 5 OR id IN (SELECT research_id FROM research_piece GROUP BY research_id HAVING COUNT(*) > 5)`)
	if err := migrateAgentLessonColumns(db); err != nil {
		return err
	}
	return nil
}

// migrateAgentLessonColumns adds per-user ownership and the enabled flag.
// Existing rows default to enabled=1. Lessons predate ownership, so
// owner_user_id starts as ”. The legacy claim is decided once
// (agent_lesson_owner_backfill): zero accounts waits for a later startup,
// exactly one account receives the unowned rows that exist at that moment,
// and more than one account records a decision that never assigns. A later
// drop to a single account must not pick up historical unowned lessons.
// If plat_users is missing, the claim is skipped and startup continues.
// Distillation records the owner for lessons created after this migration.
// Uniqueness is per owner and source session so two users can each keep a
// lesson from session id "default". The composite unique index is created
// before the old session-only index is dropped, so a failure in between does
// not leave the table without a unique key. A legacy row whose session the
// claimed user already owns is deleted inside the claim; the user's lesson
// stays. A claim error is logged and does not fail startup.
func migrateAgentLessonColumns(db *sql.DB) error {
	if err := sqliteAddColumnIfMissing(db, "agent_lesson", "enabled", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := sqliteAddColumnIfMissing(db, "agent_lesson", "owner_user_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_lesson_owner_session ON agent_lesson(owner_user_id, source_session_id)`); err != nil {
		return err
	}
	if _, err := db.Exec(`DROP INDEX IF EXISTS idx_agent_lesson_session`); err != nil {
		return err
	}
	// ensureTranSQLiteSchema creates plat_users before this runs, including on a
	// brand-new database. Skip the claim when that table is absent so a partial
	// SQLite file cannot fail startup, and do not record a decision yet.
	hasUsers, err := sqliteTableExists(db, "plat_users")
	if err != nil || !hasUsers {
		return err
	}
	if err := claimLegacyAgentLessonOwners(db); err != nil {
		log.Printf("agent lesson owner claim skipped: %v", err)
		return nil
	}
	return nil
}

// claimLegacyAgentLessonOwners assigns pre-ownership lessons at most once.
// A missing decision row with zero accounts returns without writing, so the
// bootstrap admin created after the first schema pass can claim on a later start.
func claimLegacyAgentLessonOwners(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_lesson_owner_backfill (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			decided_at TEXT NOT NULL,
			claimed_user_id TEXT NOT NULL DEFAULT ''
		)`); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var decided int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM agent_lesson_owner_backfill WHERE id = 1`).Scan(&decided); err != nil {
		return err
	}
	if decided > 0 {
		return tx.Commit()
	}
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM plat_users`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return tx.Commit()
	}
	claimed := ""
	if n == 1 {
		if err := tx.QueryRow(`SELECT id FROM plat_users ORDER BY created_at ASC, id ASC LIMIT 1`).Scan(&claimed); err != nil {
			return err
		}
		if err := deleteLegacyLessonsConflictingWithOwner(tx, claimed); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE agent_lesson SET owner_user_id = ? WHERE owner_user_id = ''`, claimed); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(
		`INSERT INTO agent_lesson_owner_backfill (id, decided_at, claimed_user_id) VALUES (1, datetime('now'), ?)`,
		claimed,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// deleteLegacyLessonsConflictingWithOwner removes unowned rows whose session
// the claimed user already has a lesson for. The owned row is the one harvest
// stored after the account existed; assigning the legacy row would violate
// idx_agent_lesson_owner_session and stop startup.
func deleteLegacyLessonsConflictingWithOwner(tx *sql.Tx, ownerID string) error {
	rows, err := tx.Query(`
		SELECT legacy.id
		FROM agent_lesson AS legacy
		INNER JOIN agent_lesson AS owned
		  ON owned.source_session_id = legacy.source_session_id
		 AND owned.owner_user_id = ?
		WHERE legacy.owner_user_id = ''`, ownerID)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := tx.Exec(`DELETE FROM agent_lesson WHERE id = ? AND owner_user_id = ''`, id); err != nil {
			return err
		}
	}
	return nil
}
