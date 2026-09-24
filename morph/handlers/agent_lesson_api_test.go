package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func lessonHandlers(t *testing.T, sqlDB *sql.DB) (*Handlers, *morphdb.PlatUser, *morphdb.PlatUser) {
	t.Helper()
	ts := &morphdb.TranSQL{DB: sqlDB}
	if err := ts.EnsurePlatUsersTable(context.Background()); err != nil {
		t.Fatal(err)
	}
	h := &Handlers{TranMySQL: ts, jwtCfg: auth.LoadTokenConfig()}
	suffix := strings.ReplaceAll(t.Name(), "/", "-")
	a := seedPlatUser(t, ts, "a-"+suffix+"@test.local", "a-"+suffix, "secret", false)
	b := seedPlatUser(t, ts, "b-"+suffix+"@test.local", "b-"+suffix, "secret", false)
	return h, a, b
}

func setBearer(t *testing.T, h *Handlers, req *http.Request, u *morphdb.PlatUser) {
	t.Helper()
	req.Header.Set("Authorization", "Bearer "+bearerFor(t, h, u))
}

func agentLessonRouter(h *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/agent-lessons", h.ListAgentLessons)
	r.PATCH("/api/agent-lessons/:id", h.PatchAgentLesson)
	r.DELETE("/api/agent-lessons/:id", h.DeleteAgentLesson)
	r.GET("/api/skills", h.ListSkills)
	return r
}

func insertAgentLesson(t *testing.T, sqlDB *sql.DB, id, owner, session, rule string, enabled bool) {
	t.Helper()
	flag := 0
	if enabled {
		flag = 1
	}
	_, err := sqlDB.Exec(`
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at, enabled, owner_user_id)
		VALUES (?, 'when testing', ?, ?, '2026-01-01T00:00:00Z', ?, ?)`,
		id, rule, session, flag, owner)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildAgentLessonsContextOnlyEnabledForCurrentUser(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h, userA, userB := lessonHandlers(t, sqlDB)
	insertAgentLesson(t, sqlDB, "a-on", userA.ID, "sess-a", "enabled-a rule", true)
	insertAgentLesson(t, sqlDB, "a-off", userA.ID, "sess-a-off", "disabled-a rule", false)
	insertAgentLesson(t, sqlDB, "b-on", userB.ID, "sess-b", "secret-b rule", true)

	got := h.lessonsPromptForUser(userA.ID)
	if !strings.Contains(got, "enabled-a rule") {
		t.Fatalf("missing enabled lesson:\n%s", got)
	}
	if strings.Contains(got, "disabled-a rule") || strings.Contains(got, "secret-b rule") {
		t.Fatalf("prompt included disabled or other user's lesson:\n%s", got)
	}
	if !strings.Contains(h.lessonsPromptForUser(userB.ID), "secret-b rule") {
		t.Fatal("owner b should see their enabled lesson")
	}
	if h.lessonsPromptForUser("") != "" || h.buildAgentLessonsContext(nil) != "" {
		t.Fatal("empty or unauthenticated caller must not receive lessons")
	}

	spoof := httptest.NewRecorder()
	spoofReq := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	spoofReq.Header.Set("X-User-ID", userA.ID)
	agentLessonRouter(h).ServeHTTP(spoof, spoofReq)
	if spoof.Code != http.StatusOK || strings.Contains(spoof.Body.String(), "enabled-a rule") {
		t.Fatalf("skills listing trusted X-User-ID: %d %s", spoof.Code, spoof.Body.String())
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	req.Header.Set("X-User-ID", userB.ID)
	setBearer(t, h, req, userA)
	agentLessonRouter(h).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("skills status %d body %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "disabled-a rule") || strings.Contains(w.Body.String(), "secret-b rule") {
		t.Fatalf("skills listing leaked lessons: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "enabled-a rule") {
		t.Fatalf("skills listing missing enabled lesson: %s", w.Body.String())
	}

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	c.Request.Header.Set("X-User-ID", userB.ID)
	setBearer(t, h, c.Request, userA)
	prompt := h.buildAgentLessonsContext(c)
	if !strings.Contains(prompt, "enabled-a rule") || strings.Contains(prompt, "secret-b rule") || strings.Contains(prompt, "disabled-a rule") {
		t.Fatalf("prompt trusted the spoofed header:\n%s", prompt)
	}
}

func TestAgentLessonPatchDeleteAndCrossUserIsolation(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h, userA, userB := lessonHandlers(t, sqlDB)
	insertAgentLesson(t, sqlDB, "lesson-a", userA.ID, "sess-a", "rule-a", true)
	insertAgentLesson(t, sqlDB, "lesson-b", userB.ID, "sess-b", "rule-b", true)
	r := agentLessonRouter(h)

	list := func(user *morphdb.PlatUser) (int, string) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/agent-lessons", nil)
		req.Header.Set("X-User-ID", userB.ID)
		setBearer(t, h, req, user)
		r.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}
	code, body := list(userA)
	if code != http.StatusOK || !strings.Contains(body, "rule-a") || strings.Contains(body, "rule-b") {
		t.Fatalf("list isolation: %d %s", code, body)
	}
	var listed struct {
		Total   int `json:"total"`
		Lessons []struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
			Owner   string `json:"owner_user_id"`
		} `json:"lessons"`
	}
	if err := json.Unmarshal([]byte(body), &listed); err != nil {
		t.Fatal(err)
	}
	if listed.Total != 1 || len(listed.Lessons) != 1 || listed.Lessons[0].ID != "lesson-a" || !listed.Lessons[0].Enabled || listed.Lessons[0].Owner != userA.ID {
		t.Fatalf("list shape: %+v", listed)
	}

	patch := func(user *morphdb.PlatUser, id, raw string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/agent-lessons/"+id, strings.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userB.ID)
		setBearer(t, h, req, user)
		r.ServeHTTP(w, req)
		return w
	}
	if w := patch(userA, "lesson-b", `{"enabled":false}`); w.Code != http.StatusNotFound {
		t.Fatalf("cross-user patch: %d %s", w.Code, w.Body.String())
	}
	still, err := h.TranMySQL.GetAgentLessonForOwner(context.Background(), userB.ID, "lesson-b")
	if err != nil || still == nil || !still.Enabled {
		t.Fatalf("other user's lesson changed: %+v err=%v", still, err)
	}
	if w := patch(userA, "missing", `{"enabled":false}`); w.Code != http.StatusNotFound {
		t.Fatalf("missing patch: %d %s", w.Code, w.Body.String())
	}
	if w := patch(userA, "lesson-a", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("missing enabled: %d %s", w.Code, w.Body.String())
	}
	w := patch(userA, "lesson-a", `{"enabled":false}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"enabled":false`) {
		t.Fatalf("patch: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(h.lessonsPromptForUser(userA.ID), "rule-a") {
		t.Fatal("disabled lesson still injected")
	}
	code, body = list(userA)
	if code != http.StatusOK || !strings.Contains(body, `"enabled":false`) {
		t.Fatalf("disabled lesson missing from owner list: %d %s", code, body)
	}

	del := func(user *morphdb.PlatUser, id string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/agent-lessons/"+id, nil)
		req.Header.Set("X-User-ID", userB.ID)
		setBearer(t, h, req, user)
		r.ServeHTTP(w, req)
		return w
	}
	if w := del(userA, "lesson-b"); w.Code != http.StatusNotFound {
		t.Fatalf("cross-user delete: %d %s", w.Code, w.Body.String())
	}
	still, err = h.TranMySQL.GetAgentLessonForOwner(context.Background(), userB.ID, "lesson-b")
	if err != nil || still == nil {
		t.Fatalf("cross-user delete removed lesson: %+v err=%v", still, err)
	}
	if w := del(userA, "lesson-a"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatalf("delete: %d %s", w.Code, w.Body.String())
	}
	if w := del(userA, "lesson-a"); w.Code != http.StatusNotFound {
		t.Fatalf("second delete: %d %s", w.Code, w.Body.String())
	}
	code, body = list(userA)
	if code != http.StatusOK || strings.Contains(body, "lesson-a") || !strings.Contains(body, `"total":0`) {
		t.Fatalf("list after delete: %d %s", code, body)
	}
	code, body = list(userB)
	if code != http.StatusOK || !strings.Contains(body, "rule-b") {
		t.Fatalf("user b list: %d %s", code, body)
	}
}

func TestAgentLessonSpoofedHeaderIsRejected(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h, userA, userB := lessonHandlers(t, sqlDB)
	insertAgentLesson(t, sqlDB, "lesson-b", userB.ID, "sess-b", "secret-b rule", true)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h.AuthzMiddleware())
	r.GET("/api/agent-lessons", h.ListAgentLessons)
	r.PATCH("/api/agent-lessons/:id", h.PatchAgentLesson)
	r.DELETE("/api/agent-lessons/:id", h.DeleteAgentLesson)

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/agent-lessons", ""},
		{http.MethodPatch, "/api/agent-lessons/lesson-b", `{"enabled":false}`},
		{http.MethodDelete, "/api/agent-lessons/lesson-b", ""},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userB.ID)
		req.Header.Set("X-User-Role", "admin")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s spoofed header status %d body %s", tc.method, w.Code, w.Body.String())
		}
	}
	still, err := h.TranMySQL.GetAgentLessonForOwner(context.Background(), userB.ID, "lesson-b")
	if err != nil || still == nil || !still.Enabled {
		t.Fatalf("spoofed header changed lesson: %+v err=%v", still, err)
	}

	// auth_user_id copied from the header, and no bearer, is still untrusted.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/agent-lessons", nil)
	c.Request.Header.Set("X-User-ID", userB.ID)
	c.Set("auth_user_id", userB.ID)
	h.ListAgentLessons(c)
	if w.Code != http.StatusUnauthorized || strings.Contains(w.Body.String(), "secret-b rule") {
		t.Fatalf("context identity from header: %d %s", w.Code, w.Body.String())
	}

	// A real token wins over a spoofed header.
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/agent-lessons", nil)
	req.Header.Set("X-User-ID", userB.ID)
	setBearer(t, h, req, userA)
	agentLessonRouter(h).ServeHTTP(w, req)
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "secret-b rule") || !strings.Contains(w.Body.String(), `"total":0`) {
		t.Fatalf("bearer lost to spoofed header: %d %s", w.Code, w.Body.String())
	}
}

func TestAgentLessonMissingStoreWithoutBearerIsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/agent-lessons", nil)
	c.Request.Header.Set("X-User-ID", "someone-else")
	h.ListAgentLessons(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("header-only request with no store: %d %s", w.Code, w.Body.String())
	}
}

func TestAgentLessonBearerWithoutStoreIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/agent-lessons", nil)
	c.Request.Header.Set("Authorization", "Bearer not-a-real-token")
	c.Request.Header.Set("X-User-ID", "someone-else")
	h.ListAgentLessons(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("bearer with no store: %d %s", w.Code, w.Body.String())
	}
}

func TestAgentLessonsRequireSameAuthAsOperatorAPIs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	r := gin.New()
	r.Use(h.AuthzMiddleware())
	r.GET("/api/agent-lessons", h.ListAgentLessons)
	r.PATCH("/api/agent-lessons/:id", h.PatchAgentLesson)
	r.DELETE("/api/agent-lessons/:id", h.DeleteAgentLesson)

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/agent-lessons"},
		{http.MethodPatch, "/api/agent-lessons/x"},
		{http.MethodDelete, "/api/agent-lessons/x"},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"enabled":false}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status %d body %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}
