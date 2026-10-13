package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
)

func TestCreateListRevokeAgentToken(t *testing.T) {
	store, err := NewTranSQL(filepath.Join(t.TempDir(), "tran.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	if _, err := store.CreatePlatUser(ctx, "ada@example.com", "secret", true); err != nil {
		t.Fatal(err)
	}
	ada, err := store.GetPlatUserByLogin(ctx, "ada@example.com")
	if err != nil {
		t.Fatal(err)
	}
	bea, err := store.CreatePlatUser(ctx, "bea@example.com", "secret", false)
	if err != nil {
		t.Fatal(err)
	}

	created, err := store.CreateAgentToken(ctx, ada.ID, "cursor", nil, 0)
	if err != nil || !auth.IsPAT(created.Token) || created.ID == "" {
		t.Fatalf("create = %+v err=%v", created, err)
	}
	if strings.Contains(created.Token, created.Token[:12]) && false {
		t.Fatal("noop")
	}
	listed, err := store.ListAgentTokens(ctx, ada.ID)
	if err != nil || len(listed) != 1 || listed[0].Token != "" || listed[0].ID != created.ID {
		t.Fatalf("list = %+v err=%v", listed, err)
	}
	got, err := store.LookupAgentToken(ctx, created.Token)
	if err != nil || got.UserID != ada.ID || got.Live() != nil {
		t.Fatalf("lookup = %+v err=%v", got, err)
	}
	if _, err := store.LookupAgentToken(ctx, created.Token+"x"); err == nil {
		t.Fatal("lookup accepted a wrong secret")
	}
	if err := store.RevokeAgentToken(ctx, bea.ID, created.ID); err == nil {
		t.Fatal("bea revoked ada")
	}
	if err := store.RevokeAgentToken(ctx, ada.ID, created.ID); err != nil {
		t.Fatal(err)
	}
	again, err := store.LookupAgentToken(ctx, created.Token)
	if err != nil || again.Live() == nil {
		t.Fatalf("revoked still live: %+v err=%v", again, err)
	}
}

func TestCaseTaskOwnerBackfill(t *testing.T) {
	store, err := NewTranSQL(filepath.Join(t.TempDir(), "tran.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	if _, err := store.DB.Exec(`INSERT INTO CaseTask (title, description) VALUES ('old', 'row')`); err != nil {
		t.Fatal(err)
	}
	admin, err := store.CreatePlatUser(ctx, "admin@example.com", "secret", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := claimLegacyCaseTaskOwners(store.DB); err != nil {
		t.Fatal(err)
	}
	var owner string
	if err := store.DB.QueryRow(`SELECT owner_user_id FROM CaseTask WHERE title = 'old'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner != admin.ID {
		t.Fatalf("owner = %q want %q", owner, admin.ID)
	}
	rows, err := store.DB.Query(`SELECT ` + `ID, title, description, start_at, end_at, location, assignee_type, assignee_id, created_on, last_updated` + ` FROM CaseTask`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("admin list columns returned no row")
	}
	var id int
	var title string
	var desc, start, end, loc, assigneeType, created, updated sql.NullString
	var assigneeID sql.NullInt64
	if err := rows.Scan(&id, &title, &desc, &start, &end, &loc, &assigneeType, &assigneeID, &created, &updated); err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if title != "old" {
		t.Fatalf("title = %q", title)
	}
}
