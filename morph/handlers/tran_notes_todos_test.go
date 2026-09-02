package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func openUserNoteTodoDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", "file:notes-todo-"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.Exec(`CREATE TABLE user_note_todo (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID INTEGER NOT NULL,
		ItemType TEXT NOT NULL,
		Title TEXT NULL,
		Body TEXT NULL,
		Completed INTEGER NOT NULL DEFAULT 0,
		DeadlineAt TEXT NULL,
		CreatedOn TEXT DEFAULT CURRENT_TIMESTAMP,
		LastUpdated TEXT DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal(err)
	}
	return sqlDB
}

func TestScanUserNoteTodoAcceptsSQLiteTextTimestamps(t *testing.T) {
	sqlDB := openUserNoteTodoDB(t)
	_, err := sqlDB.Exec(`INSERT INTO user_note_todo
		(ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated)
		VALUES (1, 1, 'todo', 'Site', 'Walk the lot', 0, '2026-08-24 09:15:00', '2026-08-23 10:00:00', '2026-08-23 11:30:00')`)
	if err != nil {
		t.Fatal(err)
	}
	row := sqlDB.QueryRow(`SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated FROM user_note_todo WHERE ID = 1`)
	got, err := scanUserNoteTodo(row)
	if err != nil {
		t.Fatalf("scanUserNoteTodo: %v", err)
	}
	if got.Title == nil || *got.Title != "Site" {
		t.Fatalf("title=%v", got.Title)
	}
	if got.DeadlineAt == nil || got.DeadlineAt.Day() != 24 || got.DeadlineAt.Hour() != 9 {
		t.Fatalf("deadline_at not parsed: %v", got.DeadlineAt)
	}
	if got.CreatedOn == nil || got.CreatedOn.Year() != 2026 || got.CreatedOn.Month() != 8 || got.CreatedOn.Day() != 23 {
		t.Fatalf("created_on not parsed: %v", got.CreatedOn)
	}
	if got.LastUpdated == nil || got.LastUpdated.Hour() != 11 || got.LastUpdated.Minute() != 30 {
		t.Fatalf("last_updated not parsed: %v", got.LastUpdated)
	}
}

func TestCreateUserNoteTodoThenListAcceptsSQLiteText(t *testing.T) {
	sqlDB := openUserNoteTodoDB(t)
	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	r := gin.New()
	r.POST("/api/tran/notes-todos", h.CreateUserNoteTodo)
	r.GET("/api/tran/notes-todos", h.ListUserNotesTodos)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/notes-todos", strings.NewReader(`{"item_type":"todo","title":"Call depot","body":"Ask for keys"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status %d body %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tran/notes-todos", nil)
	req.Header.Set("X-User-ID", "1")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d body %s", w.Code, w.Body.String())
	}
	if strings.Contains(strings.ToLower(w.Body.String()), "unsupported scan") {
		t.Fatalf("list scan failed: %s", w.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d items: %s", len(list), w.Body.String())
	}
	if list[0]["title"] != "Call depot" {
		t.Fatalf("title=%v", list[0]["title"])
	}
}
