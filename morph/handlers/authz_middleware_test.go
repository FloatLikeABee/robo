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
	r.POST("/api/chat", authzEcho)
	r.GET("/api/admin/users", authzOK)
	r.GET("/api/data-collector/entities", authzOK)
	return r
}

func authzEcho(c *gin.Context) {
	id, _ := c.Get("auth_user_id")
	role, _ := c.Get("auth_role")
	c.JSON(http.StatusOK, gin.H{"id": id, "role": role})
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

func countResearch(t *testing.T, ts *morphdb.TranSQL) int {
	t.Helper()
	var n int
	if err := ts.DB.QueryRow(`SELECT COUNT(*) FROM research`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestAnonymousMorphDataMutationsReturn401(t *testing.T) {
	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	r := authzEngine(h)
	before := countResearch(t, ts)
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
	if got := countResearch(t, ts); got != before {
		t.Fatalf("research row count %d, want %d", got, before)
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

func TestPublicMorphReadPathShape(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{name: "exact research", method: http.MethodGet, path: "/api/tran/public/research/slug", want: true},
		{name: "exact big note", method: http.MethodHead, path: "/api/tran/public/big-notes/a-note", want: true},
		{name: "exact timeline", method: http.MethodGet, path: "/api/tran/public/timelines/a-timeline", want: true},
		{name: "query is not part of the path", method: http.MethodGet, path: "/api/tran/public/research/slug", want: true},
		{name: "empty segment", method: http.MethodGet, path: "/api/tran/public/research//slug", want: false},
		{name: "trailing slash", method: http.MethodGet, path: "/api/tran/public/research/slug/", want: false},
		{name: "extra segment", method: http.MethodGet, path: "/api/tran/public/research/slug/extra", want: false},
		{name: "decoded percent-2F", method: http.MethodGet, path: "/api/tran/public/research/slug/extra", want: false},
		{name: "dot segment", method: http.MethodGet, path: "/api/tran/public/research/.", want: false},
		{name: "dotdot segment", method: http.MethodGet, path: "/api/tran/public/research/..", want: false},
		{name: "kind case", method: http.MethodGet, path: "/api/tran/public/Research/slug", want: false},
		{name: "post", method: http.MethodPost, path: "/api/tran/public/research/slug", want: false},
		{name: "unknown kind", method: http.MethodGet, path: "/api/tran/public/other/slug", want: false},
		{name: "empty slug", method: http.MethodGet, path: "/api/tran/public/research/", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPublicMorphRead(tc.method, tc.path); got != tc.want {
				t.Fatalf("isPublicMorphRead(%s, %s) = %v, want %v", tc.method, tc.path, got, tc.want)
			}
		})
	}

	ts := openAuthzResearchDB(t)
	if _, err := ts.DB.Exec(
		`INSERT INTO research (user_id, owner_key, title, prompt, status, markdown_content, html_content, published_slug, published_path)
		 VALUES (1, 'anon', 'Public', 'p', 'complete', '# Hello', '<p>Hello</p>', 'public-topic', '/api/tran/public/research/public-topic')`,
	); err != nil {
		t.Fatal(err)
	}
	hh := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	engine := authzEngine(hh)

	open := doAuthz(engine, http.MethodGet, "/api/tran/public/research/public-topic?utm=1", "", "", nil)
	if open.Code != http.StatusOK {
		t.Fatalf("query string GET status %d body %s", open.Code, open.Body.String())
	}

	loose := []string{
		"/api/tran/public/research//public-topic",
		"/api/tran/public/research/public-topic/",
		"/api/tran/public/research/public-topic/extra",
		"/api/tran/public/Research/public-topic",
		"/api/tran/public/research/.",
		"/api/tran/public/research/..",
		"/api/tran/public/research/slug%2Fextra",
	}
	for _, path := range loose {
		w := doAuthz(engine, http.MethodGet, path, "", "", nil)
		if strings.HasSuffix(path, "/") && !strings.Contains(path, "//") {
			if strings.Contains(w.Body.String(), "Hello") {
				t.Errorf("GET %s served published HTML: %s", path, w.Body.String())
			}
			if w.Code == http.StatusMovedPermanently {
				loc := w.Header().Get("Location")
				if loc != "/api/tran/public/research/public-topic" {
					t.Errorf("GET %s redirect location %q", path, loc)
				}
				continue
			}
		}
		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s status %d, want 401, body %s", path, w.Code, w.Body.String())
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
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	token := "signed-token-value"
	wantAuth := "Bearer " + token
	router := gin.New()
	router.POST("/api/tran/research", func(c *gin.Context) {
		if got := c.GetHeader("Authorization"); got != wantAuth {
			t.Errorf("Authorization %q, want %q", got, wantAuth)
		}
		if got := c.GetHeader("X-User-ID"); got != "" {
			t.Errorf("X-User-ID %q, want empty", got)
		}
		if got := c.GetHeader("X-User-Role"); got != "" {
			t.Errorf("X-User-Role %q, want empty", got)
		}
		c.Status(http.StatusCreated)
	})
	h.ginEngine = router

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", nil)
	c.Request.Header.Set("Authorization", wantAuth)
	c.Request.Header.Set("X-User-ID", "spoof")
	c.Request.Header.Set("X-User-Role", "admin")

	code, body := h.execManagementAPI(c, http.MethodPost, "/api/tran/research", "", []byte(`{"prompt":"from the tool loop"}`))
	if code != http.StatusCreated {
		t.Fatalf("internal create status %d body %s", code, body)
	}
}

func TestManagementAPIWithoutAuthorizationIs401(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })

	h, _, _ := authzHandlers(t)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", nil)
	c.Request.Header.Set("X-User-ID", "admin")
	c.Request.Header.Set("X-User-Role", "admin")

	code, body := h.execManagementAPI(c, http.MethodPost, "/api/tran/research", "", []byte(`{"prompt":"no token"}`))
	if code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401, body %s", code, body)
	}
}

func TestManagementAPICreatesResearchWithForwardedJWT(t *testing.T) {
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

func TestHeaderOnlyIdentityIsRejected(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })

	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	r := authzEngine(h)
	before := countResearch(t, ts)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "7")
	hdr.Set("X-User-Role", "admin")
	hdr.Set("X-User-Email", "spoof@example.com")

	for _, tc := range []struct {
		method, path, body string
	}{
		{http.MethodPost, "/api/chat", `{"message":"hi"}`},
		{http.MethodGet, "/api/admin/users", ``},
		{http.MethodGet, "/api/data-collector/entities", ``},
		{http.MethodPost, "/api/tran/research", `{"prompt":"header only"}`},
	} {
		w := doAuthz(r, tc.method, tc.path, tc.body, "", hdr)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status %d, want 401, body %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
	if got := countResearch(t, ts); got != before {
		t.Fatalf("research row count %d, want %d", got, before)
	}
}

func TestInvalidBearerPlusUserIDIs401(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })

	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	r := authzEngine(h)
	before := countResearch(t, ts)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "7")
	hdr.Set("X-User-Role", "admin")
	w := doAuthz(r, http.MethodPost, "/api/tran/research", `{"prompt":"bad token"}`, "not-a-jwt", hdr)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if got := countResearch(t, ts); got != before {
		t.Fatalf("research row count %d, want %d", got, before)
	}
}

func TestSpoofedHeadersDoNotOverrideToken(t *testing.T) {
	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	employee := seedPlatUser(t, ts, "employee@test.local", "employee", "secret", false)
	r := authzEngine(h)
	token := bearerFor(t, h, employee)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "someone-else")
	hdr.Set("X-User-Role", "admin")
	hdr.Set("X-User-Roles", "admin")
	hdr.Set("X-User-Email", "admin@example.com")

	admin := doAuthz(r, http.MethodGet, "/api/admin/users", "", token, hdr)
	if admin.Code != http.StatusForbidden {
		t.Fatalf("admin status %d, want 403, body %s", admin.Code, admin.Body.String())
	}

	chat := doAuthz(r, http.MethodPost, "/api/chat", `{"message":"hi"}`, token, hdr)
	if chat.Code != http.StatusOK {
		t.Fatalf("chat status %d body %s", chat.Code, chat.Body.String())
	}
	var echoed struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := json.Unmarshal(chat.Body.Bytes(), &echoed); err != nil {
		t.Fatal(err)
	}
	if echoed.ID != employee.ID {
		t.Fatalf("user id %q, want token user %q", echoed.ID, employee.ID)
	}
	if echoed.Role != "employee" {
		t.Fatalf("role %q, want employee", echoed.Role)
	}
}

func TestResolveUserScopeIgnoresIdentityHeaders(t *testing.T) {
	ts := openAuthzResearchDB(t)
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	employee := seedPlatUser(t, ts, "scope@test.local", "scopeuser", "secret", false)
	token := bearerFor(t, h, employee)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-User-ID", "someone-else")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-User-Email", "admin@example.com")

	scope, ok := h.ResolveUserScope(req)
	if !ok {
		t.Fatal("expected a session from the JWT")
	}
	if scope.UserID != employee.ID || scope.IsAdmin || scope.Role != "employee" || scope.Email != employee.Email {
		t.Fatalf("scope %+v, want user %s employee", scope, employee.ID)
	}

	headerOnly := httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	headerOnly.Header.Set("X-User-ID", employee.ID)
	headerOnly.Header.Set("X-User-Role", "admin")
	if _, ok := h.ResolveUserScope(headerOnly); ok {
		t.Fatal("identity headers without a JWT must not resolve a session")
	}
}

func TestOptionsIsNotUnauthorized(t *testing.T) {
	_, r, _ := authzHandlers(t)
	w := doAuthz(r, http.MethodOptions, "/api/chat", "", "", nil)
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("OPTIONS /api/chat status 401, want pass-through")
	}
}

func TestInvalidJWTDoesNotFallThroughToAnonymous(t *testing.T) {
	_, r, _ := authzHandlers(t)
	w := doAuthz(r, http.MethodPost, "/api/tran/research", `{"prompt":"bad token"}`, "not-a-jwt", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}
