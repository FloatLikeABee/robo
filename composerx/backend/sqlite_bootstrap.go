package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema_sqlite.sql
var schemaSQLiteFS embed.FS

func openComposerXSQLite(path string) (*sql.DB, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("sqlite path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sqlite mkdir: %w", err)
	}
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	if err := applySchemaSQLite(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite schema: %w", err)
	}
	if err := seedBuiltinUser(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite seed user: %w", err)
	}
	return db, nil
}

func applySchemaSQLite(db *sql.DB) error {
	raw, err := schemaSQLiteFS.ReadFile("schema_sqlite.sql")
	if err != nil {
		return err
	}
	for _, stmt := range splitSQLStatements(string(raw)) {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", truncateSQL(stmt, 80), err)
		}
	}
	return nil
}

func seedBuiltinUser(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := db.Exec(
		`INSERT INTO users (id, email, name) VALUES (1, '__builtin_seed@local', 'System')`,
	)
	return err
}

func splitSQLStatements(script string) []string {
	lines := strings.Split(script, "\n")
	var cleaned []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "--") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	joined := strings.Join(cleaned, "\n")
	parts := strings.Split(joined, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func truncateSQL(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
