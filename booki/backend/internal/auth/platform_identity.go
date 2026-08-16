package auth

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// EnsurePlatformIdentity links a UsersPanel account to a Booki org + user (creates on first login).
func EnsurePlatformIdentity(db *sql.DB, email, displayName string, platformRoles []string) (userID, orgID int64, bookiRole string, err error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return 0, 0, "", fmt.Errorf("missing email")
	}
	if displayName == "" {
		displayName = email
	}
	bookiRole = mapBookiRole(platformRoles)

	var existingRole string
	err = db.QueryRow(
		`SELECT id, organization_id, role FROM users WHERE email = ? AND is_active = 1`,
		email,
	).Scan(&userID, &orgID, &existingRole)
	if err == nil {
		return userID, orgID, existingRole, nil
	}
	if err != sql.ErrNoRows {
		return 0, 0, "", err
	}

	placeholderHash, err := bcrypt.GenerateFromPassword([]byte("platform-auth-only"), bcrypt.DefaultCost)
	if err != nil {
		return 0, 0, "", err
	}

	orgLabel := displayName + "'s Ledger"
	if displayName == email {
		orgLabel = "Organization"
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, 0, "", err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO organizations (name) VALUES (?)`, orgLabel)
	if err != nil {
		return 0, 0, "", fmt.Errorf("org: %w", err)
	}
	orgID, err = res.LastInsertId()
	if err != nil {
		return 0, 0, "", err
	}
	if err := seedChartOfAccounts(tx, orgID); err != nil {
		return 0, 0, "", fmt.Errorf("chart: %w", err)
	}
	ures, err := tx.Exec(
		`INSERT INTO users (organization_id, name, email, password_hash, role) VALUES (?,?,?,?,?)`,
		orgID, displayName, email, string(placeholderHash), bookiRole,
	)
	if err != nil {
		return 0, 0, "", fmt.Errorf("user: %w", err)
	}
	userID, err = ures.LastInsertId()
	if err != nil {
		return 0, 0, "", err
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, "", err
	}
	return userID, orgID, bookiRole, nil
}
