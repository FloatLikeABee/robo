package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func managementLessonCaller(t *testing.T) (*Handlers, *morphdb.PlatUser) {
	t.Helper()
	managementExactQueryCache.Flush()
	t.Cleanup(func() { managementExactQueryCache.Flush() })
	sqlDB := openHarnessSQL(t)
	h, userA, _ := lessonHandlers(t, sqlDB)
	database, err := morphdb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	h.db = database
	return h, userA
}

func managementAsk(t *testing.T, h *Handlers, user *morphdb.PlatUser, bearer bool, userID, prompt string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", nil)
	if bearer {
		setBearer(t, h, c.Request, user)
	}
	c.Request.Header.Set("X-User-ID", userID)
	reply, _, err := h.chatWithManagementTools(c, userID, "default", prompt, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return reply
}

func waitExactCache(t *testing.T, reply string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, item := range managementExactQueryCache.Items() {
			if s, ok := item.Object.(string); ok && s == reply {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("exact-query cache was not populated")
}

func TestManagementExactCacheMissesAfterLessonDisabled(t *testing.T) {
	h, userA := managementLessonCaller(t)
	insertAgentLesson(t, h.TranMySQL.DB, "lesson-a", userA.ID, "sess-a", "polluting rule", true)
	calls := 0
	h.managementReply = func(ctx context.Context, prompt string) (string, error) {
		calls++
		return "stale-with-lesson", nil
	}
	prompt := "what is the refund policy for this exact cache test"
	if got := managementAsk(t, h, userA, true, "someone-else", prompt); got != "stale-with-lesson" {
		t.Fatalf("first reply %q", got)
	}
	waitExactCache(t, "stale-with-lesson")
	before := calls
	if got := managementAsk(t, h, userA, true, "someone-else", prompt); calls != before || got != "stale-with-lesson" {
		t.Fatalf("repeat missed cache: %q calls %d -> %d", got, before, calls)
	}
	if _, err := h.TranMySQL.SetAgentLessonEnabled(context.Background(), userA.ID, "lesson-a", false); err != nil {
		t.Fatal(err)
	}
	h.managementReply = func(ctx context.Context, prompt string) (string, error) {
		calls++
		return "fresh-without-lesson", nil
	}
	if got := managementAsk(t, h, userA, true, "someone-else", prompt); got != "fresh-without-lesson" {
		t.Fatalf("disabled lesson still served cache: %q calls=%d", got, calls)
	}
}

func TestManagementExactCacheIgnoresUntrustedUserID(t *testing.T) {
	h, userA := managementLessonCaller(t)
	calls := 0
	h.managementReply = func(ctx context.Context, prompt string) (string, error) {
		calls++
		return "bearer-reply", nil
	}
	prompt := "cached only for the bearer on this prompt"
	if got := managementAsk(t, h, userA, true, userA.ID, prompt); got != "bearer-reply" {
		t.Fatalf("bearer reply %q", got)
	}
	waitExactCache(t, "bearer-reply")
	before := calls
	if got := managementAsk(t, h, userA, true, userA.ID, prompt); calls != before || got != "bearer-reply" {
		t.Fatalf("bearer repeat missed cache: %q calls %d -> %d", got, before, calls)
	}
	h.managementReply = func(ctx context.Context, prompt string) (string, error) {
		calls++
		return "header-reply", nil
	}
	if got := managementAsk(t, h, userA, false, userA.ID, prompt); got != "header-reply" {
		t.Fatalf("spoofed header read the bearer cache: %q", got)
	}
}
