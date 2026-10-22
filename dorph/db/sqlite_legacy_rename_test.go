package db

import "testing"

func TestRenameLegacyTablesMovesKnowledge(t *testing.T) {
	sqlDB := openMemorySQLite(t)
	if _, err := sqlDB.Exec(`
		CREATE TABLE morph_knowledge_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			filename TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT '',
			kind TEXT NOT NULL,
			storage_path TEXT NOT NULL,
			byte_size INTEGER NOT NULL DEFAULT 0
		)`); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO morph_knowledge_files (title, filename, kind, storage_path) VALUES ('kept', 'a.txt', 'text', 'a.txt')`); err != nil {
		t.Fatal(err)
	}
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	var title string
	if err := sqlDB.QueryRow(`SELECT title FROM dorph_knowledge_files`).Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "kept" {
		t.Fatalf("title=%q want kept", title)
	}
	hasOld, err := sqliteTableExists(sqlDB, "morph_knowledge_files")
	if err != nil {
		t.Fatal(err)
	}
	if hasOld {
		t.Fatal("legacy table still present")
	}
}

func TestRenameLegacyTablesDropsEmptyNewTable(t *testing.T) {
	sqlDB := openMemorySQLite(t)
	if _, err := sqlDB.Exec(`CREATE TABLE morph_agent_context_cache (user_id TEXT, session_id TEXT, fingerprint TEXT, blob TEXT, updated_at TEXT, PRIMARY KEY (user_id, session_id))`); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO morph_agent_context_cache VALUES ('u', 's', 'fp', 'blob', 't')`); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`CREATE TABLE dorph_agent_context_cache (user_id TEXT, session_id TEXT, fingerprint TEXT, blob TEXT, updated_at TEXT, PRIMARY KEY (user_id, session_id))`); err != nil {
		t.Fatal(err)
	}
	if err := ensureTranSQLiteSchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	var blob string
	if err := sqlDB.QueryRow(`SELECT blob FROM dorph_agent_context_cache`).Scan(&blob); err != nil {
		t.Fatal(err)
	}
	if blob != "blob" {
		t.Fatalf("blob=%q want blob", blob)
	}
}
