package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func openLegacyCaseTaskDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	stmts := []string{
		`CREATE TABLE CaseTask (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NULL,
			start_at TEXT NULL,
			end_at TEXT NULL,
			location TEXT NULL,
			assignee_type TEXT NOT NULL,
			assignee_id INTEGER NOT NULL,
			created_on TEXT,
			last_updated TEXT
		)`,
		`CREATE TABLE CaseTaskAssignee (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			case_task_id INTEGER NOT NULL,
			assignee_kind TEXT NOT NULL,
			assignee_id INTEGER NOT NULL,
			created_on TEXT
		)`,
	}
	for _, s := range stmts {
		if _, err := sqlDB.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return sqlDB
}

func caseTaskCreateRouter(sqlDB *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	r := gin.New()
	r.POST("/api/tran/case-tasks", h.CreateCaseTask)
	return r
}

func postCreateCaseTask(r http.Handler, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/case-tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestCreateCaseTaskUnassignedAgainstNotNullAssigneeType(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	r := caseTaskCreateRouter(sqlDB)

	w := postCreateCaseTask(r, `{"title":"Site visit","start_at":"2026-09-01T00:00","end_at":"2026-09-01T23:59","assignee_type":null,"assignee_id":null}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if strings.Contains(strings.ToLower(w.Body.String()), "constraint") {
		t.Fatalf("must not return constraint error: %s", w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	id, _ := out["id"].(float64)
	if id <= 0 {
		t.Fatalf("expected id, got %s", w.Body.String())
	}
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM CaseTask WHERE ID = ?`, int(id)).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 row, got %d", n)
	}
}

func TestCreateCaseTaskRejectsMissingStartAt(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	r := caseTaskCreateRouter(sqlDB)

	w := postCreateCaseTask(r, `{"title":"Site visit","end_at":"2026-09-01T23:59"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "start") {
		t.Fatalf("expected start date validation, got %s", w.Body.String())
	}
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM CaseTask`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected no row, got %d", n)
	}
}

func TestCreateCaseTaskRejectsMissingEndAt(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	r := caseTaskCreateRouter(sqlDB)

	w := postCreateCaseTask(r, `{"title":"Site visit","start_at":"2026-09-01T00:00"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "end") {
		t.Fatalf("expected end date validation, got %s", w.Body.String())
	}
}

func TestCreateCaseTaskAcceptsStartAndEnd(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	r := caseTaskCreateRouter(sqlDB)

	w := postCreateCaseTask(r, `{"title":"Site visit","start_at":"2026-09-01T00:00","end_at":"2026-09-01T23:59"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	id, _ := out["id"].(float64)
	if id <= 0 {
		t.Fatalf("expected id, got %s", w.Body.String())
	}
	var startAt, endAt sql.NullString
	if err := sqlDB.QueryRow(`SELECT start_at, end_at FROM CaseTask WHERE ID = ?`, int(id)).Scan(&startAt, &endAt); err != nil {
		t.Fatal(err)
	}
	if !startAt.Valid || !strings.Contains(startAt.String, "2026-09-01") {
		t.Fatalf("start_at=%v", startAt)
	}
	if !endAt.Valid || !strings.Contains(endAt.String, "2026-09-01") {
		t.Fatalf("end_at=%v", endAt)
	}
}

func TestUpdateCaseTaskRejectsMissingDates(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	r := gin.New()
	r.POST("/api/tran/case-tasks", h.CreateCaseTask)
	r.PUT("/api/tran/case-tasks/:id", h.UpdateCaseTask)

	created := postCreateCaseTask(r, `{"title":"Site visit","start_at":"2026-09-01T00:00","end_at":"2026-09-01T23:59"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("create status %d body %s", created.Code, created.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	id := int(out["id"].(float64))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tran/case-tasks/%d", id), strings.NewReader(`{"title":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestCreateCaseTaskRejectsBlankTitle(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	r := caseTaskCreateRouter(sqlDB)

	w := postCreateCaseTask(r, `{"title":"   "}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "title") {
		t.Fatalf("expected title validation, got %s", w.Body.String())
	}
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM CaseTask`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected no row, got %d", n)
	}
}
