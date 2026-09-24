package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"idongivaflyinfa/config"

	"golang.org/x/crypto/bcrypt"
)

// HasStoredDefaultAdminPassword reports whether an admin or the bootstrap
// identity still accepts the published development password.
func (m *TranSQL) HasStoredDefaultAdminPassword(ctx context.Context, email, username string) (bool, error) {
	ids, err := m.defaultAdminPasswordUserIDs(ctx, email, username)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

// GuardStoredDefaultAdminPassword refuses startup when a privileged account
// still accepts the development password. When rotate is true and newPassword
// is strong, it replaces those hashes and keeps each account id.
// The error text does not include password values.
func (m *TranSQL) GuardStoredDefaultAdminPassword(ctx context.Context, email, username string, rotate bool, newPassword string) (int, error) {
	ids, err := m.defaultAdminPasswordUserIDs(ctx, email, username)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if !rotate {
		return 0, fmt.Errorf("MORPH_ENV=production: an admin account in the database still accepts the development default password. Set ADMIN_PASSWORD to a unique password and start once with MORPH_ROTATE_DEFAULT_ADMIN=1 to replace it, then unset MORPH_ROTATE_DEFAULT_ADMIN. See docs/security-hosting-checklist.md")
	}
	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" || strings.EqualFold(newPassword, config.DefaultAdminPassword) || len(newPassword) < config.MinAdminPasswordLength {
		return 0, fmt.Errorf("MORPH_ENV=production: MORPH_ROTATE_DEFAULT_ADMIN is set, but ADMIN_PASSWORD is missing, still the development default, or shorter than %d characters. Set ADMIN_PASSWORD to a unique password of at least %d characters", config.MinAdminPasswordLength, config.MinAdminPasswordLength)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, id := range ids {
		if _, err := m.DB.ExecContext(ctx, `UPDATE plat_users SET password_hash = ?, updated_at = ? WHERE id = ?`, string(hash), now, id); err != nil {
			return 0, err
		}
	}
	left, err := m.defaultAdminPasswordUserIDs(ctx, email, username)
	if err != nil {
		return 0, err
	}
	if len(left) > 0 {
		return 0, fmt.Errorf("MORPH_ENV=production: an admin account in the database still accepts the development default password. Update that account before hosting")
	}
	return len(ids), nil
}

func (m *TranSQL) defaultAdminPasswordUserIDs(ctx context.Context, email, username string) ([]string, error) {
	if m == nil || m.DB == nil {
		return nil, fmt.Errorf("sql store not configured")
	}
	rows, err := m.DB.QueryContext(ctx, `SELECT id, email, username, COALESCE(password_hash,''), roles FROM plat_users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.ToLower(strings.TrimSpace(username))
	var ids []string
	for rows.Next() {
		var id, rowEmail, rowUser, hash, rolesRaw string
		if err := rows.Scan(&id, &rowEmail, &rowUser, &hash, &rolesRaw); err != nil {
			return nil, err
		}
		if !defaultAdminCandidate(email, username, rowEmail, rowUser, parseRolesJSON(rolesRaw)) {
			continue
		}
		if !VerifyPassword(hash, config.DefaultAdminPassword) {
			continue
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func defaultAdminCandidate(cfgEmail, cfgUser, rowEmail, rowUser string, roles []string) bool {
	if (&PlatUser{Roles: roles}).IsAdmin() {
		return true
	}
	if strings.EqualFold(rowEmail, config.DefaultAdminEmail) || strings.EqualFold(rowUser, config.DefaultAdminUsername) {
		return true
	}
	if cfgEmail != "" && strings.EqualFold(rowEmail, cfgEmail) {
		return true
	}
	if cfgUser != "" && strings.EqualFold(rowUser, cfgUser) {
		return true
	}
	return false
}
