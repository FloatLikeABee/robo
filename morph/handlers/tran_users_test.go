package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"

	_ "modernc.org/sqlite"
)

func openTranUserDB(t *testing.T) *morphdb.TranSQL {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "-")
	sqlDB, err := sql.Open("sqlite", "file:"+name+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if _, err := sqlDB.Exec(`CREATE TABLE "User" (
		UserID INTEGER PRIMARY KEY AUTOINCREMENT,
		LoginID TEXT NULL,
		FirstName TEXT NULL,
		LastName TEXT NOT NULL DEFAULT '',
		Email TEXT NULL,
		Phone TEXT NULL,
		Title TEXT NULL,
		Administrator INTEGER NOT NULL DEFAULT 0,
		Deactivated INTEGER NOT NULL DEFAULT 0,
		DeactivatedDate TEXT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	ts := &morphdb.TranSQL{DB: sqlDB}
	if err := ts.EnsurePlatUsersTable(context.Background()); err != nil {
		t.Fatal(err)
	}
	return ts
}

func tranUserRouter(h *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h.AuthzMiddleware())
	r.GET("/api/tran/users/me", h.GetTranUserMe)
	r.PUT("/api/tran/users/me", h.UpdateTranUserMe)
	r.GET("/api/tran/users", h.ListTranUsers)
	r.POST("/api/tran/users", h.CreateTranUser)
	r.PUT("/api/tran/users/:id", h.UpdateTranUser)
	r.DELETE("/api/tran/users/:id", h.DeleteTranUser)
	return r
}

func tranUserHandlers(t *testing.T) (*Handlers, *gin.Engine, *morphdb.TranSQL) {
	t.Helper()
	ts := openTranUserDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	return h, tranUserRouter(h), ts
}

func insertTranUser(t *testing.T, ts *morphdb.TranSQL, email, login, last string, admin, deactivated int) int {
	t.Helper()
	res, err := ts.DB.Exec(
		`INSERT INTO "User" (LoginID, LastName, Email, Administrator, Deactivated) VALUES (?, ?, ?, ?, ?)`,
		login, last, email, admin, deactivated)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return int(id)
}

func countTranUsers(t *testing.T, ts *morphdb.TranSQL) int {
	t.Helper()
	var n int
	if err := ts.DB.QueryRow(`SELECT COUNT(*) FROM "User"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func tranUserField(t *testing.T, ts *morphdb.TranSQL, id int, col string) string {
	t.Helper()
	var v sql.NullString
	q := `SELECT ` + col + ` FROM "User" WHERE UserID = ?`
	if err := ts.DB.QueryRow(q, id).Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v.String
}

func notesTranUserID(h *Handlers, email string) int {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/tran/notes-todos", nil)
	c.Set("auth_email", email)
	return h.tranUserIDFromContext(c)
}

func TestTranUserAdminRoutes(t *testing.T) {
	h, r, ts := tranUserHandlers(t)
	admin := seedPlatUser(t, ts, "admin-users@test.local", "admin-users", "secret", true)
	employee := seedPlatUser(t, ts, "employee-users@test.local", "employee-users", "secret", false)
	adminTok := bearerFor(t, h, admin)
	empTok := bearerFor(t, h, employee)
	ownID := insertTranUser(t, ts, employee.Email, employee.Email, "Employee", 1, 0)

	spoof := http.Header{}
	spoof.Set("X-User-Role", "admin")
	spoof.Set("X-User-ID", "1")

	t.Run("no session is 401", func(t *testing.T) {
		before := countTranUsers(t, ts)
		for _, tc := range []struct{ method, path, body string }{
			{http.MethodPost, "/api/tran/users", `{"last_name":"Nope"}`},
			{http.MethodPut, "/api/tran/users/" + strconv.Itoa(ownID), `{"last_name":"Nope"}`},
			{http.MethodDelete, "/api/tran/users/" + strconv.Itoa(ownID), ``},
		} {
			w := doAuthz(r, tc.method, tc.path, tc.body, "", spoof)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s %s: status %d, want 401, body %s", tc.method, tc.path, w.Code, w.Body.String())
			}
		}
		if got := countTranUsers(t, ts); got != before {
			t.Fatalf("user count %d, want %d", got, before)
		}
		if tranUserField(t, ts, ownID, "Administrator") != "1" {
			t.Fatal("anonymous request changed administrator")
		}
	})

	t.Run("non-admin is 403", func(t *testing.T) {
		before := countTranUsers(t, ts)
		for _, tc := range []struct{ method, path, body string }{
			{http.MethodPost, "/api/tran/users", `{"last_name":"Nope"}`},
			{http.MethodPut, "/api/tran/users/" + strconv.Itoa(ownID), `{"last_name":"Hacked","email":"victim@test.local"}`},
			{http.MethodDelete, "/api/tran/users/" + strconv.Itoa(ownID), ``},
		} {
			w := doAuthz(r, tc.method, tc.path, tc.body, empTok, spoof)
			if w.Code != http.StatusForbidden {
				t.Errorf("%s %s: status %d, want 403, body %s", tc.method, tc.path, w.Code, w.Body.String())
			}
		}
		if got := countTranUsers(t, ts); got != before {
			t.Fatalf("user count %d, want %d", got, before)
		}
		if tranUserField(t, ts, ownID, "Email") != employee.Email {
			t.Fatalf("email changed to %q", tranUserField(t, ts, ownID, "Email"))
		}
		if tranUserField(t, ts, ownID, "Deactivated") != "0" {
			t.Fatal("non-admin deactivated the row")
		}
	})

	t.Run("self-promotion is 403", func(t *testing.T) {
		w := doAuthz(r, http.MethodPut, "/api/tran/users/"+strconv.Itoa(ownID), `{"administrator":true}`, empTok, nil)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status %d, want 403, body %s", w.Code, w.Body.String())
		}
		if tranUserField(t, ts, ownID, "Administrator") != "1" {
			t.Fatal("administrator flag changed")
		}
		// The seeded flag is already 1 to prove Tran Administrator is not the platform role.
		// Reset and try promotion from false.
		if _, err := ts.DB.Exec(`UPDATE "User" SET Administrator = 0 WHERE UserID = ?`, ownID); err != nil {
			t.Fatal(err)
		}
		w = doAuthz(r, http.MethodPut, "/api/tran/users/"+strconv.Itoa(ownID), `{"administrator":true}`, empTok, spoof)
		if w.Code != http.StatusForbidden {
			t.Fatalf("promotion status %d, want 403, body %s", w.Code, w.Body.String())
		}
		if tranUserField(t, ts, ownID, "Administrator") != "0" {
			t.Fatal("self-promotion stuck")
		}
	})

	t.Run("stale admin token is 403", func(t *testing.T) {
		tok := bearerFor(t, h, admin)
		off := false
		if _, err := ts.UpdatePlatUser(context.Background(), admin.ID, "", "", &off); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			on := true
			_, _ = ts.UpdatePlatUser(context.Background(), admin.ID, "", "", &on)
		})
		before := countTranUsers(t, ts)
		w := doAuthz(r, http.MethodPost, "/api/tran/users", `{"last_name":"Stale"}`, tok, nil)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status %d, want 403, body %s", w.Code, w.Body.String())
		}
		if got := countTranUsers(t, ts); got != before {
			t.Fatalf("user count %d, want %d", got, before)
		}
	})

	t.Run("admin succeeds", func(t *testing.T) {
		w := doAuthz(r, http.MethodPost, "/api/tran/users", `{"last_name":"Ada","email":"ada@test.local"}`, adminTok, nil)
		if w.Code != http.StatusCreated {
			t.Fatalf("create status %d, want 201, body %s", w.Code, w.Body.String())
		}
		var created struct {
			User struct {
				ID int `json:"id"`
			} `json:"user"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
			t.Fatal(err)
		}
		if created.User.ID <= 0 {
			t.Fatalf("create body %s", w.Body.String())
		}
		w = doAuthz(r, http.MethodPut, "/api/tran/users/"+strconv.Itoa(created.User.ID), `{"email":"ada-renamed@test.local","last_name":"Lovelace"}`, adminTok, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("update status %d, want 200, body %s", w.Code, w.Body.String())
		}
		if got := tranUserField(t, ts, created.User.ID, "Email"); got != "ada-renamed@test.local" {
			t.Fatalf("admin email update got %q", got)
		}
		w = doAuthz(r, http.MethodDelete, "/api/tran/users/"+strconv.Itoa(created.User.ID), ``, adminTok, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("delete status %d, want 200, body %s", w.Code, w.Body.String())
		}
		if tranUserField(t, ts, created.User.ID, "Deactivated") != "1" {
			t.Fatal("admin delete did not deactivate")
		}
	})

	t.Run("signed-in list is not 403", func(t *testing.T) {
		w := doAuthz(r, http.MethodGet, "/api/tran/users", ``, empTok, nil)
		if w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized {
			t.Fatalf("list status %d, body %s", w.Code, w.Body.String())
		}
	})
}

func TestTranUserMeDoesNotChangeEmailOrRole(t *testing.T) {
	h, r, ts := tranUserHandlers(t)
	employee := seedPlatUser(t, ts, "me-user@test.local", "me-user", "secret", false)
	id := insertTranUser(t, ts, employee.Email, employee.Email, "Old", 0, 0)
	tok := bearerFor(t, h, employee)

	w := doAuthz(r, http.MethodPut, "/api/tran/users/me", `{"email":"other@test.local","last_name":"New"}`, tok, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200, body %s", w.Code, w.Body.String())
	}
	if got := tranUserField(t, ts, id, "Email"); got != employee.Email {
		t.Fatalf("email %q, want %q", got, employee.Email)
	}
	if got := tranUserField(t, ts, id, "LastName"); got != "New" {
		t.Fatalf("last name %q, want New", got)
	}

	w = doAuthz(r, http.MethodPut, "/api/tran/users/me", `{"email":"other@test.local"}`, tok, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("email-only status %d, want 400, body %s", w.Code, w.Body.String())
	}
	if got := tranUserField(t, ts, id, "Email"); got != employee.Email {
		t.Fatalf("email-only wrote %q", got)
	}

	w = doAuthz(r, http.MethodPut, "/api/tran/users/me", `{"administrator":true,"deactivated":true,"last_name":"Newer"}`, tok, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("role body status %d, body %s", w.Code, w.Body.String())
	}
	if tranUserField(t, ts, id, "Administrator") != "0" || tranUserField(t, ts, id, "Deactivated") != "0" {
		t.Fatal("profile save changed administrator or deactivated")
	}
}

func TestTranUserTakeoverSequenceFails(t *testing.T) {
	h, r, ts := tranUserHandlers(t)
	a := seedPlatUser(t, ts, "attacker@test.local", "attacker", "secret", false)
	b := seedPlatUser(t, ts, "victim@test.local", "victim", "secret", false)
	idA := insertTranUser(t, ts, a.Email, a.Email, "Attacker", 0, 0)
	idB := insertTranUser(t, ts, b.Email, b.Email, "Victim", 0, 0)
	tok := bearerFor(t, h, a)

	w := doAuthz(r, http.MethodPut, "/api/tran/users/"+strconv.Itoa(idB), `{"email":"`+a.Email+`"}`, tok, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("rewrite status %d, want 403, body %s", w.Code, w.Body.String())
	}
	w = doAuthz(r, http.MethodDelete, "/api/tran/users/"+strconv.Itoa(idA), ``, tok, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("deactivate status %d, want 403, body %s", w.Code, w.Body.String())
	}
	if got := tranUserField(t, ts, idB, "Email"); got != b.Email {
		t.Fatalf("victim email %q", got)
	}
	if tranUserField(t, ts, idA, "Deactivated") != "0" {
		t.Fatal("attacker row was deactivated")
	}

	w = doAuthz(r, http.MethodGet, "/api/tran/users/me", ``, tok, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("me status %d, body %s", w.Code, w.Body.String())
	}
	var me struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.ID != idA || me.ID == idB {
		t.Fatalf("resolved profile id %d, want %d", me.ID, idA)
	}
	if got := notesTranUserID(h, a.Email); got != idA {
		t.Fatalf("notes user id %d, want %d", got, idA)
	}
}

func TestAmbiguousTranEmailFailsClosed(t *testing.T) {
	h, r, ts := tranUserHandlers(t)
	user := seedPlatUser(t, ts, "shared@test.local", "shared", "secret", false)
	id1 := insertTranUser(t, ts, user.Email, "one", "One", 0, 0)
	id2 := insertTranUser(t, ts, user.Email, "two", "Two", 0, 0)
	tok := bearerFor(t, h, user)
	before := countTranUsers(t, ts)

	w := doAuthz(r, http.MethodGet, "/api/tran/users/me", ``, tok, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409, body %s", w.Code, w.Body.String())
	}
	if got := countTranUsers(t, ts); got != before {
		t.Fatalf("user count %d, want %d", got, before)
	}
	got := notesTranUserID(h, user.Email)
	if got == id1 || got == id2 {
		t.Fatalf("notes resolver picked %d", got)
	}

	w = doAuthz(r, http.MethodPut, "/api/tran/users/me", `{"last_name":"Nope"}`, tok, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("put status %d, want 409, body %s", w.Code, w.Body.String())
	}
	if tranUserField(t, ts, id1, "LastName") == "Nope" || tranUserField(t, ts, id2, "LastName") == "Nope" {
		t.Fatal("ambiguous put wrote a last name")
	}
}

func TestAmbiguousTranLoginIDFailsClosed(t *testing.T) {
	h, r, ts := tranUserHandlers(t)
	user := seedPlatUser(t, ts, "login-shared@test.local", "login-shared", "secret", false)
	insertTranUser(t, ts, "other-a@test.local", user.Email, "A", 0, 0)
	insertTranUser(t, ts, "other-b@test.local", user.Email, "B", 0, 0)
	before := countTranUsers(t, ts)
	w := doAuthz(r, http.MethodGet, "/api/tran/users/me", ``, bearerFor(t, h, user), nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409, body %s", w.Code, w.Body.String())
	}
	if got := countTranUsers(t, ts); got != before {
		t.Fatalf("user count %d, want %d", got, before)
	}
}

func TestTranEmailResolutionKeepsSingleAndMissing(t *testing.T) {
	h, r, ts := tranUserHandlers(t)
	user := seedPlatUser(t, ts, "only@test.local", "only", "secret", false)
	active := insertTranUser(t, ts, user.Email, user.Email, "Only", 0, 0)
	insertTranUser(t, ts, user.Email, "old-login", "Gone", 0, 1)
	tok := bearerFor(t, h, user)

	w := doAuthz(r, http.MethodGet, "/api/tran/users/me", ``, tok, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("single status %d, body %s", w.Code, w.Body.String())
	}
	var me struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.ID != active {
		t.Fatalf("id %d, want active %d", me.ID, active)
	}

	fresh := seedPlatUser(t, ts, "brand-new@test.local", "brand-new", "secret", false)
	before := countTranUsers(t, ts)
	w = doAuthz(r, http.MethodGet, "/api/tran/users/me", ``, bearerFor(t, h, fresh), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("create status %d, body %s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if countTranUsers(t, ts) != before+1 || me.ID <= 0 {
		t.Fatalf("count %d id %d", countTranUsers(t, ts), me.ID)
	}
	w = doAuthz(r, http.MethodGet, "/api/tran/users/me", ``, bearerFor(t, h, fresh), nil)
	var again struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &again); err != nil {
		t.Fatal(err)
	}
	if again.ID != me.ID {
		t.Fatalf("second id %d, want %d", again.ID, me.ID)
	}
}
