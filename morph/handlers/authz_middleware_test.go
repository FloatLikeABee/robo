package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func authzOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func openAuthzResearchDB(t *testing.T) *morphdb.TranSQL {
	t.Helper()
	db := openResearchDB(t)
	ts := &morphdb.TranSQL{DB: db}
	if err := ts.EnsurePlatUsersTable(context.Background()); err != nil {
		t.Fatal(err)
	}
	return ts
}

func authzEngine(h *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h.AuthzMiddleware())
	r.POST("/api/tran/research", h.CreateResearch)
	r.GET("/api/tran/research", authzOK)
	r.GET("/api/tran/research/:id", h.GetResearch)
	r.PATCH("/api/tran/research/:id", h.PatchResearchMarkdown)
	r.PUT("/api/tran/research/:id", authzOK)
	r.POST("/api/tran/research/:id/publish", h.PublishResearch)
	r.DELETE("/api/tran/research/:id", h.DeleteResearch)
	r.GET("/api/tran/public/research/:slug", h.ServePublicResearch)
	r.GET("/api/tran/public/big-notes/:slug", authzOK)
	r.GET("/api/tran/public/timelines/:slug", authzOK)
	r.POST("/api/forms/templates", authzOK)
	r.GET("/api/forms/templates", authzOK)
	r.POST("/api/knowledge/files", authzOK)
	r.GET("/api/knowledge/files", authzOK)
	r.POST("/api/graph/search", authzOK)
	r.GET("/api/graph/health", authzOK)
	r.POST("/api/auth/login", authzOK)
	return r
}

func authzHandlers(t *testing.T) (*Handlers, *gin.Engine, *morphdb.PlatUser) {
	t.Helper()
	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	user := seedPlatUser(t, ts, "notes@test.local", "notes", "secret", false)
	r := authzEngine(h)
	h.ginEngine = r
	return h, r, user
}

func doAuthz(r http.Handler, method, path, body, bearer string, extra http.Header) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	for k, vals := range extra {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAnonymousMorphDataMutationsReturn401(t *testing.T) {
	_, r, _ := authzHandlers(t)
	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/tran/research", `{"prompt":"unauth create"}`},
		{http.MethodPatch, "/api/tran/research/1", `{"markdown_content":"# x"}`},
		{http.MethodPut, "/api/tran/research/1", `{"markdown_content":"# x"}`},
		{http.MethodPost, "/api/tran/research/1/publish", ``},
		{http.MethodDelete, "/api/tran/research/1", ``},
		{http.MethodPost, "/api/forms/templates", `{"name":"x"}`},
		{http.MethodPost, "/api/knowledge/files", `{}`},
		{http.MethodPost, "/api/graph/search", `{"query":"x"}`},
		// A public prefix is not an open route. Only the published GET/HEAD pages are.
		{http.MethodPost, "/api/tran/public/research/some-slug", `{}`},
		{http.MethodGet, "/api/tran/public/other/some-slug", ``},
	}
	for _, tc := range cases {
		w := doAuthz(r, tc.method, tc.path, tc.body, "", nil)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status %d, want 401, body %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}

func TestAnonymousPrivateReadsStayOpen(t *testing.T) {
	_, r, _ := authzHandlers(t)
	for _, path := range []string{
		"/api/tran/research",
		"/api/forms/templates",
		"/api/knowledge/files",
		"/api/graph/health",
	} {
		w := doAuthz(r, http.MethodGet, path, "", "", nil)
		if w.Code == http.StatusUnauthorized {
			t.Errorf("GET %s should stay readable without a session, got 401 %s", path, w.Body.String())
		}
		if w.Code != http.StatusOK {
			t.Errorf("GET %s status %d, want 200, body %s", path, w.Code, w.Body.String())
		}
	}
}

func TestPublicPublishedPagesStayOpen(t *testing.T) {
	ts := openAuthzResearchDB(t)
	if _, err := ts.DB.Exec(
		`INSERT INTO research (user_id, owner_key, title, prompt, status, markdown_content, html_content, published_slug, published_path)
		 VALUES (1, 'anon', 'Public', 'p', 'complete', '# Hello', '<p>Hello</p>', 'public-topic', '/api/tran/public/research/public-topic')`,
	); err != nil {
		t.Fatal(err)
	}
	hh := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	engine := authzEngine(hh)

	for _, path := range []string{
		"/api/tran/public/research/public-topic",
		"/api/tran/public/big-notes/a-note",
		"/api/tran/public/timelines/a-timeline",
	} {
		w := doAuthz(engine, http.MethodGet, path, "", "", nil)
		if w.Code == http.StatusUnauthorized {
			t.Fatalf("GET %s should stay public, got 401 %s", path, w.Body.String())
		}
		if path == "/api/tran/public/research/public-topic" && w.Code != http.StatusOK {
			t.Fatalf("GET %s status %d body %s", path, w.Code, w.Body.String())
		}
		hw := doAuthz(engine, http.MethodHead, path, "", "", nil)
		if hw.Code == http.StatusUnauthorized {
			t.Fatalf("HEAD %s should stay public, got 401", path)
		}
	}

	// Login stays public.
	login := doAuthz(engine, http.MethodPost, "/api/auth/login", `{}`, "", nil)
	if login.Code == http.StatusUnauthorized {
		t.Fatalf("login should stay public, got 401")
	}
}

func TestJWTAllowsResearchCreateAndPublish(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })

	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	user := seedPlatUser(t, ts, "researcher@test.local", "researcher", "secret", false)
	r := authzEngine(h)
	token := bearerFor(t, h, user)

	created := doAuthz(r, http.MethodPost, "/api/tran/research", `{"prompt":"What is graphene?"}`, token, nil)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status %d body %s", created.Code, created.Body.String())
	}
	var doc researchDoc
	if err := json.Unmarshal(created.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ID <= 0 {
		t.Fatalf("expected created id, body %s", created.Body.String())
	}

	if _, err := ts.DB.Exec(`UPDATE research SET markdown_content = ? WHERE id = ?`, "# Graphene\n\nBody.", doc.ID); err != nil {
		t.Fatal(err)
	}
	id := strconv.Itoa(doc.ID)
	patched := doAuthz(r, http.MethodPatch, "/api/tran/research/"+id, `{"markdown_content":"# Graphene\n\nEdited."}`, token, nil)
	if patched.Code != http.StatusOK {
		t.Fatalf("patch status %d body %s", patched.Code, patched.Body.String())
	}
	published := doAuthz(r, http.MethodPost, "/api/tran/research/"+id+"/publish", ``, token, nil)
	if published.Code != http.StatusOK {
		t.Fatalf("publish status %d body %s", published.Code, published.Body.String())
	}
	var pub researchDoc
	if err := json.Unmarshal(published.Body.Bytes(), &pub); err != nil {
		t.Fatal(err)
	}
	if pub.PublishedSlug == nil || strings.TrimSpace(*pub.PublishedSlug) == "" {
		t.Fatalf("expected published slug, body %s", published.Body.String())
	}
	publicPath := "/api/tran/public/research/" + *pub.PublishedSlug
	open := doAuthz(r, http.MethodGet, publicPath, "", "", nil)
	if open.Code != http.StatusOK {
		t.Fatalf("public GET %s status %d body %s", publicPath, open.Code, open.Body.String())
	}
	if !strings.Contains(open.Body.String(), "Graphene") {
		t.Fatalf("public HTML missing content: %s", open.Body.String())
	}

	// Same JWT still allows a write on the sibling open prefixes.
	for _, path := range []string{"/api/forms/templates", "/api/knowledge/files", "/api/graph/search"} {
		w := doAuthz(r, http.MethodPost, path, `{"query":"x","name":"x"}`, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("POST %s with JWT status %d body %s", path, w.Code, w.Body.String())
		}
	}
}

func TestManagementAPIMutationUsesCallerJWT(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })

	h, _, user := authzHandlers(t)
	token := bearerFor(t, h, user)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	code, body := h.execManagementAPI(c, http.MethodPost, "/api/tran/research", "", []byte(`{"prompt":"from the tool loop"}`))
	if code != http.StatusCreated {
		t.Fatalf("internal create status %d body %s", code, body)
	}
}

// Header impersonation stays until issue #23. This guards the boundary of this story.
func TestLegacyUserIDHeaderStillAllowsMutation(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })

	_, r, _ := authzHandlers(t)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "7")
	hdr.Set("X-User-Role", "employee")
	w := doAuthz(r, http.MethodPost, "/api/tran/research", `{"prompt":"legacy header"}`, "", hdr)
	if w.Code != http.StatusCreated {
		t.Fatalf("legacy header create status %d body %s", w.Code, w.Body.String())
	}
}

func TestInvalidJWTDoesNotFallThroughToAnonymous(t *testing.T) {
	_, r, _ := authzHandlers(t)
	w := doAuthz(r, http.MethodPost, "/api/tran/research", `{"prompt":"bad token"}`, "not-a-jwt", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}
