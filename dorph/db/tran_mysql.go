package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// ensureTranMySQLSchema applies small idempotent fixes for legacy MySQL snapshots
// (migrate tool / STORAGE_BACKEND=legacy only).
func ensureTranMySQLSchema(db *sql.DB) error {
	if err := ensureBigNotesSchemaMySQL(db); err != nil {
		return err
	}
	if err := ensureTimelinesSchemaMySQL(db); err != nil {
		return err
	}
	var tableCount int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = 'facility'`).Scan(&tableCount); err != nil {
		return err
	}
	if tableCount == 0 {
		return nil
	}
	var colCount int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'facility' AND column_name = 'location'`).Scan(&colCount); err != nil {
		return err
	}
	if colCount == 0 {
		if _, err := db.Exec("ALTER TABLE `facility` ADD COLUMN location JSON NULL"); err != nil {
			return err
		}
	}
	if err := (&TranSQL{DB: db}).EnsureGraphKnowledgeSchema(); err != nil {
		return err
	}
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS StoryPost (
		  ID INT NOT NULL AUTO_INCREMENT,
		  title VARCHAR(255) NOT NULL,
		  content TEXT NOT NULL,
		  author_user_id INT NOT NULL,
		  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (ID),
		  KEY ix_story_post_author (author_user_id),
		  KEY ix_story_post_created (created_on)
		)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS generic_data (
		  id INT NOT NULL AUTO_INCREMENT,
		  title VARCHAR(255) NOT NULL,
		  source_type ENUM('csv', 'json', 'pdf') NOT NULL,
		  source_filename VARCHAR(500) NULL,
		  record_count INT NOT NULL DEFAULT 0,
		  description TEXT NULL,
		  ai_analysis TEXT NULL,
		  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (id),
		  KEY ix_generic_data_source_type (source_type),
		  KEY ix_generic_data_created_on (created_on)
		)`)
	return err
}

func ensureBigNotesSchemaMySQL(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS big_note (
		  id INT NOT NULL AUTO_INCREMENT,
		  user_id INT NOT NULL,
		  owner_key VARCHAR(191) NOT NULL DEFAULT '',
		  title VARCHAR(255) NOT NULL,
		  idea TEXT NOT NULL,
		  note_kind VARCHAR(32) NOT NULL DEFAULT 'note',
		  markdown_content LONGTEXT NOT NULL,
		  html_content LONGTEXT NOT NULL,
		  questions_json LONGTEXT NULL,
		  theme VARCHAR(64) NOT NULL DEFAULT 'default',
		  published_slug VARCHAR(255) NULL,
		  published_path VARCHAR(512) NULL,
		  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (id),
		  KEY ix_big_note_user (user_id),
		  KEY ix_big_note_owner (owner_key),
		  KEY ix_big_note_created (created_on)
		)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS big_note_response (
		  id INT NOT NULL AUTO_INCREMENT,
		  big_note_id INT NOT NULL,
		  answers_json LONGTEXT NOT NULL,
		  analysis_markdown LONGTEXT NULL,
		  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (id),
		  KEY ix_big_note_response_note (big_note_id),
		  KEY ix_big_note_response_created (created_on)
		)`)
	if err != nil {
		return err
	}
	for _, stmt := range []string{
		`ALTER TABLE big_note ADD COLUMN owner_key VARCHAR(191) NOT NULL DEFAULT ''`,
		`ALTER TABLE big_note ADD COLUMN note_kind VARCHAR(32) NOT NULL DEFAULT 'note'`,
		`ALTER TABLE big_note ADD COLUMN questions_json LONGTEXT NULL`,
	} {
		if _, err := db.Exec(stmt); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			if !strings.Contains(err.Error(), "1060") {
				return err
			}
		}
	}
	return nil
}

func ensureTimelinesSchemaMySQL(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS timeline (
		  id INT NOT NULL AUTO_INCREMENT,
		  user_id INT NOT NULL,
		  owner_key VARCHAR(191) NOT NULL DEFAULT '',
		  title VARCHAR(255) NOT NULL,
		  source_summary TEXT NOT NULL,
		  source_file_name VARCHAR(512) NULL,
		  source_url VARCHAR(2048) NULL,
		  has_paste TINYINT(1) NOT NULL DEFAULT 0,
		  markdown_content LONGTEXT NOT NULL,
		  html_content LONGTEXT NOT NULL,
		  published_slug VARCHAR(255) NULL,
		  published_path VARCHAR(512) NULL,
		  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (id),
		  KEY ix_timeline_user (user_id),
		  KEY ix_timeline_owner (owner_key),
		  KEY ix_timeline_created (created_on),
		  UNIQUE KEY uq_timeline_published_slug (published_slug)
		)`)
	return err
}

// Deprecated: use NewTranSQL for embedded or NewTranMySQLLegacy for migrate/legacy.
func NewTranMySQL(dsn string) (*TranSQL, error) {
	return NewTranMySQLLegacy(dsn)
}

// IsSQLiteDSN is unused helper reserved for dialect detection.
func IsSQLitePath(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".sqlite") || strings.HasSuffix(strings.ToLower(path), ".db")
}

var _ = fmt.Sprintf
