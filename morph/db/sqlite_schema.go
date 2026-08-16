package db

import "database/sql"

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
	return nil
}
