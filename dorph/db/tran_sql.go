package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// TranSQL wraps an embedded SQLite connection for Tran relational data.
// TranMySQL remains a type alias so existing call sites compile during cutover.
type TranSQL struct {
	DB *sql.DB
}

// TranMySQL is an alias for TranSQL (legacy name used throughout handlers).
type TranMySQL = TranSQL

// NewTranSQL opens (or creates) an embedded SQLite database at path.
func NewTranSQL(path string) (*TranSQL, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("sqlite path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sqlite mkdir: %w", err)
	}
	// _busy_timeout and _pragma foreign_keys via DSN query params supported by modernc.
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	// SQLite is safest with a single writer connection from one process.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	if err := ensureTranSQLiteSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite schema ensure: %w", err)
	}
	return &TranSQL{DB: db}, nil
}

// NewTranMySQLLegacy opens a MySQL connection (migration / STORAGE_BACKEND=legacy only).
func NewTranMySQLLegacy(dsn string) (*TranSQL, error) {
	dsn = ensureClientFoundRowsDSN(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	if err := ensureTranMySQLSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("tran schema ensure: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &TranSQL{DB: db}, nil
}

// Close closes the SQL connection.
func (m *TranSQL) Close() error {
	if m != nil && m.DB != nil {
		return m.DB.Close()
	}
	return nil
}

func ensureClientFoundRowsDSN(dsn string) string {
	if strings.Contains(dsn, "clientFoundRows=") {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "clientFoundRows=true"
}

// sqliteHasColumn reports whether table.column exists.
func sqliteHasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(table)))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	col := strings.ToLower(column)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if strings.EqualFold(name, col) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func sqliteTableRowCount(db *sql.DB, table string) (int, error) {
	var n int
	err := db.QueryRow("SELECT COUNT(*) FROM " + quoteIdent(table)).Scan(&n)
	return n, err
}

// sqliteRenameTableIfLegacy moves oldName to newName when the old table is
// present. If an empty new table was already created, it is dropped first so
// the owner's leftover rows are not stranded.
func sqliteRenameTableIfLegacy(db *sql.DB, oldName, newName string) error {
	hasOld, err := sqliteTableExists(db, oldName)
	if err != nil || !hasOld {
		return err
	}
	hasNew, err := sqliteTableExists(db, newName)
	if err != nil {
		return err
	}
	if hasNew {
		n, err := sqliteTableRowCount(db, newName)
		if err != nil {
			return err
		}
		if n > 0 {
			return nil
		}
		if _, err := db.Exec("DROP TABLE " + quoteIdent(newName)); err != nil {
			return err
		}
	}
	_, err = db.Exec("ALTER TABLE " + quoteIdent(oldName) + " RENAME TO " + quoteIdent(newName))
	return err
}

func renameLegacyMorphTables(db *sql.DB) error {
	for _, p := range [][2]string{
		{"morph_knowledge_files", "dorph_knowledge_files"},
		{"morph_knowledge_chunks", "dorph_knowledge_chunks"},
		{"morph_agent_context_cache", "dorph_agent_context_cache"},
	} {
		if err := sqliteRenameTableIfLegacy(db, p[0], p[1]); err != nil {
			return err
		}
	}
	return nil
}

func mysqlRenameTableIfLegacy(db *sql.DB, oldName, newName string) error {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`,
		oldName,
	).Scan(&n)
	if err != nil || n == 0 {
		return err
	}
	var nNew int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`,
		newName,
	).Scan(&nNew); err != nil {
		return err
	}
	if nNew > 0 {
		return nil
	}
	_, err = db.Exec("RENAME TABLE `" + strings.ReplaceAll(oldName, "`", "") + "` TO `" + strings.ReplaceAll(newName, "`", "") + "`")
	return err
}

func sqliteTableExists(db *sql.DB, table string) (bool, error) {
	var name string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name = ? LIMIT 1`,
		table,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return name != "", nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func sqliteAddColumnIfMissing(db *sql.DB, table, column, decl string) error {
	ok, err := sqliteHasColumn(db, table, column)
	if err != nil || ok {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", quoteIdent(table), quoteIdent(column), decl))
	return err
}

// isSQLite reports whether the underlying driver is SQLite.
func (m *TranSQL) isSQLite() bool {
	if m == nil || m.DB == nil {
		return false
	}
	var v string
	err := m.DB.QueryRow(`SELECT sqlite_version()`).Scan(&v)
	return err == nil && v != ""
}

// NowRFC3339 is a small helper for timestamps written as TEXT.
func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
