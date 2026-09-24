package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

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
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertAgentLesson(t, sqlDB, "a-on", "user-a", "sess-a", "enabled-a rule", true)
	insertAgentLesson(t, sqlDB, "a-off", "user-a", "sess-a-off", "disabled-a rule", false)
	insertAgentLesson(t, sqlDB, "b-on", "user-b", "sess-b", "secret-b rule", true)

	got := h.buildAgentLessonsContext("user-a")
	if !strings.Contains(got, "enabled-a rule") {
		t.Fatalf("missing enabled lesson:\n%s", got)
	}
	if strings.Contains(got, "disabled-a rule") || strings.Contains(got, "secret-b rule") {
		t.Fatalf("prompt included disabled or other user's lesson:\n%s", got)
	}
	if !strings.Contains(h.buildAgentLessonsContext("user-b"), "secret-b rule") {
		t.Fatal("owner b should see their enabled lesson")
	}
	if h.buildAgentLessonsContext("") != "" {
		t.Fatal("empty user must not receive unowned or any lessons")
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	req.Header.Set("X-User-ID", "user-a")
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
}

func TestAgentLessonPatchDeleteAndCrossUserIsolation(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertAgentLesson(t, sqlDB, "lesson-a", "user-a", "sess-a", "rule-a", true)
	insertAgentLesson(t, sqlDB, "lesson-b", "user-b", "sess-b", "rule-b", true)
	r := agentLessonRouter(h)

	list := func(user string) (int, string) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/agent-lessons", nil)
		req.Header.Set("X-User-ID", user)
		r.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}
	code, body := list("user-a")
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
	if listed.Total != 1 || len(listed.Lessons) != 1 || listed.Lessons[0].ID != "lesson-a" || !listed.Lessons[0].Enabled || listed.Lessons[0].Owner != "user-a" {
		t.Fatalf("list shape: %+v", listed)
	}

	patch := func(user, id, raw string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/agent-lessons/"+id, strings.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user)
		r.ServeHTTP(w, req)
		return w
	}
	if w := patch("user-a", "lesson-b", `{"enabled":false}`); w.Code != http.StatusNotFound {
		t.Fatalf("cross-user patch: %d %s", w.Code, w.Body.String())
	}
	still, err := h.TranMySQL.GetAgentLessonForOwner(context.Background(), "user-b", "lesson-b")
	if err != nil || still == nil || !still.Enabled {
		t.Fatalf("other user's lesson changed: %+v err=%v", still, err)
	}
	if w := patch("user-a", "missing", `{"enabled":false}`); w.Code != http.StatusNotFound {
		t.Fatalf("missing patch: %d %s", w.Code, w.Body.String())
	}
	if w := patch("user-a", "lesson-a", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("missing enabled: %d %s", w.Code, w.Body.String())
	}
	w := patch("user-a", "lesson-a", `{"enabled":false}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"enabled":false`) {
		t.Fatalf("patch: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(h.buildAgentLessonsContext("user-a"), "rule-a") {
		t.Fatal("disabled lesson still injected")
	}
	code, body = list("user-a")
	if code != http.StatusOK || !strings.Contains(body, `"enabled":false`) {
		t.Fatalf("disabled lesson missing from owner list: %d %s", code, body)
	}

	del := func(user, id string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/agent-lessons/"+id, nil)
		req.Header.Set("X-User-ID", user)
		r.ServeHTTP(w, req)
		return w
	}
	if w := del("user-a", "lesson-b"); w.Code != http.StatusNotFound {
		t.Fatalf("cross-user delete: %d %s", w.Code, w.Body.String())
	}
	still, err = h.TranMySQL.GetAgentLessonForOwner(context.Background(), "user-b", "lesson-b")
	if err != nil || still == nil {
		t.Fatalf("cross-user delete removed lesson: %+v err=%v", still, err)
	}
	if w := del("user-a", "lesson-a"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatalf("delete: %d %s", w.Code, w.Body.String())
	}
	if w := del("user-a", "lesson-a"); w.Code != http.StatusNotFound {
		t.Fatalf("second delete: %d %s", w.Code, w.Body.String())
	}
	code, body = list("user-a")
	if code != http.StatusOK || strings.Contains(body, "lesson-a") || !strings.Contains(body, `"total":0`) {
		t.Fatalf("list after delete: %d %s", code, body)
	}
	code, body = list("user-b")
	if code != http.StatusOK || !strings.Contains(body, "rule-b") {
		t.Fatalf("user b list: %d %s", code, body)
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
