package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func TestAgentNotesStayWithTheSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tran.sqlite")
	store, err := morphdb.NewTranSQL(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	h := &Handlers{TranMySQL: store, jwtCfg: auth.LoadTokenConfig()}
	userA := seedPlatUser(t, store, "ada@example.com", "ada-notes", "secret", false)
	userB := seedPlatUser(t, store, "bea@example.com", "bea-notes", "secret", false)
	orphan := seedPlatUser(t, store, "orphan@example.com", "orphan-notes", "secret", false)
	ambiguous := seedPlatUser(t, store, "both@example.com", "both-notes", "secret", false)

	insertNoteUser := func(email string, deactivated int) int {
		t.Helper()
		res, err := store.DB.Exec(`INSERT INTO "User" (LastName, Email, Deactivated) VALUES ('N', ?, ?)`, email, deactivated)
		if err != nil {
			t.Fatal(err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		return int(id)
	}
	insertNote := func(userID int, itemType, title, body string, completed int) int {
		t.Helper()
		res, err := store.DB.Exec(`INSERT INTO user_note_todo (UserID, ItemType, Title, Body, Completed, CreatedOn)
			VALUES (?, ?, ?, ?, ?, '2026-10-01 12:00:00')`, userID, itemType, title, body, completed)
		if err != nil {
			t.Fatal(err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		return int(id)
	}

	demoID := insertNoteUser("demo@example.com", 0)
	adaID := insertNoteUser(" Ada@example.com ", 0)
	beaID := insertNoteUser("bea@example.com", 0)
	insertNoteUser("both@example.com", 0)
	insertNoteUser("both@example.com", 0)
	insertNoteUser("ada@example.com", 1)

	demoNote := insertNote(demoID, "note", "[morph-mcp] User one secret", "source: morph-mcp\n\ndemo", 0)
	adaNote := insertNote(adaID, "note", "[morph-mcp] Shift report", "source: morph-mcp\n\ndock 4 is clear", 0)
	adaDone := insertNote(adaID, "note", "[morph-mcp] Closed", "source: morph-mcp\n\ndone", 1)
	insertNote(adaID, "note", "[morph-mcp] Kept title", "the body was edited", 0)
	insertNote(adaID, "note", "Renamed by hand", "source: morph-mcp\n\nstill the agent", 0)
	insertNote(adaID, "note", "Groceries", "milk", 0)
	insertNote(adaID, "todo", "[morph-mcp] Not a note", "source: morph-mcp\n\ntodo", 0)
	insertNote(adaID, "note", "notes [morph-mcp] later", "hello source: morph-mcp", 0)
	beaNote := insertNote(beaID, "note", "[morph-mcp] Bea secret", "source: morph-mcp\n\nprivate", 0)
	_ = demoNote

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h.AuthzMiddleware())
	r.GET("/api/tran/agent-notes", h.ListAgentNotes)
	r.GET("/api/tran/agent-notes/:id", h.GetAgentNote)

	call := func(method, raw string, user *morphdb.PlatUser, headerUser string) (int, string) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, raw, nil)
		if headerUser != "" {
			req.Header.Set("X-User-ID", headerUser)
		}
		if user != nil {
			setBearer(t, h, req, user)
		}
		r.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}

	if code, body := call(http.MethodGet, "/api/tran/agent-notes", nil, ""); code != http.StatusUnauthorized || strings.Contains(body, "User one secret") || strings.Contains(body, "Shift report") || strings.Contains(body, "Bea secret") {
		t.Fatalf("logged-out list: %d %s", code, body)
	}
	if code, body := call(http.MethodGet, "/api/tran/agent-notes/"+itoa(adaNote), nil, userA.ID); code != http.StatusUnauthorized || strings.Contains(body, "dock 4") || strings.Contains(body, "Shift report") {
		t.Fatalf("header-only get: %d %s", code, body)
	}

	code, body := call(http.MethodGet, "/api/tran/agent-notes?user_id="+itoa(beaID), userA, userB.ID)
	if code != http.StatusOK || !strings.Contains(body, "Shift report") || !strings.Contains(body, "Kept title") || !strings.Contains(body, "Renamed by hand") || !strings.Contains(body, "Closed") {
		t.Fatalf("ada list: %d %s", code, body)
	}
	if strings.Contains(body, "Bea secret") || strings.Contains(body, "User one secret") || strings.Contains(body, "Groceries") || strings.Contains(body, "Not a note") || strings.Contains(body, "later") {
		t.Fatalf("ada list leaked: %s", body)
	}
	var listed []map[string]any
	if err := json.Unmarshal([]byte(body), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 4 {
		t.Fatalf("ada count %d: %s", len(listed), body)
	}
	seenDone := false
	for _, row := range listed {
		if row["title"] == "[morph-mcp] Closed" {
			seenDone = row["completed"] == true
		}
	}
	if !seenDone {
		t.Fatalf("done status: %s", body)
	}

	if code, body := call(http.MethodGet, "/api/tran/agent-notes/"+itoa(adaNote), userA, ""); code != http.StatusOK || !strings.Contains(body, "dock 4 is clear") {
		t.Fatalf("ada get: %d %s", code, body)
	}
	if code, body := call(http.MethodGet, "/api/tran/agent-notes/"+itoa(beaNote), userA, ""); code != http.StatusNotFound || strings.Contains(body, "Bea secret") || strings.Contains(body, "private") {
		t.Fatalf("cross-user get: %d %s", code, body)
	}
	if code, body := call(http.MethodGet, "/api/tran/agent-notes/"+itoa(adaDone), userB, ""); code != http.StatusNotFound || strings.Contains(body, "Closed") {
		t.Fatalf("bea get ada: %d %s", code, body)
	}
	if code, body := call(http.MethodGet, "/api/tran/agent-notes?user_id="+itoa(adaID), userB, userA.ID); code != http.StatusOK || strings.Contains(body, "Shift report") || strings.Contains(body, "dock 4") || !strings.Contains(body, "Bea secret") {
		t.Fatalf("user_id override: %d %s", code, body)
	}
	if code, body := call(http.MethodGet, "/api/tran/agent-notes", orphan, ""); code != http.StatusConflict || strings.Contains(body, "User one secret") || strings.Contains(body, "Shift report") {
		t.Fatalf("no tran user: %d %s", code, body)
	}
	if code, body := call(http.MethodGet, "/api/tran/agent-notes/"+itoa(demoNote), ambiguous, ""); code != http.StatusConflict || strings.Contains(body, "User one secret") {
		t.Fatalf("ambiguous email: %d %s", code, body)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
