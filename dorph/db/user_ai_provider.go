package db

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// UserAIProvider is one saved chat provider for a signed-in user.
type UserAIProvider struct {
	UserID   string
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

func (s *TranSQL) GetUserAIProvider(ctx context.Context, userID string) (UserAIProvider, bool, error) {
	var row UserAIProvider
	if s == nil || s.DB == nil {
		return row, false, sql.ErrConnDone
	}
	err := s.DB.QueryRowContext(ctx, `
SELECT user_id, provider, base_url, api_key, model
FROM user_ai_provider WHERE user_id = ?`, strings.TrimSpace(userID)).
		Scan(&row.UserID, &row.Provider, &row.BaseURL, &row.APIKey, &row.Model)
	if err == sql.ErrNoRows {
		return UserAIProvider{}, false, nil
	}
	if err != nil {
		return UserAIProvider{}, false, err
	}
	return row, true, nil
}

// SaveUserAIProvider upserts the row. An empty APIKey keeps the stored key.
func (s *TranSQL) SaveUserAIProvider(ctx context.Context, row UserAIProvider) error {
	if s == nil || s.DB == nil {
		return sql.ErrConnDone
	}
	row.UserID = strings.TrimSpace(row.UserID)
	row.Provider = strings.TrimSpace(row.Provider)
	row.BaseURL = strings.TrimSpace(row.BaseURL)
	row.Model = strings.TrimSpace(row.Model)
	key := row.APIKey
	if strings.TrimSpace(key) == "" {
		if prev, ok, err := s.GetUserAIProvider(ctx, row.UserID); err != nil {
			return err
		} else if ok {
			key = prev.APIKey
		}
	}
	_, err := s.DB.ExecContext(ctx, `
INSERT INTO user_ai_provider (user_id, provider, base_url, api_key, model, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
  provider = excluded.provider,
  base_url = excluded.base_url,
  api_key = excluded.api_key,
  model = excluded.model,
  updated_at = excluded.updated_at`,
		row.UserID, row.Provider, row.BaseURL, key, row.Model, time.Now().UTC().Format(time.RFC3339))
	return err
}
