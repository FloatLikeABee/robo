package db

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUserAIProviderMasksAndKeepsKey(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if _, err := sqlDB.Exec(`CREATE TABLE user_ai_provider (
		user_id TEXT NOT NULL PRIMARY KEY,
		provider TEXT NOT NULL,
		base_url TEXT NOT NULL DEFAULT '',
		api_key TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	store := &TranSQL{DB: sqlDB}
	ctx := context.Background()
	if err := store.SaveUserAIProvider(ctx, UserAIProvider{
		UserID: "u1", Provider: "openai", BaseURL: "https://api.example/v1", APIKey: "secret-key", Model: "gpt-4.1",
	}); err != nil {
		t.Fatal(err)
	}
	got, ok, err := store.GetUserAIProvider(ctx, "u1")
	if err != nil || !ok {
		t.Fatalf("get: %v ok=%v", err, ok)
	}
	view := map[string]any{
		"provider": got.Provider,
		"base_url": got.BaseURL,
		"model":    got.Model,
		"key_set":  got.APIKey != "",
	}
	if view["key_set"] != true {
		t.Fatal("key_set")
	}
	if _, leaked := view["api_key"]; leaked {
		t.Fatal("public view includes the key")
	}
	if err := store.SaveUserAIProvider(ctx, UserAIProvider{
		UserID: "u1", Provider: "openai", BaseURL: "https://api.example/v1", APIKey: "", Model: "gpt-4.1",
	}); err != nil {
		t.Fatal(err)
	}
	got, _, err = store.GetUserAIProvider(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.APIKey != "secret-key" {
		t.Fatalf("empty update dropped the key: %q", got.APIKey)
	}
}
