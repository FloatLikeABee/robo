package mcp_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/config"
	morphdb "idongivaflyinfa/db"
	"idongivaflyinfa/mcp"
)

func TestMyTasksAreScopedReadOnlyAndVisibleBesideWAL(t *testing.T) {
	path, writer := seedTwoUsers(t)
	ctx := context.Background()

	ro, err := mcp.OpenReadOnly(path)
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	t.Cleanup(func() { _ = ro.Close() })

	listed, err := mcp.ListMyTasks(ctx, ro, "ada-id", mcp.TaskFilter{})
	if err != nil {
		t.Fatalf("ListMyTasks: %v", err)
	}
	if listed.Limit != 50 {
		t.Fatalf("default limit = %d", listed.Limit)
	}
	if len(listed.Tasks) != 2 {
		t.Fatalf("tasks = %+v", listed.Tasks)
	}
	blob := taskBlob(listed.Tasks)
	if !strings.Contains(blob, "Ada open") || !strings.Contains(blob, "Ada done") {
		t.Fatalf("missing caller rows: %s", blob)
	}
	if strings.Contains(blob, "Bea private") || strings.Contains(blob, "Shared board case") {
		t.Fatalf("list leaked another row: %s", blob)
	}

	openOnly, err := mcp.ListMyTasks(ctx, ro, "ada-id", mcp.TaskFilter{Status: "open"})
	if err != nil {
		t.Fatal(err)
	}
	if len(openOnly.Tasks) != 1 || openOnly.Tasks[0].Title != "Ada open" || openOnly.Tasks[0].Status != "open" {
		t.Fatalf("open filter = %+v", openOnly.Tasks)
	}
	if openOnly.Tasks[0].Body != "walk the lot" || openOnly.Tasks[0].ItemType != "todo" {
		t.Fatalf("open task = %+v", openOnly.Tasks[0])
	}
	if openOnly.Tasks[0].DeadlineAt == "" || openOnly.Tasks[0].CreatedOn == "" {
		t.Fatalf("dates = %+v", openOnly.Tasks[0])
	}

	capped, err := mcp.ListMyTasks(ctx, ro, "ada-id", mcp.TaskFilter{Limit: 500})
	if err != nil {
		t.Fatal(err)
	}
	if capped.Limit != 100 || len(capped.Tasks) > 100 {
		t.Fatalf("cap = %+v len %d", capped.Limit, len(capped.Tasks))
	}

	mine, err := mcp.GetMyTask(ctx, ro, "ada-id", openOnly.Tasks[0].ID)
	if err != nil {
		t.Fatalf("GetMyTask: %v", err)
	}
	if mine.Title != "Ada open" || mine.Body != "walk the lot" || mine.Status != "open" {
		t.Fatalf("get = %+v", mine)
	}

	var beaID int
	if err := writer.QueryRow(`SELECT ID FROM user_note_todo WHERE Title = 'Bea private'`).Scan(&beaID); err != nil {
		t.Fatal(err)
	}
	_, err = mcp.GetMyTask(ctx, ro, "ada-id", beaID)
	if !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("foreign get = %v", err)
	}
	if strings.Contains(err.Error(), "Bea private") || strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("foreign error = %q", err)
	}

	if _, err := ro.Exec(`INSERT INTO user_note_todo (UserID, ItemType, Title) VALUES (1, 'todo', 'nope')`); err == nil {
		t.Fatal("read-only connection accepted a write")
	}

	if _, err := writer.Exec(`INSERT INTO user_note_todo (UserID, ItemType, Title, Body, Completed) VALUES (1, 'note', 'Ada later', 'still mine', 0)`); err != nil {
		t.Fatalf("writer commit: %v", err)
	}
	again, err := mcp.ListMyTasks(ctx, ro, "ada-id", mcp.TaskFilter{Type: "note"})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Tasks) != 1 || again.Tasks[0].Title != "Ada later" {
		t.Fatalf("note filter after WAL commit = %+v", again.Tasks)
	}

	bea, err := mcp.ListMyTasks(ctx, ro, "bea-id", mcp.TaskFilter{})
	if err != nil {
		t.Fatal(err)
	}
	beaBlob := taskBlob(bea.Tasks)
	if !strings.Contains(beaBlob, "Bea private") || strings.Contains(beaBlob, "Ada open") || strings.Contains(beaBlob, "Ada later") {
		t.Fatalf("bea list = %s", beaBlob)
	}

	injected, err := mcp.ListMyTasks(ctx, ro, "ada-id' OR 1=1 --", mcp.TaskFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(injected.Tasks) != 0 {
		t.Fatalf("injection returned rows: %+v", injected.Tasks)
	}
	if _, err := mcp.GetMyTask(ctx, ro, "ada-id' OR 1=1 --", openOnly.Tasks[0].ID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("injection get = %v", err)
	}

	if _, err := writer.Exec(`UPDATE "User" SET Deactivated = 1 WHERE Email = 'ada@example.com'`); err != nil {
		t.Fatal(err)
	}
	gone, err := mcp.ListMyTasks(ctx, ro, "ada-id", mcp.TaskFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(gone.Tasks) != 0 {
		t.Fatalf("deactivated user still listed: %+v", gone.Tasks)
	}
}

func TestOpenReadOnlyDoesNotCreateMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.sqlite")
	_, err := mcp.OpenReadOnly(missing)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, statErr := os.Stat(missing); !os.IsNotExist(statErr) {
		t.Fatalf("stat = %v", statErr)
	}
}

func TestRejectJWTSecret(t *testing.T) {
	for _, secret := range []string{"", "   ", config.DefaultJWTSecret, strings.ToUpper(config.DefaultJWTSecret)} {
		err := mcp.RejectJWTSecret(secret)
		if err == nil {
			t.Fatalf("accepted %q", secret)
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Fatalf("error = %q", err)
		}
		if strings.Contains(err.Error(), config.DefaultJWTSecret) {
			t.Fatalf("error echoed the secret: %q", err)
		}
	}
	if err := mcp.RejectJWTSecret("stdio-test-secret"); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmPlatUserRejectsDeletedSubject(t *testing.T) {
	path, writer := seedTwoUsers(t)
	ctx := context.Background()
	if err := mcp.ConfirmPlatUser(ctx, writer, "ada-id"); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Exec(`DELETE FROM plat_users WHERE id = 'ada-id'`); err != nil {
		t.Fatal(err)
	}
	err := mcp.ConfirmPlatUser(ctx, writer, "ada-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "ada-id") {
		t.Fatalf("error included the subject: %q", err)
	}
	ro, err := mcp.OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ro.Close() })
	if err := mcp.ConfirmPlatUser(ctx, ro, "bea-id"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckStartupValidatesJWTBeforeServing(t *testing.T) {
	path, _ := seedTwoUsers(t)
	ctx := context.Background()
	secret := "0123456789abcdef0123456789abcdef"
	t.Setenv("TRAN_SQLITE_PATH", path)
	t.Setenv("MORPH_ENV", "production")
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	t.Setenv("JWT_SECRET", "stdio-test-secret")
	t.Setenv(mcp.TokenEnv, "not-a-token")
	if _, _, err := mcp.CheckStartup(ctx); err == nil {
		t.Fatal("production accepted a short JWT_SECRET")
	} else if strings.Contains(err.Error(), "stdio-test-secret") {
		t.Fatalf("error echoed the secret: %q", err)
	}

	t.Setenv("JWT_SECRET", secret)
	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	if _, _, err := mcp.CheckStartup(ctx); err == nil {
		t.Fatal("production accepted JWT_EXPIRY_HOURS=876000")
	}

	t.Setenv("JWT_EXPIRY_HOURS", "24")
	longLived, err := auth.EncodeToken(auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 876000}, "ada-id", "ada@example.com", "ada", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(mcp.TokenEnv, longLived)
	if _, _, err := mcp.CheckStartup(ctx); err == nil {
		t.Fatal("production accepted a token whose lifetime is 876000 hours")
	} else if strings.Contains(err.Error(), longLived) {
		t.Fatal("error included the token")
	}

	other, err := auth.EncodeToken(auth.TokenConfig{Secret: []byte("other-secret-0123456789abcdef"), ExpiryHours: 24}, "ada-id", "ada@example.com", "ada", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(mcp.TokenEnv, other)
	if _, _, err := mcp.CheckStartup(ctx); err == nil {
		t.Fatal("accepted a token signed with a different secret")
	} else if strings.Contains(err.Error(), other) || strings.Contains(err.Error(), "ada-id") {
		t.Fatalf("error leaked token or subject: %q", err)
	}

	good, err := auth.EncodeToken(auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 24}, "ada-id", "ada@example.com", "ada", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(mcp.TokenEnv, good)
	id, db, err := mcp.CheckStartup(ctx)
	if err != nil {
		t.Fatalf("valid production token: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if id.UserID != "ada-id" {
		t.Fatalf("identity = %+v", id)
	}
	listed, err := mcp.ListMyTasks(ctx, db, id.UserID, mcp.TaskFilter{Status: "open"})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 1 || listed.Tasks[0].Title != "Ada open" || strings.Contains(taskBlob(listed.Tasks), "Bea private") {
		t.Fatalf("scoped list = %+v", listed.Tasks)
	}
}

func TestRecheckTokenRejectsExpiredWithoutEcho(t *testing.T) {
	secret := []byte("stdio-test-secret")
	expiredCfg := auth.TokenConfig{Secret: secret, ExpiryHours: -1}
	tok, err := auth.EncodeToken(expiredCfg, "ada-id", "ada@example.com", "ada", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	err = mcp.RecheckToken(tok, auth.TokenConfig{Secret: secret, ExpiryHours: 24})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), tok) {
		t.Fatal("error included the token")
	}
}

func seedTwoUsers(t *testing.T) (string, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tran.sqlite")
	store, err := morphdb.NewTranSQL(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	db := store.DB
	now := "2026-09-24T00:00:00Z"
	for _, row := range []struct{ id, email, username string }{
		{"ada-id", "ada@example.com", "ada"},
		{"bea-id", "bea@example.com", "bea"},
	} {
		_, err = db.Exec(`INSERT INTO plat_users (
			id, email, username, is_verified, roles, permissions, default_channel_id, created_at, updated_at
		) VALUES (?, ?, ?, 1, '["employee"]', '[]', 'ch', ?, ?)`, row.id, row.email, row.username, now, now)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec(`INSERT INTO "User" (LoginID, FirstName, LastName, Email, Deactivated) VALUES
		('ada', 'Ada', 'Lovelace', 'ada@example.com', 0),
		('bea', 'Bea', 'Knuth', 'bea@example.com', 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO user_note_todo (UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated)
		VALUES
		(1, 'todo', 'Ada open', 'walk the lot', 0, '2026-09-01 09:00:00', '2026-08-01 10:00:00', '2026-08-02 10:00:00'),
		(1, 'todo', 'Ada done', 'sent the note', 1, NULL, '2026-08-03 10:00:00', '2026-08-04 10:00:00'),
		(2, 'todo', 'Bea private', 'do not leak', 0, NULL, '2026-08-05 10:00:00', '2026-08-05 10:00:00')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO CaseTask (title, description) VALUES ('Shared board case', 'not mine')`); err != nil {
		t.Fatal(err)
	}
	return path, db
}

func taskBlob(tasks []mcp.Task) string {
	var b strings.Builder
	for _, task := range tasks {
		b.WriteString(task.Title)
		b.WriteByte('\n')
		b.WriteString(task.Body)
		b.WriteByte('\n')
	}
	return b.String()
}
