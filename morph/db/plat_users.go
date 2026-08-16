package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// PlatUser is a login account in plat_users (shared Morph auth).
type PlatUser struct {
	ID               string
	Email            string
	Username         string
	PasswordHash     string
	IsVerified       bool
	Roles            []string
	PermissionsJSON  string
	DefaultChannelID string
	CreatedAt        string
	UpdatedAt        string
}

func (u *PlatUser) IsAdmin() bool {
	for _, r := range u.Roles {
		if strings.EqualFold(strings.TrimSpace(r), "Admin") {
			return true
		}
	}
	return false
}

func (u *PlatUser) Public() map[string]any {
	roles := u.Roles
	if roles == nil {
		roles = []string{}
	}
	return map[string]any{
		"id":                 u.ID,
		"email":              u.Email,
		"username":           u.Username,
		"roles":              roles,
		"permissions":        []string{},
		"default_channel_id": u.DefaultChannelID,
		"is_admin":           u.IsAdmin(),
	}
}

func rolesJSON(roles []string) string {
	if roles == nil {
		roles = []string{}
	}
	b, _ := json.Marshal(roles)
	return string(b)
}

func parseRolesJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var roles []string
	if err := json.Unmarshal([]byte(raw), &roles); err != nil {
		return nil
	}
	return roles
}

// EnsurePlatUsersTable creates plat_users if missing (idempotent; SQLite + MySQL).
func (m *TranSQL) EnsurePlatUsersTable(ctx context.Context) error {
	if m == nil || m.DB == nil {
		return fmt.Errorf("sql store not configured")
	}
	if m.isSQLite() {
		_, err := m.DB.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS plat_users (
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
)`)
		if err != nil {
			return err
		}
		_, _ = m.DB.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_plat_users_email ON plat_users(email)`)
		_, _ = m.DB.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_plat_users_username ON plat_users(username)`)
		_ = sqliteAddColumnIfMissing(m.DB, "plat_users", "permissions", "TEXT DEFAULT '[]'")
		_, _ = m.DB.ExecContext(ctx, `UPDATE plat_users SET permissions = '[]' WHERE permissions IS NULL OR permissions = ''`)
		return nil
	}
	_, err := m.DB.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS plat_users (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NULL,
    google_id VARCHAR(255) NULL,
    is_verified TINYINT(1) NOT NULL DEFAULT 1,
    roles TEXT NOT NULL,
    permissions TEXT NOT NULL,
    default_channel_id VARCHAR(64) NOT NULL,
    verification_token VARCHAR(512) NULL,
    verification_expires_at VARCHAR(64) NULL,
    reset_token VARCHAR(512) NULL,
    reset_expires_at VARCHAR(64) NULL,
    created_at VARCHAR(64) NOT NULL,
    updated_at VARCHAR(64) NOT NULL,
    UNIQUE KEY uk_plat_users_email (email),
    UNIQUE KEY uk_plat_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		return err
	}
	_, _ = m.DB.ExecContext(ctx, `ALTER TABLE plat_users ADD COLUMN permissions TEXT NULL`)
	_, _ = m.DB.ExecContext(ctx, `UPDATE plat_users SET permissions = '[]' WHERE permissions IS NULL OR permissions = ''`)
	return nil
}

// EnsureBootstrapAdmin creates the env-configured admin if missing (by email or username).
// If the account exists, password is left unchanged.
func (m *TranSQL) EnsureBootstrapAdmin(ctx context.Context, email, username, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return nil
	}
	if username == "" {
		if i := strings.IndexByte(email, '@'); i > 0 {
			username = email[:i]
		} else {
			username = email
		}
	}
	if err := m.EnsurePlatUsersTable(ctx); err != nil {
		return err
	}
	var id string
	err := m.DB.QueryRowContext(ctx, `SELECT id FROM plat_users WHERE email = ? OR LOWER(username) = LOWER(?) LIMIT 1`, email, username).Scan(&id)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	uid := uuid.NewString()
	channel := "ch_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	_, err = m.DB.ExecContext(ctx, `
INSERT INTO plat_users (
    id, email, username, password_hash, google_id, is_verified, roles, permissions,
    default_channel_id, verification_token, verification_expires_at,
    reset_token, reset_expires_at, created_at, updated_at
) VALUES (?, ?, ?, ?, NULL, 1, ?, '[]', ?, NULL, NULL, NULL, NULL, ?, ?)`,
		uid, email, username, string(hash), rolesJSON([]string{"Admin"}), channel, now, now)
	return err
}

// EnsureBootstrapAdminForce creates or resets the bootstrap admin password (dev recovery).
func (m *TranSQL) EnsureBootstrapAdminForce(ctx context.Context, email, username, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return fmt.Errorf("email and password required")
	}
	if username == "" {
		if i := strings.IndexByte(email, '@'); i > 0 {
			username = email[:i]
		} else {
			username = email
		}
	}
	if err := m.EnsurePlatUsersTable(ctx); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var id string
	err = m.DB.QueryRowContext(ctx, `SELECT id FROM plat_users WHERE email = ? OR LOWER(username) = LOWER(?) LIMIT 1`, email, username).Scan(&id)
	if err == nil {
		_, err = m.DB.ExecContext(ctx, `
UPDATE plat_users SET email = ?, username = ?, password_hash = ?, roles = ?, is_verified = 1, updated_at = ?
WHERE id = ?`, email, username, string(hash), rolesJSON([]string{"Admin"}), now, id)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}
	uid := uuid.NewString()
	channel := "ch_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	_, err = m.DB.ExecContext(ctx, `
INSERT INTO plat_users (
    id, email, username, password_hash, google_id, is_verified, roles, permissions,
    default_channel_id, verification_token, verification_expires_at,
    reset_token, reset_expires_at, created_at, updated_at
) VALUES (?, ?, ?, ?, NULL, 1, ?, '[]', ?, NULL, NULL, NULL, NULL, ?, ?)`,
		uid, email, username, string(hash), rolesJSON([]string{"Admin"}), channel, now, now)
	return err
}

func (m *TranSQL) GetPlatUserByEmail(ctx context.Context, email string) (*PlatUser, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return m.scanPlatUser(ctx, `SELECT id, email, username, COALESCE(password_hash,''), is_verified, roles, COALESCE(permissions,'[]'), default_channel_id, created_at, updated_at
FROM plat_users WHERE email = ? LIMIT 1`, email)
}

func (m *TranSQL) GetPlatUserByUsername(ctx context.Context, username string) (*PlatUser, error) {
	username = strings.TrimSpace(username)
	return m.scanPlatUser(ctx, `SELECT id, email, username, COALESCE(password_hash,''), is_verified, roles, COALESCE(permissions,'[]'), default_channel_id, created_at, updated_at
FROM plat_users WHERE LOWER(username) = LOWER(?) LIMIT 1`, username)
}

// GetPlatUserByLogin finds a user by email or username.
func (m *TranSQL) GetPlatUserByLogin(ctx context.Context, login string) (*PlatUser, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, sql.ErrNoRows
	}
	if strings.Contains(login, "@") {
		return m.GetPlatUserByEmail(ctx, login)
	}
	u, err := m.GetPlatUserByUsername(ctx, login)
	if err == nil {
		return u, nil
	}
	// Also try as email if someone typed a non-@ identifier that was stored as email.
	return m.GetPlatUserByEmail(ctx, strings.ToLower(login))
}

func (m *TranSQL) GetPlatUserByID(ctx context.Context, id string) (*PlatUser, error) {
	return m.scanPlatUser(ctx, `SELECT id, email, username, COALESCE(password_hash,''), is_verified, roles, COALESCE(permissions,'[]'), default_channel_id, created_at, updated_at
FROM plat_users WHERE id = ? LIMIT 1`, id)
}

func (m *TranSQL) scanPlatUser(ctx context.Context, q string, args ...any) (*PlatUser, error) {
	if m == nil || m.DB == nil {
		return nil, fmt.Errorf("sql store not configured")
	}
	var u PlatUser
	var verified int
	var rolesRaw string
	err := m.DB.QueryRowContext(ctx, q, args...).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &verified, &rolesRaw, &u.PermissionsJSON, &u.DefaultChannelID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.IsVerified = verified != 0
	u.Roles = parseRolesJSON(rolesRaw)
	return &u, nil
}

func (m *TranSQL) ListPlatUsers(ctx context.Context, limit int) ([]PlatUser, error) {
	if m == nil || m.DB == nil {
		return nil, fmt.Errorf("mysql not configured")
	}
	if limit < 1 {
		limit = 500
	}
	rows, err := m.DB.QueryContext(ctx, `SELECT id, email, username, COALESCE(password_hash,''), is_verified, roles, COALESCE(permissions,'[]'), default_channel_id, created_at, updated_at
FROM plat_users ORDER BY email ASC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PlatUser
	for rows.Next() {
		var u PlatUser
		var verified int
		var rolesRaw string
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &verified, &rolesRaw, &u.PermissionsJSON, &u.DefaultChannelID, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.IsVerified = verified != 0
		u.Roles = parseRolesJSON(rolesRaw)
		out = append(out, u)
	}
	return out, rows.Err()
}

func (m *TranSQL) CountPlatAdmins(ctx context.Context) (int, error) {
	rows, err := m.ListPlatUsers(ctx, 5000)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, u := range rows {
		if u.IsAdmin() {
			n++
		}
	}
	return n, nil
}

func (m *TranSQL) CreatePlatUser(ctx context.Context, email, password string, isAdmin bool) (*PlatUser, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("email must be a valid email address")
	}
	if len(password) < 4 {
		return nil, fmt.Errorf("password is required")
	}
	var existing string
	err := m.DB.QueryRowContext(ctx, `SELECT id FROM plat_users WHERE email = ? OR username = ? LIMIT 1`, email, email).Scan(&existing)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, err
	}
	roles := []string{}
	if isAdmin {
		roles = []string{"Admin"}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	uid := uuid.NewString()
	channel := "ch_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	_, err = m.DB.ExecContext(ctx, `
INSERT INTO plat_users (
    id, email, username, password_hash, google_id, is_verified, roles, permissions,
    default_channel_id, verification_token, verification_expires_at,
    reset_token, reset_expires_at, created_at, updated_at
) VALUES (?, ?, ?, ?, NULL, 1, ?, '[]', ?, NULL, NULL, NULL, NULL, ?, ?)`,
		uid, email, email, string(hash), rolesJSON(roles), channel, now, now)
	if err != nil {
		return nil, err
	}
	return m.GetPlatUserByID(ctx, uid)
}

func (m *TranSQL) UpdatePlatUser(ctx context.Context, id, email, password string, isAdmin *bool) (*PlatUser, error) {
	u, err := m.GetPlatUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if email = strings.ToLower(strings.TrimSpace(email)); email != "" {
		if !strings.Contains(email, "@") {
			return nil, fmt.Errorf("email must be a valid email address")
		}
		var other string
		err := m.DB.QueryRowContext(ctx, `SELECT id FROM plat_users WHERE email = ? AND id != ? LIMIT 1`, email, id).Scan(&other)
		if err == nil {
			return nil, fmt.Errorf("email already in use")
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
		u.Email = email
		u.Username = email
	}
	hash := u.PasswordHash
	if password = strings.TrimSpace(password); password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return nil, err
		}
		hash = string(b)
	}
	roles := u.Roles
	if isAdmin != nil {
		if *isAdmin {
			roles = []string{"Admin"}
		} else {
			roles = []string{}
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = m.DB.ExecContext(ctx, `
UPDATE plat_users SET email = ?, username = ?, password_hash = ?, roles = ?, updated_at = ? WHERE id = ?`,
		u.Email, u.Username, hash, rolesJSON(roles), now, id)
	if err != nil {
		return nil, err
	}
	return m.GetPlatUserByID(ctx, id)
}

func (m *TranSQL) DeletePlatUser(ctx context.Context, id string) error {
	res, err := m.DB.ExecContext(ctx, `DELETE FROM plat_users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func VerifyPassword(hash, password string) bool {
	if strings.TrimSpace(hash) == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
