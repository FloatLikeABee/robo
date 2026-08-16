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
