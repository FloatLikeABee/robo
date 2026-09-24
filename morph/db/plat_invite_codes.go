package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const inviteCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const invitePasswordCharset = "abcdefghjkmnpqrstuvwxyz23456789"
const inviteUsernameCharset = "abcdefghijklmnopqrstuvwxyz"

// PlatInviteCode is a single-use signup invitation.
type PlatInviteCode struct {
	ID               string
	Code             string
	CreatedBy        string
	CreatedAt        string
	RedeemedAt       string
	RedeemedByUserID string
}

func (c PlatInviteCode) Public() map[string]any {
	out := map[string]any{
		"id":         c.ID,
		"code":       c.Code,
		"created_by": c.CreatedBy,
		"created_at": c.CreatedAt,
		"used":       c.RedeemedAt != "",
	}
	if c.RedeemedAt != "" {
		out["redeemed_at"] = c.RedeemedAt
	}
	if c.RedeemedByUserID != "" {
		out["redeemed_by_user_id"] = c.RedeemedByUserID
	}
	return out
}

// EnsurePlatInviteCodesTable creates plat_invite_codes if missing.
func (m *TranSQL) EnsurePlatInviteCodesTable(ctx context.Context) error {
	if m == nil || m.DB == nil {
		return fmt.Errorf("sql store not configured")
	}
	_, err := m.DB.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS plat_invite_codes (
    id TEXT NOT NULL PRIMARY KEY,
    code TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at TEXT NOT NULL,
    redeemed_at TEXT NULL,
    redeemed_by_user_id TEXT NULL
)`)
	if err != nil {
		return err
	}
	_, _ = m.DB.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_plat_invite_codes_code ON plat_invite_codes(code)`)
	return nil
}

func randomFromCharset(n int, charset string) (string, error) {
	if n < 1 || charset == "" {
		return "", fmt.Errorf("invalid random length")
	}
	max := big.NewInt(int64(len(charset)))
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[idx.Int64()]
	}
	return string(out), nil
}

// GenerateInviteCode returns a human-shareable single-use code string.
func GenerateInviteCode() (string, error) {
	return randomFromCharset(10, inviteCodeCharset)
}

func generateInviteUsername() (string, error) {
	suffix, err := randomFromCharset(4, inviteUsernameCharset)
	if err != nil {
		return "", err
	}
	return "user" + suffix, nil
}

func generateInvitePassword() (string, error) {
	return randomFromCharset(8, invitePasswordCharset)
}

// CreateInviteCode stores a new unused invitation code.
func (m *TranSQL) CreateInviteCode(ctx context.Context, createdBy string) (*PlatInviteCode, error) {
	if err := m.EnsurePlatInviteCodesTable(ctx); err != nil {
		return nil, err
	}
	createdBy = strings.TrimSpace(createdBy)
	if createdBy == "" {
		return nil, fmt.Errorf("created_by required")
	}
	for attempt := 0; attempt < 8; attempt++ {
		code, err := GenerateInviteCode()
		if err != nil {
			return nil, err
		}
		now := time.Now().UTC().Format(time.RFC3339)
		id := uuid.NewString()
		_, err = m.DB.ExecContext(ctx, `
INSERT INTO plat_invite_codes (id, code, created_by, created_at, redeemed_at, redeemed_by_user_id)
VALUES (?, ?, ?, ?, NULL, NULL)`, id, code, createdBy, now)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				continue
			}
			return nil, err
		}
		return &PlatInviteCode{ID: id, Code: code, CreatedBy: createdBy, CreatedAt: now}, nil
	}
	return nil, fmt.Errorf("could not generate unique invite code")
}

func (m *TranSQL) ListInviteCodes(ctx context.Context, limit int) ([]PlatInviteCode, error) {
	if err := m.EnsurePlatInviteCodesTable(ctx); err != nil {
		return nil, err
	}
	if limit < 1 {
		limit = 200
	}
	rows, err := m.DB.QueryContext(ctx, `
SELECT id, code, created_by, created_at, COALESCE(redeemed_at,''), COALESCE(redeemed_by_user_id,'')
FROM plat_invite_codes ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PlatInviteCode
	for rows.Next() {
		var c PlatInviteCode
		if err := rows.Scan(&c.ID, &c.Code, &c.CreatedBy, &c.CreatedAt, &c.RedeemedAt, &c.RedeemedByUserID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RedeemInviteCode marks a code used and creates a new plat_users row.
func (m *TranSQL) RedeemInviteCode(ctx context.Context, code string) (username, password string, err error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return "", "", fmt.Errorf("code required")
	}
	if err := m.EnsurePlatInviteCodesTable(ctx); err != nil {
		return "", "", err
	}
	if err := m.EnsurePlatUsersTable(ctx); err != nil {
		return "", "", err
	}
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = tx.Rollback() }()

	var inviteID, redeemedAt string
	err = tx.QueryRowContext(ctx, `
SELECT id, COALESCE(redeemed_at,'') FROM plat_invite_codes WHERE code = ? LIMIT 1`, code).Scan(&inviteID, &redeemedAt)
	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("invalid invite code")
	}
	if err != nil {
		return "", "", err
	}
	if redeemedAt != "" {
		return "", "", fmt.Errorf("invite code already used")
	}

	username, password, err = m.createPlatUserFromInviteTx(ctx, tx)
	if err != nil {
		return "", "", err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var userID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM plat_users WHERE username = ? LIMIT 1`, username).Scan(&userID)
	if err != nil {
		return "", "", err
	}
	res, err := tx.ExecContext(ctx, `
UPDATE plat_invite_codes SET redeemed_at = ?, redeemed_by_user_id = ? WHERE id = ? AND redeemed_at IS NULL`,
		now, userID, inviteID)
	if err != nil {
		return "", "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", "", fmt.Errorf("invite code already used")
	}
	if err := tx.Commit(); err != nil {
		return "", "", err
	}
	return username, password, nil
}

// CreatePlatUserFromInvite creates a non-admin user with generated credentials (no transaction).
func (m *TranSQL) CreatePlatUserFromInvite(ctx context.Context) (username, password string, err error) {
	if err := m.EnsurePlatUsersTable(ctx); err != nil {
		return "", "", err
	}
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = tx.Rollback() }()
	username, password, err = m.createPlatUserFromInviteTx(ctx, tx)
	if err != nil {
		return "", "", err
	}
	if err := tx.Commit(); err != nil {
		return "", "", err
	}
	return username, password, nil
}

func (m *TranSQL) createPlatUserFromInviteTx(ctx context.Context, tx *sql.Tx) (username, password string, err error) {
	for attempt := 0; attempt < 12; attempt++ {
		username, err = generateInviteUsername()
		if err != nil {
			return "", "", err
		}
		var existing string
		err = tx.QueryRowContext(ctx, `SELECT id FROM plat_users WHERE LOWER(username) = LOWER(?) LIMIT 1`, username).Scan(&existing)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			return "", "", err
		}
	}
	password, err = generateInvitePassword()
	if err != nil {
		return "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", "", err
	}
	email := strings.ToLower(username) + "@invite.local"
	now := time.Now().UTC().Format(time.RFC3339)
	uid := uuid.NewString()
	channel := "ch_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	_, err = tx.ExecContext(ctx, `
INSERT INTO plat_users (
    id, email, username, password_hash, google_id, is_verified, roles, permissions,
    default_channel_id, verification_token, verification_expires_at,
    reset_token, reset_expires_at, created_at, updated_at
) VALUES (?, ?, ?, ?, NULL, 1, ?, '[]', ?, NULL, NULL, NULL, NULL, ?, ?)`,
		uid, email, username, string(hash), rolesJSON([]string{}), channel, now, now)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}
