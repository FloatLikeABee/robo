package auth

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DevUserEmail is the auto-seeded account used with POST /auth/dev-login (development only).
// Matches UsersPanel default bootstrap admin so you can use one identity across Tran* apps and Booki locally.
const DevUserEmail = "admin@example.com"

// DevUserPassword must match UsersPanel dev default (see UsersPanel bootstrap) unless you override BOOTSTRAP_*.
const DevUserPassword = "AdminExample2026!"

const devOrgName = "Local Dev"

// EnsureDevIdentity creates a default org + owner if none exists (development convenience).
func EnsureDevIdentity(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, DevUserEmail).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(DevUserPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO organizations (name) VALUES (?)`, devOrgName)
	if err != nil {
		return fmt.Errorf("dev org: %w", err)
	}
	orgID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if err := seedChartOfAccounts(tx, orgID); err != nil {
		return fmt.Errorf("dev chart: %w", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO users (organization_id, name, email, password_hash, role) VALUES (?,?,?,?, 'owner')`,
		orgID, "Dev User", DevUserEmail, string(hash),
	); err != nil {
		return fmt.Errorf("dev user: %w", err)
	}
	return tx.Commit()
}
