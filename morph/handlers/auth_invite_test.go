package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"idongivaflyinfa/auth"
	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"

	_ "modernc.org/sqlite"
)

func openAuthHarnessSQL(t *testing.T) *morphdb.TranSQL {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", "file:auth-"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	ts := &morphdb.TranSQL{DB: sqlDB}
	ctx := context.Background()
	if err := ts.EnsurePlatUsersTable(ctx); err != nil {
		t.Fatal(err)
	}
	if err := ts.EnsurePlatInviteCodesTable(ctx); err != nil {
		t.Fatal(err)
	}
	return ts
}

func seedPlatUser(t *testing.T, ts *morphdb.TranSQL, email, username, password string, admin bool) *morphdb.PlatUser {
	t.Helper()
	u, err := ts.CreatePlatUser(context.Background(), email, password, admin)
	if err != nil {
		t.Fatal(err)
	}
	if username != "" && username != email {
		updated, err := ts.UpdatePlatUserCredentials(context.Background(), u.ID, username, "", "")
		if err != nil {
			t.Fatal(err)
		}
		return updated
	}
	return u
}

func bearerFor(t *testing.T, h *Handlers, u *morphdb.PlatUser) string {
	t.Helper()
	token, err := auth.EncodeToken(h.jwtCfg, u.ID, u.Email, u.Username, u.Roles, u.DefaultChannelID)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRedeemInviteCodeCreatesUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := openAuthHarnessSQL(t)
	admin := seedPlatUser(t, ts, "admin@test.local", "admin", "secret", true)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	inv, err := ts.CreateInviteCode(context.Background(), admin.ID)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.POST("/api/invite/redeem", h.RedeemInviteCode)

	body, _ := json.Marshal(map[string]string{"code": inv.Code})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/invite/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("redeem status %d body=%s", w.Code, w.Body.String())
	}
	var out map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["username"] == "" || out["password"] == "" {
		t.Fatalf("expected credentials, got %+v", out)
	}
	login, err := ts.GetPlatUserByLogin(context.Background(), out["username"])
	if err != nil || !morphdb.VerifyPassword(login.PasswordHash, out["password"]) {
		t.Fatalf("created user login failed: %v", err)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/invite/redeem", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusGone {
		t.Fatalf("expected 410 on double redeem, got %d", w2.Code)
	}
}

func TestPatchMorphAuthMeUpdatesUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := openAuthHarnessSQL(t)
	u := seedPlatUser(t, ts, "user@test.local", "oldname", "secret", false)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	token := bearerFor(t, h, u)

	router := gin.New()
	router.PATCH("/api/auth/me", h.PatchMorphAuthMe)

	body, _ := json.Marshal(map[string]string{"username": "newname"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/auth/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("patch status %d body=%s", w.Code, w.Body.String())
	}
	got, err := ts.GetPlatUserByUsername(context.Background(), "newname")
	if err != nil || got.ID != u.ID {
		t.Fatalf("username not updated: %v", err)
	}
}

func TestPatchMorphAuthMePasswordRequiresCurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := openAuthHarnessSQL(t)
	u := seedPlatUser(t, ts, "user2@test.local", "user2", "secret", false)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	token := bearerFor(t, h, u)

	router := gin.New()
	router.PATCH("/api/auth/me", h.PatchMorphAuthMe)

	body, _ := json.Marshal(map[string]string{"password": "newpass", "current_password": "wrong"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/auth/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
	}
}
