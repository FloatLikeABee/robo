package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// RepairLegacySchema adds columns expected by the app if an older table exists
// (CREATE TABLE IF NOT EXISTS does not alter existing tables).
func RepairLegacySchema(db *sql.DB) error {
	alters := []string{
		"ADD COLUMN organization_id INTEGER NULL",
		"ADD COLUMN name TEXT NULL",
		"ADD COLUMN email TEXT NULL",
		"ADD COLUMN password_hash TEXT NULL",
		"ADD COLUMN role TEXT NULL DEFAULT 'owner'",
		"ADD COLUMN is_active INTEGER DEFAULT 1",
		"ADD COLUMN created_at TEXT NULL DEFAULT CURRENT_TIMESTAMP",
		"ADD COLUMN updated_at TEXT NULL DEFAULT CURRENT_TIMESTAMP",
	}
	for _, frag := range alters {
		q := "ALTER TABLE users " + frag
		if _, err := db.Exec(q); err != nil && !isSQLiteDuplicateColumn(err) {
			if !isSQLiteNoSuchTable(err) {
				return fmt.Errorf("repair users: %s: %w", q, err)
			}
		}
	}
	whFragments := []string{
		"ADD COLUMN detail TEXT NULL",
		"ADD COLUMN record_date TEXT NULL DEFAULT (CURRENT_DATE)",
	}
	for _, frag := range whFragments {
		q := "ALTER TABLE warehouses " + frag
		if _, err := db.Exec(q); err != nil && !isSQLiteDuplicateColumn(err) {
			if !isSQLiteNoSuchTable(err) {
				return fmt.Errorf("repair warehouses: %s: %w", q, err)
			}
		}
	}
	prodFragments := []string{
		"ADD COLUMN description TEXT NOT NULL DEFAULT ''",
		"ADD COLUMN detail TEXT NULL",
		"ADD COLUMN record_date TEXT NULL DEFAULT (CURRENT_DATE)",
	}
	for _, frag := range prodFragments {
		q := "ALTER TABLE products " + frag
		if _, err := db.Exec(q); err != nil && !isSQLiteDuplicateColumn(err) {
			if !isSQLiteNoSuchTable(err) {
				return fmt.Errorf("repair products: %s: %w", q, err)
			}
		}
	}
	return nil
}

func isSQLiteDuplicateColumn(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate column name")
}

func isSQLiteNoSuchTable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table")
}
