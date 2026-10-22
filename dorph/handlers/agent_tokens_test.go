package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	dorphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func tokenRouter(t *testing.T) (*gin.Engine, *Handlers, *dorphdb.PlatUser, *dorphdb.PlatUser) {
	t.Helper()
	store, err := dorphdb.NewTranSQL(filepath.Join(t.TempDir(), "tran.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	h := &Handlers{TranMySQL: store, jwtCfg: auth.LoadTokenConfig()}
	ada := seedPlatUser(t, store, "ada@test.local", "ada", "secret", true)
	bea := seedPlatUser(t, store, "bea@test.local", "bea", "secret", false)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/agent-tokens", h.CreateAgentToken)
	r.GET("/api/agent-tokens", h.ListAgentTokens)
	r.DELETE("/api/agent-tokens/:id", h.RevokeAgentToken)
	return r, h, ada, bea
}

func TestAgentTokenCreateListRevoke(t *testing.T) {
	r, h, ada, bea := tokenRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/agent-tokens", bytes.NewBufferString(`{"name":"cursor"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no auth = %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/agent-tokens", bytes.NewBufferString(`{"name":"cursor"}`))
	req.Header.Set("Content-Type", "application/json")
	setBearer(t, h, req, ada)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create = %d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	tok, _ := created["token"].(string)
	id, _ := created["id"].(string)
	if !auth.IsPAT(tok) || id == "" {
		t.Fatalf("created = %#v", created)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/agent-tokens", nil)
	setBearer(t, h, req, ada)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), tok) {
		t.Fatalf("list leaked token: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), id) {
		t.Fatalf("list missing id: %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/agent-tokens", nil)
	setBearer(t, h, req, bea)
	r.ServeHTTP(w, req)
	if strings.Contains(w.Body.String(), id) {
		t.Fatalf("bea saw ada token: %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/agent-tokens/"+id, nil)
	setBearer(t, h, req, bea)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("bea revoke = %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/agent-tokens/"+id, nil)
	setBearer(t, h, req, ada)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ada revoke = %d %s", w.Code, w.Body.String())
	}
}
